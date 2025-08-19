package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)


// AnthropicProvider implements LLMProvider for Anthropic Claude
type AnthropicProvider struct {
	*BaseProvider
	client  *http.Client
	baseURL string
	tracer  trace.Tracer
}

// AnthropicRequest represents a request to Anthropic API
type AnthropicRequest struct {
	Model       string                 `json:"model"`
	MaxTokens   int                    `json:"max_tokens"`
	Messages    []AnthropicMessage     `json:"messages"`
	System      string                 `json:"system,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	TopP        float64                `json:"top_p,omitempty"`
	Tools       []AnthropicTool        `json:"tools,omitempty"`
	ToolChoice  interface{}            `json:"tool_choice,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AnthropicMessage represents a message in Anthropic format
type AnthropicMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // Can be string or array of content blocks
}

// AnthropicTool represents a tool in Anthropic format
type AnthropicTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// AnthropicResponse represents a response from Anthropic API
type AnthropicResponse struct {
	ID           string             `json:"id"`
	Type         string             `json:"type"`
	Role         string             `json:"role"`
	Content      []AnthropicContent `json:"content"`
	Model        string             `json:"model"`
	StopReason   string             `json:"stop_reason"`
	StopSequence string             `json:"stop_sequence,omitempty"`
	Usage        AnthropicUsage     `json:"usage"`
}

// AnthropicContent represents content in Anthropic response
type AnthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	// Tool use fields
	ID    string                 `json:"id,omitempty"`
	Name  string                 `json:"name,omitempty"`
	Input map[string]interface{} `json:"input,omitempty"`
}

// AnthropicUsage represents usage information from Anthropic
type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(config *LLMConfig) (LLMProvider, error) {
	base := NewBaseProvider(config, "anthropic")
	if err := base.ValidateConfig(); err != nil {
		return nil, err
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	client := &http.Client{
		Timeout: timeout,
	}

	tracer := otel.Tracer("llm.anthropic")

	return &AnthropicProvider{
		BaseProvider: base,
		client:       client,
		baseURL:      baseURL,
		tracer:       tracer,
	}, nil
}

// GenerateResponse generates a single response using Anthropic
func (p *AnthropicProvider) GenerateResponse(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	ctx, span := p.tracer.Start(ctx, "anthropic.generate_response")
	defer span.End()

	span.SetAttributes(
		attribute.String("llm.provider", "anthropic"),
		attribute.String("llm.model", req.Model),
		attribute.Int("llm.max_tokens", req.MaxTokens),
		attribute.Float64("llm.temperature", req.Temperature),
	)

	prepared := p.PrepareRequest(req)

	var response *GenerateResponse
	err := p.WithRetry(ctx, func() error {
		anthropicReq := p.convertToAnthropicRequest(prepared)

		resp, err := p.makeRequest(ctx, "/messages", anthropicReq)
		if err != nil {
			return err
		}

		response = p.convertFromAnthropicResponse(resp)
		return nil
	})

	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Record usage metrics
	if response.Usage.TotalTokens > 0 {
		span.SetAttributes(
			attribute.Int("llm.usage.prompt_tokens", response.Usage.PromptTokens),
			attribute.Int("llm.usage.completion_tokens", response.Usage.CompletionTokens),
			attribute.Int("llm.usage.total_tokens", response.Usage.TotalTokens),
		)
	}

	return response, nil
}

// StreamResponse generates a streaming response using Anthropic
func (p *AnthropicProvider) StreamResponse(ctx context.Context, req *GenerateRequest) (<-chan *StreamChunk, error) {
	ctx, span := p.tracer.Start(ctx, "anthropic.stream_response")
	defer span.End()

	span.SetAttributes(
		attribute.String("llm.provider", "anthropic"),
		attribute.String("llm.model", req.Model),
		attribute.Bool("llm.stream", true),
	)

	prepared := p.PrepareRequest(req)
	prepared.Stream = true

	// Anthropic streaming implementation would go here
	// For now, return an error indicating streaming is not implemented
	return nil, NewLLMError(
		"not_implemented",
		"streaming is not yet implemented for Anthropic provider",
		"feature_error",
		"anthropic",
	)
}

// GetModels returns available Anthropic models
func (p *AnthropicProvider) GetModels(ctx context.Context) ([]string, error) {
	// Anthropic doesn't have a models endpoint, so return known models
	return []string{
		"claude-3-5-sonnet-20241022",
		"claude-3-5-haiku-20241022",
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
	}, nil
}

// makeRequest makes an HTTP request to Anthropic API
func (p *AnthropicProvider) makeRequest(ctx context.Context, endpoint string, reqBody interface{}) (*AnthropicResponse, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, p.handleAnthropicError(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, p.handleHTTPError(resp.StatusCode, body)
	}

	var anthropicResp AnthropicResponse
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &anthropicResp, nil
}

// convertToAnthropicRequest converts our request format to Anthropic format
func (p *AnthropicProvider) convertToAnthropicRequest(req *GenerateRequest) *AnthropicRequest {
	messages := make([]AnthropicMessage, 0, len(req.Messages))

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			continue // System messages are handled separately in Anthropic
		}

		messages = append(messages, AnthropicMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	anthropicReq := &AnthropicRequest{
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Messages:    messages,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stream:      req.Stream,
	}

	// Set system prompt
	if req.SystemPrompt != "" {
		anthropicReq.System = req.SystemPrompt
	} else {
		// Check for system message in messages
		for _, msg := range req.Messages {
			if msg.Role == "system" {
				anthropicReq.System = msg.Content
				break
			}
		}
	}

	// Convert tools if present
	if len(req.Tools) > 0 {
		tools := make([]AnthropicTool, len(req.Tools))
		for i, tool := range req.Tools {
			tools[i] = AnthropicTool{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				InputSchema: tool.Function.Parameters,
			}
		}
		anthropicReq.Tools = tools

		if req.ToolChoice != nil {
			anthropicReq.ToolChoice = req.ToolChoice
		}
	}

	return anthropicReq
}

// convertFromAnthropicResponse converts Anthropic response to our format
func (p *AnthropicProvider) convertFromAnthropicResponse(resp *AnthropicResponse) *GenerateResponse {
	// Combine all text content
	var content string
	var toolCalls []ToolCall

	for _, c := range resp.Content {
		if c.Type == "text" {
			content += c.Text
		} else if c.Type == "tool_use" {
			// Convert tool use to tool call
			argsJSON, _ := json.Marshal(c.Input)
			toolCalls = append(toolCalls, ToolCall{
				ID:   c.ID,
				Type: "function",
				Function: Function{
					Name:      c.Name,
					Arguments: string(argsJSON),
				},
			})
		}
	}

	message := Message{
		Role:      resp.Role,
		Content:   content,
		ToolCalls: toolCalls,
	}

	choice := Choice{
		Index:        0,
		Message:      message,
		FinishReason: resp.StopReason,
	}

	return &GenerateResponse{
		ID:      resp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   resp.Model,
		Choices: []Choice{choice},
		Usage: Usage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
	}
}

// handleAnthropicError converts Anthropic errors to our error format
func (p *AnthropicProvider) handleAnthropicError(err error) error {
	return NewLLMError(
		"request_error",
		err.Error(),
		"client_error",
		"anthropic",
	)
}

// handleHTTPError handles HTTP errors from Anthropic API
func (p *AnthropicProvider) handleHTTPError(statusCode int, body []byte) error {
	var errorResp struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(body, &errorResp); err != nil {
		return NewLLMError(
			fmt.Sprintf("http_%d", statusCode),
			string(body),
			"api_error",
			"anthropic",
		)
	}

	var code string
	switch statusCode {
	case 400:
		code = "invalid_request"
	case 401:
		code = "authentication_error"
	case 403:
		code = "permission_error"
	case 429:
		code = "rate_limit_exceeded"
	case 500, 502, 503, 504:
		code = "server_error"
	default:
		code = fmt.Sprintf("http_%d", statusCode)
	}

	return NewLLMError(
		code,
		errorResp.Message,
		"api_error",
		"anthropic",
	)
}

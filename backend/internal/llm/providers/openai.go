package providers

import (
	"context"
	"net/http"
	"time"

	"github.com/sashabaranov/go-openai"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// OpenAIProvider implements LLMProvider for OpenAI
type OpenAIProvider struct {
	*BaseProvider
	client *openai.Client
	tracer trace.Tracer
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config *LLMConfig) (LLMProvider, error) {
	base := NewBaseProvider(config, "openai")
	if err := base.ValidateConfig(); err != nil {
		return nil, err
	}

	// Create OpenAI client configuration
	clientConfig := openai.DefaultConfig(config.APIKey)
	if config.BaseURL != "" {
		clientConfig.BaseURL = config.BaseURL
	}

	// Set timeout if specified
	timeout := 30 * time.Second
	if config.Timeout > 0 {
		timeout = config.Timeout
	}
	clientConfig.HTTPClient = &http.Client{
		Timeout: timeout,
	}

	client := openai.NewClientWithConfig(clientConfig)
	tracer := otel.Tracer("llm.openai")

	return &OpenAIProvider{
		BaseProvider: base,
		client:       client,
		tracer:       tracer,
	}, nil
}

// GenerateResponse generates a single response using OpenAI
func (p *OpenAIProvider) GenerateResponse(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	ctx, span := p.tracer.Start(ctx, "openai.generate_response")
	defer span.End()

	span.SetAttributes(
		attribute.String("llm.provider", "openai"),
		attribute.String("llm.model", req.Model),
		attribute.Int("llm.max_tokens", req.MaxTokens),
		attribute.Float64("llm.temperature", req.Temperature),
	)

	prepared := p.PrepareRequest(req)

	var response *GenerateResponse
	err := p.WithRetry(ctx, func() error {
		openaiReq := p.convertToOpenAIRequest(prepared)

		resp, err := p.client.CreateChatCompletion(ctx, openaiReq)
		if err != nil {
			return p.handleOpenAIError(err)
		}

		response = p.convertFromOpenAIResponse(&resp)
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

// StreamResponse generates a streaming response using OpenAI
func (p *OpenAIProvider) StreamResponse(ctx context.Context, req *GenerateRequest) (<-chan *StreamChunk, error) {
	ctx, span := p.tracer.Start(ctx, "openai.stream_response")
	defer span.End()

	span.SetAttributes(
		attribute.String("llm.provider", "openai"),
		attribute.String("llm.model", req.Model),
		attribute.Bool("llm.stream", true),
	)

	prepared := p.PrepareRequest(req)
	prepared.Stream = true

	openaiReq := p.convertToOpenAIRequest(prepared)

	stream, err := p.client.CreateChatCompletionStream(ctx, openaiReq)
	if err != nil {
		span.RecordError(err)
		return nil, p.handleOpenAIError(err)
	}

	chunks := make(chan *StreamChunk, 10)

	go func() {
		defer close(chunks)
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if err != nil {
				if err.Error() == "EOF" || err.Error() == "io: read/write on closed pipe" {
					return
				}
				// Send error as a chunk (could be handled differently)
				return
			}

			chunk := p.convertFromOpenAIStreamResponse(&response)
			select {
			case chunks <- chunk:
			case <-ctx.Done():
				return
			}
		}
	}()

	return chunks, nil
}

// GetModels returns available OpenAI models
func (p *OpenAIProvider) GetModels(ctx context.Context) ([]string, error) {
	ctx, span := p.tracer.Start(ctx, "openai.get_models")
	defer span.End()

	models, err := p.client.ListModels(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, p.handleOpenAIError(err)
	}

	modelNames := make([]string, len(models.Models))
	for i, model := range models.Models {
		modelNames[i] = model.ID
	}

	return modelNames, nil
}

// convertToOpenAIRequest converts our request format to OpenAI format
func (p *OpenAIProvider) convertToOpenAIRequest(req *GenerateRequest) openai.ChatCompletionRequest {
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}

		// Convert tool calls if present
		if len(msg.ToolCalls) > 0 {
			toolCalls := make([]openai.ToolCall, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				toolCalls[j] = openai.ToolCall{
					ID:   tc.ID,
					Type: openai.ToolType(tc.Type),
					Function: openai.FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
			messages[i].ToolCalls = toolCalls
		}

		if msg.ToolCallID != "" {
			messages[i].ToolCallID = msg.ToolCallID
		}
	}

	// Add system prompt as first message if specified
	if req.SystemPrompt != "" {
		systemMessages := p.AddSystemMessage(req.Messages, req.SystemPrompt)
		messages = make([]openai.ChatCompletionMessage, len(systemMessages))
		for i, msg := range systemMessages {
			messages[i] = openai.ChatCompletionMessage{
				Role:    msg.Role,
				Content: msg.Content,
			}
		}
	}

	openaiReq := openai.ChatCompletionRequest{
		Model:       req.Model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: float32(req.Temperature),
		TopP:        float32(req.TopP),
		Stream:      req.Stream,
	}

	// Convert tools if present
	if len(req.Tools) > 0 {
		tools := make([]openai.Tool, len(req.Tools))
		for i, tool := range req.Tools {
			tools[i] = openai.Tool{
				Type: openai.ToolType(tool.Type),
				Function: &openai.FunctionDefinition{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  tool.Function.Parameters,
				},
			}
		}
		openaiReq.Tools = tools

		if req.ToolChoice != nil {
			openaiReq.ToolChoice = req.ToolChoice
		}
	}

	return openaiReq
}

// convertFromOpenAIResponse converts OpenAI response to our format
func (p *OpenAIProvider) convertFromOpenAIResponse(resp *openai.ChatCompletionResponse) *GenerateResponse {
	choices := make([]Choice, len(resp.Choices))
	for i, choice := range resp.Choices {
		message := Message{
			Role:    choice.Message.Role,
			Content: choice.Message.Content,
		}

		// Convert tool calls if present
		if len(choice.Message.ToolCalls) > 0 {
			toolCalls := make([]ToolCall, len(choice.Message.ToolCalls))
			for j, tc := range choice.Message.ToolCalls {
				toolCalls[j] = ToolCall{
					ID:   tc.ID,
					Type: string(tc.Type),
					Function: struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					}{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
			message.ToolCalls = toolCalls
		}

		choices[i] = Choice{
			Index:        choice.Index,
			Message:      message,
			FinishReason: string(choice.FinishReason),
		}
	}

	return &GenerateResponse{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
		Choices: choices,
		Usage: Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
		SystemFingerprint: resp.SystemFingerprint,
	}
}

// convertFromOpenAIStreamResponse converts OpenAI stream response to our format
func (p *OpenAIProvider) convertFromOpenAIStreamResponse(resp *openai.ChatCompletionStreamResponse) *StreamChunk {
	choices := make([]StreamChoice, len(resp.Choices))
	for i, choice := range resp.Choices {
		delta := MessageDelta{
			Role:    choice.Delta.Role,
			Content: choice.Delta.Content,
		}

		// Convert tool calls if present
		if len(choice.Delta.ToolCalls) > 0 {
			toolCalls := make([]ToolCall, len(choice.Delta.ToolCalls))
			for j, tc := range choice.Delta.ToolCalls {
				toolCalls[j] = ToolCall{
					ID:   tc.ID,
					Type: string(tc.Type),
					Function: struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					}{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
			delta.ToolCalls = toolCalls
		}

		var finishReason *string
		if choice.FinishReason != "" {
			fr := string(choice.FinishReason)
			finishReason = &fr
		}

		choices[i] = StreamChoice{
			Index:        choice.Index,
			Delta:        delta,
			FinishReason: finishReason,
		}
	}

	var usage *Usage
	if resp.Usage != nil {
		usage = &Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
	}

	return &StreamChunk{
		ID:                resp.ID,
		Object:            resp.Object,
		Created:           resp.Created,
		Model:             resp.Model,
		Choices:           choices,
		Usage:             usage,
		SystemFingerprint: resp.SystemFingerprint,
	}
}

// handleOpenAIError converts OpenAI errors to our error format
func (p *OpenAIProvider) handleOpenAIError(err error) error {
	if apiErr, ok := err.(*openai.APIError); ok {
		code := "unknown"
		if apiErr.Code != nil {
			if codeStr, ok := apiErr.Code.(string); ok {
				code = codeStr
			}
		}
		return NewLLMError(
			code,
			apiErr.Message,
			apiErr.Type,
			"openai",
		)
	}

	return NewLLMError(
		"unknown_error",
		err.Error(),
		"client_error",
		"openai",
	)
}

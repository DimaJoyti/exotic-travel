package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// OllamaClient handles communication with Ollama API
type OllamaClient struct {
	baseURL    string
	httpClient *http.Client
	tracer     trace.Tracer
}

// OllamaGenerateRequest represents a request to Ollama's generate endpoint
type OllamaGenerateRequest struct {
	Model    string                 `json:"model"`
	Prompt   string                 `json:"prompt"`
	System   string                 `json:"system,omitempty"`
	Template string                 `json:"template,omitempty"`
	Context  []int                  `json:"context,omitempty"`
	Stream   bool                   `json:"stream"`
	Raw      bool                   `json:"raw,omitempty"`
	Format   string                 `json:"format,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// OllamaChatRequest represents a request to Ollama's chat endpoint
type OllamaChatRequest struct {
	Model    string                 `json:"model"`
	Messages []OllamaChatMessage    `json:"messages"`
	Stream   bool                   `json:"stream"`
	Format   string                 `json:"format,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// OllamaChatMessage represents a chat message
type OllamaChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OllamaResponse represents a response from Ollama
type OllamaResponse struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Response           string    `json:"response"`
	Done               bool      `json:"done"`
	Context            []int     `json:"context,omitempty"`
	TotalDuration      int64     `json:"total_duration,omitempty"`
	LoadDuration       int64     `json:"load_duration,omitempty"`
	PromptEvalCount    int       `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64     `json:"prompt_eval_duration,omitempty"`
	EvalCount          int       `json:"eval_count,omitempty"`
	EvalDuration       int64     `json:"eval_duration,omitempty"`
}

// OllamaChatResponse represents a chat response from Ollama
type OllamaChatResponse struct {
	Model     string            `json:"model"`
	CreatedAt time.Time         `json:"created_at"`
	Message   OllamaChatMessage `json:"message"`
	Done      bool              `json:"done"`
}

// OllamaModel represents a model in Ollama
type OllamaModel struct {
	Name       string            `json:"name"`
	ModifiedAt time.Time         `json:"modified_at"`
	Size       int64             `json:"size"`
	Digest     string            `json:"digest"`
	Details    OllamaModelDetail `json:"details"`
}

// OllamaModelDetail contains model details
type OllamaModelDetail struct {
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

// OllamaModelsResponse represents the response from the models endpoint
type OllamaModelsResponse struct {
	Models []OllamaModel `json:"models"`
}

// OllamaPullRequest represents a request to pull a model
type OllamaPullRequest struct {
	Name   string `json:"name"`
	Stream bool   `json:"stream"`
}

// OllamaPullResponse represents a response from pulling a model
type OllamaPullResponse struct {
	Status    string `json:"status"`
	Digest    string `json:"digest,omitempty"`
	Total     int64  `json:"total,omitempty"`
	Completed int64  `json:"completed,omitempty"`
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient(baseURL string, timeout time.Duration) *OllamaClient {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	return &OllamaClient{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
		tracer: otel.Tracer("ollama.client"),
	}
}

// Generate sends a generate request to Ollama
func (c *OllamaClient) Generate(ctx context.Context, req *OllamaGenerateRequest) (*OllamaResponse, error) {
	ctx, span := c.tracer.Start(ctx, "ollama.generate")
	defer span.End()

	span.SetAttributes(
		attribute.String("ollama.model", req.Model),
		attribute.Bool("ollama.stream", req.Stream),
	)

	url := fmt.Sprintf("%s/api/generate", c.baseURL)
	result, err := c.sendRequest(ctx, url, req, &OllamaResponse{})
	if err != nil {
		return nil, err
	}
	return result.(*OllamaResponse), nil
}

// Chat sends a chat request to Ollama
func (c *OllamaClient) Chat(ctx context.Context, req *OllamaChatRequest) (*OllamaChatResponse, error) {
	ctx, span := c.tracer.Start(ctx, "ollama.chat")
	defer span.End()

	span.SetAttributes(
		attribute.String("ollama.model", req.Model),
		attribute.Bool("ollama.stream", req.Stream),
		attribute.Int("ollama.messages_count", len(req.Messages)),
	)

	url := fmt.Sprintf("%s/api/chat", c.baseURL)
	result, err := c.sendRequest(ctx, url, req, &OllamaChatResponse{})
	if err != nil {
		return nil, err
	}
	return result.(*OllamaChatResponse), nil
}

// GenerateStream sends a streaming generate request to Ollama
func (c *OllamaClient) GenerateStream(ctx context.Context, req *OllamaGenerateRequest) (<-chan *OllamaResponse, error) {
	ctx, span := c.tracer.Start(ctx, "ollama.generate_stream")
	defer span.End()

	req.Stream = true
	url := fmt.Sprintf("%s/api/generate", c.baseURL)
	return c.sendStreamRequest(ctx, url, req)
}

// ChatStream sends a streaming chat request to Ollama
func (c *OllamaClient) ChatStream(ctx context.Context, req *OllamaChatRequest) (<-chan *OllamaChatResponse, error) {
	ctx, span := c.tracer.Start(ctx, "ollama.chat_stream")
	defer span.End()

	req.Stream = true
	url := fmt.Sprintf("%s/api/chat", c.baseURL)
	return c.sendChatStreamRequest(ctx, url, req)
}

// ListModels lists available models
func (c *OllamaClient) ListModels(ctx context.Context) (*OllamaModelsResponse, error) {
	ctx, span := c.tracer.Start(ctx, "ollama.list_models")
	defer span.End()

	url := fmt.Sprintf("%s/api/tags", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		err := fmt.Errorf("ollama API error: %d - %s", resp.StatusCode, string(body))
		span.RecordError(err)
		return nil, err
	}

	var result OllamaModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	span.SetAttributes(attribute.Int("ollama.models_count", len(result.Models)))
	return &result, nil
}

// PullModel pulls a model from Ollama registry
func (c *OllamaClient) PullModel(ctx context.Context, modelName string) error {
	ctx, span := c.tracer.Start(ctx, "ollama.pull_model")
	defer span.End()

	span.SetAttributes(attribute.String("ollama.model_name", modelName))

	req := &OllamaPullRequest{
		Name:   modelName,
		Stream: false,
	}

	url := fmt.Sprintf("%s/api/pull", c.baseURL)
	_, err := c.sendRequest(ctx, url, req, &OllamaPullResponse{})
	return err
}

// DeleteModel deletes a model
func (c *OllamaClient) DeleteModel(ctx context.Context, modelName string) error {
	ctx, span := c.tracer.Start(ctx, "ollama.delete_model")
	defer span.End()

	span.SetAttributes(attribute.String("ollama.model_name", modelName))

	url := fmt.Sprintf("%s/api/delete", c.baseURL)
	req := map[string]string{"name": modelName}
	
	_, err := c.sendRequest(ctx, url, req, nil)
	return err
}

// Health checks if Ollama is healthy
func (c *OllamaClient) Health(ctx context.Context) error {
	ctx, span := c.tracer.Start(ctx, "ollama.health")
	defer span.End()

	url := fmt.Sprintf("%s/api/tags", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("ollama health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("ollama health check failed with status: %d", resp.StatusCode)
		span.RecordError(err)
		return err
	}

	return nil
}

// sendRequest sends a request to Ollama and decodes the response
func (c *OllamaClient) sendRequest(ctx context.Context, url string, reqBody interface{}, respBody interface{}) (interface{}, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error: %d - %s", resp.StatusCode, string(body))
	}

	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return respBody, nil
	}

	return nil, nil
}

// sendStreamRequest sends a streaming request to Ollama
func (c *OllamaClient) sendStreamRequest(ctx context.Context, url string, reqBody interface{}) (<-chan *OllamaResponse, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("ollama API error: %d - %s", resp.StatusCode, string(body))
	}

	ch := make(chan *OllamaResponse)
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		for {
			var response OllamaResponse
			if err := decoder.Decode(&response); err != nil {
				if err != io.EOF {
					// Log error but don't send to channel to avoid blocking
					fmt.Printf("Error decoding stream response: %v\n", err)
				}
				return
			}

			select {
			case ch <- &response:
				if response.Done {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

// sendChatStreamRequest sends a streaming chat request to Ollama
func (c *OllamaClient) sendChatStreamRequest(ctx context.Context, url string, reqBody interface{}) (<-chan *OllamaChatResponse, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("ollama API error: %d - %s", resp.StatusCode, string(body))
	}

	ch := make(chan *OllamaChatResponse)
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		for {
			var response OllamaChatResponse
			if err := decoder.Decode(&response); err != nil {
				if err != io.EOF {
					fmt.Printf("Error decoding chat stream response: %v\n", err)
				}
				return
			}

			select {
			case ch <- &response:
				if response.Done {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

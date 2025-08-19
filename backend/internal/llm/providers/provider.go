package providers

import (
	"context"
	"fmt"
	"time"
)

// Local type definitions to avoid import cycles

// Message represents a conversation message
type Message struct {
	Role       string                 `json:"role"`      // "system", "user", "assistant", "tool"
	Content    string                 `json:"content"`
	Name       string                 `json:"name,omitempty"`
	ToolCalls  []ToolCall             `json:"tool_calls,omitempty"`
	ToolCallID string                 `json:"tool_call_id,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ToolCall represents a function/tool call from the LLM
type ToolCall struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"` // "function"
	Function Function `json:"function"`
}

// Function represents a function call
type Function struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string
}

// Tool represents an external tool that can be called by the LLM
type Tool struct {
	Type     string       `json:"type"` // "function"
	Function ToolFunction `json:"function"`
}

// ToolFunction defines a function that can be called
type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"` // JSON Schema
}

// GenerateRequest represents a request to generate text
type GenerateRequest struct {
	Messages     []Message              `json:"messages"`
	Model        string                 `json:"model,omitempty"`
	MaxTokens    int                    `json:"max_tokens,omitempty"`
	Temperature  float64                `json:"temperature,omitempty"`
	TopP         float64                `json:"top_p,omitempty"`
	SystemPrompt string                 `json:"system_prompt,omitempty"`
	Tools        []Tool                 `json:"tools,omitempty"`
	ToolChoice   interface{}            `json:"tool_choice,omitempty"`
	Stream       bool                   `json:"stream,omitempty"`
	Stop         []string               `json:"stop,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// GenerateResponse represents the response from text generation
type GenerateResponse struct {
	ID                string                 `json:"id"`
	Object            string                 `json:"object"`
	Created           int64                  `json:"created"`
	Model             string                 `json:"model"`
	Choices           []Choice               `json:"choices"`
	Usage             Usage                  `json:"usage"`
	SystemFingerprint string                 `json:"system_fingerprint,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// Choice represents a generated choice
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
	Logprobs     *struct {
		Content []TokenLogprob `json:"content"`
	} `json:"logprobs,omitempty"`
}

// TokenLogprob represents token probability information
type TokenLogprob struct {
	Token   string  `json:"token"`
	Logprob float64 `json:"logprob"`
	Bytes   []int   `json:"bytes,omitempty"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChunk represents a streaming response chunk
type StreamChunk struct {
	ID                string                 `json:"id"`
	Object            string                 `json:"object"`
	Created           int64                  `json:"created"`
	Model             string                 `json:"model"`
	Choices           []StreamChoice         `json:"choices"`
	Usage             *Usage                 `json:"usage,omitempty"`
	SystemFingerprint string                 `json:"system_fingerprint,omitempty"`
	Done              bool                   `json:"done,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// StreamChoice represents a streaming choice
type StreamChoice struct {
	Index        int           `json:"index"`
	Delta        MessageDelta  `json:"delta"`
	FinishReason *string       `json:"finish_reason"`
	Logprobs     *struct {
		Content []TokenLogprob `json:"content"`
	} `json:"logprobs,omitempty"`
}

// MessageDelta represents incremental message content
type MessageDelta struct {
	Role      string     `json:"role,omitempty"`
	Content   string     `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// LLMProvider defines the interface for LLM providers
type LLMProvider interface {
	// GenerateResponse generates a single response
	GenerateResponse(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
	
	// StreamResponse generates a streaming response
	StreamResponse(ctx context.Context, req *GenerateRequest) (<-chan *StreamChunk, error)
	
	// GetModels returns available models
	GetModels(ctx context.Context) ([]string, error)
	
	// GetName returns the provider name
	GetName() string
	
	// Close cleans up resources
	Close() error
}

// LLMConfig represents configuration for LLM providers
type LLMConfig struct {
	Provider    string                 `json:"provider"`
	APIKey      string                 `json:"api_key"`
	BaseURL     string                 `json:"base_url,omitempty"`
	Model       string                 `json:"model"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	TopP        float64                `json:"top_p,omitempty"`
	Timeout     time.Duration          `json:"timeout,omitempty"`
	RetryConfig *RetryConfig           `json:"retry_config,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxRetries      int           `json:"max_retries"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffFactor   float64       `json:"backoff_factor"`
	RetryableErrors []string      `json:"retryable_errors"`
}

// DefaultRetryConfig returns a sensible default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    3,
		InitialDelay:  time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		RetryableErrors: []string{
			"rate_limit_exceeded",
			"server_error",
			"timeout",
			"connection_error",
		},
	}
}

// LLMError represents an error from LLM operations
type LLMError struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Type     string `json:"type"`
	Provider string `json:"provider"`
}

func (e *LLMError) Error() string {
	return e.Message
}

// NewLLMError creates a new LLM error
func NewLLMError(code, message, errorType, provider string) *LLMError {
	return &LLMError{
		Code:     code,
		Message:  message,
		Type:     errorType,
		Provider: provider,
	}
}

// ProviderType represents different LLM provider types
type ProviderType string

const (
	ProviderOpenAI    ProviderType = "openai"
	ProviderAnthropic ProviderType = "anthropic"
	ProviderLocal     ProviderType = "local"
)

// BaseProvider provides common functionality for all providers
type BaseProvider struct {
	config      *LLMConfig
	name        string
	retryConfig *RetryConfig
}

// NewBaseProvider creates a new base provider
func NewBaseProvider(config *LLMConfig, name string) *BaseProvider {
	if config == nil {
		panic("config cannot be nil")
	}
	if name == "" {
		panic("provider name cannot be empty")
	}

	retryConfig := config.RetryConfig
	if retryConfig == nil {
		retryConfig = DefaultRetryConfig()
	}

	return &BaseProvider{
		config:      config,
		name:        name,
		retryConfig: retryConfig,
	}
}

// GetName returns the provider name
func (p *BaseProvider) GetName() string {
	return p.name
}

// GetConfig returns the provider configuration
func (p *BaseProvider) GetConfig() *LLMConfig {
	return p.config
}

// WithRetry executes a function with retry logic
func (p *BaseProvider) WithRetry(ctx context.Context, operation func() error) error {
	if operation == nil {
		return fmt.Errorf("operation cannot be nil")
	}
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}

	var lastErr error
	delay := p.retryConfig.InitialDelay

	for attempt := 0; attempt <= p.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Continue with retry
			}

			// Calculate next delay with exponential backoff
			delay = time.Duration(float64(delay) * p.retryConfig.BackoffFactor)
			if delay > p.retryConfig.MaxDelay {
				delay = p.retryConfig.MaxDelay
			}
		}

		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !p.isRetryableError(err) {
			return err
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", p.retryConfig.MaxRetries, lastErr)
}

// isRetryableError checks if an error should trigger a retry
func (p *BaseProvider) isRetryableError(err error) bool {
	if llmErr, ok := err.(*LLMError); ok {
		for _, retryableCode := range p.retryConfig.RetryableErrors {
			if llmErr.Code == retryableCode {
				return true
			}
		}
	}
	return false
}

// ValidateConfig validates the provider configuration
func (p *BaseProvider) ValidateConfig() error {
	if p.config.APIKey == "" {
		return fmt.Errorf("API key is required for provider %s", p.name)
	}

	if p.config.Model == "" {
		return fmt.Errorf("model is required for provider %s", p.name)
	}

	if p.config.MaxTokens < 0 {
		return fmt.Errorf("max_tokens must be non-negative for provider %s", p.name)
	}

	if p.config.Temperature < 0 || p.config.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2 for provider %s", p.name)
	}

	if p.config.TopP < 0 || p.config.TopP > 1 {
		return fmt.Errorf("top_p must be between 0 and 1 for provider %s", p.name)
	}

	return nil
}

// Close provides default cleanup (can be overridden)
func (p *BaseProvider) Close() error {
	return nil
}

// PrepareRequest prepares a request with default values
func (p *BaseProvider) PrepareRequest(req *GenerateRequest) *GenerateRequest {
	if req == nil {
		return nil
	}

	// Create a copy to avoid modifying the original
	prepared := *req

	// Set defaults from config if not specified in request
	if prepared.Model == "" {
		prepared.Model = p.config.Model
	}

	if prepared.MaxTokens == 0 && p.config.MaxTokens > 0 {
		prepared.MaxTokens = p.config.MaxTokens
	}

	if prepared.Temperature == 0 && p.config.Temperature > 0 {
		prepared.Temperature = p.config.Temperature
	}

	if prepared.TopP == 0 && p.config.TopP > 0 {
		prepared.TopP = p.config.TopP
	}

	// Add system prompt if specified in config and not in request
	if prepared.SystemPrompt == "" && len(prepared.Messages) > 0 {
		if prepared.Messages[0].Role != "system" {
			// Check if we have a system prompt in config metadata
			if p.config.Metadata != nil {
				if systemPrompt, ok := p.config.Metadata["system_prompt"].(string); ok && systemPrompt != "" {
					prepared.SystemPrompt = systemPrompt
				}
			}
		}
	}

	return &prepared
}

// AddSystemMessage adds a system message to the beginning of messages if needed
func (p *BaseProvider) AddSystemMessage(messages []Message, systemPrompt string) []Message {
	if systemPrompt == "" {
		return messages
	}

	// Check if first message is already a system message
	if len(messages) > 0 && messages[0].Role == "system" {
		return messages
	}

	// Prepend system message
	systemMessage := Message{
		Role:    "system",
		Content: systemPrompt,
	}

	return append([]Message{systemMessage}, messages...)
}

package llm

import (
	"context"
	"time"
)

// Message represents a conversation message
type Message struct {
	Role      string                 `json:"role"`      // "system", "user", "assistant", "tool"
	Content   string                 `json:"content"`
	ToolCalls []ToolCall             `json:"tool_calls,omitempty"`
	ToolCallID string                `json:"tool_call_id,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ToolCall represents a function/tool call from the LLM
type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // "function"
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"` // JSON string
	} `json:"function"`
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

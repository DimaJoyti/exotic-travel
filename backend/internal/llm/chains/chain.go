package chains

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Local type definitions to avoid import cycles

// Message represents a message in a conversation
type Message struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	Name      string     `json:"name,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall represents a tool call in a message
type ToolCall struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// Function represents a function call
type Function struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// LLMTool represents a tool that can be called by the LLM
type LLMTool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction represents a tool function definition
type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Choice represents a choice in the response
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// GenerateRequest represents a request to generate text
type GenerateRequest struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	Stop        []string  `json:"stop,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
	Tools       []LLMTool `json:"tools,omitempty"`
}

// GenerateResponse represents a response from text generation
type GenerateResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// LLMProvider interface for LLM providers
type LLMProvider interface {
	GenerateResponse(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
	StreamResponse(ctx context.Context, req *GenerateRequest) (<-chan *StreamChunk, error)
	GetModels(ctx context.Context) ([]string, error)
	GetName() string
	Close() error
}

// StreamChunk represents a chunk in a streaming response
type StreamChunk struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Done    bool     `json:"done"`
}

// ChainInput represents input to a chain
type ChainInput struct {
	Variables map[string]interface{} `json:"variables"`
	Messages  []Message              `json:"messages,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// ChainOutput represents output from a chain
type ChainOutput struct {
	Result   interface{}            `json:"result"`
	Messages []Message              `json:"messages,omitempty"`
	Context  map[string]interface{} `json:"context,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ChainStep represents a single step in a chain
type ChainStep interface {
	Execute(ctx context.Context, input *ChainInput) (*ChainOutput, error)
	GetName() string
	GetDescription() string
}

// Chain represents a sequence of steps that can be executed
type Chain interface {
	Execute(ctx context.Context, input *ChainInput) (*ChainOutput, error)
	AddStep(step ChainStep) Chain
	SetMemory(memory Memory) Chain
	GetSteps() []ChainStep
	GetName() string
}

// Memory interface for storing and retrieving conversation context
type Memory interface {
	Store(ctx context.Context, key string, value interface{}) error
	Retrieve(ctx context.Context, key string) (interface{}, error)
	Clear(ctx context.Context, key string) error
	GetHistory(ctx context.Context, key string, limit int) ([]Message, error)
	AddMessage(ctx context.Context, key string, message Message) error
}

// BaseChain provides common functionality for all chains
type BaseChain struct {
	name        string
	description string
	steps       []ChainStep
	memory      Memory
	tracer      trace.Tracer
}

// NewBaseChain creates a new base chain
func NewBaseChain(name, description string) *BaseChain {
	return &BaseChain{
		name:        name,
		description: description,
		steps:       make([]ChainStep, 0),
		tracer:      otel.Tracer("llm.chains"),
	}
}

// GetName returns the chain name
func (c *BaseChain) GetName() string {
	return c.name
}

// GetDescription returns the chain description
func (c *BaseChain) GetDescription() string {
	return c.description
}

// GetSteps returns the chain steps
func (c *BaseChain) GetSteps() []ChainStep {
	return c.steps
}

// AddStep adds a step to the chain
func (c *BaseChain) AddStep(step ChainStep) Chain {
	c.steps = append(c.steps, step)
	return c
}

// SetMemory sets the memory for the chain
func (c *BaseChain) SetMemory(memory Memory) Chain {
	c.memory = memory
	return c
}

// Execute executes the chain (to be implemented by specific chain types)
func (c *BaseChain) Execute(ctx context.Context, input *ChainInput) (*ChainOutput, error) {
	return nil, fmt.Errorf("execute method must be implemented by specific chain type")
}

// LLMStep represents a step that calls an LLM
type LLMStep struct {
	name        string
	description string
	provider    LLMProvider
	prompt      string
	config      *GenerateRequest
	tracer      trace.Tracer
}

// NewLLMStep creates a new LLM step
func NewLLMStep(name, description string, provider LLMProvider, prompt string) *LLMStep {
	return &LLMStep{
		name:        name,
		description: description,
		provider:    provider,
		prompt:      prompt,
		config:      &GenerateRequest{},
		tracer:      otel.Tracer("llm.chains.llm_step"),
	}
}

// WithConfig sets the LLM configuration for the step
func (s *LLMStep) WithConfig(config *GenerateRequest) *LLMStep {
	s.config = config
	return s
}

// GetName returns the step name
func (s *LLMStep) GetName() string {
	return s.name
}

// GetDescription returns the step description
func (s *LLMStep) GetDescription() string {
	return s.description
}

// Execute executes the LLM step
func (s *LLMStep) Execute(ctx context.Context, input *ChainInput) (*ChainOutput, error) {
	ctx, span := s.tracer.Start(ctx, "llm_step.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("step.name", s.name),
		attribute.String("step.type", "llm"),
		attribute.String("llm.provider", s.provider.GetName()),
	)

	// Prepare the prompt by substituting variables
	prompt, err := s.substituteVariables(s.prompt, input.Variables)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to substitute variables in prompt: %w", err)
	}

	// Prepare messages
	messages := input.Messages
	if len(messages) == 0 {
		messages = []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		}
	} else {
		// Add the prompt as the latest user message
		messages = append(messages, Message{
			Role:    "user",
			Content: prompt,
		})
	}

	// Prepare the request
	req := &GenerateRequest{
		Messages:    messages,
		Model:       s.config.Model,
		MaxTokens:   s.config.MaxTokens,
		Temperature: s.config.Temperature,
		TopP:        s.config.TopP,
		Tools:       s.config.Tools,
	}

	// Call the LLM
	response, err := s.provider.GenerateResponse(ctx, req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// Extract the result
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from LLM")
	}

	choice := response.Choices[0]

	// Prepare output
	output := &ChainOutput{
		Result:   choice.Message.Content,
		Messages: append(messages, choice.Message),
		Context:  input.Context,
		Metadata: map[string]interface{}{
			"usage":         response.Usage,
			"finish_reason": choice.FinishReason,
			"model":         response.Model,
		},
	}

	// Handle tool calls if present
	if len(choice.Message.ToolCalls) > 0 {
		output.Metadata["tool_calls"] = choice.Message.ToolCalls
	}

	return output, nil
}

// substituteVariables replaces variables in the prompt template
func (s *LLMStep) substituteVariables(prompt string, variables map[string]interface{}) (string, error) {
	// Simple variable substitution - in a real implementation, you might use
	// a more sophisticated template engine like text/template
	result := prompt

	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		valueStr := fmt.Sprintf("%v", value)
		result = fmt.Sprintf(result, placeholder, valueStr)
	}

	return result, nil
}

// ToolStep represents a step that calls an external tool
type ToolStep struct {
	name        string
	description string
	tool        Tool
	tracer      trace.Tracer
}

// Tool interface for external tools
type Tool interface {
	Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)
	GetName() string
	GetDescription() string
	GetSchema() map[string]interface{}
}

// NewToolStep creates a new tool step
func NewToolStep(name, description string, tool Tool) *ToolStep {
	return &ToolStep{
		name:        name,
		description: description,
		tool:        tool,
		tracer:      otel.Tracer("llm.chains.tool_step"),
	}
}

// GetName returns the step name
func (s *ToolStep) GetName() string {
	return s.name
}

// GetDescription returns the step description
func (s *ToolStep) GetDescription() string {
	return s.description
}

// Execute executes the tool step
func (s *ToolStep) Execute(ctx context.Context, input *ChainInput) (*ChainOutput, error) {
	ctx, span := s.tracer.Start(ctx, "tool_step.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("step.name", s.name),
		attribute.String("step.type", "tool"),
		attribute.String("tool.name", s.tool.GetName()),
	)

	// Execute the tool
	result, err := s.tool.Execute(ctx, input.Variables)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	// Prepare output
	output := &ChainOutput{
		Result:  result,
		Context: input.Context,
		Metadata: map[string]interface{}{
			"tool_name":   s.tool.GetName(),
			"tool_result": result,
		},
	}

	return output, nil
}

// TransformStep represents a step that transforms data
type TransformStep struct {
	name        string
	description string
	transformer func(ctx context.Context, input *ChainInput) (*ChainOutput, error)
	tracer      trace.Tracer
}

// NewTransformStep creates a new transform step
func NewTransformStep(name, description string, transformer func(ctx context.Context, input *ChainInput) (*ChainOutput, error)) *TransformStep {
	return &TransformStep{
		name:        name,
		description: description,
		transformer: transformer,
		tracer:      otel.Tracer("llm.chains.transform_step"),
	}
}

// GetName returns the step name
func (s *TransformStep) GetName() string {
	return s.name
}

// GetDescription returns the step description
func (s *TransformStep) GetDescription() string {
	return s.description
}

// Execute executes the transform step
func (s *TransformStep) Execute(ctx context.Context, input *ChainInput) (*ChainOutput, error) {
	ctx, span := s.tracer.Start(ctx, "transform_step.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("step.name", s.name),
		attribute.String("step.type", "transform"),
	)

	return s.transformer(ctx, input)
}

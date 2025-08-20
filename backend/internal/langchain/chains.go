package langchain

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Chain represents a sequence of operations that can be executed
type Chain interface {
	// Execute runs the chain with the given input
	Execute(ctx context.Context, input map[string]interface{}) (*ChainResult, error)

	// GetName returns the chain's name
	GetName() string

	// GetDescription returns the chain's description
	GetDescription() string

	// Validate validates the chain configuration
	Validate() error
}

// ChainResult represents the result of a chain execution
type ChainResult struct {
	Output    map[string]interface{} `json:"output"`
	Metadata  map[string]interface{} `json:"metadata"`
	Duration  time.Duration          `json:"duration"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error,omitempty"`
	ChainName string                 `json:"chain_name"`
}

// BaseChain provides common functionality for all chains
type BaseChain struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	tracer      trace.Tracer           `json:"-"`
}

// NewBaseChain creates a new base chain
func NewBaseChain(name, description string) *BaseChain {
	return &BaseChain{
		Name:        name,
		Description: description,
		Config:      make(map[string]interface{}),
		tracer:      otel.Tracer("langchain.chain"),
	}
}

// GetName returns the chain name
func (c *BaseChain) GetName() string {
	return c.Name
}

// GetDescription returns the chain description
func (c *BaseChain) GetDescription() string {
	return c.Description
}

// Validate validates the base chain
func (c *BaseChain) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("chain name cannot be empty")
	}
	return nil
}

// SequentialChain executes multiple chains in sequence
type SequentialChain struct {
	*BaseChain
	Chains []Chain `json:"chains"`
}

// NewSequentialChain creates a new sequential chain
func NewSequentialChain(name, description string, chains ...Chain) *SequentialChain {
	return &SequentialChain{
		BaseChain: NewBaseChain(name, description),
		Chains:    chains,
	}
}

// Execute executes all chains in sequence
func (c *SequentialChain) Execute(ctx context.Context, input map[string]interface{}) (*ChainResult, error) {
	ctx, span := c.tracer.Start(ctx, "sequential_chain.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("chain.name", c.Name),
		attribute.Int("chain.count", len(c.Chains)),
	)

	startTime := time.Now()
	currentInput := input
	allOutputs := make(map[string]interface{})
	metadata := make(map[string]interface{})

	for i, chain := range c.Chains {
		chainSpan := trace.SpanFromContext(ctx)
		chainSpan.SetAttributes(
			attribute.String("current_chain", chain.GetName()),
			attribute.Int("chain_index", i),
		)

		result, err := chain.Execute(ctx, currentInput)
		if err != nil {
			span.RecordError(err)
			return &ChainResult{
				Output:    allOutputs,
				Metadata:  metadata,
				Duration:  time.Since(startTime),
				Success:   false,
				Error:     fmt.Sprintf("chain %d (%s) failed: %v", i, chain.GetName(), err),
				ChainName: c.Name,
			}, err
		}

		// Merge outputs for next chain
		for key, value := range result.Output {
			currentInput[key] = value
			allOutputs[key] = value
		}

		// Collect metadata
		metadata[fmt.Sprintf("chain_%d_metadata", i)] = result.Metadata
		metadata[fmt.Sprintf("chain_%d_duration", i)] = result.Duration
	}

	return &ChainResult{
		Output:    allOutputs,
		Metadata:  metadata,
		Duration:  time.Since(startTime),
		Success:   true,
		ChainName: c.Name,
	}, nil
}

// Validate validates the sequential chain
func (c *SequentialChain) Validate() error {
	if err := c.BaseChain.Validate(); err != nil {
		return err
	}

	if len(c.Chains) == 0 {
		return fmt.Errorf("sequential chain must have at least one chain")
	}

	for i, chain := range c.Chains {
		if err := chain.Validate(); err != nil {
			return fmt.Errorf("chain %d validation failed: %w", i, err)
		}
	}

	return nil
}

// ParallelChain executes multiple chains in parallel
type ParallelChain struct {
	*BaseChain
	Chains []Chain `json:"chains"`
}

// NewParallelChain creates a new parallel chain
func NewParallelChain(name, description string, chains ...Chain) *ParallelChain {
	return &ParallelChain{
		BaseChain: NewBaseChain(name, description),
		Chains:    chains,
	}
}

// Execute executes all chains in parallel
func (c *ParallelChain) Execute(ctx context.Context, input map[string]interface{}) (*ChainResult, error) {
	ctx, span := c.tracer.Start(ctx, "parallel_chain.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("chain.name", c.Name),
		attribute.Int("chain.count", len(c.Chains)),
	)

	startTime := time.Now()

	// Create channels for results
	type chainResult struct {
		index  int
		result *ChainResult
		err    error
	}

	resultChan := make(chan chainResult, len(c.Chains))

	// Execute chains in parallel
	for i, chain := range c.Chains {
		go func(index int, ch Chain) {
			result, err := ch.Execute(ctx, input)
			resultChan <- chainResult{
				index:  index,
				result: result,
				err:    err,
			}
		}(i, chain)
	}

	// Collect results
	allOutputs := make(map[string]interface{})
	metadata := make(map[string]interface{})
	var errors []string

	for i := 0; i < len(c.Chains); i++ {
		select {
		case result := <-resultChan:
			if result.err != nil {
				errors = append(errors, fmt.Sprintf("chain %d (%s) failed: %v",
					result.index, c.Chains[result.index].GetName(), result.err))
				span.RecordError(result.err)
			} else {
				// Merge outputs with chain prefix
				chainName := c.Chains[result.index].GetName()
				for key, value := range result.result.Output {
					prefixedKey := fmt.Sprintf("%s_%s", chainName, key)
					allOutputs[prefixedKey] = value
				}

				// Collect metadata
				metadata[fmt.Sprintf("chain_%d_metadata", result.index)] = result.result.Metadata
				metadata[fmt.Sprintf("chain_%d_duration", result.index)] = result.result.Duration
			}
		case <-ctx.Done():
			return &ChainResult{
				Output:    allOutputs,
				Metadata:  metadata,
				Duration:  time.Since(startTime),
				Success:   false,
				Error:     "execution cancelled",
				ChainName: c.Name,
			}, ctx.Err()
		}
	}

	success := len(errors) == 0
	var errorMsg string
	if !success {
		errorMsg = fmt.Sprintf("parallel execution had %d errors: %v", len(errors), errors)
	}

	return &ChainResult{
		Output:    allOutputs,
		Metadata:  metadata,
		Duration:  time.Since(startTime),
		Success:   success,
		Error:     errorMsg,
		ChainName: c.Name,
	}, nil
}

// Validate validates the parallel chain
func (c *ParallelChain) Validate() error {
	if err := c.BaseChain.Validate(); err != nil {
		return err
	}

	if len(c.Chains) == 0 {
		return fmt.Errorf("parallel chain must have at least one chain")
	}

	for i, chain := range c.Chains {
		if err := chain.Validate(); err != nil {
			return fmt.Errorf("chain %d validation failed: %w", i, err)
		}
	}

	return nil
}

// ConditionalChain executes chains based on conditions
type ConditionalChain struct {
	*BaseChain
	Condition    func(ctx context.Context, input map[string]interface{}) (bool, error) `json:"-"`
	TrueChain    Chain                                                                 `json:"true_chain"`
	FalseChain   Chain                                                                 `json:"false_chain"`
	DefaultChain Chain                                                                 `json:"default_chain"`
}

// NewConditionalChain creates a new conditional chain
func NewConditionalChain(name, description string, condition func(ctx context.Context, input map[string]interface{}) (bool, error), trueChain, falseChain Chain) *ConditionalChain {
	return &ConditionalChain{
		BaseChain:  NewBaseChain(name, description),
		Condition:  condition,
		TrueChain:  trueChain,
		FalseChain: falseChain,
	}
}

// Execute executes the appropriate chain based on the condition
func (c *ConditionalChain) Execute(ctx context.Context, input map[string]interface{}) (*ChainResult, error) {
	ctx, span := c.tracer.Start(ctx, "conditional_chain.execute")
	defer span.End()

	span.SetAttributes(attribute.String("chain.name", c.Name))

	startTime := time.Now()

	// Evaluate condition
	conditionResult, err := c.Condition(ctx, input)
	if err != nil {
		span.RecordError(err)
		return &ChainResult{
			Output:    make(map[string]interface{}),
			Metadata:  map[string]interface{}{"condition_error": err.Error()},
			Duration:  time.Since(startTime),
			Success:   false,
			Error:     fmt.Sprintf("condition evaluation failed: %v", err),
			ChainName: c.Name,
		}, err
	}

	span.SetAttributes(attribute.Bool("condition.result", conditionResult))

	// Execute appropriate chain
	var targetChain Chain
	var chainType string

	if conditionResult && c.TrueChain != nil {
		targetChain = c.TrueChain
		chainType = "true"
	} else if !conditionResult && c.FalseChain != nil {
		targetChain = c.FalseChain
		chainType = "false"
	} else if c.DefaultChain != nil {
		targetChain = c.DefaultChain
		chainType = "default"
	} else {
		return &ChainResult{
			Output:    make(map[string]interface{}),
			Metadata:  map[string]interface{}{"condition_result": conditionResult},
			Duration:  time.Since(startTime),
			Success:   true,
			ChainName: c.Name,
		}, nil
	}

	span.SetAttributes(attribute.String("executed_chain", chainType))

	result, err := targetChain.Execute(ctx, input)
	if err != nil {
		span.RecordError(err)
		return &ChainResult{
			Output:    make(map[string]interface{}),
			Metadata:  map[string]interface{}{"condition_result": conditionResult, "chain_type": chainType},
			Duration:  time.Since(startTime),
			Success:   false,
			Error:     fmt.Sprintf("%s chain execution failed: %v", chainType, err),
			ChainName: c.Name,
		}, err
	}

	// Add condition metadata
	result.Metadata["condition_result"] = conditionResult
	result.Metadata["executed_chain"] = chainType
	result.ChainName = c.Name

	return result, nil
}

// Validate validates the conditional chain
func (c *ConditionalChain) Validate() error {
	if err := c.BaseChain.Validate(); err != nil {
		return err
	}

	if c.Condition == nil {
		return fmt.Errorf("condition function cannot be nil")
	}

	if c.TrueChain == nil && c.FalseChain == nil && c.DefaultChain == nil {
		return fmt.Errorf("at least one chain (true, false, or default) must be provided")
	}

	if c.TrueChain != nil {
		if err := c.TrueChain.Validate(); err != nil {
			return fmt.Errorf("true chain validation failed: %w", err)
		}
	}

	if c.FalseChain != nil {
		if err := c.FalseChain.Validate(); err != nil {
			return fmt.Errorf("false chain validation failed: %w", err)
		}
	}

	if c.DefaultChain != nil {
		if err := c.DefaultChain.Validate(); err != nil {
			return fmt.Errorf("default chain validation failed: %w", err)
		}
	}

	return nil
}

// ChainBuilder provides a fluent interface for building chains
type ChainBuilder struct {
	chains []Chain
}

// NewChainBuilder creates a new chain builder
func NewChainBuilder() *ChainBuilder {
	return &ChainBuilder{
		chains: make([]Chain, 0),
	}
}

// Add adds a chain to the builder
func (b *ChainBuilder) Add(chain Chain) *ChainBuilder {
	b.chains = append(b.chains, chain)
	return b
}

// BuildSequential builds a sequential chain
func (b *ChainBuilder) BuildSequential(name, description string) *SequentialChain {
	return NewSequentialChain(name, description, b.chains...)
}

// BuildParallel builds a parallel chain
func (b *ChainBuilder) BuildParallel(name, description string) *ParallelChain {
	return NewParallelChain(name, description, b.chains...)
}

// LLMChain represents a chain that calls an LLM
type LLMChain struct {
	*BaseChain
	LLMProvider    interface{}     `json:"-"` // LLM provider interface
	PromptTemplate *PromptTemplate `json:"prompt_template"`
	OutputParser   OutputParser    `json:"output_parser"`
	Memory         Memory          `json:"memory"`
	SessionID      string          `json:"session_id"`
}

// NewLLMChain creates a new LLM chain
func NewLLMChain(name, description string, llmProvider interface{}, promptTemplate *PromptTemplate) *LLMChain {
	return &LLMChain{
		BaseChain:      NewBaseChain(name, description),
		LLMProvider:    llmProvider,
		PromptTemplate: promptTemplate,
	}
}

// SetOutputParser sets the output parser for the chain
func (c *LLMChain) SetOutputParser(parser OutputParser) *LLMChain {
	c.OutputParser = parser
	return c
}

// SetMemory sets the memory for the chain
func (c *LLMChain) SetMemory(memory Memory, sessionID string) *LLMChain {
	c.Memory = memory
	c.SessionID = sessionID
	return c
}

// Execute executes the LLM chain
func (c *LLMChain) Execute(ctx context.Context, input map[string]interface{}) (*ChainResult, error) {
	ctx, span := c.tracer.Start(ctx, "llm_chain.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("chain.name", c.Name),
		attribute.String("chain.type", "llm"),
	)

	startTime := time.Now()

	// Add conversation history if memory is available
	if c.Memory != nil && c.SessionID != "" {
		messages, err := c.Memory.GetMessages(ctx, c.SessionID, 10) // Get last 10 messages
		if err == nil && len(messages) > 0 {
			// Convert messages to conversation history
			history := ""
			for _, msg := range messages {
				history += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
			}
			input["conversation_history"] = history
		}
	}

	// Render prompt
	prompt, err := c.PromptTemplate.Render(ctx, input)
	if err != nil {
		span.RecordError(err)
		return &ChainResult{
			Output:    make(map[string]interface{}),
			Metadata:  map[string]interface{}{"error": err.Error()},
			Duration:  time.Since(startTime),
			Success:   false,
			Error:     fmt.Sprintf("prompt rendering failed: %v", err),
			ChainName: c.Name,
		}, err
	}

	span.SetAttributes(attribute.String("prompt.rendered", prompt))

	// TODO: Call actual LLM provider here
	// For now, simulate LLM response
	llmResponse := fmt.Sprintf("LLM response to: %s", prompt)

	// Store user message in memory
	if c.Memory != nil && c.SessionID != "" {
		userMessage := &Message{
			SessionID: c.SessionID,
			Role:      "user",
			Content:   fmt.Sprintf("%v", input["query"]),
			Metadata:  map[string]interface{}{"chain": c.Name},
			Timestamp: time.Now(),
		}
		c.Memory.AddMessage(ctx, userMessage)
	}

	// Parse output if parser is available
	var parsedOutput interface{} = llmResponse
	if c.OutputParser != nil {
		parsed, err := c.OutputParser.Parse(ctx, llmResponse)
		if err != nil {
			span.RecordError(err)
			// Continue with raw output if parsing fails
		} else {
			parsedOutput = parsed
		}
	}

	// Store assistant response in memory
	if c.Memory != nil && c.SessionID != "" {
		assistantMessage := &Message{
			SessionID: c.SessionID,
			Role:      "assistant",
			Content:   llmResponse,
			Metadata:  map[string]interface{}{"chain": c.Name, "parsed": parsedOutput},
			Timestamp: time.Now(),
		}
		c.Memory.AddMessage(ctx, assistantMessage)
	}

	result := &ChainResult{
		Output: map[string]interface{}{
			"response":      llmResponse,
			"parsed_output": parsedOutput,
			"prompt_used":   prompt,
		},
		Metadata: map[string]interface{}{
			"llm_provider":    "simulated", // TODO: Get actual provider name
			"prompt_template": c.PromptTemplate.Name,
			"session_id":      c.SessionID,
		},
		Duration:  time.Since(startTime),
		Success:   true,
		ChainName: c.Name,
	}

	span.SetAttributes(
		attribute.String("response.length", fmt.Sprintf("%d", len(llmResponse))),
		attribute.Bool("output.parsed", c.OutputParser != nil),
	)

	return result, nil
}

// Validate validates the LLM chain
func (c *LLMChain) Validate() error {
	if err := c.BaseChain.Validate(); err != nil {
		return err
	}

	if c.LLMProvider == nil {
		return fmt.Errorf("LLM provider cannot be nil")
	}

	if c.PromptTemplate == nil {
		return fmt.Errorf("prompt template cannot be nil")
	}

	if err := c.PromptTemplate.Validate(); err != nil {
		return fmt.Errorf("prompt template validation failed: %w", err)
	}

	if c.OutputParser != nil {
		if err := c.OutputParser.Validate(); err != nil {
			return fmt.Errorf("output parser validation failed: %w", err)
		}
	}

	return nil
}

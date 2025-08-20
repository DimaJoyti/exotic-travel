package langgraph

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Node represents a node in the graph
type Node interface {
	// GetID returns the unique identifier of the node
	GetID() string
	
	// GetName returns the human-readable name of the node
	GetName() string
	
	// GetType returns the type of the node
	GetType() string
	
	// Execute executes the node with the given state and returns the updated state
	Execute(ctx context.Context, state *State) (*State, error)
	
	// Validate validates the node configuration
	Validate() error
}

// BaseNode provides common functionality for all nodes
type BaseNode struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	tracer      trace.Tracer           `json:"-"`
}

// NewBaseNode creates a new base node
func NewBaseNode(id, name, nodeType string) *BaseNode {
	return &BaseNode{
		ID:     id,
		Name:   name,
		Type:   nodeType,
		Config: make(map[string]interface{}),
		tracer: otel.Tracer("langgraph.node"),
	}
}

// GetID returns the node ID
func (n *BaseNode) GetID() string {
	return n.ID
}

// GetName returns the node name
func (n *BaseNode) GetName() string {
	return n.Name
}

// GetType returns the node type
func (n *BaseNode) GetType() string {
	return n.Type
}

// Validate validates the base node
func (n *BaseNode) Validate() error {
	if n.ID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}
	if n.Name == "" {
		return fmt.Errorf("node name cannot be empty")
	}
	if n.Type == "" {
		return fmt.Errorf("node type cannot be empty")
	}
	return nil
}

// LLMNode represents a node that calls an LLM
type LLMNode struct {
	*BaseNode
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	PromptTemplate string `json:"prompt_template"`
	OutputKey    string `json:"output_key"`
	LLMProvider  interface{} `json:"-"` // Will be set to actual LLM provider
}

// NewLLMNode creates a new LLM node
func NewLLMNode(id, name, provider, model, promptTemplate, outputKey string) *LLMNode {
	return &LLMNode{
		BaseNode:       NewBaseNode(id, name, "llm"),
		Provider:       provider,
		Model:          model,
		PromptTemplate: promptTemplate,
		OutputKey:      outputKey,
	}
}

// Execute executes the LLM node
func (n *LLMNode) Execute(ctx context.Context, state *State) (*State, error) {
	ctx, span := n.tracer.Start(ctx, "llm_node.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("node.id", n.ID),
		attribute.String("node.type", n.Type),
		attribute.String("llm.provider", n.Provider),
		attribute.String("llm.model", n.Model),
	)

	// Create a new state based on the input state
	newState := state.Clone()

	// Render prompt template with state data
	prompt, err := n.renderPrompt(state)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to render prompt: %w", err)
	}

	span.SetAttributes(attribute.String("llm.prompt", prompt))

	// TODO: Call actual LLM provider here
	// For now, simulate LLM response
	response := fmt.Sprintf("LLM response for prompt: %s", prompt)
	
	// Store response in state
	newState.Set(n.OutputKey, response)
	newState.SetMetadata("last_llm_call", map[string]interface{}{
		"node_id":  n.ID,
		"provider": n.Provider,
		"model":    n.Model,
		"prompt":   prompt,
		"response": response,
		"timestamp": time.Now(),
	})

	span.SetAttributes(attribute.String("llm.response", response))
	return newState, nil
}

// renderPrompt renders the prompt template with state data
func (n *LLMNode) renderPrompt(state *State) (string, error) {
	// Simple template rendering - replace {{key}} with state values
	prompt := n.PromptTemplate
	
	// For now, just return the template as-is
	// TODO: Implement proper template rendering with state substitution
	return prompt, nil
}

// Validate validates the LLM node
func (n *LLMNode) Validate() error {
	if err := n.BaseNode.Validate(); err != nil {
		return err
	}
	if n.Provider == "" {
		return fmt.Errorf("LLM provider cannot be empty")
	}
	if n.Model == "" {
		return fmt.Errorf("LLM model cannot be empty")
	}
	if n.PromptTemplate == "" {
		return fmt.Errorf("prompt template cannot be empty")
	}
	if n.OutputKey == "" {
		return fmt.Errorf("output key cannot be empty")
	}
	return nil
}

// ToolNode represents a node that calls a tool
type ToolNode struct {
	*BaseNode
	ToolName   string                 `json:"tool_name"`
	InputKeys  []string               `json:"input_keys"`
	OutputKey  string                 `json:"output_key"`
	ToolConfig map[string]interface{} `json:"tool_config"`
	Tool       interface{}            `json:"-"` // Will be set to actual tool
}

// NewToolNode creates a new tool node
func NewToolNode(id, name, toolName string, inputKeys []string, outputKey string) *ToolNode {
	return &ToolNode{
		BaseNode:   NewBaseNode(id, name, "tool"),
		ToolName:   toolName,
		InputKeys:  inputKeys,
		OutputKey:  outputKey,
		ToolConfig: make(map[string]interface{}),
	}
}

// Execute executes the tool node
func (n *ToolNode) Execute(ctx context.Context, state *State) (*State, error) {
	ctx, span := n.tracer.Start(ctx, "tool_node.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("node.id", n.ID),
		attribute.String("node.type", n.Type),
		attribute.String("tool.name", n.ToolName),
	)

	// Create a new state based on the input state
	newState := state.Clone()

	// Gather input values from state
	inputs := make(map[string]interface{})
	for _, key := range n.InputKeys {
		if value, exists := state.Get(key); exists {
			inputs[key] = value
		}
	}

	span.SetAttributes(attribute.Int("tool.inputs_count", len(inputs)))

	// TODO: Call actual tool here
	// For now, simulate tool execution
	result := fmt.Sprintf("Tool %s executed with inputs: %v", n.ToolName, inputs)
	
	// Store result in state
	newState.Set(n.OutputKey, result)
	newState.SetMetadata("last_tool_call", map[string]interface{}{
		"node_id":   n.ID,
		"tool_name": n.ToolName,
		"inputs":    inputs,
		"result":    result,
		"timestamp": time.Now(),
	})

	span.SetAttributes(attribute.String("tool.result", result))
	return newState, nil
}

// Validate validates the tool node
func (n *ToolNode) Validate() error {
	if err := n.BaseNode.Validate(); err != nil {
		return err
	}
	if n.ToolName == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	if n.OutputKey == "" {
		return fmt.Errorf("output key cannot be empty")
	}
	return nil
}

// FunctionNode represents a node that executes a custom function
type FunctionNode struct {
	*BaseNode
	Function  func(ctx context.Context, state *State) (*State, error) `json:"-"`
	InputKeys []string `json:"input_keys"`
	OutputKey string   `json:"output_key"`
}

// NewFunctionNode creates a new function node
func NewFunctionNode(id, name string, fn func(ctx context.Context, state *State) (*State, error)) *FunctionNode {
	return &FunctionNode{
		BaseNode: NewBaseNode(id, name, "function"),
		Function: fn,
	}
}

// Execute executes the function node
func (n *FunctionNode) Execute(ctx context.Context, state *State) (*State, error) {
	ctx, span := n.tracer.Start(ctx, "function_node.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("node.id", n.ID),
		attribute.String("node.type", n.Type),
	)

	if n.Function == nil {
		err := fmt.Errorf("function is nil")
		span.RecordError(err)
		return nil, err
	}

	// Execute the function
	newState, err := n.Function(ctx, state)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("function execution failed: %w", err)
	}

	// Add metadata about function execution
	newState.SetMetadata("last_function_call", map[string]interface{}{
		"node_id":   n.ID,
		"timestamp": time.Now(),
	})

	return newState, nil
}

// Validate validates the function node
func (n *FunctionNode) Validate() error {
	if err := n.BaseNode.Validate(); err != nil {
		return err
	}
	if n.Function == nil {
		return fmt.Errorf("function cannot be nil")
	}
	return nil
}

// ConditionalNode represents a node that performs conditional logic
type ConditionalNode struct {
	*BaseNode
	Condition func(ctx context.Context, state *State) (bool, error) `json:"-"`
	TrueKey   string `json:"true_key"`
	FalseKey  string `json:"false_key"`
}

// NewConditionalNode creates a new conditional node
func NewConditionalNode(id, name string, condition func(ctx context.Context, state *State) (bool, error)) *ConditionalNode {
	return &ConditionalNode{
		BaseNode:  NewBaseNode(id, name, "conditional"),
		Condition: condition,
		TrueKey:   "condition_result_true",
		FalseKey:  "condition_result_false",
	}
}

// Execute executes the conditional node
func (n *ConditionalNode) Execute(ctx context.Context, state *State) (*State, error) {
	ctx, span := n.tracer.Start(ctx, "conditional_node.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("node.id", n.ID),
		attribute.String("node.type", n.Type),
	)

	if n.Condition == nil {
		err := fmt.Errorf("condition function is nil")
		span.RecordError(err)
		return nil, err
	}

	// Create a new state based on the input state
	newState := state.Clone()

	// Evaluate condition
	result, err := n.Condition(ctx, state)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("condition evaluation failed: %w", err)
	}

	span.SetAttributes(attribute.Bool("condition.result", result))

	// Store condition result in state
	if result {
		newState.Set(n.TrueKey, true)
		newState.Delete(n.FalseKey)
	} else {
		newState.Set(n.FalseKey, true)
		newState.Delete(n.TrueKey)
	}

	// Add metadata about condition evaluation
	newState.SetMetadata("last_condition_eval", map[string]interface{}{
		"node_id":   n.ID,
		"result":    result,
		"timestamp": time.Now(),
	})

	return newState, nil
}

// Validate validates the conditional node
func (n *ConditionalNode) Validate() error {
	if err := n.BaseNode.Validate(); err != nil {
		return err
	}
	if n.Condition == nil {
		return fmt.Errorf("condition function cannot be nil")
	}
	return nil
}

// StartNode represents the entry point of a graph
type StartNode struct {
	*BaseNode
	InitialData map[string]interface{} `json:"initial_data"`
}

// NewStartNode creates a new start node
func NewStartNode(id, name string) *StartNode {
	return &StartNode{
		BaseNode:    NewBaseNode(id, name, "start"),
		InitialData: make(map[string]interface{}),
	}
}

// Execute executes the start node
func (n *StartNode) Execute(ctx context.Context, state *State) (*State, error) {
	ctx, span := n.tracer.Start(ctx, "start_node.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("node.id", n.ID),
		attribute.String("node.type", n.Type),
	)

	// Create a new state and add initial data
	newState := state.Clone()
	
	for key, value := range n.InitialData {
		newState.Set(key, value)
	}

	newState.SetMetadata("execution_started", time.Now())
	
	return newState, nil
}

// EndNode represents an exit point of a graph
type EndNode struct {
	*BaseNode
	FinalizeFunction func(ctx context.Context, state *State) error `json:"-"`
}

// NewEndNode creates a new end node
func NewEndNode(id, name string) *EndNode {
	return &EndNode{
		BaseNode: NewBaseNode(id, name, "end"),
	}
}

// Execute executes the end node
func (n *EndNode) Execute(ctx context.Context, state *State) (*State, error) {
	ctx, span := n.tracer.Start(ctx, "end_node.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("node.id", n.ID),
		attribute.String("node.type", n.Type),
	)

	// Create a new state
	newState := state.Clone()

	// Execute finalize function if provided
	if n.FinalizeFunction != nil {
		if err := n.FinalizeFunction(ctx, newState); err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("finalize function failed: %w", err)
		}
	}

	newState.SetMetadata("execution_completed", time.Now())
	
	return newState, nil
}

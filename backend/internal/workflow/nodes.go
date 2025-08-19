package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Local type definitions to avoid import cycles (should match types.go definitions)

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

// Choice represents a choice in an LLM response
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// GenerateRequest represents a request to generate text
type GenerateRequest struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
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
	Close() error
}

// LLMNode represents a node that calls an LLM
type LLMNode struct {
	*BaseNode
	Provider    string       `json:"provider"`
	Model       string       `json:"model,omitempty"`
	Prompt      string       `json:"prompt"`
	MaxTokens   int          `json:"max_tokens,omitempty"`
	Temperature float64      `json:"temperature,omitempty"`
	Tools       []LLMTool    `json:"tools,omitempty"`
	llmProvider LLMProvider  `json:"-"`
	tracer      trace.Tracer `json:"-"`
}

// NewLLMNode creates a new LLM node
func NewLLMNode(id, name string, provider LLMProvider, prompt string) *LLMNode {
	return &LLMNode{
		BaseNode: &BaseNode{
			ID:          id,
			Type:        NodeTypeLLM,
			Name:        name,
			Description: "LLM node that generates responses using language models",
			Config:      make(map[string]interface{}),
			Metadata:    make(map[string]interface{}),
		},
		Prompt:      prompt,
		llmProvider: provider,
		tracer:      otel.Tracer("workflow.nodes.llm"),
	}
}

// Execute executes the LLM node
func (n *LLMNode) Execute(ctx context.Context, state *WorkflowState) (*NodeOutput, error) {
	ctx, span := n.tracer.Start(ctx, "llm_node.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("node.id", n.ID),
		attribute.String("node.type", string(n.Type)),
		attribute.String("llm.provider", n.Provider),
	)

	// Substitute variables in prompt
	prompt, err := n.substituteVariables(n.Prompt, state.Data)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to substitute variables: %w", err)
	}

	// Prepare messages
	messages := state.Messages
	if len(messages) == 0 || messages[len(messages)-1].Role != "user" {
		messages = append(messages, Message{
			Role:    "user",
			Content: prompt,
		})
	}

	// Create LLM request
	req := &GenerateRequest{
		Messages:    messages,
		Model:       n.Model,
		MaxTokens:   n.MaxTokens,
		Temperature: n.Temperature,
		Tools:       n.Tools,
	}

	// Call LLM
	response, err := n.llmProvider.GenerateResponse(ctx, req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from LLM")
	}

	choice := response.Choices[0]

	// Prepare output
	output := &NodeOutput{
		Data: map[string]interface{}{
			"llm_response":  choice.Message.Content,
			"finish_reason": choice.FinishReason,
		},
		Messages: append(messages, choice.Message),
		Metadata: map[string]interface{}{
			"usage":    response.Usage,
			"model":    response.Model,
			"provider": n.Provider,
		},
	}

	// Handle tool calls
	if len(choice.Message.ToolCalls) > 0 {
		output.Data["tool_calls"] = choice.Message.ToolCalls
		output.Metadata["has_tool_calls"] = true
	}

	return output, nil
}

// substituteVariables replaces variables in the prompt
func (n *LLMNode) substituteVariables(prompt string, data map[string]interface{}) (string, error) {
	// Simple variable substitution - in production, use a proper template engine
	result := prompt
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		valueStr := fmt.Sprintf("%v", value)
		result = replaceAll(result, placeholder, valueStr)
	}
	return result, nil
}

// ToolNode represents a node that calls an external tool
type ToolNode struct {
	*BaseNode
	Tool   Tool         `json:"-"`
	tracer trace.Tracer `json:"-"`
}

// Tool interface for external tools
type Tool interface {
	Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)
	GetName() string
	GetDescription() string
	GetSchema() map[string]interface{}
}

// NewToolNode creates a new tool node
func NewToolNode(id, name string, tool Tool) *ToolNode {
	return &ToolNode{
		BaseNode: &BaseNode{
			ID:          id,
			Type:        NodeTypeTool,
			Name:        name,
			Description: fmt.Sprintf("Tool node that executes %s", tool.GetName()),
			Config:      make(map[string]interface{}),
			Metadata:    make(map[string]interface{}),
		},
		Tool:   tool,
		tracer: otel.Tracer("workflow.nodes.tool"),
	}
}

// Execute executes the tool node
func (n *ToolNode) Execute(ctx context.Context, state *WorkflowState) (*NodeOutput, error) {
	ctx, span := n.tracer.Start(ctx, "tool_node.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("node.id", n.ID),
		attribute.String("node.type", string(n.Type)),
		attribute.String("tool.name", n.Tool.GetName()),
	)

	// Execute the tool
	result, err := n.Tool.Execute(ctx, state.Data)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	// Prepare output
	output := &NodeOutput{
		Data: result,
		Metadata: map[string]interface{}{
			"tool_name":   n.Tool.GetName(),
			"tool_result": result,
		},
	}

	return output, nil
}

// DecisionNode represents a node that makes decisions based on conditions
type DecisionNode struct {
	*BaseNode
	Conditions map[string]EdgeCondition `json:"conditions"`
	Default    string                   `json:"default,omitempty"`
	tracer     trace.Tracer             `json:"-"`
}

// NewDecisionNode creates a new decision node
func NewDecisionNode(id, name string) *DecisionNode {
	return &DecisionNode{
		BaseNode: &BaseNode{
			ID:          id,
			Type:        NodeTypeDecision,
			Name:        name,
			Description: "Decision node that routes based on conditions",
			Config:      make(map[string]interface{}),
			Metadata:    make(map[string]interface{}),
		},
		Conditions: make(map[string]EdgeCondition),
		tracer:     otel.Tracer("workflow.nodes.decision"),
	}
}

// AddCondition adds a condition for routing
func (n *DecisionNode) AddCondition(nextNode string, condition EdgeCondition) {
	n.Conditions[nextNode] = condition
}

// SetDefault sets the default next node
func (n *DecisionNode) SetDefault(nextNode string) {
	n.Default = nextNode
}

// Execute executes the decision node
func (n *DecisionNode) Execute(ctx context.Context, state *WorkflowState) (*NodeOutput, error) {
	ctx, span := n.tracer.Start(ctx, "decision_node.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("node.id", n.ID),
		attribute.String("node.type", string(n.Type)),
		attribute.Int("conditions.count", len(n.Conditions)),
	)

	// Evaluate conditions
	for nextNode, condition := range n.Conditions {
		shouldRoute, err := condition.Evaluate(ctx, state)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to evaluate condition for %s: %w", nextNode, err)
		}

		if shouldRoute {
			span.SetAttributes(attribute.String("decision.next_node", nextNode))
			return &NodeOutput{
				Data:     map[string]interface{}{"decision": nextNode},
				NextNode: nextNode,
				Metadata: map[string]interface{}{
					"condition_met": condition.GetDescription(),
				},
			}, nil
		}
	}

	// No condition met, use default
	if n.Default != "" {
		span.SetAttributes(attribute.String("decision.next_node", n.Default))
		return &NodeOutput{
			Data:     map[string]interface{}{"decision": n.Default},
			NextNode: n.Default,
			Metadata: map[string]interface{}{
				"used_default": true,
			},
		}, nil
	}

	// No default, end workflow
	return &NodeOutput{
		Data: map[string]interface{}{"decision": "no_condition_met"},
		Metadata: map[string]interface{}{
			"no_condition_met": true,
		},
	}, nil
}

// ParallelNode represents a node that executes multiple sub-nodes in parallel
type ParallelNode struct {
	*BaseNode
	SubNodes       []Node       `json:"-"`
	MaxConcurrency int          `json:"max_concurrency"`
	tracer         trace.Tracer `json:"-"`
}

// NewParallelNode creates a new parallel node
func NewParallelNode(id, name string, maxConcurrency int) *ParallelNode {
	if maxConcurrency <= 0 {
		maxConcurrency = 5
	}

	return &ParallelNode{
		BaseNode: &BaseNode{
			ID:          id,
			Type:        NodeTypeParallel,
			Name:        name,
			Description: "Parallel node that executes multiple sub-nodes concurrently",
			Config:      make(map[string]interface{}),
			Metadata:    make(map[string]interface{}),
		},
		SubNodes:       make([]Node, 0),
		MaxConcurrency: maxConcurrency,
		tracer:         otel.Tracer("workflow.nodes.parallel"),
	}
}

// AddSubNode adds a sub-node to execute in parallel
func (n *ParallelNode) AddSubNode(node Node) {
	n.SubNodes = append(n.SubNodes, node)
}

// Execute executes the parallel node
func (n *ParallelNode) Execute(ctx context.Context, state *WorkflowState) (*NodeOutput, error) {
	ctx, span := n.tracer.Start(ctx, "parallel_node.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("node.id", n.ID),
		attribute.String("node.type", string(n.Type)),
		attribute.Int("sub_nodes.count", len(n.SubNodes)),
		attribute.Int("max_concurrency", n.MaxConcurrency),
	)

	if len(n.SubNodes) == 0 {
		return &NodeOutput{
			Data: map[string]interface{}{"results": []interface{}{}},
		}, nil
	}

	// Create semaphore for concurrency control
	semaphore := make(chan struct{}, n.MaxConcurrency)
	results := make(chan nodeResult, len(n.SubNodes))
	var wg sync.WaitGroup

	// Execute sub-nodes in parallel
	for i, subNode := range n.SubNodes {
		wg.Add(1)
		go func(index int, node Node) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Create a copy of state for this sub-node
			subState := n.copyState(state)

			// Execute sub-node
			output, err := node.Execute(ctx, subState)

			results <- nodeResult{
				Index:  index,
				Node:   node,
				Output: output,
				Error:  err,
			}
		}(i, subNode)
	}

	// Wait for all sub-nodes to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	nodeResults := make([]nodeResult, len(n.SubNodes))
	var errors []error

	for result := range results {
		nodeResults[result.Index] = result
		if result.Error != nil {
			errors = append(errors, fmt.Errorf("sub-node %s failed: %w", result.Node.GetID(), result.Error))
		}
	}

	// Check for errors
	if len(errors) > 0 {
		span.RecordError(errors[0])
		return nil, errors[0]
	}

	// Combine results
	combinedResults := make([]interface{}, len(nodeResults))
	var allMessages []Message
	combinedData := make(map[string]interface{})

	for i, result := range nodeResults {
		if result.Output != nil {
			combinedResults[i] = result.Output.Data

			// Combine messages
			if result.Output.Messages != nil {
				allMessages = append(allMessages, result.Output.Messages...)
			}

			// Combine data with sub-node prefix
			if result.Output.Data != nil {
				subNodeKey := fmt.Sprintf("subnode_%d_%s", i, result.Node.GetID())
				combinedData[subNodeKey] = result.Output.Data
			}
		}
	}

	return &NodeOutput{
		Data:     map[string]interface{}{"results": combinedResults, "combined": combinedData},
		Messages: allMessages,
		Metadata: map[string]interface{}{
			"sub_nodes_executed": len(n.SubNodes),
			"parallel_execution": true,
		},
	}, nil
}

// nodeResult represents the result of a sub-node execution
type nodeResult struct {
	Index  int
	Node   Node
	Output *NodeOutput
	Error  error
}

// copyState creates a copy of workflow state for sub-node execution
func (n *ParallelNode) copyState(state *WorkflowState) *WorkflowState {
	return &WorkflowState{
		ID:          state.ID,
		WorkflowID:  state.WorkflowID,
		Status:      state.Status,
		CurrentNode: state.CurrentNode,
		Data:        copyDataMap(state.Data),
		Messages:    append([]Message{}, state.Messages...),
		History:     state.History, // Share history
		CreatedAt:   state.CreatedAt,
		UpdatedAt:   time.Now(),
		Metadata:    copyDataMap(state.Metadata),
	}
}

// TransformNode represents a node that transforms data
type TransformNode struct {
	*BaseNode
	Transformer func(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) `json:"-"`
	tracer      trace.Tracer                                                                           `json:"-"`
}

// NewTransformNode creates a new transform node
func NewTransformNode(id, name string, transformer func(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error)) *TransformNode {
	return &TransformNode{
		BaseNode: &BaseNode{
			ID:          id,
			Type:        NodeTypeTransform,
			Name:        name,
			Description: "Transform node that processes and transforms data",
			Config:      make(map[string]interface{}),
			Metadata:    make(map[string]interface{}),
		},
		Transformer: transformer,
		tracer:      otel.Tracer("workflow.nodes.transform"),
	}
}

// Execute executes the transform node
func (n *TransformNode) Execute(ctx context.Context, state *WorkflowState) (*NodeOutput, error) {
	ctx, span := n.tracer.Start(ctx, "transform_node.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("node.id", n.ID),
		attribute.String("node.type", string(n.Type)),
	)

	// Apply transformation
	transformedData, err := n.Transformer(ctx, state.Data)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("transformation failed: %w", err)
	}

	return &NodeOutput{
		Data: transformedData,
		Metadata: map[string]interface{}{
			"transformation_applied": true,
		},
	}, nil
}

// Helper function to replace all occurrences
func replaceAll(s, old, new string) string {
	for {
		newS := replace(s, old, new)
		if newS == s {
			break
		}
		s = newS
	}
	return s
}

// Helper function to replace first occurrence
func replace(s, old, new string) string {
	if old == "" {
		return s
	}

	index := indexOf(s, old)
	if index == -1 {
		return s
	}

	return s[:index] + new + s[index+len(old):]
}

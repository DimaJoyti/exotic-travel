package workflow

import (
	"context"
	"time"
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

// WorkflowState represents the state of a workflow execution
type WorkflowState struct {
	ID          string                 `json:"id"`
	WorkflowID  string                 `json:"workflow_id"`
	Status      WorkflowStatus         `json:"status"`
	CurrentNode string                 `json:"current_node"`
	Data        map[string]interface{} `json:"data"`
	Messages    []Message              `json:"messages,omitempty"`
	History     []NodeExecution        `json:"history"`
	Error       *WorkflowError         `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// WorkflowStatus represents the status of a workflow
type WorkflowStatus string

const (
	StatusPending   WorkflowStatus = "pending"
	StatusRunning   WorkflowStatus = "running"
	StatusCompleted WorkflowStatus = "completed"
	StatusFailed    WorkflowStatus = "failed"
	StatusCancelled WorkflowStatus = "cancelled"
	StatusPaused    WorkflowStatus = "paused"
)

// NodeExecution represents the execution of a single node
type NodeExecution struct {
	NodeID    string                 `json:"node_id"`
	NodeType  string                 `json:"node_type"`
	StartTime time.Time              `json:"start_time"`
	EndTime   *time.Time             `json:"end_time,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Input     map[string]interface{} `json:"input"`
	Output    map[string]interface{} `json:"output,omitempty"`
	Error     *WorkflowError         `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// WorkflowError represents an error in workflow execution
type WorkflowError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	NodeID    string                 `json:"node_id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *WorkflowError) Error() string {
	return e.Message
}

// WorkflowInput represents input to a workflow
type WorkflowInput struct {
	UserID      string                 `json:"user_id,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	Query       string                 `json:"query,omitempty"`
	Data        map[string]interface{} `json:"data"`
	Messages    []Message              `json:"messages,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
}

// WorkflowOutput represents output from a workflow
type WorkflowOutput struct {
	Result   interface{}            `json:"result"`
	Messages []Message              `json:"messages,omitempty"`
	Data     map[string]interface{} `json:"data"`
	State    *WorkflowState         `json:"state"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Node represents a node in the workflow graph
type Node interface {
	// GetID returns the unique identifier for this node
	GetID() string

	// GetType returns the type of this node
	GetType() string

	// Execute executes the node with the given state
	Execute(ctx context.Context, state *WorkflowState) (*NodeOutput, error)

	// GetInputSchema returns the JSON schema for expected input
	GetInputSchema() map[string]interface{}

	// GetOutputSchema returns the JSON schema for expected output
	GetOutputSchema() map[string]interface{}

	// Validate validates the node configuration
	Validate() error
}

// NodeOutput represents the output from a node execution
type NodeOutput struct {
	Data     map[string]interface{} `json:"data"`
	Messages []Message              `json:"messages,omitempty"`
	NextNode string                 `json:"next_node,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Edge represents a connection between nodes
type Edge struct {
	ID        string                 `json:"id"`
	FromNode  string                 `json:"from_node"`
	ToNode    string                 `json:"to_node"`
	Condition EdgeCondition          `json:"condition,omitempty"`
	Weight    float64                `json:"weight,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// EdgeCondition represents a condition for traversing an edge
type EdgeCondition interface {
	// Evaluate evaluates the condition based on the current state
	Evaluate(ctx context.Context, state *WorkflowState) (bool, error)

	// GetDescription returns a human-readable description of the condition
	GetDescription() string
}

// WorkflowGraph represents a workflow as a directed graph
type WorkflowGraph interface {
	// GetID returns the workflow graph ID
	GetID() string

	// GetName returns the workflow graph name
	GetName() string

	// AddNode adds a node to the graph
	AddNode(node Node) error

	// AddEdge adds an edge to the graph
	AddEdge(edge *Edge) error

	// GetNode retrieves a node by ID
	GetNode(nodeID string) (Node, error)

	// GetEdges retrieves all edges from a node
	GetEdges(fromNodeID string) ([]*Edge, error)

	// GetStartNode returns the starting node ID
	GetStartNode() string

	// SetStartNode sets the starting node ID
	SetStartNode(nodeID string) error

	// Validate validates the graph structure
	Validate() error

	// Execute executes the workflow
	Execute(ctx context.Context, input *WorkflowInput) (*WorkflowOutput, error)
}

// WorkflowExecutor executes workflows
type WorkflowExecutor interface {
	// Execute executes a workflow
	Execute(ctx context.Context, graph WorkflowGraph, input *WorkflowInput) (*WorkflowOutput, error)

	// ExecuteAsync executes a workflow asynchronously
	ExecuteAsync(ctx context.Context, graph WorkflowGraph, input *WorkflowInput) (string, error)

	// GetExecution retrieves an execution by ID
	GetExecution(executionID string) (*WorkflowState, error)

	// CancelExecution cancels a running execution
	CancelExecution(executionID string) error

	// PauseExecution pauses a running execution
	PauseExecution(executionID string) error

	// ResumeExecution resumes a paused execution
	ResumeExecution(executionID string) error
}

// NodeType represents different types of nodes
type NodeType string

const (
	NodeTypeLLM       NodeType = "llm"
	NodeTypeTool      NodeType = "tool"
	NodeTypeDecision  NodeType = "decision"
	NodeTypeParallel  NodeType = "parallel"
	NodeTypeLoop      NodeType = "loop"
	NodeTypeCondition NodeType = "condition"
	NodeTypeTransform NodeType = "transform"
	NodeTypeInput     NodeType = "input"
	NodeTypeOutput    NodeType = "output"
	NodeTypeStart     NodeType = "start"
	NodeTypeEnd       NodeType = "end"
)

// BaseNode provides common functionality for all nodes
type BaseNode struct {
	ID          string                 `json:"id"`
	Type        NodeType               `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// GetID returns the node ID
func (n *BaseNode) GetID() string {
	return n.ID
}

// GetType returns the node type
func (n *BaseNode) GetType() string {
	return string(n.Type)
}

// GetInputSchema returns the default input schema
func (n *BaseNode) GetInputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"data": map[string]interface{}{
				"type": "object",
			},
		},
	}
}

// GetOutputSchema returns the default output schema
func (n *BaseNode) GetOutputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"data": map[string]interface{}{
				"type": "object",
			},
		},
	}
}

// Validate provides basic validation
func (n *BaseNode) Validate() error {
	if n.ID == "" {
		return &WorkflowError{
			Code:      "invalid_node",
			Message:   "node ID cannot be empty",
			Timestamp: time.Now(),
		}
	}

	if n.Type == "" {
		return &WorkflowError{
			Code:      "invalid_node",
			Message:   "node type cannot be empty",
			Timestamp: time.Now(),
		}
	}

	return nil
}

// SimpleCondition represents a simple condition based on data values
type SimpleCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, gt, lt, gte, lte, contains, exists
	Value    interface{} `json:"value"`
}

// Evaluate evaluates the simple condition
func (c *SimpleCondition) Evaluate(ctx context.Context, state *WorkflowState) (bool, error) {
	value, exists := state.Data[c.Field]
	if !exists {
		if c.Operator == "exists" {
			return false, nil
		}
		return false, &WorkflowError{
			Code:      "condition_error",
			Message:   "field not found: " + c.Field,
			Timestamp: time.Now(),
		}
	}

	switch c.Operator {
	case "exists":
		return true, nil
	case "eq":
		return value == c.Value, nil
	case "ne":
		return value != c.Value, nil
	case "contains":
		if str, ok := value.(string); ok {
			if searchStr, ok := c.Value.(string); ok {
				return contains(str, searchStr), nil
			}
		}
		return false, nil
	default:
		return false, &WorkflowError{
			Code:      "condition_error",
			Message:   "unsupported operator: " + c.Operator,
			Timestamp: time.Now(),
		}
	}
}

// GetDescription returns a description of the condition
func (c *SimpleCondition) GetDescription() string {
	return "Simple condition: " + c.Field + " " + c.Operator + " " + toString(c.Value)
}

// Helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(substr) > 0 && indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// WorkflowRegistry manages workflow definitions
type WorkflowRegistry interface {
	// RegisterWorkflow registers a workflow
	RegisterWorkflow(graph WorkflowGraph) error

	// GetWorkflow retrieves a workflow by ID
	GetWorkflow(workflowID string) (WorkflowGraph, error)

	// ListWorkflows lists all registered workflows
	ListWorkflows() []string

	// UnregisterWorkflow removes a workflow
	UnregisterWorkflow(workflowID string) error
}

package langgraph

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ExecutionResult represents the result of a graph execution
type ExecutionResult struct {
	ExecutionID string                 `json:"execution_id"`
	GraphID     string                 `json:"graph_id"`
	StateID     string                 `json:"state_id"`
	Status      string                 `json:"status"` // "running", "completed", "failed", "cancelled"
	StartTime   time.Time              `json:"start_time"`
	EndTime     *time.Time             `json:"end_time,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Error       string                 `json:"error,omitempty"`
	NodesVisited []string              `json:"nodes_visited"`
	FinalState  *State                 `json:"final_state,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ExecutionOptions configures graph execution
type ExecutionOptions struct {
	MaxIterations int           `json:"max_iterations"`
	Timeout       time.Duration `json:"timeout"`
	EnableTracing bool          `json:"enable_tracing"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// DefaultExecutionOptions returns default execution options
func DefaultExecutionOptions() *ExecutionOptions {
	return &ExecutionOptions{
		MaxIterations: 100,
		Timeout:       5 * time.Minute,
		EnableTracing: true,
		Metadata:      make(map[string]interface{}),
	}
}

// GraphExecutor manages graph execution
type GraphExecutor struct {
	stateManager StateManager
	tracer       trace.Tracer
	executions   map[string]*ExecutionResult
	mutex        sync.RWMutex
}

// NewGraphExecutor creates a new graph executor
func NewGraphExecutor(stateManager StateManager) *GraphExecutor {
	return &GraphExecutor{
		stateManager: stateManager,
		tracer:       otel.Tracer("langgraph.executor"),
		executions:   make(map[string]*ExecutionResult),
	}
}

// Execute executes a graph with the given input
func (e *GraphExecutor) Execute(ctx context.Context, graph *Graph, input map[string]interface{}, options *ExecutionOptions) (*ExecutionResult, error) {
	if options == nil {
		options = DefaultExecutionOptions()
	}

	// Create execution context with timeout
	execCtx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()

	// Start tracing
	execCtx, span := e.tracer.Start(execCtx, "graph_executor.execute")
	defer span.End()

	// Generate execution ID
	executionID := uuid.New().String()
	stateID := uuid.New().String()

	span.SetAttributes(
		attribute.String("execution.id", executionID),
		attribute.String("graph.id", graph.ID),
		attribute.String("state.id", stateID),
	)

	// Create execution result
	result := &ExecutionResult{
		ExecutionID:  executionID,
		GraphID:      graph.ID,
		StateID:      stateID,
		Status:       "running",
		StartTime:    time.Now(),
		NodesVisited: make([]string, 0),
		Metadata:     options.Metadata,
	}

	// Store execution result
	e.mutex.Lock()
	e.executions[executionID] = result
	e.mutex.Unlock()

	// Create initial state
	state := NewState(stateID, graph.ID)
	for key, value := range input {
		state.Set(key, value)
	}

	// Execute graph
	finalState, err := e.executeGraph(execCtx, graph, state, options, result)
	
	// Update execution result
	endTime := time.Now()
	result.EndTime = &endTime
	result.Duration = endTime.Sub(result.StartTime)

	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		span.RecordError(err)
	} else {
		result.Status = "completed"
		result.FinalState = finalState
	}

	span.SetAttributes(
		attribute.String("execution.status", result.Status),
		attribute.Int64("execution.duration_ms", result.Duration.Milliseconds()),
		attribute.Int("execution.nodes_visited", len(result.NodesVisited)),
	)

	return result, err
}

// executeGraph performs the actual graph execution
func (e *GraphExecutor) executeGraph(ctx context.Context, graph *Graph, initialState *State, options *ExecutionOptions, result *ExecutionResult) (*State, error) {
	ctx, span := e.tracer.Start(ctx, "graph_executor.execute_graph")
	defer span.End()

	// Validate graph
	if err := graph.Validate(); err != nil {
		return nil, fmt.Errorf("graph validation failed: %w", err)
	}

	// Save initial state
	if err := e.stateManager.SaveState(ctx, initialState); err != nil {
		return nil, fmt.Errorf("failed to save initial state: %w", err)
	}

	currentNodeID := graph.EntryPoint
	currentState := initialState
	iteration := 0

	for iteration < options.MaxIterations {
		iteration++

		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		span.SetAttributes(
			attribute.String("current.node", currentNodeID),
			attribute.Int("execution.iteration", iteration),
		)

		// Add to visited nodes
		result.NodesVisited = append(result.NodesVisited, currentNodeID)

		// Check if we've reached an exit point
		if e.isExitPoint(graph, currentNodeID) {
			span.SetAttributes(attribute.Bool("execution.completed", true))
			return currentState, nil
		}

		// Get current node
		node, exists := graph.Nodes[currentNodeID]
		if !exists {
			return nil, fmt.Errorf("node %s not found", currentNodeID)
		}

		// Execute current node
		nodeCtx, nodeSpan := e.tracer.Start(ctx, fmt.Sprintf("node.%s.execute", currentNodeID))
		newState, err := node.Execute(nodeCtx, currentState)
		nodeSpan.End()

		if err != nil {
			return nil, fmt.Errorf("node %s execution failed: %w", currentNodeID, err)
		}

		// Save updated state
		if err := e.stateManager.SaveState(ctx, newState); err != nil {
			return nil, fmt.Errorf("failed to save state after node %s: %w", currentNodeID, err)
		}

		currentState = newState

		// Determine next node
		nextNodeID, err := e.getNextNode(ctx, graph, currentNodeID, currentState)
		if err != nil {
			return nil, fmt.Errorf("failed to determine next node from %s: %w", currentNodeID, err)
		}

		if nextNodeID == "" {
			// No next node, execution complete
			span.SetAttributes(attribute.Bool("execution.completed", true))
			return currentState, nil
		}

		currentNodeID = nextNodeID
	}

	return nil, fmt.Errorf("maximum iterations (%d) reached, possible infinite loop", options.MaxIterations)
}

// getNextNode determines the next node to execute
func (e *GraphExecutor) getNextNode(ctx context.Context, graph *Graph, currentNodeID string, state *State) (string, error) {
	ctx, span := e.tracer.Start(ctx, "graph_executor.get_next_node")
	defer span.End()

	span.SetAttributes(attribute.String("current.node", currentNodeID))

	edges := graph.GetEdges(currentNodeID)
	if len(edges) == 0 {
		// No outgoing edges, execution ends
		return "", nil
	}

	// Evaluate edges in order
	for _, edge := range edges {
		if edge.Condition == nil {
			// Unconditional edge
			span.SetAttributes(attribute.String("next.node", edge.To))
			return edge.To, nil
		}

		// Evaluate condition
		conditionMet, err := edge.Condition.Evaluate(ctx, state)
		if err != nil {
			span.RecordError(err)
			return "", fmt.Errorf("failed to evaluate condition for edge %s->%s: %w", edge.From, edge.To, err)
		}

		if conditionMet {
			span.SetAttributes(
				attribute.String("next.node", edge.To),
				attribute.String("edge.condition", edge.Description),
			)
			return edge.To, nil
		}
	}

	// No condition matched, execution ends
	return "", nil
}

// isExitPoint checks if a node is an exit point
func (e *GraphExecutor) isExitPoint(graph *Graph, nodeID string) bool {
	for _, exitPoint := range graph.ExitPoints {
		if exitPoint == nodeID {
			return true
		}
	}
	return false
}

// GetExecution retrieves an execution result by ID
func (e *GraphExecutor) GetExecution(executionID string) (*ExecutionResult, bool) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	result, exists := e.executions[executionID]
	return result, exists
}

// ListExecutions returns all execution results
func (e *GraphExecutor) ListExecutions() []*ExecutionResult {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	results := make([]*ExecutionResult, 0, len(e.executions))
	for _, result := range e.executions {
		results = append(results, result)
	}

	return results
}

// CancelExecution cancels a running execution (if possible)
func (e *GraphExecutor) CancelExecution(executionID string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	result, exists := e.executions[executionID]
	if !exists {
		return fmt.Errorf("execution %s not found", executionID)
	}

	if result.Status != "running" {
		return fmt.Errorf("execution %s is not running (status: %s)", executionID, result.Status)
	}

	// Mark as cancelled
	result.Status = "cancelled"
	endTime := time.Now()
	result.EndTime = &endTime
	result.Duration = endTime.Sub(result.StartTime)

	return nil
}

// CleanupExecutions removes old execution results
func (e *GraphExecutor) CleanupExecutions(maxAge time.Duration) int {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	cutoff := time.Now().Add(-maxAge)
	removed := 0

	for id, result := range e.executions {
		if result.StartTime.Before(cutoff) {
			delete(e.executions, id)
			removed++
		}
	}

	return removed
}

// ExecuteAsync executes a graph asynchronously
func (e *GraphExecutor) ExecuteAsync(ctx context.Context, graph *Graph, input map[string]interface{}, options *ExecutionOptions) (string, <-chan *ExecutionResult, error) {
	if options == nil {
		options = DefaultExecutionOptions()
	}

	// Generate execution ID
	executionID := uuid.New().String()

	// Create result channel
	resultChan := make(chan *ExecutionResult, 1)

	// Start execution in goroutine
	go func() {
		defer close(resultChan)

		result, err := e.Execute(ctx, graph, input, options)
		if err != nil {
			// Create error result if execution failed to start
			if result == nil {
				result = &ExecutionResult{
					ExecutionID: executionID,
					GraphID:     graph.ID,
					Status:      "failed",
					StartTime:   time.Now(),
					Error:       err.Error(),
					Metadata:    options.Metadata,
				}
				endTime := time.Now()
				result.EndTime = &endTime
				result.Duration = endTime.Sub(result.StartTime)
			}
		}

		resultChan <- result
	}()

	return executionID, resultChan, nil
}

// GetExecutionStats returns statistics about executions
func (e *GraphExecutor) GetExecutionStats() map[string]interface{} {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_executions": len(e.executions),
		"status_counts":    make(map[string]int),
		"avg_duration_ms":  0.0,
		"total_duration_ms": 0.0,
	}

	statusCounts := make(map[string]int)
	var totalDuration time.Duration

	for _, result := range e.executions {
		statusCounts[result.Status]++
		totalDuration += result.Duration
	}

	stats["status_counts"] = statusCounts
	stats["total_duration_ms"] = totalDuration.Milliseconds()

	if len(e.executions) > 0 {
		stats["avg_duration_ms"] = float64(totalDuration.Milliseconds()) / float64(len(e.executions))
	}

	return stats
}

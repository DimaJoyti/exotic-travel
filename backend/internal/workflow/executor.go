package workflow

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

// Executor implements WorkflowExecutor
type Executor struct {
	executions map[string]*WorkflowState
	mutex      sync.RWMutex
	tracer     trace.Tracer
}

// NewExecutor creates a new workflow executor
func NewExecutor() *Executor {
	return &Executor{
		executions: make(map[string]*WorkflowState),
		tracer:     otel.Tracer("workflow.executor"),
	}
}

// Execute executes a workflow synchronously
func (e *Executor) Execute(ctx context.Context, graph WorkflowGraph, input *WorkflowInput) (*WorkflowOutput, error) {
	return graph.Execute(ctx, input)
}

// ExecuteAsync executes a workflow asynchronously
func (e *Executor) ExecuteAsync(ctx context.Context, graph WorkflowGraph, input *WorkflowInput) (string, error) {
	executionID := uuid.New().String()
	
	// Create initial state
	state := &WorkflowState{
		ID:          executionID,
		WorkflowID:  graph.GetID(),
		Status:      StatusPending,
		CurrentNode: graph.GetStartNode(),
		Data:        input.Data,
		Messages:    input.Messages,
		History:     make([]NodeExecution, 0),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}
	
	// Store execution
	e.mutex.Lock()
	e.executions[executionID] = state
	e.mutex.Unlock()
	
	// Start execution in goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				e.mutex.Lock()
				state.Status = StatusFailed
				state.Error = &WorkflowError{
					Code:      "execution_panic",
					Message:   fmt.Sprintf("execution panicked: %v", r),
					Timestamp: time.Now(),
				}
				state.UpdatedAt = time.Now()
				e.mutex.Unlock()
			}
		}()
		
		// Execute the workflow
		finalState, err := e.executeGraph(ctx, graph, state)
		
		e.mutex.Lock()
		if err != nil {
			finalState.Status = StatusFailed
			finalState.Error = &WorkflowError{
				Code:      "execution_error",
				Message:   err.Error(),
				Timestamp: time.Now(),
			}
		} else {
			finalState.Status = StatusCompleted
		}
		finalState.UpdatedAt = time.Now()
		e.executions[executionID] = finalState
		e.mutex.Unlock()
	}()
	
	return executionID, nil
}

// GetExecution retrieves an execution by ID
func (e *Executor) GetExecution(executionID string) (*WorkflowState, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	
	state, exists := e.executions[executionID]
	if !exists {
		return nil, fmt.Errorf("execution not found: %s", executionID)
	}
	
	// Return a copy to prevent external modification
	return e.copyState(state), nil
}

// CancelExecution cancels a running execution
func (e *Executor) CancelExecution(executionID string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	state, exists := e.executions[executionID]
	if !exists {
		return fmt.Errorf("execution not found: %s", executionID)
	}
	
	if state.Status != StatusRunning && state.Status != StatusPending {
		return fmt.Errorf("execution cannot be cancelled, current status: %s", state.Status)
	}
	
	state.Status = StatusCancelled
	state.UpdatedAt = time.Now()
	
	return nil
}

// PauseExecution pauses a running execution
func (e *Executor) PauseExecution(executionID string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	state, exists := e.executions[executionID]
	if !exists {
		return fmt.Errorf("execution not found: %s", executionID)
	}
	
	if state.Status != StatusRunning {
		return fmt.Errorf("execution cannot be paused, current status: %s", state.Status)
	}
	
	state.Status = StatusPaused
	state.UpdatedAt = time.Now()
	
	return nil
}

// ResumeExecution resumes a paused execution
func (e *Executor) ResumeExecution(executionID string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	state, exists := e.executions[executionID]
	if !exists {
		return fmt.Errorf("execution not found: %s", executionID)
	}
	
	if state.Status != StatusPaused {
		return fmt.Errorf("execution cannot be resumed, current status: %s", state.Status)
	}
	
	state.Status = StatusRunning
	state.UpdatedAt = time.Now()
	
	return nil
}

// executeGraph executes a workflow graph
func (e *Executor) executeGraph(ctx context.Context, graph WorkflowGraph, state *WorkflowState) (*WorkflowState, error) {
	ctx, span := e.tracer.Start(ctx, "executor.execute_graph")
	defer span.End()
	
	span.SetAttributes(
		attribute.String("workflow.id", graph.GetID()),
		attribute.String("execution.id", state.ID),
	)
	
	state.Status = StatusRunning
	state.UpdatedAt = time.Now()
	
	currentNodeID := state.CurrentNode
	maxIterations := 100 // Prevent infinite loops
	iterations := 0
	
	for currentNodeID != "" && iterations < maxIterations {
		// Check for cancellation
		select {
		case <-ctx.Done():
			state.Status = StatusCancelled
			return state, ctx.Err()
		default:
		}
		
		// Check if execution is paused
		e.mutex.RLock()
		if state.Status == StatusPaused {
			e.mutex.RUnlock()
			return state, nil
		}
		e.mutex.RUnlock()
		
		// Get current node
		node, err := graph.GetNode(currentNodeID)
		if err != nil {
			span.RecordError(err)
			return state, fmt.Errorf("failed to get node %s: %w", currentNodeID, err)
		}
		
		// Execute node
		nodeExecution := NodeExecution{
			NodeID:    currentNodeID,
			NodeType:  node.GetType(),
			StartTime: time.Now(),
			Input:     copyDataMap(state.Data),
			Metadata:  make(map[string]interface{}),
		}
		
		nodeSpan := fmt.Sprintf("node.%s.execute", currentNodeID)
		nodeCtx, nodeSpanTracer := e.tracer.Start(ctx, nodeSpan)
		
		nodeSpanTracer.SetAttributes(
			attribute.String("node.id", currentNodeID),
			attribute.String("node.type", node.GetType()),
		)
		
		output, err := node.Execute(nodeCtx, state)
		endTime := time.Now()
		nodeExecution.EndTime = &endTime
		nodeExecution.Duration = endTime.Sub(nodeExecution.StartTime)
		
		if err != nil {
			nodeSpanTracer.RecordError(err)
			nodeSpanTracer.End()
			
			nodeExecution.Error = &WorkflowError{
				Code:      "node_execution_error",
				Message:   err.Error(),
				NodeID:    currentNodeID,
				Timestamp: time.Now(),
			}
			
			state.History = append(state.History, nodeExecution)
			state.Error = nodeExecution.Error
			span.RecordError(err)
			return state, fmt.Errorf("node %s execution failed: %w", currentNodeID, err)
		}
		
		nodeSpanTracer.End()
		
		// Update state with node output
		if output.Data != nil {
			for k, v := range output.Data {
				state.Data[k] = v
			}
		}
		
		if output.Messages != nil {
			state.Messages = append(state.Messages, output.Messages...)
		}
		
		nodeExecution.Output = output.Data
		if output.Metadata != nil {
			nodeExecution.Metadata = output.Metadata
		}
		
		state.History = append(state.History, nodeExecution)
		state.UpdatedAt = time.Now()
		
		// Determine next node
		nextNodeID, err := e.determineNextNode(ctx, graph, currentNodeID, state, output)
		if err != nil {
			span.RecordError(err)
			return state, fmt.Errorf("failed to determine next node: %w", err)
		}
		
		// Update current node
		state.CurrentNode = nextNodeID
		currentNodeID = nextNodeID
		
		iterations++
	}
	
	if iterations >= maxIterations {
		return state, fmt.Errorf("workflow exceeded maximum iterations (%d)", maxIterations)
	}
	
	// Workflow completed successfully
	state.Status = StatusCompleted
	state.CurrentNode = ""
	state.UpdatedAt = time.Now()
	
	return state, nil
}

// determineNextNode determines the next node to execute
func (e *Executor) determineNextNode(ctx context.Context, graph WorkflowGraph, currentNodeID string, state *WorkflowState, output *NodeOutput) (string, error) {
	// If node output specifies next node, use it
	if output.NextNode != "" {
		return output.NextNode, nil
	}
	
	// Get edges from current node
	edges, err := graph.GetEdges(currentNodeID)
	if err != nil {
		return "", err
	}
	
	// If no edges, workflow is complete
	if len(edges) == 0 {
		return "", nil
	}
	
	// Evaluate edge conditions
	for _, edge := range edges {
		if edge.Condition == nil {
			// No condition, take this edge
			return edge.ToNode, nil
		}
		
		// Evaluate condition
		shouldTake, err := edge.Condition.Evaluate(ctx, state)
		if err != nil {
			return "", fmt.Errorf("failed to evaluate edge condition: %w", err)
		}
		
		if shouldTake {
			return edge.ToNode, nil
		}
	}
	
	// No edge condition was satisfied
	return "", fmt.Errorf("no edge condition was satisfied from node %s", currentNodeID)
}

// copyState creates a deep copy of workflow state
func (e *Executor) copyState(state *WorkflowState) *WorkflowState {
	copy := &WorkflowState{
		ID:          state.ID,
		WorkflowID:  state.WorkflowID,
		Status:      state.Status,
		CurrentNode: state.CurrentNode,
		Data:        copyDataMap(state.Data),
		Messages:    make([]Message, len(state.Messages)),
		History:     make([]NodeExecution, len(state.History)),
		CreatedAt:   state.CreatedAt,
		UpdatedAt:   state.UpdatedAt,
		Metadata:    copyDataMap(state.Metadata),
	}
	
	// Copy messages
	for i, msg := range state.Messages {
		copy.Messages[i] = msg
	}
	
	// Copy history
	for i, exec := range state.History {
		copy.History[i] = exec
	}
	
	// Copy error if present
	if state.Error != nil {
		copy.Error = &WorkflowError{
			Code:      state.Error.Code,
			Message:   state.Error.Message,
			NodeID:    state.Error.NodeID,
			Timestamp: state.Error.Timestamp,
			Details:   copyDataMap(state.Error.Details),
		}
	}
	
	return copy
}

// copyDataMap creates a shallow copy of a data map
func copyDataMap(original map[string]interface{}) map[string]interface{} {
	if original == nil {
		return make(map[string]interface{})
	}
	
	copy := make(map[string]interface{})
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

// ListExecutions returns all execution IDs
func (e *Executor) ListExecutions() []string {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	
	ids := make([]string, 0, len(e.executions))
	for id := range e.executions {
		ids = append(ids, id)
	}
	return ids
}

// CleanupCompletedExecutions removes completed executions older than the specified duration
func (e *Executor) CleanupCompletedExecutions(olderThan time.Duration) int {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	cutoff := time.Now().Add(-olderThan)
	removed := 0
	
	for id, state := range e.executions {
		if (state.Status == StatusCompleted || state.Status == StatusFailed || state.Status == StatusCancelled) &&
			state.UpdatedAt.Before(cutoff) {
			delete(e.executions, id)
			removed++
		}
	}
	
	return removed
}

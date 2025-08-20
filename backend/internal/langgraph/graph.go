package langgraph

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Graph represents a state graph for agent execution
type Graph struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Nodes        map[string]Node        `json:"nodes"`
	Edges        map[string][]Edge      `json:"edges"`
	EntryPoint   string                 `json:"entry_point"`
	ExitPoints   []string               `json:"exit_points"`
	StateManager StateManager           `json:"-"`
	mutex        sync.RWMutex           `json:"-"`
	tracer       trace.Tracer           `json:"-"`
}

// NewGraph creates a new graph
func NewGraph(id, name string, stateManager StateManager) *Graph {
	return &Graph{
		ID:           id,
		Name:         name,
		Nodes:        make(map[string]Node),
		Edges:        make(map[string][]Edge),
		ExitPoints:   make([]string, 0),
		StateManager: stateManager,
		tracer:       otel.Tracer("langgraph.graph"),
	}
}

// AddNode adds a node to the graph
func (g *Graph) AddNode(node Node) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if _, exists := g.Nodes[node.GetID()]; exists {
		return fmt.Errorf("node with ID %s already exists", node.GetID())
	}

	g.Nodes[node.GetID()] = node
	g.Edges[node.GetID()] = make([]Edge, 0)

	return nil
}

// AddEdge adds an edge between two nodes
func (g *Graph) AddEdge(edge Edge) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Validate that source and target nodes exist
	if _, exists := g.Nodes[edge.From]; !exists {
		return fmt.Errorf("source node %s does not exist", edge.From)
	}
	if _, exists := g.Nodes[edge.To]; !exists {
		return fmt.Errorf("target node %s does not exist", edge.To)
	}

	g.Edges[edge.From] = append(g.Edges[edge.From], edge)
	return nil
}

// SetEntryPoint sets the entry point of the graph
func (g *Graph) SetEntryPoint(nodeID string) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if _, exists := g.Nodes[nodeID]; !exists {
		return fmt.Errorf("node %s does not exist", nodeID)
	}

	g.EntryPoint = nodeID
	return nil
}

// AddExitPoint adds an exit point to the graph
func (g *Graph) AddExitPoint(nodeID string) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if _, exists := g.Nodes[nodeID]; !exists {
		return fmt.Errorf("node %s does not exist", nodeID)
	}

	g.ExitPoints = append(g.ExitPoints, nodeID)
	return nil
}

// Execute executes the graph with the given initial state
func (g *Graph) Execute(ctx context.Context, initialState *State) (*State, error) {
	ctx, span := g.tracer.Start(ctx, "graph.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("graph.id", g.ID),
		attribute.String("graph.name", g.Name),
		attribute.String("state.id", initialState.ID),
	)

	if g.EntryPoint == "" {
		err := fmt.Errorf("no entry point defined for graph %s", g.ID)
		span.RecordError(err)
		return nil, err
	}

	// Save initial state
	if err := g.StateManager.SaveState(ctx, initialState); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to save initial state: %w", err)
	}

	// Start execution from entry point
	currentNodeID := g.EntryPoint
	currentState := initialState
	maxIterations := 100 // Prevent infinite loops
	iteration := 0

	for iteration < maxIterations {
		iteration++
		
		span.SetAttributes(
			attribute.String("current.node", currentNodeID),
			attribute.Int("execution.iteration", iteration),
		)

		// Check if we've reached an exit point
		if g.isExitPoint(currentNodeID) {
			span.SetAttributes(attribute.Bool("execution.completed", true))
			return currentState, nil
		}

		// Get current node
		node, exists := g.Nodes[currentNodeID]
		if !exists {
			err := fmt.Errorf("node %s not found", currentNodeID)
			span.RecordError(err)
			return nil, err
		}

		// Execute current node
		nodeCtx, nodeSpan := g.tracer.Start(ctx, fmt.Sprintf("node.%s.execute", currentNodeID))
		newState, err := node.Execute(nodeCtx, currentState)
		nodeSpan.End()

		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("node %s execution failed: %w", currentNodeID, err)
		}

		// Save updated state
		if err := g.StateManager.SaveState(ctx, newState); err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to save state after node %s: %w", currentNodeID, err)
		}

		currentState = newState

		// Determine next node
		nextNodeID, err := g.getNextNode(ctx, currentNodeID, currentState)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to determine next node from %s: %w", currentNodeID, err)
		}

		if nextNodeID == "" {
			// No next node, execution complete
			span.SetAttributes(attribute.Bool("execution.completed", true))
			return currentState, nil
		}

		currentNodeID = nextNodeID
	}

	err := fmt.Errorf("maximum iterations (%d) reached, possible infinite loop", maxIterations)
	span.RecordError(err)
	return nil, err
}

// getNextNode determines the next node to execute based on edges and conditions
func (g *Graph) getNextNode(ctx context.Context, currentNodeID string, state *State) (string, error) {
	ctx, span := g.tracer.Start(ctx, "graph.get_next_node")
	defer span.End()

	span.SetAttributes(attribute.String("current.node", currentNodeID))

	edges, exists := g.Edges[currentNodeID]
	if !exists || len(edges) == 0 {
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
func (g *Graph) isExitPoint(nodeID string) bool {
	for _, exitPoint := range g.ExitPoints {
		if exitPoint == nodeID {
			return true
		}
	}
	return false
}

// GetNode retrieves a node by ID
func (g *Graph) GetNode(nodeID string) (Node, bool) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	node, exists := g.Nodes[nodeID]
	return node, exists
}

// GetEdges retrieves all edges from a node
func (g *Graph) GetEdges(nodeID string) []Edge {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	edges, exists := g.Edges[nodeID]
	if !exists {
		return make([]Edge, 0)
	}

	// Return a copy to prevent external modifications
	result := make([]Edge, len(edges))
	copy(result, edges)
	return result
}

// Validate validates the graph structure
func (g *Graph) Validate() error {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	// Check if entry point is set
	if g.EntryPoint == "" {
		return fmt.Errorf("no entry point defined")
	}

	// Check if entry point exists
	if _, exists := g.Nodes[g.EntryPoint]; !exists {
		return fmt.Errorf("entry point node %s does not exist", g.EntryPoint)
	}

	// Check if all exit points exist
	for _, exitPoint := range g.ExitPoints {
		if _, exists := g.Nodes[exitPoint]; !exists {
			return fmt.Errorf("exit point node %s does not exist", exitPoint)
		}
	}

	// Check if all edges reference existing nodes
	for fromNode, edges := range g.Edges {
		if _, exists := g.Nodes[fromNode]; !exists {
			return fmt.Errorf("edge source node %s does not exist", fromNode)
		}

		for _, edge := range edges {
			if _, exists := g.Nodes[edge.To]; !exists {
				return fmt.Errorf("edge target node %s does not exist", edge.To)
			}
		}
	}

	// Check for reachability (basic check)
	reachable := g.getReachableNodes()
	for nodeID := range g.Nodes {
		if !reachable[nodeID] && nodeID != g.EntryPoint {
			return fmt.Errorf("node %s is not reachable from entry point", nodeID)
		}
	}

	return nil
}

// getReachableNodes returns a set of nodes reachable from the entry point
func (g *Graph) getReachableNodes() map[string]bool {
	reachable := make(map[string]bool)
	visited := make(map[string]bool)

	var dfs func(nodeID string)
	dfs = func(nodeID string) {
		if visited[nodeID] {
			return
		}
		visited[nodeID] = true
		reachable[nodeID] = true

		for _, edge := range g.Edges[nodeID] {
			dfs(edge.To)
		}
	}

	if g.EntryPoint != "" {
		dfs(g.EntryPoint)
	}

	return reachable
}

// GetNodeCount returns the number of nodes in the graph
func (g *Graph) GetNodeCount() int {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return len(g.Nodes)
}

// GetEdgeCount returns the total number of edges in the graph
func (g *Graph) GetEdgeCount() int {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	count := 0
	for _, edges := range g.Edges {
		count += len(edges)
	}
	return count
}

// Clone creates a deep copy of the graph
func (g *Graph) Clone() *Graph {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	clone := &Graph{
		ID:           g.ID,
		Name:         g.Name,
		Description:  g.Description,
		Nodes:        make(map[string]Node),
		Edges:        make(map[string][]Edge),
		EntryPoint:   g.EntryPoint,
		ExitPoints:   make([]string, len(g.ExitPoints)),
		StateManager: g.StateManager,
		tracer:       g.tracer,
	}

	// Copy nodes
	for id, node := range g.Nodes {
		clone.Nodes[id] = node // Note: This is a shallow copy of nodes
	}

	// Copy edges
	for from, edges := range g.Edges {
		clone.Edges[from] = make([]Edge, len(edges))
		copy(clone.Edges[from], edges)
	}

	// Copy exit points
	copy(clone.ExitPoints, g.ExitPoints)

	return clone
}

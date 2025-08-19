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

// Graph implements WorkflowGraph
type Graph struct {
	id          string
	name        string
	description string
	nodes       map[string]Node
	edges       map[string][]*Edge
	startNode   string
	metadata    map[string]interface{}
	mutex       sync.RWMutex
	tracer      trace.Tracer
}

// NewGraph creates a new workflow graph
func NewGraph(id, name, description string) *Graph {
	return &Graph{
		id:          id,
		name:        name,
		description: description,
		nodes:       make(map[string]Node),
		edges:       make(map[string][]*Edge),
		metadata:    make(map[string]interface{}),
		tracer:      otel.Tracer("workflow.graph"),
	}
}

// GetID returns the workflow graph ID
func (g *Graph) GetID() string {
	return g.id
}

// GetName returns the workflow graph name
func (g *Graph) GetName() string {
	return g.name
}

// AddNode adds a node to the graph
func (g *Graph) AddNode(node Node) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	if err := node.Validate(); err != nil {
		return fmt.Errorf("invalid node: %w", err)
	}
	
	nodeID := node.GetID()
	if _, exists := g.nodes[nodeID]; exists {
		return fmt.Errorf("node already exists: %s", nodeID)
	}
	
	g.nodes[nodeID] = node
	
	// Initialize edges map for this node
	if g.edges[nodeID] == nil {
		g.edges[nodeID] = make([]*Edge, 0)
	}
	
	return nil
}

// AddEdge adds an edge to the graph
func (g *Graph) AddEdge(edge *Edge) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	// Validate edge
	if edge.FromNode == "" || edge.ToNode == "" {
		return fmt.Errorf("edge must have both from and to nodes")
	}
	
	// Check that both nodes exist
	if _, exists := g.nodes[edge.FromNode]; !exists {
		return fmt.Errorf("from node does not exist: %s", edge.FromNode)
	}
	
	if _, exists := g.nodes[edge.ToNode]; !exists {
		return fmt.Errorf("to node does not exist: %s", edge.ToNode)
	}
	
	// Generate edge ID if not provided
	if edge.ID == "" {
		edge.ID = fmt.Sprintf("%s->%s", edge.FromNode, edge.ToNode)
	}
	
	// Add edge to the from node's edge list
	g.edges[edge.FromNode] = append(g.edges[edge.FromNode], edge)
	
	return nil
}

// GetNode retrieves a node by ID
func (g *Graph) GetNode(nodeID string) (Node, error) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	node, exists := g.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node not found: %s", nodeID)
	}
	
	return node, nil
}

// GetEdges retrieves all edges from a node
func (g *Graph) GetEdges(fromNodeID string) ([]*Edge, error) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	edges, exists := g.edges[fromNodeID]
	if !exists {
		return []*Edge{}, nil
	}
	
	// Return a copy to prevent external modification
	result := make([]*Edge, len(edges))
	copy(result, edges)
	
	return result, nil
}

// GetStartNode returns the starting node ID
func (g *Graph) GetStartNode() string {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	return g.startNode
}

// SetStartNode sets the starting node ID
func (g *Graph) SetStartNode(nodeID string) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	if _, exists := g.nodes[nodeID]; !exists {
		return fmt.Errorf("start node does not exist: %s", nodeID)
	}
	
	g.startNode = nodeID
	return nil
}

// Validate validates the graph structure
func (g *Graph) Validate() error {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	// Check that we have at least one node
	if len(g.nodes) == 0 {
		return fmt.Errorf("graph must have at least one node")
	}
	
	// Check that start node is set and exists
	if g.startNode == "" {
		return fmt.Errorf("start node must be set")
	}
	
	if _, exists := g.nodes[g.startNode]; !exists {
		return fmt.Errorf("start node does not exist: %s", g.startNode)
	}
	
	// Validate all nodes
	for nodeID, node := range g.nodes {
		if err := node.Validate(); err != nil {
			return fmt.Errorf("invalid node %s: %w", nodeID, err)
		}
	}
	
	// Check for cycles (simple DFS-based cycle detection)
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	
	for nodeID := range g.nodes {
		if !visited[nodeID] {
			if g.hasCycle(nodeID, visited, recStack) {
				return fmt.Errorf("graph contains cycles")
			}
		}
	}
	
	return nil
}

// hasCycle performs DFS to detect cycles
func (g *Graph) hasCycle(nodeID string, visited, recStack map[string]bool) bool {
	visited[nodeID] = true
	recStack[nodeID] = true
	
	edges := g.edges[nodeID]
	for _, edge := range edges {
		if !visited[edge.ToNode] {
			if g.hasCycle(edge.ToNode, visited, recStack) {
				return true
			}
		} else if recStack[edge.ToNode] {
			return true
		}
	}
	
	recStack[nodeID] = false
	return false
}

// Execute executes the workflow
func (g *Graph) Execute(ctx context.Context, input *WorkflowInput) (*WorkflowOutput, error) {
	ctx, span := g.tracer.Start(ctx, "graph.execute")
	defer span.End()
	
	span.SetAttributes(
		attribute.String("workflow.id", g.id),
		attribute.String("workflow.name", g.name),
	)
	
	// Validate graph before execution
	if err := g.Validate(); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("invalid graph: %w", err)
	}
	
	// Create initial workflow state
	state := &WorkflowState{
		ID:          uuid.New().String(),
		WorkflowID:  g.id,
		Status:      StatusRunning,
		CurrentNode: g.startNode,
		Data:        input.Data,
		Messages:    input.Messages,
		History:     make([]NodeExecution, 0),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}
	
	// Copy input context to state data
	if input.Context != nil {
		for k, v := range input.Context {
			state.Data[k] = v
		}
	}
	
	// Add input metadata
	if input.UserID != "" {
		state.Metadata["user_id"] = input.UserID
	}
	if input.SessionID != "" {
		state.Metadata["session_id"] = input.SessionID
	}
	if input.Query != "" {
		state.Data["query"] = input.Query
	}
	
	// Execute the workflow
	executor := NewExecutor()
	finalState, err := executor.executeGraph(ctx, g, state)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	
	// Create output
	output := &WorkflowOutput{
		Result:   finalState.Data["result"],
		Messages: finalState.Messages,
		Data:     finalState.Data,
		State:    finalState,
		Metadata: finalState.Metadata,
	}
	
	return output, nil
}

// GetAllNodes returns all nodes in the graph
func (g *Graph) GetAllNodes() map[string]Node {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	// Return a copy to prevent external modification
	result := make(map[string]Node)
	for k, v := range g.nodes {
		result[k] = v
	}
	
	return result
}

// GetAllEdges returns all edges in the graph
func (g *Graph) GetAllEdges() map[string][]*Edge {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	// Return a copy to prevent external modification
	result := make(map[string][]*Edge)
	for k, v := range g.edges {
		edges := make([]*Edge, len(v))
		copy(edges, v)
		result[k] = edges
	}
	
	return result
}

// RemoveNode removes a node from the graph
func (g *Graph) RemoveNode(nodeID string) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	// Check if node exists
	if _, exists := g.nodes[nodeID]; !exists {
		return fmt.Errorf("node not found: %s", nodeID)
	}
	
	// Remove node
	delete(g.nodes, nodeID)
	
	// Remove all edges from this node
	delete(g.edges, nodeID)
	
	// Remove all edges to this node
	for fromNode, edges := range g.edges {
		filteredEdges := make([]*Edge, 0)
		for _, edge := range edges {
			if edge.ToNode != nodeID {
				filteredEdges = append(filteredEdges, edge)
			}
		}
		g.edges[fromNode] = filteredEdges
	}
	
	// Update start node if necessary
	if g.startNode == nodeID {
		g.startNode = ""
	}
	
	return nil
}

// RemoveEdge removes an edge from the graph
func (g *Graph) RemoveEdge(edgeID string) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	// Find and remove the edge
	for fromNode, edges := range g.edges {
		for i, edge := range edges {
			if edge.ID == edgeID {
				// Remove edge from slice
				g.edges[fromNode] = append(edges[:i], edges[i+1:]...)
				return nil
			}
		}
	}
	
	return fmt.Errorf("edge not found: %s", edgeID)
}

// Clone creates a deep copy of the graph
func (g *Graph) Clone() *Graph {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	clone := NewGraph(g.id+"_clone", g.name+"_clone", g.description)
	
	// Copy nodes (assuming nodes are immutable or have their own clone methods)
	for nodeID, node := range g.nodes {
		clone.nodes[nodeID] = node
	}
	
	// Copy edges
	for fromNode, edges := range g.edges {
		clonedEdges := make([]*Edge, len(edges))
		for i, edge := range edges {
			clonedEdges[i] = &Edge{
				ID:        edge.ID,
				FromNode:  edge.FromNode,
				ToNode:    edge.ToNode,
				Condition: edge.Condition,
				Weight:    edge.Weight,
				Metadata:  copyMap(edge.Metadata),
			}
		}
		clone.edges[fromNode] = clonedEdges
	}
	
	clone.startNode = g.startNode
	clone.metadata = copyMap(g.metadata)
	
	return clone
}

// copyMap creates a shallow copy of a map
func copyMap(original map[string]interface{}) map[string]interface{} {
	if original == nil {
		return nil
	}
	
	copy := make(map[string]interface{})
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

// GetMetadata returns the graph metadata
func (g *Graph) GetMetadata() map[string]interface{} {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	return copyMap(g.metadata)
}

// SetMetadata sets graph metadata
func (g *Graph) SetMetadata(key string, value interface{}) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	g.metadata[key] = value
}

// GetDescription returns the graph description
func (g *Graph) GetDescription() string {
	return g.description
}

// SetDescription sets the graph description
func (g *Graph) SetDescription(description string) {
	g.description = description
}

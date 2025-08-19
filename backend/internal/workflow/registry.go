package workflow

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Registry implements WorkflowRegistry
type Registry struct {
	workflows map[string]WorkflowGraph
	mutex     sync.RWMutex
	tracer    trace.Tracer
}

// NewRegistry creates a new workflow registry
func NewRegistry() *Registry {
	return &Registry{
		workflows: make(map[string]WorkflowGraph),
		tracer:    otel.Tracer("workflow.registry"),
	}
}

// RegisterWorkflow registers a workflow
func (r *Registry) RegisterWorkflow(graph WorkflowGraph) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	workflowID := graph.GetID()
	if workflowID == "" {
		return fmt.Errorf("workflow ID cannot be empty")
	}

	// Validate the workflow
	if err := graph.Validate(); err != nil {
		return fmt.Errorf("invalid workflow: %w", err)
	}

	// Check if workflow already exists
	if _, exists := r.workflows[workflowID]; exists {
		return fmt.Errorf("workflow already registered: %s", workflowID)
	}

	r.workflows[workflowID] = graph
	return nil
}

// GetWorkflow retrieves a workflow by ID
func (r *Registry) GetWorkflow(workflowID string) (WorkflowGraph, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	workflow, exists := r.workflows[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	return workflow, nil
}

// ListWorkflows lists all registered workflows
func (r *Registry) ListWorkflows() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	workflows := make([]string, 0, len(r.workflows))
	for id := range r.workflows {
		workflows = append(workflows, id)
	}
	return workflows
}

// UnregisterWorkflow removes a workflow
func (r *Registry) UnregisterWorkflow(workflowID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.workflows[workflowID]; !exists {
		return fmt.Errorf("workflow not found: %s", workflowID)
	}

	delete(r.workflows, workflowID)
	return nil
}

// GetWorkflowInfo returns information about a workflow
func (r *Registry) GetWorkflowInfo(workflowID string) (*WorkflowInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	workflow, exists := r.workflows[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	// Get workflow details
	nodes := make([]NodeInfo, 0)
	edges := make([]EdgeInfo, 0)

	// Extract node information
	if graph, ok := workflow.(*Graph); ok {
		allNodes := graph.GetAllNodes()
		for nodeID, node := range allNodes {
			nodeInfo := NodeInfo{
				ID:          nodeID,
				Type:        node.GetType(),
				Name:        node.GetID(), // Assuming ID is the name for now
				Description: "",           // Would need to extract from node if available
			}
			nodes = append(nodes, nodeInfo)
		}

		// Extract edge information
		allEdges := graph.GetAllEdges()
		for fromNode, nodeEdges := range allEdges {
			for _, edge := range nodeEdges {
				edgeInfo := EdgeInfo{
					ID:           edge.ID,
					FromNode:     fromNode,
					ToNode:       edge.ToNode,
					HasCondition: edge.Condition != nil,
				}
				if edge.Condition != nil {
					edgeInfo.ConditionDescription = edge.Condition.GetDescription()
				}
				edges = append(edges, edgeInfo)
			}
		}
	}

	return &WorkflowInfo{
		ID:          workflow.GetID(),
		Name:        workflow.GetName(),
		Description: "", // Would need to add description to WorkflowGraph interface
		StartNode:   workflow.GetStartNode(),
		Nodes:       nodes,
		Edges:       edges,
	}, nil
}

// WorkflowInfo represents information about a workflow
type WorkflowInfo struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	StartNode   string     `json:"start_node"`
	Nodes       []NodeInfo `json:"nodes"`
	Edges       []EdgeInfo `json:"edges"`
}

// NodeInfo represents information about a node
type NodeInfo struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// EdgeInfo represents information about an edge
type EdgeInfo struct {
	ID                   string `json:"id"`
	FromNode             string `json:"from_node"`
	ToNode               string `json:"to_node"`
	HasCondition         bool   `json:"has_condition"`
	ConditionDescription string `json:"condition_description,omitempty"`
}

// WorkflowBuilder helps build workflows programmatically
type WorkflowBuilder struct {
	graph  *Graph
	tracer trace.Tracer
}

// NewWorkflowBuilder creates a new workflow builder
func NewWorkflowBuilder(id, name, description string) *WorkflowBuilder {
	return &WorkflowBuilder{
		graph:  NewGraph(id, name, description),
		tracer: otel.Tracer("workflow.builder"),
	}
}

// AddLLMNode adds an LLM node to the workflow
func (b *WorkflowBuilder) AddLLMNode(id, name string, provider LLMProvider, prompt string) *WorkflowBuilder {
	node := NewLLMNode(id, name, provider, prompt)
	b.graph.AddNode(node)
	return b
}

// AddToolNode adds a tool node to the workflow
func (b *WorkflowBuilder) AddToolNode(id, name string, tool Tool) *WorkflowBuilder {
	node := NewToolNode(id, name, tool)
	b.graph.AddNode(node)
	return b
}

// AddDecisionNode adds a decision node to the workflow
func (b *WorkflowBuilder) AddDecisionNode(id, name string) *WorkflowBuilder {
	node := NewDecisionNode(id, name)
	b.graph.AddNode(node)
	return b
}

// AddParallelNode adds a parallel node to the workflow
func (b *WorkflowBuilder) AddParallelNode(id, name string, maxConcurrency int) *WorkflowBuilder {
	node := NewParallelNode(id, name, maxConcurrency)
	b.graph.AddNode(node)
	return b
}

// AddTransformNode adds a transform node to the workflow
func (b *WorkflowBuilder) AddTransformNode(id, name string, transformer func(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error)) *WorkflowBuilder {
	node := NewTransformNode(id, name, transformer)
	b.graph.AddNode(node)
	return b
}

// AddEdge adds an edge between two nodes
func (b *WorkflowBuilder) AddEdge(fromNode, toNode string, condition EdgeCondition) *WorkflowBuilder {
	edge := &Edge{
		FromNode:  fromNode,
		ToNode:    toNode,
		Condition: condition,
	}
	b.graph.AddEdge(edge)
	return b
}

// AddSimpleEdge adds a simple edge without conditions
func (b *WorkflowBuilder) AddSimpleEdge(fromNode, toNode string) *WorkflowBuilder {
	return b.AddEdge(fromNode, toNode, nil)
}

// SetStartNode sets the starting node
func (b *WorkflowBuilder) SetStartNode(nodeID string) *WorkflowBuilder {
	b.graph.SetStartNode(nodeID)
	return b
}

// Build builds and returns the workflow graph
func (b *WorkflowBuilder) Build() (*Graph, error) {
	if err := b.graph.Validate(); err != nil {
		return nil, fmt.Errorf("invalid workflow: %w", err)
	}

	return b.graph, nil
}

// BuildAndRegister builds the workflow and registers it
func (b *WorkflowBuilder) BuildAndRegister(registry *Registry) error {
	graph, err := b.Build()
	if err != nil {
		return err
	}

	return registry.RegisterWorkflow(graph)
}

// WorkflowTemplate represents a reusable workflow template
type WorkflowTemplate struct {
	ID          string                                              `json:"id"`
	Name        string                                              `json:"name"`
	Description string                                              `json:"description"`
	Parameters  []TemplateParameter                                 `json:"parameters"`
	Builder     func(params map[string]interface{}) (*Graph, error) `json:"-"`
}

// TemplateParameter represents a parameter for a workflow template
type TemplateParameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
}

// TemplateRegistry manages workflow templates
type TemplateRegistry struct {
	templates map[string]*WorkflowTemplate
	mutex     sync.RWMutex
}

// NewTemplateRegistry creates a new template registry
func NewTemplateRegistry() *TemplateRegistry {
	return &TemplateRegistry{
		templates: make(map[string]*WorkflowTemplate),
	}
}

// RegisterTemplate registers a workflow template
func (tr *TemplateRegistry) RegisterTemplate(template *WorkflowTemplate) error {
	tr.mutex.Lock()
	defer tr.mutex.Unlock()

	if template.ID == "" {
		return fmt.Errorf("template ID cannot be empty")
	}

	if _, exists := tr.templates[template.ID]; exists {
		return fmt.Errorf("template already registered: %s", template.ID)
	}

	tr.templates[template.ID] = template
	return nil
}

// GetTemplate retrieves a template by ID
func (tr *TemplateRegistry) GetTemplate(templateID string) (*WorkflowTemplate, error) {
	tr.mutex.RLock()
	defer tr.mutex.RUnlock()

	template, exists := tr.templates[templateID]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	return template, nil
}

// ListTemplates lists all registered templates
func (tr *TemplateRegistry) ListTemplates() []string {
	tr.mutex.RLock()
	defer tr.mutex.RUnlock()

	templates := make([]string, 0, len(tr.templates))
	for id := range tr.templates {
		templates = append(templates, id)
	}
	return templates
}

// InstantiateTemplate creates a workflow instance from a template
func (tr *TemplateRegistry) InstantiateTemplate(templateID string, params map[string]interface{}) (*Graph, error) {
	template, err := tr.GetTemplate(templateID)
	if err != nil {
		return nil, err
	}

	// Validate parameters
	if err := tr.validateParameters(template, params); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Add default values for missing parameters
	enrichedParams := tr.addDefaultValues(template, params)

	// Build workflow from template
	return template.Builder(enrichedParams)
}

// validateParameters validates template parameters
func (tr *TemplateRegistry) validateParameters(template *WorkflowTemplate, params map[string]interface{}) error {
	for _, param := range template.Parameters {
		value, exists := params[param.Name]

		if param.Required && !exists {
			return fmt.Errorf("required parameter missing: %s", param.Name)
		}

		if exists {
			// Basic type validation
			if err := tr.validateParameterType(param, value); err != nil {
				return fmt.Errorf("invalid parameter %s: %w", param.Name, err)
			}
		}
	}

	return nil
}

// validateParameterType validates parameter type
func (tr *TemplateRegistry) validateParameterType(param TemplateParameter, value interface{}) error {
	switch param.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "int":
		if _, ok := value.(int); !ok {
			return fmt.Errorf("expected int, got %T", value)
		}
	case "float":
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("expected float, got %T", value)
		}
	case "bool":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected bool, got %T", value)
		}
	}

	return nil
}

// addDefaultValues adds default values for missing parameters
func (tr *TemplateRegistry) addDefaultValues(template *WorkflowTemplate, params map[string]interface{}) map[string]interface{} {
	enriched := make(map[string]interface{})

	// Copy provided parameters
	for k, v := range params {
		enriched[k] = v
	}

	// Add defaults for missing parameters
	for _, param := range template.Parameters {
		if _, exists := enriched[param.Name]; !exists && param.Default != nil {
			enriched[param.Name] = param.Default
		}
	}

	return enriched
}

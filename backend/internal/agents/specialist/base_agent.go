package specialist

import (
	"context"
	"fmt"
	"time"

	"github.com/exotic-travel-booking/backend/internal/langgraph"
	"github.com/exotic-travel-booking/backend/internal/llm/providers"
	"github.com/exotic-travel-booking/backend/internal/tools"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// AgentRequest represents a request to a specialist agent
type AgentRequest struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	SessionID   string                 `json:"session_id"`
	AgentType   string                 `json:"agent_type"`
	Query       string                 `json:"query"`
	Parameters  map[string]interface{} `json:"parameters"`
	Context     map[string]interface{} `json:"context"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
}

// AgentResponse represents a response from a specialist agent
type AgentResponse struct {
	ID          string                 `json:"id"`
	RequestID   string                 `json:"request_id"`
	AgentType   string                 `json:"agent_type"`
	Status      string                 `json:"status"` // "success", "error", "partial"
	Result      interface{}            `json:"result"`
	Error       string                 `json:"error,omitempty"`
	Confidence  float64                `json:"confidence"`
	Suggestions []string               `json:"suggestions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	Duration    time.Duration          `json:"duration"`
	CreatedAt   time.Time              `json:"created_at"`
}

// BaseAgent provides common functionality for all specialist agents
type BaseAgent struct {
	ID           string                    `json:"id"`
	Name         string                    `json:"name"`
	Type         string                    `json:"type"`
	Description  string                    `json:"description"`
	LLMProvider  providers.LLMProvider     `json:"-"`
	ToolRegistry *tools.ToolRegistry       `json:"-"`
	StateManager langgraph.StateManager    `json:"-"`
	GraphBuilder *langgraph.GraphBuilder   `json:"-"`
	tracer       trace.Tracer              `json:"-"`
}

// NewBaseAgent creates a new base agent
func NewBaseAgent(id, name, agentType, description string, llmProvider providers.LLMProvider, toolRegistry *tools.ToolRegistry, stateManager langgraph.StateManager) *BaseAgent {
	return &BaseAgent{
		ID:           id,
		Name:         name,
		Type:         agentType,
		Description:  description,
		LLMProvider:  llmProvider,
		ToolRegistry: toolRegistry,
		StateManager: stateManager,
		GraphBuilder: langgraph.NewGraphBuilder(fmt.Sprintf("%s Graph", name), stateManager),
		tracer:       otel.Tracer(fmt.Sprintf("agent.%s", agentType)),
	}
}

// GetID returns the agent ID
func (a *BaseAgent) GetID() string {
	return a.ID
}

// GetName returns the agent name
func (a *BaseAgent) GetName() string {
	return a.Name
}

// GetType returns the agent type
func (a *BaseAgent) GetType() string {
	return a.Type
}

// GetDescription returns the agent description
func (a *BaseAgent) GetDescription() string {
	return a.Description
}

// ProcessRequest processes a request using the agent's specialized logic
func (a *BaseAgent) ProcessRequest(ctx context.Context, request *AgentRequest) (*AgentResponse, error) {
	ctx, span := a.tracer.Start(ctx, "base_agent.process_request")
	defer span.End()

	span.SetAttributes(
		attribute.String("agent.id", a.ID),
		attribute.String("agent.type", a.Type),
		attribute.String("request.id", request.ID),
		attribute.String("user.id", request.UserID),
	)

	startTime := time.Now()

	// Create response
	response := &AgentResponse{
		ID:        fmt.Sprintf("resp_%d", time.Now().UnixNano()),
		RequestID: request.ID,
		AgentType: a.Type,
		Status:    "success",
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
	}

	// This is a base implementation - should be overridden by specific agents
	response.Result = map[string]interface{}{
		"message": fmt.Sprintf("Base agent %s processed request", a.Name),
		"query":   request.Query,
	}
	response.Confidence = 0.5

	response.Duration = time.Since(startTime)
	span.SetAttributes(
		attribute.String("response.status", response.Status),
		attribute.Float64("response.confidence", response.Confidence),
		attribute.Int64("response.duration_ms", response.Duration.Milliseconds()),
	)

	return response, nil
}

// CreateLLMNode creates an LLM node for the agent's graph
func (a *BaseAgent) CreateLLMNode(nodeID, nodeName, promptTemplate, outputKey string) langgraph.Node {
	return langgraph.NewLLMNode(nodeID, nodeName, a.LLMProvider.GetName(), "llama3.2", promptTemplate, outputKey)
}

// CreateToolNode creates a tool node for the agent's graph
func (a *BaseAgent) CreateToolNode(nodeID, nodeName, toolName string, inputKeys []string, outputKey string) langgraph.Node {
	return langgraph.NewToolNode(nodeID, nodeName, toolName, inputKeys, outputKey)
}

// ExecuteLLM executes an LLM request
func (a *BaseAgent) ExecuteLLM(ctx context.Context, prompt string, maxTokens int) (string, error) {
	ctx, span := a.tracer.Start(ctx, "base_agent.execute_llm")
	defer span.End()

	span.SetAttributes(
		attribute.String("llm.provider", a.LLMProvider.GetName()),
		attribute.Int("llm.max_tokens", maxTokens),
	)

	req := &providers.GenerateRequest{
		Messages: []providers.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   maxTokens,
		Temperature: 0.7,
	}

	response, err := a.LLMProvider.GenerateResponse(ctx, req)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("LLM request failed: %w", err)
	}

	if len(response.Choices) == 0 {
		err := fmt.Errorf("no response choices returned")
		span.RecordError(err)
		return "", err
	}

	result := response.Choices[0].Message.Content
	span.SetAttributes(attribute.String("llm.response_length", fmt.Sprintf("%d", len(result))))

	return result, nil
}

// ExecuteTool executes a tool
func (a *BaseAgent) ExecuteTool(ctx context.Context, toolName string, input map[string]interface{}) (interface{}, error) {
	ctx, span := a.tracer.Start(ctx, "base_agent.execute_tool")
	defer span.End()

	span.SetAttributes(
		attribute.String("tool.name", toolName),
		attribute.Int("tool.input_params", len(input)),
	)

	_, err := a.ToolRegistry.GetTool(toolName)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("tool not found: %w", err)
	}

	// TODO: Execute tool with proper interface
	// For now, return a mock result
	result := map[string]interface{}{
		"tool":   toolName,
		"input":  input,
		"result": fmt.Sprintf("Tool %s executed successfully", toolName),
	}

	return result, nil
}

// ValidateRequest validates an agent request
func (a *BaseAgent) ValidateRequest(request *AgentRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if request.ID == "" {
		return fmt.Errorf("request ID cannot be empty")
	}

	if request.UserID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	if request.Query == "" {
		return fmt.Errorf("query cannot be empty")
	}

	if request.AgentType != a.Type {
		return fmt.Errorf("agent type mismatch: expected %s, got %s", a.Type, request.AgentType)
	}

	return nil
}

// CreateErrorResponse creates an error response
func (a *BaseAgent) CreateErrorResponse(requestID string, err error) *AgentResponse {
	return &AgentResponse{
		ID:        fmt.Sprintf("resp_%d", time.Now().UnixNano()),
		RequestID: requestID,
		AgentType: a.Type,
		Status:    "error",
		Error:     err.Error(),
		Confidence: 0.0,
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
	}
}

// ExtractParameters extracts typed parameters from the request
func (a *BaseAgent) ExtractParameters(request *AgentRequest) *RequestParameters {
	params := &RequestParameters{
		Raw: request.Parameters,
	}

	if request.Parameters == nil {
		return params
	}

	// Extract common travel parameters
	if dest, ok := request.Parameters["destination"].(string); ok {
		params.Destination = dest
	}

	if origin, ok := request.Parameters["origin"].(string); ok {
		params.Origin = origin
	}

	if startDate, ok := request.Parameters["start_date"].(string); ok {
		params.StartDate = startDate
	}

	if endDate, ok := request.Parameters["end_date"].(string); ok {
		params.EndDate = endDate
	}

	if travelers, ok := request.Parameters["travelers"].(int); ok {
		params.Travelers = travelers
	} else if travelersFloat, ok := request.Parameters["travelers"].(float64); ok {
		params.Travelers = int(travelersFloat)
	}

	if budget, ok := request.Parameters["budget"].(int); ok {
		params.Budget = budget
	} else if budgetFloat, ok := request.Parameters["budget"].(float64); ok {
		params.Budget = int(budgetFloat)
	}

	if preferences, ok := request.Parameters["preferences"].([]interface{}); ok {
		params.Preferences = make([]string, len(preferences))
		for i, pref := range preferences {
			if prefStr, ok := pref.(string); ok {
				params.Preferences[i] = prefStr
			}
		}
	}

	return params
}

// RequestParameters represents extracted and typed request parameters
type RequestParameters struct {
	Raw         map[string]interface{} `json:"raw"`
	Destination string                 `json:"destination"`
	Origin      string                 `json:"origin"`
	StartDate   string                 `json:"start_date"`
	EndDate     string                 `json:"end_date"`
	Travelers   int                    `json:"travelers"`
	Budget      int                    `json:"budget"`
	Preferences []string               `json:"preferences"`
}

// Agent interface defines the contract for all specialist agents
type Agent interface {
	// GetID returns the agent's unique identifier
	GetID() string

	// GetName returns the agent's human-readable name
	GetName() string

	// GetType returns the agent's type
	GetType() string

	// GetDescription returns the agent's description
	GetDescription() string

	// ProcessRequest processes a request and returns a response
	ProcessRequest(ctx context.Context, request *AgentRequest) (*AgentResponse, error)

	// GetCapabilities returns the agent's capabilities
	GetCapabilities() []string

	// GetSupportedParameters returns the parameters this agent supports
	GetSupportedParameters() []string
}

// AgentCapability represents a capability that an agent can perform
type AgentCapability struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// AgentMetrics represents metrics for an agent
type AgentMetrics struct {
	TotalRequests     int           `json:"total_requests"`
	SuccessfulRequests int          `json:"successful_requests"`
	FailedRequests    int           `json:"failed_requests"`
	AverageLatency    time.Duration `json:"average_latency"`
	AverageConfidence float64       `json:"average_confidence"`
	LastRequestTime   time.Time     `json:"last_request_time"`
}

// GetSuccessRate returns the success rate as a percentage
func (m *AgentMetrics) GetSuccessRate() float64 {
	if m.TotalRequests == 0 {
		return 0.0
	}
	return float64(m.SuccessfulRequests) / float64(m.TotalRequests) * 100.0
}

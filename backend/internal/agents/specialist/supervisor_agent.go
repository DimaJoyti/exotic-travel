package specialist

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/exotic-travel-booking/backend/internal/langgraph"
	"github.com/exotic-travel-booking/backend/internal/llm/providers"
	"github.com/exotic-travel-booking/backend/internal/tools"
	"go.opentelemetry.io/otel/attribute"
)

// SupervisorAgent coordinates multiple specialist agents
type SupervisorAgent struct {
	*BaseAgent
	flightAgent    *FlightAgent
	hotelAgent     *HotelAgent
	itineraryAgent *ItineraryAgent
	graph          *langgraph.Graph
	metrics        *AgentMetrics
}

// NewSupervisorAgent creates a new supervisor agent
func NewSupervisorAgent(llmProvider providers.LLMProvider, toolRegistry *tools.ToolRegistry, stateManager langgraph.StateManager) *SupervisorAgent {
	baseAgent := NewBaseAgent(
		"supervisor-agent",
		"Travel Planning Supervisor",
		"supervisor",
		"Coordinates multiple specialist agents for comprehensive travel planning",
		llmProvider,
		toolRegistry,
		stateManager,
	)

	agent := &SupervisorAgent{
		BaseAgent:      baseAgent,
		flightAgent:    NewFlightAgent(llmProvider, toolRegistry, stateManager),
		hotelAgent:     NewHotelAgent(llmProvider, toolRegistry, stateManager),
		itineraryAgent: NewItineraryAgent(llmProvider, toolRegistry, stateManager),
		metrics:        &AgentMetrics{},
	}

	// Build the supervisor coordination graph
	graph, err := agent.buildSupervisorGraph()
	if err != nil {
		fmt.Printf("Warning: Failed to build supervisor graph: %v\n", err)
	} else {
		agent.graph = graph
	}

	return agent
}

// ProcessRequest processes a comprehensive travel planning request
func (a *SupervisorAgent) ProcessRequest(ctx context.Context, request *AgentRequest) (*AgentResponse, error) {
	ctx, span := a.tracer.Start(ctx, "supervisor_agent.process_request")
	defer span.End()

	span.SetAttributes(
		attribute.String("agent.type", "supervisor"),
		attribute.String("request.id", request.ID),
	)

	startTime := time.Now()
	a.metrics.TotalRequests++
	a.metrics.LastRequestTime = startTime

	// Validate request
	if err := a.ValidateRequest(request); err != nil {
		a.metrics.FailedRequests++
		return a.CreateErrorResponse(request.ID, err), err
	}

	// Extract parameters
	params := a.ExtractParameters(request)

	// Determine which agents to involve based on the query
	agentPlan := a.determineAgentPlan(request.Query, params)

	// Process using graph if available, otherwise use direct coordination
	var result interface{}
	var confidence float64
	var err error

	if a.graph != nil {
		result, confidence, err = a.processWithGraph(ctx, request, params, agentPlan)
	} else {
		result, confidence, err = a.processDirectly(ctx, request, params, agentPlan)
	}

	// Create response
	response := &AgentResponse{
		ID:        fmt.Sprintf("supervisor_resp_%d", time.Now().UnixNano()),
		RequestID: request.ID,
		AgentType: a.Type,
		CreatedAt: time.Now(),
		Duration:  time.Since(startTime),
		Metadata: map[string]interface{}{
			"agents_involved": agentPlan.AgentsToInvolve,
			"execution_plan":  agentPlan.ExecutionPlan,
		},
	}

	if err != nil {
		a.metrics.FailedRequests++
		response.Status = "error"
		response.Error = err.Error()
		response.Confidence = 0.0
		span.RecordError(err)
	} else {
		a.metrics.SuccessfulRequests++
		response.Status = "success"
		response.Result = result
		response.Confidence = confidence
		response.Suggestions = a.generateSuggestions(params, result)
	}

	// Update metrics
	a.updateMetrics(response)

	span.SetAttributes(
		attribute.String("response.status", response.Status),
		attribute.Float64("response.confidence", response.Confidence),
		attribute.Int64("response.duration_ms", response.Duration.Milliseconds()),
		attribute.StringSlice("agents.involved", agentPlan.AgentsToInvolve),
	)

	return response, nil
}

// AgentPlan represents the plan for coordinating specialist agents
type AgentPlan struct {
	AgentsToInvolve []string `json:"agents_to_involve"`
	ExecutionPlan   string   `json:"execution_plan"`
	Priority        string   `json:"priority"` // "flights_first", "hotels_first", "itinerary_first", "parallel"
}

// determineAgentPlan determines which agents to involve and how to coordinate them
func (a *SupervisorAgent) determineAgentPlan(query string, params *RequestParameters) *AgentPlan {
	plan := &AgentPlan{
		AgentsToInvolve: []string{},
		Priority:        "parallel",
	}

	queryLower := strings.ToLower(query)

	// Determine which agents to involve based on query content
	if strings.Contains(queryLower, "flight") || strings.Contains(queryLower, "fly") || strings.Contains(queryLower, "airline") {
		plan.AgentsToInvolve = append(plan.AgentsToInvolve, "flight")
	}

	if strings.Contains(queryLower, "hotel") || strings.Contains(queryLower, "accommodation") || strings.Contains(queryLower, "stay") {
		plan.AgentsToInvolve = append(plan.AgentsToInvolve, "hotel")
	}

	if strings.Contains(queryLower, "itinerary") || strings.Contains(queryLower, "plan") || strings.Contains(queryLower, "schedule") || strings.Contains(queryLower, "activities") {
		plan.AgentsToInvolve = append(plan.AgentsToInvolve, "itinerary")
	}

	// If no specific agents mentioned, involve all for comprehensive planning
	if len(plan.AgentsToInvolve) == 0 {
		plan.AgentsToInvolve = []string{"flight", "hotel", "itinerary"}
		plan.ExecutionPlan = "comprehensive travel planning"
	} else {
		plan.ExecutionPlan = fmt.Sprintf("focused planning for: %s", strings.Join(plan.AgentsToInvolve, ", "))
	}

	// Determine execution priority
	if len(plan.AgentsToInvolve) == 1 {
		plan.Priority = "single_agent"
	} else if params.Budget > 0 && params.Budget < 1000 {
		plan.Priority = "flights_first" // Budget travelers often prioritize flights
	} else {
		plan.Priority = "parallel"
	}

	return plan
}

// processWithGraph processes the request using the LangGraph workflow
func (a *SupervisorAgent) processWithGraph(ctx context.Context, request *AgentRequest, params *RequestParameters, plan *AgentPlan) (interface{}, float64, error) {
	// Create initial state
	state := langgraph.NewState(fmt.Sprintf("supervisor_%s", request.ID), a.graph.ID)
	state.SetMultiple(map[string]interface{}{
		"query":            request.Query,
		"destination":      params.Destination,
		"origin":           params.Origin,
		"start_date":       params.StartDate,
		"end_date":         params.EndDate,
		"travelers":        params.Travelers,
		"budget":           params.Budget,
		"preferences":      params.Preferences,
		"agents_to_involve": plan.AgentsToInvolve,
		"execution_plan":   plan.ExecutionPlan,
	})

	// Execute graph
	finalState, err := a.graph.Execute(ctx, state)
	if err != nil {
		return nil, 0.0, fmt.Errorf("graph execution failed: %w", err)
	}

	// Extract results from all involved agents
	result := map[string]interface{}{
		"travel_plan": map[string]interface{}{
			"destination": params.Destination,
			"dates":       fmt.Sprintf("%s to %s", params.StartDate, params.EndDate),
			"travelers":   params.Travelers,
			"budget":      params.Budget,
		},
		"execution_summary": plan,
	}

	// Add results from each agent
	if flightResults, exists := finalState.Get("flight_results"); exists {
		result["flights"] = flightResults
	}

	if hotelResults, exists := finalState.Get("hotel_results"); exists {
		result["hotels"] = hotelResults
	}

	if itineraryResults, exists := finalState.Get("itinerary_results"); exists {
		result["itinerary"] = itineraryResults
	}

	if finalRecommendations, exists := finalState.Get("final_recommendations"); exists {
		result["recommendations"] = finalRecommendations
	}

	confidence := a.calculateOverallConfidence(finalState, plan)
	return result, confidence, nil
}

// processDirectly processes the request by directly coordinating specialist agents
func (a *SupervisorAgent) processDirectly(ctx context.Context, request *AgentRequest, params *RequestParameters, plan *AgentPlan) (interface{}, float64, error) {
	result := map[string]interface{}{
		"travel_plan": map[string]interface{}{
			"destination": params.Destination,
			"dates":       fmt.Sprintf("%s to %s", params.StartDate, params.EndDate),
			"travelers":   params.Travelers,
			"budget":      params.Budget,
		},
		"execution_summary": plan,
	}

	var totalConfidence float64
	agentCount := 0

	// Execute agents based on plan
	for _, agentType := range plan.AgentsToInvolve {
		agentRequest := &AgentRequest{
			ID:         fmt.Sprintf("%s_%s", request.ID, agentType),
			UserID:     request.UserID,
			SessionID:  request.SessionID,
			AgentType:  agentType,
			Query:      request.Query,
			Parameters: request.Parameters,
			Context:    request.Context,
			Metadata:   request.Metadata,
			CreatedAt:  time.Now(),
		}

		var agentResponse *AgentResponse
		var err error

		switch agentType {
		case "flight":
			agentResponse, err = a.flightAgent.ProcessRequest(ctx, agentRequest)
		case "hotel":
			agentResponse, err = a.hotelAgent.ProcessRequest(ctx, agentRequest)
		case "itinerary":
			agentResponse, err = a.itineraryAgent.ProcessRequest(ctx, agentRequest)
		default:
			continue
		}

		if err != nil {
			result[fmt.Sprintf("%s_error", agentType)] = err.Error()
		} else {
			result[agentType] = agentResponse.Result
			totalConfidence += agentResponse.Confidence
			agentCount++
		}
	}

	// Generate final recommendations
	finalRecommendations, err := a.generateFinalRecommendations(ctx, result, params)
	if err == nil {
		result["recommendations"] = finalRecommendations
	}

	// Calculate overall confidence
	var confidence float64
	if agentCount > 0 {
		confidence = totalConfidence / float64(agentCount)
	} else {
		confidence = 0.3
	}

	return result, confidence, nil
}

// generateFinalRecommendations generates final coordinated recommendations
func (a *SupervisorAgent) generateFinalRecommendations(ctx context.Context, results map[string]interface{}, params *RequestParameters) (string, error) {
	prompt := fmt.Sprintf(`Based on the comprehensive travel planning results, provide final coordinated recommendations:

Travel Plan: %v

Destination: %s
Budget: $%d
Travelers: %d
Preferences: %v

Provide:
1. Overall trip summary and highlights
2. Booking priority and timeline
3. Budget optimization strategies
4. Coordination tips between flights, hotels, and activities
5. Final checklist and preparation advice
6. Contingency planning suggestions

Format as a comprehensive travel planning summary.`,
		results, params.Destination, params.Budget, params.Travelers, params.Preferences)

	return a.ExecuteLLM(ctx, prompt, 800)
}

// calculateOverallConfidence calculates the overall confidence based on all agent results
func (a *SupervisorAgent) calculateOverallConfidence(state *langgraph.State, plan *AgentPlan) float64 {
	confidence := 0.2 // Base confidence

	// Increase confidence based on successful agent executions
	agentResults := 0
	if state.Has("flight_results") {
		agentResults++
	}
	if state.Has("hotel_results") {
		agentResults++
	}
	if state.Has("itinerary_results") {
		agentResults++
	}

	confidence += float64(agentResults) * 0.25

	// Bonus for comprehensive planning
	if len(plan.AgentsToInvolve) >= 3 {
		confidence += 0.05
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// generateSuggestions generates helpful suggestions
func (a *SupervisorAgent) generateSuggestions(params *RequestParameters, result interface{}) []string {
	suggestions := []string{}

	suggestions = append(suggestions, "Review all recommendations from different specialists")
	suggestions = append(suggestions, "Book flights and hotels as soon as you're satisfied with options")
	suggestions = append(suggestions, "Keep your itinerary flexible for spontaneous experiences")
	suggestions = append(suggestions, "Consider travel insurance for comprehensive protection")
	suggestions = append(suggestions, "Download offline maps and translation apps")

	return suggestions
}

// buildSupervisorGraph builds the LangGraph workflow for supervisor coordination
func (a *SupervisorAgent) buildSupervisorGraph() (*langgraph.Graph, error) {
	builder := langgraph.NewGraphBuilder("Supervisor Coordination Graph", a.StateManager)

	// Add nodes
	builder.AddStartNode("start", "Start Coordination")

	// Agent coordination nodes (simplified - in practice, these would call the actual agents)
	builder.AddFunctionNode("coordinate_flights", "Coordinate Flight Search", func(ctx context.Context, state *langgraph.State) (*langgraph.State, error) {
		newState := state.Clone()
		// Simulate flight agent coordination
		newState.Set("flight_results", "Flight search coordinated")
		return newState, nil
	})

	builder.AddFunctionNode("coordinate_hotels", "Coordinate Hotel Search", func(ctx context.Context, state *langgraph.State) (*langgraph.State, error) {
		newState := state.Clone()
		// Simulate hotel agent coordination
		newState.Set("hotel_results", "Hotel search coordinated")
		return newState, nil
	})

	builder.AddFunctionNode("coordinate_itinerary", "Coordinate Itinerary Planning", func(ctx context.Context, state *langgraph.State) (*langgraph.State, error) {
		newState := state.Clone()
		// Simulate itinerary agent coordination
		newState.Set("itinerary_results", "Itinerary planning coordinated")
		return newState, nil
	})

	// Final coordination node
	finalPrompt := `Provide final coordinated recommendations based on all specialist results:

Flight Results: {{.flight_results}}
Hotel Results: {{.hotel_results}}
Itinerary Results: {{.itinerary_results}}

Destination: {{.destination}}
Budget: ${{.budget}}
Travelers: {{.travelers}}

Generate comprehensive final recommendations and coordination advice.`

	builder.AddLLMNode("final_coordination", "Final Coordination", a.LLMProvider.GetName(), "llama3.2", finalPrompt, "final_recommendations")

	builder.AddEndNode("end", "Complete Coordination")

	// Connect nodes (simplified linear flow)
	builder.From("start").ConnectTo("coordinate_flights").ConnectTo("coordinate_hotels").ConnectTo("coordinate_itinerary").ConnectTo("final_coordination").ConnectTo("end")

	return builder.Build()
}

// updateMetrics updates agent metrics
func (a *SupervisorAgent) updateMetrics(response *AgentResponse) {
	// Update average latency
	if a.metrics.TotalRequests > 1 {
		a.metrics.AverageLatency = (a.metrics.AverageLatency*time.Duration(a.metrics.TotalRequests-1) + response.Duration) / time.Duration(a.metrics.TotalRequests)
	} else {
		a.metrics.AverageLatency = response.Duration
	}

	// Update average confidence
	if a.metrics.TotalRequests > 1 {
		a.metrics.AverageConfidence = (a.metrics.AverageConfidence*float64(a.metrics.TotalRequests-1) + response.Confidence) / float64(a.metrics.TotalRequests)
	} else {
		a.metrics.AverageConfidence = response.Confidence
	}
}

// GetCapabilities returns the supervisor agent's capabilities
func (a *SupervisorAgent) GetCapabilities() []string {
	return []string{
		"agent_coordination",
		"comprehensive_planning",
		"multi_agent_orchestration",
		"travel_optimization",
		"resource_allocation",
		"conflict_resolution",
		"priority_management",
		"quality_assurance",
	}
}

// GetSupportedParameters returns the parameters this agent supports
func (a *SupervisorAgent) GetSupportedParameters() []string {
	return []string{
		"destination",
		"origin",
		"start_date",
		"end_date",
		"travelers",
		"budget",
		"preferences",
		"priority",
		"agent_selection",
		"coordination_style",
	}
}

// GetMetrics returns the agent's performance metrics
func (a *SupervisorAgent) GetMetrics() *AgentMetrics {
	return a.metrics
}

// GetSpecialistAgents returns the specialist agents managed by this supervisor
func (a *SupervisorAgent) GetSpecialistAgents() map[string]Agent {
	return map[string]Agent{
		"flight":    a.flightAgent,
		"hotel":     a.hotelAgent,
		"itinerary": a.itineraryAgent,
	}
}

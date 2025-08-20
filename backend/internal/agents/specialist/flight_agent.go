package specialist

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/exotic-travel-booking/backend/internal/langgraph"
	"github.com/exotic-travel-booking/backend/internal/llm/providers"
	"github.com/exotic-travel-booking/backend/internal/tools"
	"go.opentelemetry.io/otel/attribute"
)

// FlightAgent specializes in flight search and booking assistance
type FlightAgent struct {
	*BaseAgent
	graph   *langgraph.Graph
	metrics *AgentMetrics
}

// NewFlightAgent creates a new flight specialist agent
func NewFlightAgent(llmProvider providers.LLMProvider, toolRegistry *tools.ToolRegistry, stateManager langgraph.StateManager) *FlightAgent {
	baseAgent := NewBaseAgent(
		"flight-agent",
		"Flight Search Specialist",
		"flight",
		"Specializes in flight search, comparison, and booking assistance",
		llmProvider,
		toolRegistry,
		stateManager,
	)

	agent := &FlightAgent{
		BaseAgent: baseAgent,
		metrics:   &AgentMetrics{},
	}

	// Build the flight search graph
	graph, err := agent.buildFlightSearchGraph()
	if err != nil {
		// Log error but don't fail creation
		fmt.Printf("Warning: Failed to build flight search graph: %v\n", err)
	} else {
		agent.graph = graph
	}

	return agent
}

// ProcessRequest processes a flight search request
func (a *FlightAgent) ProcessRequest(ctx context.Context, request *AgentRequest) (*AgentResponse, error) {
	ctx, span := a.tracer.Start(ctx, "flight_agent.process_request")
	defer span.End()

	span.SetAttributes(
		attribute.String("agent.type", "flight"),
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

	// Validate flight-specific parameters
	if err := a.validateFlightParameters(params); err != nil {
		a.metrics.FailedRequests++
		return a.CreateErrorResponse(request.ID, err), err
	}

	// Process using graph if available, otherwise use direct processing
	var result interface{}
	var confidence float64
	var err error

	if a.graph != nil {
		result, confidence, err = a.processWithGraph(ctx, request, params)
	} else {
		result, confidence, err = a.processDirectly(ctx, request, params)
	}

	// Create response
	response := &AgentResponse{
		ID:        fmt.Sprintf("flight_resp_%d", time.Now().UnixNano()),
		RequestID: request.ID,
		AgentType: a.Type,
		CreatedAt: time.Now(),
		Duration:  time.Since(startTime),
		Metadata:  make(map[string]interface{}),
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
	)

	return response, nil
}

// processWithGraph processes the request using the LangGraph workflow
func (a *FlightAgent) processWithGraph(ctx context.Context, request *AgentRequest, params *RequestParameters) (interface{}, float64, error) {
	// Create initial state
	state := langgraph.NewState(fmt.Sprintf("flight_%s", request.ID), a.graph.ID)
	state.SetMultiple(map[string]interface{}{
		"query":       request.Query,
		"origin":      params.Origin,
		"destination": params.Destination,
		"start_date":  params.StartDate,
		"end_date":    params.EndDate,
		"travelers":   params.Travelers,
		"budget":      params.Budget,
		"preferences": params.Preferences,
	})

	// Execute graph
	finalState, err := a.graph.Execute(ctx, state)
	if err != nil {
		return nil, 0.0, fmt.Errorf("graph execution failed: %w", err)
	}

	// Extract results
	flightOptions, _ := finalState.Get("flight_options")
	analysis, _ := finalState.Get("flight_analysis")
	recommendations, _ := finalState.Get("recommendations")

	result := map[string]interface{}{
		"flight_options":  flightOptions,
		"analysis":        analysis,
		"recommendations": recommendations,
		"search_criteria": map[string]interface{}{
			"origin":      params.Origin,
			"destination": params.Destination,
			"start_date":  params.StartDate,
			"end_date":    params.EndDate,
			"travelers":   params.Travelers,
			"budget":      params.Budget,
		},
	}

	// Calculate confidence based on results
	confidence := a.calculateConfidence(flightOptions, params)

	return result, confidence, nil
}

// processDirectly processes the request without using the graph
func (a *FlightAgent) processDirectly(ctx context.Context, request *AgentRequest, params *RequestParameters) (interface{}, float64, error) {
	// Step 1: Search flights using tool
	flightOptions, err := a.searchFlights(ctx, params)
	if err != nil {
		return nil, 0.0, fmt.Errorf("flight search failed: %w", err)
	}

	// Step 2: Analyze options using LLM
	analysis, err := a.analyzeFlightOptions(ctx, flightOptions, params)
	if err != nil {
		return nil, 0.0, fmt.Errorf("flight analysis failed: %w", err)
	}

	// Step 3: Generate recommendations
	recommendations, err := a.generateRecommendations(ctx, flightOptions, analysis, params)
	if err != nil {
		return nil, 0.0, fmt.Errorf("recommendation generation failed: %w", err)
	}

	result := map[string]interface{}{
		"flight_options":  flightOptions,
		"analysis":        analysis,
		"recommendations": recommendations,
		"search_criteria": map[string]interface{}{
			"origin":      params.Origin,
			"destination": params.Destination,
			"start_date":  params.StartDate,
			"end_date":    params.EndDate,
			"travelers":   params.Travelers,
			"budget":      params.Budget,
		},
	}

	confidence := a.calculateConfidence(flightOptions, params)
	return result, confidence, nil
}

// searchFlights searches for flights using the flight search tool
func (a *FlightAgent) searchFlights(ctx context.Context, params *RequestParameters) (interface{}, error) {
	input := map[string]interface{}{
		"origin":      params.Origin,
		"destination": params.Destination,
		"start_date":  params.StartDate,
		"end_date":    params.EndDate,
		"travelers":   params.Travelers,
		"budget":      params.Budget,
	}

	return a.ExecuteTool(ctx, "flight_search", input)
}

// analyzeFlightOptions analyzes flight options using LLM
func (a *FlightAgent) analyzeFlightOptions(ctx context.Context, flightOptions interface{}, params *RequestParameters) (string, error) {
	prompt := fmt.Sprintf(`Analyze the following flight options for a trip from %s to %s:

Flight Options: %v

Travel Details:
- Dates: %s to %s
- Travelers: %d
- Budget: $%d
- Preferences: %v

Please provide:
1. Summary of available options
2. Price comparison and trends
3. Best value recommendations
4. Timing considerations
5. Airline and route analysis

Keep the analysis concise but informative.`,
		params.Origin, params.Destination, flightOptions,
		params.StartDate, params.EndDate, params.Travelers, params.Budget, params.Preferences)

	return a.ExecuteLLM(ctx, prompt, 500)
}

// generateRecommendations generates flight recommendations using LLM
func (a *FlightAgent) generateRecommendations(ctx context.Context, flightOptions interface{}, analysis string, params *RequestParameters) (string, error) {
	prompt := fmt.Sprintf(`Based on the flight analysis, provide specific recommendations:

Analysis: %s

Flight Options: %v

Budget: $%d
Preferences: %v

Provide:
1. Top 3 recommended flights with reasons
2. Money-saving tips
3. Alternative dates or routes if beneficial
4. Booking timing advice
5. Additional considerations (baggage, seat selection, etc.)

Format as actionable recommendations.`,
		analysis, flightOptions, params.Budget, params.Preferences)

	return a.ExecuteLLM(ctx, prompt, 400)
}

// validateFlightParameters validates flight-specific parameters
func (a *FlightAgent) validateFlightParameters(params *RequestParameters) error {
	if params.Origin == "" {
		return fmt.Errorf("origin airport/city is required")
	}

	if params.Destination == "" {
		return fmt.Errorf("destination airport/city is required")
	}

	if params.StartDate == "" {
		return fmt.Errorf("departure date is required")
	}

	if params.Travelers <= 0 {
		return fmt.Errorf("number of travelers must be greater than 0")
	}

	if params.Travelers > 9 {
		return fmt.Errorf("maximum 9 travelers supported")
	}

	// Validate date format (basic check)
	if !a.isValidDateFormat(params.StartDate) {
		return fmt.Errorf("invalid start date format, expected YYYY-MM-DD")
	}

	if params.EndDate != "" && !a.isValidDateFormat(params.EndDate) {
		return fmt.Errorf("invalid end date format, expected YYYY-MM-DD")
	}

	return nil
}

// isValidDateFormat checks if a date string is in YYYY-MM-DD format
func (a *FlightAgent) isValidDateFormat(dateStr string) bool {
	parts := strings.Split(dateStr, "-")
	if len(parts) != 3 {
		return false
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil || year < 2024 || year > 2030 {
		return false
	}

	month, err := strconv.Atoi(parts[1])
	if err != nil || month < 1 || month > 12 {
		return false
	}

	day, err := strconv.Atoi(parts[2])
	if err != nil || day < 1 || day > 31 {
		return false
	}

	return true
}

// calculateConfidence calculates confidence score based on results
func (a *FlightAgent) calculateConfidence(flightOptions interface{}, params *RequestParameters) float64 {
	confidence := 0.5 // Base confidence

	// Increase confidence if we have flight options
	if flightOptions != nil {
		confidence += 0.3
	}

	// Increase confidence if all required parameters are provided
	if params.Origin != "" && params.Destination != "" && params.StartDate != "" {
		confidence += 0.2
	}

	// Ensure confidence is between 0 and 1
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// generateSuggestions generates helpful suggestions
func (a *FlightAgent) generateSuggestions(params *RequestParameters, result interface{}) []string {
	suggestions := []string{}

	if params.Budget > 0 && params.Budget < 500 {
		suggestions = append(suggestions, "Consider budget airlines or connecting flights to reduce costs")
	}

	if params.EndDate == "" {
		suggestions = append(suggestions, "Specify return date for round-trip options")
	}

	if len(params.Preferences) == 0 {
		suggestions = append(suggestions, "Add preferences (direct flights, specific airlines, etc.) for better recommendations")
	}

	suggestions = append(suggestions, "Book 2-8 weeks in advance for best prices")
	suggestions = append(suggestions, "Consider flexible dates for potential savings")

	return suggestions
}

// buildFlightSearchGraph builds the LangGraph workflow for flight search
func (a *FlightAgent) buildFlightSearchGraph() (*langgraph.Graph, error) {
	builder := langgraph.NewGraphBuilder("Flight Search Graph", a.StateManager)

	// Add nodes
	builder.AddStartNode("start", "Start Flight Search")

	// Flight search node
	builder.AddToolNode("search_flights", "Search Flights", "flight_search",
		[]string{"origin", "destination", "start_date", "end_date", "travelers", "budget"},
		"flight_options")

	// Analysis node
	analysisPrompt := `Analyze the flight search results:

Origin: {{.origin}}
Destination: {{.destination}}
Dates: {{.start_date}} to {{.end_date}}
Travelers: {{.travelers}}
Budget: ${{.budget}}

Flight Options: {{.flight_options}}

Provide a comprehensive analysis including:
1. Price range and average costs
2. Flight duration and routing options
3. Airline comparison
4. Best value identification
5. Timing recommendations

Keep analysis factual and helpful.`

	builder.AddLLMNode("analyze_flights", "Analyze Flight Options", a.LLMProvider.GetName(), "llama3.2", analysisPrompt, "flight_analysis")

	// Recommendations node
	recommendationPrompt := `Based on the flight analysis, provide specific recommendations:

Analysis: {{.flight_analysis}}
Budget: ${{.budget}}
Preferences: {{.preferences}}

Generate:
1. Top 3 flight recommendations with clear reasons
2. Money-saving strategies
3. Booking timing advice
4. Alternative options to consider

Format as clear, actionable recommendations.`

	builder.AddLLMNode("generate_recommendations", "Generate Recommendations", a.LLMProvider.GetName(), "llama3.2", recommendationPrompt, "recommendations")

	builder.AddEndNode("end", "Complete Flight Search")

	// Connect nodes
	builder.From("start").ConnectTo("search_flights").ConnectTo("analyze_flights").ConnectTo("generate_recommendations").ConnectTo("end")

	return builder.Build()
}

// updateMetrics updates agent metrics
func (a *FlightAgent) updateMetrics(response *AgentResponse) {
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

// GetCapabilities returns the flight agent's capabilities
func (a *FlightAgent) GetCapabilities() []string {
	return []string{
		"flight_search",
		"price_comparison",
		"route_analysis",
		"airline_comparison",
		"booking_recommendations",
		"travel_timing_advice",
		"budget_optimization",
	}
}

// GetSupportedParameters returns the parameters this agent supports
func (a *FlightAgent) GetSupportedParameters() []string {
	return []string{
		"origin",
		"destination",
		"start_date",
		"end_date",
		"travelers",
		"budget",
		"preferences",
		"airline_preference",
		"class_preference",
		"direct_flights_only",
	}
}

// GetMetrics returns the agent's performance metrics
func (a *FlightAgent) GetMetrics() *AgentMetrics {
	return a.metrics
}

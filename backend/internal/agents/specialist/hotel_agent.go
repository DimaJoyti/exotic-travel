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

// HotelAgent specializes in hotel search and accommodation recommendations
type HotelAgent struct {
	*BaseAgent
	graph   *langgraph.Graph
	metrics *AgentMetrics
}

// NewHotelAgent creates a new hotel specialist agent
func NewHotelAgent(llmProvider providers.LLMProvider, toolRegistry *tools.ToolRegistry, stateManager langgraph.StateManager) *HotelAgent {
	baseAgent := NewBaseAgent(
		"hotel-agent",
		"Hotel Search Specialist",
		"hotel",
		"Specializes in hotel search, accommodation recommendations, and booking assistance",
		llmProvider,
		toolRegistry,
		stateManager,
	)

	agent := &HotelAgent{
		BaseAgent: baseAgent,
		metrics:   &AgentMetrics{},
	}

	// Build the hotel search graph
	graph, err := agent.buildHotelSearchGraph()
	if err != nil {
		fmt.Printf("Warning: Failed to build hotel search graph: %v\n", err)
	} else {
		agent.graph = graph
	}

	return agent
}

// ProcessRequest processes a hotel search request
func (a *HotelAgent) ProcessRequest(ctx context.Context, request *AgentRequest) (*AgentResponse, error) {
	ctx, span := a.tracer.Start(ctx, "hotel_agent.process_request")
	defer span.End()

	span.SetAttributes(
		attribute.String("agent.type", "hotel"),
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

	// Validate hotel-specific parameters
	if err := a.validateHotelParameters(params); err != nil {
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
		ID:        fmt.Sprintf("hotel_resp_%d", time.Now().UnixNano()),
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
func (a *HotelAgent) processWithGraph(ctx context.Context, request *AgentRequest, params *RequestParameters) (interface{}, float64, error) {
	// Create initial state
	state := langgraph.NewState(fmt.Sprintf("hotel_%s", request.ID), a.graph.ID)
	state.SetMultiple(map[string]interface{}{
		"query":       request.Query,
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
	hotelOptions, _ := finalState.Get("hotel_options")
	analysis, _ := finalState.Get("hotel_analysis")
	recommendations, _ := finalState.Get("recommendations")
	locationInfo, _ := finalState.Get("location_info")

	result := map[string]interface{}{
		"hotel_options":    hotelOptions,
		"analysis":         analysis,
		"recommendations":  recommendations,
		"location_info":    locationInfo,
		"search_criteria": map[string]interface{}{
			"destination": params.Destination,
			"start_date":  params.StartDate,
			"end_date":    params.EndDate,
			"travelers":   params.Travelers,
			"budget":      params.Budget,
		},
	}

	confidence := a.calculateConfidence(hotelOptions, params)
	return result, confidence, nil
}

// processDirectly processes the request without using the graph
func (a *HotelAgent) processDirectly(ctx context.Context, request *AgentRequest, params *RequestParameters) (interface{}, float64, error) {
	// Step 1: Get location information
	locationInfo, err := a.getLocationInfo(ctx, params)
	if err != nil {
		return nil, 0.0, fmt.Errorf("location info failed: %w", err)
	}

	// Step 2: Search hotels using tool
	hotelOptions, err := a.searchHotels(ctx, params)
	if err != nil {
		return nil, 0.0, fmt.Errorf("hotel search failed: %w", err)
	}

	// Step 3: Analyze options using LLM
	analysis, err := a.analyzeHotelOptions(ctx, hotelOptions, params, locationInfo)
	if err != nil {
		return nil, 0.0, fmt.Errorf("hotel analysis failed: %w", err)
	}

	// Step 4: Generate recommendations
	recommendations, err := a.generateRecommendations(ctx, hotelOptions, analysis, params)
	if err != nil {
		return nil, 0.0, fmt.Errorf("recommendation generation failed: %w", err)
	}

	result := map[string]interface{}{
		"hotel_options":    hotelOptions,
		"analysis":         analysis,
		"recommendations":  recommendations,
		"location_info":    locationInfo,
		"search_criteria": map[string]interface{}{
			"destination": params.Destination,
			"start_date":  params.StartDate,
			"end_date":    params.EndDate,
			"travelers":   params.Travelers,
			"budget":      params.Budget,
		},
	}

	confidence := a.calculateConfidence(hotelOptions, params)
	return result, confidence, nil
}

// getLocationInfo gets information about the destination
func (a *HotelAgent) getLocationInfo(ctx context.Context, params *RequestParameters) (interface{}, error) {
	input := map[string]interface{}{
		"destination": params.Destination,
		"query":       fmt.Sprintf("hotels and accommodation areas in %s", params.Destination),
	}

	return a.ExecuteTool(ctx, "location", input)
}

// searchHotels searches for hotels using the hotel search tool
func (a *HotelAgent) searchHotels(ctx context.Context, params *RequestParameters) (interface{}, error) {
	input := map[string]interface{}{
		"destination": params.Destination,
		"start_date":  params.StartDate,
		"end_date":    params.EndDate,
		"travelers":   params.Travelers,
		"budget":      params.Budget,
	}

	return a.ExecuteTool(ctx, "hotel_search", input)
}

// analyzeHotelOptions analyzes hotel options using LLM
func (a *HotelAgent) analyzeHotelOptions(ctx context.Context, hotelOptions interface{}, params *RequestParameters, locationInfo interface{}) (string, error) {
	prompt := fmt.Sprintf(`Analyze the following hotel options for %s:

Hotel Options: %v

Location Information: %v

Travel Details:
- Dates: %s to %s
- Travelers: %d
- Budget: $%d per night
- Preferences: %v

Please provide:
1. Overview of available accommodation types
2. Price range analysis and value assessment
3. Location and neighborhood analysis
4. Amenities and facilities comparison
5. Best options for different traveler types

Keep the analysis comprehensive but concise.`,
		params.Destination, hotelOptions, locationInfo,
		params.StartDate, params.EndDate, params.Travelers, params.Budget, params.Preferences)

	return a.ExecuteLLM(ctx, prompt, 600)
}

// generateRecommendations generates hotel recommendations using LLM
func (a *HotelAgent) generateRecommendations(ctx context.Context, hotelOptions interface{}, analysis string, params *RequestParameters) (string, error) {
	prompt := fmt.Sprintf(`Based on the hotel analysis, provide specific recommendations:

Analysis: %s

Hotel Options: %v

Budget: $%d per night
Travelers: %d
Preferences: %v

Provide:
1. Top 3 hotel recommendations with detailed reasons
2. Best neighborhoods to stay in
3. Money-saving tips and alternatives
4. Booking timing and strategy advice
5. Special considerations (amenities, location, etc.)

Format as clear, actionable recommendations.`,
		analysis, hotelOptions, params.Budget, params.Travelers, params.Preferences)

	return a.ExecuteLLM(ctx, prompt, 500)
}

// validateHotelParameters validates hotel-specific parameters
func (a *HotelAgent) validateHotelParameters(params *RequestParameters) error {
	if params.Destination == "" {
		return fmt.Errorf("destination is required")
	}

	if params.StartDate == "" {
		return fmt.Errorf("check-in date is required")
	}

	if params.EndDate == "" {
		return fmt.Errorf("check-out date is required")
	}

	if params.Travelers <= 0 {
		return fmt.Errorf("number of travelers must be greater than 0")
	}

	if params.Travelers > 10 {
		return fmt.Errorf("maximum 10 travelers supported")
	}

	// Validate date format
	if !a.isValidDateFormat(params.StartDate) {
		return fmt.Errorf("invalid check-in date format, expected YYYY-MM-DD")
	}

	if !a.isValidDateFormat(params.EndDate) {
		return fmt.Errorf("invalid check-out date format, expected YYYY-MM-DD")
	}

	return nil
}

// isValidDateFormat checks if a date string is in YYYY-MM-DD format
func (a *HotelAgent) isValidDateFormat(dateStr string) bool {
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
func (a *HotelAgent) calculateConfidence(hotelOptions interface{}, params *RequestParameters) float64 {
	confidence := 0.4 // Base confidence

	// Increase confidence if we have hotel options
	if hotelOptions != nil {
		confidence += 0.3
	}

	// Increase confidence if all required parameters are provided
	if params.Destination != "" && params.StartDate != "" && params.EndDate != "" {
		confidence += 0.3
	}

	// Ensure confidence is between 0 and 1
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// generateSuggestions generates helpful suggestions
func (a *HotelAgent) generateSuggestions(params *RequestParameters, result interface{}) []string {
	suggestions := []string{}

	if params.Budget > 0 && params.Budget < 100 {
		suggestions = append(suggestions, "Consider hostels, guesthouses, or vacation rentals for budget accommodation")
	}

	if params.Travelers > 4 {
		suggestions = append(suggestions, "Look into vacation rentals or connecting rooms for large groups")
	}

	if len(params.Preferences) == 0 {
		suggestions = append(suggestions, "Specify preferences (location, amenities, hotel type) for better recommendations")
	}

	suggestions = append(suggestions, "Book directly with hotels for potential perks and better cancellation policies")
	suggestions = append(suggestions, "Check reviews on multiple platforms before booking")
	suggestions = append(suggestions, "Consider location vs. price trade-offs based on your itinerary")

	return suggestions
}

// buildHotelSearchGraph builds the LangGraph workflow for hotel search
func (a *HotelAgent) buildHotelSearchGraph() (*langgraph.Graph, error) {
	builder := langgraph.NewGraphBuilder("Hotel Search Graph", a.StateManager)

	// Add nodes
	builder.AddStartNode("start", "Start Hotel Search")

	// Location research node
	builder.AddToolNode("location_research", "Research Location", "location",
		[]string{"destination"},
		"location_info")

	// Hotel search node
	builder.AddToolNode("search_hotels", "Search Hotels", "hotel_search",
		[]string{"destination", "start_date", "end_date", "travelers", "budget"},
		"hotel_options")

	// Analysis node
	analysisPrompt := `Analyze the hotel search results:

Destination: {{.destination}}
Dates: {{.start_date}} to {{.end_date}}
Travelers: {{.travelers}}
Budget: ${{.budget}} per night

Hotel Options: {{.hotel_options}}
Location Info: {{.location_info}}

Provide comprehensive analysis including:
1. Accommodation types and price ranges
2. Location and neighborhood breakdown
3. Amenities and facilities overview
4. Value for money assessment
5. Suitability for different traveler needs

Keep analysis detailed but organized.`

	builder.AddLLMNode("analyze_hotels", "Analyze Hotel Options", a.LLMProvider.GetName(), "llama3.2", analysisPrompt, "hotel_analysis")

	// Recommendations node
	recommendationPrompt := `Based on the hotel analysis, provide specific recommendations:

Analysis: {{.hotel_analysis}}
Budget: ${{.budget}} per night
Travelers: {{.travelers}}
Preferences: {{.preferences}}

Generate:
1. Top 3 hotel recommendations with detailed explanations
2. Best neighborhoods for different needs
3. Money-saving strategies and alternatives
4. Booking tips and timing advice
5. Special considerations and warnings

Format as clear, actionable recommendations.`

	builder.AddLLMNode("generate_recommendations", "Generate Recommendations", a.LLMProvider.GetName(), "llama3.2", recommendationPrompt, "recommendations")

	builder.AddEndNode("end", "Complete Hotel Search")

	// Connect nodes
	builder.From("start").ConnectTo("location_research").ConnectTo("search_hotels").ConnectTo("analyze_hotels").ConnectTo("generate_recommendations").ConnectTo("end")

	return builder.Build()
}

// updateMetrics updates agent metrics
func (a *HotelAgent) updateMetrics(response *AgentResponse) {
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

// GetCapabilities returns the hotel agent's capabilities
func (a *HotelAgent) GetCapabilities() []string {
	return []string{
		"hotel_search",
		"accommodation_recommendations",
		"location_analysis",
		"price_comparison",
		"amenity_analysis",
		"neighborhood_guidance",
		"booking_advice",
		"budget_optimization",
	}
}

// GetSupportedParameters returns the parameters this agent supports
func (a *HotelAgent) GetSupportedParameters() []string {
	return []string{
		"destination",
		"start_date",
		"end_date",
		"travelers",
		"budget",
		"preferences",
		"hotel_type",
		"amenities",
		"location_preference",
		"star_rating",
	}
}

// GetMetrics returns the agent's performance metrics
func (a *HotelAgent) GetMetrics() *AgentMetrics {
	return a.metrics
}

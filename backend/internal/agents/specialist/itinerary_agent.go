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

// ItineraryAgent specializes in creating detailed travel itineraries
type ItineraryAgent struct {
	*BaseAgent
	graph   *langgraph.Graph
	metrics *AgentMetrics
}

// NewItineraryAgent creates a new itinerary specialist agent
func NewItineraryAgent(llmProvider providers.LLMProvider, toolRegistry *tools.ToolRegistry, stateManager langgraph.StateManager) *ItineraryAgent {
	baseAgent := NewBaseAgent(
		"itinerary-agent",
		"Itinerary Planning Specialist",
		"itinerary",
		"Specializes in creating detailed, personalized travel itineraries",
		llmProvider,
		toolRegistry,
		stateManager,
	)

	agent := &ItineraryAgent{
		BaseAgent: baseAgent,
		metrics:   &AgentMetrics{},
	}

	// Build the itinerary planning graph
	graph, err := agent.buildItineraryPlanningGraph()
	if err != nil {
		fmt.Printf("Warning: Failed to build itinerary planning graph: %v\n", err)
	} else {
		agent.graph = graph
	}

	return agent
}

// ProcessRequest processes an itinerary planning request
func (a *ItineraryAgent) ProcessRequest(ctx context.Context, request *AgentRequest) (*AgentResponse, error) {
	ctx, span := a.tracer.Start(ctx, "itinerary_agent.process_request")
	defer span.End()

	span.SetAttributes(
		attribute.String("agent.type", "itinerary"),
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

	// Validate itinerary-specific parameters
	if err := a.validateItineraryParameters(params); err != nil {
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
		ID:        fmt.Sprintf("itinerary_resp_%d", time.Now().UnixNano()),
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
func (a *ItineraryAgent) processWithGraph(ctx context.Context, request *AgentRequest, params *RequestParameters) (interface{}, float64, error) {
	// Create initial state
	state := langgraph.NewState(fmt.Sprintf("itinerary_%s", request.ID), a.graph.ID)
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
	destinationInfo, _ := finalState.Get("destination_info")
	weatherInfo, _ := finalState.Get("weather_info")
	attractions, _ := finalState.Get("attractions")
	itinerary, _ := finalState.Get("itinerary")
	recommendations, _ := finalState.Get("recommendations")

	result := map[string]interface{}{
		"itinerary":        itinerary,
		"destination_info": destinationInfo,
		"weather_info":     weatherInfo,
		"attractions":      attractions,
		"recommendations":  recommendations,
		"trip_summary": map[string]interface{}{
			"destination": params.Destination,
			"duration":    a.calculateTripDuration(params.StartDate, params.EndDate),
			"travelers":   params.Travelers,
			"budget":      params.Budget,
		},
	}

	confidence := a.calculateConfidence(itinerary, params)
	return result, confidence, nil
}

// processDirectly processes the request without using the graph
func (a *ItineraryAgent) processDirectly(ctx context.Context, request *AgentRequest, params *RequestParameters) (interface{}, float64, error) {
	// Step 1: Research destination
	destinationInfo, err := a.researchDestination(ctx, params)
	if err != nil {
		return nil, 0.0, fmt.Errorf("destination research failed: %w", err)
	}

	// Step 2: Get weather information
	weatherInfo, err := a.getWeatherInfo(ctx, params)
	if err != nil {
		return nil, 0.0, fmt.Errorf("weather info failed: %w", err)
	}

	// Step 3: Find attractions and activities
	attractions, err := a.findAttractions(ctx, params)
	if err != nil {
		return nil, 0.0, fmt.Errorf("attractions search failed: %w", err)
	}

	// Step 4: Create detailed itinerary
	itinerary, err := a.createItinerary(ctx, params, destinationInfo, weatherInfo, attractions)
	if err != nil {
		return nil, 0.0, fmt.Errorf("itinerary creation failed: %w", err)
	}

	// Step 5: Generate additional recommendations
	recommendations, err := a.generateDetailedRecommendations(ctx, params, itinerary)
	if err != nil {
		return nil, 0.0, fmt.Errorf("recommendations generation failed: %w", err)
	}

	result := map[string]interface{}{
		"itinerary":        itinerary,
		"destination_info": destinationInfo,
		"weather_info":     weatherInfo,
		"attractions":      attractions,
		"recommendations":  recommendations,
		"trip_summary": map[string]interface{}{
			"destination": params.Destination,
			"duration":    a.calculateTripDuration(params.StartDate, params.EndDate),
			"travelers":   params.Travelers,
			"budget":      params.Budget,
		},
	}

	confidence := a.calculateConfidence(itinerary, params)
	return result, confidence, nil
}

// researchDestination researches the destination
func (a *ItineraryAgent) researchDestination(ctx context.Context, params *RequestParameters) (interface{}, error) {
	input := map[string]interface{}{
		"destination": params.Destination,
		"query":       fmt.Sprintf("travel information for %s", params.Destination),
	}

	return a.ExecuteTool(ctx, "location", input)
}

// getWeatherInfo gets weather information for the destination
func (a *ItineraryAgent) getWeatherInfo(ctx context.Context, params *RequestParameters) (interface{}, error) {
	input := map[string]interface{}{
		"destination": params.Destination,
		"start_date":  params.StartDate,
		"end_date":    params.EndDate,
	}

	return a.ExecuteTool(ctx, "weather", input)
}

// findAttractions finds attractions and activities
func (a *ItineraryAgent) findAttractions(ctx context.Context, params *RequestParameters) (string, error) {
	prompt := fmt.Sprintf(`Find and list attractions and activities for %s:

Travel Details:
- Destination: %s
- Dates: %s to %s
- Travelers: %d
- Budget: $%d
- Preferences: %v

Please provide:
1. Top 10 must-see attractions
2. Cultural experiences and activities
3. Food and dining recommendations
4. Shopping areas and markets
5. Day trip options
6. Entertainment and nightlife
7. Family-friendly activities (if applicable)

Organize by category and include brief descriptions.`,
		params.Destination, params.Destination,
		params.StartDate, params.EndDate, params.Travelers, params.Budget, params.Preferences)

	return a.ExecuteLLM(ctx, prompt, 800)
}

// createItinerary creates a detailed day-by-day itinerary
func (a *ItineraryAgent) createItinerary(ctx context.Context, params *RequestParameters, destinationInfo, weatherInfo interface{}, attractions string) (string, error) {
	duration := a.calculateTripDuration(params.StartDate, params.EndDate)
	
	prompt := fmt.Sprintf(`Create a detailed %d-day itinerary for %s:

Destination Information: %v
Weather Forecast: %v
Available Attractions: %s

Travel Details:
- Dates: %s to %s
- Travelers: %d
- Budget: $%d total
- Preferences: %v

Create a day-by-day itinerary including:
1. Daily schedule with specific times
2. Morning, afternoon, and evening activities
3. Restaurant recommendations for each meal
4. Transportation between locations
5. Estimated costs for each activity
6. Alternative options for bad weather
7. Rest periods and flexibility
8. Local tips and cultural notes

Format as a clear, chronological itinerary with practical details.`,
		duration, params.Destination, destinationInfo, weatherInfo, attractions,
		params.StartDate, params.EndDate, params.Travelers, params.Budget, params.Preferences)

	return a.ExecuteLLM(ctx, prompt, 1200)
}

// generateDetailedRecommendations generates additional recommendations
func (a *ItineraryAgent) generateDetailedRecommendations(ctx context.Context, params *RequestParameters, itinerary string) (string, error) {
	prompt := fmt.Sprintf(`Based on the created itinerary, provide additional recommendations:

Itinerary: %s

Travel Details:
- Destination: %s
- Budget: $%d
- Travelers: %d
- Preferences: %v

Provide:
1. Packing recommendations based on weather and activities
2. Money-saving tips and budget optimization
3. Transportation advice and options
4. Safety tips and cultural etiquette
5. Emergency contacts and important information
6. Booking priorities and timing
7. Flexibility suggestions for changes
8. Local apps and resources

Format as practical, actionable advice.`,
		itinerary, params.Destination, params.Budget, params.Travelers, params.Preferences)

	return a.ExecuteLLM(ctx, prompt, 600)
}

// validateItineraryParameters validates itinerary-specific parameters
func (a *ItineraryAgent) validateItineraryParameters(params *RequestParameters) error {
	if params.Destination == "" {
		return fmt.Errorf("destination is required")
	}

	if params.StartDate == "" {
		return fmt.Errorf("start date is required")
	}

	if params.EndDate == "" {
		return fmt.Errorf("end date is required")
	}

	if params.Travelers <= 0 {
		return fmt.Errorf("number of travelers must be greater than 0")
	}

	// Validate date format
	if !a.isValidDateFormat(params.StartDate) {
		return fmt.Errorf("invalid start date format, expected YYYY-MM-DD")
	}

	if !a.isValidDateFormat(params.EndDate) {
		return fmt.Errorf("invalid end date format, expected YYYY-MM-DD")
	}

	// Validate trip duration
	duration := a.calculateTripDuration(params.StartDate, params.EndDate)
	if duration <= 0 {
		return fmt.Errorf("end date must be after start date")
	}

	if duration > 30 {
		return fmt.Errorf("maximum trip duration is 30 days")
	}

	return nil
}

// isValidDateFormat checks if a date string is in YYYY-MM-DD format
func (a *ItineraryAgent) isValidDateFormat(dateStr string) bool {
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

// calculateTripDuration calculates the duration of the trip in days
func (a *ItineraryAgent) calculateTripDuration(startDate, endDate string) int {
	// Simple calculation - in a real implementation, use proper date parsing
	startParts := strings.Split(startDate, "-")
	endParts := strings.Split(endDate, "-")
	
	if len(startParts) != 3 || len(endParts) != 3 {
		return 1 // Default to 1 day
	}
	
	startDay, _ := strconv.Atoi(startParts[2])
	endDay, _ := strconv.Atoi(endParts[2])
	
	// Simplified calculation - assumes same month
	duration := endDay - startDay + 1
	if duration <= 0 {
		return 1
	}
	
	return duration
}

// calculateConfidence calculates confidence score based on results
func (a *ItineraryAgent) calculateConfidence(itinerary interface{}, params *RequestParameters) float64 {
	confidence := 0.3 // Base confidence

	// Increase confidence if we have an itinerary
	if itinerary != nil {
		confidence += 0.4
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
func (a *ItineraryAgent) generateSuggestions(params *RequestParameters, result interface{}) []string {
	suggestions := []string{}

	duration := a.calculateTripDuration(params.StartDate, params.EndDate)
	
	if duration > 7 {
		suggestions = append(suggestions, "Consider breaking long trips into multiple destinations")
	}

	if params.Budget > 0 && params.Budget < 100*duration {
		suggestions = append(suggestions, "Budget appears tight - focus on free activities and local experiences")
	}

	if len(params.Preferences) == 0 {
		suggestions = append(suggestions, "Add interests and preferences for more personalized recommendations")
	}

	suggestions = append(suggestions, "Book popular attractions in advance to avoid disappointment")
	suggestions = append(suggestions, "Keep some flexibility in your schedule for spontaneous discoveries")
	suggestions = append(suggestions, "Research local customs and etiquette before traveling")

	return suggestions
}

// buildItineraryPlanningGraph builds the LangGraph workflow for itinerary planning
func (a *ItineraryAgent) buildItineraryPlanningGraph() (*langgraph.Graph, error) {
	builder := langgraph.NewGraphBuilder("Itinerary Planning Graph", a.StateManager)

	// Add nodes
	builder.AddStartNode("start", "Start Itinerary Planning")

	// Research nodes
	builder.AddToolNode("research_destination", "Research Destination", "location",
		[]string{"destination"},
		"destination_info")

	builder.AddToolNode("get_weather", "Get Weather Info", "weather",
		[]string{"destination", "start_date", "end_date"},
		"weather_info")

	// Attractions research node
	attractionsPrompt := `Research attractions and activities for {{.destination}}:

Travel dates: {{.start_date}} to {{.end_date}}
Travelers: {{.travelers}}
Budget: ${{.budget}}
Preferences: {{.preferences}}

Find and categorize:
1. Must-see attractions and landmarks
2. Cultural experiences and museums
3. Food and dining options
4. Activities and entertainment
5. Shopping and markets
6. Day trip possibilities

Provide detailed information with descriptions and recommendations.`

	builder.AddLLMNode("find_attractions", "Find Attractions", a.LLMProvider.GetName(), "llama3.2", attractionsPrompt, "attractions")

	// Itinerary creation node
	itineraryPrompt := `Create a detailed day-by-day itinerary:

Destination: {{.destination}}
Dates: {{.start_date}} to {{.end_date}}
Travelers: {{.travelers}}
Budget: ${{.budget}}
Preferences: {{.preferences}}

Available information:
- Destination info: {{.destination_info}}
- Weather: {{.weather_info}}
- Attractions: {{.attractions}}

Create a comprehensive itinerary with:
- Daily schedules with specific times
- Activities for morning, afternoon, evening
- Restaurant recommendations
- Transportation details
- Cost estimates
- Weather alternatives
- Local tips and cultural notes

Format as a clear, practical day-by-day guide.`

	builder.AddLLMNode("create_itinerary", "Create Itinerary", a.LLMProvider.GetName(), "llama3.2", itineraryPrompt, "itinerary")

	// Recommendations node
	recommendationsPrompt := `Provide additional travel recommendations:

Itinerary: {{.itinerary}}
Destination: {{.destination}}
Budget: ${{.budget}}
Travelers: {{.travelers}}

Generate:
1. Packing recommendations
2. Money-saving tips
3. Transportation advice
4. Safety and cultural tips
5. Booking priorities
6. Flexibility suggestions
7. Local resources and apps

Format as practical, actionable advice.`

	builder.AddLLMNode("generate_recommendations", "Generate Recommendations", a.LLMProvider.GetName(), "llama3.2", recommendationsPrompt, "recommendations")

	builder.AddEndNode("end", "Complete Itinerary Planning")

	// Connect nodes
	builder.From("start").ConnectTo("research_destination").ConnectTo("get_weather").ConnectTo("find_attractions").ConnectTo("create_itinerary").ConnectTo("generate_recommendations").ConnectTo("end")

	return builder.Build()
}

// updateMetrics updates agent metrics
func (a *ItineraryAgent) updateMetrics(response *AgentResponse) {
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

// GetCapabilities returns the itinerary agent's capabilities
func (a *ItineraryAgent) GetCapabilities() []string {
	return []string{
		"itinerary_planning",
		"destination_research",
		"activity_recommendations",
		"schedule_optimization",
		"budget_planning",
		"weather_integration",
		"cultural_guidance",
		"logistics_planning",
	}
}

// GetSupportedParameters returns the parameters this agent supports
func (a *ItineraryAgent) GetSupportedParameters() []string {
	return []string{
		"destination",
		"start_date",
		"end_date",
		"travelers",
		"budget",
		"preferences",
		"interests",
		"activity_level",
		"travel_style",
		"special_requirements",
	}
}

// GetMetrics returns the agent's performance metrics
func (a *ItineraryAgent) GetMetrics() *AgentMetrics {
	return a.metrics
}

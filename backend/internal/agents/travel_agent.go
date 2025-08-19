package agents

import (
	"context"
	"fmt"
	"time"

	"github.com/exotic-travel-booking/backend/internal/tools"
	"github.com/exotic-travel-booking/backend/internal/workflow"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)


// LLMManagerInterface defines the interface for LLM manager
type LLMManagerInterface interface {
	GenerateResponse(ctx context.Context, providerName string, req *workflow.GenerateRequest) (*workflow.GenerateResponse, error)
	GetProvider(name string) (workflow.LLMProvider, error)
	ListProviders() []string
	AddProvider(name string, provider workflow.LLMProvider) error
	RemoveProvider(name string) error
	SetDefaultProvider(name string) error
	Close() error
}

// ConversationManagerInterface defines the interface for conversation manager
type ConversationManagerInterface interface {
	StartConversation(ctx context.Context, conversationID string) error
	AddMessage(ctx context.Context, conversationID string, message workflow.Message) error
	GetHistory(ctx context.Context, conversationID string, limit int) ([]workflow.Message, error)
	GenerateResponse(ctx context.Context, conversationID string, req *workflow.GenerateRequest) (*workflow.GenerateResponse, error)
	ClearConversation(ctx context.Context, conversationID string) error
	ListConversations() []string
}

// TravelAgent represents a comprehensive travel planning agent
type TravelAgent struct {
	llmManager          LLMManagerInterface
	toolRegistry        *tools.ToolRegistry
	workflowRegistry    *workflow.Registry
	conversationManager ConversationManagerInterface
	tracer              trace.Tracer
}

// TravelRequest represents a travel planning request
type TravelRequest struct {
	UserID      string                 `json:"user_id"`
	SessionID   string                 `json:"session_id"`
	Query       string                 `json:"query"`
	Destination string                 `json:"destination,omitempty"`
	Origin      string                 `json:"origin,omitempty"`
	StartDate   string                 `json:"start_date,omitempty"`
	EndDate     string                 `json:"end_date,omitempty"`
	Travelers   int                    `json:"travelers,omitempty"`
	Budget      string                 `json:"budget,omitempty"`
	Interests   []string               `json:"interests,omitempty"`
	TravelStyle string                 `json:"travel_style,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// TravelResponse represents a travel planning response
type TravelResponse struct {
	Response        string                 `json:"response"`
	Recommendations []Recommendation       `json:"recommendations,omitempty"`
	Itinerary       *Itinerary             `json:"itinerary,omitempty"`
	NextSteps       []string               `json:"next_steps,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// Recommendation represents a travel recommendation
type Recommendation struct {
	Type        string                 `json:"type"` // flight, hotel, activity, restaurant
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Price       *Price                 `json:"price,omitempty"`
	Rating      float64                `json:"rating,omitempty"`
	URL         string                 `json:"url,omitempty"`
	Images      []string               `json:"images,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// Itinerary represents a travel itinerary
type Itinerary struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Destination string         `json:"destination"`
	StartDate   string         `json:"start_date"`
	EndDate     string         `json:"end_date"`
	Days        []ItineraryDay `json:"days"`
	TotalCost   *Price         `json:"total_cost,omitempty"`
	Notes       string         `json:"notes,omitempty"`
}

// ItineraryDay represents a single day in an itinerary
type ItineraryDay struct {
	Date       string              `json:"date"`
	Activities []ItineraryActivity `json:"activities"`
	Meals      []ItineraryMeal     `json:"meals,omitempty"`
	Notes      string              `json:"notes,omitempty"`
}

// ItineraryActivity represents an activity in the itinerary
type ItineraryActivity struct {
	Time        string                 `json:"time"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Location    string                 `json:"location"`
	Duration    string                 `json:"duration,omitempty"`
	Cost        *Price                 `json:"cost,omitempty"`
	BookingURL  string                 `json:"booking_url,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// ItineraryMeal represents a meal recommendation
type ItineraryMeal struct {
	Time        string `json:"time"` // breakfast, lunch, dinner
	Restaurant  string `json:"restaurant"`
	Cuisine     string `json:"cuisine,omitempty"`
	Location    string `json:"location"`
	Price       *Price `json:"price,omitempty"`
	Reservation string `json:"reservation,omitempty"`
}

// Price represents pricing information
type Price struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// NewTravelAgent creates a new travel agent
func NewTravelAgent(llmManager LLMManagerInterface, toolRegistry *tools.ToolRegistry) *TravelAgent {
	workflowRegistry := workflow.NewRegistry()
	
	// Create conversation manager - this would need to be injected or created elsewhere
	// to avoid the circular dependency. For now, using nil as placeholder.
	var conversationManager ConversationManagerInterface

	agent := &TravelAgent{
		llmManager:          llmManager,
		toolRegistry:        toolRegistry,
		workflowRegistry:    workflowRegistry,
		conversationManager: conversationManager,
		tracer:              otel.Tracer("agents.travel"),
	}

	// Initialize travel workflows
	agent.initializeWorkflows()

	return agent
}

// ProcessRequest processes a travel request
func (a *TravelAgent) ProcessRequest(ctx context.Context, req *TravelRequest) (*TravelResponse, error) {
	ctx, span := a.tracer.Start(ctx, "travel_agent.process_request")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", req.UserID),
		attribute.String("session.id", req.SessionID),
		attribute.String("request.query", req.Query),
	)

	// Determine the appropriate workflow based on the request
	workflowID, err := a.determineWorkflow(ctx, req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to determine workflow: %w", err)
	}

	span.SetAttributes(attribute.String("workflow.id", workflowID))

	// Get the workflow
	workflowGraph, err := a.workflowRegistry.GetWorkflow(workflowID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	// Prepare workflow input
	workflowInput := &workflow.WorkflowInput{
		UserID:    req.UserID,
		SessionID: req.SessionID,
		Query:     req.Query,
		Data:      a.prepareWorkflowData(req),
		Context:   req.Context,
	}

	// Execute the workflow
	workflowOutput, err := workflowGraph.Execute(ctx, workflowInput)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("workflow execution failed: %w", err)
	}

	// Convert workflow output to travel response
	response := a.convertWorkflowOutput(workflowOutput)

	return response, nil
}

// determineWorkflow determines which workflow to use based on the request
func (a *TravelAgent) determineWorkflow(ctx context.Context, req *TravelRequest) (string, error) {
	// Use LLM to classify the request intent
	intentPrompt := fmt.Sprintf(`
Analyze the following travel request and determine the appropriate workflow:

User Query: %s
Destination: %s
Travel Dates: %s to %s
Travelers: %d
Budget: %s

Available workflows:
1. "complete_trip_planning" - For comprehensive trip planning with flights, hotels, and itinerary
2. "flight_search" - For flight-only searches
3. "hotel_search" - For accommodation-only searches
4. "itinerary_planning" - For activity and itinerary planning
5. "travel_recommendations" - For general travel advice and recommendations

Respond with just the workflow ID that best matches the request.
`, req.Query, req.Destination, req.StartDate, req.EndDate, req.Travelers, req.Budget)

	llmReq := &workflow.GenerateRequest{
		Messages: []workflow.Message{
			{
				Role:    "user",
				Content: intentPrompt,
			},
		},
		MaxTokens:   100,
		Temperature: 0.1,
	}

	response, err := a.llmManager.GenerateResponse(ctx, "", llmReq)
	if err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	workflowID := response.Choices[0].Message.Content

	// Validate workflow ID
	validWorkflows := []string{
		"complete_trip_planning",
		"flight_search",
		"hotel_search",
		"itinerary_planning",
		"travel_recommendations",
	}

	for _, valid := range validWorkflows {
		if workflowID == valid {
			return workflowID, nil
		}
	}

	// Default to complete trip planning if classification fails
	return "complete_trip_planning", nil
}

// prepareWorkflowData prepares data for workflow execution
func (a *TravelAgent) prepareWorkflowData(req *TravelRequest) map[string]interface{} {
	data := map[string]interface{}{
		"query":      req.Query,
		"user_id":    req.UserID,
		"session_id": req.SessionID,
		"timestamp":  time.Now().Unix(),
	}

	if req.Destination != "" {
		data["destination"] = req.Destination
	}

	if req.Origin != "" {
		data["origin"] = req.Origin
	}

	if req.StartDate != "" {
		data["start_date"] = req.StartDate
	}

	if req.EndDate != "" {
		data["end_date"] = req.EndDate
	}

	if req.Travelers > 0 {
		data["travelers"] = req.Travelers
	}

	if req.Budget != "" {
		data["budget"] = req.Budget
	}

	if len(req.Interests) > 0 {
		data["interests"] = req.Interests
	}

	if req.TravelStyle != "" {
		data["travel_style"] = req.TravelStyle
	}

	if req.Preferences != nil {
		data["preferences"] = req.Preferences
	}

	return data
}

// convertWorkflowOutput converts workflow output to travel response
func (a *TravelAgent) convertWorkflowOutput(output *workflow.WorkflowOutput) *TravelResponse {
	response := &TravelResponse{
		Metadata: make(map[string]interface{}),
	}

	// Extract response text
	if result, ok := output.Result.(string); ok {
		response.Response = result
	} else if output.Data != nil {
		if responseText, ok := output.Data["response"].(string); ok {
			response.Response = responseText
		}
	}

	// Extract recommendations
	if recommendations, ok := output.Data["recommendations"].([]interface{}); ok {
		response.Recommendations = make([]Recommendation, len(recommendations))
		for i, rec := range recommendations {
			if recMap, ok := rec.(map[string]interface{}); ok {
				response.Recommendations[i] = a.convertRecommendation(recMap)
			}
		}
	}

	// Extract itinerary
	if itinerary, ok := output.Data["itinerary"].(map[string]interface{}); ok {
		response.Itinerary = a.convertItinerary(itinerary)
	}

	// Extract next steps
	if nextSteps, ok := output.Data["next_steps"].([]interface{}); ok {
		response.NextSteps = make([]string, len(nextSteps))
		for i, step := range nextSteps {
			if stepStr, ok := step.(string); ok {
				response.NextSteps[i] = stepStr
			}
		}
	}

	// Add metadata
	response.Metadata["workflow_id"] = output.State.WorkflowID
	response.Metadata["execution_id"] = output.State.ID
	response.Metadata["execution_time"] = output.State.UpdatedAt.Sub(output.State.CreatedAt).Seconds()

	return response
}

// convertRecommendation converts a recommendation map to Recommendation struct
func (a *TravelAgent) convertRecommendation(recMap map[string]interface{}) Recommendation {
	rec := Recommendation{}

	if recType, ok := recMap["type"].(string); ok {
		rec.Type = recType
	}

	if title, ok := recMap["title"].(string); ok {
		rec.Title = title
	}

	if description, ok := recMap["description"].(string); ok {
		rec.Description = description
	}

	if rating, ok := recMap["rating"].(float64); ok {
		rec.Rating = rating
	}

	if url, ok := recMap["url"].(string); ok {
		rec.URL = url
	}

	if images, ok := recMap["images"].([]interface{}); ok {
		rec.Images = make([]string, len(images))
		for i, img := range images {
			if imgStr, ok := img.(string); ok {
				rec.Images[i] = imgStr
			}
		}
	}

	if price, ok := recMap["price"].(map[string]interface{}); ok {
		rec.Price = &Price{}
		if amount, ok := price["amount"].(float64); ok {
			rec.Price.Amount = amount
		}
		if currency, ok := price["currency"].(string); ok {
			rec.Price.Currency = currency
		}
	}

	if details, ok := recMap["details"].(map[string]interface{}); ok {
		rec.Details = details
	}

	return rec
}

// convertItinerary converts an itinerary map to Itinerary struct
func (a *TravelAgent) convertItinerary(itinMap map[string]interface{}) *Itinerary {
	itin := &Itinerary{}

	if id, ok := itinMap["id"].(string); ok {
		itin.ID = id
	}

	if title, ok := itinMap["title"].(string); ok {
		itin.Title = title
	}

	if destination, ok := itinMap["destination"].(string); ok {
		itin.Destination = destination
	}

	if startDate, ok := itinMap["start_date"].(string); ok {
		itin.StartDate = startDate
	}

	if endDate, ok := itinMap["end_date"].(string); ok {
		itin.EndDate = endDate
	}

	if notes, ok := itinMap["notes"].(string); ok {
		itin.Notes = notes
	}

	// Convert days
	if days, ok := itinMap["days"].([]interface{}); ok {
		itin.Days = make([]ItineraryDay, len(days))
		for i, day := range days {
			if dayMap, ok := day.(map[string]interface{}); ok {
				itin.Days[i] = a.convertItineraryDay(dayMap)
			}
		}
	}

	// Convert total cost
	if totalCost, ok := itinMap["total_cost"].(map[string]interface{}); ok {
		itin.TotalCost = &Price{}
		if amount, ok := totalCost["amount"].(float64); ok {
			itin.TotalCost.Amount = amount
		}
		if currency, ok := totalCost["currency"].(string); ok {
			itin.TotalCost.Currency = currency
		}
	}

	return itin
}

// convertItineraryDay converts an itinerary day map to ItineraryDay struct
func (a *TravelAgent) convertItineraryDay(dayMap map[string]interface{}) ItineraryDay {
	day := ItineraryDay{}

	if date, ok := dayMap["date"].(string); ok {
		day.Date = date
	}

	if notes, ok := dayMap["notes"].(string); ok {
		day.Notes = notes
	}

	// Convert activities
	if activities, ok := dayMap["activities"].([]interface{}); ok {
		day.Activities = make([]ItineraryActivity, len(activities))
		for i, activity := range activities {
			if actMap, ok := activity.(map[string]interface{}); ok {
				day.Activities[i] = a.convertItineraryActivity(actMap)
			}
		}
	}

	// Convert meals
	if meals, ok := dayMap["meals"].([]interface{}); ok {
		day.Meals = make([]ItineraryMeal, len(meals))
		for i, meal := range meals {
			if mealMap, ok := meal.(map[string]interface{}); ok {
				day.Meals[i] = a.convertItineraryMeal(mealMap)
			}
		}
	}

	return day
}

// convertItineraryActivity converts an activity map to ItineraryActivity struct
func (a *TravelAgent) convertItineraryActivity(actMap map[string]interface{}) ItineraryActivity {
	activity := ItineraryActivity{}

	if time, ok := actMap["time"].(string); ok {
		activity.Time = time
	}

	if title, ok := actMap["title"].(string); ok {
		activity.Title = title
	}

	if description, ok := actMap["description"].(string); ok {
		activity.Description = description
	}

	if location, ok := actMap["location"].(string); ok {
		activity.Location = location
	}

	if duration, ok := actMap["duration"].(string); ok {
		activity.Duration = duration
	}

	if bookingURL, ok := actMap["booking_url"].(string); ok {
		activity.BookingURL = bookingURL
	}

	if cost, ok := actMap["cost"].(map[string]interface{}); ok {
		activity.Cost = &Price{}
		if amount, ok := cost["amount"].(float64); ok {
			activity.Cost.Amount = amount
		}
		if currency, ok := cost["currency"].(string); ok {
			activity.Cost.Currency = currency
		}
	}

	if details, ok := actMap["details"].(map[string]interface{}); ok {
		activity.Details = details
	}

	return activity
}

// convertItineraryMeal converts a meal map to ItineraryMeal struct
func (a *TravelAgent) convertItineraryMeal(mealMap map[string]interface{}) ItineraryMeal {
	meal := ItineraryMeal{}

	if time, ok := mealMap["time"].(string); ok {
		meal.Time = time
	}

	if restaurant, ok := mealMap["restaurant"].(string); ok {
		meal.Restaurant = restaurant
	}

	if cuisine, ok := mealMap["cuisine"].(string); ok {
		meal.Cuisine = cuisine
	}

	if location, ok := mealMap["location"].(string); ok {
		meal.Location = location
	}

	if reservation, ok := mealMap["reservation"].(string); ok {
		meal.Reservation = reservation
	}

	if price, ok := mealMap["price"].(map[string]interface{}); ok {
		meal.Price = &Price{}
		if amount, ok := price["amount"].(float64); ok {
			meal.Price.Amount = amount
		}
		if currency, ok := price["currency"].(string); ok {
			meal.Price.Currency = currency
		}
	}

	return meal
}

// initializeWorkflows initializes all travel workflows
func (a *TravelAgent) initializeWorkflows() {
	// Initialize complete trip planning workflow
	a.initializeCompleteTripPlanningWorkflow()

	// Initialize flight search workflow
	a.initializeFlightSearchWorkflow()

	// Initialize hotel search workflow
	a.initializeHotelSearchWorkflow()

	// Initialize itinerary planning workflow
	a.initializeItineraryPlanningWorkflow()

	// Initialize travel recommendations workflow
	a.initializeTravelRecommendationsWorkflow()
}

// initializeCompleteTripPlanningWorkflow creates the complete trip planning workflow
func (a *TravelAgent) initializeCompleteTripPlanningWorkflow() {
	builder := workflow.NewWorkflowBuilder(
		"complete_trip_planning",
		"Complete Trip Planning",
		"Comprehensive workflow for planning flights, hotels, and itinerary",
	)

	// Get LLM provider
	provider, _ := a.llmManager.GetProvider("")

	// Intent extraction node
	builder.AddLLMNode(
		"intent_extraction",
		"Extract Travel Intent",
		provider,
		`Extract detailed travel requirements from the user query:
Query: {{.query}}
Destination: {{.destination}}
Dates: {{.start_date}} to {{.end_date}}
Travelers: {{.travelers}}
Budget: {{.budget}}

Extract and structure the travel requirements including specific preferences for flights, hotels, and activities.`,
	)

	// Parallel search node for flights and hotels
	builder.AddParallelNode("parallel_search", "Search Flights and Hotels", 2)

	// Flight search node
	flightTool, _ := a.toolRegistry.GetTool("flight_search")
	builder.AddToolNode("flight_search", "Search Flights", flightTool)

	// Hotel search node
	hotelTool, _ := a.toolRegistry.GetTool("hotel_search")
	builder.AddToolNode("hotel_search", "Search Hotels", hotelTool)

	// Weather check node
	weatherTool, _ := a.toolRegistry.GetTool("weather")
	builder.AddToolNode("weather_check", "Check Weather", weatherTool)

	// Itinerary planning node
	builder.AddLLMNode(
		"itinerary_planning",
		"Plan Itinerary",
		provider,
		`Create a detailed itinerary based on:
Destination: {{.destination}}
Dates: {{.start_date}} to {{.end_date}}
Weather: {{.weather}}
Flight options: {{.flights}}
Hotel options: {{.hotels}}
User interests: {{.interests}}
Budget: {{.budget}}

Create a day-by-day itinerary with activities, meals, and recommendations.`,
	)

	// Final recommendation node
	builder.AddLLMNode(
		"final_recommendations",
		"Generate Final Recommendations",
		provider,
		`Provide final travel recommendations based on:
Flight options: {{.flights}}
Hotel options: {{.hotels}}
Itinerary: {{.itinerary}}
Weather forecast: {{.weather}}
Budget considerations: {{.budget}}

Present the best options with clear reasoning and next steps for booking.`,
	)

	// Add edges
	builder.AddSimpleEdge("intent_extraction", "parallel_search")
	builder.AddSimpleEdge("parallel_search", "weather_check")
	builder.AddSimpleEdge("weather_check", "itinerary_planning")
	builder.AddSimpleEdge("itinerary_planning", "final_recommendations")

	// Set start node
	builder.SetStartNode("intent_extraction")

	// Build and register
	builder.BuildAndRegister(a.workflowRegistry)
}

// initializeFlightSearchWorkflow creates the flight search workflow
func (a *TravelAgent) initializeFlightSearchWorkflow() {
	builder := workflow.NewWorkflowBuilder(
		"flight_search",
		"Flight Search",
		"Workflow for searching and recommending flights",
	)

	provider, _ := a.llmManager.GetProvider("")
	flightTool, _ := a.toolRegistry.GetTool("flight_search")

	// Flight search node
	builder.AddToolNode("search_flights", "Search Flights", flightTool)

	// Flight analysis node
	builder.AddLLMNode(
		"analyze_flights",
		"Analyze Flight Options",
		provider,
		`Analyze the flight search results and provide recommendations:
Flight options: {{.flights}}
User preferences: {{.preferences}}
Budget: {{.budget}}

Provide detailed analysis of the best flight options with pros and cons.`,
	)

	// Add edges
	builder.AddSimpleEdge("search_flights", "analyze_flights")

	// Set start node
	builder.SetStartNode("search_flights")

	// Build and register
	builder.BuildAndRegister(a.workflowRegistry)
}

// initializeHotelSearchWorkflow creates the hotel search workflow
func (a *TravelAgent) initializeHotelSearchWorkflow() {
	builder := workflow.NewWorkflowBuilder(
		"hotel_search",
		"Hotel Search",
		"Workflow for searching and recommending hotels",
	)

	provider, _ := a.llmManager.GetProvider("")
	hotelTool, _ := a.toolRegistry.GetTool("hotel_search")

	// Hotel search node
	builder.AddToolNode("search_hotels", "Search Hotels", hotelTool)

	// Hotel analysis node
	builder.AddLLMNode(
		"analyze_hotels",
		"Analyze Hotel Options",
		provider,
		`Analyze the hotel search results and provide recommendations:
Hotel options: {{.hotels}}
User preferences: {{.preferences}}
Budget: {{.budget}}
Location preferences: {{.location_preferences}}

Provide detailed analysis of the best hotel options with location, amenities, and value considerations.`,
	)

	// Add edges
	builder.AddSimpleEdge("search_hotels", "analyze_hotels")

	// Set start node
	builder.SetStartNode("search_hotels")

	// Build and register
	builder.BuildAndRegister(a.workflowRegistry)
}

// initializeItineraryPlanningWorkflow creates the itinerary planning workflow
func (a *TravelAgent) initializeItineraryPlanningWorkflow() {
	builder := workflow.NewWorkflowBuilder(
		"itinerary_planning",
		"Itinerary Planning",
		"Workflow for creating detailed travel itineraries",
	)

	provider, _ := a.llmManager.GetProvider("")
	locationTool, _ := a.toolRegistry.GetTool("location")
	weatherTool, _ := a.toolRegistry.GetTool("weather")

	// Location research node
	builder.AddToolNode("location_research", "Research Locations", locationTool)

	// Weather check node
	builder.AddToolNode("weather_check", "Check Weather", weatherTool)

	// Itinerary creation node
	builder.AddLLMNode(
		"create_itinerary",
		"Create Detailed Itinerary",
		provider,
		`Create a detailed day-by-day itinerary:
Destination: {{.destination}}
Dates: {{.start_date}} to {{.end_date}}
Travelers: {{.travelers}}
Interests: {{.interests}}
Weather forecast: {{.weather}}
Local attractions: {{.locations}}
Budget: {{.budget}}

Create a comprehensive itinerary with activities, timing, costs, and practical tips.`,
	)

	// Add edges
	builder.AddSimpleEdge("location_research", "weather_check")
	builder.AddSimpleEdge("weather_check", "create_itinerary")

	// Set start node
	builder.SetStartNode("location_research")

	// Build and register
	builder.BuildAndRegister(a.workflowRegistry)
}

// initializeTravelRecommendationsWorkflow creates the travel recommendations workflow
func (a *TravelAgent) initializeTravelRecommendationsWorkflow() {
	builder := workflow.NewWorkflowBuilder(
		"travel_recommendations",
		"Travel Recommendations",
		"Workflow for providing general travel advice and recommendations",
	)

	provider, _ := a.llmManager.GetProvider("")

	// Analysis node
	builder.AddLLMNode(
		"analyze_request",
		"Analyze Travel Request",
		provider,
		`Analyze the travel request and provide personalized recommendations:
Query: {{.query}}
Destination: {{.destination}}
Travel style: {{.travel_style}}
Interests: {{.interests}}
Budget: {{.budget}}

Provide comprehensive travel advice, tips, and recommendations tailored to the user's preferences.`,
	)

	// Set start node
	builder.SetStartNode("analyze_request")

	// Build and register
	builder.BuildAndRegister(a.workflowRegistry)
}

package langgraph

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// GraphBuilder provides a fluent interface for building graphs
type GraphBuilder struct {
	graph        *Graph
	currentNode  string
	stateManager StateManager
}

// NewGraphBuilder creates a new graph builder
func NewGraphBuilder(name string, stateManager StateManager) *GraphBuilder {
	graphID := uuid.New().String()
	graph := NewGraph(graphID, name, stateManager)

	return &GraphBuilder{
		graph:        graph,
		stateManager: stateManager,
	}
}

// SetDescription sets the graph description
func (b *GraphBuilder) SetDescription(description string) *GraphBuilder {
	b.graph.Description = description
	return b
}

// AddStartNode adds a start node and sets it as the entry point
func (b *GraphBuilder) AddStartNode(id, name string) *GraphBuilder {
	node := NewStartNode(id, name)
	b.graph.AddNode(node)
	b.graph.SetEntryPoint(id)
	b.currentNode = id
	return b
}

// AddEndNode adds an end node and sets it as an exit point
func (b *GraphBuilder) AddEndNode(id, name string) *GraphBuilder {
	node := NewEndNode(id, name)
	b.graph.AddNode(node)
	b.graph.AddExitPoint(id)
	return b
}

// AddLLMNode adds an LLM node
func (b *GraphBuilder) AddLLMNode(id, name, provider, model, promptTemplate, outputKey string) *GraphBuilder {
	node := NewLLMNode(id, name, provider, model, promptTemplate, outputKey)
	b.graph.AddNode(node)
	return b
}

// AddToolNode adds a tool node
func (b *GraphBuilder) AddToolNode(id, name, toolName string, inputKeys []string, outputKey string) *GraphBuilder {
	node := NewToolNode(id, name, toolName, inputKeys, outputKey)
	b.graph.AddNode(node)
	return b
}

// AddFunctionNode adds a function node
func (b *GraphBuilder) AddFunctionNode(id, name string, fn func(ctx context.Context, state *State) (*State, error)) *GraphBuilder {
	node := NewFunctionNode(id, name, fn)
	b.graph.AddNode(node)
	return b
}

// AddConditionalNode adds a conditional node
func (b *GraphBuilder) AddConditionalNode(id, name string, condition func(ctx context.Context, state *State) (bool, error)) *GraphBuilder {
	node := NewConditionalNode(id, name, condition)
	b.graph.AddNode(node)
	return b
}

// AddNode adds a custom node
func (b *GraphBuilder) AddNode(node Node) *GraphBuilder {
	b.graph.AddNode(node)
	return b
}

// ConnectTo creates an unconditional edge from the current node to the target node
func (b *GraphBuilder) ConnectTo(targetNodeID string) *GraphBuilder {
	if b.currentNode == "" {
		panic("no current node set, use From() first")
	}

	edge := NewEdge(b.currentNode, targetNodeID, fmt.Sprintf("%s -> %s", b.currentNode, targetNodeID))
	b.graph.AddEdge(edge)
	b.currentNode = targetNodeID
	return b
}

// ConnectToIf creates a conditional edge from the current node to the target node
func (b *GraphBuilder) ConnectToIf(targetNodeID string, condition Condition) *GraphBuilder {
	if b.currentNode == "" {
		panic("no current node set, use From() first")
	}

	edge := NewConditionalEdge(b.currentNode, targetNodeID, condition.GetDescription(), condition)
	b.graph.AddEdge(edge)
	return b
}

// From sets the current node for subsequent connections
func (b *GraphBuilder) From(nodeID string) *GraphBuilder {
	b.currentNode = nodeID
	return b
}

// AddEdge adds a custom edge
func (b *GraphBuilder) AddEdge(edge Edge) *GraphBuilder {
	b.graph.AddEdge(edge)
	return b
}

// Build returns the constructed graph
func (b *GraphBuilder) Build() (*Graph, error) {
	if err := b.graph.Validate(); err != nil {
		return nil, fmt.Errorf("graph validation failed: %w", err)
	}

	return b.graph, nil
}

// TravelGraphBuilder provides specialized methods for building travel-related graphs
type TravelGraphBuilder struct {
	*GraphBuilder
}

// NewTravelGraphBuilder creates a new travel graph builder
func NewTravelGraphBuilder(name string, stateManager StateManager) *TravelGraphBuilder {
	return &TravelGraphBuilder{
		GraphBuilder: NewGraphBuilder(name, stateManager),
	}
}

// AddDestinationResearchNode adds a node for destination research
func (b *TravelGraphBuilder) AddDestinationResearchNode(id, name string) *TravelGraphBuilder {
	prompt := `Research the destination: {{.destination}}
Provide information about:
- Best time to visit
- Key attractions
- Local culture and customs
- Transportation options
- Estimated costs

Destination: {{.destination}}
Travel dates: {{.start_date}} to {{.end_date}}
Number of travelers: {{.travelers}}
Budget: {{.budget}}`

	b.AddLLMNode(id, name, "ollama", "llama3.2", prompt, "destination_info")
	return b
}

// AddFlightSearchNode adds a node for flight search
func (b *TravelGraphBuilder) AddFlightSearchNode(id, name string) *TravelGraphBuilder {
	inputKeys := []string{"origin", "destination", "start_date", "end_date", "travelers"}
	b.AddToolNode(id, name, "flight_search", inputKeys, "flight_options")
	return b
}

// AddHotelSearchNode adds a node for hotel search
func (b *TravelGraphBuilder) AddHotelSearchNode(id, name string) *TravelGraphBuilder {
	inputKeys := []string{"destination", "start_date", "end_date", "travelers", "budget"}
	b.AddToolNode(id, name, "hotel_search", inputKeys, "hotel_options")
	return b
}

// AddWeatherCheckNode adds a node for weather checking
func (b *TravelGraphBuilder) AddWeatherCheckNode(id, name string) *TravelGraphBuilder {
	inputKeys := []string{"destination", "start_date", "end_date"}
	b.AddToolNode(id, name, "weather", inputKeys, "weather_forecast")
	return b
}

// AddItineraryPlanningNode adds a node for itinerary planning
func (b *TravelGraphBuilder) AddItineraryPlanningNode(id, name string) *TravelGraphBuilder {
	prompt := `Create a detailed itinerary for the trip:

Destination: {{.destination}}
Dates: {{.start_date}} to {{.end_date}}
Travelers: {{.travelers}}
Budget: {{.budget}}

Available information:
- Destination info: {{.destination_info}}
- Weather forecast: {{.weather_forecast}}
- Flight options: {{.flight_options}}
- Hotel options: {{.hotel_options}}

Create a day-by-day itinerary including:
- Daily activities and attractions
- Meal recommendations
- Transportation between locations
- Time estimates for each activity
- Budget breakdown
- Tips and recommendations`

	b.AddLLMNode(id, name, "ollama", "llama3.2", prompt, "itinerary")
	return b
}

// AddBudgetAnalysisNode adds a node for budget analysis
func (b *TravelGraphBuilder) AddBudgetAnalysisNode(id, name string) *TravelGraphBuilder {
	fn := func(ctx context.Context, state *State) (*State, error) {
		newState := state.Clone()

		// Get budget and cost information
		budget, _ := state.GetInt("budget")

		// Calculate estimated costs (simplified)
		flightCost := 500 // Default flight cost
		hotelCost := 150  // Default hotel cost per night
		dailyCost := 100  // Default daily expenses

		// Get trip duration
		// TODO: Calculate actual duration from dates
		duration := 7 // Default 7 days

		totalCost := flightCost + (hotelCost * duration) + (dailyCost * duration)

		budgetAnalysis := map[string]interface{}{
			"total_estimated_cost": totalCost,
			"budget":               budget,
			"remaining_budget":     budget - totalCost,
			"budget_sufficient":    budget >= totalCost,
			"cost_breakdown": map[string]int{
				"flights": flightCost,
				"hotels":  hotelCost * duration,
				"daily":   dailyCost * duration,
			},
		}

		newState.Set("budget_analysis", budgetAnalysis)
		return newState, nil
	}

	b.AddFunctionNode(id, name, fn)
	return b
}

// AddBudgetCheckCondition adds a condition to check if budget is sufficient
func (b *TravelGraphBuilder) AddBudgetCheckCondition() Condition {
	return NewFunctionCondition("budget is sufficient", func(ctx context.Context, state *State) (bool, error) {
		analysis, exists := state.GetMap("budget_analysis")
		if !exists {
			return false, fmt.Errorf("budget analysis not found")
		}

		sufficient, ok := analysis["budget_sufficient"].(bool)
		if !ok {
			return false, fmt.Errorf("budget_sufficient not found or not boolean")
		}

		return sufficient, nil
	})
}

// BuildCompleteTripPlanningGraph builds a complete trip planning graph
func (b *TravelGraphBuilder) BuildCompleteTripPlanningGraph() (*Graph, error) {
	// Add nodes
	b.AddStartNode("start", "Start Trip Planning")
	b.AddDestinationResearchNode("research", "Research Destination")
	b.AddWeatherCheckNode("weather", "Check Weather")
	b.AddFlightSearchNode("flights", "Search Flights")
	b.AddHotelSearchNode("hotels", "Search Hotels")
	b.AddBudgetAnalysisNode("budget", "Analyze Budget")
	b.AddItineraryPlanningNode("itinerary", "Plan Itinerary")
	b.AddEndNode("end", "Complete Planning")

	// Connect nodes with flow
	b.From("start").ConnectTo("research").
		ConnectTo("weather").
		ConnectTo("flights").
		ConnectTo("hotels").
		ConnectTo("budget")

	// Add conditional flow based on budget
	budgetCondition := b.AddBudgetCheckCondition()
	b.From("budget").
		ConnectToIf("itinerary", budgetCondition).
		ConnectToIf("end", NewNotCondition(budgetCondition))

	b.From("itinerary").ConnectTo("end")

	return b.Build()
}

// BuildFlightSearchGraph builds a focused flight search graph
func (b *TravelGraphBuilder) BuildFlightSearchGraph() (*Graph, error) {
	b.AddStartNode("start", "Start Flight Search")
	b.AddFlightSearchNode("search", "Search Flights")
	b.AddEndNode("end", "Flight Search Complete")

	b.From("start").ConnectTo("search").ConnectTo("end")

	return b.Build()
}

// BuildHotelSearchGraph builds a focused hotel search graph
func (b *TravelGraphBuilder) BuildHotelSearchGraph() (*Graph, error) {
	b.AddStartNode("start", "Start Hotel Search")
	b.AddHotelSearchNode("search", "Search Hotels")
	b.AddEndNode("end", "Hotel Search Complete")

	b.From("start").ConnectTo("search").ConnectTo("end")

	return b.Build()
}

// BuildItineraryPlanningGraph builds a focused itinerary planning graph
func (b *TravelGraphBuilder) BuildItineraryPlanningGraph() (*Graph, error) {
	b.AddStartNode("start", "Start Itinerary Planning")
	b.AddDestinationResearchNode("research", "Research Destination")
	b.AddWeatherCheckNode("weather", "Check Weather")
	b.AddItineraryPlanningNode("plan", "Create Itinerary")
	b.AddEndNode("end", "Itinerary Complete")

	b.From("start").ConnectTo("research").
		ConnectTo("weather").
		ConnectTo("plan").
		ConnectTo("end")

	return b.Build()
}

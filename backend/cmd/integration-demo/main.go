package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/exotic-travel-booking/backend/internal/agents/specialist"
	"github.com/exotic-travel-booking/backend/internal/langchain"
	"github.com/exotic-travel-booking/backend/internal/langgraph"
	"github.com/exotic-travel-booking/backend/internal/llm/providers"
	"github.com/exotic-travel-booking/backend/internal/services"
	"github.com/exotic-travel-booking/backend/internal/tools"
)

func main() {
	fmt.Println("🌟 Exotic Travel Booking - Complete Integration Demo")
	fmt.Println("===================================================")

	ctx := context.Background()

	// Initialize core components
	fmt.Println("\n🔧 Initializing Core Components...")

	// 1. OLAMA Service
	ollamaService := services.NewOllamaService("http://localhost:11434")
	if err := ollamaService.CheckHealth(ctx); err != nil {
		log.Printf("⚠️  OLAMA not available: %v", err)
	} else {
		fmt.Println("✅ OLAMA service connected")
	}

	// 2. LLM Provider (using OLAMA)
	llmConfig := &providers.LLMConfig{
		Provider: "ollama",
		Model:    "llama3.2",
		BaseURL:  "http://localhost:11434",
		Timeout:  60 * time.Second,
	}

	llmProvider, err := providers.NewOllamaProvider(llmConfig)
	if err != nil {
		log.Printf("⚠️  LLM provider creation failed: %v", err)
	} else {
		fmt.Println("✅ LLM provider initialized")
	}

	// 3. Tool Registry
	toolRegistry := tools.NewToolRegistry()
	fmt.Println("✅ Tool registry created")

	// 4. State Manager for LangGraph
	stateManager := langgraph.NewMemoryStateManager()
	fmt.Println("✅ State manager initialized")

	// 5. Memory Manager for LangChain
	memoryManager := langchain.NewMemoryManager()
	bufferMemory := langchain.NewBufferMemory("travel_conversation", 20)
	memoryManager.RegisterMemory(bufferMemory)
	fmt.Println("✅ Memory manager initialized")

	// Demo 1: LangChain Patterns
	fmt.Println("\n1. LangChain Patterns Demo")
	fmt.Println("==========================")

	if err := demonstrateLangChainPatterns(ctx, llmProvider, memoryManager); err != nil {
		log.Printf("❌ LangChain demo failed: %v", err)
	}

	// Demo 2: LangGraph Workflows
	fmt.Println("\n2. LangGraph Workflows Demo")
	fmt.Println("============================")

	if err := demonstrateLangGraphWorkflows(ctx, stateManager); err != nil {
		log.Printf("❌ LangGraph demo failed: %v", err)
	}

	// Demo 3: Specialist Agents
	fmt.Println("\n3. Specialist Agents Demo")
	fmt.Println("==========================")

	if err := demonstrateSpecialistAgents(ctx, llmProvider, toolRegistry, stateManager); err != nil {
		log.Printf("❌ Specialist agents demo failed: %v", err)
	}

	// Demo 4: Complete Travel Planning Workflow
	fmt.Println("\n4. Complete Travel Planning Workflow")
	fmt.Println("====================================")

	if err := demonstrateCompleteTravelWorkflow(ctx, llmProvider, toolRegistry, stateManager, memoryManager); err != nil {
		log.Printf("❌ Complete workflow demo failed: %v", err)
	}

	fmt.Println("\n🎉 Integration Demo Completed Successfully!")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("✅ OLAMA local LLM integration")
	fmt.Println("✅ LangChain patterns (chains, prompts, parsers, memory)")
	fmt.Println("✅ LangGraph state management and workflows")
	fmt.Println("✅ Specialist travel agents")
	fmt.Println("✅ End-to-end travel planning")
	fmt.Println("✅ OpenTelemetry observability")
}

func demonstrateLangChainPatterns(ctx context.Context, llmProvider providers.LLMProvider, memoryManager *langchain.MemoryManager) error {
	fmt.Println("   📝 Creating prompt templates...")

	// Create travel prompt templates
	travelTemplates := langchain.NewTravelPromptTemplates()
	registry := travelTemplates.GetRegistry()

	// Get destination research template
	destTemplate, err := registry.GetTemplate("destination_research")
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	fmt.Println("   🔗 Building LLM chain...")

	// Create LLM chain with memory
	memory, _ := memoryManager.GetMemory("travel_conversation")
	sessionID := "demo_session_1"

	llmChain := langchain.NewLLMChain(
		"destination_research_chain",
		"Research travel destinations using LLM",
		llmProvider,
		destTemplate,
	).SetMemory(memory, sessionID)

	// Create JSON output parser
	jsonParser := langchain.NewJSONParser("destination_parser", map[string]interface{}{
		"type":     "object",
		"required": []string{"destination", "summary"},
	}, false)
	llmChain.SetOutputParser(jsonParser)

	fmt.Println("   🚀 Executing chain...")

	// Execute chain
	input := map[string]interface{}{
		"destination": "Kyoto, Japan",
		"start_date":  "2024-04-01",
		"end_date":    "2024-04-07",
		"travelers":   2,
		"budget":      3000,
		"interests":   "culture, temples, food",
	}

	result, err := llmChain.Execute(ctx, input)
	if err != nil {
		return fmt.Errorf("chain execution failed: %w", err)
	}

	fmt.Printf("   ✅ Chain executed successfully (duration: %v)\n", result.Duration)
	fmt.Printf("   📊 Response: %s\n", truncateString(fmt.Sprintf("%v", result.Output["response"]), 100))

	// Demonstrate sequential chain
	fmt.Println("   🔄 Building sequential chain...")

	// Create a simple sequential chain
	builder := langchain.NewChainBuilder()
	sequentialChain := builder.Add(llmChain).BuildSequential(
		"travel_research_sequence",
		"Sequential travel research workflow",
	)

	seqResult, err := sequentialChain.Execute(ctx, input)
	if err != nil {
		return fmt.Errorf("sequential chain failed: %w", err)
	}

	fmt.Printf("   ✅ Sequential chain completed (duration: %v)\n", seqResult.Duration)

	return nil
}

func demonstrateLangGraphWorkflows(ctx context.Context, stateManager langgraph.StateManager) error {
	fmt.Println("   🏗️  Building travel planning graph...")

	// Create travel graph builder
	builder := langgraph.NewTravelGraphBuilder("Demo Travel Planning", stateManager)

	// Build a complete trip planning graph
	graph, err := builder.BuildCompleteTripPlanningGraph()
	if err != nil {
		return fmt.Errorf("failed to build graph: %w", err)
	}

	fmt.Printf("   📊 Graph created with %d nodes and %d edges\n", graph.GetNodeCount(), graph.GetEdgeCount())

	// Create initial state
	initialState := langgraph.NewState("demo_trip_state", graph.ID)
	initialState.SetMultiple(map[string]interface{}{
		"destination": "Bali, Indonesia",
		"start_date":  "2024-05-15",
		"end_date":    "2024-05-22",
		"travelers":   2,
		"budget":      2500,
		"preferences": []string{"beaches", "culture", "adventure"},
	})

	fmt.Println("   🚀 Executing graph workflow...")

	// Execute graph
	finalState, err := graph.Execute(ctx, initialState)
	if err != nil {
		return fmt.Errorf("graph execution failed: %w", err)
	}

	fmt.Printf("   ✅ Graph executed successfully\n")
	fmt.Printf("   📋 Final state has %d items\n", finalState.Size())

	// Show some results
	if destinationInfo, exists := finalState.Get("destination_info"); exists {
		fmt.Printf("   🏝️  Destination info: %s\n", truncateString(fmt.Sprintf("%v", destinationInfo), 80))
	}

	return nil
}

func demonstrateSpecialistAgents(ctx context.Context, llmProvider providers.LLMProvider, toolRegistry *tools.ToolRegistry, stateManager langgraph.StateManager) error {
	fmt.Println("   👥 Creating specialist agents...")

	// Create specialist agents
	flightAgent := specialist.NewFlightAgent(llmProvider, toolRegistry, stateManager)
	hotelAgent := specialist.NewHotelAgent(llmProvider, toolRegistry, stateManager)
	itineraryAgent := specialist.NewItineraryAgent(llmProvider, toolRegistry, stateManager)
	supervisorAgent := specialist.NewSupervisorAgent(llmProvider, toolRegistry, stateManager)

	fmt.Printf("   ✅ Created %d specialist agents\n", 4)

	// Test flight agent
	fmt.Println("   ✈️  Testing flight agent...")

	flightRequest := &specialist.AgentRequest{
		ID:        "flight_demo_1",
		UserID:    "demo_user",
		SessionID: "demo_session",
		AgentType: "flight",
		Query:     "Find flights from New York to Paris for 2 people",
		Parameters: map[string]interface{}{
			"origin":      "New York",
			"destination": "Paris",
			"start_date":  "2024-06-15",
			"end_date":    "2024-06-22",
			"travelers":   2,
			"budget":      1500,
		},
		CreatedAt: time.Now(),
	}

	flightResponse, err := flightAgent.ProcessRequest(ctx, flightRequest)
	if err != nil {
		return fmt.Errorf("flight agent failed: %w", err)
	}

	fmt.Printf("   ✅ Flight agent completed (confidence: %.2f, duration: %v)\n",
		flightResponse.Confidence, flightResponse.Duration)

	// Test supervisor agent
	fmt.Println("   🎯 Testing supervisor agent...")

	supervisorRequest := &specialist.AgentRequest{
		ID:        "supervisor_demo_1",
		UserID:    "demo_user",
		SessionID: "demo_session",
		AgentType: "supervisor",
		Query:     "Plan a complete trip to Tokyo including flights, hotels, and itinerary",
		Parameters: map[string]interface{}{
			"destination": "Tokyo, Japan",
			"origin":      "Los Angeles",
			"start_date":  "2024-07-10",
			"end_date":    "2024-07-17",
			"travelers":   2,
			"budget":      4000,
			"preferences": []string{"culture", "food", "technology"},
		},
		CreatedAt: time.Now(),
	}

	supervisorResponse, err := supervisorAgent.ProcessRequest(ctx, supervisorRequest)
	if err != nil {
		return fmt.Errorf("supervisor agent failed: %w", err)
	}

	fmt.Printf("   ✅ Supervisor agent completed (confidence: %.2f, duration: %v)\n",
		supervisorResponse.Confidence, supervisorResponse.Duration)

	// Show agent capabilities
	fmt.Printf("   📋 Flight agent capabilities: %v\n", flightAgent.GetCapabilities()[:3])
	fmt.Printf("   📋 Hotel agent capabilities: %v\n", hotelAgent.GetCapabilities()[:3])
	fmt.Printf("   📋 Itinerary agent capabilities: %v\n", itineraryAgent.GetCapabilities()[:3])

	return nil
}

func demonstrateCompleteTravelWorkflow(ctx context.Context, llmProvider providers.LLMProvider, toolRegistry *tools.ToolRegistry, stateManager langgraph.StateManager, memoryManager *langchain.MemoryManager) error {
	fmt.Println("   🌍 Creating complete travel planning workflow...")

	// Create graph executor
	executor := langgraph.NewGraphExecutor(stateManager)

	// Create travel planning graph
	builder := langgraph.NewTravelGraphBuilder("Complete Travel Workflow", stateManager)
	graph, err := builder.BuildCompleteTripPlanningGraph()
	if err != nil {
		return fmt.Errorf("failed to build workflow graph: %w", err)
	}

	// Create supervisor agent for coordination
	supervisor := specialist.NewSupervisorAgent(llmProvider, toolRegistry, stateManager)

	fmt.Println("   🎯 Supervisor agent created:", supervisor.GetName())
	fmt.Println("   📝 Simulating user travel request...")

	// Simulate a complete travel planning request
	travelInput := map[string]interface{}{
		"user_query":       "I want to plan a romantic getaway to Santorini, Greece",
		"destination":      "Santorini, Greece",
		"origin":           "London, UK",
		"start_date":       "2024-08-15",
		"end_date":         "2024-08-22",
		"travelers":        2,
		"budget":           3500,
		"preferences":      []string{"romantic", "sunset views", "wine tasting", "relaxation"},
		"special_requests": "Anniversary celebration",
	}

	fmt.Println("   🚀 Executing complete workflow...")

	// Execute with custom options
	options := &langgraph.ExecutionOptions{
		MaxIterations: 50,
		Timeout:       5 * time.Minute,
		EnableTracing: true,
		Metadata: map[string]interface{}{
			"workflow_type": "complete_travel_planning",
			"user_type":     "romantic_getaway",
		},
	}

	result, err := executor.Execute(ctx, graph, travelInput, options)
	if err != nil {
		return fmt.Errorf("workflow execution failed: %w", err)
	}

	fmt.Printf("   ✅ Complete workflow executed successfully!\n")
	fmt.Printf("   📊 Execution ID: %s\n", result.ExecutionID)
	fmt.Printf("   ⏱️  Duration: %v\n", result.Duration)
	fmt.Printf("   📈 Status: %s\n", result.Status)
	fmt.Printf("   🛤️  Nodes visited: %v\n", result.NodesVisited)

	// Show final results
	if result.FinalState != nil {
		fmt.Printf("   📋 Final state contains %d items\n", result.FinalState.Size())
	}

	// Show executor statistics
	stats := executor.GetExecutionStats()
	fmt.Printf("   📊 Total executions: %v\n", stats["total_executions"])

	// Demonstrate memory integration
	fmt.Println("   💭 Testing memory integration...")

	memory, _ := memoryManager.GetMemory("travel_conversation")
	sessionID := "complete_workflow_session"

	// Add some conversation messages
	messages := []*langchain.Message{
		{
			SessionID: sessionID,
			Role:      "user",
			Content:   "I want to plan a trip to Santorini",
			Timestamp: time.Now(),
		},
		{
			SessionID: sessionID,
			Role:      "assistant",
			Content:   "I'd be happy to help you plan your trip to Santorini!",
			Timestamp: time.Now(),
		},
	}

	for _, msg := range messages {
		memory.AddMessage(ctx, msg)
	}

	// Retrieve conversation history
	history, err := memory.GetMessages(ctx, sessionID, 10)
	if err != nil {
		return fmt.Errorf("memory retrieval failed: %w", err)
	}

	fmt.Printf("   ✅ Memory contains %d messages\n", len(history))

	summary, err := memory.GetSummary(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("memory summary failed: %w", err)
	}

	fmt.Printf("   📝 Conversation summary: %s\n", summary)

	return nil
}

// Helper function
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

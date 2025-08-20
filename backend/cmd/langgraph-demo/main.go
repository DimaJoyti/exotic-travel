package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/exotic-travel-booking/backend/internal/langgraph"
)

func main() {
	fmt.Println("🔗 LangGraph Integration Demo for Exotic Travel Booking")
	fmt.Println("======================================================")

	ctx := context.Background()

	// Create state manager
	stateManager := langgraph.NewMemoryStateManager()

	// Demo 1: Basic Graph Execution
	fmt.Println("\n1. Basic Graph Execution Demo")
	fmt.Println("-----------------------------")
	
	if err := basicGraphDemo(ctx, stateManager); err != nil {
		log.Printf("❌ Basic graph demo failed: %v", err)
	}

	// Demo 2: Conditional Graph
	fmt.Println("\n2. Conditional Graph Demo")
	fmt.Println("-------------------------")
	
	if err := conditionalGraphDemo(ctx, stateManager); err != nil {
		log.Printf("❌ Conditional graph demo failed: %v", err)
	}

	// Demo 3: Travel Planning Graph
	fmt.Println("\n3. Travel Planning Graph Demo")
	fmt.Println("-----------------------------")
	
	if err := travelPlanningDemo(ctx, stateManager); err != nil {
		log.Printf("❌ Travel planning demo failed: %v", err)
	}

	// Demo 4: Graph Executor
	fmt.Println("\n4. Graph Executor Demo")
	fmt.Println("----------------------")
	
	if err := executorDemo(ctx, stateManager); err != nil {
		log.Printf("❌ Executor demo failed: %v", err)
	}

	// Demo 5: Async Execution
	fmt.Println("\n5. Async Execution Demo")
	fmt.Println("-----------------------")
	
	if err := asyncExecutionDemo(ctx, stateManager); err != nil {
		log.Printf("❌ Async execution demo failed: %v", err)
	}

	fmt.Println("\n🎉 LangGraph integration demo completed successfully!")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("✅ State management with persistence")
	fmt.Println("✅ Graph-based agent execution")
	fmt.Println("✅ Conditional routing and loops")
	fmt.Println("✅ Node orchestration")
	fmt.Println("✅ Travel-specific workflows")
	fmt.Println("✅ Async execution with monitoring")
}

func basicGraphDemo(ctx context.Context, stateManager langgraph.StateManager) error {
	// Create a simple linear graph
	builder := langgraph.NewGraphBuilder("Basic Demo Graph", stateManager)
	
	graph, err := builder.
		AddStartNode("start", "Start Processing").
		AddFunctionNode("step1", "Step 1", func(ctx context.Context, state *langgraph.State) (*langgraph.State, error) {
			newState := state.Clone()
			newState.Set("step1_completed", true)
			newState.Set("message", "Step 1 executed")
			fmt.Println("   📍 Executing Step 1")
			return newState, nil
		}).
		AddFunctionNode("step2", "Step 2", func(ctx context.Context, state *langgraph.State) (*langgraph.State, error) {
			newState := state.Clone()
			newState.Set("step2_completed", true)
			message, _ := newState.GetString("message")
			newState.Set("message", message+" -> Step 2 executed")
			fmt.Println("   📍 Executing Step 2")
			return newState, nil
		}).
		AddEndNode("end", "End Processing").
		From("start").ConnectTo("step1").ConnectTo("step2").ConnectTo("end").
		Build()

	if err != nil {
		return fmt.Errorf("failed to build graph: %w", err)
	}

	// Execute the graph
	initialState := langgraph.NewState("demo-state-1", graph.ID)
	initialState.Set("input", "Hello LangGraph!")

	finalState, err := graph.Execute(ctx, initialState)
	if err != nil {
		return fmt.Errorf("graph execution failed: %w", err)
	}

	// Display results
	message, _ := finalState.GetString("message")
	fmt.Printf("   ✅ Final message: %s\n", message)
	fmt.Printf("   📊 State size: %d items\n", finalState.Size())

	return nil
}

func conditionalGraphDemo(ctx context.Context, stateManager langgraph.StateManager) error {
	// Create a graph with conditional routing
	builder := langgraph.NewGraphBuilder("Conditional Demo Graph", stateManager)
	
	// Add nodes
	builder.AddStartNode("start", "Start")
	builder.AddFunctionNode("check_budget", "Check Budget", func(ctx context.Context, state *langgraph.State) (*langgraph.State, error) {
		newState := state.Clone()
		budget, _ := newState.GetInt("budget")
		newState.Set("budget_sufficient", budget >= 1000)
		fmt.Printf("   💰 Budget check: $%d (sufficient: %t)\n", budget, budget >= 1000)
		return newState, nil
	})
	builder.AddFunctionNode("luxury_plan", "Create Luxury Plan", func(ctx context.Context, state *langgraph.State) (*langgraph.State, error) {
		newState := state.Clone()
		newState.Set("plan_type", "luxury")
		newState.Set("plan", "5-star hotels, first-class flights, private tours")
		fmt.Println("   🏨 Creating luxury travel plan")
		return newState, nil
	})
	builder.AddFunctionNode("budget_plan", "Create Budget Plan", func(ctx context.Context, state *langgraph.State) (*langgraph.State, error) {
		newState := state.Clone()
		newState.Set("plan_type", "budget")
		newState.Set("plan", "3-star hotels, economy flights, group tours")
		fmt.Println("   🎒 Creating budget travel plan")
		return newState, nil
	})
	builder.AddEndNode("end", "End")

	// Add conditional edges
	budgetCondition := langgraph.NewStateValueCondition("budget_sufficient", true, "equals")
	noBudgetCondition := langgraph.NewStateValueCondition("budget_sufficient", false, "equals")

	builder.From("start").ConnectTo("check_budget")
	builder.From("check_budget").ConnectToIf("luxury_plan", budgetCondition)
	builder.From("check_budget").ConnectToIf("budget_plan", noBudgetCondition)
	builder.From("luxury_plan").ConnectTo("end")
	builder.From("budget_plan").ConnectTo("end")

	graph, err := builder.Build()
	if err != nil {
		return fmt.Errorf("failed to build conditional graph: %w", err)
	}

	// Test with high budget
	fmt.Println("   🧪 Testing with high budget ($2000)")
	initialState := langgraph.NewState("demo-state-2a", graph.ID)
	initialState.Set("budget", 2000)

	finalState, err := graph.Execute(ctx, initialState)
	if err != nil {
		return fmt.Errorf("high budget execution failed: %w", err)
	}

	planType, _ := finalState.GetString("plan_type")
	plan, _ := finalState.GetString("plan")
	fmt.Printf("   ✅ Plan type: %s\n", planType)
	fmt.Printf("   📋 Plan: %s\n", plan)

	// Test with low budget
	fmt.Println("\n   🧪 Testing with low budget ($500)")
	initialState2 := langgraph.NewState("demo-state-2b", graph.ID)
	initialState2.Set("budget", 500)

	finalState2, err := graph.Execute(ctx, initialState2)
	if err != nil {
		return fmt.Errorf("low budget execution failed: %w", err)
	}

	planType2, _ := finalState2.GetString("plan_type")
	plan2, _ := finalState2.GetString("plan")
	fmt.Printf("   ✅ Plan type: %s\n", planType2)
	fmt.Printf("   📋 Plan: %s\n", plan2)

	return nil
}

func travelPlanningDemo(ctx context.Context, stateManager langgraph.StateManager) error {
	// Create a travel-specific graph using the travel builder
	builder := langgraph.NewTravelGraphBuilder("Travel Planning Demo", stateManager)
	
	// Build a simplified travel planning graph
	builder.AddStartNode("start", "Start Travel Planning")
	builder.AddDestinationResearchNode("research", "Research Destination")
	builder.AddWeatherCheckNode("weather", "Check Weather")
	builder.AddBudgetAnalysisNode("budget", "Analyze Budget")
	builder.AddEndNode("end", "Complete Planning")

	// Connect the nodes
	builder.From("start").ConnectTo("research").ConnectTo("weather").ConnectTo("budget").ConnectTo("end")

	graph, err := builder.Build()
	if err != nil {
		return fmt.Errorf("failed to build travel graph: %w", err)
	}

	// Execute with travel data
	initialState := langgraph.NewState("travel-demo-state", graph.ID)
	initialState.SetMultiple(map[string]interface{}{
		"destination": "Tokyo, Japan",
		"start_date":  "2024-06-15",
		"end_date":    "2024-06-22",
		"travelers":   2,
		"budget":      3000,
	})

	fmt.Println("   🗾 Planning trip to Tokyo, Japan")
	fmt.Println("   📅 Dates: June 15-22, 2024")
	fmt.Println("   👥 Travelers: 2")
	fmt.Println("   💰 Budget: $3,000")

	finalState, err := graph.Execute(ctx, initialState)
	if err != nil {
		return fmt.Errorf("travel planning execution failed: %w", err)
	}

	// Display results
	destinationInfo, _ := finalState.GetString("destination_info")
	weatherForecast, _ := finalState.GetString("weather_forecast")
	budgetAnalysis, _ := finalState.GetMap("budget_analysis")

	fmt.Printf("   ✅ Destination research completed: %s\n", destinationInfo)
	fmt.Printf("   🌤️  Weather forecast: %s\n", weatherForecast)
	fmt.Printf("   💹 Budget analysis: sufficient = %v\n", budgetAnalysis["budget_sufficient"])

	return nil
}

func executorDemo(ctx context.Context, stateManager langgraph.StateManager) error {
	// Create graph executor
	executor := langgraph.NewGraphExecutor(stateManager)

	// Create a simple graph
	builder := langgraph.NewGraphBuilder("Executor Demo Graph", stateManager)
	graph, err := builder.
		AddStartNode("start", "Start").
		AddFunctionNode("process", "Process Data", func(ctx context.Context, state *langgraph.State) (*langgraph.State, error) {
			newState := state.Clone()
			counter, _ := newState.GetInt("counter")
			newState.Set("counter", counter+1)
			newState.Set("processed_at", time.Now().Format(time.RFC3339))
			fmt.Printf("   🔄 Processing... counter = %d\n", counter+1)
			return newState, nil
		}).
		AddEndNode("end", "End").
		From("start").ConnectTo("process").ConnectTo("end").
		Build()

	if err != nil {
		return fmt.Errorf("failed to build executor demo graph: %w", err)
	}

	// Execute with custom options
	input := map[string]interface{}{
		"counter": 0,
		"data":    "test data",
	}

	options := &langgraph.ExecutionOptions{
		MaxIterations: 50,
		Timeout:       30 * time.Second,
		EnableTracing: true,
		Metadata: map[string]interface{}{
			"demo": "executor",
		},
	}

	result, err := executor.Execute(ctx, graph, input, options)
	if err != nil {
		return fmt.Errorf("executor execution failed: %w", err)
	}

	// Display execution results
	fmt.Printf("   ✅ Execution ID: %s\n", result.ExecutionID)
	fmt.Printf("   ⏱️  Duration: %v\n", result.Duration)
	fmt.Printf("   📊 Status: %s\n", result.Status)
	fmt.Printf("   🛤️  Nodes visited: %v\n", result.NodesVisited)

	counter, _ := result.FinalState.GetInt("counter")
	fmt.Printf("   🔢 Final counter: %d\n", counter)

	return nil
}

func asyncExecutionDemo(ctx context.Context, stateManager langgraph.StateManager) error {
	// Create graph executor
	executor := langgraph.NewGraphExecutor(stateManager)

	// Create a graph with some delay
	builder := langgraph.NewGraphBuilder("Async Demo Graph", stateManager)
	graph, err := builder.
		AddStartNode("start", "Start Async").
		AddFunctionNode("long_task", "Long Running Task", func(ctx context.Context, state *langgraph.State) (*langgraph.State, error) {
			fmt.Println("   ⏳ Starting long running task...")
			time.Sleep(1 * time.Second) // Simulate work
			newState := state.Clone()
			newState.Set("task_completed", true)
			newState.Set("result", "Long task completed successfully")
			fmt.Println("   ✅ Long running task completed")
			return newState, nil
		}).
		AddEndNode("end", "End Async").
		From("start").ConnectTo("long_task").ConnectTo("end").
		Build()

	if err != nil {
		return fmt.Errorf("failed to build async demo graph: %w", err)
	}

	// Execute asynchronously
	input := map[string]interface{}{
		"async_test": true,
	}

	executionID, resultChan, err := executor.ExecuteAsync(ctx, graph, input, nil)
	if err != nil {
		return fmt.Errorf("failed to start async execution: %w", err)
	}

	fmt.Printf("   🚀 Started async execution: %s\n", executionID)
	fmt.Println("   ⏳ Waiting for completion...")

	// Wait for result with timeout
	select {
	case result := <-resultChan:
		fmt.Printf("   ✅ Async execution completed: %s\n", result.Status)
		fmt.Printf("   ⏱️  Duration: %v\n", result.Duration)
		
		taskResult, _ := result.FinalState.GetString("result")
		fmt.Printf("   📋 Result: %s\n", taskResult)

	case <-time.After(10 * time.Second):
		return fmt.Errorf("async execution timed out")
	}

	// Show executor stats
	stats := executor.GetExecutionStats()
	fmt.Printf("   📊 Executor stats: %v\n", stats)

	return nil
}

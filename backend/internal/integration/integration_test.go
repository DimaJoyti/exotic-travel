package integration

import (
	"context"
	"testing"
	"time"

	"github.com/exotic-travel-booking/backend/internal/agents/specialist"
	"github.com/exotic-travel-booking/backend/internal/langchain"
	"github.com/exotic-travel-booking/backend/internal/langgraph"
	"github.com/exotic-travel-booking/backend/internal/llm/providers"
	"github.com/exotic-travel-booking/backend/internal/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompleteIntegration(t *testing.T) {
	ctx := context.Background()

	// Setup components
	stateManager := langgraph.NewMemoryStateManager()
	toolRegistry := tools.NewToolRegistry()
	
	// Mock LLM provider for testing
	llmConfig := &providers.LLMConfig{
		Provider: "mock",
		Model:    "test-model",
		BaseURL:  "http://localhost:11434",
		Timeout:  30 * time.Second,
	}

	// Create a mock provider that doesn't require actual OLAMA
	llmProvider := &MockLLMProvider{config: llmConfig}

	t.Run("LangChain Integration", func(t *testing.T) {
		testLangChainIntegration(t, ctx, llmProvider)
	})

	t.Run("LangGraph Integration", func(t *testing.T) {
		testLangGraphIntegration(t, ctx, stateManager)
	})

	t.Run("Specialist Agents Integration", func(t *testing.T) {
		testSpecialistAgentsIntegration(t, ctx, llmProvider, toolRegistry, stateManager)
	})

	t.Run("End-to-End Workflow", func(t *testing.T) {
		testEndToEndWorkflow(t, ctx, llmProvider, toolRegistry, stateManager)
	})
}

func testLangChainIntegration(t *testing.T, ctx context.Context, llmProvider providers.LLMProvider) {
	// Test prompt templates
	t.Run("Prompt Templates", func(t *testing.T) {
		travelTemplates := langchain.NewTravelPromptTemplates()
		registry := travelTemplates.GetRegistry()

		// Test template retrieval
		template, err := registry.GetTemplate("destination_research")
		require.NoError(t, err)
		assert.NotNil(t, template)
		assert.Equal(t, "destination_research", template.Name)

		// Test template rendering
		vars := map[string]interface{}{
			"destination": "Tokyo",
			"start_date":  "2024-06-01",
			"end_date":    "2024-06-07",
			"travelers":   2,
			"budget":      3000,
			"interests":   "culture, food",
		}

		rendered, err := template.Render(ctx, vars)
		require.NoError(t, err)
		assert.Contains(t, rendered, "Tokyo")
		assert.Contains(t, rendered, "2024-06-01")
	})

	// Test output parsers
	t.Run("Output Parsers", func(t *testing.T) {
		// JSON parser
		jsonParser := langchain.NewJSONParser("test_json", nil, false)
		jsonOutput := `{"destination": "Paris", "budget": 2000}`
		
		result, err := jsonParser.Parse(ctx, jsonOutput)
		require.NoError(t, err)
		
		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Paris", resultMap["destination"])

		// List parser
		listParser := langchain.NewListParser("test_list", "\n", true, true)
		listOutput := "1. First item\n2. Second item\n3. Third item"
		
		listResult, err := listParser.Parse(ctx, listOutput)
		require.NoError(t, err)
		
		items, ok := listResult.([]string)
		require.True(t, ok)
		assert.Len(t, items, 3)
		assert.Equal(t, "First item", items[0])
	})

	// Test memory
	t.Run("Memory", func(t *testing.T) {
		memory := langchain.NewBufferMemory("test_memory", 10)
		sessionID := "test_session"

		// Add messages
		message1 := &langchain.Message{
			SessionID: sessionID,
			Role:      "user",
			Content:   "Hello",
			Timestamp: time.Now(),
		}

		message2 := &langchain.Message{
			SessionID: sessionID,
			Role:      "assistant",
			Content:   "Hi there!",
			Timestamp: time.Now(),
		}

		err := memory.AddMessage(ctx, message1)
		require.NoError(t, err)

		err = memory.AddMessage(ctx, message2)
		require.NoError(t, err)

		// Retrieve messages
		messages, err := memory.GetMessages(ctx, sessionID, 10)
		require.NoError(t, err)
		assert.Len(t, messages, 2)
		assert.Equal(t, "Hello", messages[0].Content)
		assert.Equal(t, "Hi there!", messages[1].Content)
	})

	// Test chains
	t.Run("Chains", func(t *testing.T) {
		// Create a simple LLM chain
		template := langchain.NewPromptTemplate(
			"test_template",
			"Answer this question: {{.question}}",
			[]string{"question"},
		)

		chain := langchain.NewLLMChain(
			"test_chain",
			"Test LLM chain",
			llmProvider,
			template,
		)

		input := map[string]interface{}{
			"question": "What is the capital of France?",
		}

		result, err := chain.Execute(ctx, input)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.NotEmpty(t, result.Output["response"])
	})
}

func testLangGraphIntegration(t *testing.T, ctx context.Context, stateManager langgraph.StateManager) {
	t.Run("State Management", func(t *testing.T) {
		// Test state operations
		state := langgraph.NewState("test_state", "test_graph")
		
		// Set and get values
		state.Set("key1", "value1")
		state.Set("key2", 42)
		state.Set("key3", true)

		value1, exists := state.GetString("key1")
		assert.True(t, exists)
		assert.Equal(t, "value1", value1)

		value2, exists := state.GetInt("key2")
		assert.True(t, exists)
		assert.Equal(t, 42, value2)

		value3, exists := state.GetBool("key3")
		assert.True(t, exists)
		assert.True(t, value3)

		// Test state manager
		err := stateManager.SaveState(ctx, state)
		require.NoError(t, err)

		loadedState, err := stateManager.LoadState(ctx, "test_state")
		require.NoError(t, err)
		assert.Equal(t, state.ID, loadedState.ID)
	})

	t.Run("Graph Execution", func(t *testing.T) {
		// Create a simple graph
		graph := langgraph.NewGraph("test_graph", "Test Graph", stateManager)

		// Add nodes
		startNode := langgraph.NewStartNode("start", "Start")
		endNode := langgraph.NewEndNode("end", "End")

		err := graph.AddNode(startNode)
		require.NoError(t, err)

		err = graph.AddNode(endNode)
		require.NoError(t, err)

		// Set entry and exit points
		err = graph.SetEntryPoint("start")
		require.NoError(t, err)

		err = graph.AddExitPoint("end")
		require.NoError(t, err)

		// Add edge
		edge := langgraph.NewEdge("start", "end", "start to end")
		err = graph.AddEdge(edge)
		require.NoError(t, err)

		// Execute graph
		initialState := langgraph.NewState("test_execution", "test_graph")
		finalState, err := graph.Execute(ctx, initialState)
		require.NoError(t, err)
		assert.NotNil(t, finalState)
	})

	t.Run("Graph Builder", func(t *testing.T) {
		builder := langgraph.NewGraphBuilder("Test Builder Graph", stateManager)

		graph, err := builder.
			AddStartNode("start", "Start").
			AddFunctionNode("process", "Process", func(ctx context.Context, state *langgraph.State) (*langgraph.State, error) {
				newState := state.Clone()
				newState.Set("processed", true)
				return newState, nil
			}).
			AddEndNode("end", "End").
			From("start").ConnectTo("process").ConnectTo("end").
			Build()

		require.NoError(t, err)
		assert.Equal(t, 3, graph.GetNodeCount())
		assert.Equal(t, 2, graph.GetEdgeCount())
	})
}

func testSpecialistAgentsIntegration(t *testing.T, ctx context.Context, llmProvider providers.LLMProvider, toolRegistry *tools.ToolRegistry, stateManager langgraph.StateManager) {
	t.Run("Flight Agent", func(t *testing.T) {
		agent := specialist.NewFlightAgent(llmProvider, toolRegistry, stateManager)

		request := &specialist.AgentRequest{
			ID:        "test_flight_1",
			UserID:    "test_user",
			SessionID: "test_session",
			AgentType: "flight",
			Query:     "Find flights from NYC to LAX",
			Parameters: map[string]interface{}{
				"origin":      "NYC",
				"destination": "LAX",
				"start_date":  "2024-07-01",
				"travelers":   2,
				"budget":      800,
			},
			CreatedAt: time.Now(),
		}

		response, err := agent.ProcessRequest(ctx, request)
		require.NoError(t, err)
		assert.Equal(t, "success", response.Status)
		assert.Greater(t, response.Confidence, 0.0)
		assert.NotEmpty(t, response.Result)
	})

	t.Run("Hotel Agent", func(t *testing.T) {
		agent := specialist.NewHotelAgent(llmProvider, toolRegistry, stateManager)

		request := &specialist.AgentRequest{
			ID:        "test_hotel_1",
			UserID:    "test_user",
			SessionID: "test_session",
			AgentType: "hotel",
			Query:     "Find hotels in Paris",
			Parameters: map[string]interface{}{
				"destination": "Paris",
				"start_date":  "2024-07-01",
				"end_date":    "2024-07-05",
				"travelers":   2,
				"budget":      200,
			},
			CreatedAt: time.Now(),
		}

		response, err := agent.ProcessRequest(ctx, request)
		require.NoError(t, err)
		assert.Equal(t, "success", response.Status)
		assert.Greater(t, response.Confidence, 0.0)
	})

	t.Run("Supervisor Agent", func(t *testing.T) {
		agent := specialist.NewSupervisorAgent(llmProvider, toolRegistry, stateManager)

		request := &specialist.AgentRequest{
			ID:        "test_supervisor_1",
			UserID:    "test_user",
			SessionID: "test_session",
			AgentType: "supervisor",
			Query:     "Plan a complete trip to Rome",
			Parameters: map[string]interface{}{
				"destination": "Rome",
				"origin":      "New York",
				"start_date":  "2024-08-01",
				"end_date":    "2024-08-07",
				"travelers":   2,
				"budget":      3000,
			},
			CreatedAt: time.Now(),
		}

		response, err := agent.ProcessRequest(ctx, request)
		require.NoError(t, err)
		assert.Equal(t, "success", response.Status)
		assert.Greater(t, response.Confidence, 0.0)
	})
}

func testEndToEndWorkflow(t *testing.T, ctx context.Context, llmProvider providers.LLMProvider, toolRegistry *tools.ToolRegistry, stateManager langgraph.StateManager) {
	t.Run("Complete Travel Planning", func(t *testing.T) {
		// Create executor
		executor := langgraph.NewGraphExecutor(stateManager)

		// Create travel planning graph
		builder := langgraph.NewTravelGraphBuilder("E2E Test Graph", stateManager)
		graph, err := builder.BuildFlightSearchGraph()
		require.NoError(t, err)

		// Execute workflow
		input := map[string]interface{}{
			"origin":      "Boston",
			"destination": "London",
			"start_date":  "2024-09-01",
			"end_date":    "2024-09-08",
			"travelers":   2,
			"budget":      2000,
		}

		options := &langgraph.ExecutionOptions{
			MaxIterations: 20,
			Timeout:       30 * time.Second,
			EnableTracing: true,
		}

		result, err := executor.Execute(ctx, graph, input, options)
		require.NoError(t, err)
		assert.Equal(t, "completed", result.Status)
		assert.NotEmpty(t, result.ExecutionID)
		assert.Greater(t, result.Duration, time.Duration(0))
	})
}

// MockLLMProvider for testing
type MockLLMProvider struct {
	config *providers.LLMConfig
}

func (m *MockLLMProvider) GetName() string {
	return "mock"
}

func (m *MockLLMProvider) GenerateResponse(ctx context.Context, req *providers.GenerateRequest) (*providers.GenerateResponse, error) {
	// Return a mock response
	return &providers.GenerateResponse{
		Choices: []providers.Choice{
			{
				Message: providers.Message{
					Role:    "assistant",
					Content: "Mock response for testing",
				},
			},
		},
		Usage: providers.Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}, nil
}

func (m *MockLLMProvider) StreamResponse(ctx context.Context, req *providers.GenerateRequest) (<-chan *providers.StreamChunk, error) {
	ch := make(chan *providers.StreamChunk, 1)
	go func() {
		defer close(ch)
		ch <- &providers.StreamChunk{
			Choices: []providers.StreamChoice{
				{
					Delta: providers.MessageDelta{
						Role:    "assistant",
						Content: "Mock stream response",
					},
				},
			},
			Done: true,
		}
	}()
	return ch, nil
}

func (m *MockLLMProvider) ValidateConfig() error {
	return nil
}

func (m *MockLLMProvider) GetModels(ctx context.Context) ([]string, error) {
	return []string{"test-model"}, nil
}

func (m *MockLLMProvider) Close() error {
	return nil
}

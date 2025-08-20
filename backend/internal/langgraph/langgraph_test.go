package langgraph

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestState_BasicOperations(t *testing.T) {
	state := NewState("test-state", "test-graph")

	// Test Set and Get
	state.Set("key1", "value1")
	value, exists := state.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", value)

	// Test GetString
	str, ok := state.GetString("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", str)

	// Test SetMultiple
	state.SetMultiple(map[string]interface{}{
		"key2": 42,
		"key3": true,
	})

	intVal, ok := state.GetInt("key2")
	assert.True(t, ok)
	assert.Equal(t, 42, intVal)

	boolVal, ok := state.GetBool("key3")
	assert.True(t, ok)
	assert.True(t, boolVal)

	// Test Has and Delete
	assert.True(t, state.Has("key1"))
	state.Delete("key1")
	assert.False(t, state.Has("key1"))

	// Test Keys and Size
	keys := state.Keys()
	assert.Contains(t, keys, "key2")
	assert.Contains(t, keys, "key3")
	assert.Equal(t, 2, state.Size())
}

func TestState_Clone(t *testing.T) {
	original := NewState("test-state", "test-graph")
	original.Set("key1", "value1")
	original.Set("nested", map[string]interface{}{
		"inner": "value",
	})

	clone := original.Clone()

	// Verify clone has same data
	value, exists := clone.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", value)

	// Verify modifications don't affect original
	clone.Set("key2", "value2")
	assert.True(t, clone.Has("key2"))
	assert.False(t, original.Has("key2"))

	// Verify deep copy of nested structures
	nested, _ := clone.GetMap("nested")
	nested["inner"] = "modified"
	
	originalNested, _ := original.GetMap("nested")
	assert.Equal(t, "value", originalNested["inner"])
}

func TestMemoryStateManager(t *testing.T) {
	ctx := context.Background()
	manager := NewMemoryStateManager()

	state := NewState("test-state", "test-graph")
	state.Set("key1", "value1")

	// Test SaveState
	err := manager.SaveState(ctx, state)
	require.NoError(t, err)

	// Test LoadState
	loaded, err := manager.LoadState(ctx, "test-state")
	require.NoError(t, err)
	assert.Equal(t, state.ID, loaded.ID)
	assert.Equal(t, state.GraphID, loaded.GraphID)

	value, exists := loaded.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", value)

	// Test ListStates
	states, err := manager.ListStates(ctx, nil)
	require.NoError(t, err)
	assert.Len(t, states, 1)

	// Test filtered listing
	states, err = manager.ListStates(ctx, map[string]interface{}{
		"graph_id": "test-graph",
	})
	require.NoError(t, err)
	assert.Len(t, states, 1)

	states, err = manager.ListStates(ctx, map[string]interface{}{
		"graph_id": "non-existent",
	})
	require.NoError(t, err)
	assert.Len(t, states, 0)

	// Test DeleteState
	err = manager.DeleteState(ctx, "test-state")
	require.NoError(t, err)

	_, err = manager.LoadState(ctx, "test-state")
	assert.Error(t, err)
}

func TestNodes(t *testing.T) {
	ctx := context.Background()
	state := NewState("test-state", "test-graph")

	t.Run("StartNode", func(t *testing.T) {
		node := NewStartNode("start", "Start Node")
		node.InitialData["initial_key"] = "initial_value"

		newState, err := node.Execute(ctx, state)
		require.NoError(t, err)

		value, exists := newState.Get("initial_key")
		assert.True(t, exists)
		assert.Equal(t, "initial_value", value)
	})

	t.Run("FunctionNode", func(t *testing.T) {
		fn := func(ctx context.Context, state *State) (*State, error) {
			newState := state.Clone()
			newState.Set("function_result", "success")
			return newState, nil
		}

		node := NewFunctionNode("function", "Function Node", fn)
		newState, err := node.Execute(ctx, state)
		require.NoError(t, err)

		value, exists := newState.Get("function_result")
		assert.True(t, exists)
		assert.Equal(t, "success", value)
	})

	t.Run("ConditionalNode", func(t *testing.T) {
		condition := func(ctx context.Context, state *State) (bool, error) {
			return true, nil
		}

		node := NewConditionalNode("conditional", "Conditional Node", condition)
		newState, err := node.Execute(ctx, state)
		require.NoError(t, err)

		value, exists := newState.Get("condition_result_true")
		assert.True(t, exists)
		assert.True(t, value.(bool))
	})
}

func TestConditions(t *testing.T) {
	ctx := context.Background()
	state := NewState("test-state", "test-graph")
	state.Set("key1", "value1")
	state.Set("number", 42)
	state.Set("flag", true)

	t.Run("StateKeyCondition", func(t *testing.T) {
		// Test exists
		condition := NewStateKeyCondition("key1")
		result, err := condition.Evaluate(ctx, state)
		require.NoError(t, err)
		assert.True(t, result)

		// Test not exists
		condition = NewStateKeyCondition("non-existent")
		result, err = condition.Evaluate(ctx, state)
		require.NoError(t, err)
		assert.False(t, result)

		// Test equals
		condition = NewStateValueCondition("key1", "value1", "equals")
		result, err = condition.Evaluate(ctx, state)
		require.NoError(t, err)
		assert.True(t, result)

		// Test not equals
		condition = NewStateValueCondition("key1", "different", "not_equals")
		result, err = condition.Evaluate(ctx, state)
		require.NoError(t, err)
		assert.True(t, result)

		// Test greater
		condition = NewStateValueCondition("number", 40, "greater")
		result, err = condition.Evaluate(ctx, state)
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("FunctionCondition", func(t *testing.T) {
		fn := func(ctx context.Context, state *State) (bool, error) {
			flag, _ := state.GetBool("flag")
			return flag, nil
		}

		condition := NewFunctionCondition("check flag", fn)
		result, err := condition.Evaluate(ctx, state)
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("AndCondition", func(t *testing.T) {
		cond1 := NewStateKeyCondition("key1")
		cond2 := NewStateKeyCondition("number")
		
		condition := NewAndCondition(cond1, cond2)
		result, err := condition.Evaluate(ctx, state)
		require.NoError(t, err)
		assert.True(t, result)

		// Add a false condition
		cond3 := NewStateKeyCondition("non-existent")
		condition = NewAndCondition(cond1, cond2, cond3)
		result, err = condition.Evaluate(ctx, state)
		require.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("OrCondition", func(t *testing.T) {
		cond1 := NewStateKeyCondition("key1")
		cond2 := NewStateKeyCondition("non-existent")
		
		condition := NewOrCondition(cond1, cond2)
		result, err := condition.Evaluate(ctx, state)
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("NotCondition", func(t *testing.T) {
		cond := NewStateKeyCondition("non-existent")
		condition := NewNotCondition(cond)
		result, err := condition.Evaluate(ctx, state)
		require.NoError(t, err)
		assert.True(t, result)
	})
}

func TestGraph(t *testing.T) {
	ctx := context.Background()
	stateManager := NewMemoryStateManager()
	
	t.Run("BasicGraph", func(t *testing.T) {
		graph := NewGraph("test-graph", "Test Graph", stateManager)

		// Add nodes
		startNode := NewStartNode("start", "Start")
		endNode := NewEndNode("end", "End")
		
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
		edge := NewEdge("start", "end", "start to end")
		err = graph.AddEdge(edge)
		require.NoError(t, err)

		// Validate
		err = graph.Validate()
		require.NoError(t, err)

		// Execute
		initialState := NewState("test-state", "test-graph")
		finalState, err := graph.Execute(ctx, initialState)
		require.NoError(t, err)
		assert.NotNil(t, finalState)
	})

	t.Run("ConditionalGraph", func(t *testing.T) {
		graph := NewGraph("conditional-graph", "Conditional Graph", stateManager)

		// Add nodes
		startNode := NewStartNode("start", "Start")
		startNode.InitialData["condition_value"] = true
		
		conditionalNode := NewConditionalNode("conditional", "Conditional", func(ctx context.Context, state *State) (bool, error) {
			value, _ := state.GetBool("condition_value")
			return value, nil
		})
		
		trueNode := NewFunctionNode("true_path", "True Path", func(ctx context.Context, state *State) (*State, error) {
			newState := state.Clone()
			newState.Set("path_taken", "true")
			return newState, nil
		})
		
		falseNode := NewFunctionNode("false_path", "False Path", func(ctx context.Context, state *State) (*State, error) {
			newState := state.Clone()
			newState.Set("path_taken", "false")
			return newState, nil
		})
		
		endNode := NewEndNode("end", "End")

		// Add all nodes
		graph.AddNode(startNode)
		graph.AddNode(conditionalNode)
		graph.AddNode(trueNode)
		graph.AddNode(falseNode)
		graph.AddNode(endNode)

		// Set entry and exit points
		graph.SetEntryPoint("start")
		graph.AddExitPoint("end")

		// Add edges
		graph.AddEdge(NewEdge("start", "conditional", "start to conditional"))
		
		// Conditional edges
		trueCondition := NewStateKeyCondition("condition_result_true")
		falseCondition := NewStateKeyCondition("condition_result_false")
		
		graph.AddEdge(NewConditionalEdge("conditional", "true_path", "condition is true", trueCondition))
		graph.AddEdge(NewConditionalEdge("conditional", "false_path", "condition is false", falseCondition))
		
		graph.AddEdge(NewEdge("true_path", "end", "true path to end"))
		graph.AddEdge(NewEdge("false_path", "end", "false path to end"))

		// Execute
		initialState := NewState("test-state", "conditional-graph")
		finalState, err := graph.Execute(ctx, initialState)
		require.NoError(t, err)

		pathTaken, exists := finalState.Get("path_taken")
		assert.True(t, exists)
		assert.Equal(t, "true", pathTaken)
	})
}

func TestGraphBuilder(t *testing.T) {
	stateManager := NewMemoryStateManager()
	
	t.Run("BasicBuilder", func(t *testing.T) {
		builder := NewGraphBuilder("Test Graph", stateManager)
		
		graph, err := builder.
			AddStartNode("start", "Start").
			AddFunctionNode("process", "Process", func(ctx context.Context, state *State) (*State, error) {
				newState := state.Clone()
				newState.Set("processed", true)
				return newState, nil
			}).
			AddEndNode("end", "End").
			From("start").ConnectTo("process").ConnectTo("end").
			Build()
		
		require.NoError(t, err)
		assert.Equal(t, "Test Graph", graph.Name)
		assert.Equal(t, 3, graph.GetNodeCount())
		assert.Equal(t, 2, graph.GetEdgeCount())
	})

	t.Run("TravelGraphBuilder", func(t *testing.T) {
		builder := NewTravelGraphBuilder("Travel Planning", stateManager)
		
		graph, err := builder.BuildFlightSearchGraph()
		require.NoError(t, err)
		
		assert.Equal(t, "Travel Planning", graph.Name)
		assert.Equal(t, 3, graph.GetNodeCount()) // start, search, end
		assert.Equal(t, 2, graph.GetEdgeCount()) // start->search, search->end
	})
}

func TestGraphExecutor(t *testing.T) {
	ctx := context.Background()
	stateManager := NewMemoryStateManager()
	executor := NewGraphExecutor(stateManager)
	
	// Create a simple graph
	builder := NewGraphBuilder("Test Graph", stateManager)
	graph, err := builder.
		AddStartNode("start", "Start").
		AddFunctionNode("process", "Process", func(ctx context.Context, state *State) (*State, error) {
			newState := state.Clone()
			counter, _ := newState.GetInt("counter")
			newState.Set("counter", counter+1)
			return newState, nil
		}).
		AddEndNode("end", "End").
		From("start").ConnectTo("process").ConnectTo("end").
		Build()
	
	require.NoError(t, err)

	// Execute graph
	input := map[string]interface{}{
		"counter": 0,
	}
	
	result, err := executor.Execute(ctx, graph, input, nil)
	require.NoError(t, err)
	
	assert.Equal(t, "completed", result.Status)
	assert.NotNil(t, result.FinalState)
	
	counter, exists := result.FinalState.Get("counter")
	assert.True(t, exists)
	assert.Equal(t, 1, counter)
	
	assert.Contains(t, result.NodesVisited, "start")
	assert.Contains(t, result.NodesVisited, "process")
	assert.Contains(t, result.NodesVisited, "end")
}

func TestGraphExecutor_Async(t *testing.T) {
	ctx := context.Background()
	stateManager := NewMemoryStateManager()
	executor := NewGraphExecutor(stateManager)
	
	// Create a graph with a delay
	builder := NewGraphBuilder("Async Test Graph", stateManager)
	graph, err := builder.
		AddStartNode("start", "Start").
		AddFunctionNode("delay", "Delay", func(ctx context.Context, state *State) (*State, error) {
			time.Sleep(100 * time.Millisecond) // Small delay
			newState := state.Clone()
			newState.Set("delayed", true)
			return newState, nil
		}).
		AddEndNode("end", "End").
		From("start").ConnectTo("delay").ConnectTo("end").
		Build()
	
	require.NoError(t, err)

	// Execute asynchronously
	input := map[string]interface{}{
		"test": "async",
	}
	
	executionID, resultChan, err := executor.ExecuteAsync(ctx, graph, input, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, executionID)

	// Wait for result
	select {
	case result := <-resultChan:
		assert.Equal(t, "completed", result.Status)
		assert.NotNil(t, result.FinalState)
		
		delayed, exists := result.FinalState.Get("delayed")
		assert.True(t, exists)
		assert.True(t, delayed.(bool))
		
	case <-time.After(5 * time.Second):
		t.Fatal("Execution timed out")
	}
}

package langchain

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptTemplate(t *testing.T) {
	ctx := context.Background()

	t.Run("Basic Template", func(t *testing.T) {
		template := NewPromptTemplate(
			"test_template",
			"Hello {{.name}}, you are {{.age}} years old.",
			[]string{"name", "age"},
		)

		vars := map[string]interface{}{
			"name": "Alice",
			"age":  30,
		}

		result, err := template.Render(ctx, vars)
		require.NoError(t, err)
		assert.Equal(t, "Hello Alice, you are 30 years old.", result)
	})

	t.Run("Missing Variables", func(t *testing.T) {
		template := NewPromptTemplate(
			"test_template",
			"Hello {{.name}}, you are {{.age}} years old.",
			[]string{"name", "age"},
		)

		vars := map[string]interface{}{
			"name": "Alice",
			// missing age
		}

		_, err := template.Render(ctx, vars)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required variables")
	})

	t.Run("Partial Variables", func(t *testing.T) {
		template := NewPromptTemplate(
			"test_template",
			"Hello {{.name}}, you are from {{.country}}.",
			[]string{"name", "country"},
		)

		template.SetPartial("country", "USA")

		vars := map[string]interface{}{
			"name": "Bob",
		}

		result, err := template.Render(ctx, vars)
		require.NoError(t, err)
		assert.Equal(t, "Hello Bob, you are from USA.", result)
	})

	t.Run("Template Validation", func(t *testing.T) {
		template := NewPromptTemplate(
			"test_template",
			"Valid template with {{.variable}}",
			[]string{"variable"},
		)

		err := template.Validate()
		assert.NoError(t, err)

		// Test invalid template
		invalidTemplate := &PromptTemplate{
			Name:     "",
			Template: "{{.invalid syntax",
		}

		err = invalidTemplate.Validate()
		assert.Error(t, err)
	})
}

func TestChatPromptTemplate(t *testing.T) {
	ctx := context.Background()

	t.Run("Multi-Message Template", func(t *testing.T) {
		systemTemplate := NewPromptTemplate(
			"system",
			"You are a helpful assistant specializing in {{.domain}}.",
			[]string{"domain"},
		)

		userTemplate := NewPromptTemplate(
			"user",
			"Please help me with: {{.query}}",
			[]string{"query"},
		)

		chatTemplate := NewChatPromptTemplate(
			"travel_assistant",
			[]MessageTemplate{
				{Role: "system", Template: systemTemplate},
				{Role: "user", Template: userTemplate},
			},
		)

		vars := map[string]interface{}{
			"domain": "travel planning",
			"query":  "planning a trip to Japan",
		}

		messages, err := chatTemplate.RenderMessages(ctx, vars)
		require.NoError(t, err)
		assert.Len(t, messages, 2)
		assert.Equal(t, "system", messages[0].Role)
		assert.Contains(t, messages[0].Content, "travel planning")
		assert.Equal(t, "user", messages[1].Role)
		assert.Contains(t, messages[1].Content, "planning a trip to Japan")
	})
}

func TestOutputParsers(t *testing.T) {
	ctx := context.Background()

	t.Run("JSON Parser", func(t *testing.T) {
		parser := NewJSONParser("test_json", nil, false)

		// Test valid JSON
		jsonOutput := `{"name": "Paris", "country": "France", "population": 2161000}`
		result, err := parser.Parse(ctx, jsonOutput)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Paris", resultMap["name"])
		assert.Equal(t, "France", resultMap["country"])

		// Test JSON in markdown code block
		markdownOutput := "```json\n{\"city\": \"Tokyo\"}\n```"
		result2, err := parser.Parse(ctx, markdownOutput)
		require.NoError(t, err)

		resultMap2, ok := result2.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Tokyo", resultMap2["city"])

		// Test invalid JSON
		_, err = parser.Parse(ctx, "invalid json")
		assert.Error(t, err)
	})

	t.Run("List Parser", func(t *testing.T) {
		parser := NewListParser("test_list", "\n", true, true)

		listOutput := "1. First item\n2. Second item\n3. Third item"
		result, err := parser.Parse(ctx, listOutput)
		require.NoError(t, err)

		items, ok := result.([]string)
		require.True(t, ok)
		assert.Len(t, items, 3)
		assert.Equal(t, "First item", items[0])
		assert.Equal(t, "Second item", items[1])
		assert.Equal(t, "Third item", items[2])
	})

	t.Run("Key-Value Parser", func(t *testing.T) {
		parser := NewKeyValueParser("test_kv", ":", "\n", true)

		kvOutput := "name: Alice\nage: 30\ncity: New York"
		result, err := parser.Parse(ctx, kvOutput)
		require.NoError(t, err)

		kvMap, ok := result.(map[string]string)
		require.True(t, ok)
		assert.Equal(t, "Alice", kvMap["name"])
		assert.Equal(t, "30", kvMap["age"])
		assert.Equal(t, "New York", kvMap["city"])
	})

	t.Run("Number Parser", func(t *testing.T) {
		intParser := NewNumberParser("test_int", "int", 0)

		// Test integer parsing
		result, err := intParser.Parse(ctx, "The answer is 42")
		require.NoError(t, err)
		assert.Equal(t, 42, result)

		floatParser := NewNumberParser("test_float", "float", 0.0)

		// Test float parsing
		result2, err := floatParser.Parse(ctx, "The price is $19.99")
		require.NoError(t, err)
		assert.Equal(t, 19.99, result2)
	})

	t.Run("Regex Parser", func(t *testing.T) {
		parser := NewRegexParser(
			"email_parser",
			`(\w+)@(\w+\.\w+)`,
			[]string{"username", "domain"},
			false,
		)

		result, err := parser.Parse(ctx, "Contact us at support@example.com for help")
		require.NoError(t, err)

		resultMap, ok := result.(map[string]string)
		require.True(t, ok)
		assert.Equal(t, "support@example.com", resultMap["full_match"])
		assert.Equal(t, "support", resultMap["username"])
		assert.Equal(t, "example.com", resultMap["domain"])
	})
}

func TestMemory(t *testing.T) {
	ctx := context.Background()
	sessionID := "test_session"

	t.Run("Buffer Memory", func(t *testing.T) {
		memory := NewBufferMemory("test_buffer", 5)

		// Add messages
		messages := []*Message{
			{SessionID: sessionID, Role: "user", Content: "Hello", Timestamp: time.Now()},
			{SessionID: sessionID, Role: "assistant", Content: "Hi there!", Timestamp: time.Now()},
			{SessionID: sessionID, Role: "user", Content: "How are you?", Timestamp: time.Now()},
			{SessionID: sessionID, Role: "assistant", Content: "I'm doing well!", Timestamp: time.Now()},
		}

		for _, msg := range messages {
			err := memory.AddMessage(ctx, msg)
			require.NoError(t, err)
		}

		// Retrieve messages
		retrieved, err := memory.GetMessages(ctx, sessionID, 0)
		require.NoError(t, err)
		assert.Len(t, retrieved, 4)
		assert.Equal(t, "Hello", retrieved[0].Content)
		assert.Equal(t, "I'm doing well!", retrieved[3].Content)

		// Test limit
		limited, err := memory.GetMessages(ctx, sessionID, 2)
		require.NoError(t, err)
		assert.Len(t, limited, 2)
		assert.Equal(t, "How are you?", limited[0].Content)

		// Test summary
		summary, err := memory.GetSummary(ctx, sessionID)
		require.NoError(t, err)
		assert.Contains(t, summary, "2 user messages")
		assert.Contains(t, summary, "2 assistant responses")
	})

	t.Run("Window Memory", func(t *testing.T) {
		memory := NewWindowMemory("test_window", 3)

		// Add more messages than window size
		for i := 0; i < 5; i++ {
			msg := &Message{
				SessionID: sessionID,
				Role:      "user",
				Content:   fmt.Sprintf("Message %d", i),
				Timestamp: time.Now(),
			}
			err := memory.AddMessage(ctx, msg)
			require.NoError(t, err)
		}

		// Should only keep last 3 messages
		retrieved, err := memory.GetMessages(ctx, sessionID, 0)
		require.NoError(t, err)
		assert.Len(t, retrieved, 3)
		assert.Equal(t, "Message 2", retrieved[0].Content)
		assert.Equal(t, "Message 4", retrieved[2].Content)
	})

	t.Run("Memory Manager", func(t *testing.T) {
		manager := NewMemoryManager()

		buffer := NewBufferMemory("buffer", 10)
		window := NewWindowMemory("window", 5)

		manager.RegisterMemory(buffer)
		manager.RegisterMemory(window)

		// Test retrieval
		retrieved, err := manager.GetMemory("buffer")
		require.NoError(t, err)
		assert.Equal(t, "buffer", retrieved.GetName())

		// Test listing
		names := manager.ListMemories()
		assert.Contains(t, names, "buffer")
		assert.Contains(t, names, "window")

		// Test not found
		_, err = manager.GetMemory("nonexistent")
		assert.Error(t, err)
	})
}

func TestChains(t *testing.T) {
	ctx := context.Background()

	t.Run("Sequential Chain", func(t *testing.T) {
		// Create mock chains
		chain1 := &MockChain{name: "chain1", output: map[string]interface{}{"step1": "done"}}
		chain2 := &MockChain{name: "chain2", output: map[string]interface{}{"step2": "done"}}

		sequential := NewSequentialChain("test_sequential", "Test sequential chain", chain1, chain2)

		input := map[string]interface{}{"start": "value"}
		result, err := sequential.Execute(ctx, input)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "done", result.Output["step1"])
		assert.Equal(t, "done", result.Output["step2"])
	})

	t.Run("Parallel Chain", func(t *testing.T) {
		// Create mock chains
		chain1 := &MockChain{name: "chain1", output: map[string]interface{}{"result1": "value1"}}
		chain2 := &MockChain{name: "chain2", output: map[string]interface{}{"result2": "value2"}}

		parallel := NewParallelChain("test_parallel", "Test parallel chain", chain1, chain2)

		input := map[string]interface{}{"start": "value"}
		result, err := parallel.Execute(ctx, input)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "value1", result.Output["chain1_result1"])
		assert.Equal(t, "value2", result.Output["chain2_result2"])
	})

	t.Run("Conditional Chain", func(t *testing.T) {
		trueChain := &MockChain{name: "true", output: map[string]interface{}{"path": "true"}}
		falseChain := &MockChain{name: "false", output: map[string]interface{}{"path": "false"}}

		condition := func(ctx context.Context, input map[string]interface{}) (bool, error) {
			value, exists := input["condition"]
			if !exists {
				return false, nil
			}
			return value.(bool), nil
		}

		conditional := NewConditionalChain("test_conditional", "Test conditional chain", condition, trueChain, falseChain)

		// Test true condition
		input := map[string]interface{}{"condition": true}
		result, err := conditional.Execute(ctx, input)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "true", result.Output["path"])

		// Test false condition
		input["condition"] = false
		result, err = conditional.Execute(ctx, input)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "false", result.Output["path"])
	})
}

func TestTravelPromptTemplates(t *testing.T) {
	t.Run("Travel Templates", func(t *testing.T) {
		travelTemplates := NewTravelPromptTemplates()
		registry := travelTemplates.GetRegistry()

		// Test destination research template
		template, err := registry.GetTemplate("destination_research")
		require.NoError(t, err)
		assert.Equal(t, "destination_research", template.Name)
		assert.Contains(t, template.InputVars, "destination")
		assert.Contains(t, template.InputVars, "budget")

		// Test flight analysis template
		flightTemplate, err := registry.GetTemplate("flight_analysis")
		require.NoError(t, err)
		assert.Equal(t, "flight_analysis", flightTemplate.Name)
		assert.Contains(t, flightTemplate.InputVars, "origin")
		assert.Contains(t, flightTemplate.InputVars, "destination")

		// Test template listing
		templates := registry.ListTemplates()
		assert.Contains(t, templates, "destination_research")
		assert.Contains(t, templates, "flight_analysis")
		assert.Contains(t, templates, "hotel_recommendations")
		assert.Contains(t, templates, "itinerary_planning")
	})
}

// MockChain for testing
type MockChain struct {
	name   string
	output map[string]interface{}
	err    error
}

func (m *MockChain) Execute(ctx context.Context, input map[string]interface{}) (*ChainResult, error) {
	if m.err != nil {
		return nil, m.err
	}

	return &ChainResult{
		Output:    m.output,
		Metadata:  map[string]interface{}{"chain": m.name},
		Duration:  time.Millisecond * 10,
		Success:   true,
		ChainName: m.name,
	}, nil
}

func (m *MockChain) GetName() string {
	return m.name
}

func (m *MockChain) GetDescription() string {
	return fmt.Sprintf("Mock chain: %s", m.name)
}

func (m *MockChain) Validate() error {
	return nil
}

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/exotic-travel-booking/backend/internal/langchain"
	"github.com/exotic-travel-booking/backend/internal/langgraph"
	"github.com/exotic-travel-booking/backend/internal/llm/providers"
	"github.com/exotic-travel-booking/backend/internal/rag"
	"github.com/exotic-travel-booking/backend/internal/services"
	"github.com/exotic-travel-booking/backend/internal/tools"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestHandler() (*AIHandler, *fiber.App) {
	// Create mock components
	embeddingService := rag.NewMockEmbeddingService(384, "mock-embed")
	vectorStore := rag.NewMemoryVectorStore(embeddingService)
	retriever := rag.NewTravelRetriever(vectorStore)
	
	ragConfig := &rag.RAGConfig{
		MaxContextLength:   1000,
		RetrievalLimit:     3,
		RelevanceThreshold: 0.1,
		IncludeSourceInfo:  true,
		MaxTokens:          500,
		Temperature:        0.7,
	}

	mockLLM := &MockLLMProvider{}
	ragChain := rag.NewTravelRAGChain(retriever, mockLLM, ragConfig)

	memoryManager := langchain.NewMemoryManager()
	bufferMemory := langchain.NewBufferMemory("travel_conversation", 20)
	memoryManager.RegisterMemory(bufferMemory)

	stateManager := langgraph.NewMemoryStateManager()
	toolRegistry := tools.NewToolRegistry()
	ollamaService := services.NewOllamaService("http://localhost:11434")

	// Add sample data to vector store
	ctx := context.Background()
	sampleDoc := &rag.Document{
		ID:      "test_doc",
		Content: "Paris is a beautiful city with the Eiffel Tower and great food.",
		Metadata: map[string]interface{}{
			"destination":   "Paris",
			"content_type":  "city_guide",
			"document_type": "travel_guide",
		},
	}
	vectorStore.AddDocument(ctx, sampleDoc)

	handler := NewAIHandler(
		mockLLM,
		ragChain,
		memoryManager,
		stateManager,
		toolRegistry,
		ollamaService,
	)

	app := fiber.New()
	return handler, app
}

func TestAIHandler_Chat(t *testing.T) {
	handler, app := setupTestHandler()

	t.Run("Valid Chat Request", func(t *testing.T) {
		app.Post("/chat", handler.Chat)

		reqBody := ChatRequest{
			Message:   "Tell me about Paris",
			SessionID: "test_session_1",
			UserID:    "test_user",
			Stream:    false,
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/chat", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 10000) // 10 second timeout
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var response ChatResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "test_session_1", response.SessionID)
		assert.NotEmpty(t, response.Response)
		assert.NotEmpty(t, response.MessageID)
		assert.Greater(t, response.TokensUsed, 0)
		assert.Greater(t, response.Duration, time.Duration(0))
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		app.Post("/chat", handler.Chat)

		req := httptest.NewRequest("POST", "/chat", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Missing Required Fields", func(t *testing.T) {
		app.Post("/chat", handler.Chat)

		reqBody := ChatRequest{
			// Missing Message and SessionID
			UserID: "test_user",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/chat", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		// Should still work but with empty message
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}

func TestAIHandler_QueryKnowledgeBase(t *testing.T) {
	handler, app := setupTestHandler()

	t.Run("Valid Knowledge Base Query", func(t *testing.T) {
		app.Post("/knowledge/query", handler.QueryKnowledgeBase)

		reqBody := RAGQueryRequest{
			Query: "What can you tell me about Paris?",
			Limit: 5,
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/knowledge/query", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 10000)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var response RAGQueryResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "What can you tell me about Paris?", response.Query)
		assert.NotEmpty(t, response.Answer)
		assert.NotNil(t, response.Sources)
		assert.Greater(t, response.Duration, time.Duration(0))
	})

	t.Run("Destination-Specific Query", func(t *testing.T) {
		app.Post("/knowledge/query", handler.QueryKnowledgeBase)

		reqBody := RAGQueryRequest{
			Query:       "What are the best attractions?",
			Destination: "Paris",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/knowledge/query", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 10000)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var response RAGQueryResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.Answer)
		assert.Contains(t, response.Metadata, "destination")
	})

	t.Run("Category-Based Query", func(t *testing.T) {
		app.Post("/knowledge/query", handler.QueryKnowledgeBase)

		reqBody := RAGQueryRequest{
			Query:    "Tell me about city guides",
			Category: "city_guide",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/knowledge/query", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 10000)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var response RAGQueryResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.Answer)
		assert.Contains(t, response.Metadata, "category")
	})
}

func TestAIHandler_ProcessAgentRequest(t *testing.T) {
	handler, app := setupTestHandler()

	t.Run("Valid Flight Agent Request", func(t *testing.T) {
		app.Post("/agents/request", handler.ProcessAgentRequest)

		reqBody := AgentRequest{
			AgentType: "flight",
			Query:     "Find flights from NYC to Paris",
			SessionID: "test_session",
			UserID:    "test_user",
			Parameters: map[string]interface{}{
				"origin":      "NYC",
				"destination": "Paris",
				"start_date":  "2024-06-15",
				"travelers":   2,
			},
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/agents/request", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 10000)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var response AgentResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "flight", response.AgentType)
		assert.Equal(t, "test_session", response.SessionID)
		assert.NotEmpty(t, response.RequestID)
		assert.NotEmpty(t, response.Response)
		assert.Greater(t, response.Confidence, 0.0)
		assert.Greater(t, response.Duration, time.Duration(0))
	})

	t.Run("Valid Hotel Agent Request", func(t *testing.T) {
		app.Post("/agents/request", handler.ProcessAgentRequest)

		reqBody := AgentRequest{
			AgentType: "hotel",
			Query:     "Find hotels in Tokyo",
			SessionID: "test_session",
			UserID:    "test_user",
			Parameters: map[string]interface{}{
				"destination": "Tokyo",
				"start_date":  "2024-07-01",
				"end_date":    "2024-07-05",
				"travelers":   2,
			},
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/agents/request", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 10000)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var response AgentResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "hotel", response.AgentType)
		assert.NotEmpty(t, response.Response)
	})

	t.Run("Invalid Agent Type", func(t *testing.T) {
		app.Post("/agents/request", handler.ProcessAgentRequest)

		reqBody := AgentRequest{
			AgentType: "invalid_agent",
			Query:     "Some query",
			SessionID: "test_session",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/agents/request", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

func TestAIHandler_GetConversationHistory(t *testing.T) {
	handler, app := setupTestHandler()

	t.Run("Valid Session ID", func(t *testing.T) {
		// First, add some messages to the session
		ctx := context.Background()
		memory, _ := handler.memoryManager.GetMemory("travel_conversation")
		
		messages := []*langchain.Message{
			{SessionID: "test_session", Role: "user", Content: "Hello", Timestamp: time.Now()},
			{SessionID: "test_session", Role: "assistant", Content: "Hi there!", Timestamp: time.Now()},
		}

		for _, msg := range messages {
			memory.AddMessage(ctx, msg)
		}

		app.Get("/history/:sessionId", handler.GetConversationHistory)

		req := httptest.NewRequest("GET", "/history/test_session", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var response ConversationHistoryResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "test_session", response.SessionID)
		assert.Len(t, response.Messages, 2)
		assert.NotEmpty(t, response.Summary)
	})

	t.Run("Empty Session ID", func(t *testing.T) {
		app.Get("/history/:sessionId", handler.GetConversationHistory)

		req := httptest.NewRequest("GET", "/history/", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode) // Fiber returns 404 for missing params
	})
}

func TestAIHandler_ClearConversation(t *testing.T) {
	handler, app := setupTestHandler()

	t.Run("Valid Clear Request", func(t *testing.T) {
		app.Delete("/clear", handler.ClearConversation)

		reqBody := ClearConversationRequest{
			SessionID: "test_session_clear",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("DELETE", "/clear", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "message")
		assert.Equal(t, "test_session_clear", response["session_id"])
	})
}

// MockLLMProvider for testing
type MockLLMProvider struct{}

func (m *MockLLMProvider) GetName() string {
	return "mock-llm"
}

func (m *MockLLMProvider) GenerateResponse(ctx context.Context, req *providers.GenerateRequest) (*providers.GenerateResponse, error) {
	return &providers.GenerateResponse{
		Choices: []providers.Choice{
			{
				Message: providers.Message{
					Role:    "assistant",
					Content: "This is a mock response for testing purposes.",
				},
			},
		},
		Usage: providers.Usage{
			PromptTokens:     50,
			CompletionTokens: 25,
			TotalTokens:      75,
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

func (m *MockLLMProvider) GetModels(ctx context.Context) ([]string, error) {
	return []string{"mock-model"}, nil
}

func (m *MockLLMProvider) Close() error {
	return nil
}

func (m *MockLLMProvider) ValidateConfig() error {
	return nil
}

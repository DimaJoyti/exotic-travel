package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/exotic-travel-booking/backend/internal/agents/specialist"
	"github.com/exotic-travel-booking/backend/internal/langchain"
	"github.com/exotic-travel-booking/backend/internal/langgraph"
	"github.com/exotic-travel-booking/backend/internal/llm/providers"
	"github.com/exotic-travel-booking/backend/internal/rag"
	"github.com/exotic-travel-booking/backend/internal/services"
	"github.com/exotic-travel-booking/backend/internal/tools"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// AIHandler handles AI-related API endpoints
type AIHandler struct {
	llmProvider     providers.LLMProvider
	ragChain        *rag.TravelRAGChain
	memoryManager   *langchain.MemoryManager
	stateManager    langgraph.StateManager
	toolRegistry    *tools.ToolRegistry
	ollamaService   *services.OllamaService
	flightAgent     *specialist.FlightAgent
	hotelAgent      *specialist.HotelAgent
	itineraryAgent  *specialist.ItineraryAgent
	supervisorAgent *specialist.SupervisorAgent
	tracer          trace.Tracer
}

// NewAIHandler creates a new AI handler
func NewAIHandler(
	llmProvider providers.LLMProvider,
	ragChain *rag.TravelRAGChain,
	memoryManager *langchain.MemoryManager,
	stateManager langgraph.StateManager,
	toolRegistry *tools.ToolRegistry,
	ollamaService *services.OllamaService,
) *AIHandler {
	// Create specialist agents
	flightAgent := specialist.NewFlightAgent(llmProvider, toolRegistry, stateManager)
	hotelAgent := specialist.NewHotelAgent(llmProvider, toolRegistry, stateManager)
	itineraryAgent := specialist.NewItineraryAgent(llmProvider, toolRegistry, stateManager)
	supervisorAgent := specialist.NewSupervisorAgent(llmProvider, toolRegistry, stateManager)

	return &AIHandler{
		llmProvider:     llmProvider,
		ragChain:        ragChain,
		memoryManager:   memoryManager,
		stateManager:    stateManager,
		toolRegistry:    toolRegistry,
		ollamaService:   ollamaService,
		flightAgent:     flightAgent,
		hotelAgent:      hotelAgent,
		itineraryAgent:  itineraryAgent,
		supervisorAgent: supervisorAgent,
		tracer:          otel.Tracer("api.ai_handler"),
	}
}

// ChatRequest represents a chat request
type ChatRequest struct {
	Message     string                 `json:"message" validate:"required"`
	SessionID   string                 `json:"session_id" validate:"required"`
	UserID      string                 `json:"user_id"`
	Context     map[string]interface{} `json:"context"`
	Stream      bool                   `json:"stream"`
	MaxTokens   int                    `json:"max_tokens"`
	Temperature float64                `json:"temperature"`
}

// ChatResponse represents a chat response
type ChatResponse struct {
	Response   string                 `json:"response"`
	SessionID  string                 `json:"session_id"`
	MessageID  string                 `json:"message_id"`
	Sources    []*rag.RetrievalResult `json:"sources,omitempty"`
	Metadata   map[string]interface{} `json:"metadata"`
	TokensUsed int                    `json:"tokens_used"`
	Duration   time.Duration          `json:"duration"`
	Timestamp  time.Time              `json:"timestamp"`
}

// StreamChunk represents a streaming response chunk
type StreamChunk struct {
	Content   string                 `json:"content"`
	SessionID string                 `json:"session_id"`
	MessageID string                 `json:"message_id"`
	Done      bool                   `json:"done"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Chat handles general chat requests with RAG
func (h *AIHandler) Chat(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "ai_handler.chat")
	defer span.End()

	var req ChatRequest
	if err := c.BodyParser(&req); err != nil {
		span.RecordError(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	span.SetAttributes(
		attribute.String("session_id", req.SessionID),
		attribute.String("user_id", req.UserID),
		attribute.Bool("stream", req.Stream),
		attribute.Int("message_length", len(req.Message)),
	)

	if req.Stream {
		return h.handleStreamingChat(c, ctx, &req)
	}

	return h.handleRegularChat(c, ctx, &req)
}

// handleRegularChat handles non-streaming chat requests
func (h *AIHandler) handleRegularChat(c *fiber.Ctx, ctx context.Context, req *ChatRequest) error {
	startTime := time.Now()

	// Add user message to memory
	memory, err := h.memoryManager.GetMemory("travel_conversation")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get conversation memory",
		})
	}

	userMessage := &langchain.Message{
		SessionID: req.SessionID,
		Role:      "user",
		Content:   req.Message,
		Timestamp: time.Now(),
	}

	if err := memory.AddMessage(ctx, userMessage); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save user message",
		})
	}

	// Use RAG to generate response
	ragResult, err := h.ragChain.Query(ctx, req.Message)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate response",
		})
	}

	// Add assistant message to memory
	assistantMessage := &langchain.Message{
		SessionID: req.SessionID,
		Role:      "assistant",
		Content:   ragResult.Answer,
		Timestamp: time.Now(),
	}

	if err := memory.AddMessage(ctx, assistantMessage); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save assistant message",
		})
	}

	messageID := fmt.Sprintf("msg_%d", time.Now().UnixNano())

	response := &ChatResponse{
		Response:   ragResult.Answer,
		SessionID:  req.SessionID,
		MessageID:  messageID,
		Sources:    ragResult.Sources,
		TokensUsed: ragResult.TokensUsed,
		Duration:   time.Since(startTime),
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"rag_sources_count":   len(ragResult.Sources),
			"rag_retrieval_time":  ragResult.RetrievalTime,
			"rag_generation_time": ragResult.GenerationTime,
			"context_length":      ragResult.Metadata["context_length"],
			"model":               ragResult.Metadata["model"],
		},
	}

	return c.JSON(response)
}

// handleStreamingChat handles streaming chat requests
func (h *AIHandler) handleStreamingChat(c *fiber.Ctx, ctx context.Context, req *ChatRequest) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Access-Control-Allow-Origin", "*")

	// Add user message to memory
	memory, err := h.memoryManager.GetMemory("travel_conversation")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get conversation memory",
		})
	}

	userMessage := &langchain.Message{
		SessionID: req.SessionID,
		Role:      "user",
		Content:   req.Message,
		Timestamp: time.Now(),
	}

	if err := memory.AddMessage(ctx, userMessage); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save user message",
		})
	}

	messageID := fmt.Sprintf("msg_%d", time.Now().UnixNano())

	// First, get context from RAG
	ragResult, err := h.ragChain.Query(ctx, req.Message)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve context",
		})
	}

	// Stream the response
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		// Send initial metadata
		metadataChunk := &StreamChunk{
			SessionID: req.SessionID,
			MessageID: messageID,
			Done:      false,
			Metadata: map[string]interface{}{
				"sources_count": len(ragResult.Sources),
				"sources":       ragResult.Sources,
			},
		}

		data, _ := json.Marshal(metadataChunk)
		fmt.Fprintf(w, "data: %s\n\n", data)
		w.Flush()

		// Stream the actual response content
		content := ragResult.Answer
		words := splitIntoWords(content)

		for i, word := range words {
			chunk := &StreamChunk{
				Content:   word + " ",
				SessionID: req.SessionID,
				MessageID: messageID,
				Done:      i == len(words)-1,
			}

			data, _ := json.Marshal(chunk)
			fmt.Fprintf(w, "data: %s\n\n", data)
			w.Flush()

			// Simulate streaming delay
			time.Sleep(50 * time.Millisecond)
		}

		// Send final chunk
		finalChunk := &StreamChunk{
			SessionID: req.SessionID,
			MessageID: messageID,
			Done:      true,
			Metadata: map[string]interface{}{
				"tokens_used": ragResult.TokensUsed,
				"duration":    ragResult.Duration,
			},
		}

		data, _ = json.Marshal(finalChunk)
		fmt.Fprintf(w, "data: %s\n\n", data)
		w.Flush()
	})

	// Add assistant message to memory
	assistantMessage := &langchain.Message{
		SessionID: req.SessionID,
		Role:      "assistant",
		Content:   ragResult.Answer,
		Timestamp: time.Now(),
	}

	if err := memory.AddMessage(ctx, assistantMessage); err != nil {
		// Log error but don't fail the response
		fmt.Printf("Failed to save assistant message: %v\n", err)
	}

	return nil
}

// AgentRequest represents a request to a specialist agent
type AgentRequest struct {
	AgentType  string                 `json:"agent_type" validate:"required"`
	Query      string                 `json:"query" validate:"required"`
	SessionID  string                 `json:"session_id" validate:"required"`
	UserID     string                 `json:"user_id"`
	Parameters map[string]interface{} `json:"parameters"`
	Stream     bool                   `json:"stream"`
}

// AgentResponse represents a response from a specialist agent
type AgentResponse struct {
	AgentType   string                 `json:"agent_type"`
	Response    string                 `json:"response"`
	SessionID   string                 `json:"session_id"`
	RequestID   string                 `json:"request_id"`
	Confidence  float64                `json:"confidence"`
	Suggestions []string               `json:"suggestions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	Duration    time.Duration          `json:"duration"`
	Timestamp   time.Time              `json:"timestamp"`
}

// ProcessAgentRequest handles requests to specialist agents
func (h *AIHandler) ProcessAgentRequest(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "ai_handler.process_agent_request")
	defer span.End()

	var req AgentRequest
	if err := c.BodyParser(&req); err != nil {
		span.RecordError(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	span.SetAttributes(
		attribute.String("agent_type", req.AgentType),
		attribute.String("session_id", req.SessionID),
		attribute.String("user_id", req.UserID),
	)

	// Select appropriate agent
	var agent specialist.Agent
	switch req.AgentType {
	case "flight":
		agent = h.flightAgent
	case "hotel":
		agent = h.hotelAgent
	case "itinerary":
		agent = h.itineraryAgent
	case "supervisor":
		agent = h.supervisorAgent
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Unknown agent type: %s", req.AgentType),
		})
	}

	// Create agent request
	agentReq := &specialist.AgentRequest{
		ID:         fmt.Sprintf("req_%d", time.Now().UnixNano()),
		UserID:     req.UserID,
		SessionID:  req.SessionID,
		AgentType:  req.AgentType,
		Query:      req.Query,
		Parameters: req.Parameters,
		CreatedAt:  time.Now(),
	}

	// Process request
	startTime := time.Now()
	agentResp, err := agent.ProcessRequest(ctx, agentReq)
	if err != nil {
		span.RecordError(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process agent request",
		})
	}

	response := &AgentResponse{
		AgentType:   req.AgentType,
		Response:    fmt.Sprintf("%v", agentResp.Result),
		SessionID:   req.SessionID,
		RequestID:   agentReq.ID,
		Confidence:  agentResp.Confidence,
		Suggestions: agentResp.Suggestions,
		Duration:    time.Since(startTime),
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"agent_status":   agentResp.Status,
			"agent_duration": agentResp.Duration,
			"capabilities":   agent.GetCapabilities(),
		},
	}

	return c.JSON(response)
}

// RAGQueryRequest represents a RAG query request
type RAGQueryRequest struct {
	Query       string                 `json:"query" validate:"required"`
	Destination string                 `json:"destination"`
	Category    string                 `json:"category"`
	Filter      map[string]interface{} `json:"filter"`
	Limit       int                    `json:"limit"`
}

// RAGQueryResponse represents a RAG query response
type RAGQueryResponse struct {
	Query    string                 `json:"query"`
	Answer   string                 `json:"answer"`
	Sources  []*rag.RetrievalResult `json:"sources"`
	Context  string                 `json:"context"`
	Metadata map[string]interface{} `json:"metadata"`
	Duration time.Duration          `json:"duration"`
}

// QueryKnowledgeBase handles RAG queries to the travel knowledge base
func (h *AIHandler) QueryKnowledgeBase(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "ai_handler.query_knowledge_base")
	defer span.End()

	var req RAGQueryRequest
	if err := c.BodyParser(&req); err != nil {
		span.RecordError(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	span.SetAttributes(
		attribute.String("query", req.Query),
		attribute.String("destination", req.Destination),
		attribute.String("category", req.Category),
	)

	var result *rag.RAGResult
	var err error

	// Choose appropriate query method based on request
	if req.Destination != "" {
		result, err = h.ragChain.QueryDestination(ctx, req.Destination, req.Query)
	} else if req.Category != "" {
		result, err = h.ragChain.QueryByCategory(ctx, req.Category, req.Query)
	} else if req.Filter != nil {
		result, err = h.ragChain.QueryWithFilter(ctx, req.Query, req.Filter)
	} else {
		result, err = h.ragChain.Query(ctx, req.Query)
	}

	if err != nil {
		span.RecordError(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to query knowledge base",
		})
	}

	response := &RAGQueryResponse{
		Query:    req.Query,
		Answer:   result.Answer,
		Sources:  result.Sources,
		Context:  result.Context,
		Duration: result.Duration,
		Metadata: result.Metadata,
	}

	return c.JSON(response)
}

// ConversationHistoryRequest represents a request for conversation history
type ConversationHistoryRequest struct {
	SessionID string `json:"session_id" validate:"required"`
	Limit     int    `json:"limit"`
}

// ConversationHistoryResponse represents conversation history
type ConversationHistoryResponse struct {
	SessionID string                 `json:"session_id"`
	Messages  []*langchain.Message   `json:"messages"`
	Summary   string                 `json:"summary"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// GetConversationHistory retrieves conversation history for a session
func (h *AIHandler) GetConversationHistory(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "ai_handler.get_conversation_history")
	defer span.End()

	sessionID := c.Params("sessionId")
	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Session ID is required",
		})
	}

	limitStr := c.Query("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	span.SetAttributes(
		attribute.String("session_id", sessionID),
		attribute.Int("limit", limit),
	)

	memory, err := h.memoryManager.GetMemory("travel_conversation")
	if err != nil {
		span.RecordError(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get conversation memory",
		})
	}

	messages, err := memory.GetMessages(ctx, sessionID, limit)
	if err != nil {
		span.RecordError(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve conversation history",
		})
	}

	summary, err := memory.GetSummary(ctx, sessionID)
	if err != nil {
		summary = "No summary available"
	}

	response := &ConversationHistoryResponse{
		SessionID: sessionID,
		Messages:  messages,
		Summary:   summary,
		Metadata: map[string]interface{}{
			"message_count": len(messages),
			"memory_type":   memory.GetName(),
		},
	}

	return c.JSON(response)
}

// ClearConversationRequest represents a request to clear conversation
type ClearConversationRequest struct {
	SessionID string `json:"session_id" validate:"required"`
}

// ClearConversation clears conversation history for a session
func (h *AIHandler) ClearConversation(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "ai_handler.clear_conversation")
	defer span.End()

	var req ClearConversationRequest
	if err := c.BodyParser(&req); err != nil {
		span.RecordError(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	span.SetAttributes(attribute.String("session_id", req.SessionID))

	memory, err := h.memoryManager.GetMemory("travel_conversation")
	if err != nil {
		span.RecordError(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get conversation memory",
		})
	}

	if err := memory.Clear(ctx, req.SessionID); err != nil {
		span.RecordError(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to clear conversation",
		})
	}

	return c.JSON(fiber.Map{
		"message":    "Conversation cleared successfully",
		"session_id": req.SessionID,
	})
}

// Helper function to split text into words
func splitIntoWords(text string) []string {
	words := []string{}
	current := ""

	for _, char := range text {
		if char == ' ' || char == '\n' || char == '\t' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		words = append(words, current)
	}

	return words
}

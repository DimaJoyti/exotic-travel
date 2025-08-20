package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/exotic-travel-booking/backend/internal/api/handlers"
	"github.com/exotic-travel-booking/backend/internal/api/routes"
	"github.com/exotic-travel-booking/backend/internal/langchain"
	"github.com/exotic-travel-booking/backend/internal/langgraph"
	"github.com/exotic-travel-booking/backend/internal/llm/providers"
	"github.com/exotic-travel-booking/backend/internal/rag"
	"github.com/exotic-travel-booking/backend/internal/services"
	"github.com/exotic-travel-booking/backend/internal/tools"
	"github.com/gofiber/fiber/v2"
)

func main() {
	fmt.Println("🚀 Starting Exotic Travel Booking AI Server")
	fmt.Println("===========================================")

	ctx := context.Background()

	// Initialize core components
	fmt.Println("🔧 Initializing core components...")

	// 1. OLAMA Service
	ollamaService := services.NewOllamaService("http://localhost:11434")
	if err := ollamaService.CheckHealth(ctx); err != nil {
		log.Printf("⚠️  OLAMA not available: %v", err)
	} else {
		fmt.Println("✅ OLAMA service connected")
	}

	// 2. LLM Provider (use mock for demo if OLAMA not available)
	llmConfig := &providers.LLMConfig{
		Provider: "ollama",
		Model:    "llama3.2",
		BaseURL:  "http://localhost:11434",
		Timeout:  60 * time.Second,
	}

	llmProvider, err := providers.NewOllamaProvider(llmConfig)
	if err != nil {
		log.Printf("⚠️  OLAMA provider failed, using mock: %v", err)
		// Use mock provider for demo
		llmProvider = &MockLLMProvider{}
	}
	fmt.Println("✅ LLM provider initialized")

	// 3. Embedding Service and Vector Store (use mock for demo)
	embeddingService := rag.NewMockEmbeddingService(384, "mock-embed")

	vectorStore := rag.NewMemoryVectorStore(embeddingService)
	fmt.Println("✅ Vector store initialized")

	// 4. Load sample travel knowledge
	if err := loadSampleKnowledge(ctx, vectorStore); err != nil {
		log.Printf("⚠️  Failed to load sample knowledge: %v", err)
	} else {
		fmt.Println("✅ Sample travel knowledge loaded")
	}

	// 5. RAG System
	retriever := rag.NewTravelRetriever(vectorStore)
	ragConfig := &rag.RAGConfig{
		MaxContextLength:   4000,
		RetrievalLimit:     5,
		RelevanceThreshold: 0.1,
		IncludeSourceInfo:  true,
		MaxTokens:          1000,
		Temperature:        0.7,
	}
	ragChain := rag.NewTravelRAGChain(retriever, llmProvider, ragConfig)
	fmt.Println("✅ RAG system initialized")

	// 6. Memory Management
	memoryManager := langchain.NewMemoryManager()
	bufferMemory := langchain.NewBufferMemory("travel_conversation", 50)
	memoryManager.RegisterMemory(bufferMemory)
	fmt.Println("✅ Memory management initialized")

	// 7. State Management
	stateManager := langgraph.NewMemoryStateManager()
	fmt.Println("✅ State management initialized")

	// 8. Tool Registry
	toolRegistry := tools.NewToolRegistry()
	fmt.Println("✅ Tool registry initialized")

	// 9. Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Exotic Travel AI API",
		ServerHeader: "Exotic Travel AI",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	})

	// 10. Setup basic middleware
	// Note: Using basic Fiber middleware for now

	// 11. Create handlers
	aiHandler := handlers.NewAIHandler(
		llmProvider,
		ragChain,
		memoryManager,
		stateManager,
		toolRegistry,
		ollamaService,
	)

	healthHandler := handlers.NewHealthHandler(ollamaService, llmProvider)

	fmt.Println("✅ Handlers initialized")

	// 12. Setup routes
	routes.SetupAIRoutes(app, aiHandler, healthHandler)
	routes.SetupWebSocketRoutes(app, aiHandler)

	// Root endpoint
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":   "Exotic Travel Booking AI API",
			"version":   "1.0.0",
			"status":    "running",
			"timestamp": time.Now(),
			"endpoints": []string{
				"GET /api/v1/health",
				"POST /api/v1/ai/chat",
				"POST /api/v1/ai/knowledge/query",
				"POST /api/v1/ai/agents/request",
				"GET /api/v1/ai/chat/history/:sessionId",
			},
		})
	})

	fmt.Println("✅ Routes configured")

	// 13. Start server
	port := getEnv("PORT", "8081")
	fmt.Printf("🌐 Server starting on port %s\n", port)
	fmt.Println("\n📋 Available Endpoints:")
	fmt.Println("   GET  /                              - API information")
	fmt.Println("   GET  /api/v1/health                 - Health check")
	fmt.Println("   POST /api/v1/ai/chat                - AI chat")
	fmt.Println("   POST /api/v1/ai/knowledge/query     - Knowledge base query")
	fmt.Println("   POST /api/v1/ai/agents/request      - Specialist agent request")
	fmt.Println("   GET  /api/v1/ai/chat/history/:id    - Conversation history")

	// Graceful shutdown
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("❌ Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\n🛑 Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
	}

	fmt.Println("✅ Server shutdown complete")
}

// loadSampleKnowledge loads sample travel knowledge into the vector store
func loadSampleKnowledge(ctx context.Context, vectorStore rag.VectorStore) error {
	sampleDocs := []*rag.Document{
		{
			ID:      "paris_guide",
			Content: "Paris, the City of Light, is France's capital and most populous city. Famous for the Eiffel Tower, Louvre Museum, Notre-Dame Cathedral, and Champs-Élysées. Best visited in spring (April-June) or fall (September-November). Must-try: croissants, macarons, and wine. Budget: €100-200 per day for mid-range travel.",
			Metadata: map[string]interface{}{
				"destination":   "Paris",
				"country":       "France",
				"content_type":  "city_guide",
				"document_type": "travel_guide",
				"themes":        []string{"culture", "history", "food", "art"},
			},
		},
		{
			ID:      "tokyo_food",
			Content: "Tokyo offers incredible culinary experiences from street food to Michelin-starred restaurants. Must-try: sushi at Tsukiji Outer Market, ramen in Shibuya, tempura, wagyu beef, and traditional kaiseki. Food tours available. Budget: ¥3000-8000 per meal. Best food districts: Shibuya, Harajuku, Ginza, and Shinjuku.",
			Metadata: map[string]interface{}{
				"destination":   "Tokyo",
				"country":       "Japan",
				"content_type":  "food",
				"document_type": "travel_guide",
				"themes":        []string{"food", "culture", "local_experience"},
			},
		},
		{
			ID:      "bali_beaches",
			Content: "Bali's beaches offer something for everyone. Kuta Beach for surfing and nightlife, Seminyak for luxury resorts and beach clubs, Uluwatu for dramatic cliffs and world-class surf breaks, and Sanur for family-friendly activities. Best time: April-October (dry season). Activities: surfing, snorkeling, beach clubs, sunset viewing, and water sports.",
			Metadata: map[string]interface{}{
				"destination":   "Bali",
				"country":       "Indonesia",
				"content_type":  "beaches",
				"document_type": "travel_guide",
				"themes":        []string{"beaches", "surfing", "relaxation", "adventure", "luxury"},
			},
		},
		{
			ID:      "iceland_nature",
			Content: "Iceland is a land of fire and ice with stunning natural wonders. See the Northern Lights (September-March), Blue Lagoon geothermal spa, Gullfoss waterfall, Geysir hot springs, and Jökulsárlón glacier lagoon. Ring Road trip highly recommended. Best time: June-August for hiking and midnight sun, September-March for Northern Lights. Budget: $150-300 per day.",
			Metadata: map[string]interface{}{
				"destination":   "Iceland",
				"country":       "Iceland",
				"content_type":  "nature",
				"document_type": "travel_guide",
				"themes":        []string{"nature", "adventure", "photography", "unique_experiences", "northern_lights"},
			},
		},
		{
			ID:      "santorini_romance",
			Content: "Santorini, Greece is perfect for romantic getaways and honeymoons. Famous for white-washed buildings with blue domes, spectacular sunsets in Oia, and volcanic beaches. Wine tasting tours, luxury cave hotels, and intimate cliffside restaurants. Best time: April-October. Perfect for couples seeking romance, luxury, and stunning views.",
			Metadata: map[string]interface{}{
				"destination":   "Santorini",
				"country":       "Greece",
				"content_type":  "romance",
				"document_type": "travel_guide",
				"themes":        []string{"romantic", "luxury", "wine", "sunset_views", "honeymoon"},
			},
		},
		{
			ID:      "new_york_city",
			Content: "New York City, the Big Apple, offers endless attractions and experiences. Visit Times Square, Central Park, Statue of Liberty, Empire State Building, and Broadway shows. World-class museums like MoMA and Metropolitan Museum. Diverse neighborhoods: SoHo, Greenwich Village, Chinatown. Best time: April-June and September-November. Budget: $200-400 per day.",
			Metadata: map[string]interface{}{
				"destination":   "New York City",
				"country":       "USA",
				"content_type":  "city_guide",
				"document_type": "travel_guide",
				"themes":        []string{"urban", "culture", "entertainment", "shopping", "museums"},
			},
		},
	}

	return vectorStore.AddDocuments(ctx, sampleDocs)
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// MockLLMProvider for demo when OLAMA is not available
type MockLLMProvider struct{}

func (m *MockLLMProvider) GetName() string {
	return "mock-llm"
}

func (m *MockLLMProvider) GenerateResponse(ctx context.Context, req *providers.GenerateRequest) (*providers.GenerateResponse, error) {
	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	return &providers.GenerateResponse{
		Choices: []providers.Choice{
			{
				Message: providers.Message{
					Role:    "assistant",
					Content: "This is a mock response for the AI travel assistant. I can help you with travel planning, destination recommendations, and booking assistance.",
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
						Content: "Mock stream response for travel assistance",
					},
				},
			},
			Done: true,
		}
	}()
	return ch, nil
}

func (m *MockLLMProvider) GetModels(ctx context.Context) ([]string, error) {
	return []string{"mock-travel-model"}, nil
}

func (m *MockLLMProvider) Close() error {
	return nil
}

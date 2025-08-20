package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/exotic-travel-booking/backend/internal/llm/providers"
	"github.com/exotic-travel-booking/backend/internal/rag"
)

func main() {
	fmt.Println("🔍 RAG (Retrieval-Augmented Generation) System Demo")
	fmt.Println("==================================================")

	ctx := context.Background()

	// Demo 1: Vector Store and Embeddings
	fmt.Println("\n1. Vector Store and Embeddings Demo")
	fmt.Println("===================================")

	if err := demonstrateVectorStore(ctx); err != nil {
		log.Printf("❌ Vector store demo failed: %v", err)
	}

	// Demo 2: Document Loading and Processing
	fmt.Println("\n2. Document Loading and Processing Demo")
	fmt.Println("======================================")

	if err := demonstrateDocumentLoading(ctx); err != nil {
		log.Printf("❌ Document loading demo failed: %v", err)
	}

	// Demo 3: Semantic Search and Retrieval
	fmt.Println("\n3. Semantic Search and Retrieval Demo")
	fmt.Println("=====================================")

	if err := demonstrateRetrieval(ctx); err != nil {
		log.Printf("❌ Retrieval demo failed: %v", err)
	}

	// Demo 4: RAG Chain - Retrieval + Generation
	fmt.Println("\n4. RAG Chain - Retrieval + Generation Demo")
	fmt.Println("==========================================")

	if err := demonstrateRAGChain(ctx); err != nil {
		log.Printf("❌ RAG chain demo failed: %v", err)
	}

	// Demo 5: Travel-Specific RAG
	fmt.Println("\n5. Travel-Specific RAG Demo")
	fmt.Println("===========================")

	if err := demonstrateTravelRAG(ctx); err != nil {
		log.Printf("❌ Travel RAG demo failed: %v", err)
	}

	fmt.Println("\n🎉 RAG System Demo Completed Successfully!")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("✅ Vector embeddings and similarity search")
	fmt.Println("✅ Document loading and text splitting")
	fmt.Println("✅ Semantic retrieval with ranking")
	fmt.Println("✅ Retrieval-augmented generation")
	fmt.Println("✅ Travel-specific knowledge base")
	fmt.Println("✅ Context-aware question answering")
}

func demonstrateVectorStore(ctx context.Context) error {
	fmt.Println("   🧮 Creating embedding service and vector store...")

	// Create mock embedding service for demo
	embeddingService := rag.NewMockEmbeddingService(384, "nomic-embed-text")
	vectorStore := rag.NewMemoryVectorStore(embeddingService)

	fmt.Printf("   ✅ Embedding service created (dimension: %d, model: %s)\n",
		embeddingService.GetDimension(), embeddingService.GetModel())

	// Add sample travel documents
	travelDocs := []*rag.Document{
		{
			ID:      "paris_guide",
			Content: "Paris, the City of Light, is France's capital and most populous city. Famous for the Eiffel Tower, Louvre Museum, Notre-Dame Cathedral, and Champs-Élysées. Best visited in spring (April-June) or fall (September-November). Must-try: croissants, macarons, and wine. Budget: €100-200 per day.",
			Metadata: map[string]interface{}{
				"destination":   "Paris",
				"country":       "France",
				"content_type":  "city_guide",
				"document_type": "travel_guide",
				"themes":        []string{"culture", "history", "food"},
			},
		},
		{
			ID:      "tokyo_food",
			Content: "Tokyo offers incredible culinary experiences from street food to Michelin-starred restaurants. Must-try: sushi at Tsukiji Market, ramen in Shibuya, tempura, and wagyu beef. Food tours available. Budget: ¥3000-8000 per meal. Best food districts: Shibuya, Harajuku, Ginza.",
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
			Content: "Bali's beaches offer something for everyone. Kuta Beach for surfing and nightlife, Seminyak for luxury resorts, Uluwatu for dramatic cliffs, and Sanur for family-friendly activities. Best time: April-October (dry season). Activities: surfing, snorkeling, beach clubs, sunset viewing.",
			Metadata: map[string]interface{}{
				"destination":   "Bali",
				"country":       "Indonesia",
				"content_type":  "beaches",
				"document_type": "travel_guide",
				"themes":        []string{"beaches", "surfing", "relaxation", "adventure"},
			},
		},
		{
			ID:      "iceland_nature",
			Content: "Iceland is a land of fire and ice with stunning natural wonders. See the Northern Lights (September-March), Blue Lagoon geothermal spa, Gullfoss waterfall, and Geysir hot springs. Ring Road trip recommended. Best time: June-August for hiking, September-March for Northern Lights.",
			Metadata: map[string]interface{}{
				"destination":   "Iceland",
				"country":       "Iceland",
				"content_type":  "nature",
				"document_type": "travel_guide",
				"themes":        []string{"nature", "adventure", "photography", "unique_experiences"},
			},
		},
		{
			ID:      "santorini_romance",
			Content: "Santorini, Greece is perfect for romantic getaways. Famous for white-washed buildings, blue domes, and spectacular sunsets in Oia. Wine tasting tours, luxury resorts, and intimate restaurants. Best time: April-October. Perfect for honeymoons and anniversaries.",
			Metadata: map[string]interface{}{
				"destination":   "Santorini",
				"country":       "Greece",
				"content_type":  "romance",
				"document_type": "travel_guide",
				"themes":        []string{"romantic", "luxury", "wine", "sunset_views"},
			},
		},
	}

	fmt.Println("   📚 Adding travel documents to vector store...")
	err := vectorStore.AddDocuments(ctx, travelDocs)
	if err != nil {
		return fmt.Errorf("failed to add documents: %w", err)
	}

	fmt.Printf("   ✅ Added %d travel documents\n", len(travelDocs))

	// Test similarity search
	fmt.Println("   🔍 Testing similarity search...")

	searchQueries := []string{
		"romantic sunset destinations",
		"best food experiences in Asia",
		"natural wonders and outdoor activities",
	}

	for _, query := range searchQueries {
		results, err := vectorStore.SearchByText(ctx, query, 2, 0.1)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		fmt.Printf("   🎯 Query: '%s'\n", query)
		for i, result := range results {
			fmt.Printf("      %d. %s (similarity: %.3f) - %s\n",
				i+1, result.Document.Metadata["destination"], result.Similarity,
				truncateString(result.Document.Content, 60))
		}
	}

	// Show vector store statistics
	stats, err := vectorStore.GetStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	fmt.Printf("   📊 Vector store stats: %d documents, %d dimensions, %d bytes\n",
		stats.DocumentCount, stats.VectorDim, stats.StorageSize)

	return nil
}

func demonstrateDocumentLoading(ctx context.Context) error {
	fmt.Println("   📄 Testing document loading and text splitting...")

	// Create text splitter
	textSplitter := rag.NewRecursiveCharacterTextSplitter(200, 50)

	// Create string document loader
	loader := rag.NewStringDocumentLoader(textSplitter)

	// Sample long travel content
	longContent := `
	Complete Guide to Visiting Japan

	Japan is an incredible destination that offers a perfect blend of ancient traditions and modern innovation. From bustling cities like Tokyo and Osaka to serene temples in Kyoto, there's something for every traveler.

	Transportation: The JR Pass is essential for traveling between cities. Local trains and subways are efficient and punctual. Consider getting a Suica or Pasmo card for local transportation.

	Accommodation: Options range from traditional ryokans to modern hotels. Capsule hotels offer a unique experience for budget travelers. Book early during cherry blossom season (March-May).

	Food: Japanese cuisine is diverse and delicious. Try sushi, ramen, tempura, and local specialties in each region. Don't miss the food markets and street food scenes.

	Culture: Respect local customs, bow when greeting, remove shoes when entering homes or temples. Learn basic Japanese phrases - locals appreciate the effort.

	Best Time to Visit: Spring (March-May) for cherry blossoms, autumn (September-November) for fall colors, or winter (December-February) for snow activities.
	`

	metadata := map[string]interface{}{
		"destination": "Japan",
		"type":        "comprehensive_guide",
		"author":      "Travel Expert",
	}

	fmt.Println("   ✂️  Splitting long document into chunks...")

	// Load and split document
	doc := loader.LoadFromString(longContent, metadata)
	chunks := textSplitter.SplitDocuments([]*rag.Document{doc})

	fmt.Printf("   ✅ Split document into %d chunks:\n", len(chunks))

	for i, chunk := range chunks {
		fmt.Printf("      Chunk %d: %s... (%d chars)\n",
			i+1, truncateString(chunk.Content, 50), len(chunk.Content))
	}

	// Test travel document loader
	fmt.Println("   🌍 Testing travel-specific document processing...")

	travelContent := "Visit beautiful Santorini, Greece for romantic sunsets and luxury wine tasting. The island offers stunning beaches, traditional villages, and world-class restaurants. Perfect for honeymoons and special occasions."

	// Create a string loader for the travel content
	stringLoader := rag.NewStringDocumentLoader(nil)
	enhancedDoc := stringLoader.LoadFromString(travelContent, map[string]interface{}{
		"source":      "demo",
		"destination": "Santorini",
		"type":        "travel_guide",
	})

	fmt.Printf("   ✅ Document loaded with metadata:\n")
	fmt.Printf("      Document ID: %s\n", enhancedDoc.ID)
	fmt.Printf("      Content length: %d\n", enhancedDoc.Metadata["content_length"])
	fmt.Printf("      Loader type: %s\n", enhancedDoc.Metadata["loader_type"])
	fmt.Printf("      Destination: %s\n", enhancedDoc.Metadata["destination"])

	return nil
}

func demonstrateRetrieval(ctx context.Context) error {
	fmt.Println("   🎯 Setting up retrieval system...")

	// Setup vector store with travel data
	embeddingService := rag.NewMockEmbeddingService(384, "nomic-embed-text")
	vectorStore := rag.NewMemoryVectorStore(embeddingService)

	// Add diverse travel documents
	travelDocs := []*rag.Document{
		{
			ID:      "paris_budget",
			Content: "Budget travel in Paris: Stay in hostels (€25-40/night), eat at bistros and markets, use metro day passes, visit free museums on first Sundays. Total budget: €50-80/day.",
			Metadata: map[string]interface{}{
				"destination":   "Paris",
				"content_type":  "budget",
				"document_type": "travel_guide",
			},
		},
		{
			ID:      "paris_luxury",
			Content: "Luxury Paris experience: Stay at Le Meurice or Ritz, dine at Michelin-starred restaurants, private Louvre tours, shopping on Champs-Élysées. Budget: €500-1000/day.",
			Metadata: map[string]interface{}{
				"destination":   "Paris",
				"content_type":  "luxury",
				"document_type": "travel_guide",
			},
		},
		{
			ID:      "tokyo_culture",
			Content: "Tokyo cultural experiences: Visit Senso-ji Temple, participate in tea ceremony, watch sumo wrestling, explore traditional neighborhoods like Asakusa and Ueno.",
			Metadata: map[string]interface{}{
				"destination":   "Tokyo",
				"content_type":  "culture",
				"document_type": "travel_guide",
			},
		},
	}

	err := vectorStore.AddDocuments(ctx, travelDocs)
	if err != nil {
		return fmt.Errorf("failed to add documents: %w", err)
	}

	// Create retrievers
	basicRetriever := rag.NewVectorStoreRetriever(vectorStore)
	travelRetriever := rag.NewTravelRetriever(vectorStore)

	fmt.Println("   🔍 Testing basic retrieval...")

	// Test basic retrieval
	query := "affordable travel options in Europe"
	results, err := basicRetriever.Retrieve(ctx, query, 3)
	if err != nil {
		return fmt.Errorf("retrieval failed: %w", err)
	}

	fmt.Printf("   📋 Query: '%s'\n", query)
	fmt.Printf("   📊 Found %d results:\n", len(results))

	for i, result := range results {
		fmt.Printf("      %d. %s - %s (relevance: %.3f)\n",
			i+1, result.Document.Metadata["destination"],
			result.Document.Metadata["content_type"], result.Relevance)
		fmt.Printf("         %s\n", result.Explanation)
	}

	fmt.Println("   🌍 Testing travel-specific retrieval...")

	// Test destination-specific retrieval
	destResults, err := travelRetriever.RetrieveForDestination(ctx, "Paris", "luxury experiences", 2)
	if err != nil {
		return fmt.Errorf("destination retrieval failed: %w", err)
	}

	fmt.Printf("   🎯 Destination query: 'luxury experiences in Paris'\n")
	fmt.Printf("   📊 Found %d destination-specific results:\n", len(destResults))

	for i, result := range destResults {
		fmt.Printf("      %d. %s (score: %.3f)\n",
			i+1, result.Document.Metadata["content_type"], result.Score)
		fmt.Printf("         %s\n", truncateString(result.Document.Content, 80))
	}

	// Test category-based retrieval
	categoryResults, err := travelRetriever.RetrieveByCategory(ctx, "culture", "traditional experiences", 2)
	if err != nil {
		return fmt.Errorf("category retrieval failed: %w", err)
	}

	fmt.Printf("   🏛️  Category query: 'traditional experiences' in 'culture' category\n")
	fmt.Printf("   📊 Found %d category results:\n", len(categoryResults))

	for i, result := range categoryResults {
		fmt.Printf("      %d. %s - %s\n",
			i+1, result.Document.Metadata["destination"],
			truncateString(result.Document.Content, 60))
	}

	return nil
}

func demonstrateRAGChain(ctx context.Context) error {
	fmt.Println("   🔗 Setting up RAG chain...")

	// Setup components
	embeddingService := rag.NewMockEmbeddingService(384, "nomic-embed-text")
	vectorStore := rag.NewMemoryVectorStore(embeddingService)
	retriever := rag.NewVectorStoreRetriever(vectorStore)

	// Add comprehensive travel knowledge
	knowledgeBase := []*rag.Document{
		{
			ID:      "japan_overview",
			Content: "Japan is an island nation in East Asia known for its rich culture, advanced technology, and delicious cuisine. Major cities include Tokyo, Osaka, and Kyoto. The country offers everything from ancient temples to modern skyscrapers, making it perfect for diverse travel experiences.",
			Metadata: map[string]interface{}{
				"destination": "Japan",
				"type":        "overview",
			},
		},
		{
			ID:      "japan_seasons",
			Content: "Japan has four distinct seasons. Spring (March-May) features cherry blossoms and mild weather. Summer (June-August) is hot and humid with festivals. Autumn (September-November) offers beautiful fall colors. Winter (December-February) brings snow and winter sports opportunities.",
			Metadata: map[string]interface{}{
				"destination": "Japan",
				"type":        "seasonal_guide",
			},
		},
		{
			ID:      "japan_transportation",
			Content: "Japan has an excellent transportation system. The JR Pass allows unlimited travel on JR trains for 7, 14, or 21 days. Shinkansen (bullet trains) connect major cities quickly. Local trains and subways are punctual and efficient. IC cards like Suica make local travel convenient.",
			Metadata: map[string]interface{}{
				"destination": "Japan",
				"type":        "transportation",
			},
		},
	}

	err := vectorStore.AddDocuments(ctx, knowledgeBase)
	if err != nil {
		return fmt.Errorf("failed to add knowledge base: %w", err)
	}

	// Create mock LLM provider
	mockLLM := &MockLLMProvider{}

	// Create RAG chain
	config := &rag.RAGConfig{
		MaxContextLength:   2000,
		RetrievalLimit:     3,
		RelevanceThreshold: 0.1,
		IncludeSourceInfo:  true,
		MaxTokens:          500,
		Temperature:        0.7,
	}

	ragChain := rag.NewRAGChain(retriever, mockLLM, config)

	fmt.Println("   🤖 Testing RAG question answering...")

	// Test questions
	questions := []string{
		"What's the best time to visit Japan?",
		"How do I get around Japan?",
		"What makes Japan a unique travel destination?",
	}

	for _, question := range questions {
		fmt.Printf("   ❓ Question: %s\n", question)

		result, err := ragChain.Query(ctx, question)
		if err != nil {
			return fmt.Errorf("RAG query failed: %w", err)
		}

		fmt.Printf("   🤖 Answer: %s\n", result.Answer)
		fmt.Printf("   📊 Sources used: %d, Tokens: %d, Duration: %v\n",
			len(result.Sources), result.TokensUsed, result.Duration)
		fmt.Printf("   ⏱️  Retrieval: %v, Generation: %v\n",
			result.RetrievalTime, result.GenerationTime)
		fmt.Println()
	}

	return nil
}

func demonstrateTravelRAG(ctx context.Context) error {
	fmt.Println("   ✈️  Setting up travel-specific RAG system...")

	// Setup travel knowledge base
	embeddingService := rag.NewMockEmbeddingService(384, "nomic-embed-text")
	vectorStore := rag.NewMemoryVectorStore(embeddingService)
	travelRetriever := rag.NewTravelRetriever(vectorStore)

	// Add destination-specific knowledge
	destinationGuides := []*rag.Document{
		{
			ID:      "bali_complete",
			Content: "Bali, Indonesia is a tropical paradise offering diverse experiences. Ubud for culture and rice terraces, Seminyak for beaches and nightlife, Canggu for surfing, and Uluwatu for dramatic cliffs. Best time: April-October (dry season). Activities include temple visits, volcano hikes, cooking classes, and spa treatments.",
			Metadata: map[string]interface{}{
				"destinations":  []string{"Bali", "Indonesia"},
				"document_type": "travel_guide",
				"content_type":  "comprehensive",
				"themes":        []string{"beaches", "culture", "adventure", "relaxation"},
			},
		},
		{
			ID:      "bali_accommodation",
			Content: "Bali accommodation ranges from budget hostels (€10-20/night) to luxury resorts (€200-500/night). Popular areas: Ubud for jungle retreats, Seminyak for beach resorts, Canggu for surf lodges. Many offer yoga classes, spa services, and cultural activities.",
			Metadata: map[string]interface{}{
				"destinations":  []string{"Bali", "Indonesia"},
				"document_type": "travel_guide",
				"content_type":  "accommodation",
				"themes":        []string{"budget", "luxury", "wellness"},
			},
		},
	}

	err := vectorStore.AddDocuments(ctx, destinationGuides)
	if err != nil {
		return fmt.Errorf("failed to add destination guides: %w", err)
	}

	// Create travel RAG chain
	mockLLM := &MockLLMProvider{}
	travelRAG := rag.NewTravelRAGChain(travelRetriever, mockLLM, nil)

	fmt.Println("   🌴 Testing destination-specific queries...")

	// Test destination queries
	destQueries := []struct {
		destination string
		question    string
	}{
		{"Bali", "What are the best areas to stay?"},
		{"Bali", "What activities can I do there?"},
		{"Indonesia", "Tell me about tropical destinations"},
	}

	for _, query := range destQueries {
		fmt.Printf("   🎯 Destination: %s, Question: %s\n", query.destination, query.question)

		result, err := travelRAG.QueryDestination(ctx, query.destination, query.question)
		if err != nil {
			return fmt.Errorf("destination query failed: %w", err)
		}

		fmt.Printf("   ✈️  Travel Advice: %s\n", result.Answer)
		fmt.Printf("   📍 Destination: %s, Method: %s\n",
			result.Metadata["destination"], result.Metadata["retrieval_method"])
		fmt.Println()
	}

	fmt.Println("   🏷️  Testing category-based queries...")

	// Test category queries
	categoryResult, err := travelRAG.QueryByCategory(ctx, "accommodation", "luxury resort options")
	if err != nil {
		return fmt.Errorf("category query failed: %w", err)
	}

	fmt.Printf("   🏨 Category: accommodation, Question: luxury resort options\n")
	fmt.Printf("   💎 Luxury Advice: %s\n", categoryResult.Answer)
	fmt.Printf("   🏷️  Category: %s\n", categoryResult.Metadata["category"])

	return nil
}

// Helper function
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// MockLLMProvider for demo
type MockLLMProvider struct{}

func (m *MockLLMProvider) GetName() string {
	return "mock-travel-llm"
}

func (m *MockLLMProvider) GenerateResponse(ctx context.Context, req *providers.GenerateRequest) (*providers.GenerateResponse, error) {
	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	// Generate contextual response based on the prompt
	prompt := req.Messages[0].Content
	var response string

	if contains(prompt, "best time") || contains(prompt, "season") {
		response = "Based on the travel information provided, the best time to visit depends on your preferences. Spring and autumn generally offer the most pleasant weather with moderate temperatures and beautiful scenery."
	} else if contains(prompt, "transportation") || contains(prompt, "get around") {
		response = "According to the travel guides, the transportation system is excellent with various options including trains, local transit, and passes that make getting around convenient and efficient."
	} else if contains(prompt, "accommodation") || contains(prompt, "stay") {
		response = "The destination offers diverse accommodation options ranging from budget-friendly hostels to luxury resorts, with different areas catering to different preferences and budgets."
	} else if contains(prompt, "activities") || contains(prompt, "do") {
		response = "There are numerous activities available including cultural experiences, outdoor adventures, culinary tours, and relaxation options, providing something for every type of traveler."
	} else {
		response = "Based on the comprehensive travel information provided, this destination offers a unique blend of experiences that cater to diverse interests, making it an excellent choice for travelers seeking both adventure and cultural enrichment."
	}

	return &providers.GenerateResponse{
		Choices: []providers.Choice{
			{
				Message: providers.Message{
					Role:    "assistant",
					Content: response,
				},
			},
		},
		Usage: providers.Usage{
			PromptTokens:     len(prompt) / 4, // Rough token estimate
			CompletionTokens: len(response) / 4,
			TotalTokens:      (len(prompt) + len(response)) / 4,
		},
	}, nil
}

func (m *MockLLMProvider) StreamResponse(ctx context.Context, req *providers.GenerateRequest) (<-chan *providers.StreamChunk, error) {
	// Not implemented for demo
	return nil, fmt.Errorf("streaming not implemented in demo")
}

func (m *MockLLMProvider) GetModels(ctx context.Context) ([]string, error) {
	return []string{"mock-travel-model"}, nil
}

func (m *MockLLMProvider) Close() error {
	return nil
}

func (m *MockLLMProvider) ValidateConfig() error {
	return nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			strings.Contains(strings.ToLower(s), strings.ToLower(substr))))
}

package rag

import (
	"context"
	"testing"
	"time"

	"github.com/exotic-travel-booking/backend/internal/llm/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVectorStore(t *testing.T) {
	ctx := context.Background()

	t.Run("Memory Vector Store", func(t *testing.T) {
		embeddingService := NewMockEmbeddingService(384, "mock-model")
		vectorStore := NewMemoryVectorStore(embeddingService)

		// Test adding documents
		doc1 := &Document{
			ID:      "doc1",
			Content: "Paris is the capital of France and known for the Eiffel Tower.",
			Metadata: map[string]interface{}{
				"destination":   "Paris",
				"country":       "France",
				"content_type":  "city_guide",
				"document_type": "travel_guide",
			},
		}

		err := vectorStore.AddDocument(ctx, doc1)
		require.NoError(t, err)

		// Test retrieving document
		retrieved, err := vectorStore.GetDocument(ctx, "doc1")
		require.NoError(t, err)
		assert.Equal(t, doc1.ID, retrieved.ID)
		assert.Equal(t, doc1.Content, retrieved.Content)
		assert.NotEmpty(t, retrieved.Vector)

		// Test search by text
		results, err := vectorStore.SearchByText(ctx, "Eiffel Tower Paris", 5, 0.1)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "doc1", results[0].Document.ID)
		assert.Greater(t, results[0].Similarity, 0.0)

		// Test stats
		stats, err := vectorStore.GetStats(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, stats.DocumentCount)
		assert.Equal(t, 384, stats.VectorDim)
	})

	t.Run("Document Operations", func(t *testing.T) {
		embeddingService := NewMockEmbeddingService(384, "mock-model")
		vectorStore := NewMemoryVectorStore(embeddingService)

		// Add multiple documents
		docs := []*Document{
			{
				ID:      "tokyo",
				Content: "Tokyo is Japan's capital, famous for sushi, temples, and modern technology.",
				Metadata: map[string]interface{}{
					"destination": "Tokyo",
					"country":     "Japan",
					"themes":      []string{"culture", "food", "technology"},
				},
			},
			{
				ID:      "london",
				Content: "London is the UK capital with Big Ben, Thames River, and rich history.",
				Metadata: map[string]interface{}{
					"destination": "London",
					"country":     "UK",
					"themes":      []string{"history", "culture"},
				},
			},
		}

		err := vectorStore.AddDocuments(ctx, docs)
		require.NoError(t, err)

		// Test filtering
		filter := map[string]interface{}{
			"country": "Japan",
		}

		filtered, err := vectorStore.ListDocuments(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, filtered, 1)
		assert.Equal(t, "tokyo", filtered[0].ID)

		// Test update
		docs[0].Content = "Tokyo is Japan's amazing capital with incredible sushi and temples."
		err = vectorStore.UpdateDocument(ctx, docs[0])
		require.NoError(t, err)

		updated, err := vectorStore.GetDocument(ctx, "tokyo")
		require.NoError(t, err)
		assert.Contains(t, updated.Content, "amazing")

		// Test delete
		err = vectorStore.DeleteDocument(ctx, "london")
		require.NoError(t, err)

		_, err = vectorStore.GetDocument(ctx, "london")
		assert.Error(t, err)
	})
}

func TestEmbeddingService(t *testing.T) {
	ctx := context.Background()

	t.Run("Mock Embedding Service", func(t *testing.T) {
		service := NewMockEmbeddingService(384, "mock-model")

		// Test single embedding
		embedding, err := service.GenerateEmbedding(ctx, "Test text for embedding")
		require.NoError(t, err)
		assert.Len(t, embedding, 384)
		assert.Equal(t, "mock-model", service.GetModel())
		assert.Equal(t, 384, service.GetDimension())

		// Test batch embeddings
		texts := []string{
			"First text",
			"Second text",
			"Third text",
		}

		embeddings, err := service.GenerateEmbeddings(ctx, texts)
		require.NoError(t, err)
		assert.Len(t, embeddings, 3)
		for _, emb := range embeddings {
			assert.Len(t, emb, 384)
		}

		// Test consistency (same text should produce same embedding)
		embedding1, err := service.GenerateEmbedding(ctx, "consistent text")
		require.NoError(t, err)

		embedding2, err := service.GenerateEmbedding(ctx, "consistent text")
		require.NoError(t, err)

		assert.Equal(t, embedding1, embedding2)
	})

	t.Run("Batch Embedding Service", func(t *testing.T) {
		mockService := NewMockEmbeddingService(384, "mock-model")
		batchService := NewBatchEmbeddingService(mockService, 2)

		texts := []string{"text1", "text2", "text3", "text4", "text5"}
		embeddings, err := batchService.GenerateEmbeddings(ctx, texts)
		require.NoError(t, err)
		assert.Len(t, embeddings, 5)
	})
}

func TestDocumentLoader(t *testing.T) {
	t.Run("String Document Loader", func(t *testing.T) {
		textSplitter := NewRecursiveCharacterTextSplitter(100, 20)
		loader := NewStringDocumentLoader(textSplitter)

		content := "This is a test document about travel to Paris. Paris is beautiful with many attractions like the Eiffel Tower, Louvre Museum, and Notre-Dame Cathedral. The city offers great food, culture, and history."

		metadata := map[string]interface{}{
			"destination": "Paris",
			"type":        "guide",
		}

		doc := loader.LoadFromString(content, metadata)
		assert.NotEmpty(t, doc.ID)
		assert.Equal(t, content, doc.Content)
		assert.Equal(t, "Paris", doc.Metadata["destination"])
		assert.Equal(t, "string", doc.Metadata["loader_type"])

		// Test multiple documents
		contents := []string{
			"Guide to Tokyo with sushi and temples.",
			"London travel guide with Big Ben and Thames.",
		}

		metadataList := []map[string]interface{}{
			{"destination": "Tokyo"},
			{"destination": "London"},
		}

		docs := loader.LoadFromStrings(contents, metadataList)
		assert.Len(t, docs, 2)
		assert.Equal(t, "Tokyo", docs[0].Metadata["destination"])
		assert.Equal(t, "London", docs[1].Metadata["destination"])
	})

	t.Run("Text Splitter", func(t *testing.T) {
		splitter := NewRecursiveCharacterTextSplitter(50, 10)

		text := "This is a long text that needs to be split into smaller chunks for better processing. Each chunk should be around 50 characters with some overlap between chunks."

		chunks := splitter.SplitText(text)
		assert.Greater(t, len(chunks), 1)

		// Check that chunks have reasonable sizes
		for _, chunk := range chunks {
			assert.LessOrEqual(t, len(chunk), 80) // Allow some flexibility for overlap
		}

		// Test document splitting
		doc := &Document{
			ID:      "test_doc",
			Content: text,
			Metadata: map[string]interface{}{
				"source": "test",
			},
		}

		splitDocs := splitter.SplitDocuments([]*Document{doc})
		assert.Greater(t, len(splitDocs), 1)

		for i, splitDoc := range splitDocs {
			assert.Contains(t, splitDoc.ID, "chunk")
			assert.Equal(t, i, splitDoc.Metadata["chunk_index"])
			assert.Equal(t, "test_doc", splitDoc.Metadata["parent_document_id"])
		}
	})

	t.Run("Travel Document Loader", func(t *testing.T) {
		textSplitter := NewRecursiveCharacterTextSplitter(200, 50)
		loader := NewTravelDocumentLoader(textSplitter)

		// Test travel metadata enhancement
		doc := &Document{
			ID:       "travel_doc",
			Content:  "Visit beautiful Santorini, Greece for romantic sunsets and luxury resorts. The island offers amazing beaches, traditional villages, and excellent wine tasting experiences.",
			Metadata: make(map[string]interface{}),
		}

		loader.enhanceTravelMetadata(doc)

		assert.Equal(t, "travel_guide", doc.Metadata["document_type"])
		assert.Contains(t, doc.Metadata["themes"], "romantic")

		if destinations, exists := doc.Metadata["destinations"]; exists {
			destList := destinations.([]string)
			assert.Greater(t, len(destList), 0)
		}
	})
}

func TestRetriever(t *testing.T) {
	ctx := context.Background()

	t.Run("Vector Store Retriever", func(t *testing.T) {
		// Setup vector store with test data
		embeddingService := NewMockEmbeddingService(384, "mock-model")
		vectorStore := NewMemoryVectorStore(embeddingService)

		testDocs := []*Document{
			{
				ID:      "paris_guide",
				Content: "Paris travel guide with Eiffel Tower, Louvre, and romantic cafes.",
				Metadata: map[string]interface{}{
					"destination":   "Paris",
					"content_type":  "attractions",
					"document_type": "travel_guide",
				},
			},
			{
				ID:      "tokyo_food",
				Content: "Tokyo food guide featuring sushi, ramen, and traditional Japanese cuisine.",
				Metadata: map[string]interface{}{
					"destination":   "Tokyo",
					"content_type":  "food",
					"document_type": "travel_guide",
				},
			},
			{
				ID:      "london_history",
				Content: "London historical sites including Big Ben, Tower of London, and British Museum.",
				Metadata: map[string]interface{}{
					"destination":   "London",
					"content_type":  "culture",
					"document_type": "travel_guide",
				},
			},
		}

		err := vectorStore.AddDocuments(ctx, testDocs)
		require.NoError(t, err)

		retriever := NewVectorStoreRetriever(vectorStore)

		// Test basic retrieval
		results, err := retriever.Retrieve(ctx, "Eiffel Tower Paris attractions", 2)
		require.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, "paris_guide", results[0].Document.ID)
		assert.Greater(t, results[0].Relevance, 0.0)
		assert.NotEmpty(t, results[0].Explanation)

		// Test filtered retrieval
		filter := map[string]interface{}{
			"content_type": "food",
		}

		filteredResults, err := retriever.RetrieveWithFilter(ctx, "Japanese cuisine", filter, 5)
		require.NoError(t, err)
		assert.Len(t, filteredResults, 1)
		assert.Equal(t, "tokyo_food", filteredResults[0].Document.ID)

		// Test retrieval options
		options := &RetrievalOptions{
			Limit:           3,
			Threshold:       0.0,
			RerankResults:   true,
			IncludeMetadata: true,
		}

		optionResults, err := retriever.GetRelevantDocuments(ctx, "historical sites", options)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(optionResults), 3)
	})

	t.Run("Travel Retriever", func(t *testing.T) {
		embeddingService := NewMockEmbeddingService(384, "mock-model")
		vectorStore := NewMemoryVectorStore(embeddingService)

		// Add travel-specific test data
		travelDoc := &Document{
			ID:      "santorini_guide",
			Content: "Santorini, Greece offers stunning sunsets, white architecture, and volcanic beaches.",
			Metadata: map[string]interface{}{
				"destinations":  []string{"Santorini", "Greece"},
				"document_type": "travel_guide",
				"content_type":  "attractions",
				"themes":        []string{"romantic", "beaches", "architecture"},
			},
		}

		err := vectorStore.AddDocument(ctx, travelDoc)
		require.NoError(t, err)

		travelRetriever := NewTravelRetriever(vectorStore)

		// Test destination-specific retrieval
		results, err := travelRetriever.RetrieveForDestination(ctx, "Santorini", "romantic sunset views", 5)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "santorini_guide", results[0].Document.ID)

		// Test category-based retrieval
		categoryResults, err := travelRetriever.RetrieveByCategory(ctx, "attractions", "beautiful views", 5)
		require.NoError(t, err)
		assert.Len(t, categoryResults, 1)
	})
}

func TestRAGChain(t *testing.T) {
	ctx := context.Background()

	t.Run("Basic RAG Chain", func(t *testing.T) {
		// Setup components
		embeddingService := NewMockEmbeddingService(384, "mock-model")
		vectorStore := NewMemoryVectorStore(embeddingService)
		retriever := NewVectorStoreRetriever(vectorStore)

		// Add test documents
		testDoc := &Document{
			ID:      "paris_info",
			Content: "Paris is the capital of France, famous for the Eiffel Tower, Louvre Museum, and delicious cuisine. Best time to visit is spring or fall.",
			Metadata: map[string]interface{}{
				"destination": "Paris",
				"type":        "city_guide",
			},
		}

		err := vectorStore.AddDocument(ctx, testDoc)
		require.NoError(t, err)

		// Create mock LLM provider
		mockLLM := &MockLLMProvider{}

		// Create RAG chain
		config := &RAGConfig{
			MaxContextLength:   1000,
			RetrievalLimit:     3,
			RelevanceThreshold: 0.1,
			IncludeSourceInfo:  true,
			MaxTokens:          500,
			Temperature:        0.7,
		}

		ragChain := NewRAGChain(retriever, mockLLM, config)

		// Test query
		result, err := ragChain.Query(ctx, "What can you tell me about Paris?")
		require.NoError(t, err)

		assert.Equal(t, "What can you tell me about Paris?", result.Query)
		assert.NotEmpty(t, result.Answer)
		assert.Len(t, result.Sources, 1)
		assert.NotEmpty(t, result.Context)
		assert.Greater(t, result.Duration, time.Duration(0))
		assert.Equal(t, "mock-llm", result.Metadata["model"])
	})

	t.Run("Travel RAG Chain", func(t *testing.T) {
		embeddingService := NewMockEmbeddingService(384, "mock-model")
		vectorStore := NewMemoryVectorStore(embeddingService)
		travelRetriever := NewTravelRetriever(vectorStore)

		// Add travel document
		travelDoc := &Document{
			ID:      "bali_guide",
			Content: "Bali, Indonesia is a tropical paradise with beautiful beaches, ancient temples, and rich culture. Perfect for relaxation and adventure.",
			Metadata: map[string]interface{}{
				"destinations":  []string{"Bali", "Indonesia"},
				"document_type": "travel_guide",
				"content_type":  "general",
				"themes":        []string{"beaches", "culture", "adventure"},
			},
		}

		err := vectorStore.AddDocument(ctx, travelDoc)
		require.NoError(t, err)

		mockLLM := &MockLLMProvider{}
		travelRAG := NewTravelRAGChain(travelRetriever, mockLLM, nil)

		// Test destination query
		result, err := travelRAG.QueryDestination(ctx, "Bali", "What activities are available?")
		require.NoError(t, err)

		assert.NotEmpty(t, result.Answer)
		assert.Equal(t, "Bali", result.Metadata["destination"])
		assert.Equal(t, "destination_specific", result.Metadata["retrieval_method"])

		// Test category query
		categoryResult, err := travelRAG.QueryByCategory(ctx, "general", "tropical destinations")
		require.NoError(t, err)

		assert.NotEmpty(t, categoryResult.Answer)
		assert.Equal(t, "general", categoryResult.Metadata["category"])
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
					Content: "This is a mock response based on the provided context about travel destinations.",
				},
			},
		},
		Usage: providers.Usage{
			PromptTokens:     100,
			CompletionTokens: 50,
			TotalTokens:      150,
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

package rag

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Retriever defines the interface for document retrieval
type Retriever interface {
	// Retrieve retrieves relevant documents for a query
	Retrieve(ctx context.Context, query string, limit int) ([]*RetrievalResult, error)
	
	// RetrieveWithFilter retrieves documents with metadata filtering
	RetrieveWithFilter(ctx context.Context, query string, filter map[string]interface{}, limit int) ([]*RetrievalResult, error)
	
	// GetRelevantDocuments gets documents relevant to a query with scoring
	GetRelevantDocuments(ctx context.Context, query string, options *RetrievalOptions) ([]*RetrievalResult, error)
}

// RetrievalResult represents a retrieved document with relevance information
type RetrievalResult struct {
	Document    *Document `json:"document"`
	Score       float64   `json:"score"`
	Relevance   float64   `json:"relevance"`
	Explanation string    `json:"explanation"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// RetrievalOptions configures retrieval behavior
type RetrievalOptions struct {
	Limit           int                    `json:"limit"`
	Threshold       float64                `json:"threshold"`
	Filter          map[string]interface{} `json:"filter"`
	RerankResults   bool                   `json:"rerank_results"`
	IncludeMetadata bool                   `json:"include_metadata"`
	MaxTokens       int                    `json:"max_tokens"`
}

// VectorStoreRetriever implements Retriever using a vector store
type VectorStoreRetriever struct {
	vectorStore VectorStore
	tracer      trace.Tracer
}

// NewVectorStoreRetriever creates a new vector store retriever
func NewVectorStoreRetriever(vectorStore VectorStore) *VectorStoreRetriever {
	return &VectorStoreRetriever{
		vectorStore: vectorStore,
		tracer:      otel.Tracer("rag.retriever"),
	}
}

// Retrieve retrieves relevant documents for a query
func (vsr *VectorStoreRetriever) Retrieve(ctx context.Context, query string, limit int) ([]*RetrievalResult, error) {
	options := &RetrievalOptions{
		Limit:           limit,
		Threshold:       0.0,
		RerankResults:   true,
		IncludeMetadata: true,
	}
	
	return vsr.GetRelevantDocuments(ctx, query, options)
}

// RetrieveWithFilter retrieves documents with metadata filtering
func (vsr *VectorStoreRetriever) RetrieveWithFilter(ctx context.Context, query string, filter map[string]interface{}, limit int) ([]*RetrievalResult, error) {
	options := &RetrievalOptions{
		Limit:           limit,
		Threshold:       0.0,
		Filter:          filter,
		RerankResults:   true,
		IncludeMetadata: true,
	}
	
	return vsr.GetRelevantDocuments(ctx, query, options)
}

// GetRelevantDocuments gets documents relevant to a query with scoring
func (vsr *VectorStoreRetriever) GetRelevantDocuments(ctx context.Context, query string, options *RetrievalOptions) ([]*RetrievalResult, error) {
	ctx, span := vsr.tracer.Start(ctx, "vector_retriever.get_relevant_documents")
	defer span.End()

	span.SetAttributes(
		attribute.String("query", query),
		attribute.Int("limit", options.Limit),
		attribute.Float64("threshold", options.Threshold),
	)

	// Search vector store
	searchResults, err := vsr.vectorStore.SearchByText(ctx, query, options.Limit*2, options.Threshold) // Get more results for reranking
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// Convert to retrieval results
	var results []*RetrievalResult
	for _, searchResult := range searchResults {
		// Apply metadata filter if specified
		if options.Filter != nil && !vsr.matchesFilter(searchResult.Document, options.Filter) {
			continue
		}

		result := &RetrievalResult{
			Document:  searchResult.Document,
			Score:     searchResult.Score,
			Relevance: searchResult.Similarity,
			Metadata:  make(map[string]interface{}),
		}

		// Add retrieval metadata
		result.Metadata["retrieval_method"] = "vector_similarity"
		result.Metadata["similarity_score"] = searchResult.Similarity
		result.Metadata["query"] = query

		if options.IncludeMetadata {
			for k, v := range searchResult.Document.Metadata {
				result.Metadata[k] = v
			}
		}

		results = append(results, result)
	}

	// Rerank results if requested
	if options.RerankResults {
		results = vsr.rerankResults(ctx, query, results)
	}

	// Apply final limit
	if len(results) > options.Limit {
		results = results[:options.Limit]
	}

	// Generate explanations
	for i, result := range results {
		result.Explanation = vsr.generateExplanation(query, result, i+1)
	}

	span.SetAttributes(
		attribute.Int("results.count", len(results)),
		attribute.Float64("top_score", func() float64 {
			if len(results) > 0 {
				return results[0].Score
			}
			return 0.0
		}()),
	)

	return results, nil
}

// matchesFilter checks if a document matches the given filter
func (vsr *VectorStoreRetriever) matchesFilter(doc *Document, filter map[string]interface{}) bool {
	for key, expectedValue := range filter {
		actualValue, exists := doc.Metadata[key]
		if !exists || actualValue != expectedValue {
			return false
		}
	}
	return true
}

// rerankResults reranks results using additional scoring factors
func (vsr *VectorStoreRetriever) rerankResults(ctx context.Context, query string, results []*RetrievalResult) []*RetrievalResult {
	ctx, span := vsr.tracer.Start(ctx, "vector_retriever.rerank_results")
	defer span.End()

	queryLower := strings.ToLower(query)
	queryWords := strings.Fields(queryLower)

	for _, result := range results {
		// Calculate additional relevance factors
		contentLower := strings.ToLower(result.Document.Content)
		
		// Keyword matching score
		keywordScore := vsr.calculateKeywordScore(queryWords, contentLower)
		
		// Recency score (newer documents get slight boost)
		recencyScore := vsr.calculateRecencyScore(result.Document.Created)
		
		// Content quality score (based on length and structure)
		qualityScore := vsr.calculateQualityScore(result.Document.Content)
		
		// Combine scores with weights
		combinedScore := (result.Relevance * 0.7) + (keywordScore * 0.2) + (recencyScore * 0.05) + (qualityScore * 0.05)
		
		result.Score = combinedScore
		result.Metadata["keyword_score"] = keywordScore
		result.Metadata["recency_score"] = recencyScore
		result.Metadata["quality_score"] = qualityScore
	}

	// Sort by combined score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	span.SetAttributes(attribute.Int("reranked.count", len(results)))
	return results
}

// calculateKeywordScore calculates keyword matching score
func (vsr *VectorStoreRetriever) calculateKeywordScore(queryWords []string, content string) float64 {
	if len(queryWords) == 0 {
		return 0.0
	}

	matches := 0
	for _, word := range queryWords {
		if strings.Contains(content, word) {
			matches++
		}
	}

	return float64(matches) / float64(len(queryWords))
}

// calculateRecencyScore calculates recency score
func (vsr *VectorStoreRetriever) calculateRecencyScore(created time.Time) float64 {
	if created.IsZero() {
		return 0.0
	}

	daysSinceCreation := time.Since(created).Hours() / 24
	
	// Newer documents get higher scores, with diminishing returns
	if daysSinceCreation < 30 {
		return 1.0
	} else if daysSinceCreation < 365 {
		return 0.8
	} else {
		return 0.5
	}
}

// calculateQualityScore calculates content quality score
func (vsr *VectorStoreRetriever) calculateQualityScore(content string) float64 {
	length := len(content)
	
	// Optimal length range gets highest score
	if length >= 200 && length <= 2000 {
		return 1.0
	} else if length >= 100 && length <= 5000 {
		return 0.8
	} else if length >= 50 {
		return 0.6
	} else {
		return 0.3
	}
}

// generateExplanation generates an explanation for why a document was retrieved
func (vsr *VectorStoreRetriever) generateExplanation(query string, result *RetrievalResult, rank int) string {
	explanation := fmt.Sprintf("Ranked #%d with %.2f%% relevance", rank, result.Relevance*100)
	
	if keywordScore, exists := result.Metadata["keyword_score"]; exists {
		if score, ok := keywordScore.(float64); ok && score > 0.5 {
			explanation += fmt.Sprintf(", high keyword match (%.1f%%)", score*100)
		}
	}
	
	if contentType, exists := result.Document.Metadata["content_type"]; exists {
		explanation += fmt.Sprintf(", %s content", contentType)
	}
	
	return explanation
}

// HybridRetriever combines multiple retrieval methods
type HybridRetriever struct {
	retrievers []Retriever
	weights    []float64
	tracer     trace.Tracer
}

// NewHybridRetriever creates a new hybrid retriever
func NewHybridRetriever(retrievers []Retriever, weights []float64) *HybridRetriever {
	if len(weights) != len(retrievers) {
		// Default to equal weights
		weights = make([]float64, len(retrievers))
		for i := range weights {
			weights[i] = 1.0 / float64(len(retrievers))
		}
	}

	return &HybridRetriever{
		retrievers: retrievers,
		weights:    weights,
		tracer:     otel.Tracer("rag.hybrid_retriever"),
	}
}

// Retrieve retrieves documents using multiple methods and combines results
func (hr *HybridRetriever) Retrieve(ctx context.Context, query string, limit int) ([]*RetrievalResult, error) {
	ctx, span := hr.tracer.Start(ctx, "hybrid_retriever.retrieve")
	defer span.End()

	span.SetAttributes(
		attribute.String("query", query),
		attribute.Int("retrievers.count", len(hr.retrievers)),
		attribute.Int("limit", limit),
	)

	// Collect results from all retrievers
	var allResults []*RetrievalResult
	resultMap := make(map[string]*RetrievalResult)

	for i, retriever := range hr.retrievers {
		results, err := retriever.Retrieve(ctx, query, limit*2) // Get more results for fusion
		if err != nil {
			span.RecordError(err)
			continue // Skip failed retrievers
		}

		weight := hr.weights[i]
		for _, result := range results {
			docID := result.Document.ID
			
			if existing, exists := resultMap[docID]; exists {
				// Combine scores using weighted average
				existing.Score = (existing.Score + result.Score*weight) / 2
				existing.Relevance = (existing.Relevance + result.Relevance*weight) / 2
			} else {
				// New result
				newResult := &RetrievalResult{
					Document:    result.Document,
					Score:       result.Score * weight,
					Relevance:   result.Relevance * weight,
					Explanation: result.Explanation,
					Metadata:    make(map[string]interface{}),
				}

				// Copy metadata
				for k, v := range result.Metadata {
					newResult.Metadata[k] = v
				}
				newResult.Metadata["retrieval_method"] = "hybrid"
				newResult.Metadata["weight"] = weight

				resultMap[docID] = newResult
				allResults = append(allResults, newResult)
			}
		}
	}

	// Sort by combined score
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Score > allResults[j].Score
	})

	// Apply limit
	if len(allResults) > limit {
		allResults = allResults[:limit]
	}

	span.SetAttributes(attribute.Int("results.final_count", len(allResults)))
	return allResults, nil
}

// RetrieveWithFilter retrieves documents with filtering
func (hr *HybridRetriever) RetrieveWithFilter(ctx context.Context, query string, filter map[string]interface{}, limit int) ([]*RetrievalResult, error) {
	// For simplicity, delegate to first retriever that supports filtering
	for _, retriever := range hr.retrievers {
		if vsr, ok := retriever.(*VectorStoreRetriever); ok {
			return vsr.RetrieveWithFilter(ctx, query, filter, limit)
		}
	}
	
	// Fallback to regular retrieve
	return hr.Retrieve(ctx, query, limit)
}

// GetRelevantDocuments gets relevant documents with options
func (hr *HybridRetriever) GetRelevantDocuments(ctx context.Context, query string, options *RetrievalOptions) ([]*RetrievalResult, error) {
	// For simplicity, delegate to first retriever
	if len(hr.retrievers) > 0 {
		if vsr, ok := hr.retrievers[0].(*VectorStoreRetriever); ok {
			return vsr.GetRelevantDocuments(ctx, query, options)
		}
	}
	
	return hr.Retrieve(ctx, query, options.Limit)
}

// TravelRetriever specializes in travel-related document retrieval
type TravelRetriever struct {
	*VectorStoreRetriever
	tracer trace.Tracer
}

// NewTravelRetriever creates a new travel-specific retriever
func NewTravelRetriever(vectorStore VectorStore) *TravelRetriever {
	return &TravelRetriever{
		VectorStoreRetriever: NewVectorStoreRetriever(vectorStore),
		tracer:               otel.Tracer("rag.travel_retriever"),
	}
}

// RetrieveForDestination retrieves documents relevant to a specific destination
func (tr *TravelRetriever) RetrieveForDestination(ctx context.Context, destination, query string, limit int) ([]*RetrievalResult, error) {
	ctx, span := tr.tracer.Start(ctx, "travel_retriever.retrieve_for_destination")
	defer span.End()

	span.SetAttributes(
		attribute.String("destination", destination),
		attribute.String("query", query),
	)

	// Enhance query with destination context
	enhancedQuery := fmt.Sprintf("%s %s travel guide information", destination, query)

	// Use destination filter
	filter := map[string]interface{}{
		"document_type": "travel_guide",
	}

	options := &RetrievalOptions{
		Limit:           limit,
		Threshold:       0.1,
		Filter:          filter,
		RerankResults:   true,
		IncludeMetadata: true,
	}

	results, err := tr.GetRelevantDocuments(ctx, enhancedQuery, options)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Post-process results for travel-specific ranking
	tr.applyTravelRanking(destination, results)

	span.SetAttributes(attribute.Int("travel_results.count", len(results)))
	return results, nil
}

// RetrieveByCategory retrieves documents by travel category
func (tr *TravelRetriever) RetrieveByCategory(ctx context.Context, category, query string, limit int) ([]*RetrievalResult, error) {
	filter := map[string]interface{}{
		"content_type": category,
	}

	return tr.RetrieveWithFilter(ctx, query, filter, limit)
}

// applyTravelRanking applies travel-specific ranking to results
func (tr *TravelRetriever) applyTravelRanking(destination string, results []*RetrievalResult) {
	destinationLower := strings.ToLower(destination)

	for _, result := range results {
		// Boost score if destination is mentioned in content
		contentLower := strings.ToLower(result.Document.Content)
		if strings.Contains(contentLower, destinationLower) {
			result.Score *= 1.2
		}

		// Boost score if destination is in metadata
		if destinations, exists := result.Document.Metadata["destinations"]; exists {
			if destList, ok := destinations.([]string); ok {
				for _, dest := range destList {
					if strings.Contains(strings.ToLower(dest), destinationLower) {
						result.Score *= 1.1
						break
					}
				}
			}
		}
	}

	// Re-sort by updated scores
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
}

package rag

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Document represents a document in the vector store
type Document struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
	Vector   []float64              `json:"vector"`
	Created  time.Time              `json:"created"`
	Updated  time.Time              `json:"updated"`
}

// SearchResult represents a search result with similarity score
type SearchResult struct {
	Document   *Document `json:"document"`
	Score      float64   `json:"score"`
	Similarity float64   `json:"similarity"`
}

// VectorStore defines the interface for vector storage and retrieval
type VectorStore interface {
	// AddDocument adds a document to the vector store
	AddDocument(ctx context.Context, doc *Document) error

	// AddDocuments adds multiple documents to the vector store
	AddDocuments(ctx context.Context, docs []*Document) error

	// GetDocument retrieves a document by ID
	GetDocument(ctx context.Context, id string) (*Document, error)

	// UpdateDocument updates an existing document
	UpdateDocument(ctx context.Context, doc *Document) error

	// DeleteDocument removes a document from the store
	DeleteDocument(ctx context.Context, id string) error

	// Search performs similarity search
	Search(ctx context.Context, queryVector []float64, limit int, threshold float64) ([]*SearchResult, error)

	// SearchByText performs text-based search (requires embedding service)
	SearchByText(ctx context.Context, query string, limit int, threshold float64) ([]*SearchResult, error)

	// ListDocuments returns all documents with optional filtering
	ListDocuments(ctx context.Context, filter map[string]interface{}) ([]*Document, error)

	// GetStats returns statistics about the vector store
	GetStats(ctx context.Context) (*VectorStoreStats, error)

	// Clear removes all documents from the store
	Clear(ctx context.Context) error
}

// VectorStoreStats represents statistics about the vector store
type VectorStoreStats struct {
	DocumentCount int       `json:"document_count"`
	VectorDim     int       `json:"vector_dimension"`
	LastUpdated   time.Time `json:"last_updated"`
	StorageSize   int64     `json:"storage_size_bytes"`
}

// MemoryVectorStore implements VectorStore using in-memory storage
type MemoryVectorStore struct {
	documents        map[string]*Document
	embeddingService EmbeddingService
	mutex            sync.RWMutex
	tracer           trace.Tracer
}

// NewMemoryVectorStore creates a new in-memory vector store
func NewMemoryVectorStore(embeddingService EmbeddingService) *MemoryVectorStore {
	return &MemoryVectorStore{
		documents:        make(map[string]*Document),
		embeddingService: embeddingService,
		tracer:           otel.Tracer("rag.vector_store"),
	}
}

// AddDocument adds a document to the vector store
func (vs *MemoryVectorStore) AddDocument(ctx context.Context, doc *Document) error {
	ctx, span := vs.tracer.Start(ctx, "memory_vector_store.add_document")
	defer span.End()

	span.SetAttributes(
		attribute.String("document.id", doc.ID),
		attribute.Int("document.content_length", len(doc.Content)),
	)

	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	// Set timestamps
	now := time.Now()
	if doc.Created.IsZero() {
		doc.Created = now
	}
	doc.Updated = now

	// Generate embedding if not provided
	if len(doc.Vector) == 0 && vs.embeddingService != nil {
		vector, err := vs.embeddingService.GenerateEmbedding(ctx, doc.Content)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to generate embedding: %w", err)
		}
		doc.Vector = vector
	}

	vs.documents[doc.ID] = doc

	span.SetAttributes(
		attribute.Int("vector.dimension", len(doc.Vector)),
		attribute.Int("store.total_documents", len(vs.documents)),
	)

	return nil
}

// AddDocuments adds multiple documents to the vector store
func (vs *MemoryVectorStore) AddDocuments(ctx context.Context, docs []*Document) error {
	ctx, span := vs.tracer.Start(ctx, "memory_vector_store.add_documents")
	defer span.End()

	span.SetAttributes(attribute.Int("documents.count", len(docs)))

	for i, doc := range docs {
		if err := vs.AddDocument(ctx, doc); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to add document %d: %w", i, err)
		}
	}

	return nil
}

// GetDocument retrieves a document by ID
func (vs *MemoryVectorStore) GetDocument(ctx context.Context, id string) (*Document, error) {
	ctx, span := vs.tracer.Start(ctx, "memory_vector_store.get_document")
	defer span.End()

	span.SetAttributes(attribute.String("document.id", id))

	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	doc, exists := vs.documents[id]
	if !exists {
		err := fmt.Errorf("document not found: %s", id)
		span.RecordError(err)
		return nil, err
	}

	// Return a copy to prevent external modification
	return vs.copyDocument(doc), nil
}

// UpdateDocument updates an existing document
func (vs *MemoryVectorStore) UpdateDocument(ctx context.Context, doc *Document) error {
	ctx, span := vs.tracer.Start(ctx, "memory_vector_store.update_document")
	defer span.End()

	span.SetAttributes(attribute.String("document.id", doc.ID))

	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	existing, exists := vs.documents[doc.ID]
	if !exists {
		err := fmt.Errorf("document not found: %s", doc.ID)
		span.RecordError(err)
		return err
	}

	// Preserve creation time
	doc.Created = existing.Created
	doc.Updated = time.Now()

	// Regenerate embedding if content changed
	if doc.Content != existing.Content && vs.embeddingService != nil {
		vector, err := vs.embeddingService.GenerateEmbedding(ctx, doc.Content)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to generate embedding: %w", err)
		}
		doc.Vector = vector
	}

	vs.documents[doc.ID] = doc
	return nil
}

// DeleteDocument removes a document from the store
func (vs *MemoryVectorStore) DeleteDocument(ctx context.Context, id string) error {
	ctx, span := vs.tracer.Start(ctx, "memory_vector_store.delete_document")
	defer span.End()

	span.SetAttributes(attribute.String("document.id", id))

	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	if _, exists := vs.documents[id]; !exists {
		err := fmt.Errorf("document not found: %s", id)
		span.RecordError(err)
		return err
	}

	delete(vs.documents, id)
	return nil
}

// Search performs similarity search using cosine similarity
func (vs *MemoryVectorStore) Search(ctx context.Context, queryVector []float64, limit int, threshold float64) ([]*SearchResult, error) {
	ctx, span := vs.tracer.Start(ctx, "memory_vector_store.search")
	defer span.End()

	span.SetAttributes(
		attribute.Int("query.vector_dimension", len(queryVector)),
		attribute.Int("search.limit", limit),
		attribute.Float64("search.threshold", threshold),
	)

	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	var results []*SearchResult

	for _, doc := range vs.documents {
		if len(doc.Vector) == 0 {
			continue // Skip documents without vectors
		}

		similarity := vs.cosineSimilarity(queryVector, doc.Vector)
		if similarity >= threshold {
			results = append(results, &SearchResult{
				Document:   vs.copyDocument(doc),
				Score:      similarity,
				Similarity: similarity,
			})
		}
	}

	// Sort by similarity (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	// Apply limit
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	span.SetAttributes(
		attribute.Int("search.results_count", len(results)),
		attribute.Int("store.total_documents", len(vs.documents)),
	)

	return results, nil
}

// SearchByText performs text-based search by generating embeddings
func (vs *MemoryVectorStore) SearchByText(ctx context.Context, query string, limit int, threshold float64) ([]*SearchResult, error) {
	ctx, span := vs.tracer.Start(ctx, "memory_vector_store.search_by_text")
	defer span.End()

	span.SetAttributes(
		attribute.String("query.text", query),
		attribute.Int("search.limit", limit),
		attribute.Float64("search.threshold", threshold),
	)

	if vs.embeddingService == nil {
		err := fmt.Errorf("embedding service not available")
		span.RecordError(err)
		return nil, err
	}

	// Generate embedding for query
	queryVector, err := vs.embeddingService.GenerateEmbedding(ctx, query)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	return vs.Search(ctx, queryVector, limit, threshold)
}

// ListDocuments returns all documents with optional filtering
func (vs *MemoryVectorStore) ListDocuments(ctx context.Context, filter map[string]interface{}) ([]*Document, error) {
	ctx, span := vs.tracer.Start(ctx, "memory_vector_store.list_documents")
	defer span.End()

	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	var documents []*Document

	for _, doc := range vs.documents {
		if vs.matchesFilter(doc, filter) {
			documents = append(documents, vs.copyDocument(doc))
		}
	}

	span.SetAttributes(
		attribute.Int("documents.total", len(vs.documents)),
		attribute.Int("documents.filtered", len(documents)),
	)

	return documents, nil
}

// GetStats returns statistics about the vector store
func (vs *MemoryVectorStore) GetStats(ctx context.Context) (*VectorStoreStats, error) {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	var vectorDim int
	var lastUpdated time.Time
	var storageSize int64

	for _, doc := range vs.documents {
		if len(doc.Vector) > 0 && vectorDim == 0 {
			vectorDim = len(doc.Vector)
		}
		if doc.Updated.After(lastUpdated) {
			lastUpdated = doc.Updated
		}
		storageSize += int64(len(doc.Content))
	}

	return &VectorStoreStats{
		DocumentCount: len(vs.documents),
		VectorDim:     vectorDim,
		LastUpdated:   lastUpdated,
		StorageSize:   storageSize,
	}, nil
}

// Clear removes all documents from the store
func (vs *MemoryVectorStore) Clear(ctx context.Context) error {
	ctx, span := vs.tracer.Start(ctx, "memory_vector_store.clear")
	defer span.End()

	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	count := len(vs.documents)
	vs.documents = make(map[string]*Document)

	span.SetAttributes(attribute.Int("documents.cleared", count))
	return nil
}

// cosineSimilarity calculates cosine similarity between two vectors
func (vs *MemoryVectorStore) cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float64

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0.0 || normB == 0.0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// copyDocument creates a deep copy of a document
func (vs *MemoryVectorStore) copyDocument(doc *Document) *Document {
	docCopy := &Document{
		ID:       doc.ID,
		Content:  doc.Content,
		Vector:   make([]float64, len(doc.Vector)),
		Created:  doc.Created,
		Updated:  doc.Updated,
		Metadata: make(map[string]interface{}),
	}

	copy(docCopy.Vector, doc.Vector)

	for k, v := range doc.Metadata {
		docCopy.Metadata[k] = v
	}

	return docCopy
}

// matchesFilter checks if a document matches the given filter
func (vs *MemoryVectorStore) matchesFilter(doc *Document, filter map[string]interface{}) bool {
	if filter == nil {
		return true
	}

	for key, expectedValue := range filter {
		if actualValue, exists := doc.Metadata[key]; !exists || actualValue != expectedValue {
			return false
		}
	}

	return true
}

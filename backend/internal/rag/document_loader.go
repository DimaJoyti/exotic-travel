package rag

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DocumentLoader defines the interface for loading documents
type DocumentLoader interface {
	// LoadDocument loads a single document from a source
	LoadDocument(ctx context.Context, source string) (*Document, error)

	// LoadDocuments loads multiple documents from a source
	LoadDocuments(ctx context.Context, source string) ([]*Document, error)

	// GetSupportedFormats returns the supported file formats
	GetSupportedFormats() []string
}

// TextSplitter defines the interface for splitting text into chunks
type TextSplitter interface {
	// SplitText splits text into chunks
	SplitText(text string) []string

	// SplitDocuments splits documents into smaller chunks
	SplitDocuments(docs []*Document) []*Document
}

// FileDocumentLoader loads documents from files
type FileDocumentLoader struct {
	textSplitter TextSplitter
	tracer       trace.Tracer
}

// NewFileDocumentLoader creates a new file document loader
func NewFileDocumentLoader(textSplitter TextSplitter) *FileDocumentLoader {
	return &FileDocumentLoader{
		textSplitter: textSplitter,
		tracer:       otel.Tracer("rag.document_loader"),
	}
}

// LoadDocument loads a single document from a file
func (fdl *FileDocumentLoader) LoadDocument(ctx context.Context, filePath string) (*Document, error) {
	ctx, span := fdl.tracer.Start(ctx, "file_loader.load_document")
	defer span.End()

	span.SetAttributes(attribute.String("file.path", filePath))

	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Create document
	doc := &Document{
		ID:      generateDocumentID(filePath),
		Content: string(content),
		Metadata: map[string]interface{}{
			"source":      filePath,
			"filename":    filepath.Base(filePath),
			"extension":   filepath.Ext(filePath),
			"size":        info.Size(),
			"modified":    info.ModTime(),
			"loader_type": "file",
		},
		Created: time.Now(),
	}

	span.SetAttributes(
		attribute.String("document.id", doc.ID),
		attribute.Int("document.size", len(doc.Content)),
	)

	return doc, nil
}

// LoadDocuments loads multiple documents from a directory
func (fdl *FileDocumentLoader) LoadDocuments(ctx context.Context, dirPath string) ([]*Document, error) {
	ctx, span := fdl.tracer.Start(ctx, "file_loader.load_documents")
	defer span.End()

	span.SetAttributes(attribute.String("directory.path", dirPath))

	var documents []*Document

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file format is supported
		if !fdl.isSupportedFormat(path) {
			return nil
		}

		doc, err := fdl.LoadDocument(ctx, path)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to load document %s: %w", path, err)
		}

		// Split document if text splitter is provided
		if fdl.textSplitter != nil {
			chunks := fdl.textSplitter.SplitDocuments([]*Document{doc})
			documents = append(documents, chunks...)
		} else {
			documents = append(documents, doc)
		}

		return nil
	})

	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	span.SetAttributes(attribute.Int("documents.loaded", len(documents)))
	return documents, nil
}

// GetSupportedFormats returns the supported file formats
func (fdl *FileDocumentLoader) GetSupportedFormats() []string {
	return []string{".txt", ".md", ".json", ".csv"}
}

// isSupportedFormat checks if a file format is supported
func (fdl *FileDocumentLoader) isSupportedFormat(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	for _, supported := range fdl.GetSupportedFormats() {
		if ext == supported {
			return true
		}
	}
	return false
}

// RecursiveCharacterTextSplitter splits text by characters with overlap
type RecursiveCharacterTextSplitter struct {
	chunkSize    int
	chunkOverlap int
	separators   []string
	tracer       trace.Tracer
}

// NewRecursiveCharacterTextSplitter creates a new recursive character text splitter
func NewRecursiveCharacterTextSplitter(chunkSize, chunkOverlap int) *RecursiveCharacterTextSplitter {
	return &RecursiveCharacterTextSplitter{
		chunkSize:    chunkSize,
		chunkOverlap: chunkOverlap,
		separators:   []string{"\n\n", "\n", " ", ""},
		tracer:       otel.Tracer("rag.text_splitter"),
	}
}

// SplitText splits text into chunks
func (rcts *RecursiveCharacterTextSplitter) SplitText(text string) []string {
	_, span := rcts.tracer.Start(context.Background(), "text_splitter.split_text")
	defer span.End()

	span.SetAttributes(
		attribute.Int("text.length", len(text)),
		attribute.Int("chunk.size", rcts.chunkSize),
		attribute.Int("chunk.overlap", rcts.chunkOverlap),
	)

	if len(text) <= rcts.chunkSize {
		return []string{text}
	}

	chunks := rcts.splitTextRecursive(text, rcts.separators)

	span.SetAttributes(attribute.Int("chunks.count", len(chunks)))
	return chunks
}

// SplitDocuments splits documents into smaller chunks
func (rcts *RecursiveCharacterTextSplitter) SplitDocuments(docs []*Document) []*Document {
	_, span := rcts.tracer.Start(context.Background(), "text_splitter.split_documents")
	defer span.End()

	span.SetAttributes(attribute.Int("documents.input", len(docs)))

	var result []*Document

	for _, doc := range docs {
		chunks := rcts.SplitText(doc.Content)

		for i, chunk := range chunks {
			chunkDoc := &Document{
				ID:       fmt.Sprintf("%s_chunk_%d", doc.ID, i),
				Content:  chunk,
				Metadata: make(map[string]interface{}),
				Created:  time.Now(),
			}

			// Copy metadata from original document
			for k, v := range doc.Metadata {
				chunkDoc.Metadata[k] = v
			}

			// Add chunk-specific metadata
			chunkDoc.Metadata["chunk_index"] = i
			chunkDoc.Metadata["total_chunks"] = len(chunks)
			chunkDoc.Metadata["parent_document_id"] = doc.ID

			result = append(result, chunkDoc)
		}
	}

	span.SetAttributes(attribute.Int("documents.output", len(result)))
	return result
}

// splitTextRecursive recursively splits text using different separators
func (rcts *RecursiveCharacterTextSplitter) splitTextRecursive(text string, separators []string) []string {
	if len(text) <= rcts.chunkSize {
		return []string{text}
	}

	if len(separators) == 0 {
		// No more separators, split by character count
		return rcts.splitByLength(text)
	}

	separator := separators[0]
	remainingSeparators := separators[1:]

	if separator == "" {
		// Split by character count
		return rcts.splitByLength(text)
	}

	splits := strings.Split(text, separator)
	var chunks []string
	var currentChunk strings.Builder

	for _, split := range splits {
		testChunk := currentChunk.String()
		if testChunk != "" {
			testChunk += separator
		}
		testChunk += split

		if len(testChunk) <= rcts.chunkSize {
			if currentChunk.Len() > 0 {
				currentChunk.WriteString(separator)
			}
			currentChunk.WriteString(split)
		} else {
			// Current chunk is full, process it
			if currentChunk.Len() > 0 {
				chunk := currentChunk.String()
				if len(chunk) > rcts.chunkSize {
					// Chunk is still too large, split recursively
					subChunks := rcts.splitTextRecursive(chunk, remainingSeparators)
					chunks = append(chunks, subChunks...)
				} else {
					chunks = append(chunks, chunk)
				}
				currentChunk.Reset()
			}

			// Start new chunk with current split
			if len(split) > rcts.chunkSize {
				// Split is too large, split recursively
				subChunks := rcts.splitTextRecursive(split, remainingSeparators)
				chunks = append(chunks, subChunks...)
			} else {
				currentChunk.WriteString(split)
			}
		}
	}

	// Add remaining chunk
	if currentChunk.Len() > 0 {
		chunk := currentChunk.String()
		if len(chunk) > rcts.chunkSize {
			subChunks := rcts.splitTextRecursive(chunk, remainingSeparators)
			chunks = append(chunks, subChunks...)
		} else {
			chunks = append(chunks, chunk)
		}
	}

	return rcts.addOverlap(chunks)
}

// splitByLength splits text by character count
func (rcts *RecursiveCharacterTextSplitter) splitByLength(text string) []string {
	var chunks []string

	for i := 0; i < len(text); i += rcts.chunkSize {
		end := i + rcts.chunkSize
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[i:end])
	}

	return rcts.addOverlap(chunks)
}

// addOverlap adds overlap between chunks
func (rcts *RecursiveCharacterTextSplitter) addOverlap(chunks []string) []string {
	if rcts.chunkOverlap == 0 || len(chunks) <= 1 {
		return chunks
	}

	var result []string

	for i, chunk := range chunks {
		if i == 0 {
			result = append(result, chunk)
			continue
		}

		// Add overlap from previous chunk
		prevChunk := chunks[i-1]
		overlapStart := len(prevChunk) - rcts.chunkOverlap
		if overlapStart < 0 {
			overlapStart = 0
		}

		overlap := prevChunk[overlapStart:]
		overlappedChunk := overlap + chunk
		result = append(result, overlappedChunk)
	}

	return result
}

// TravelDocumentLoader specializes in loading travel-related documents
type TravelDocumentLoader struct {
	*FileDocumentLoader
	tracer trace.Tracer
}

// NewTravelDocumentLoader creates a new travel document loader
func NewTravelDocumentLoader(textSplitter TextSplitter) *TravelDocumentLoader {
	return &TravelDocumentLoader{
		FileDocumentLoader: NewFileDocumentLoader(textSplitter),
		tracer:             otel.Tracer("rag.travel_loader"),
	}
}

// LoadTravelGuides loads travel guide documents with enhanced metadata
func (tdl *TravelDocumentLoader) LoadTravelGuides(ctx context.Context, guidesPath string) ([]*Document, error) {
	ctx, span := tdl.tracer.Start(ctx, "travel_loader.load_guides")
	defer span.End()

	documents, err := tdl.LoadDocuments(ctx, guidesPath)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Enhance documents with travel-specific metadata
	for _, doc := range documents {
		tdl.enhanceTravelMetadata(doc)
	}

	span.SetAttributes(attribute.Int("travel_guides.loaded", len(documents)))
	return documents, nil
}

// enhanceTravelMetadata adds travel-specific metadata to documents
func (tdl *TravelDocumentLoader) enhanceTravelMetadata(doc *Document) {
	content := strings.ToLower(doc.Content)

	// Extract destinations mentioned in the document
	destinations := tdl.extractDestinations(content)
	if len(destinations) > 0 {
		doc.Metadata["destinations"] = destinations
	}

	// Categorize content type
	doc.Metadata["content_type"] = tdl.categorizeContent(content)

	// Extract travel themes
	themes := tdl.extractTravelThemes(content)
	if len(themes) > 0 {
		doc.Metadata["themes"] = themes
	}

	// Add document type
	doc.Metadata["document_type"] = "travel_guide"
}

// extractDestinations extracts destination names from content
func (tdl *TravelDocumentLoader) extractDestinations(content string) []string {
	// Simple pattern matching for common destination patterns
	patterns := []string{
		`\b([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*),\s*([A-Z][a-z]+)\b`, // City, Country
		`\b([A-Z][a-z]+\s+[A-Z][a-z]+)\b`,                       // Two-word places
	}

	var destinations []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(content, -1)

		for _, match := range matches {
			if len(match) > 1 {
				dest := strings.TrimSpace(match[0])
				if !seen[dest] && len(dest) > 3 {
					destinations = append(destinations, dest)
					seen[dest] = true
				}
			}
		}
	}

	return destinations
}

// categorizeContent categorizes the type of travel content
func (tdl *TravelDocumentLoader) categorizeContent(content string) string {
	keywords := map[string][]string{
		"accommodation":  {"hotel", "resort", "hostel", "accommodation", "stay", "lodge"},
		"transportation": {"flight", "train", "bus", "car", "transport", "airline", "airport"},
		"attractions":    {"museum", "park", "monument", "attraction", "sightseeing", "tour"},
		"food":           {"restaurant", "food", "cuisine", "dining", "cafe", "bar", "local dishes"},
		"culture":        {"culture", "history", "tradition", "festival", "art", "music"},
		"adventure":      {"hiking", "climbing", "diving", "adventure", "outdoor", "sports"},
		"budget":         {"budget", "cheap", "affordable", "cost", "price", "money"},
		"luxury":         {"luxury", "premium", "exclusive", "high-end", "upscale"},
	}

	scores := make(map[string]int)

	for category, words := range keywords {
		for _, word := range words {
			if strings.Contains(content, word) {
				scores[category]++
			}
		}
	}

	// Find category with highest score
	maxScore := 0
	category := "general"

	for cat, score := range scores {
		if score > maxScore {
			maxScore = score
			category = cat
		}
	}

	return category
}

// extractTravelThemes extracts travel themes from content
func (tdl *TravelDocumentLoader) extractTravelThemes(content string) []string {
	themes := []string{
		"beach", "mountain", "city", "nature", "wildlife", "photography",
		"family", "romantic", "solo", "group", "business", "leisure",
		"summer", "winter", "spring", "autumn", "tropical", "desert",
	}

	var found []string
	for _, theme := range themes {
		if strings.Contains(content, theme) {
			found = append(found, theme)
		}
	}

	return found
}

// generateDocumentID generates a unique document ID from file path
func generateDocumentID(filePath string) string {
	// Use file path and modification time for uniqueness
	return fmt.Sprintf("doc_%x", []byte(filePath))
}

// StringDocumentLoader loads documents from strings
type StringDocumentLoader struct {
	textSplitter TextSplitter
	tracer       trace.Tracer
}

// NewStringDocumentLoader creates a new string document loader
func NewStringDocumentLoader(textSplitter TextSplitter) *StringDocumentLoader {
	return &StringDocumentLoader{
		textSplitter: textSplitter,
		tracer:       otel.Tracer("rag.string_loader"),
	}
}

// LoadFromString loads a document from a string
func (sdl *StringDocumentLoader) LoadFromString(content string, metadata map[string]interface{}) *Document {
	doc := &Document{
		ID:       fmt.Sprintf("string_doc_%d", time.Now().UnixNano()),
		Content:  content,
		Metadata: metadata,
		Created:  time.Now(),
	}

	if doc.Metadata == nil {
		doc.Metadata = make(map[string]interface{})
	}

	doc.Metadata["loader_type"] = "string"
	doc.Metadata["content_length"] = len(content)

	return doc
}

// LoadFromStrings loads multiple documents from strings
func (sdl *StringDocumentLoader) LoadFromStrings(contents []string, metadataList []map[string]interface{}) []*Document {
	var documents []*Document

	for i, content := range contents {
		var metadata map[string]interface{}
		if i < len(metadataList) {
			metadata = metadataList[i]
		}

		doc := sdl.LoadFromString(content, metadata)

		// Split document if text splitter is provided
		if sdl.textSplitter != nil {
			chunks := sdl.textSplitter.SplitDocuments([]*Document{doc})
			documents = append(documents, chunks...)
		} else {
			documents = append(documents, doc)
		}
	}

	return documents
}

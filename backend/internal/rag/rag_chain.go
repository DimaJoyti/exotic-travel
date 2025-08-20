package rag

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/exotic-travel-booking/backend/internal/langchain"
	"github.com/exotic-travel-booking/backend/internal/llm/providers"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// RAGChain implements Retrieval-Augmented Generation
type RAGChain struct {
	retriever      Retriever
	llmProvider    providers.LLMProvider
	promptTemplate *langchain.PromptTemplate
	outputParser   langchain.OutputParser
	tracer         trace.Tracer
	config         *RAGConfig
}

// RAGConfig configures RAG behavior
type RAGConfig struct {
	MaxContextLength   int     `json:"max_context_length"`
	RetrievalLimit     int     `json:"retrieval_limit"`
	RelevanceThreshold float64 `json:"relevance_threshold"`
	IncludeSourceInfo  bool    `json:"include_source_info"`
	ContextSeparator   string  `json:"context_separator"`
	MaxTokens          int     `json:"max_tokens"`
	Temperature        float64 `json:"temperature"`
}

// RAGResult represents the result of a RAG operation
type RAGResult struct {
	Query          string                 `json:"query"`
	Answer         string                 `json:"answer"`
	Sources        []*RetrievalResult     `json:"sources"`
	Context        string                 `json:"context"`
	Metadata       map[string]interface{} `json:"metadata"`
	Duration       time.Duration          `json:"duration"`
	TokensUsed     int                    `json:"tokens_used"`
	RetrievalTime  time.Duration          `json:"retrieval_time"`
	GenerationTime time.Duration          `json:"generation_time"`
}

// NewRAGChain creates a new RAG chain
func NewRAGChain(retriever Retriever, llmProvider providers.LLMProvider, config *RAGConfig) *RAGChain {
	if config == nil {
		config = &RAGConfig{
			MaxContextLength:   4000,
			RetrievalLimit:     5,
			RelevanceThreshold: 0.1,
			IncludeSourceInfo:  true,
			ContextSeparator:   "\n\n---\n\n",
			MaxTokens:          1000,
			Temperature:        0.7,
		}
	}

	// Create default prompt template for RAG
	promptTemplate := langchain.NewPromptTemplate(
		"rag_prompt",
		`You are a helpful travel assistant. Use the following context to answer the user's question about travel.

Context:
{{.context}}

Question: {{.question}}

Instructions:
- Provide a comprehensive and helpful answer based on the context provided
- If the context doesn't contain enough information, say so clearly
- Include specific details from the context when relevant
- Be accurate and don't make up information not present in the context
- Format your response in a clear and organized manner

Answer:`,
		[]string{"context", "question"},
	)

	return &RAGChain{
		retriever:      retriever,
		llmProvider:    llmProvider,
		promptTemplate: promptTemplate,
		tracer:         otel.Tracer("rag.chain"),
		config:         config,
	}
}

// SetPromptTemplate sets a custom prompt template
func (rag *RAGChain) SetPromptTemplate(template *langchain.PromptTemplate) *RAGChain {
	rag.promptTemplate = template
	return rag
}

// SetOutputParser sets an output parser
func (rag *RAGChain) SetOutputParser(parser langchain.OutputParser) *RAGChain {
	rag.outputParser = parser
	return rag
}

// Query performs a RAG query
func (rag *RAGChain) Query(ctx context.Context, question string) (*RAGResult, error) {
	ctx, span := rag.tracer.Start(ctx, "rag_chain.query")
	defer span.End()

	startTime := time.Now()

	span.SetAttributes(
		attribute.String("question", question),
		attribute.Int("retrieval_limit", rag.config.RetrievalLimit),
		attribute.Float64("relevance_threshold", rag.config.RelevanceThreshold),
	)

	result := &RAGResult{
		Query:    question,
		Metadata: make(map[string]interface{}),
	}

	// Step 1: Retrieve relevant documents
	retrievalStart := time.Now()
	sources, err := rag.retriever.Retrieve(ctx, question, rag.config.RetrievalLimit)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("retrieval failed: %w", err)
	}
	result.RetrievalTime = time.Since(retrievalStart)

	// Filter sources by relevance threshold
	var filteredSources []*RetrievalResult
	for _, source := range sources {
		if source.Relevance >= rag.config.RelevanceThreshold {
			filteredSources = append(filteredSources, source)
		}
	}
	result.Sources = filteredSources

	span.SetAttributes(
		attribute.Int("sources.retrieved", len(sources)),
		attribute.Int("sources.filtered", len(filteredSources)),
	)

	// Step 2: Build context from retrieved documents
	context := rag.buildContext(filteredSources)
	result.Context = context

	// Step 3: Generate answer using LLM
	generationStart := time.Now()
	answer, tokensUsed, err := rag.generateAnswer(ctx, question, context)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("generation failed: %w", err)
	}
	result.GenerationTime = time.Since(generationStart)

	result.Answer = answer
	result.TokensUsed = tokensUsed
	result.Duration = time.Since(startTime)

	// Add metadata
	result.Metadata["retrieval_method"] = "vector_similarity"
	result.Metadata["context_length"] = len(context)
	result.Metadata["sources_count"] = len(filteredSources)
	result.Metadata["model"] = rag.llmProvider.GetName()

	span.SetAttributes(
		attribute.String("answer.preview", truncateString(answer, 100)),
		attribute.Int("tokens.used", tokensUsed),
		attribute.Int64("duration.ms", result.Duration.Milliseconds()),
	)

	return result, nil
}

// QueryWithFilter performs a RAG query with metadata filtering
func (rag *RAGChain) QueryWithFilter(ctx context.Context, question string, filter map[string]interface{}) (*RAGResult, error) {
	ctx, span := rag.tracer.Start(ctx, "rag_chain.query_with_filter")
	defer span.End()

	startTime := time.Now()

	// Retrieve with filter
	retrievalStart := time.Now()
	sources, err := rag.retriever.RetrieveWithFilter(ctx, question, filter, rag.config.RetrievalLimit)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("filtered retrieval failed: %w", err)
	}
	retrievalTime := time.Since(retrievalStart)

	// Build context and generate answer
	context := rag.buildContext(sources)

	generationStart := time.Now()
	answer, tokensUsed, err := rag.generateAnswer(ctx, question, context)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("generation failed: %w", err)
	}
	generationTime := time.Since(generationStart)

	result := &RAGResult{
		Query:          question,
		Answer:         answer,
		Sources:        sources,
		Context:        context,
		Duration:       time.Since(startTime),
		TokensUsed:     tokensUsed,
		RetrievalTime:  retrievalTime,
		GenerationTime: generationTime,
		Metadata: map[string]interface{}{
			"filter":           filter,
			"retrieval_method": "filtered_vector_similarity",
			"context_length":   len(context),
			"sources_count":    len(sources),
			"model":            rag.llmProvider.GetName(),
		},
	}

	return result, nil
}

// buildContext builds context string from retrieved documents
func (rag *RAGChain) buildContext(sources []*RetrievalResult) string {
	if len(sources) == 0 {
		return "No relevant information found."
	}

	var contextParts []string
	totalLength := 0

	for i, source := range sources {
		content := source.Document.Content

		// Add source information if configured
		if rag.config.IncludeSourceInfo {
			sourceInfo := fmt.Sprintf("Source %d", i+1)
			if filename, exists := source.Document.Metadata["filename"]; exists {
				sourceInfo += fmt.Sprintf(" (%s)", filename)
			}
			if contentType, exists := source.Document.Metadata["content_type"]; exists {
				sourceInfo += fmt.Sprintf(" [%s]", contentType)
			}
			content = fmt.Sprintf("%s:\n%s", sourceInfo, content)
		}

		// Check if adding this content would exceed max context length
		if totalLength+len(content) > rag.config.MaxContextLength {
			// Truncate content to fit
			remaining := rag.config.MaxContextLength - totalLength
			if remaining > 100 { // Only add if we have reasonable space
				content = content[:remaining] + "..."
				contextParts = append(contextParts, content)
			}
			break
		}

		contextParts = append(contextParts, content)
		totalLength += len(content)
	}

	return strings.Join(contextParts, rag.config.ContextSeparator)
}

// generateAnswer generates an answer using the LLM
func (rag *RAGChain) generateAnswer(ctx context.Context, question, context string) (string, int, error) {
	// Render prompt template
	vars := map[string]interface{}{
		"question": question,
		"context":  context,
	}

	prompt, err := rag.promptTemplate.Render(ctx, vars)
	if err != nil {
		return "", 0, fmt.Errorf("failed to render prompt: %w", err)
	}

	// Generate response using LLM
	req := &providers.GenerateRequest{
		Messages: []providers.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   rag.config.MaxTokens,
		Temperature: rag.config.Temperature,
	}

	resp, err := rag.llmProvider.GenerateResponse(ctx, req)
	if err != nil {
		return "", 0, fmt.Errorf("LLM generation failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", 0, fmt.Errorf("no response choices returned")
	}

	answer := resp.Choices[0].Message.Content

	// Apply output parser if configured
	if rag.outputParser != nil {
		parsed, err := rag.outputParser.Parse(ctx, answer)
		if err != nil {
			// If parsing fails, return original answer
			return answer, resp.Usage.TotalTokens, nil
		}

		if parsedStr, ok := parsed.(string); ok {
			answer = parsedStr
		}
	}

	return answer, resp.Usage.TotalTokens, nil
}

// TravelRAGChain specializes RAG for travel queries
type TravelRAGChain struct {
	*RAGChain
	tracer trace.Tracer
}

// NewTravelRAGChain creates a new travel-specific RAG chain
func NewTravelRAGChain(retriever Retriever, llmProvider providers.LLMProvider, config *RAGConfig) *TravelRAGChain {
	// Create travel-specific prompt template
	travelPrompt := langchain.NewPromptTemplate(
		"travel_rag_prompt",
		`You are an expert travel advisor with extensive knowledge about destinations worldwide. Use the following travel information to provide helpful and accurate advice.

Travel Information:
{{.context}}

Traveler's Question: {{.question}}

Instructions:
- Provide detailed, practical travel advice based on the information provided
- Include specific recommendations for activities, accommodations, dining, and transportation when available
- Mention any important travel tips, cultural considerations, or safety information
- If budget information is available, provide cost-effective suggestions
- Be enthusiastic and inspiring while remaining factual
- If the information is insufficient, clearly state what additional details would be helpful

Travel Advice:`,
		[]string{"context", "question"},
	)

	ragChain := NewRAGChain(retriever, llmProvider, config)
	ragChain.SetPromptTemplate(travelPrompt)

	return &TravelRAGChain{
		RAGChain: ragChain,
		tracer:   otel.Tracer("rag.travel_chain"),
	}
}

// QueryDestination queries for destination-specific information
func (trag *TravelRAGChain) QueryDestination(ctx context.Context, destination, question string) (*RAGResult, error) {
	ctx, span := trag.tracer.Start(ctx, "travel_rag.query_destination")
	defer span.End()

	span.SetAttributes(
		attribute.String("destination", destination),
		attribute.String("question", question),
	)

	// Use travel retriever if available
	if travelRetriever, ok := trag.retriever.(*TravelRetriever); ok {
		// Retrieve destination-specific documents
		sources, err := travelRetriever.RetrieveForDestination(ctx, destination, question, trag.config.RetrievalLimit)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("destination retrieval failed: %w", err)
		}

		// Build context and generate answer
		context := trag.buildContext(sources)

		generationStart := time.Now()
		answer, tokensUsed, err := trag.generateAnswer(ctx, question, context)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("generation failed: %w", err)
		}
		generationTime := time.Since(generationStart)

		result := &RAGResult{
			Query:          question,
			Answer:         answer,
			Sources:        sources,
			Context:        context,
			TokensUsed:     tokensUsed,
			GenerationTime: generationTime,
			Metadata: map[string]interface{}{
				"destination":      destination,
				"retrieval_method": "destination_specific",
				"context_length":   len(context),
				"sources_count":    len(sources),
				"model":            trag.llmProvider.GetName(),
			},
		}

		return result, nil
	}

	// Fallback to regular query with destination in question
	enhancedQuestion := fmt.Sprintf("%s travel information for %s", question, destination)
	return trag.Query(ctx, enhancedQuestion)
}

// QueryByCategory queries for category-specific travel information
func (trag *TravelRAGChain) QueryByCategory(ctx context.Context, category, question string) (*RAGResult, error) {
	filter := map[string]interface{}{
		"content_type": category,
	}

	result, err := trag.QueryWithFilter(ctx, question, filter)
	if err != nil {
		return nil, err
	}

	result.Metadata["category"] = category
	return result, nil
}

// Helper function
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

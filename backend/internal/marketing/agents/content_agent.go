package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/exotic-travel-booking/backend/internal/llm/providers"
	"github.com/exotic-travel-booking/backend/internal/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ContentAgent handles AI-powered content generation for marketing
type ContentAgent struct {
	llmProvider providers.LLMProvider
	tracer      trace.Tracer
}

// NewContentAgent creates a new content generation agent
func NewContentAgent(llmProvider providers.LLMProvider) *ContentAgent {
	return &ContentAgent{
		llmProvider: llmProvider,
		tracer:      otel.Tracer("marketing.content_agent"),
	}
}

// ContentRequest represents a request for content generation
type ContentRequest struct {
	Type           models.ContentType `json:"type"`
	Platform       string             `json:"platform"`
	BrandVoice     BrandVoice         `json:"brand_voice"`
	TargetAudience TargetAudience     `json:"target_audience"`
	Objectives     []string           `json:"objectives"`
	Keywords       []string           `json:"keywords"`
	Tone           string             `json:"tone"`
	Length         string             `json:"length"`
	CallToAction   string             `json:"call_to_action"`
	Context        string             `json:"context"`
	Constraints    []string           `json:"constraints"`
}

// BrandVoice represents brand voice guidelines
type BrandVoice struct {
	Personality    []string `json:"personality"`
	Values         []string `json:"values"`
	DoList         []string `json:"do_list"`
	DontList       []string `json:"dont_list"`
	ExampleContent []string `json:"example_content"`
}

// TargetAudience represents the target audience for content
type TargetAudience struct {
	Demographics map[string]interface{} `json:"demographics"`
	Interests    []string               `json:"interests"`
	PainPoints   []string               `json:"pain_points"`
	Goals        []string               `json:"goals"`
	Platforms    []string               `json:"platforms"`
}

// ContentResponse represents the generated content
type ContentResponse struct {
	Title        string            `json:"title"`
	Body         string            `json:"body"`
	Hashtags     []string          `json:"hashtags,omitempty"`
	SEOData      SEOData           `json:"seo_data,omitempty"`
	Variations   []ContentVariation `json:"variations,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
	Confidence   float64           `json:"confidence"`
	Suggestions  []string          `json:"suggestions"`
}

// ContentVariation represents A/B testing variations
type ContentVariation struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	Hashtags    []string `json:"hashtags,omitempty"`
	TestFocus   string `json:"test_focus"`
	Description string `json:"description"`
}

// SEOData represents SEO optimization data
type SEOData struct {
	MetaTitle       string   `json:"meta_title"`
	MetaDescription string   `json:"meta_description"`
	Keywords        []string `json:"keywords"`
	ReadabilityScore float64 `json:"readability_score"`
	KeywordDensity  map[string]float64 `json:"keyword_density"`
}

// GenerateContent creates AI-powered marketing content
func (ca *ContentAgent) GenerateContent(ctx context.Context, req ContentRequest) (*ContentResponse, error) {
	ctx, span := ca.tracer.Start(ctx, "content_agent.generate_content")
	defer span.End()

	span.SetAttributes(
		attribute.String("content.type", string(req.Type)),
		attribute.String("content.platform", req.Platform),
		attribute.String("content.tone", req.Tone),
	)

	// Build the prompt based on content type and requirements
	prompt := ca.buildContentPrompt(req)

	// Generate content using LLM
	response, err := ca.llmProvider.GenerateCompletion(ctx, providers.CompletionRequest{
		Messages: []providers.Message{
			{
				Role:    "system",
				Content: ca.getSystemPrompt(req.Type),
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   2000,
	})

	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// Parse the response
	contentResp, err := ca.parseContentResponse(response.Content, req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to parse content response: %w", err)
	}

	// Generate A/B testing variations if requested
	if req.Type == models.ContentTypeAd || req.Type == models.ContentTypeSocialPost {
		variations, err := ca.generateVariations(ctx, req, contentResp)
		if err == nil {
			contentResp.Variations = variations
		}
	}

	// Generate SEO data for web content
	if req.Type == models.ContentTypeBlog || req.Type == models.ContentTypeLanding {
		seoData, err := ca.generateSEOData(ctx, req, contentResp)
		if err == nil {
			contentResp.SEOData = *seoData
		}
	}

	span.SetAttributes(
		attribute.Float64("content.confidence", contentResp.Confidence),
		attribute.Int("content.variations_count", len(contentResp.Variations)),
	)

	return contentResp, nil
}

// buildContentPrompt creates a detailed prompt for content generation
func (ca *ContentAgent) buildContentPrompt(req ContentRequest) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf("Generate %s content for %s platform.\n\n", req.Type, req.Platform))

	// Brand voice
	if len(req.BrandVoice.Personality) > 0 {
		prompt.WriteString(fmt.Sprintf("Brand Personality: %s\n", strings.Join(req.BrandVoice.Personality, ", ")))
	}
	if len(req.BrandVoice.Values) > 0 {
		prompt.WriteString(fmt.Sprintf("Brand Values: %s\n", strings.Join(req.BrandVoice.Values, ", ")))
	}

	// Target audience
	if len(req.TargetAudience.Interests) > 0 {
		prompt.WriteString(fmt.Sprintf("Target Interests: %s\n", strings.Join(req.TargetAudience.Interests, ", ")))
	}
	if len(req.TargetAudience.PainPoints) > 0 {
		prompt.WriteString(fmt.Sprintf("Audience Pain Points: %s\n", strings.Join(req.TargetAudience.PainPoints, ", ")))
	}

	// Objectives and keywords
	if len(req.Objectives) > 0 {
		prompt.WriteString(fmt.Sprintf("Objectives: %s\n", strings.Join(req.Objectives, ", ")))
	}
	if len(req.Keywords) > 0 {
		prompt.WriteString(fmt.Sprintf("Keywords to include: %s\n", strings.Join(req.Keywords, ", ")))
	}

	// Tone and style
	prompt.WriteString(fmt.Sprintf("Tone: %s\n", req.Tone))
	prompt.WriteString(fmt.Sprintf("Length: %s\n", req.Length))

	if req.CallToAction != "" {
		prompt.WriteString(fmt.Sprintf("Call to Action: %s\n", req.CallToAction))
	}

	if req.Context != "" {
		prompt.WriteString(fmt.Sprintf("Context: %s\n", req.Context))
	}

	// Constraints
	if len(req.Constraints) > 0 {
		prompt.WriteString(fmt.Sprintf("Constraints: %s\n", strings.Join(req.Constraints, ", ")))
	}

	prompt.WriteString("\nPlease provide the content in JSON format with title, body, and relevant metadata.")

	return prompt.String()
}

// getSystemPrompt returns the system prompt for different content types
func (ca *ContentAgent) getSystemPrompt(contentType models.ContentType) string {
	switch contentType {
	case models.ContentTypeAd:
		return "You are an expert advertising copywriter. Create compelling, conversion-focused ad copy that drives action while maintaining brand consistency."
	case models.ContentTypeSocialPost:
		return "You are a social media expert. Create engaging, shareable content that resonates with the target audience and encourages interaction."
	case models.ContentTypeEmail:
		return "You are an email marketing specialist. Create compelling email content that drives opens, clicks, and conversions while building relationships."
	case models.ContentTypeBlog:
		return "You are a content marketing expert. Create informative, engaging blog content that provides value while supporting business objectives."
	case models.ContentTypeLanding:
		return "You are a conversion copywriter. Create persuasive landing page content that guides visitors toward the desired action."
	default:
		return "You are a marketing content expert. Create high-quality, engaging content that aligns with brand voice and marketing objectives."
	}
}

// parseContentResponse parses the LLM response into structured content
func (ca *ContentAgent) parseContentResponse(response string, req ContentRequest) (*ContentResponse, error) {
	// Try to parse as JSON first
	var jsonResp ContentResponse
	if err := json.Unmarshal([]byte(response), &jsonResp); err == nil {
		jsonResp.Confidence = 0.85 // Default confidence
		return &jsonResp, nil
	}

	// Fallback: parse as plain text
	lines := strings.Split(response, "\n")
	contentResp := &ContentResponse{
		Metadata:   make(map[string]interface{}),
		Confidence: 0.75,
	}

	// Simple parsing logic for plain text responses
	var bodyLines []string
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if i == 0 && !strings.Contains(line, ":") {
			contentResp.Title = line
		} else {
			bodyLines = append(bodyLines, line)
		}
	}

	contentResp.Body = strings.Join(bodyLines, "\n")

	// Extract hashtags for social content
	if req.Type == models.ContentTypeSocialPost {
		contentResp.Hashtags = ca.extractHashtags(contentResp.Body)
	}

	return contentResp, nil
}

// extractHashtags extracts hashtags from content
func (ca *ContentAgent) extractHashtags(content string) []string {
	words := strings.Fields(content)
	var hashtags []string
	
	for _, word := range words {
		if strings.HasPrefix(word, "#") {
			hashtags = append(hashtags, word)
		}
	}
	
	return hashtags
}

// generateVariations creates A/B testing variations
func (ca *ContentAgent) generateVariations(ctx context.Context, req ContentRequest, original *ContentResponse) ([]ContentVariation, error) {
	// Implementation for generating A/B test variations
	// This would use the LLM to create different versions focusing on different aspects
	return []ContentVariation{}, nil
}

// generateSEOData creates SEO optimization data
func (ca *ContentAgent) generateSEOData(ctx context.Context, req ContentRequest, content *ContentResponse) (*SEOData, error) {
	// Implementation for SEO data generation
	// This would analyze the content and generate SEO recommendations
	return &SEOData{}, nil
}

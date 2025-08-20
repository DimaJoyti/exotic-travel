package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/exotic-travel-booking/backend/internal/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// VisualAgent handles AI-powered visual content generation
type VisualAgent struct {
	openAIAPIKey string
	dalleModel   string
	httpClient   *http.Client
	tracer       trace.Tracer
}

// NewVisualAgent creates a new visual content generation agent
func NewVisualAgent(openAIAPIKey string) *VisualAgent {
	return &VisualAgent{
		openAIAPIKey: openAIAPIKey,
		dalleModel:   "dall-e-3",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		tracer: otel.Tracer("marketing.visual_agent"),
	}
}

// VisualRequest represents a request for visual content generation
type VisualRequest struct {
	Type           models.AssetType       `json:"type"`
	Platform       string                 `json:"platform"`
	BrandGuidelines BrandVisualGuidelines `json:"brand_guidelines"`
	ContentContext string                 `json:"content_context"`
	Style          string                 `json:"style"`
	Dimensions     ImageDimensions        `json:"dimensions"`
	ColorScheme    []string               `json:"color_scheme"`
	Elements       []string               `json:"elements"`
	Mood           string                 `json:"mood"`
	Quality        string                 `json:"quality"`
	Variations     int                    `json:"variations"`
}

// BrandVisualGuidelines represents brand visual identity guidelines
type BrandVisualGuidelines struct {
	LogoURL        string            `json:"logo_url"`
	ColorPalette   map[string]string `json:"color_palette"`
	Typography     map[string]string `json:"typography"`
	VisualStyle    string            `json:"visual_style"`
	BrandElements  []string          `json:"brand_elements"`
	DoList         []string          `json:"do_list"`
	DontList       []string          `json:"dont_list"`
}

// ImageDimensions represents image size requirements
type ImageDimensions struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Ratio  string `json:"ratio"`
}

// VisualResponse represents the generated visual content
type VisualResponse struct {
	Images      []GeneratedImage       `json:"images"`
	Metadata    map[string]interface{} `json:"metadata"`
	Prompt      string                 `json:"prompt"`
	Style       string                 `json:"style"`
	Confidence  float64                `json:"confidence"`
	Suggestions []string               `json:"suggestions"`
}

// GeneratedImage represents a single generated image
type GeneratedImage struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	Prompt      string `json:"prompt"`
	Dimensions  ImageDimensions `json:"dimensions"`
	Format      string `json:"format"`
	Quality     string `json:"quality"`
	Variation   int    `json:"variation"`
}

// DALL-E API structures
type dalleRequest struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	N              int    `json:"n"`
	Size           string `json:"size"`
	Quality        string `json:"quality"`
	Style          string `json:"style"`
	ResponseFormat string `json:"response_format"`
}

type dalleResponse struct {
	Created int64 `json:"created"`
	Data    []struct {
		URL           string `json:"url"`
		RevisedPrompt string `json:"revised_prompt"`
	} `json:"data"`
}

// GenerateVisualContent creates AI-powered visual content
func (va *VisualAgent) GenerateVisualContent(ctx context.Context, req VisualRequest) (*VisualResponse, error) {
	ctx, span := va.tracer.Start(ctx, "visual_agent.generate_visual_content")
	defer span.End()

	span.SetAttributes(
		attribute.String("visual.type", string(req.Type)),
		attribute.String("visual.platform", req.Platform),
		attribute.String("visual.style", req.Style),
	)

	// Build the prompt based on request
	prompt := va.buildVisualPrompt(req)

	// Generate images using DALL-E
	images, err := va.generateWithDALLE(ctx, prompt, req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to generate visual content: %w", err)
	}

	// Create response
	response := &VisualResponse{
		Images:      images,
		Prompt:      prompt,
		Style:       req.Style,
		Confidence:  0.85,
		Metadata: map[string]interface{}{
			"platform":    req.Platform,
			"type":        req.Type,
			"dimensions":  req.Dimensions,
			"generated_at": time.Now(),
		},
		Suggestions: va.generateSuggestions(req),
	}

	span.SetAttributes(
		attribute.Int("images.count", len(images)),
		attribute.Float64("visual.confidence", response.Confidence),
	)

	return response, nil
}

// buildVisualPrompt creates a detailed prompt for image generation
func (va *VisualAgent) buildVisualPrompt(req VisualRequest) string {
	var prompt bytes.Buffer

	// Base description
	prompt.WriteString(fmt.Sprintf("Create a %s for %s platform. ", req.Type, req.Platform))

	// Content context
	if req.ContentContext != "" {
		prompt.WriteString(fmt.Sprintf("Content context: %s. ", req.ContentContext))
	}

	// Style and mood
	prompt.WriteString(fmt.Sprintf("Style: %s. Mood: %s. ", req.Style, req.Mood))

	// Brand guidelines
	if req.BrandGuidelines.VisualStyle != "" {
		prompt.WriteString(fmt.Sprintf("Brand visual style: %s. ", req.BrandGuidelines.VisualStyle))
	}

	// Color scheme
	if len(req.ColorScheme) > 0 {
		prompt.WriteString(fmt.Sprintf("Use colors: %v. ", req.ColorScheme))
	}

	// Elements to include
	if len(req.Elements) > 0 {
		prompt.WriteString(fmt.Sprintf("Include elements: %v. ", req.Elements))
	}

	// Brand do's and don'ts
	if len(req.BrandGuidelines.DoList) > 0 {
		prompt.WriteString(fmt.Sprintf("Brand requirements: %v. ", req.BrandGuidelines.DoList))
	}
	if len(req.BrandGuidelines.DontList) > 0 {
		prompt.WriteString(fmt.Sprintf("Avoid: %v. ", req.BrandGuidelines.DontList))
	}

	// Platform-specific optimizations
	prompt.WriteString(va.getPlatformOptimizations(req.Platform))

	// Quality and technical requirements
	prompt.WriteString("High quality, professional, marketing-ready image. ")

	return prompt.String()
}

// generateWithDALLE calls the DALL-E API to generate images
func (va *VisualAgent) generateWithDALLE(ctx context.Context, prompt string, req VisualRequest) ([]GeneratedImage, error) {
	// Prepare DALL-E request
	dalleReq := dalleRequest{
		Model:          va.dalleModel,
		Prompt:         prompt,
		N:              max(1, min(req.Variations, 4)), // DALL-E 3 supports max 1 image per request
		Size:           va.getDalleSize(req.Dimensions),
		Quality:        req.Quality,
		Style:          va.mapStyleToDALLE(req.Style),
		ResponseFormat: "url",
	}

	// Convert to JSON
	reqBody, err := json.Marshal(dalleReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DALL-E request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/images/generations", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+va.openAIAPIKey)

	// Make request
	resp, err := va.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make DALL-E request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read DALL-E response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DALL-E API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var dalleResp dalleResponse
	if err := json.Unmarshal(body, &dalleResp); err != nil {
		return nil, fmt.Errorf("failed to parse DALL-E response: %w", err)
	}

	// Convert to our format
	var images []GeneratedImage
	for i, img := range dalleResp.Data {
		images = append(images, GeneratedImage{
			ID:         fmt.Sprintf("img_%d_%d", dalleResp.Created, i),
			URL:        img.URL,
			Prompt:     img.RevisedPrompt,
			Dimensions: req.Dimensions,
			Format:     "png",
			Quality:    req.Quality,
			Variation:  i + 1,
		})
	}

	return images, nil
}

// getDalleSize converts dimensions to DALL-E size format
func (va *VisualAgent) getDalleSize(dims ImageDimensions) string {
	// DALL-E 3 supported sizes
	if dims.Width == 1024 && dims.Height == 1024 {
		return "1024x1024"
	}
	if dims.Width == 1792 && dims.Height == 1024 {
		return "1792x1024"
	}
	if dims.Width == 1024 && dims.Height == 1792 {
		return "1024x1792"
	}
	
	// Default to square
	return "1024x1024"
}

// mapStyleToDALLE maps our style to DALL-E style parameter
func (va *VisualAgent) mapStyleToDALLE(style string) string {
	switch style {
	case "natural", "realistic", "photographic":
		return "natural"
	case "vivid", "artistic", "creative":
		return "vivid"
	default:
		return "natural"
	}
}

// getPlatformOptimizations returns platform-specific optimization instructions
func (va *VisualAgent) getPlatformOptimizations(platform string) string {
	switch platform {
	case "instagram":
		return "Optimized for Instagram: vibrant colors, high contrast, mobile-friendly composition. "
	case "facebook":
		return "Optimized for Facebook: clear focal point, readable text overlay space, engaging composition. "
	case "twitter":
		return "Optimized for Twitter: simple composition, bold elements, works well at small sizes. "
	case "linkedin":
		return "Optimized for LinkedIn: professional appearance, business-appropriate, clean design. "
	case "youtube":
		return "Optimized for YouTube: thumbnail-ready, clear subject, compelling visual hook. "
	case "google":
		return "Optimized for Google Ads: clear product focus, minimal text, strong call-to-action visual. "
	default:
		return "Optimized for digital marketing: versatile, high-impact, professional quality. "
	}
}

// generateSuggestions creates optimization suggestions
func (va *VisualAgent) generateSuggestions(req VisualRequest) []string {
	suggestions := []string{}

	// Platform-specific suggestions
	switch req.Platform {
	case "instagram":
		suggestions = append(suggestions, "Consider adding Instagram Stories format (9:16)")
		suggestions = append(suggestions, "Test with and without text overlay")
	case "facebook":
		suggestions = append(suggestions, "Create carousel variations for better engagement")
		suggestions = append(suggestions, "Consider video thumbnail version")
	case "linkedin":
		suggestions = append(suggestions, "Create professional headshot variations")
		suggestions = append(suggestions, "Consider infographic-style layouts")
	}

	// General suggestions
	suggestions = append(suggestions, "A/B test different color variations")
	suggestions = append(suggestions, "Consider seasonal adaptations")
	suggestions = append(suggestions, "Test mobile vs desktop optimized versions")

	return suggestions
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

package content

import (
	"context"
	"fmt"
	"time"

	"github.com/exotic-travel-booking/backend/internal/marketing/agents"
	"github.com/exotic-travel-booking/backend/internal/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Generator handles content generation workflows
type Generator struct {
	contentAgent *agents.ContentAgent
	repository   Repository
	tracer       trace.Tracer
}

// Repository interface for content persistence
type Repository interface {
	CreateContent(ctx context.Context, content *models.Content) error
	UpdateContent(ctx context.Context, content *models.Content) error
	GetContentByID(ctx context.Context, id int) (*models.Content, error)
	GetContentByCampaign(ctx context.Context, campaignID int) ([]models.Content, error)
	GetBrandByID(ctx context.Context, id int) (*models.Brand, error)
	GetCampaignByID(ctx context.Context, id int) (*models.Campaign, error)
}

// NewGenerator creates a new content generator
func NewGenerator(contentAgent *agents.ContentAgent, repo Repository) *Generator {
	return &Generator{
		contentAgent: contentAgent,
		repository:   repo,
		tracer:       otel.Tracer("marketing.content_generator"),
	}
}

// GenerationRequest represents a request to generate content
type GenerationRequest struct {
	CampaignID      int                    `json:"campaign_id"`
	ContentType     models.ContentType     `json:"content_type"`
	Platform        string                 `json:"platform"`
	Title           string                 `json:"title,omitempty"`
	Brief           string                 `json:"brief"`
	Keywords        []string               `json:"keywords"`
	Tone            string                 `json:"tone"`
	Length          string                 `json:"length"`
	CallToAction    string                 `json:"call_to_action"`
	GenerateVariations bool                `json:"generate_variations"`
	VariationCount  int                    `json:"variation_count"`
	CreatedBy       int                    `json:"created_by"`
}

// GenerationResponse represents the result of content generation
type GenerationResponse struct {
	Content    *models.Content   `json:"content"`
	Variations []models.Content  `json:"variations,omitempty"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// Generate creates new marketing content using AI
func (g *Generator) Generate(ctx context.Context, req GenerationRequest) (*GenerationResponse, error) {
	ctx, span := g.tracer.Start(ctx, "content_generator.generate")
	defer span.End()

	span.SetAttributes(
		attribute.Int("campaign.id", req.CampaignID),
		attribute.String("content.type", string(req.ContentType)),
		attribute.String("content.platform", req.Platform),
		attribute.Bool("generate.variations", req.GenerateVariations),
	)

	// Get campaign and brand information
	campaign, err := g.repository.GetCampaignByID(ctx, req.CampaignID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get campaign: %w", err)
	}

	brand, err := g.repository.GetBrandByID(ctx, campaign.BrandID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get brand: %w", err)
	}

	// Build content generation request
	contentReq := g.buildContentRequest(req, campaign, brand)

	// Generate content using AI agent
	aiResponse, err := g.contentAgent.GenerateContent(ctx, contentReq)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// Create main content record
	content := &models.Content{
		CampaignID:  req.CampaignID,
		Type:        req.ContentType,
		Title:       aiResponse.Title,
		Body:        aiResponse.Body,
		Platform:    req.Platform,
		BrandVoice:  g.extractBrandVoice(brand),
		SEOData:     g.convertSEOData(aiResponse.SEOData),
		Metadata:    g.convertMetadata(aiResponse.Metadata),
		Status:      models.ContentStatusDraft,
		CreatedBy:   req.CreatedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save main content
	if err := g.repository.CreateContent(ctx, content); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to save content: %w", err)
	}

	response := &GenerationResponse{
		Content:  content,
		Metadata: aiResponse.Metadata,
	}

	// Generate and save variations if requested
	if req.GenerateVariations && len(aiResponse.Variations) > 0 {
		variations, err := g.createVariations(ctx, content, aiResponse.Variations, req.CreatedBy)
		if err != nil {
			// Log error but don't fail the main generation
			span.RecordError(err)
		} else {
			response.Variations = variations
		}
	}

	span.SetAttributes(
		attribute.Int("content.id", content.ID),
		attribute.Int("variations.count", len(response.Variations)),
	)

	return response, nil
}

// buildContentRequest converts the generation request to an agent request
func (g *Generator) buildContentRequest(req GenerationRequest, campaign *models.Campaign, brand *models.Brand) agents.ContentRequest {
	// Extract brand voice from brand guidelines
	brandVoice := agents.BrandVoice{}
	if brand.VoiceGuidelines != nil {
		if personality, ok := brand.VoiceGuidelines["personality"].([]interface{}); ok {
			for _, p := range personality {
				if str, ok := p.(string); ok {
					brandVoice.Personality = append(brandVoice.Personality, str)
				}
			}
		}
		if values, ok := brand.VoiceGuidelines["values"].([]interface{}); ok {
			for _, v := range values {
				if str, ok := v.(string); ok {
					brandVoice.Values = append(brandVoice.Values, str)
				}
			}
		}
	}

	// Extract target audience from campaign
	targetAudience := agents.TargetAudience{}
	if campaign.TargetAudience != nil {
		if demographics, ok := campaign.TargetAudience["demographics"].(map[string]interface{}); ok {
			targetAudience.Demographics = demographics
		}
		if interests, ok := campaign.TargetAudience["interests"].([]interface{}); ok {
			for _, i := range interests {
				if str, ok := i.(string); ok {
					targetAudience.Interests = append(targetAudience.Interests, str)
				}
			}
		}
	}

	// Extract objectives from campaign
	var objectives []string
	if campaign.Objectives != nil {
		if objList, ok := campaign.Objectives["objectives"].([]interface{}); ok {
			for _, obj := range objList {
				if str, ok := obj.(string); ok {
					objectives = append(objectives, str)
				}
			}
		}
	}

	return agents.ContentRequest{
		Type:           req.ContentType,
		Platform:       req.Platform,
		BrandVoice:     brandVoice,
		TargetAudience: targetAudience,
		Objectives:     objectives,
		Keywords:       req.Keywords,
		Tone:           req.Tone,
		Length:         req.Length,
		CallToAction:   req.CallToAction,
		Context:        req.Brief,
	}
}

// createVariations creates and saves content variations
func (g *Generator) createVariations(ctx context.Context, parent *models.Content, variations []agents.ContentVariation, createdBy int) ([]models.Content, error) {
	var results []models.Content
	variationGroup := fmt.Sprintf("var_%d_%d", parent.ID, time.Now().Unix())

	for _, variation := range variations {
		content := &models.Content{
			CampaignID:      parent.CampaignID,
			Type:            parent.Type,
			Title:           variation.Title,
			Body:            variation.Body,
			Platform:        parent.Platform,
			BrandVoice:      parent.BrandVoice,
			Status:          models.ContentStatusDraft,
			VariationGroup:  &variationGroup,
			ParentContentID: &parent.ID,
			CreatedBy:       createdBy,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		// Add variation-specific metadata
		metadata := make(models.JSON)
		metadata["test_focus"] = variation.TestFocus
		metadata["description"] = variation.Description
		metadata["variation_id"] = variation.ID
		content.Metadata = metadata

		if err := g.repository.CreateContent(ctx, content); err != nil {
			return nil, fmt.Errorf("failed to save variation: %w", err)
		}

		results = append(results, *content)
	}

	return results, nil
}

// extractBrandVoice extracts brand voice as a string summary
func (g *Generator) extractBrandVoice(brand *models.Brand) string {
	if brand.VoiceGuidelines == nil {
		return "professional"
	}

	var voice []string
	if personality, ok := brand.VoiceGuidelines["personality"].([]interface{}); ok {
		for _, p := range personality {
			if str, ok := p.(string); ok {
				voice = append(voice, str)
			}
		}
	}

	if len(voice) > 0 {
		return voice[0] // Return primary personality trait
	}

	return "professional"
}

// convertSEOData converts agent SEO data to model JSON
func (g *Generator) convertSEOData(seoData agents.SEOData) models.JSON {
	result := make(models.JSON)
	result["meta_title"] = seoData.MetaTitle
	result["meta_description"] = seoData.MetaDescription
	result["keywords"] = seoData.Keywords
	result["readability_score"] = seoData.ReadabilityScore
	result["keyword_density"] = seoData.KeywordDensity
	return result
}

// convertMetadata converts agent metadata to model JSON
func (g *Generator) convertMetadata(metadata map[string]interface{}) models.JSON {
	result := make(models.JSON)
	for k, v := range metadata {
		result[k] = v
	}
	return result
}

// GetContentHistory retrieves content generation history for a campaign
func (g *Generator) GetContentHistory(ctx context.Context, campaignID int) ([]models.Content, error) {
	ctx, span := g.tracer.Start(ctx, "content_generator.get_history")
	defer span.End()

	span.SetAttributes(attribute.Int("campaign.id", campaignID))

	contents, err := g.repository.GetContentByCampaign(ctx, campaignID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get content history: %w", err)
	}

	span.SetAttributes(attribute.Int("content.count", len(contents)))
	return contents, nil
}

// RegenerateContent creates a new version of existing content
func (g *Generator) RegenerateContent(ctx context.Context, contentID int, modifications map[string]interface{}) (*models.Content, error) {
	ctx, span := g.tracer.Start(ctx, "content_generator.regenerate")
	defer span.End()

	span.SetAttributes(attribute.Int("content.id", contentID))

	// Get existing content
	existing, err := g.repository.GetContentByID(ctx, contentID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get existing content: %w", err)
	}

	// Apply modifications and regenerate
	// Implementation would modify the generation request based on the modifications map
	// and create new content

	return existing, nil
}

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/exotic-travel-booking/backend/internal/marketing/content"
	"github.com/exotic-travel-booking/backend/internal/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// MarketingHandler handles marketing-related API endpoints
type MarketingHandler struct {
	contentGenerator *content.Generator
	tracer           trace.Tracer
}

// NewMarketingHandler creates a new marketing handler
func NewMarketingHandler(contentGenerator *content.Generator) *MarketingHandler {
	return &MarketingHandler{
		contentGenerator: contentGenerator,
		tracer:           otel.Tracer("api.marketing_handler"),
	}
}

// GenerateContentRequest represents the API request for content generation
type GenerateContentRequest struct {
	CampaignID         int                `json:"campaign_id"`
	ContentType        models.ContentType `json:"content_type"`
	Platform           string             `json:"platform"`
	Title              string             `json:"title,omitempty"`
	Brief              string             `json:"brief"`
	Keywords           []string           `json:"keywords"`
	Tone               string             `json:"tone"`
	Length             string             `json:"length"`
	CallToAction       string             `json:"call_to_action"`
	GenerateVariations bool               `json:"generate_variations"`
	VariationCount     int                `json:"variation_count"`
}

// GenerateContentResponse represents the API response for content generation
type GenerateContentResponse struct {
	Success    bool                   `json:"success"`
	Content    *models.Content        `json:"content"`
	Variations []models.Content       `json:"variations,omitempty"`
	Metadata   map[string]interface{} `json:"metadata"`
	Message    string                 `json:"message,omitempty"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}

// GenerateContent handles POST /api/v1/marketing/content/generate
func (h *MarketingHandler) GenerateContent(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "marketing_handler.generate_content")
	defer span.End()

	if r.Method != http.MethodPost {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req GenerateContentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validateGenerateContentRequest(req); err != nil {
		span.RecordError(err)
		h.writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get user ID from context (assuming it's set by auth middleware)
	userID := h.getUserIDFromContext(r)
	if userID == 0 {
		h.writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	span.SetAttributes(
		attribute.Int("campaign.id", req.CampaignID),
		attribute.String("content.type", string(req.ContentType)),
		attribute.String("content.platform", req.Platform),
		attribute.Int("user.id", userID),
	)

	// Convert to service request
	genReq := content.GenerationRequest{
		CampaignID:         req.CampaignID,
		ContentType:        req.ContentType,
		Platform:           req.Platform,
		Title:              req.Title,
		Brief:              req.Brief,
		Keywords:           req.Keywords,
		Tone:               req.Tone,
		Length:             req.Length,
		CallToAction:       req.CallToAction,
		GenerateVariations: req.GenerateVariations,
		VariationCount:     req.VariationCount,
		CreatedBy:          userID,
	}

	// Generate content
	result, err := h.contentGenerator.Generate(ctx, genReq)
	if err != nil {
		span.RecordError(err)
		h.writeError(w, "Failed to generate content", http.StatusInternalServerError)
		return
	}

	// Return response
	response := GenerateContentResponse{
		Success:    true,
		Content:    result.Content,
		Variations: result.Variations,
		Metadata:   result.Metadata,
		Message:    "Content generated successfully",
	}

	span.SetAttributes(
		attribute.Int("content.id", result.Content.ID),
		attribute.Int("variations.count", len(result.Variations)),
	)

	h.writeJSON(w, response, http.StatusOK)
}

// GetContentHistory handles GET /api/v1/marketing/content/history/{campaignId}
func (h *MarketingHandler) GetContentHistory(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "marketing_handler.get_content_history")
	defer span.End()

	if r.Method != http.MethodGet {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract campaign ID from URL path
	campaignIDStr := r.URL.Path[len("/api/v1/marketing/content/history/"):]
	campaignID, err := strconv.Atoi(campaignIDStr)
	if err != nil {
		h.writeError(w, "Invalid campaign ID", http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.Int("campaign.id", campaignID))

	// Get content history
	contents, err := h.contentGenerator.GetContentHistory(ctx, campaignID)
	if err != nil {
		span.RecordError(err)
		h.writeError(w, "Failed to get content history", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":  true,
		"contents": contents,
		"count":    len(contents),
	}

	span.SetAttributes(attribute.Int("content.count", len(contents)))
	h.writeJSON(w, response, http.StatusOK)
}

// RegenerateContent handles POST /api/v1/marketing/content/{contentId}/regenerate
func (h *MarketingHandler) RegenerateContent(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "marketing_handler.regenerate_content")
	defer span.End()

	if r.Method != http.MethodPost {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract content ID from URL path
	contentIDStr := r.URL.Path[len("/api/v1/marketing/content/"):]
	contentIDStr = contentIDStr[:len(contentIDStr)-len("/regenerate")]
	contentID, err := strconv.Atoi(contentIDStr)
	if err != nil {
		h.writeError(w, "Invalid content ID", http.StatusBadRequest)
		return
	}

	// Parse modifications from request body
	var modifications map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&modifications); err != nil {
		span.RecordError(err)
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.Int("content.id", contentID))

	// Regenerate content
	newContent, err := h.contentGenerator.RegenerateContent(ctx, contentID, modifications)
	if err != nil {
		span.RecordError(err)
		h.writeError(w, "Failed to regenerate content", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"content": newContent,
		"message": "Content regenerated successfully",
	}

	h.writeJSON(w, response, http.StatusOK)
}

// validateGenerateContentRequest validates the content generation request
func (h *MarketingHandler) validateGenerateContentRequest(req GenerateContentRequest) error {
	if req.CampaignID <= 0 {
		return fmt.Errorf("campaign_id is required")
	}

	if req.ContentType == "" {
		return fmt.Errorf("content_type is required")
	}

	if req.Platform == "" {
		return fmt.Errorf("platform is required")
	}

	if req.Brief == "" {
		return fmt.Errorf("brief is required")
	}

	// Validate content type
	validTypes := []models.ContentType{
		models.ContentTypeAd,
		models.ContentTypeSocialPost,
		models.ContentTypeEmail,
		models.ContentTypeBlog,
		models.ContentTypeLanding,
		models.ContentTypeVideo,
	}

	valid := false
	for _, validType := range validTypes {
		if req.ContentType == validType {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid content_type")
	}

	return nil
}

// getUserIDFromContext extracts user ID from request context
func (h *MarketingHandler) getUserIDFromContext(r *http.Request) int {
	// This would typically extract the user ID from JWT token or session
	// For now, return a placeholder
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		if id, err := strconv.Atoi(userID); err == nil {
			return id
		}
	}
	return 1 // Default user for demo
}

// writeJSON writes a JSON response
func (h *MarketingHandler) writeJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log error but don't expose it to client
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// writeError writes an error response
func (h *MarketingHandler) writeError(w http.ResponseWriter, message string, statusCode int) {
	response := ErrorResponse{
		Success: false,
		Error:   message,
	}
	
	h.writeJSON(w, response, statusCode)
}

// HealthCheck handles GET /api/v1/marketing/health
func (h *MarketingHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"success":   true,
		"service":   "marketing-api",
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	}

	h.writeJSON(w, response, http.StatusOK)
}

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/exotic-travel-booking/backend/internal/marketing/integrations"
	"github.com/exotic-travel-booking/backend/internal/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// IntegrationHandler handles integration-related API endpoints
type IntegrationHandler struct {
	integrationManager *integrations.IntegrationManager
	tracer             trace.Tracer
}

// NewIntegrationHandler creates a new integration handler
func NewIntegrationHandler(integrationManager *integrations.IntegrationManager) *IntegrationHandler {
	return &IntegrationHandler{
		integrationManager: integrationManager,
		tracer:             otel.Tracer("api.integration_handler"),
	}
}

// ConnectPlatformRequest represents the request to connect a platform
type ConnectPlatformRequest struct {
	Platform    string            `json:"platform"`
	Credentials map[string]string `json:"credentials"`
}

// ConnectPlatformResponse represents the response for platform connection
type ConnectPlatformResponse struct {
	Success     bool                `json:"success"`
	Integration *models.Integration `json:"integration"`
	Message     string              `json:"message"`
}

// ListIntegrationsResponse represents the response for listing integrations
type ListIntegrationsResponse struct {
	Success      bool                                        `json:"success"`
	Integrations []models.Integration                        `json:"integrations"`
	Health       map[string]integrations.IntegrationHealth  `json:"health"`
	Available    []string                                    `json:"available_platforms"`
}

// SyncCampaignRequest represents the request to sync a campaign
type SyncCampaignRequest struct {
	CampaignID int      `json:"campaign_id"`
	Platforms  []string `json:"platforms"`
}

// SyncCampaignResponse represents the response for campaign sync
type SyncCampaignResponse struct {
	Success   bool                                           `json:"success"`
	Results   map[string][]*integrations.PlatformCampaign   `json:"results"`
	Message   string                                         `json:"message"`
}

// ConnectPlatform handles POST /api/v1/marketing/integrations/connect
func (h *IntegrationHandler) ConnectPlatform(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "integration_handler.connect_platform")
	defer span.End()

	if r.Method != http.MethodPost {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req ConnectPlatformRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Platform == "" {
		h.writeError(w, "Platform is required", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID := h.getUserIDFromContext(r)
	if userID == 0 {
		h.writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	span.SetAttributes(
		attribute.String("platform", req.Platform),
		attribute.Int("user.id", userID),
	)

	// Connect to platform
	integration, err := h.integrationManager.ConnectPlatform(ctx, req.Platform, userID, req.Credentials)
	if err != nil {
		span.RecordError(err)
		h.writeError(w, fmt.Sprintf("Failed to connect to %s: %v", req.Platform, err), http.StatusBadRequest)
		return
	}

	response := ConnectPlatformResponse{
		Success:     true,
		Integration: integration,
		Message:     fmt.Sprintf("Successfully connected to %s", req.Platform),
	}

	span.SetAttributes(
		attribute.Int("integration.id", integration.ID),
		attribute.String("integration.account_id", integration.AccountID),
	)

	h.writeJSON(w, response, http.StatusOK)
}

// ListIntegrations handles GET /api/v1/marketing/integrations
func (h *IntegrationHandler) ListIntegrations(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "integration_handler.list_integrations")
	defer span.End()

	if r.Method != http.MethodGet {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context
	userID := h.getUserIDFromContext(r)
	if userID == 0 {
		h.writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	span.SetAttributes(attribute.Int("user.id", userID))

	// Get user integrations (this would come from repository)
	integrations := []models.Integration{
		{
			ID:          1,
			Platform:    "google_ads",
			AccountID:   "123456789",
			Status:      models.IntegrationStatusActive,
			CreatedBy:   userID,
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          2,
			Platform:    "facebook_ads",
			AccountID:   "act_987654321",
			Status:      models.IntegrationStatusActive,
			CreatedBy:   userID,
			CreatedAt:   time.Now().Add(-48 * time.Hour),
			UpdatedAt:   time.Now(),
		},
	}

	// Get integration health
	health, err := h.integrationManager.GetIntegrationHealth(ctx, userID)
	if err != nil {
		span.RecordError(err)
		// Continue with empty health map
		health = make(map[string]integrations.IntegrationHealth)
	}

	// Get available platforms
	available := h.integrationManager.ListAvailableIntegrations()

	response := ListIntegrationsResponse{
		Success:      true,
		Integrations: integrations,
		Health:       health,
		Available:    available,
	}

	span.SetAttributes(
		attribute.Int("integrations.count", len(integrations)),
		attribute.StringSlice("available.platforms", available),
	)

	h.writeJSON(w, response, http.StatusOK)
}

// DisconnectPlatform handles DELETE /api/v1/marketing/integrations/{platform}
func (h *IntegrationHandler) DisconnectPlatform(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "integration_handler.disconnect_platform")
	defer span.End()

	if r.Method != http.MethodDelete {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract platform from URL path
	platform := r.URL.Path[len("/api/v1/marketing/integrations/"):]
	if platform == "" {
		h.writeError(w, "Platform is required", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID := h.getUserIDFromContext(r)
	if userID == 0 {
		h.writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	span.SetAttributes(
		attribute.String("platform", platform),
		attribute.Int("user.id", userID),
	)

	// Disconnect platform
	if err := h.integrationManager.DisconnectPlatform(ctx, platform, userID); err != nil {
		span.RecordError(err)
		h.writeError(w, fmt.Sprintf("Failed to disconnect from %s: %v", platform, err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Successfully disconnected from %s", platform),
	}

	h.writeJSON(w, response, http.StatusOK)
}

// ValidateIntegration handles POST /api/v1/marketing/integrations/{integrationId}/validate
func (h *IntegrationHandler) ValidateIntegration(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "integration_handler.validate_integration")
	defer span.End()

	if r.Method != http.MethodPost {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract integration ID from URL path
	pathParts := r.URL.Path[len("/api/v1/marketing/integrations/"):]
	integrationIDStr := pathParts[:len(pathParts)-len("/validate")]
	integrationID, err := strconv.Atoi(integrationIDStr)
	if err != nil {
		h.writeError(w, "Invalid integration ID", http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.Int("integration.id", integrationID))

	// Validate integration
	if err := h.integrationManager.ValidateIntegration(ctx, integrationID); err != nil {
		span.RecordError(err)
		h.writeError(w, fmt.Sprintf("Integration validation failed: %v", err), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Integration is valid and working",
	}

	h.writeJSON(w, response, http.StatusOK)
}

// RefreshIntegration handles POST /api/v1/marketing/integrations/{integrationId}/refresh
func (h *IntegrationHandler) RefreshIntegration(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "integration_handler.refresh_integration")
	defer span.End()

	if r.Method != http.MethodPost {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract integration ID from URL path
	pathParts := r.URL.Path[len("/api/v1/marketing/integrations/"):]
	integrationIDStr := pathParts[:len(pathParts)-len("/refresh")]
	integrationID, err := strconv.Atoi(integrationIDStr)
	if err != nil {
		h.writeError(w, "Invalid integration ID", http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.Int("integration.id", integrationID))

	// Refresh integration
	integration, err := h.integrationManager.RefreshIntegration(ctx, integrationID)
	if err != nil {
		span.RecordError(err)
		h.writeError(w, fmt.Sprintf("Failed to refresh integration: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":     true,
		"integration": integration,
		"message":     "Integration refreshed successfully",
	}

	h.writeJSON(w, response, http.StatusOK)
}

// SyncCampaign handles POST /api/v1/marketing/integrations/sync-campaign
func (h *IntegrationHandler) SyncCampaign(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "integration_handler.sync_campaign")
	defer span.End()

	if r.Method != http.MethodPost {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req SyncCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.CampaignID <= 0 {
		h.writeError(w, "Campaign ID is required", http.StatusBadRequest)
		return
	}
	if len(req.Platforms) == 0 {
		h.writeError(w, "At least one platform is required", http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.Int("campaign.id", req.CampaignID),
		attribute.StringSlice("platforms", req.Platforms),
	)

	// Get campaign (this would come from campaign repository)
	campaign := &models.Campaign{
		ID:          req.CampaignID,
		Name:        "Sample Campaign",
		Type:        models.CampaignTypeSocial,
		Status:      models.CampaignStatusActive,
		Budget:      1000.0,
		StartDate:   time.Now(),
		CreatedBy:   h.getUserIDFromContext(r),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Sync campaign to platforms
	results, err := h.integrationManager.BulkSyncCampaigns(ctx, []*models.Campaign{campaign}, req.Platforms)
	if err != nil {
		span.RecordError(err)
		h.writeError(w, fmt.Sprintf("Failed to sync campaign: %v", err), http.StatusInternalServerError)
		return
	}

	response := SyncCampaignResponse{
		Success: true,
		Results: results,
		Message: "Campaign synced successfully",
	}

	// Count total synced campaigns
	totalSynced := 0
	for _, platformCampaigns := range results {
		totalSynced += len(platformCampaigns)
	}

	span.SetAttributes(attribute.Int("synced.campaigns", totalSynced))

	h.writeJSON(w, response, http.StatusOK)
}

// GetOAuthURL handles GET /api/v1/marketing/integrations/{platform}/oauth-url
func (h *IntegrationHandler) GetOAuthURL(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "integration_handler.get_oauth_url")
	defer span.End()

	if r.Method != http.MethodGet {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract platform from URL path
	pathParts := r.URL.Path[len("/api/v1/marketing/integrations/"):]
	platform := pathParts[:len(pathParts)-len("/oauth-url")]
	if platform == "" {
		h.writeError(w, "Platform is required", http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.String("platform", platform))

	// Generate OAuth URL based on platform
	var oauthURL string
	var scopes []string

	switch platform {
	case "google_ads":
		oauthURL = "https://accounts.google.com/oauth2/auth"
		scopes = []string{"https://www.googleapis.com/auth/adwords"}
	case "facebook_ads":
		oauthURL = "https://www.facebook.com/v18.0/dialog/oauth"
		scopes = []string{"ads_management", "ads_read", "business_management"}
	default:
		h.writeError(w, "Unsupported platform", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success":    true,
		"oauth_url":  oauthURL,
		"scopes":     scopes,
		"platform":   platform,
		"state":      fmt.Sprintf("state_%d_%s", time.Now().Unix(), platform),
	}

	h.writeJSON(w, response, http.StatusOK)
}

// Helper methods

func (h *IntegrationHandler) getUserIDFromContext(r *http.Request) int {
	// This would typically extract the user ID from JWT token or session
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		if id, err := strconv.Atoi(userID); err == nil {
			return id
		}
	}
	return 1 // Default user for demo
}

func (h *IntegrationHandler) writeJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *IntegrationHandler) writeError(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]interface{}{
		"success": false,
		"error":   message,
	}
	
	h.writeJSON(w, response, statusCode)
}

package integrations

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/exotic-travel-booking/backend/internal/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// IntegrationManager manages all platform integrations
type IntegrationManager struct {
	integrations map[string]PlatformIntegration
	repository   IntegrationRepository
	tracer       trace.Tracer
	mu           sync.RWMutex
}

// IntegrationRepository interface for integration data persistence
type IntegrationRepository interface {
	CreateIntegration(ctx context.Context, integration *models.Integration) error
	UpdateIntegration(ctx context.Context, integration *models.Integration) error
	GetIntegration(ctx context.Context, id int) (*models.Integration, error)
	GetIntegrationByPlatform(ctx context.Context, platform string, userID int) (*models.Integration, error)
	ListIntegrations(ctx context.Context, userID int) ([]models.Integration, error)
	DeleteIntegration(ctx context.Context, id int) error
}

// NewIntegrationManager creates a new integration manager
func NewIntegrationManager(repo IntegrationRepository) *IntegrationManager {
	return &IntegrationManager{
		integrations: make(map[string]PlatformIntegration),
		repository:   repo,
		tracer:       otel.Tracer("marketing.integration_manager"),
	}
}

// RegisterIntegration registers a platform integration
func (im *IntegrationManager) RegisterIntegration(platform string, integration PlatformIntegration) {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.integrations[platform] = integration
}

// GetIntegration retrieves a platform integration
func (im *IntegrationManager) GetIntegration(platform string) (PlatformIntegration, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()
	
	integration, exists := im.integrations[platform]
	if !exists {
		return nil, fmt.Errorf("integration for platform %s not found", platform)
	}
	
	return integration, nil
}

// ListAvailableIntegrations returns all available platform integrations
func (im *IntegrationManager) ListAvailableIntegrations() []string {
	im.mu.RLock()
	defer im.mu.RUnlock()
	
	platforms := make([]string, 0, len(im.integrations))
	for platform := range im.integrations {
		platforms = append(platforms, platform)
	}
	
	return platforms
}

// ConnectPlatform connects to a platform and stores the integration
func (im *IntegrationManager) ConnectPlatform(ctx context.Context, platform string, userID int, credentials map[string]string) (*models.Integration, error) {
	ctx, span := im.tracer.Start(ctx, "integration_manager.connect_platform")
	defer span.End()

	span.SetAttributes(
		attribute.String("platform", platform),
		attribute.Int("user_id", userID),
	)

	// Get platform integration
	platformIntegration, err := im.GetIntegration(platform)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get integration: %w", err)
	}

	// Authenticate with platform
	authResult, err := platformIntegration.Authenticate(ctx, credentials)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to authenticate with %s: %w", platform, err)
	}

	// Create integration record
	integration := &models.Integration{
		Platform:     platform,
		AccountID:    authResult.AccountID,
		AccessToken:  authResult.AccessToken,
		RefreshToken: authResult.RefreshToken,
		ExpiresAt:    &authResult.ExpiresAt,
		Status:       models.IntegrationStatusActive,
		Config: models.JSON{
			"account_name": authResult.AccountName,
			"scope":        authResult.Scope,
			"token_type":   authResult.TokenType,
		},
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to database
	if err := im.repository.CreateIntegration(ctx, integration); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to save integration: %w", err)
	}

	span.SetAttributes(
		attribute.Int("integration.id", integration.ID),
		attribute.String("integration.account_id", integration.AccountID),
	)

	return integration, nil
}

// RefreshIntegration refreshes an integration's access token
func (im *IntegrationManager) RefreshIntegration(ctx context.Context, integrationID int) (*models.Integration, error) {
	ctx, span := im.tracer.Start(ctx, "integration_manager.refresh_integration")
	defer span.End()

	span.SetAttributes(attribute.Int("integration.id", integrationID))

	// Get integration from database
	integration, err := im.repository.GetIntegration(ctx, integrationID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get integration: %w", err)
	}

	// Get platform integration
	platformIntegration, err := im.GetIntegration(integration.Platform)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get platform integration: %w", err)
	}

	// Refresh token
	authResult, err := platformIntegration.RefreshToken(ctx, integration.RefreshToken)
	if err != nil {
		span.RecordError(err)
		// Mark integration as expired
		integration.Status = models.IntegrationStatusExpired
		im.repository.UpdateIntegration(ctx, integration)
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update integration
	integration.AccessToken = authResult.AccessToken
	if authResult.RefreshToken != "" {
		integration.RefreshToken = authResult.RefreshToken
	}
	integration.ExpiresAt = &authResult.ExpiresAt
	integration.Status = models.IntegrationStatusActive
	integration.UpdatedAt = time.Now()

	if err := im.repository.UpdateIntegration(ctx, integration); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to update integration: %w", err)
	}

	return integration, nil
}

// ValidateIntegration validates an integration's connection
func (im *IntegrationManager) ValidateIntegration(ctx context.Context, integrationID int) error {
	ctx, span := im.tracer.Start(ctx, "integration_manager.validate_integration")
	defer span.End()

	span.SetAttributes(attribute.Int("integration.id", integrationID))

	// Get integration from database
	integration, err := im.repository.GetIntegration(ctx, integrationID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get integration: %w", err)
	}

	// Get platform integration
	platformIntegration, err := im.GetIntegration(integration.Platform)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get platform integration: %w", err)
	}

	// Validate connection
	if err := platformIntegration.ValidateConnection(ctx); err != nil {
		span.RecordError(err)
		// Mark integration as error
		integration.Status = models.IntegrationStatusError
		im.repository.UpdateIntegration(ctx, integration)
		return fmt.Errorf("integration validation failed: %w", err)
	}

	// Update status if it was in error
	if integration.Status == models.IntegrationStatusError {
		integration.Status = models.IntegrationStatusActive
		integration.UpdatedAt = time.Now()
		im.repository.UpdateIntegration(ctx, integration)
	}

	return nil
}

// SyncCampaign synchronizes a campaign with a platform
func (im *IntegrationManager) SyncCampaign(ctx context.Context, campaign *models.Campaign, platform string) (*PlatformCampaign, error) {
	ctx, span := im.tracer.Start(ctx, "integration_manager.sync_campaign")
	defer span.End()

	span.SetAttributes(
		attribute.String("platform", platform),
		attribute.Int("campaign.id", campaign.ID),
		attribute.String("campaign.name", campaign.Name),
	)

	// Get platform integration
	platformIntegration, err := im.GetIntegration(platform)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get integration: %w", err)
	}

	// Get user's integration for this platform
	integration, err := im.repository.GetIntegrationByPlatform(ctx, platform, campaign.CreatedBy)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get user integration: %w", err)
	}

	// Check if token needs refresh
	if integration.ExpiresAt != nil && time.Now().After(*integration.ExpiresAt) {
		_, err := im.RefreshIntegration(ctx, integration.ID)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to refresh integration: %w", err)
		}
	}

	// Create or update campaign on platform
	var platformCampaign *PlatformCampaign
	
	// Check if campaign already exists on platform
	if campaign.Metadata != nil {
		if platformID, exists := campaign.Metadata[platform+"_id"]; exists {
			// Update existing campaign
			platformCampaign, err = platformIntegration.UpdateCampaign(ctx, platformID.(string), campaign)
		}
	}
	
	if platformCampaign == nil {
		// Create new campaign
		platformCampaign, err = platformIntegration.CreateCampaign(ctx, campaign)
	}

	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to sync campaign: %w", err)
	}

	// Update campaign metadata with platform ID
	if campaign.Metadata == nil {
		campaign.Metadata = make(models.JSON)
	}
	campaign.Metadata[platform+"_id"] = platformCampaign.ID
	campaign.Metadata[platform+"_synced_at"] = time.Now()

	span.SetAttributes(
		attribute.String("platform_campaign.id", platformCampaign.ID),
		attribute.String("platform_campaign.status", platformCampaign.Status),
	)

	return platformCampaign, nil
}

// SyncCampaignMetrics synchronizes campaign metrics from a platform
func (im *IntegrationManager) SyncCampaignMetrics(ctx context.Context, campaignID int, platform string, timeRange TimeRange) (*CampaignMetrics, error) {
	ctx, span := im.tracer.Start(ctx, "integration_manager.sync_campaign_metrics")
	defer span.End()

	span.SetAttributes(
		attribute.String("platform", platform),
		attribute.Int("campaign.id", campaignID),
	)

	// Get platform integration
	platformIntegration, err := im.GetIntegration(platform)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get integration: %w", err)
	}

	// Get campaign to find platform ID
	// This would typically come from the campaign repository
	platformCampaignID := fmt.Sprintf("platform_campaign_%d", campaignID) // Placeholder

	// Get metrics from platform
	metrics, err := platformIntegration.GetCampaignMetrics(ctx, platformCampaignID, timeRange)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get campaign metrics: %w", err)
	}

	span.SetAttributes(
		attribute.Int64("metrics.impressions", metrics.Impressions),
		attribute.Int64("metrics.clicks", metrics.Clicks),
		attribute.Float64("metrics.spend", metrics.Spend),
	)

	return metrics, nil
}

// BulkSyncCampaigns synchronizes multiple campaigns across platforms
func (im *IntegrationManager) BulkSyncCampaigns(ctx context.Context, campaigns []*models.Campaign, platforms []string) (map[string][]*PlatformCampaign, error) {
	ctx, span := im.tracer.Start(ctx, "integration_manager.bulk_sync_campaigns")
	defer span.End()

	span.SetAttributes(
		attribute.Int("campaigns.count", len(campaigns)),
		attribute.StringSlice("platforms", platforms),
	)

	results := make(map[string][]*PlatformCampaign)
	
	for _, platform := range platforms {
		var platformCampaigns []*PlatformCampaign
		
		for _, campaign := range campaigns {
			platformCampaign, err := im.SyncCampaign(ctx, campaign, platform)
			if err != nil {
				// Log error but continue with other campaigns
				span.RecordError(err)
				continue
			}
			platformCampaigns = append(platformCampaigns, platformCampaign)
		}
		
		results[platform] = platformCampaigns
	}

	return results, nil
}

// GetIntegrationHealth returns the health status of all integrations
func (im *IntegrationManager) GetIntegrationHealth(ctx context.Context, userID int) (map[string]IntegrationHealth, error) {
	ctx, span := im.tracer.Start(ctx, "integration_manager.get_integration_health")
	defer span.End()

	span.SetAttributes(attribute.Int("user.id", userID))

	integrations, err := im.repository.ListIntegrations(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to list integrations: %w", err)
	}

	health := make(map[string]IntegrationHealth)
	
	for _, integration := range integrations {
		status := IntegrationHealthHealthy
		message := "Integration is working properly"
		
		// Check if token is expired
		if integration.ExpiresAt != nil && time.Now().After(*integration.ExpiresAt) {
			status = IntegrationHealthExpired
			message = "Access token has expired"
		} else if integration.Status == models.IntegrationStatusError {
			status = IntegrationHealthError
			message = "Integration has errors"
		} else if integration.Status == models.IntegrationStatusInactive {
			status = IntegrationHealthInactive
			message = "Integration is inactive"
		}
		
		health[integration.Platform] = IntegrationHealth{
			Platform:    integration.Platform,
			Status:      status,
			Message:     message,
			LastSync:    integration.LastSync,
			ConnectedAt: integration.CreatedAt,
		}
	}

	return health, nil
}

// IntegrationHealth represents the health status of an integration
type IntegrationHealth struct {
	Platform    string                   `json:"platform"`
	Status      IntegrationHealthStatus  `json:"status"`
	Message     string                   `json:"message"`
	LastSync    *time.Time               `json:"last_sync,omitempty"`
	ConnectedAt time.Time                `json:"connected_at"`
}

// IntegrationHealthStatus represents the health status
type IntegrationHealthStatus string

const (
	IntegrationHealthHealthy  IntegrationHealthStatus = "healthy"
	IntegrationHealthExpired  IntegrationHealthStatus = "expired"
	IntegrationHealthError    IntegrationHealthStatus = "error"
	IntegrationHealthInactive IntegrationHealthStatus = "inactive"
)

// DisconnectPlatform disconnects from a platform
func (im *IntegrationManager) DisconnectPlatform(ctx context.Context, platform string, userID int) error {
	ctx, span := im.tracer.Start(ctx, "integration_manager.disconnect_platform")
	defer span.End()

	span.SetAttributes(
		attribute.String("platform", platform),
		attribute.Int("user.id", userID),
	)

	// Get integration
	integration, err := im.repository.GetIntegrationByPlatform(ctx, platform, userID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get integration: %w", err)
	}

	// Delete integration
	if err := im.repository.DeleteIntegration(ctx, integration.ID); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete integration: %w", err)
	}

	return nil
}

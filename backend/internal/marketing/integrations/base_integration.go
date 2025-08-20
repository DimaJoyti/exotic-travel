package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/exotic-travel-booking/backend/internal/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// PlatformIntegration defines the interface for all marketing platform integrations
type PlatformIntegration interface {
	// Authentication
	Authenticate(ctx context.Context, credentials map[string]string) (*AuthResult, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error)
	ValidateConnection(ctx context.Context) error

	// Campaign Management
	CreateCampaign(ctx context.Context, campaign *models.Campaign) (*PlatformCampaign, error)
	UpdateCampaign(ctx context.Context, platformCampaignID string, campaign *models.Campaign) (*PlatformCampaign, error)
	GetCampaign(ctx context.Context, platformCampaignID string) (*PlatformCampaign, error)
	ListCampaigns(ctx context.Context, filters map[string]interface{}) ([]*PlatformCampaign, error)
	DeleteCampaign(ctx context.Context, platformCampaignID string) error

	// Content Management
	CreateAd(ctx context.Context, content *models.Content, campaignID string) (*PlatformAd, error)
	UpdateAd(ctx context.Context, platformAdID string, content *models.Content) (*PlatformAd, error)
	GetAd(ctx context.Context, platformAdID string) (*PlatformAd, error)

	// Analytics and Reporting
	GetCampaignMetrics(ctx context.Context, campaignID string, timeRange TimeRange) (*CampaignMetrics, error)
	GetAdMetrics(ctx context.Context, adID string, timeRange TimeRange) (*AdMetrics, error)
	GetAccountMetrics(ctx context.Context, timeRange TimeRange) (*AccountMetrics, error)

	// Audience Management
	CreateAudience(ctx context.Context, audience *models.Audience) (*PlatformAudience, error)
	GetAudience(ctx context.Context, audienceID string) (*PlatformAudience, error)
	ListAudiences(ctx context.Context) ([]*PlatformAudience, error)

	// Platform-specific information
	GetPlatformName() string
	GetSupportedFeatures() []string
	GetRateLimits() RateLimits
}

// AuthResult represents authentication result
type AuthResult struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
	Scope        string    `json:"scope,omitempty"`
	AccountID    string    `json:"account_id,omitempty"`
	AccountName  string    `json:"account_name,omitempty"`
}

// PlatformCampaign represents a campaign on the external platform
type PlatformCampaign struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Status           string                 `json:"status"`
	Objective        string                 `json:"objective"`
	Budget           float64                `json:"budget"`
	BudgetType       string                 `json:"budget_type"` // daily, lifetime
	StartDate        time.Time              `json:"start_date"`
	EndDate          *time.Time             `json:"end_date,omitempty"`
	TargetAudience   map[string]interface{} `json:"target_audience"`
	PlatformSpecific map[string]interface{} `json:"platform_specific"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// PlatformAd represents an ad on the external platform
type PlatformAd struct {
	ID               string                 `json:"id"`
	CampaignID       string                 `json:"campaign_id"`
	Name             string                 `json:"name"`
	Status           string                 `json:"status"`
	AdType           string                 `json:"ad_type"`
	Creative         AdCreative             `json:"creative"`
	TargetAudience   map[string]interface{} `json:"target_audience"`
	PlatformSpecific map[string]interface{} `json:"platform_specific"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// AdCreative represents the creative content of an ad
type AdCreative struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	ImageURL    string   `json:"image_url,omitempty"`
	VideoURL    string   `json:"video_url,omitempty"`
	CallToAction string  `json:"call_to_action"`
	Headlines   []string `json:"headlines,omitempty"`
	Descriptions []string `json:"descriptions,omitempty"`
}

// PlatformAudience represents an audience segment on the external platform
type PlatformAudience struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	Size             int64                  `json:"size"`
	Type             string                 `json:"type"` // custom, lookalike, saved
	Demographics     map[string]interface{} `json:"demographics"`
	Interests        []string               `json:"interests"`
	Behaviors        []string               `json:"behaviors"`
	PlatformSpecific map[string]interface{} `json:"platform_specific"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// TimeRange represents a time range for metrics
type TimeRange struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Preset    string    `json:"preset,omitempty"` // today, yesterday, last_7_days, etc.
}

// CampaignMetrics represents campaign performance metrics
type CampaignMetrics struct {
	CampaignID   string            `json:"campaign_id"`
	Impressions  int64             `json:"impressions"`
	Clicks       int64             `json:"clicks"`
	Conversions  int64             `json:"conversions"`
	Spend        float64           `json:"spend"`
	Revenue      float64           `json:"revenue"`
	CTR          float64           `json:"ctr"`
	CPC          float64           `json:"cpc"`
	CPM          float64           `json:"cpm"`
	ROAS         float64           `json:"roas"`
	Frequency    float64           `json:"frequency"`
	Reach        int64             `json:"reach"`
	CustomMetrics map[string]float64 `json:"custom_metrics,omitempty"`
	TimeRange    TimeRange         `json:"time_range"`
}

// AdMetrics represents ad performance metrics
type AdMetrics struct {
	AdID         string            `json:"ad_id"`
	CampaignID   string            `json:"campaign_id"`
	Impressions  int64             `json:"impressions"`
	Clicks       int64             `json:"clicks"`
	Conversions  int64             `json:"conversions"`
	Spend        float64           `json:"spend"`
	Revenue      float64           `json:"revenue"`
	CTR          float64           `json:"ctr"`
	CPC          float64           `json:"cpc"`
	CPM          float64           `json:"cpm"`
	ROAS         float64           `json:"roas"`
	CustomMetrics map[string]float64 `json:"custom_metrics,omitempty"`
	TimeRange    TimeRange         `json:"time_range"`
}

// AccountMetrics represents account-level metrics
type AccountMetrics struct {
	AccountID     string            `json:"account_id"`
	TotalSpend    float64           `json:"total_spend"`
	TotalRevenue  float64           `json:"total_revenue"`
	TotalClicks   int64             `json:"total_clicks"`
	TotalImpressions int64          `json:"total_impressions"`
	TotalConversions int64          `json:"total_conversions"`
	AverageCTR    float64           `json:"average_ctr"`
	AverageCPC    float64           `json:"average_cpc"`
	AverageROAS   float64           `json:"average_roas"`
	CustomMetrics map[string]float64 `json:"custom_metrics,omitempty"`
	TimeRange     TimeRange         `json:"time_range"`
}

// RateLimits represents platform rate limiting information
type RateLimits struct {
	RequestsPerMinute int `json:"requests_per_minute"`
	RequestsPerHour   int `json:"requests_per_hour"`
	RequestsPerDay    int `json:"requests_per_day"`
	BurstLimit        int `json:"burst_limit"`
}

// BaseIntegration provides common functionality for all integrations
type BaseIntegration struct {
	platformName string
	config       map[string]interface{}
	tracer       trace.Tracer
}

// NewBaseIntegration creates a new base integration
func NewBaseIntegration(platformName string, config map[string]interface{}) *BaseIntegration {
	return &BaseIntegration{
		platformName: platformName,
		config:       config,
		tracer:       otel.Tracer(fmt.Sprintf("integration.%s", platformName)),
	}
}

// GetPlatformName returns the platform name
func (b *BaseIntegration) GetPlatformName() string {
	return b.platformName
}

// GetConfig returns the integration configuration
func (b *BaseIntegration) GetConfig() map[string]interface{} {
	return b.config
}

// LogOperation logs an integration operation
func (b *BaseIntegration) LogOperation(ctx context.Context, operation string, data map[string]interface{}) {
	ctx, span := b.tracer.Start(ctx, fmt.Sprintf("%s.%s", b.platformName, operation))
	defer span.End()

	span.SetAttributes(
		attribute.String("platform", b.platformName),
		attribute.String("operation", operation),
	)

	// Add custom attributes from data
	for key, value := range data {
		switch v := value.(type) {
		case string:
			span.SetAttributes(attribute.String(key, v))
		case int:
			span.SetAttributes(attribute.Int(key, v))
		case int64:
			span.SetAttributes(attribute.Int64(key, v))
		case float64:
			span.SetAttributes(attribute.Float64(key, v))
		case bool:
			span.SetAttributes(attribute.Bool(key, v))
		}
	}
}

// HandleRateLimit handles rate limiting for API calls
func (b *BaseIntegration) HandleRateLimit(ctx context.Context, rateLimits RateLimits) error {
	// Implementation would include rate limiting logic
	// For now, this is a placeholder
	return nil
}

// ConvertToJSON converts a struct to JSON map
func (b *BaseIntegration) ConvertToJSON(data interface{}) (map[string]interface{}, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return result, nil
}

// IntegrationError represents an integration-specific error
type IntegrationError struct {
	Platform   string `json:"platform"`
	Operation  string `json:"operation"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code,omitempty"`
	RetryAfter int    `json:"retry_after,omitempty"`
}

func (e *IntegrationError) Error() string {
	return fmt.Sprintf("%s integration error in %s: %s (code: %s)", e.Platform, e.Operation, e.Message, e.Code)
}

// NewIntegrationError creates a new integration error
func NewIntegrationError(platform, operation, code, message string) *IntegrationError {
	return &IntegrationError{
		Platform:  platform,
		Operation: operation,
		Code:      code,
		Message:   message,
	}
}

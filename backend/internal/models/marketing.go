package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Campaign represents a marketing campaign
type Campaign struct {
	ID              int                    `json:"id" db:"id"`
	Name            string                 `json:"name" db:"name"`
	Description     string                 `json:"description" db:"description"`
	Type            CampaignType           `json:"type" db:"type"`
	Status          CampaignStatus         `json:"status" db:"status"`
	Budget          float64                `json:"budget" db:"budget"`
	SpentBudget     float64                `json:"spent_budget" db:"spent_budget"`
	StartDate       time.Time              `json:"start_date" db:"start_date"`
	EndDate         *time.Time             `json:"end_date,omitempty" db:"end_date"`
	TargetAudience  JSON                   `json:"target_audience" db:"target_audience"`
	Objectives      JSON                   `json:"objectives" db:"objectives"`
	Platforms       JSON                   `json:"platforms" db:"platforms"`
	CreatedBy       int                    `json:"created_by" db:"created_by"`
	BrandID         int                    `json:"brand_id" db:"brand_id"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`

	// Joined fields
	Brand    *Brand     `json:"brand,omitempty"`
	Creator  *User      `json:"creator,omitempty"`
	Contents []Content  `json:"contents,omitempty"`
	Metrics  *Metrics   `json:"metrics,omitempty"`
}

// Content represents AI-generated marketing content
type Content struct {
	ID              int           `json:"id" db:"id"`
	CampaignID      int           `json:"campaign_id" db:"campaign_id"`
	Type            ContentType   `json:"type" db:"type"`
	Title           string        `json:"title" db:"title"`
	Body            string        `json:"body" db:"body"`
	Platform        string        `json:"platform" db:"platform"`
	BrandVoice      string        `json:"brand_voice" db:"brand_voice"`
	SEOData         JSON          `json:"seo_data" db:"seo_data"`
	Metadata        JSON          `json:"metadata" db:"metadata"`
	Status          ContentStatus `json:"status" db:"status"`
	VariationGroup  *string       `json:"variation_group,omitempty" db:"variation_group"`
	ParentContentID *int          `json:"parent_content_id,omitempty" db:"parent_content_id"`
	CreatedBy       int           `json:"created_by" db:"created_by"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at" db:"updated_at"`

	// Joined fields
	Campaign *Campaign `json:"campaign,omitempty"`
	Creator  *User     `json:"creator,omitempty"`
	Assets   []Asset   `json:"assets,omitempty"`
}

// Brand represents brand identity and guidelines
type Brand struct {
	ID               int       `json:"id" db:"id"`
	Name             string    `json:"name" db:"name"`
	Description      string    `json:"description" db:"description"`
	VoiceGuidelines  JSON      `json:"voice_guidelines" db:"voice_guidelines"`
	VisualIdentity   JSON      `json:"visual_identity" db:"visual_identity"`
	ColorPalette     JSON      `json:"color_palette" db:"color_palette"`
	Typography       JSON      `json:"typography" db:"typography"`
	LogoURL          string    `json:"logo_url" db:"logo_url"`
	BrandAssets      JSON      `json:"brand_assets" db:"brand_assets"`
	CompanyID        int       `json:"company_id" db:"company_id"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// Asset represents marketing assets (images, videos, etc.)
type Asset struct {
	ID          int         `json:"id" db:"id"`
	ContentID   *int        `json:"content_id,omitempty" db:"content_id"`
	BrandID     *int        `json:"brand_id,omitempty" db:"brand_id"`
	Type        AssetType   `json:"type" db:"type"`
	Name        string      `json:"name" db:"name"`
	URL         string      `json:"url" db:"url"`
	Metadata    JSON        `json:"metadata" db:"metadata"`
	UsageRights JSON        `json:"usage_rights" db:"usage_rights"`
	CreatedBy   int         `json:"created_by" db:"created_by"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`

	// Joined fields
	Content *Content `json:"content,omitempty"`
	Brand   *Brand   `json:"brand,omitempty"`
	Creator *User    `json:"creator,omitempty"`
}

// Audience represents target audience segments
type Audience struct {
	ID           int       `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Description  string    `json:"description" db:"description"`
	Demographics JSON      `json:"demographics" db:"demographics"`
	Interests    JSON      `json:"interests" db:"interests"`
	Behaviors    JSON      `json:"behaviors" db:"behaviors"`
	PlatformData JSON      `json:"platform_data" db:"platform_data"`
	Size         int       `json:"size" db:"size"`
	CreatedBy    int       `json:"created_by" db:"created_by"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`

	// Joined fields
	Creator *User `json:"creator,omitempty"`
}

// Metrics represents campaign performance metrics
type Metrics struct {
	ID           int       `json:"id" db:"id"`
	CampaignID   int       `json:"campaign_id" db:"campaign_id"`
	Platform     string    `json:"platform" db:"platform"`
	Impressions  int64     `json:"impressions" db:"impressions"`
	Clicks       int64     `json:"clicks" db:"clicks"`
	Conversions  int64     `json:"conversions" db:"conversions"`
	Spend        float64   `json:"spend" db:"spend"`
	Revenue      float64   `json:"revenue" db:"revenue"`
	CTR          float64   `json:"ctr" db:"ctr"`
	CPC          float64   `json:"cpc" db:"cpc"`
	ROAS         float64   `json:"roas" db:"roas"`
	MetricData   JSON      `json:"metric_data" db:"metric_data"`
	RecordedAt   time.Time `json:"recorded_at" db:"recorded_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`

	// Joined fields
	Campaign *Campaign `json:"campaign,omitempty"`
}

// Integration represents platform integrations
type Integration struct {
	ID           int                 `json:"id" db:"id"`
	Platform     string              `json:"platform" db:"platform"`
	AccountID    string              `json:"account_id" db:"account_id"`
	AccessToken  string              `json:"access_token" db:"access_token"`
	RefreshToken string              `json:"refresh_token" db:"refresh_token"`
	ExpiresAt    *time.Time          `json:"expires_at,omitempty" db:"expires_at"`
	Status       IntegrationStatus   `json:"status" db:"status"`
	Config       JSON                `json:"config" db:"config"`
	LastSync     *time.Time          `json:"last_sync,omitempty" db:"last_sync"`
	CreatedBy    int                 `json:"created_by" db:"created_by"`
	CreatedAt    time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at" db:"updated_at"`

	// Joined fields
	Creator *User `json:"creator,omitempty"`
}

// Enums
type CampaignType string
const (
	CampaignTypeSocial    CampaignType = "social"
	CampaignTypeEmail     CampaignType = "email"
	CampaignTypeDisplay   CampaignType = "display"
	CampaignTypeSearch    CampaignType = "search"
	CampaignTypeVideo     CampaignType = "video"
	CampaignTypeInfluencer CampaignType = "influencer"
)

type CampaignStatus string
const (
	CampaignStatusDraft    CampaignStatus = "draft"
	CampaignStatusActive   CampaignStatus = "active"
	CampaignStatusPaused   CampaignStatus = "paused"
	CampaignStatusCompleted CampaignStatus = "completed"
	CampaignStatusCancelled CampaignStatus = "cancelled"
)

type ContentType string
const (
	ContentTypeAd         ContentType = "ad"
	ContentTypeSocialPost ContentType = "social_post"
	ContentTypeEmail      ContentType = "email"
	ContentTypeBlog       ContentType = "blog"
	ContentTypeLanding    ContentType = "landing"
	ContentTypeVideo      ContentType = "video"
)

type ContentStatus string
const (
	ContentStatusDraft     ContentStatus = "draft"
	ContentStatusReview    ContentStatus = "review"
	ContentStatusApproved  ContentStatus = "approved"
	ContentStatusPublished ContentStatus = "published"
	ContentStatusArchived  ContentStatus = "archived"
)

type AssetType string
const (
	AssetTypeImage AssetType = "image"
	AssetTypeVideo AssetType = "video"
	AssetTypeAudio AssetType = "audio"
	AssetTypeLogo  AssetType = "logo"
	AssetTypeBanner AssetType = "banner"
)

type IntegrationStatus string
const (
	IntegrationStatusActive    IntegrationStatus = "active"
	IntegrationStatusInactive  IntegrationStatus = "inactive"
	IntegrationStatusError     IntegrationStatus = "error"
	IntegrationStatusExpired   IntegrationStatus = "expired"
)

// JSON type for PostgreSQL JSON fields
type JSON map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSON", value)
	}
	
	return json.Unmarshal(bytes, j)
}

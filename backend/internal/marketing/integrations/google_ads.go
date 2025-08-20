package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/exotic-travel-booking/backend/internal/models"
	"go.opentelemetry.io/otel/attribute"
)

// GoogleAdsIntegration implements the PlatformIntegration interface for Google Ads
type GoogleAdsIntegration struct {
	*BaseIntegration
	clientID     string
	clientSecret string
	developerToken string
	customerID   string
	accessToken  string
	refreshToken string
	httpClient   *http.Client
}

// GoogleAdsConfig represents Google Ads integration configuration
type GoogleAdsConfig struct {
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
	DeveloperToken string `json:"developer_token"`
	CustomerID     string `json:"customer_id"`
	AccessToken    string `json:"access_token,omitempty"`
	RefreshToken   string `json:"refresh_token,omitempty"`
}

// NewGoogleAdsIntegration creates a new Google Ads integration
func NewGoogleAdsIntegration(config GoogleAdsConfig) *GoogleAdsIntegration {
	baseConfig := map[string]interface{}{
		"client_id":       config.ClientID,
		"client_secret":   config.ClientSecret,
		"developer_token": config.DeveloperToken,
		"customer_id":     config.CustomerID,
	}

	return &GoogleAdsIntegration{
		BaseIntegration: NewBaseIntegration("google_ads", baseConfig),
		clientID:        config.ClientID,
		clientSecret:    config.ClientSecret,
		developerToken:  config.DeveloperToken,
		customerID:      config.CustomerID,
		accessToken:     config.AccessToken,
		refreshToken:    config.RefreshToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Authenticate implements OAuth2 authentication for Google Ads
func (g *GoogleAdsIntegration) Authenticate(ctx context.Context, credentials map[string]string) (*AuthResult, error) {
	g.LogOperation(ctx, "authenticate", map[string]interface{}{
		"customer_id": g.customerID,
	})

	authCode, ok := credentials["auth_code"]
	if !ok {
		return nil, NewIntegrationError("google_ads", "authenticate", "missing_auth_code", "Authorization code is required")
	}

	// Exchange authorization code for access token
	tokenURL := "https://oauth2.googleapis.com/token"
	data := url.Values{
		"client_id":     {g.clientID},
		"client_secret": {g.clientSecret},
		"code":          {authCode},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {credentials["redirect_uri"]},
	}

	resp, err := g.httpClient.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange auth code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, NewIntegrationError("google_ads", "authenticate", "token_exchange_failed", string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	g.accessToken = tokenResp.AccessToken
	g.refreshToken = tokenResp.RefreshToken

	return &AuthResult{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		TokenType:    tokenResp.TokenType,
		Scope:        tokenResp.Scope,
		AccountID:    g.customerID,
	}, nil
}

// RefreshToken refreshes the access token
func (g *GoogleAdsIntegration) RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
	g.LogOperation(ctx, "refresh_token", nil)

	tokenURL := "https://oauth2.googleapis.com/token"
	data := url.Values{
		"client_id":     {g.clientID},
		"client_secret": {g.clientSecret},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
	}

	resp, err := g.httpClient.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read refresh response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, NewIntegrationError("google_ads", "refresh_token", "refresh_failed", string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse refresh response: %w", err)
	}

	g.accessToken = tokenResp.AccessToken

	return &AuthResult{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: refreshToken, // Keep the same refresh token
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		TokenType:    tokenResp.TokenType,
		Scope:        tokenResp.Scope,
		AccountID:    g.customerID,
	}, nil
}

// ValidateConnection validates the connection to Google Ads
func (g *GoogleAdsIntegration) ValidateConnection(ctx context.Context) error {
	g.LogOperation(ctx, "validate_connection", nil)

	// Test connection by getting customer info
	query := "SELECT customer.id, customer.descriptive_name FROM customer LIMIT 1"
	_, err := g.executeQuery(ctx, query)
	return err
}

// CreateCampaign creates a new campaign in Google Ads
func (g *GoogleAdsIntegration) CreateCampaign(ctx context.Context, campaign *models.Campaign) (*PlatformCampaign, error) {
	g.LogOperation(ctx, "create_campaign", map[string]interface{}{
		"campaign_name": campaign.Name,
		"campaign_type": campaign.Type,
	})

	// Convert campaign to Google Ads format
	googleCampaign := g.convertToGoogleCampaign(campaign)

	// Create campaign using Google Ads API
	mutation := map[string]interface{}{
		"operations": []map[string]interface{}{
			{
				"create": googleCampaign,
			},
		},
	}

	endpoint := fmt.Sprintf("customers/%s/campaigns:mutate", g.customerID)
	result, err := g.makeAPIRequest(ctx, "POST", endpoint, mutation)
	if err != nil {
		return nil, fmt.Errorf("failed to create campaign: %w", err)
	}

	// Parse response and return platform campaign
	return g.parseGoogleCampaign(result)
}

// UpdateCampaign updates an existing campaign in Google Ads
func (g *GoogleAdsIntegration) UpdateCampaign(ctx context.Context, platformCampaignID string, campaign *models.Campaign) (*PlatformCampaign, error) {
	g.LogOperation(ctx, "update_campaign", map[string]interface{}{
		"platform_campaign_id": platformCampaignID,
		"campaign_name":         campaign.Name,
	})

	// Convert campaign to Google Ads format
	googleCampaign := g.convertToGoogleCampaign(campaign)
	googleCampaign["resource_name"] = fmt.Sprintf("customers/%s/campaigns/%s", g.customerID, platformCampaignID)

	mutation := map[string]interface{}{
		"operations": []map[string]interface{}{
			{
				"update":      googleCampaign,
				"update_mask": "name,status,campaign_budget,start_date,end_date",
			},
		},
	}

	endpoint := fmt.Sprintf("customers/%s/campaigns:mutate", g.customerID)
	result, err := g.makeAPIRequest(ctx, "POST", endpoint, mutation)
	if err != nil {
		return nil, fmt.Errorf("failed to update campaign: %w", err)
	}

	return g.parseGoogleCampaign(result)
}

// GetCampaign retrieves a campaign from Google Ads
func (g *GoogleAdsIntegration) GetCampaign(ctx context.Context, platformCampaignID string) (*PlatformCampaign, error) {
	g.LogOperation(ctx, "get_campaign", map[string]interface{}{
		"platform_campaign_id": platformCampaignID,
	})

	query := fmt.Sprintf(`
		SELECT 
			campaign.id,
			campaign.name,
			campaign.status,
			campaign.advertising_channel_type,
			campaign.campaign_budget,
			campaign.start_date,
			campaign.end_date
		FROM campaign 
		WHERE campaign.id = %s`, platformCampaignID)

	result, err := g.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign: %w", err)
	}

	return g.parseGoogleCampaignFromQuery(result)
}

// ListCampaigns lists campaigns from Google Ads
func (g *GoogleAdsIntegration) ListCampaigns(ctx context.Context, filters map[string]interface{}) ([]*PlatformCampaign, error) {
	g.LogOperation(ctx, "list_campaigns", filters)

	query := `
		SELECT 
			campaign.id,
			campaign.name,
			campaign.status,
			campaign.advertising_channel_type,
			campaign.campaign_budget,
			campaign.start_date,
			campaign.end_date
		FROM campaign`

	// Add filters if provided
	if status, ok := filters["status"]; ok {
		query += fmt.Sprintf(" WHERE campaign.status = '%s'", status)
	}

	if limit, ok := filters["limit"]; ok {
		query += fmt.Sprintf(" LIMIT %v", limit)
	}

	result, err := g.executeQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list campaigns: %w", err)
	}

	return g.parseGoogleCampaignsFromQuery(result)
}

// DeleteCampaign deletes a campaign from Google Ads
func (g *GoogleAdsIntegration) DeleteCampaign(ctx context.Context, platformCampaignID string) error {
	g.LogOperation(ctx, "delete_campaign", map[string]interface{}{
		"platform_campaign_id": platformCampaignID,
	})

	// Google Ads doesn't allow deletion, only pausing
	// Update campaign status to PAUSED
	mutation := map[string]interface{}{
		"operations": []map[string]interface{}{
			{
				"update": map[string]interface{}{
					"resource_name": fmt.Sprintf("customers/%s/campaigns/%s", g.customerID, platformCampaignID),
					"status":        "PAUSED",
				},
				"update_mask": "status",
			},
		},
	}

	endpoint := fmt.Sprintf("customers/%s/campaigns:mutate", g.customerID)
	_, err := g.makeAPIRequest(ctx, "POST", endpoint, mutation)
	return err
}

// GetSupportedFeatures returns the features supported by Google Ads integration
func (g *GoogleAdsIntegration) GetSupportedFeatures() []string {
	return []string{
		"campaigns",
		"ads",
		"keywords",
		"audiences",
		"metrics",
		"conversion_tracking",
		"automated_bidding",
		"responsive_ads",
	}
}

// GetRateLimits returns the rate limits for Google Ads API
func (g *GoogleAdsIntegration) GetRateLimits() RateLimits {
	return RateLimits{
		RequestsPerMinute: 1000,
		RequestsPerHour:   10000,
		RequestsPerDay:    100000,
		BurstLimit:        100,
	}
}

// Helper methods

func (g *GoogleAdsIntegration) makeAPIRequest(ctx context.Context, method, endpoint string, data interface{}) (map[string]interface{}, error) {
	baseURL := "https://googleads.googleapis.com/v14/"
	url := baseURL + endpoint

	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request data: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Authorization", "Bearer "+g.accessToken)
	req.Header.Set("developer-token", g.developerToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, NewIntegrationError("google_ads", "api_request", strconv.Itoa(resp.StatusCode), string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}

func (g *GoogleAdsIntegration) executeQuery(ctx context.Context, query string) (map[string]interface{}, error) {
	data := map[string]interface{}{
		"query": query,
	}

	endpoint := fmt.Sprintf("customers/%s/googleAds:search", g.customerID)
	return g.makeAPIRequest(ctx, "POST", endpoint, data)
}

func (g *GoogleAdsIntegration) convertToGoogleCampaign(campaign *models.Campaign) map[string]interface{} {
	googleCampaign := map[string]interface{}{
		"name":   campaign.Name,
		"status": g.convertCampaignStatus(campaign.Status),
	}

	// Convert campaign type
	switch campaign.Type {
	case models.CampaignTypeSearch:
		googleCampaign["advertising_channel_type"] = "SEARCH"
	case models.CampaignTypeDisplay:
		googleCampaign["advertising_channel_type"] = "DISPLAY"
	case models.CampaignTypeVideo:
		googleCampaign["advertising_channel_type"] = "VIDEO"
	default:
		googleCampaign["advertising_channel_type"] = "SEARCH"
	}

	// Set budget
	if campaign.Budget > 0 {
		googleCampaign["campaign_budget"] = fmt.Sprintf("customers/%s/campaignBudgets/%d", g.customerID, int(campaign.Budget))
	}

	// Set dates
	if !campaign.StartDate.IsZero() {
		googleCampaign["start_date"] = campaign.StartDate.Format("2006-01-02")
	}
	if campaign.EndDate != nil && !campaign.EndDate.IsZero() {
		googleCampaign["end_date"] = campaign.EndDate.Format("2006-01-02")
	}

	return googleCampaign
}

func (g *GoogleAdsIntegration) convertCampaignStatus(status models.CampaignStatus) string {
	switch status {
	case models.CampaignStatusActive:
		return "ENABLED"
	case models.CampaignStatusPaused:
		return "PAUSED"
	case models.CampaignStatusCompleted, models.CampaignStatusCancelled:
		return "REMOVED"
	default:
		return "PAUSED"
	}
}

func (g *GoogleAdsIntegration) parseGoogleCampaign(data map[string]interface{}) (*PlatformCampaign, error) {
	// Implementation would parse Google Ads API response
	// This is a simplified version
	return &PlatformCampaign{
		ID:        "mock_campaign_id",
		Name:      "Mock Campaign",
		Status:    "ENABLED",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (g *GoogleAdsIntegration) parseGoogleCampaignFromQuery(data map[string]interface{}) (*PlatformCampaign, error) {
	// Implementation would parse Google Ads query response
	return g.parseGoogleCampaign(data)
}

func (g *GoogleAdsIntegration) parseGoogleCampaignsFromQuery(data map[string]interface{}) ([]*PlatformCampaign, error) {
	// Implementation would parse multiple campaigns from query response
	campaign, err := g.parseGoogleCampaignFromQuery(data)
	if err != nil {
		return nil, err
	}
	return []*PlatformCampaign{campaign}, nil
}

// Placeholder implementations for interface compliance
func (g *GoogleAdsIntegration) CreateAd(ctx context.Context, content *models.Content, campaignID string) (*PlatformAd, error) {
	return &PlatformAd{ID: "mock_ad_id", CampaignID: campaignID, Name: content.Title}, nil
}

func (g *GoogleAdsIntegration) UpdateAd(ctx context.Context, platformAdID string, content *models.Content) (*PlatformAd, error) {
	return &PlatformAd{ID: platformAdID, Name: content.Title}, nil
}

func (g *GoogleAdsIntegration) GetAd(ctx context.Context, platformAdID string) (*PlatformAd, error) {
	return &PlatformAd{ID: platformAdID}, nil
}

func (g *GoogleAdsIntegration) GetCampaignMetrics(ctx context.Context, campaignID string, timeRange TimeRange) (*CampaignMetrics, error) {
	return &CampaignMetrics{CampaignID: campaignID, TimeRange: timeRange}, nil
}

func (g *GoogleAdsIntegration) GetAdMetrics(ctx context.Context, adID string, timeRange TimeRange) (*AdMetrics, error) {
	return &AdMetrics{AdID: adID, TimeRange: timeRange}, nil
}

func (g *GoogleAdsIntegration) GetAccountMetrics(ctx context.Context, timeRange TimeRange) (*AccountMetrics, error) {
	return &AccountMetrics{AccountID: g.customerID, TimeRange: timeRange}, nil
}

func (g *GoogleAdsIntegration) CreateAudience(ctx context.Context, audience *models.Audience) (*PlatformAudience, error) {
	return &PlatformAudience{ID: "mock_audience_id", Name: audience.Name}, nil
}

func (g *GoogleAdsIntegration) GetAudience(ctx context.Context, audienceID string) (*PlatformAudience, error) {
	return &PlatformAudience{ID: audienceID}, nil
}

func (g *GoogleAdsIntegration) ListAudiences(ctx context.Context) ([]*PlatformAudience, error) {
	return []*PlatformAudience{{ID: "mock_audience_id"}}, nil
}

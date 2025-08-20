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
	"time"

	"github.com/exotic-travel-booking/backend/internal/models"
)

// FacebookAdsIntegration implements the PlatformIntegration interface for Facebook/Meta Ads
type FacebookAdsIntegration struct {
	*BaseIntegration
	appID       string
	appSecret   string
	accessToken string
	adAccountID string
	httpClient  *http.Client
}

// FacebookAdsConfig represents Facebook Ads integration configuration
type FacebookAdsConfig struct {
	AppID       string `json:"app_id"`
	AppSecret   string `json:"app_secret"`
	AccessToken string `json:"access_token,omitempty"`
	AdAccountID string `json:"ad_account_id"`
}

// NewFacebookAdsIntegration creates a new Facebook Ads integration
func NewFacebookAdsIntegration(config FacebookAdsConfig) *FacebookAdsIntegration {
	baseConfig := map[string]interface{}{
		"app_id":        config.AppID,
		"app_secret":    config.AppSecret,
		"ad_account_id": config.AdAccountID,
	}

	return &FacebookAdsIntegration{
		BaseIntegration: NewBaseIntegration("facebook_ads", baseConfig),
		appID:           config.AppID,
		appSecret:       config.AppSecret,
		accessToken:     config.AccessToken,
		adAccountID:     config.AdAccountID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Authenticate implements OAuth2 authentication for Facebook Ads
func (f *FacebookAdsIntegration) Authenticate(ctx context.Context, credentials map[string]string) (*AuthResult, error) {
	f.LogOperation(ctx, "authenticate", map[string]interface{}{
		"ad_account_id": f.adAccountID,
	})

	authCode, ok := credentials["auth_code"]
	if !ok {
		return nil, NewIntegrationError("facebook_ads", "authenticate", "missing_auth_code", "Authorization code is required")
	}

	// Exchange authorization code for access token
	tokenURL := "https://graph.facebook.com/v18.0/oauth/access_token"
	params := url.Values{
		"client_id":     {f.appID},
		"client_secret": {f.appSecret},
		"code":          {authCode},
		"redirect_uri":  {credentials["redirect_uri"]},
	}

	resp, err := f.httpClient.Get(tokenURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("failed to exchange auth code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, NewIntegrationError("facebook_ads", "authenticate", "token_exchange_failed", string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in,omitempty"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	f.accessToken = tokenResp.AccessToken

	// Get long-lived token
	longLivedToken, err := f.getLongLivedToken(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get long-lived token: %w", err)
	}

	return &AuthResult{
		AccessToken: longLivedToken.AccessToken,
		ExpiresAt:   time.Now().Add(time.Duration(longLivedToken.ExpiresIn) * time.Second),
		TokenType:   "Bearer",
		AccountID:   f.adAccountID,
	}, nil
}

// RefreshToken refreshes the access token (Facebook uses long-lived tokens)
func (f *FacebookAdsIntegration) RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
	f.LogOperation(ctx, "refresh_token", nil)

	// Facebook doesn't use refresh tokens, but we can extend the long-lived token
	return f.getLongLivedToken(ctx, refreshToken)
}

// ValidateConnection validates the connection to Facebook Ads
func (f *FacebookAdsIntegration) ValidateConnection(ctx context.Context) error {
	f.LogOperation(ctx, "validate_connection", nil)

	// Test connection by getting ad account info
	endpoint := fmt.Sprintf("act_%s", f.adAccountID)
	params := url.Values{
		"fields":       {"id,name,account_status"},
		"access_token": {f.accessToken},
	}

	_, err := f.makeAPIRequest(ctx, "GET", endpoint, params, nil)
	return err
}

// CreateCampaign creates a new campaign in Facebook Ads
func (f *FacebookAdsIntegration) CreateCampaign(ctx context.Context, campaign *models.Campaign) (*PlatformCampaign, error) {
	f.LogOperation(ctx, "create_campaign", map[string]interface{}{
		"campaign_name": campaign.Name,
		"campaign_type": campaign.Type,
	})

	endpoint := fmt.Sprintf("act_%s/campaigns", f.adAccountID)
	
	data := map[string]interface{}{
		"name":      campaign.Name,
		"objective": f.convertCampaignObjective(campaign.Type),
		"status":    f.convertCampaignStatus(campaign.Status),
	}

	// Add budget if specified
	if campaign.Budget > 0 {
		data["daily_budget"] = int(campaign.Budget * 100) // Facebook expects cents
	}

	// Add schedule if specified
	if !campaign.StartDate.IsZero() {
		data["start_time"] = campaign.StartDate.Format(time.RFC3339)
	}
	if campaign.EndDate != nil && !campaign.EndDate.IsZero() {
		data["stop_time"] = campaign.EndDate.Format(time.RFC3339)
	}

	params := url.Values{
		"access_token": {f.accessToken},
	}

	result, err := f.makeAPIRequest(ctx, "POST", endpoint, params, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create campaign: %w", err)
	}

	return f.parseFacebookCampaign(result)
}

// UpdateCampaign updates an existing campaign in Facebook Ads
func (f *FacebookAdsIntegration) UpdateCampaign(ctx context.Context, platformCampaignID string, campaign *models.Campaign) (*PlatformCampaign, error) {
	f.LogOperation(ctx, "update_campaign", map[string]interface{}{
		"platform_campaign_id": platformCampaignID,
		"campaign_name":         campaign.Name,
	})

	endpoint := platformCampaignID
	
	data := map[string]interface{}{
		"name":   campaign.Name,
		"status": f.convertCampaignStatus(campaign.Status),
	}

	// Add budget if specified
	if campaign.Budget > 0 {
		data["daily_budget"] = int(campaign.Budget * 100)
	}

	params := url.Values{
		"access_token": {f.accessToken},
	}

	result, err := f.makeAPIRequest(ctx, "POST", endpoint, params, data)
	if err != nil {
		return nil, fmt.Errorf("failed to update campaign: %w", err)
	}

	return f.parseFacebookCampaign(result)
}

// GetCampaign retrieves a campaign from Facebook Ads
func (f *FacebookAdsIntegration) GetCampaign(ctx context.Context, platformCampaignID string) (*PlatformCampaign, error) {
	f.LogOperation(ctx, "get_campaign", map[string]interface{}{
		"platform_campaign_id": platformCampaignID,
	})

	endpoint := platformCampaignID
	params := url.Values{
		"fields":       {"id,name,objective,status,daily_budget,start_time,stop_time,created_time,updated_time"},
		"access_token": {f.accessToken},
	}

	result, err := f.makeAPIRequest(ctx, "GET", endpoint, params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign: %w", err)
	}

	return f.parseFacebookCampaign(result)
}

// ListCampaigns lists campaigns from Facebook Ads
func (f *FacebookAdsIntegration) ListCampaigns(ctx context.Context, filters map[string]interface{}) ([]*PlatformCampaign, error) {
	f.LogOperation(ctx, "list_campaigns", filters)

	endpoint := fmt.Sprintf("act_%s/campaigns", f.adAccountID)
	params := url.Values{
		"fields":       {"id,name,objective,status,daily_budget,start_time,stop_time,created_time,updated_time"},
		"access_token": {f.accessToken},
	}

	// Add filters
	if limit, ok := filters["limit"]; ok {
		params.Set("limit", fmt.Sprintf("%v", limit))
	}

	if status, ok := filters["status"]; ok {
		params.Set("filtering", fmt.Sprintf(`[{"field":"campaign.status","operator":"EQUAL","value":"%s"}]`, status))
	}

	result, err := f.makeAPIRequest(ctx, "GET", endpoint, params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list campaigns: %w", err)
	}

	return f.parseFacebookCampaigns(result)
}

// DeleteCampaign deletes a campaign from Facebook Ads
func (f *FacebookAdsIntegration) DeleteCampaign(ctx context.Context, platformCampaignID string) error {
	f.LogOperation(ctx, "delete_campaign", map[string]interface{}{
		"platform_campaign_id": platformCampaignID,
	})

	endpoint := platformCampaignID
	data := map[string]interface{}{
		"status": "DELETED",
	}

	params := url.Values{
		"access_token": {f.accessToken},
	}

	_, err := f.makeAPIRequest(ctx, "POST", endpoint, params, data)
	return err
}

// GetCampaignMetrics retrieves campaign metrics from Facebook Ads
func (f *FacebookAdsIntegration) GetCampaignMetrics(ctx context.Context, campaignID string, timeRange TimeRange) (*CampaignMetrics, error) {
	f.LogOperation(ctx, "get_campaign_metrics", map[string]interface{}{
		"campaign_id": campaignID,
		"time_range":  timeRange,
	})

	endpoint := fmt.Sprintf("%s/insights", campaignID)
	params := url.Values{
		"fields": {"impressions,clicks,spend,actions,ctr,cpc,cpm"},
		"time_range": {fmt.Sprintf(`{"since":"%s","until":"%s"}`, 
			timeRange.StartDate.Format("2006-01-02"), 
			timeRange.EndDate.Format("2006-01-02"))},
		"access_token": {f.accessToken},
	}

	result, err := f.makeAPIRequest(ctx, "GET", endpoint, params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign metrics: %w", err)
	}

	return f.parseFacebookMetrics(result, campaignID, timeRange)
}

// GetSupportedFeatures returns the features supported by Facebook Ads integration
func (f *FacebookAdsIntegration) GetSupportedFeatures() []string {
	return []string{
		"campaigns",
		"ad_sets",
		"ads",
		"audiences",
		"metrics",
		"conversion_tracking",
		"lookalike_audiences",
		"custom_audiences",
		"dynamic_ads",
		"video_ads",
	}
}

// GetRateLimits returns the rate limits for Facebook Ads API
func (f *FacebookAdsIntegration) GetRateLimits() RateLimits {
	return RateLimits{
		RequestsPerMinute: 200,
		RequestsPerHour:   4800,
		RequestsPerDay:    115200,
		BurstLimit:        50,
	}
}

// Helper methods

func (f *FacebookAdsIntegration) makeAPIRequest(ctx context.Context, method, endpoint string, params url.Values, data interface{}) (map[string]interface{}, error) {
	baseURL := "https://graph.facebook.com/v18.0/"
	requestURL := baseURL + endpoint

	var body io.Reader
	if data != nil && (method == "POST" || method == "PUT") {
		// For POST/PUT requests, add data to form values
		if params == nil {
			params = url.Values{}
		}
		
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request data: %w", err)
		}
		
		// Convert JSON data to form values
		var formData map[string]interface{}
		if err := json.Unmarshal(jsonData, &formData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal data for form: %w", err)
		}
		
		for key, value := range formData {
			params.Set(key, fmt.Sprintf("%v", value))
		}
		
		body = bytes.NewBufferString(params.Encode())
	} else if params != nil {
		requestURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if method == "POST" || method == "PUT" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, NewIntegrationError("facebook_ads", "api_request", strconv.Itoa(resp.StatusCode), string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}

func (f *FacebookAdsIntegration) getLongLivedToken(ctx context.Context, shortLivedToken string) (*AuthResult, error) {
	tokenURL := "https://graph.facebook.com/v18.0/oauth/access_token"
	params := url.Values{
		"grant_type":        {"fb_exchange_token"},
		"client_id":         {f.appID},
		"client_secret":     {f.appSecret},
		"fb_exchange_token": {shortLivedToken},
	}

	resp, err := f.httpClient.Get(tokenURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("failed to get long-lived token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read long-lived token response: %w", err)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse long-lived token response: %w", err)
	}

	return &AuthResult{
		AccessToken: tokenResp.AccessToken,
		ExpiresAt:   time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		TokenType:   tokenResp.TokenType,
		AccountID:   f.adAccountID,
	}, nil
}

func (f *FacebookAdsIntegration) convertCampaignObjective(campaignType models.CampaignType) string {
	switch campaignType {
	case models.CampaignTypeSocial:
		return "REACH"
	case models.CampaignTypeDisplay:
		return "BRAND_AWARENESS"
	case models.CampaignTypeVideo:
		return "VIDEO_VIEWS"
	default:
		return "TRAFFIC"
	}
}

func (f *FacebookAdsIntegration) convertCampaignStatus(status models.CampaignStatus) string {
	switch status {
	case models.CampaignStatusActive:
		return "ACTIVE"
	case models.CampaignStatusPaused:
		return "PAUSED"
	case models.CampaignStatusCompleted, models.CampaignStatusCancelled:
		return "DELETED"
	default:
		return "PAUSED"
	}
}

func (f *FacebookAdsIntegration) parseFacebookCampaign(data map[string]interface{}) (*PlatformCampaign, error) {
	campaign := &PlatformCampaign{
		PlatformSpecific: make(map[string]interface{}),
	}

	if id, ok := data["id"].(string); ok {
		campaign.ID = id
	}
	if name, ok := data["name"].(string); ok {
		campaign.Name = name
	}
	if status, ok := data["status"].(string); ok {
		campaign.Status = status
	}
	if objective, ok := data["objective"].(string); ok {
		campaign.Objective = objective
	}

	// Parse budget
	if budget, ok := data["daily_budget"].(string); ok {
		if budgetFloat, err := strconv.ParseFloat(budget, 64); err == nil {
			campaign.Budget = budgetFloat / 100 // Convert from cents
			campaign.BudgetType = "daily"
		}
	}

	// Parse dates
	if startTime, ok := data["start_time"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, startTime); err == nil {
			campaign.StartDate = parsed
		}
	}
	if stopTime, ok := data["stop_time"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, stopTime); err == nil {
			campaign.EndDate = &parsed
		}
	}
	if createdTime, ok := data["created_time"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, createdTime); err == nil {
			campaign.CreatedAt = parsed
		}
	}
	if updatedTime, ok := data["updated_time"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, updatedTime); err == nil {
			campaign.UpdatedAt = parsed
		}
	}

	// Store original data
	campaign.PlatformSpecific = data

	return campaign, nil
}

func (f *FacebookAdsIntegration) parseFacebookCampaigns(data map[string]interface{}) ([]*PlatformCampaign, error) {
	var campaigns []*PlatformCampaign

	if dataArray, ok := data["data"].([]interface{}); ok {
		for _, item := range dataArray {
			if campaignData, ok := item.(map[string]interface{}); ok {
				campaign, err := f.parseFacebookCampaign(campaignData)
				if err != nil {
					continue // Skip invalid campaigns
				}
				campaigns = append(campaigns, campaign)
			}
		}
	}

	return campaigns, nil
}

func (f *FacebookAdsIntegration) parseFacebookMetrics(data map[string]interface{}, campaignID string, timeRange TimeRange) (*CampaignMetrics, error) {
	metrics := &CampaignMetrics{
		CampaignID: campaignID,
		TimeRange:  timeRange,
	}

	if dataArray, ok := data["data"].([]interface{}); ok && len(dataArray) > 0 {
		if metricsData, ok := dataArray[0].(map[string]interface{}); ok {
			if impressions, ok := metricsData["impressions"].(string); ok {
				if val, err := strconv.ParseInt(impressions, 10, 64); err == nil {
					metrics.Impressions = val
				}
			}
			if clicks, ok := metricsData["clicks"].(string); ok {
				if val, err := strconv.ParseInt(clicks, 10, 64); err == nil {
					metrics.Clicks = val
				}
			}
			if spend, ok := metricsData["spend"].(string); ok {
				if val, err := strconv.ParseFloat(spend, 64); err == nil {
					metrics.Spend = val
				}
			}
			if ctr, ok := metricsData["ctr"].(string); ok {
				if val, err := strconv.ParseFloat(ctr, 64); err == nil {
					metrics.CTR = val
				}
			}
			if cpc, ok := metricsData["cpc"].(string); ok {
				if val, err := strconv.ParseFloat(cpc, 64); err == nil {
					metrics.CPC = val
				}
			}
			if cpm, ok := metricsData["cpm"].(string); ok {
				if val, err := strconv.ParseFloat(cpm, 64); err == nil {
					metrics.CPM = val
				}
			}
		}
	}

	return metrics, nil
}

// Placeholder implementations for interface compliance
func (f *FacebookAdsIntegration) CreateAd(ctx context.Context, content *models.Content, campaignID string) (*PlatformAd, error) {
	return &PlatformAd{ID: "mock_ad_id", CampaignID: campaignID, Name: content.Title}, nil
}

func (f *FacebookAdsIntegration) UpdateAd(ctx context.Context, platformAdID string, content *models.Content) (*PlatformAd, error) {
	return &PlatformAd{ID: platformAdID, Name: content.Title}, nil
}

func (f *FacebookAdsIntegration) GetAd(ctx context.Context, platformAdID string) (*PlatformAd, error) {
	return &PlatformAd{ID: platformAdID}, nil
}

func (f *FacebookAdsIntegration) GetAdMetrics(ctx context.Context, adID string, timeRange TimeRange) (*AdMetrics, error) {
	return &AdMetrics{AdID: adID, TimeRange: timeRange}, nil
}

func (f *FacebookAdsIntegration) GetAccountMetrics(ctx context.Context, timeRange TimeRange) (*AccountMetrics, error) {
	return &AccountMetrics{AccountID: f.adAccountID, TimeRange: timeRange}, nil
}

func (f *FacebookAdsIntegration) CreateAudience(ctx context.Context, audience *models.Audience) (*PlatformAudience, error) {
	return &PlatformAudience{ID: "mock_audience_id", Name: audience.Name}, nil
}

func (f *FacebookAdsIntegration) GetAudience(ctx context.Context, audienceID string) (*PlatformAudience, error) {
	return &PlatformAudience{ID: audienceID}, nil
}

func (f *FacebookAdsIntegration) ListAudiences(ctx context.Context) ([]*PlatformAudience, error) {
	return []*PlatformAudience{{ID: "mock_audience_id"}}, nil
}

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
)

// LocationTool implements location services functionality
type LocationTool struct {
	*BaseTool
	client *http.Client
}

// LocationRequest represents a location services request
type LocationRequest struct {
	Query     string  `json:"query,omitempty"`     // Search query (address, place name, etc.)
	Latitude  float64 `json:"latitude,omitempty"`  // For reverse geocoding
	Longitude float64 `json:"longitude,omitempty"` // For reverse geocoding
	Radius    int     `json:"radius,omitempty"`    // Search radius in meters
	Type      string  `json:"type,omitempty"`      // Place type filter
	Language  string  `json:"language,omitempty"`  // Language for results
	Limit     int     `json:"limit,omitempty"`     // Maximum number of results
}

// LocationResponse represents a location services response
type LocationResponse struct {
	Places []Place         `json:"places"`
	Query  LocationRequest `json:"query"`
}

// Place represents a location/place
type Place struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	DisplayName      string        `json:"display_name"`
	Address          PlaceAddress  `json:"address"`
	Location         PlaceLocation `json:"location"`
	Types            []string      `json:"types"`
	Rating           float64       `json:"rating,omitempty"`
	UserRatingsTotal int           `json:"user_ratings_total,omitempty"`
	PriceLevel       int           `json:"price_level,omitempty"`
	Photos           []PlacePhoto  `json:"photos,omitempty"`
	OpeningHours     *OpeningHours `json:"opening_hours,omitempty"`
	Website          string        `json:"website,omitempty"`
	PhoneNumber      string        `json:"phone_number,omitempty"`
	BusinessStatus   string        `json:"business_status,omitempty"`
	Vicinity         string        `json:"vicinity,omitempty"`
}

// PlaceAddress represents a place address
type PlaceAddress struct {
	FormattedAddress string `json:"formatted_address"`
	StreetNumber     string `json:"street_number,omitempty"`
	Route            string `json:"route,omitempty"`
	Locality         string `json:"locality,omitempty"`
	AdminArea1       string `json:"administrative_area_level_1,omitempty"`
	AdminArea2       string `json:"administrative_area_level_2,omitempty"`
	Country          string `json:"country,omitempty"`
	CountryCode      string `json:"country_code,omitempty"`
	PostalCode       string `json:"postal_code,omitempty"`
}

// PlaceLocation represents geographic coordinates
type PlaceLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// PlacePhoto represents a place photo
type PlacePhoto struct {
	PhotoReference   string   `json:"photo_reference"`
	Height           int      `json:"height"`
	Width            int      `json:"width"`
	HTMLAttributions []string `json:"html_attributions,omitempty"`
}

// OpeningHours represents opening hours information
type OpeningHours struct {
	OpenNow     bool     `json:"open_now"`
	Periods     []Period `json:"periods,omitempty"`
	WeekdayText []string `json:"weekday_text,omitempty"`
}

// Period represents an opening/closing period
type Period struct {
	Open  TimeOfDay `json:"open"`
	Close TimeOfDay `json:"close,omitempty"`
}

// TimeOfDay represents a time of day
type TimeOfDay struct {
	Day  int    `json:"day"`  // 0=Sunday, 1=Monday, etc.
	Time string `json:"time"` // HHMM format
}

// NewLocationTool creates a new location tool
func NewLocationTool(config *ToolConfig) *LocationTool {
	if config.BaseURL == "" {
		config.BaseURL = "https://maps.googleapis.com/maps/api" // Default to Google Places API
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	return &LocationTool{
		BaseTool: NewBaseTool(config),
		client:   client,
	}
}

// Execute executes the location tool
func (t *LocationTool) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	ctx, span := t.tracer.Start(ctx, "location_tool.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("tool.name", t.GetName()),
		attribute.String("tool.type", "location"),
	)

	// Parse input
	req, err := t.parseRequest(input)
	if err != nil {
		span.RecordError(err)
		return nil, NewToolError("invalid_input", err.Error(), t.GetName(), nil)
	}

	// Execute location search with retry
	var response *LocationResponse
	err = t.WithRetry(ctx, func() error {
		var locationErr error
		response, locationErr = t.searchLocations(ctx, req)
		return locationErr
	})

	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Convert response to map
	result := map[string]interface{}{
		"places": response.Places,
		"query":  response.Query,
		"location_metadata": map[string]interface{}{
			"provider":      "google_places", // or detect from config
			"results_count": len(response.Places),
		},
	}

	span.SetAttributes(
		attribute.Int("location.results_count", len(response.Places)),
	)

	if req.Query != "" {
		span.SetAttributes(attribute.String("location.query", req.Query))
	}

	if req.Latitude != 0 && req.Longitude != 0 {
		span.SetAttributes(
			attribute.Float64("location.latitude", req.Latitude),
			attribute.Float64("location.longitude", req.Longitude),
		)
	}

	return result, nil
}

// parseRequest parses the input into a location request
func (t *LocationTool) parseRequest(input map[string]interface{}) (*LocationRequest, error) {
	req := &LocationRequest{}

	// Either query OR coordinates are required
	hasQuery := false
	hasCoordinates := false

	if query, ok := input["query"].(string); ok && query != "" {
		req.Query = query
		hasQuery = true
	}

	if latitude, ok := input["latitude"].(float64); ok {
		req.Latitude = latitude
		hasCoordinates = true
	}

	if longitude, ok := input["longitude"].(float64); ok {
		req.Longitude = longitude
	} else if hasCoordinates {
		hasCoordinates = false // Need both lat and lon
	}

	if !hasQuery && !hasCoordinates {
		return nil, fmt.Errorf("either query or coordinates (latitude and longitude) are required")
	}

	// Optional fields with defaults
	if radius, ok := input["radius"].(float64); ok {
		req.Radius = int(radius)
	} else {
		req.Radius = 5000 // Default 5km radius
	}

	if placeType, ok := input["type"].(string); ok {
		req.Type = placeType
	}

	if language, ok := input["language"].(string); ok {
		req.Language = language
	} else {
		req.Language = "en" // Default to English
	}

	if limit, ok := input["limit"].(float64); ok {
		req.Limit = int(limit)
	} else {
		req.Limit = 20 // Default limit
	}

	if req.Limit > 60 {
		req.Limit = 60 // Google Places API limit
	}

	return req, nil
}

// searchLocations performs the actual location search
func (t *LocationTool) searchLocations(ctx context.Context, req *LocationRequest) (*LocationResponse, error) {
	// For demo purposes, we'll return mock data
	// In production, this would call the actual location API (Google Places, Foursquare, etc.)

	if t.config.APIKey == "" {
		// Return mock data for development
		return t.getMockLocations(req), nil
	}

	// Real API call would go here
	return t.callLocationAPI(ctx, req)
}

// getMockLocations returns mock location data for development
func (t *LocationTool) getMockLocations(req *LocationRequest) *LocationResponse {
	places := []Place{}

	if req.Query != "" {
		// Mock search results based on query
		places = []Place{
			{
				ID:          "place_001",
				Name:        "Central Park",
				DisplayName: "Central Park, New York, NY, USA",
				Address: PlaceAddress{
					FormattedAddress: "Central Park, New York, NY, USA",
					Locality:         "New York",
					AdminArea1:       "New York",
					Country:          "United States",
					CountryCode:      "US",
					PostalCode:       "10024",
				},
				Location: PlaceLocation{
					Latitude:  40.7829,
					Longitude: -73.9654,
				},
				Types:            []string{"park", "tourist_attraction", "establishment"},
				Rating:           4.6,
				UserRatingsTotal: 89234,
				Photos: []PlacePhoto{
					{
						PhotoReference: "mock_photo_ref_1",
						Height:         400,
						Width:          600,
					},
				},
				OpeningHours: &OpeningHours{
					OpenNow: true,
					WeekdayText: []string{
						"Monday: 6:00 AM – 1:00 AM",
						"Tuesday: 6:00 AM – 1:00 AM",
						"Wednesday: 6:00 AM – 1:00 AM",
						"Thursday: 6:00 AM – 1:00 AM",
						"Friday: 6:00 AM – 1:00 AM",
						"Saturday: 6:00 AM – 1:00 AM",
						"Sunday: 6:00 AM – 1:00 AM",
					},
				},
				BusinessStatus: "OPERATIONAL",
				Vicinity:       "Manhattan",
			},
			{
				ID:          "place_002",
				Name:        "Times Square",
				DisplayName: "Times Square, New York, NY, USA",
				Address: PlaceAddress{
					FormattedAddress: "Times Square, New York, NY 10036, USA",
					Locality:         "New York",
					AdminArea1:       "New York",
					Country:          "United States",
					CountryCode:      "US",
					PostalCode:       "10036",
				},
				Location: PlaceLocation{
					Latitude:  40.7580,
					Longitude: -73.9855,
				},
				Types:            []string{"tourist_attraction", "point_of_interest", "establishment"},
				Rating:           4.3,
				UserRatingsTotal: 156789,
				Photos: []PlacePhoto{
					{
						PhotoReference: "mock_photo_ref_2",
						Height:         400,
						Width:          600,
					},
				},
				BusinessStatus: "OPERATIONAL",
				Vicinity:       "Midtown Manhattan",
			},
		}
	} else if req.Latitude != 0 && req.Longitude != 0 {
		// Mock reverse geocoding results
		places = []Place{
			{
				ID:          "place_reverse_001",
				Name:        "Nearby Restaurant",
				DisplayName: "Nearby Restaurant, City, State, Country",
				Address: PlaceAddress{
					FormattedAddress: fmt.Sprintf("123 Main St, City, State, Country"),
					StreetNumber:     "123",
					Route:            "Main St",
					Locality:         "City",
					AdminArea1:       "State",
					Country:          "Country",
					CountryCode:      "US",
					PostalCode:       "12345",
				},
				Location: PlaceLocation{
					Latitude:  req.Latitude + 0.001,
					Longitude: req.Longitude + 0.001,
				},
				Types:            []string{"restaurant", "food", "establishment"},
				Rating:           4.2,
				UserRatingsTotal: 234,
				PriceLevel:       2,
				OpeningHours: &OpeningHours{
					OpenNow: true,
				},
				PhoneNumber:    "+1-555-123-4567",
				Website:        "https://example-restaurant.com",
				BusinessStatus: "OPERATIONAL",
			},
		}
	}

	// Apply type filter if specified
	if req.Type != "" {
		filteredPlaces := make([]Place, 0)
		for _, place := range places {
			for _, placeType := range place.Types {
				if placeType == req.Type {
					filteredPlaces = append(filteredPlaces, place)
					break
				}
			}
		}
		places = filteredPlaces
	}

	// Apply limit
	if len(places) > req.Limit {
		places = places[:req.Limit]
	}

	return &LocationResponse{
		Places: places,
		Query:  *req,
	}
}

// callLocationAPI calls the actual location API
func (t *LocationTool) callLocationAPI(ctx context.Context, req *LocationRequest) (*LocationResponse, error) {
	var endpoint string
	params := url.Values{}
	params.Add("key", t.config.APIKey)

	if req.Query != "" {
		// Text search
		endpoint = fmt.Sprintf("%s/place/textsearch/json", t.config.BaseURL)
		params.Add("query", req.Query)
	} else {
		// Nearby search
		endpoint = fmt.Sprintf("%s/place/nearbysearch/json", t.config.BaseURL)
		params.Add("location", fmt.Sprintf("%f,%f", req.Latitude, req.Longitude))
		params.Add("radius", strconv.Itoa(req.Radius))
	}

	if req.Type != "" {
		params.Add("type", req.Type)
	}

	if req.Language != "" {
		params.Add("language", req.Language)
	}

	// Create request
	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())
	httpReq, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Make request
	resp, err := t.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response (this would need to be adapted to the specific API format)
	var apiResponse struct {
		Results []interface{} `json:"results"`
		Status  string        `json:"status"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if apiResponse.Status != "OK" {
		return nil, fmt.Errorf("API returned status: %s", apiResponse.Status)
	}

	// Convert API response to our format (this would need proper implementation)
	places := make([]Place, 0)
	// ... conversion logic would go here

	return &LocationResponse{
		Places: places,
		Query:  *req,
	}, nil
}

// GetSchema returns the JSON schema for the location tool
func (t *LocationTool) GetSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Search query (place name, address, etc.). Required if coordinates not provided.",
			},
			"latitude": map[string]interface{}{
				"type":        "number",
				"description": "Latitude for location-based search. Required with longitude if query not provided.",
				"minimum":     -90,
				"maximum":     90,
			},
			"longitude": map[string]interface{}{
				"type":        "number",
				"description": "Longitude for location-based search. Required with latitude if query not provided.",
				"minimum":     -180,
				"maximum":     180,
			},
			"radius": map[string]interface{}{
				"type":        "integer",
				"description": "Search radius in meters (for coordinate-based search)",
				"minimum":     1,
				"maximum":     50000,
				"default":     5000,
			},
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Filter by place type",
				"enum": []string{
					"restaurant", "lodging", "tourist_attraction", "museum", "park",
					"shopping_mall", "gas_station", "hospital", "pharmacy", "bank",
					"atm", "airport", "subway_station", "bus_station", "taxi_stand",
				},
			},
			"language": map[string]interface{}{
				"type":        "string",
				"description": "Language for results",
				"enum":        []string{"en", "es", "fr", "de", "it", "pt", "ru", "ja", "zh"},
				"default":     "en",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of results",
				"minimum":     1,
				"maximum":     60,
				"default":     20,
			},
		},
		"anyOf": []map[string]interface{}{
			{"required": []string{"query"}},
			{"required": []string{"latitude", "longitude"}},
		},
	}
}

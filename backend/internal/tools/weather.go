package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// WeatherTool implements weather information functionality
type WeatherTool struct {
	*BaseTool
	client *http.Client
}

// WeatherRequest represents a weather information request
type WeatherRequest struct {
	Location      string `json:"location"`                 // City name, coordinates, or address
	Days          int    `json:"days,omitempty"`           // Number of forecast days (1-10)
	Units         string `json:"units,omitempty"`          // metric, imperial, kelvin
	Language      string `json:"language,omitempty"`       // Language for weather descriptions
	IncludeHourly bool   `json:"include_hourly,omitempty"` // Include hourly forecast
}

// WeatherResponse represents a weather information response
type WeatherResponse struct {
	Location Location        `json:"location"`
	Current  CurrentWeather  `json:"current"`
	Forecast []DailyForecast `json:"forecast,omitempty"`
	Alerts   []WeatherAlert  `json:"alerts,omitempty"`
	Query    WeatherRequest  `json:"query"`
}

// Location represents location information
type Location struct {
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Region    string  `json:"region,omitempty"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone,omitempty"`
	LocalTime string  `json:"local_time,omitempty"`
}

// CurrentWeather represents current weather conditions
type CurrentWeather struct {
	Temperature   float64 `json:"temperature"`
	FeelsLike     float64 `json:"feels_like"`
	Humidity      int     `json:"humidity"`
	Pressure      float64 `json:"pressure"`
	Visibility    float64 `json:"visibility"`
	UVIndex       float64 `json:"uv_index"`
	WindSpeed     float64 `json:"wind_speed"`
	WindDirection int     `json:"wind_direction"`
	WindGust      float64 `json:"wind_gust,omitempty"`
	CloudCover    int     `json:"cloud_cover"`
	Condition     string  `json:"condition"`
	ConditionCode int     `json:"condition_code"`
	Icon          string  `json:"icon,omitempty"`
	LastUpdated   string  `json:"last_updated"`
	Precipitation float64 `json:"precipitation,omitempty"`
}

// DailyForecast represents daily weather forecast
type DailyForecast struct {
	Date          string           `json:"date"`
	MaxTemp       float64          `json:"max_temp"`
	MinTemp       float64          `json:"min_temp"`
	AvgTemp       float64          `json:"avg_temp"`
	Condition     string           `json:"condition"`
	Icon          string           `json:"icon,omitempty"`
	Humidity      int              `json:"humidity"`
	WindSpeed     float64          `json:"wind_speed"`
	Precipitation float64          `json:"precipitation"`
	ChanceOfRain  int              `json:"chance_of_rain"`
	ChanceOfSnow  int              `json:"chance_of_snow"`
	Sunrise       string           `json:"sunrise,omitempty"`
	Sunset        string           `json:"sunset,omitempty"`
	Moonrise      string           `json:"moonrise,omitempty"`
	Moonset       string           `json:"moonset,omitempty"`
	MoonPhase     string           `json:"moon_phase,omitempty"`
	Hourly        []HourlyForecast `json:"hourly,omitempty"`
}

// HourlyForecast represents hourly weather forecast
type HourlyForecast struct {
	Time          string  `json:"time"`
	Temperature   float64 `json:"temperature"`
	FeelsLike     float64 `json:"feels_like"`
	Condition     string  `json:"condition"`
	Icon          string  `json:"icon,omitempty"`
	Humidity      int     `json:"humidity"`
	WindSpeed     float64 `json:"wind_speed"`
	WindDirection int     `json:"wind_direction"`
	Precipitation float64 `json:"precipitation"`
	ChanceOfRain  int     `json:"chance_of_rain"`
	ChanceOfSnow  int     `json:"chance_of_snow"`
	CloudCover    int     `json:"cloud_cover"`
	UVIndex       float64 `json:"uv_index"`
}

// WeatherAlert represents weather alerts/warnings
type WeatherAlert struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Urgency     string `json:"urgency"`
	Areas       string `json:"areas"`
	Category    string `json:"category"`
	Certainty   string `json:"certainty"`
	Event       string `json:"event"`
	Note        string `json:"note,omitempty"`
	Effective   string `json:"effective"`
	Expires     string `json:"expires"`
	Instruction string `json:"instruction,omitempty"`
}

// NewWeatherTool creates a new weather tool
func NewWeatherTool(config *ToolConfig) *WeatherTool {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.weatherapi.com/v1" // Default to WeatherAPI
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	return &WeatherTool{
		BaseTool: NewBaseTool(config),
		client:   client,
	}
}

// Execute executes the weather tool
func (t *WeatherTool) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	ctx, span := t.tracer.Start(ctx, "weather_tool.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("tool.name", t.GetName()),
		attribute.String("tool.type", "weather"),
	)

	// Parse input
	req, err := t.parseRequest(input)
	if err != nil {
		span.RecordError(err)
		return nil, NewToolError("invalid_input", err.Error(), t.GetName(), nil)
	}

	// Execute weather lookup with retry
	var response *WeatherResponse
	err = t.WithRetry(ctx, func() error {
		var weatherErr error
		response, weatherErr = t.getWeather(ctx, req)
		return weatherErr
	})

	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Convert response to map
	result := map[string]interface{}{
		"location": response.Location,
		"current":  response.Current,
		"forecast": response.Forecast,
		"alerts":   response.Alerts,
		"query":    response.Query,
		"weather_metadata": map[string]interface{}{
			"request_time": time.Now().Format(time.RFC3339),
			"provider":     "weatherapi", // or detect from config
		},
	}

	span.SetAttributes(
		attribute.String("weather.location", response.Location.Name),
		attribute.Float64("weather.temperature", response.Current.Temperature),
		attribute.String("weather.condition", response.Current.Condition),
		attribute.Int("weather.forecast_days", len(response.Forecast)),
	)

	return result, nil
}

// parseRequest parses the input into a weather request
func (t *WeatherTool) parseRequest(input map[string]interface{}) (*WeatherRequest, error) {
	req := &WeatherRequest{}

	// Required fields
	if location, ok := input["location"].(string); ok {
		req.Location = location
	} else {
		return nil, fmt.Errorf("location is required")
	}

	// Optional fields with defaults
	if days, ok := input["days"].(float64); ok {
		req.Days = int(days)
	} else {
		req.Days = 3 // Default to 3-day forecast
	}

	if req.Days < 1 {
		req.Days = 1
	} else if req.Days > 10 {
		req.Days = 10
	}

	if units, ok := input["units"].(string); ok {
		req.Units = units
	} else {
		req.Units = "metric" // Default to metric
	}

	if language, ok := input["language"].(string); ok {
		req.Language = language
	} else {
		req.Language = "en" // Default to English
	}

	if includeHourly, ok := input["include_hourly"].(bool); ok {
		req.IncludeHourly = includeHourly
	}

	return req, nil
}

// getWeather performs the actual weather lookup
func (t *WeatherTool) getWeather(ctx context.Context, req *WeatherRequest) (*WeatherResponse, error) {
	// For demo purposes, we'll return mock data
	// In production, this would call the actual weather API (WeatherAPI, OpenWeatherMap, etc.)

	if t.config.APIKey == "" {
		// Return mock data for development
		return t.getMockWeather(req), nil
	}

	// Real API call would go here
	return t.callWeatherAPI(ctx, req)
}

// getMockWeather returns mock weather data for development
func (t *WeatherTool) getMockWeather(req *WeatherRequest) *WeatherResponse {
	now := time.Now()

	// Create mock location
	location := Location{
		Name:      req.Location,
		Country:   "United States",
		Region:    "State",
		Latitude:  40.7128,
		Longitude: -74.0060,
		Timezone:  "America/New_York",
		LocalTime: now.Format("2006-01-02 15:04"),
	}

	// Create mock current weather
	current := CurrentWeather{
		Temperature:   22.5,
		FeelsLike:     25.0,
		Humidity:      65,
		Pressure:      1013.2,
		Visibility:    10.0,
		UVIndex:       5.0,
		WindSpeed:     15.0,
		WindDirection: 180,
		WindGust:      20.0,
		CloudCover:    40,
		Condition:     "Partly cloudy",
		ConditionCode: 1003,
		Icon:          "//cdn.weatherapi.com/weather/64x64/day/116.png",
		LastUpdated:   now.Format("2006-01-02 15:04"),
		Precipitation: 0.0,
	}

	// Adjust for units
	if req.Units == "imperial" {
		current.Temperature = current.Temperature*9/5 + 32
		current.FeelsLike = current.FeelsLike*9/5 + 32
		current.WindSpeed = current.WindSpeed * 0.621371   // km/h to mph
		current.Visibility = current.Visibility * 0.621371 // km to miles
	}

	// Create mock forecast
	forecast := make([]DailyForecast, req.Days)
	for i := 0; i < req.Days; i++ {
		date := now.AddDate(0, 0, i)

		maxTemp := 25.0 + float64(i%3)
		minTemp := 15.0 + float64(i%3)
		avgTemp := (maxTemp + minTemp) / 2

		if req.Units == "imperial" {
			maxTemp = maxTemp*9/5 + 32
			minTemp = minTemp*9/5 + 32
			avgTemp = avgTemp*9/5 + 32
		}

		dailyForecast := DailyForecast{
			Date:          date.Format("2006-01-02"),
			MaxTemp:       maxTemp,
			MinTemp:       minTemp,
			AvgTemp:       avgTemp,
			Condition:     "Partly cloudy",
			Icon:          "//cdn.weatherapi.com/weather/64x64/day/116.png",
			Humidity:      60 + i*5,
			WindSpeed:     12.0 + float64(i),
			Precipitation: float64(i) * 0.5,
			ChanceOfRain:  20 + i*10,
			ChanceOfSnow:  0,
			Sunrise:       "06:30",
			Sunset:        "19:45",
			Moonrise:      "20:15",
			Moonset:       "07:30",
			MoonPhase:     "Waxing Crescent",
		}

		// Add hourly forecast if requested
		if req.IncludeHourly {
			hourly := make([]HourlyForecast, 24)
			for h := 0; h < 24; h++ {
				hourTime := date.Add(time.Duration(h) * time.Hour)
				temp := minTemp + (maxTemp-minTemp)*0.5*(1+0.8*float64(h-12)/12)

				hourly[h] = HourlyForecast{
					Time:          hourTime.Format("2006-01-02 15:04"),
					Temperature:   temp,
					FeelsLike:     temp + 2,
					Condition:     "Partly cloudy",
					Icon:          "//cdn.weatherapi.com/weather/64x64/day/116.png",
					Humidity:      65,
					WindSpeed:     10.0,
					WindDirection: 180,
					Precipitation: 0.0,
					ChanceOfRain:  20,
					ChanceOfSnow:  0,
					CloudCover:    40,
					UVIndex:       float64(h) / 3.0,
				}
			}
			dailyForecast.Hourly = hourly
		}

		forecast[i] = dailyForecast
	}

	// Mock alerts (empty for now)
	alerts := []WeatherAlert{}

	return &WeatherResponse{
		Location: location,
		Current:  current,
		Forecast: forecast,
		Alerts:   alerts,
		Query:    *req,
	}
}

// callWeatherAPI calls the actual weather API
func (t *WeatherTool) callWeatherAPI(ctx context.Context, req *WeatherRequest) (*WeatherResponse, error) {
	// This would implement the actual API call to WeatherAPI, OpenWeatherMap, etc.

	endpoint := fmt.Sprintf("%s/forecast.json", t.config.BaseURL)

	// Build query parameters
	params := url.Values{}
	params.Add("key", t.config.APIKey)
	params.Add("q", req.Location)
	params.Add("days", fmt.Sprintf("%d", req.Days))
	params.Add("aqi", "yes")
	params.Add("alerts", "yes")

	if req.Language != "" {
		params.Add("lang", req.Language)
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
		Location interface{} `json:"location"`
		Current  interface{} `json:"current"`
		Forecast interface{} `json:"forecast"`
		Alerts   interface{} `json:"alerts"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert API response to our format (this would need proper implementation)
	// For now, return mock data
	return t.getMockWeather(req), nil
}

// GetSchema returns the JSON schema for the weather tool
func (t *WeatherTool) GetSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"location": map[string]interface{}{
				"type":        "string",
				"description": "City name, coordinates (lat,lon), or address",
			},
			"days": map[string]interface{}{
				"type":        "integer",
				"description": "Number of forecast days (1-10)",
				"minimum":     1,
				"maximum":     10,
				"default":     3,
			},
			"units": map[string]interface{}{
				"type":        "string",
				"description": "Temperature units",
				"enum":        []string{"metric", "imperial", "kelvin"},
				"default":     "metric",
			},
			"language": map[string]interface{}{
				"type":        "string",
				"description": "Language for weather descriptions",
				"enum":        []string{"en", "es", "fr", "de", "it", "pt", "ru", "ja", "zh"},
				"default":     "en",
			},
			"include_hourly": map[string]interface{}{
				"type":        "boolean",
				"description": "Include hourly forecast data",
				"default":     false,
			},
		},
		"required": []string{"location"},
	}
}

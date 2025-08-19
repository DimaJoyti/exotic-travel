package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// FlightSearchTool implements flight search functionality
type FlightSearchTool struct {
	*BaseTool
	client *http.Client
}

// FlightSearchRequest represents a flight search request
type FlightSearchRequest struct {
	Origin           string `json:"origin"`
	Destination      string `json:"destination"`
	DepartureDate    string `json:"departure_date"`
	ReturnDate       string `json:"return_date,omitempty"`
	Adults           int    `json:"adults"`
	Children         int    `json:"children,omitempty"`
	Infants          int    `json:"infants,omitempty"`
	Class            string `json:"class,omitempty"` // economy, business, first
	MaxPrice         int    `json:"max_price,omitempty"`
	DirectFlights    bool   `json:"direct_flights,omitempty"`
	PreferredAirline string `json:"preferred_airline,omitempty"`
}

// FlightSearchResponse represents a flight search response
type FlightSearchResponse struct {
	Flights []Flight            `json:"flights"`
	Total   int                 `json:"total"`
	Query   FlightSearchRequest `json:"query"`
}

// Flight represents a flight option
type Flight struct {
	ID            string        `json:"id"`
	Airline       string        `json:"airline"`
	FlightNumber  string        `json:"flight_number"`
	Origin        Airport       `json:"origin"`
	Destination   Airport       `json:"destination"`
	DepartureTime time.Time     `json:"departure_time"`
	ArrivalTime   time.Time     `json:"arrival_time"`
	Duration      time.Duration `json:"duration"`
	Stops         int           `json:"stops"`
	Price         Price         `json:"price"`
	Class         string        `json:"class"`
	Aircraft      string        `json:"aircraft,omitempty"`
	Amenities     []string      `json:"amenities,omitempty"`
	Baggage       BaggageInfo   `json:"baggage,omitempty"`
}

// Airport represents airport information
type Airport struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	City     string `json:"city"`
	Country  string `json:"country"`
	Terminal string `json:"terminal,omitempty"`
}

// Price represents pricing information
type Price struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Taxes    float64 `json:"taxes,omitempty"`
	Fees     float64 `json:"fees,omitempty"`
}

// BaggageInfo represents baggage information
type BaggageInfo struct {
	CarryOn     string `json:"carry_on"`
	CheckedBags int    `json:"checked_bags"`
	WeightLimit string `json:"weight_limit,omitempty"`
	ExtraFees   string `json:"extra_fees,omitempty"`
}

// NewFlightSearchTool creates a new flight search tool
func NewFlightSearchTool(config *ToolConfig) *FlightSearchTool {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.amadeus.com" // Default to Amadeus API
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	return &FlightSearchTool{
		BaseTool: NewBaseTool(config),
		client:   client,
	}
}

// Execute executes the flight search
func (t *FlightSearchTool) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	ctx, span := t.tracer.Start(ctx, "flight_search_tool.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("tool.name", t.GetName()),
		attribute.String("tool.type", "flight_search"),
	)

	// Parse input
	req, err := t.parseRequest(input)
	if err != nil {
		span.RecordError(err)
		return nil, NewToolError("invalid_input", err.Error(), t.GetName(), nil)
	}

	// Execute search with retry
	var response *FlightSearchResponse
	err = t.WithRetry(ctx, func() error {
		var searchErr error
		response, searchErr = t.searchFlights(ctx, req)
		return searchErr
	})

	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Convert response to map
	result := map[string]interface{}{
		"flights": response.Flights,
		"total":   response.Total,
		"query":   response.Query,
		"search_metadata": map[string]interface{}{
			"search_time": time.Now().Format(time.RFC3339),
			"provider":    "amadeus", // or detect from config
		},
	}

	span.SetAttributes(
		attribute.Int("flights.count", len(response.Flights)),
		attribute.String("search.origin", req.Origin),
		attribute.String("search.destination", req.Destination),
	)

	return result, nil
}

// parseRequest parses the input into a flight search request
func (t *FlightSearchTool) parseRequest(input map[string]interface{}) (*FlightSearchRequest, error) {
	req := &FlightSearchRequest{}

	// Required fields
	if origin, ok := input["origin"].(string); ok {
		req.Origin = origin
	} else {
		return nil, fmt.Errorf("origin is required")
	}

	if destination, ok := input["destination"].(string); ok {
		req.Destination = destination
	} else {
		return nil, fmt.Errorf("destination is required")
	}

	if departureDate, ok := input["departure_date"].(string); ok {
		req.DepartureDate = departureDate
	} else {
		return nil, fmt.Errorf("departure_date is required")
	}

	// Optional fields with defaults
	if adults, ok := input["adults"].(float64); ok {
		req.Adults = int(adults)
	} else {
		req.Adults = 1 // Default to 1 adult
	}

	if returnDate, ok := input["return_date"].(string); ok {
		req.ReturnDate = returnDate
	}

	if children, ok := input["children"].(float64); ok {
		req.Children = int(children)
	}

	if infants, ok := input["infants"].(float64); ok {
		req.Infants = int(infants)
	}

	if class, ok := input["class"].(string); ok {
		req.Class = class
	} else {
		req.Class = "economy"
	}

	if maxPrice, ok := input["max_price"].(float64); ok {
		req.MaxPrice = int(maxPrice)
	}

	if directFlights, ok := input["direct_flights"].(bool); ok {
		req.DirectFlights = directFlights
	}

	if preferredAirline, ok := input["preferred_airline"].(string); ok {
		req.PreferredAirline = preferredAirline
	}

	return req, nil
}

// searchFlights performs the actual flight search
func (t *FlightSearchTool) searchFlights(ctx context.Context, req *FlightSearchRequest) (*FlightSearchResponse, error) {
	// For demo purposes, we'll return mock data
	// In production, this would call the actual flight API (Amadeus, Skyscanner, etc.)

	if t.config.APIKey == "" {
		// Return mock data for development
		return t.getMockFlights(req), nil
	}

	// Real API call would go here
	return t.callFlightAPI(ctx, req)
}

// getMockFlights returns mock flight data for development
func (t *FlightSearchTool) getMockFlights(req *FlightSearchRequest) *FlightSearchResponse {
	// Parse dates
	departureTime, _ := time.Parse("2006-01-02", req.DepartureDate)

	flights := []Flight{
		{
			ID:           "FL001",
			Airline:      "American Airlines",
			FlightNumber: "AA123",
			Origin: Airport{
				Code:    req.Origin,
				Name:    "Origin Airport",
				City:    "Origin City",
				Country: "US",
			},
			Destination: Airport{
				Code:    req.Destination,
				Name:    "Destination Airport",
				City:    "Destination City",
				Country: "US",
			},
			DepartureTime: departureTime.Add(8 * time.Hour),
			ArrivalTime:   departureTime.Add(11 * time.Hour),
			Duration:      3 * time.Hour,
			Stops:         0,
			Price: Price{
				Amount:   299.99,
				Currency: "USD",
				Taxes:    45.50,
				Fees:     25.00,
			},
			Class:     req.Class,
			Aircraft:  "Boeing 737",
			Amenities: []string{"WiFi", "Entertainment", "Meals"},
			Baggage: BaggageInfo{
				CarryOn:     "1 bag",
				CheckedBags: 1,
				WeightLimit: "50 lbs",
			},
		},
		{
			ID:           "FL002",
			Airline:      "Delta Airlines",
			FlightNumber: "DL456",
			Origin: Airport{
				Code:    req.Origin,
				Name:    "Origin Airport",
				City:    "Origin City",
				Country: "US",
			},
			Destination: Airport{
				Code:    req.Destination,
				Name:    "Destination Airport",
				City:    "Destination City",
				Country: "US",
			},
			DepartureTime: departureTime.Add(14 * time.Hour),
			ArrivalTime:   departureTime.Add(18 * time.Hour),
			Duration:      4 * time.Hour,
			Stops:         1,
			Price: Price{
				Amount:   249.99,
				Currency: "USD",
				Taxes:    38.75,
				Fees:     20.00,
			},
			Class:     req.Class,
			Aircraft:  "Airbus A320",
			Amenities: []string{"WiFi", "Snacks"},
			Baggage: BaggageInfo{
				CarryOn:     "1 bag",
				CheckedBags: 1,
				WeightLimit: "50 lbs",
			},
		},
	}

	// Filter by direct flights if requested
	if req.DirectFlights {
		filteredFlights := make([]Flight, 0)
		for _, flight := range flights {
			if flight.Stops == 0 {
				filteredFlights = append(filteredFlights, flight)
			}
		}
		flights = filteredFlights
	}

	// Filter by max price if specified
	if req.MaxPrice > 0 {
		filteredFlights := make([]Flight, 0)
		for _, flight := range flights {
			if flight.Price.Amount <= float64(req.MaxPrice) {
				filteredFlights = append(filteredFlights, flight)
			}
		}
		flights = filteredFlights
	}

	return &FlightSearchResponse{
		Flights: flights,
		Total:   len(flights),
		Query:   *req,
	}
}

// callFlightAPI calls the actual flight search API
func (t *FlightSearchTool) callFlightAPI(ctx context.Context, req *FlightSearchRequest) (*FlightSearchResponse, error) {
	// This would implement the actual API call to Amadeus, Skyscanner, etc.
	// For now, return an error indicating it's not implemented

	// Example Amadeus API call structure:
	endpoint := fmt.Sprintf("%s/v2/shopping/flight-offers", t.config.BaseURL)

	// Build query parameters
	params := url.Values{}
	params.Add("originLocationCode", req.Origin)
	params.Add("destinationLocationCode", req.Destination)
	params.Add("departureDate", req.DepartureDate)
	params.Add("adults", strconv.Itoa(req.Adults))

	if req.ReturnDate != "" {
		params.Add("returnDate", req.ReturnDate)
	}

	if req.Children > 0 {
		params.Add("children", strconv.Itoa(req.Children))
	}

	if req.Infants > 0 {
		params.Add("infants", strconv.Itoa(req.Infants))
	}

	if req.Class != "" {
		params.Add("travelClass", req.Class)
	}

	if req.DirectFlights {
		params.Add("nonStop", "true")
	}

	// Create request
	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())
	httpReq, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	httpReq.Header.Set("Authorization", "Bearer "+t.config.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

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
		Data []interface{} `json:"data"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert API response to our format (this would need proper implementation)
	flights := make([]Flight, 0)
	// ... conversion logic would go here

	return &FlightSearchResponse{
		Flights: flights,
		Total:   len(flights),
		Query:   *req,
	}, nil
}

// GetSchema returns the JSON schema for the flight search tool
func (t *FlightSearchTool) GetSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"origin": map[string]interface{}{
				"type":        "string",
				"description": "Origin airport code (e.g., 'JFK', 'LAX')",
			},
			"destination": map[string]interface{}{
				"type":        "string",
				"description": "Destination airport code (e.g., 'CDG', 'LHR')",
			},
			"departure_date": map[string]interface{}{
				"type":        "string",
				"description": "Departure date in YYYY-MM-DD format",
				"pattern":     "^\\d{4}-\\d{2}-\\d{2}$",
			},
			"return_date": map[string]interface{}{
				"type":        "string",
				"description": "Return date in YYYY-MM-DD format (optional for one-way)",
				"pattern":     "^\\d{4}-\\d{2}-\\d{2}$",
			},
			"adults": map[string]interface{}{
				"type":        "integer",
				"description": "Number of adult passengers",
				"minimum":     1,
				"maximum":     9,
				"default":     1,
			},
			"children": map[string]interface{}{
				"type":        "integer",
				"description": "Number of child passengers (2-11 years)",
				"minimum":     0,
				"maximum":     9,
				"default":     0,
			},
			"infants": map[string]interface{}{
				"type":        "integer",
				"description": "Number of infant passengers (under 2 years)",
				"minimum":     0,
				"maximum":     9,
				"default":     0,
			},
			"class": map[string]interface{}{
				"type":        "string",
				"description": "Travel class",
				"enum":        []string{"economy", "premium_economy", "business", "first"},
				"default":     "economy",
			},
			"max_price": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum price per person in USD",
				"minimum":     0,
			},
			"direct_flights": map[string]interface{}{
				"type":        "boolean",
				"description": "Only show direct flights",
				"default":     false,
			},
			"preferred_airline": map[string]interface{}{
				"type":        "string",
				"description": "Preferred airline code (e.g., 'AA', 'DL')",
			},
		},
		"required": []string{"origin", "destination", "departure_date"},
	}
}

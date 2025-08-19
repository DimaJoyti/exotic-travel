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

// HotelSearchTool implements hotel search functionality
type HotelSearchTool struct {
	*BaseTool
	client *http.Client
}

// HotelSearchRequest represents a hotel search request
type HotelSearchRequest struct {
	Location     string   `json:"location"`       // City, address, or coordinates
	CheckInDate  string   `json:"check_in_date"`  // YYYY-MM-DD
	CheckOutDate string   `json:"check_out_date"` // YYYY-MM-DD
	Adults       int      `json:"adults"`
	Children     int      `json:"children,omitempty"`
	Rooms        int      `json:"rooms,omitempty"`
	MinPrice     float64  `json:"min_price,omitempty"`
	MaxPrice     float64  `json:"max_price,omitempty"`
	StarRating   int      `json:"star_rating,omitempty"`   // 1-5 stars
	MinRating    float64  `json:"min_rating,omitempty"`    // Guest rating 1-10
	Amenities    []string `json:"amenities,omitempty"`     // WiFi, Pool, Gym, etc.
	PropertyType string   `json:"property_type,omitempty"` // hotel, apartment, resort, etc.
	SortBy       string   `json:"sort_by,omitempty"`       // price, rating, distance
	Currency     string   `json:"currency,omitempty"`
}

// HotelSearchResponse represents a hotel search response
type HotelSearchResponse struct {
	Hotels []Hotel            `json:"hotels"`
	Total  int                `json:"total"`
	Query  HotelSearchRequest `json:"query"`
}

// Hotel represents a hotel option
type Hotel struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Description  string        `json:"description,omitempty"`
	StarRating   int           `json:"star_rating"`
	GuestRating  float64       `json:"guest_rating"`
	ReviewCount  int           `json:"review_count"`
	Address      Address       `json:"address"`
	Location     Coordinates   `json:"location"`
	Images       []string      `json:"images,omitempty"`
	Amenities    []string      `json:"amenities"`
	RoomTypes    []RoomType    `json:"room_types"`
	Price        HotelPrice    `json:"price"`
	Availability bool          `json:"availability"`
	Distance     Distance      `json:"distance,omitempty"`
	Policies     HotelPolicies `json:"policies,omitempty"`
}

// Address represents a hotel address
type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state,omitempty"`
	Country    string `json:"country"`
	PostalCode string `json:"postal_code,omitempty"`
}

// Coordinates represents geographic coordinates
type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// RoomType represents a hotel room type
type RoomType struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	MaxGuests   int        `json:"max_guests"`
	BedType     string     `json:"bed_type,omitempty"`
	Size        string     `json:"size,omitempty"`
	Amenities   []string   `json:"amenities,omitempty"`
	Price       HotelPrice `json:"price"`
	Available   bool       `json:"available"`
}

// HotelPrice represents hotel pricing
type HotelPrice struct {
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	PerNight  float64 `json:"per_night"`
	TotalStay float64 `json:"total_stay"`
	Taxes     float64 `json:"taxes,omitempty"`
	Fees      float64 `json:"fees,omitempty"`
	Discounts float64 `json:"discounts,omitempty"`
}

// Distance represents distance information
type Distance struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
	From  string  `json:"from,omitempty"`
}

// HotelPolicies represents hotel policies
type HotelPolicies struct {
	CheckIn       string `json:"check_in,omitempty"`
	CheckOut      string `json:"check_out,omitempty"`
	Cancellation  string `json:"cancellation,omitempty"`
	PetPolicy     string `json:"pet_policy,omitempty"`
	SmokingPolicy string `json:"smoking_policy,omitempty"`
}

// NewHotelSearchTool creates a new hotel search tool
func NewHotelSearchTool(config *ToolConfig) *HotelSearchTool {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.booking.com" // Default to Booking.com API
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	return &HotelSearchTool{
		BaseTool: NewBaseTool(config),
		client:   client,
	}
}

// Execute executes the hotel search
func (t *HotelSearchTool) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	ctx, span := t.tracer.Start(ctx, "hotel_search_tool.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("tool.name", t.GetName()),
		attribute.String("tool.type", "hotel_search"),
	)

	// Parse input
	req, err := t.parseRequest(input)
	if err != nil {
		span.RecordError(err)
		return nil, NewToolError("invalid_input", err.Error(), t.GetName(), nil)
	}

	// Execute search with retry
	var response *HotelSearchResponse
	err = t.WithRetry(ctx, func() error {
		var searchErr error
		response, searchErr = t.searchHotels(ctx, req)
		return searchErr
	})

	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Convert response to map
	result := map[string]interface{}{
		"hotels": response.Hotels,
		"total":  response.Total,
		"query":  response.Query,
		"search_metadata": map[string]interface{}{
			"search_time": time.Now().Format(time.RFC3339),
			"provider":    "booking.com", // or detect from config
		},
	}

	span.SetAttributes(
		attribute.Int("hotels.count", len(response.Hotels)),
		attribute.String("search.location", req.Location),
		attribute.String("search.check_in", req.CheckInDate),
		attribute.String("search.check_out", req.CheckOutDate),
	)

	return result, nil
}

// parseRequest parses the input into a hotel search request
func (t *HotelSearchTool) parseRequest(input map[string]interface{}) (*HotelSearchRequest, error) {
	req := &HotelSearchRequest{}

	// Required fields
	if location, ok := input["location"].(string); ok {
		req.Location = location
	} else {
		return nil, fmt.Errorf("location is required")
	}

	if checkInDate, ok := input["check_in_date"].(string); ok {
		req.CheckInDate = checkInDate
	} else {
		return nil, fmt.Errorf("check_in_date is required")
	}

	if checkOutDate, ok := input["check_out_date"].(string); ok {
		req.CheckOutDate = checkOutDate
	} else {
		return nil, fmt.Errorf("check_out_date is required")
	}

	// Optional fields with defaults
	if adults, ok := input["adults"].(float64); ok {
		req.Adults = int(adults)
	} else {
		req.Adults = 2 // Default to 2 adults
	}

	if children, ok := input["children"].(float64); ok {
		req.Children = int(children)
	}

	if rooms, ok := input["rooms"].(float64); ok {
		req.Rooms = int(rooms)
	} else {
		req.Rooms = 1 // Default to 1 room
	}

	if minPrice, ok := input["min_price"].(float64); ok {
		req.MinPrice = minPrice
	}

	if maxPrice, ok := input["max_price"].(float64); ok {
		req.MaxPrice = maxPrice
	}

	if starRating, ok := input["star_rating"].(float64); ok {
		req.StarRating = int(starRating)
	}

	if minRating, ok := input["min_rating"].(float64); ok {
		req.MinRating = minRating
	}

	if amenities, ok := input["amenities"].([]interface{}); ok {
		req.Amenities = make([]string, len(amenities))
		for i, amenity := range amenities {
			if amenityStr, ok := amenity.(string); ok {
				req.Amenities[i] = amenityStr
			}
		}
	}

	if propertyType, ok := input["property_type"].(string); ok {
		req.PropertyType = propertyType
	}

	if sortBy, ok := input["sort_by"].(string); ok {
		req.SortBy = sortBy
	} else {
		req.SortBy = "price" // Default sort by price
	}

	if currency, ok := input["currency"].(string); ok {
		req.Currency = currency
	} else {
		req.Currency = "USD" // Default currency
	}

	return req, nil
}

// searchHotels performs the actual hotel search
func (t *HotelSearchTool) searchHotels(ctx context.Context, req *HotelSearchRequest) (*HotelSearchResponse, error) {
	// For demo purposes, we'll return mock data
	// In production, this would call the actual hotel API (Booking.com, Expedia, etc.)

	if t.config.APIKey == "" {
		// Return mock data for development
		return t.getMockHotels(req), nil
	}

	// Real API call would go here
	return t.callHotelAPI(ctx, req)
}

// getMockHotels returns mock hotel data for development
func (t *HotelSearchTool) getMockHotels(req *HotelSearchRequest) *HotelSearchResponse {
	// Parse dates for pricing calculation
	checkIn, _ := time.Parse("2006-01-02", req.CheckInDate)
	checkOut, _ := time.Parse("2006-01-02", req.CheckOutDate)
	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	if nights <= 0 {
		nights = 1
	}

	hotels := []Hotel{
		{
			ID:          "HTL001",
			Name:        "Grand Plaza Hotel",
			Description: "Luxury hotel in the heart of the city with stunning views and world-class amenities.",
			StarRating:  5,
			GuestRating: 9.2,
			ReviewCount: 1247,
			Address: Address{
				Street:     "123 Main Street",
				City:       req.Location,
				Country:    "US",
				PostalCode: "10001",
			},
			Location: Coordinates{
				Latitude:  40.7589,
				Longitude: -73.9851,
			},
			Images: []string{
				"https://example.com/hotel1-1.jpg",
				"https://example.com/hotel1-2.jpg",
			},
			Amenities: []string{"WiFi", "Pool", "Gym", "Spa", "Restaurant", "Room Service", "Concierge"},
			RoomTypes: []RoomType{
				{
					ID:          "ROOM001",
					Name:        "Deluxe King Room",
					Description: "Spacious room with king bed and city view",
					MaxGuests:   2,
					BedType:     "King",
					Size:        "35 sqm",
					Amenities:   []string{"WiFi", "Minibar", "Safe", "Air Conditioning"},
					Price: HotelPrice{
						Amount:    299.99,
						Currency:  req.Currency,
						PerNight:  299.99,
						TotalStay: 299.99 * float64(nights),
						Taxes:     45.00,
						Fees:      25.00,
					},
					Available: true,
				},
			},
			Price: HotelPrice{
				Amount:    299.99,
				Currency:  req.Currency,
				PerNight:  299.99,
				TotalStay: 299.99 * float64(nights),
				Taxes:     45.00,
				Fees:      25.00,
			},
			Availability: true,
			Distance: Distance{
				Value: 0.5,
				Unit:  "km",
				From:  "City Center",
			},
			Policies: HotelPolicies{
				CheckIn:       "3:00 PM",
				CheckOut:      "11:00 AM",
				Cancellation:  "Free cancellation until 24 hours before check-in",
				PetPolicy:     "Pets allowed with additional fee",
				SmokingPolicy: "Non-smoking property",
			},
		},
		{
			ID:          "HTL002",
			Name:        "Budget Inn Express",
			Description: "Comfortable and affordable accommodation with essential amenities.",
			StarRating:  3,
			GuestRating: 7.8,
			ReviewCount: 892,
			Address: Address{
				Street:     "456 Oak Avenue",
				City:       req.Location,
				Country:    "US",
				PostalCode: "10002",
			},
			Location: Coordinates{
				Latitude:  40.7505,
				Longitude: -73.9934,
			},
			Images: []string{
				"https://example.com/hotel2-1.jpg",
			},
			Amenities: []string{"WiFi", "Parking", "24h Reception", "Breakfast"},
			RoomTypes: []RoomType{
				{
					ID:          "ROOM002",
					Name:        "Standard Double Room",
					Description: "Comfortable room with two double beds",
					MaxGuests:   4,
					BedType:     "Double",
					Size:        "25 sqm",
					Amenities:   []string{"WiFi", "TV", "Air Conditioning"},
					Price: HotelPrice{
						Amount:    89.99,
						Currency:  req.Currency,
						PerNight:  89.99,
						TotalStay: 89.99 * float64(nights),
						Taxes:     13.50,
						Fees:      10.00,
					},
					Available: true,
				},
			},
			Price: HotelPrice{
				Amount:    89.99,
				Currency:  req.Currency,
				PerNight:  89.99,
				TotalStay: 89.99 * float64(nights),
				Taxes:     13.50,
				Fees:      10.00,
			},
			Availability: true,
			Distance: Distance{
				Value: 2.1,
				Unit:  "km",
				From:  "City Center",
			},
			Policies: HotelPolicies{
				CheckIn:       "2:00 PM",
				CheckOut:      "12:00 PM",
				Cancellation:  "Free cancellation until 48 hours before check-in",
				PetPolicy:     "No pets allowed",
				SmokingPolicy: "Designated smoking areas",
			},
		},
	}

	// Apply filters
	filteredHotels := make([]Hotel, 0)
	for _, hotel := range hotels {
		// Filter by price range
		if req.MinPrice > 0 && hotel.Price.PerNight < req.MinPrice {
			continue
		}
		if req.MaxPrice > 0 && hotel.Price.PerNight > req.MaxPrice {
			continue
		}

		// Filter by star rating
		if req.StarRating > 0 && hotel.StarRating < req.StarRating {
			continue
		}

		// Filter by guest rating
		if req.MinRating > 0 && hotel.GuestRating < req.MinRating {
			continue
		}

		// Filter by amenities
		if len(req.Amenities) > 0 {
			hasAllAmenities := true
			for _, reqAmenity := range req.Amenities {
				found := false
				for _, hotelAmenity := range hotel.Amenities {
					if hotelAmenity == reqAmenity {
						found = true
						break
					}
				}
				if !found {
					hasAllAmenities = false
					break
				}
			}
			if !hasAllAmenities {
				continue
			}
		}

		filteredHotels = append(filteredHotels, hotel)
	}

	return &HotelSearchResponse{
		Hotels: filteredHotels,
		Total:  len(filteredHotels),
		Query:  *req,
	}
}

// callHotelAPI calls the actual hotel search API
func (t *HotelSearchTool) callHotelAPI(ctx context.Context, req *HotelSearchRequest) (*HotelSearchResponse, error) {
	// This would implement the actual API call to Booking.com, Expedia, etc.
	// For now, return an error indicating it's not implemented

	endpoint := fmt.Sprintf("%s/v1/hotels/search", t.config.BaseURL)

	// Build query parameters
	params := url.Values{}
	params.Add("location", req.Location)
	params.Add("checkin", req.CheckInDate)
	params.Add("checkout", req.CheckOutDate)
	params.Add("adults", strconv.Itoa(req.Adults))
	params.Add("rooms", strconv.Itoa(req.Rooms))

	if req.Children > 0 {
		params.Add("children", strconv.Itoa(req.Children))
	}

	if req.MinPrice > 0 {
		params.Add("min_price", fmt.Sprintf("%.2f", req.MinPrice))
	}

	if req.MaxPrice > 0 {
		params.Add("max_price", fmt.Sprintf("%.2f", req.MaxPrice))
	}

	if req.StarRating > 0 {
		params.Add("stars", strconv.Itoa(req.StarRating))
	}

	if req.Currency != "" {
		params.Add("currency", req.Currency)
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
		Results []interface{} `json:"results"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert API response to our format (this would need proper implementation)
	hotels := make([]Hotel, 0)
	// ... conversion logic would go here

	return &HotelSearchResponse{
		Hotels: hotels,
		Total:  len(hotels),
		Query:  *req,
	}, nil
}

// GetSchema returns the JSON schema for the hotel search tool
func (t *HotelSearchTool) GetSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"location": map[string]interface{}{
				"type":        "string",
				"description": "City, address, or location to search for hotels",
			},
			"check_in_date": map[string]interface{}{
				"type":        "string",
				"description": "Check-in date in YYYY-MM-DD format",
				"pattern":     "^\\d{4}-\\d{2}-\\d{2}$",
			},
			"check_out_date": map[string]interface{}{
				"type":        "string",
				"description": "Check-out date in YYYY-MM-DD format",
				"pattern":     "^\\d{4}-\\d{2}-\\d{2}$",
			},
			"adults": map[string]interface{}{
				"type":        "integer",
				"description": "Number of adult guests",
				"minimum":     1,
				"maximum":     20,
				"default":     2,
			},
			"children": map[string]interface{}{
				"type":        "integer",
				"description": "Number of children",
				"minimum":     0,
				"maximum":     10,
				"default":     0,
			},
			"rooms": map[string]interface{}{
				"type":        "integer",
				"description": "Number of rooms needed",
				"minimum":     1,
				"maximum":     10,
				"default":     1,
			},
			"min_price": map[string]interface{}{
				"type":        "number",
				"description": "Minimum price per night",
				"minimum":     0,
			},
			"max_price": map[string]interface{}{
				"type":        "number",
				"description": "Maximum price per night",
				"minimum":     0,
			},
			"star_rating": map[string]interface{}{
				"type":        "integer",
				"description": "Minimum star rating (1-5)",
				"minimum":     1,
				"maximum":     5,
			},
			"min_rating": map[string]interface{}{
				"type":        "number",
				"description": "Minimum guest rating (1-10)",
				"minimum":     1,
				"maximum":     10,
			},
			"amenities": map[string]interface{}{
				"type":        "array",
				"description": "Required amenities",
				"items": map[string]interface{}{
					"type": "string",
					"enum": []string{"WiFi", "Pool", "Gym", "Spa", "Restaurant", "Parking", "Pet Friendly", "Business Center", "Airport Shuttle"},
				},
			},
			"property_type": map[string]interface{}{
				"type":        "string",
				"description": "Type of property",
				"enum":        []string{"hotel", "apartment", "resort", "hostel", "villa", "guesthouse"},
			},
			"sort_by": map[string]interface{}{
				"type":        "string",
				"description": "Sort results by",
				"enum":        []string{"price", "rating", "distance", "popularity"},
				"default":     "price",
			},
			"currency": map[string]interface{}{
				"type":        "string",
				"description": "Currency for prices",
				"enum":        []string{"USD", "EUR", "GBP", "JPY", "CAD", "AUD"},
				"default":     "USD",
			},
		},
		"required": []string{"location", "check_in_date", "check_out_date"},
	}
}

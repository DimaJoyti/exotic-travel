package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Simple response structures
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

type TravelPlanRequest struct {
	Query       string `json:"query"`
	Destination string `json:"destination"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Travelers   int    `json:"travelers"`
	Budget      string `json:"budget"`
}

type TravelPlanResponse struct {
	Status      string                 `json:"status"`
	Message     string                 `json:"message"`
	Itinerary   map[string]interface{} `json:"itinerary"`
	Suggestions []string               `json:"suggestions"`
}

type FlightSearchResponse struct {
	Status  string                   `json:"status"`
	Message string                   `json:"message"`
	Flights []map[string]interface{} `json:"flights"`
}

type HotelSearchResponse struct {
	Status string                   `json:"status"`
	Hotels []map[string]interface{} `json:"hotels"`
}

type WeatherResponse struct {
	Status  string                 `json:"status"`
	Weather map[string]interface{} `json:"weather"`
}

type LocationSearchResponse struct {
	Status    string                   `json:"status"`
	Locations []map[string]interface{} `json:"locations"`
}

type ToolsResponse struct {
	Status string                   `json:"status"`
	Tools  []map[string]interface{} `json:"tools"`
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Logging middleware
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// Health check handler
func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Travel planning handler
func planTripHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TravelPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Mock response for now
	response := TravelPlanResponse{
		Status:  "success",
		Message: fmt.Sprintf("Generated travel plan for %s from %s to %s", req.Destination, req.StartDate, req.EndDate),
		Itinerary: map[string]interface{}{
			"day_1": map[string]interface{}{
				"activities": []string{"Arrival and check-in", "Local exploration", "Welcome dinner"},
				"accommodation": "Hotel Example",
			},
			"day_2": map[string]interface{}{
				"activities": []string{"City tour", "Museum visit", "Local cuisine"},
				"transportation": "Walking/Metro",
			},
		},
		Suggestions: []string{
			"Book flights early for better prices",
			"Consider travel insurance",
			"Check visa requirements",
			"Pack according to weather forecast",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Flight search handler
func searchFlightsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Mock response
	response := FlightSearchResponse{
		Status:  "success",
		Message: "Found available flights",
		Flights: []map[string]interface{}{
			{
				"airline":       "Example Airlines",
				"flight_number": "EX123",
				"departure":     "2024-03-15T10:00:00Z",
				"arrival":       "2024-03-15T14:00:00Z",
				"price":         "$299",
				"duration":      "4h 0m",
			},
			{
				"airline":       "Demo Airways",
				"flight_number": "DA456",
				"departure":     "2024-03-15T15:30:00Z",
				"arrival":       "2024-03-15T19:30:00Z",
				"price":         "$349",
				"duration":      "4h 0m",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Hotel search handler
func searchHotelsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Mock response
	response := HotelSearchResponse{
		Status: "success",
		Hotels: []map[string]interface{}{
			{
				"name":         "Grand Hotel Example",
				"rating":       4.5,
				"price_per_night": "$150",
				"amenities":    []string{"WiFi", "Pool", "Gym", "Restaurant"},
				"location":     "City Center",
			},
			{
				"name":         "Boutique Inn Demo",
				"rating":       4.2,
				"price_per_night": "$120",
				"amenities":    []string{"WiFi", "Breakfast", "Parking"},
				"location":     "Historic District",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Weather handler
func getWeatherHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Mock response
	response := WeatherResponse{
		Status: "success",
		Weather: map[string]interface{}{
			"current": map[string]interface{}{
				"temperature": "22°C",
				"condition":   "Partly Cloudy",
				"humidity":    "65%",
				"wind":        "10 km/h",
			},
			"forecast": []map[string]interface{}{
				{
					"date":        "2024-03-15",
					"high":        "25°C",
					"low":         "18°C",
					"condition":   "Sunny",
					"rain_chance": "10%",
				},
				{
					"date":        "2024-03-16",
					"high":        "23°C",
					"low":         "16°C",
					"condition":   "Cloudy",
					"rain_chance": "30%",
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Location search handler
func searchLocationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Mock response
	response := LocationSearchResponse{
		Status: "success",
		Locations: []map[string]interface{}{
			{
				"name":        "Eiffel Tower",
				"type":        "landmark",
				"rating":      4.8,
				"description": "Iconic iron lattice tower in Paris",
				"coordinates": map[string]float64{"lat": 48.8584, "lng": 2.2945},
			},
			{
				"name":        "Louvre Museum",
				"type":        "museum",
				"rating":      4.7,
				"description": "World's largest art museum",
				"coordinates": map[string]float64{"lat": 48.8606, "lng": 2.3376},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Tools information handler
func getToolsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Mock response
	response := ToolsResponse{
		Status: "success",
		Tools: []map[string]interface{}{
			{
				"name":        "flight_search",
				"description": "Search for flights between destinations",
				"parameters":  []string{"origin", "destination", "departure_date", "return_date"},
			},
			{
				"name":        "hotel_search",
				"description": "Search for hotels and accommodations",
				"parameters":  []string{"location", "check_in_date", "check_out_date", "guests"},
			},
			{
				"name":        "weather_forecast",
				"description": "Get weather information and forecasts",
				"parameters":  []string{"location", "days"},
			},
			{
				"name":        "location_search",
				"description": "Search for places and points of interest",
				"parameters":  []string{"query", "location", "type"},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Root handler
func rootHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"service":     "Exotic Travel Booking - LLM Backend",
		"version":     "1.0.0",
		"status":      "running",
		"description": "LLM-powered travel booking system with intelligent planning capabilities",
		"endpoints": map[string]string{
			"health":           "/health",
			"plan_trip":        "/api/v1/travel/plan",
			"search_flights":   "/api/v1/travel/flights/search",
			"search_hotels":    "/api/v1/travel/hotels/search",
			"get_weather":      "/api/v1/travel/weather",
			"search_locations": "/api/v1/travel/locations/search",
			"get_tools":        "/api/v1/travel/tools",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Parse command line flags
	port := flag.String("port", "8081", "Port to run the server on")
	host := flag.String("host", "0.0.0.0", "Host to bind the server to")
	flag.Parse()

	// Create HTTP mux
	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/v1/travel/plan", planTripHandler)
	mux.HandleFunc("/api/v1/travel/flights/search", searchFlightsHandler)
	mux.HandleFunc("/api/v1/travel/hotels/search", searchHotelsHandler)
	mux.HandleFunc("/api/v1/travel/weather", getWeatherHandler)
	mux.HandleFunc("/api/v1/travel/locations/search", searchLocationsHandler)
	mux.HandleFunc("/api/v1/travel/tools", getToolsHandler)

	// Apply middleware
	handler := corsMiddleware(loggingMiddleware(mux))

	// Create server
	addr := fmt.Sprintf("%s:%s", *host, *port)
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("🚀 Exotic Travel Booking LLM Backend starting on %s", addr)
	log.Printf("📚 API Documentation available at http://%s", addr)
	log.Printf("❤️  Health check available at http://%s/health", addr)

	// Start server
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

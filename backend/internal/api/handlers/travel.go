package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/exotic-travel-booking/backend/internal/agents"
	"github.com/exotic-travel-booking/backend/internal/llm"
	"github.com/exotic-travel-booking/backend/internal/tools"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TravelHandler handles travel-related API requests
type TravelHandler struct {
	travelAgent  *agents.TravelAgent
	llmManager   *llm.LLMManager
	toolRegistry *tools.ToolRegistry
	tracer       trace.Tracer
}

// NewTravelHandler creates a new travel handler
func NewTravelHandler(llmManager *llm.LLMManager, toolRegistry *tools.ToolRegistry) *TravelHandler {
	// Create adapter for the agents package
	llmManagerAdapter := llm.NewAgentLLMManagerAdapter(llmManager)
	travelAgent := agents.NewTravelAgent(llmManagerAdapter, toolRegistry)

	return &TravelHandler{
		travelAgent:  travelAgent,
		llmManager:   llmManager,
		toolRegistry: toolRegistry,
		tracer:       otel.Tracer("api.handlers.travel"),
	}
}

// PlanTrip handles comprehensive trip planning requests
func (h *TravelHandler) PlanTrip(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "travel_handler.plan_trip")
	defer span.End()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req agents.TravelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Add request metadata
	if req.UserID == "" {
		req.UserID = r.Header.Get("X-User-ID")
	}
	if req.SessionID == "" {
		req.SessionID = r.Header.Get("X-Session-ID")
	}

	span.SetAttributes(
		attribute.String("user.id", req.UserID),
		attribute.String("session.id", req.SessionID),
		attribute.String("request.query", req.Query),
	)

	// Process the travel request
	response, err := h.travelAgent.ProcessRequest(ctx, &req)
	if err != nil {
		span.RecordError(err)
		http.Error(w, fmt.Sprintf("Failed to process travel request: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// SearchFlights handles flight search requests
func (h *TravelHandler) SearchFlights(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "travel_handler.search_flights")
	defer span.End()

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	query := r.URL.Query()

	input := map[string]interface{}{
		"origin":         query.Get("origin"),
		"destination":    query.Get("destination"),
		"departure_date": query.Get("departure_date"),
	}

	// Optional parameters
	if returnDate := query.Get("return_date"); returnDate != "" {
		input["return_date"] = returnDate
	}

	if adults := query.Get("adults"); adults != "" {
		if adultsInt, err := strconv.Atoi(adults); err == nil {
			input["adults"] = float64(adultsInt)
		}
	}

	if children := query.Get("children"); children != "" {
		if childrenInt, err := strconv.Atoi(children); err == nil {
			input["children"] = float64(childrenInt)
		}
	}

	if class := query.Get("class"); class != "" {
		input["class"] = class
	}

	if maxPrice := query.Get("max_price"); maxPrice != "" {
		if maxPriceInt, err := strconv.Atoi(maxPrice); err == nil {
			input["max_price"] = float64(maxPriceInt)
		}
	}

	if directFlights := query.Get("direct_flights"); directFlights == "true" {
		input["direct_flights"] = true
	}

	span.SetAttributes(
		attribute.String("flight.origin", query.Get("origin")),
		attribute.String("flight.destination", query.Get("destination")),
		attribute.String("flight.departure_date", query.Get("departure_date")),
	)

	// Execute flight search
	result, err := h.toolRegistry.ExecuteTool(ctx, "flight_search", input)
	if err != nil {
		span.RecordError(err)
		http.Error(w, fmt.Sprintf("Flight search failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// SearchHotels handles hotel search requests
func (h *TravelHandler) SearchHotels(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "travel_handler.search_hotels")
	defer span.End()

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	query := r.URL.Query()

	input := map[string]interface{}{
		"location":       query.Get("location"),
		"check_in_date":  query.Get("check_in_date"),
		"check_out_date": query.Get("check_out_date"),
	}

	// Optional parameters
	if adults := query.Get("adults"); adults != "" {
		if adultsInt, err := strconv.Atoi(adults); err == nil {
			input["adults"] = float64(adultsInt)
		}
	}

	if children := query.Get("children"); children != "" {
		if childrenInt, err := strconv.Atoi(children); err == nil {
			input["children"] = float64(childrenInt)
		}
	}

	if rooms := query.Get("rooms"); rooms != "" {
		if roomsInt, err := strconv.Atoi(rooms); err == nil {
			input["rooms"] = float64(roomsInt)
		}
	}

	if minPrice := query.Get("min_price"); minPrice != "" {
		if minPriceFloat, err := strconv.ParseFloat(minPrice, 64); err == nil {
			input["min_price"] = minPriceFloat
		}
	}

	if maxPrice := query.Get("max_price"); maxPrice != "" {
		if maxPriceFloat, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			input["max_price"] = maxPriceFloat
		}
	}

	if starRating := query.Get("star_rating"); starRating != "" {
		if starRatingInt, err := strconv.Atoi(starRating); err == nil {
			input["star_rating"] = float64(starRatingInt)
		}
	}

	if sortBy := query.Get("sort_by"); sortBy != "" {
		input["sort_by"] = sortBy
	}

	span.SetAttributes(
		attribute.String("hotel.location", query.Get("location")),
		attribute.String("hotel.check_in_date", query.Get("check_in_date")),
		attribute.String("hotel.check_out_date", query.Get("check_out_date")),
	)

	// Execute hotel search
	result, err := h.toolRegistry.ExecuteTool(ctx, "hotel_search", input)
	if err != nil {
		span.RecordError(err)
		http.Error(w, fmt.Sprintf("Hotel search failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetWeather handles weather information requests
func (h *TravelHandler) GetWeather(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "travel_handler.get_weather")
	defer span.End()

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	query := r.URL.Query()

	input := map[string]interface{}{
		"location": query.Get("location"),
	}

	// Optional parameters
	if days := query.Get("days"); days != "" {
		if daysInt, err := strconv.Atoi(days); err == nil {
			input["days"] = float64(daysInt)
		}
	}

	if units := query.Get("units"); units != "" {
		input["units"] = units
	}

	if language := query.Get("language"); language != "" {
		input["language"] = language
	}

	if includeHourly := query.Get("include_hourly"); includeHourly == "true" {
		input["include_hourly"] = true
	}

	span.SetAttributes(
		attribute.String("weather.location", query.Get("location")),
	)

	// Execute weather lookup
	result, err := h.toolRegistry.ExecuteTool(ctx, "weather", input)
	if err != nil {
		span.RecordError(err)
		http.Error(w, fmt.Sprintf("Weather lookup failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// SearchLocations handles location search requests
func (h *TravelHandler) SearchLocations(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "travel_handler.search_locations")
	defer span.End()

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	query := r.URL.Query()

	input := map[string]interface{}{}

	// Either query or coordinates are required
	if searchQuery := query.Get("query"); searchQuery != "" {
		input["query"] = searchQuery
	}

	if lat := query.Get("latitude"); lat != "" {
		if latFloat, err := strconv.ParseFloat(lat, 64); err == nil {
			input["latitude"] = latFloat
		}
	}

	if lng := query.Get("longitude"); lng != "" {
		if lngFloat, err := strconv.ParseFloat(lng, 64); err == nil {
			input["longitude"] = lngFloat
		}
	}

	// Optional parameters
	if radius := query.Get("radius"); radius != "" {
		if radiusInt, err := strconv.Atoi(radius); err == nil {
			input["radius"] = float64(radiusInt)
		}
	}

	if placeType := query.Get("type"); placeType != "" {
		input["type"] = placeType
	}

	if language := query.Get("language"); language != "" {
		input["language"] = language
	}

	if limit := query.Get("limit"); limit != "" {
		if limitInt, err := strconv.Atoi(limit); err == nil {
			input["limit"] = float64(limitInt)
		}
	}

	span.SetAttributes(
		attribute.String("location.query", query.Get("query")),
	)

	// Execute location search
	result, err := h.toolRegistry.ExecuteTool(ctx, "location", input)
	if err != nil {
		span.RecordError(err)
		http.Error(w, fmt.Sprintf("Location search failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetTools returns information about available tools
func (h *TravelHandler) GetTools(w http.ResponseWriter, r *http.Request) {
	_, span := h.tracer.Start(r.Context(), "travel_handler.get_tools")
	defer span.End()

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all tools information
	toolsInfo := h.toolRegistry.GetAllToolsInfo()

	response := map[string]interface{}{
		"tools":     toolsInfo,
		"count":     len(toolsInfo),
		"timestamp": time.Now().Unix(),
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HealthCheck provides a health check endpoint
func (h *TravelHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"services": map[string]string{
			"llm_manager":   "operational",
			"tool_registry": "operational",
			"travel_agent":  "operational",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/exotic-travel-booking/backend/internal/models"
)

// DestinationServiceInterface defines the interface for destination service
type DestinationServiceInterface interface {
	Create(ctx context.Context, req *models.CreateDestinationRequest) (*models.Destination, error)
	GetByID(ctx context.Context, id int) (*models.Destination, error)
	Update(ctx context.Context, id int, req *models.UpdateDestinationRequest) (*models.Destination, error)
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter *models.DestinationFilter) ([]*models.Destination, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*models.Destination, error)
}

// DestinationHandlers handles destination HTTP requests
type DestinationHandlers struct {
	destinationService DestinationServiceInterface
}

// NewDestinationHandlers creates new destination handlers
func NewDestinationHandlers(destinationService DestinationServiceInterface) *DestinationHandlers {
	return &DestinationHandlers{
		destinationService: destinationService,
	}
}

// Create handles destination creation (admin only)
func (h *DestinationHandlers) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateDestinationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	destination, err := h.destinationService.Create(r.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "must be") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to create destination", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(destination)
}

// GetByID handles getting a destination by ID
func (h *DestinationHandlers) GetByID(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(pathParts[len(pathParts)-1])
	if err != nil {
		http.Error(w, "Invalid destination ID", http.StatusBadRequest)
		return
	}

	destination, err := h.destinationService.GetByID(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Destination not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get destination", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(destination)
}

// Update handles destination updates (admin only)
func (h *DestinationHandlers) Update(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(pathParts[len(pathParts)-1])
	if err != nil {
		http.Error(w, "Invalid destination ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateDestinationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	destination, err := h.destinationService.Update(r.Context(), id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Destination not found", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "must be") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to update destination", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(destination)
}

// Delete handles destination deletion (admin only)
func (h *DestinationHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(pathParts[len(pathParts)-1])
	if err != nil {
		http.Error(w, "Invalid destination ID", http.StatusBadRequest)
		return
	}

	err = h.destinationService.Delete(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Destination not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete destination", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// List handles listing destinations with filtering
func (h *DestinationHandlers) List(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	filter := &models.DestinationFilter{}

	if country := r.URL.Query().Get("country"); country != "" {
		filter.Country = country
	}
	if city := r.URL.Query().Get("city"); city != "" {
		filter.City = city
	}
	if search := r.URL.Query().Get("search"); search != "" {
		filter.Search = search
	}

	if minPriceStr := r.URL.Query().Get("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			filter.MinPrice = minPrice
		}
	}

	if maxPriceStr := r.URL.Query().Get("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			filter.MaxPrice = maxPrice
		}
	}

	if durationStr := r.URL.Query().Get("duration"); durationStr != "" {
		if duration, err := strconv.Atoi(durationStr); err == nil {
			filter.Duration = duration
		}
	}

	if maxGuestsStr := r.URL.Query().Get("max_guests"); maxGuestsStr != "" {
		if maxGuests, err := strconv.Atoi(maxGuestsStr); err == nil {
			filter.MaxGuests = maxGuests
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	destinations, err := h.destinationService.List(r.Context(), filter)
	if err != nil {
		http.Error(w, "Failed to list destinations", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(destinations)
}

// Search handles full-text search of destinations
func (h *DestinationHandlers) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Search query is required", http.StatusBadRequest)
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	destinations, err := h.destinationService.Search(r.Context(), query, limit, offset)
	if err != nil {
		http.Error(w, "Failed to search destinations", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(destinations)
}

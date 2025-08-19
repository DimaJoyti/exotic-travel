package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/exotic-travel-booking/backend/internal/models"
	"github.com/exotic-travel-booking/backend/internal/services"
)

// BookingHandlers handles booking HTTP requests
type BookingHandlers struct {
	bookingService *services.BookingService
}

// NewBookingHandlers creates new booking handlers
func NewBookingHandlers(bookingService *services.BookingService) *BookingHandlers {
	return &BookingHandlers{
		bookingService: bookingService,
	}
}

// Create handles booking creation
func (h *BookingHandlers) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	var req models.CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	booking, err := h.bookingService.Create(r.Context(), userID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "must be") ||
			strings.Contains(err.Error(), "exceeds") || strings.Contains(err.Error(), "cannot be") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Destination not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to create booking", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(booking)
}

// GetByID handles getting a booking by ID
func (h *BookingHandlers) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	userRole, ok := r.Context().Value("userRole").(string)
	if !ok {
		http.Error(w, "User role not found in context", http.StatusInternalServerError)
		return
	}

	// Extract ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(pathParts[len(pathParts)-1])
	if err != nil {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return
	}

	booking, err := h.bookingService.GetByID(r.Context(), id, userID, userRole)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Booking not found", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "access denied") {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to get booking", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(booking)
}

// Update handles booking updates
func (h *BookingHandlers) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	userRole, ok := r.Context().Value("userRole").(string)
	if !ok {
		http.Error(w, "User role not found in context", http.StatusInternalServerError)
		return
	}

	// Extract ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(pathParts[len(pathParts)-1])
	if err != nil {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	booking, err := h.bookingService.Update(r.Context(), id, userID, userRole, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Booking not found", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "access denied") {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "must be") ||
			strings.Contains(err.Error(), "exceeds") || strings.Contains(err.Error(), "invalid") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to update booking", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(booking)
}

// Delete handles booking deletion
func (h *BookingHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	userRole, ok := r.Context().Value("userRole").(string)
	if !ok {
		http.Error(w, "User role not found in context", http.StatusInternalServerError)
		return
	}

	// Extract ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(pathParts[len(pathParts)-1])
	if err != nil {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return
	}

	err = h.bookingService.Delete(r.Context(), id, userID, userRole)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Booking not found", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "access denied") {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}
		if strings.Contains(err.Error(), "cannot delete") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to delete booking", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// List handles listing bookings with filtering
func (h *BookingHandlers) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	userRole, ok := r.Context().Value("userRole").(string)
	if !ok {
		http.Error(w, "User role not found in context", http.StatusInternalServerError)
		return
	}

	// Parse query parameters
	filter := &models.BookingFilter{}

	if destinationIDStr := r.URL.Query().Get("destination_id"); destinationIDStr != "" {
		if destinationID, err := strconv.Atoi(destinationIDStr); err == nil {
			filter.DestinationID = destinationID
		}
	}

	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = status
	}

	if paymentStatus := r.URL.Query().Get("payment_status"); paymentStatus != "" {
		filter.PaymentStatus = paymentStatus
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

	bookings, err := h.bookingService.List(r.Context(), userID, userRole, filter)
	if err != nil {
		http.Error(w, "Failed to list bookings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookings)
}

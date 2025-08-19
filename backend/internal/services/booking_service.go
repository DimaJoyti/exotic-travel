package services

import (
	"context"
	"fmt"
	"time"

	"github.com/exotic-travel-booking/backend/internal/models"
	"github.com/exotic-travel-booking/backend/internal/repositories"
)

// BookingService handles booking operations
type BookingService struct {
	bookingRepo     repositories.BookingRepository
	destinationRepo repositories.DestinationRepository
}

// NewBookingService creates a new booking service
func NewBookingService(bookingRepo repositories.BookingRepository, destinationRepo repositories.DestinationRepository) *BookingService {
	return &BookingService{
		bookingRepo:     bookingRepo,
		destinationRepo: destinationRepo,
	}
}

// Create creates a new booking
func (s *BookingService) Create(ctx context.Context, userID int, req *models.CreateBookingRequest) (*models.Booking, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Get destination to calculate price
	destination, err := s.destinationRepo.GetByID(ctx, req.DestinationID)
	if err != nil {
		return nil, fmt.Errorf("destination not found: %w", err)
	}

	// Validate guest count
	if req.Guests > destination.MaxGuests {
		return nil, fmt.Errorf("number of guests (%d) exceeds maximum allowed (%d)", req.Guests, destination.MaxGuests)
	}

	// Calculate total price
	totalPrice := destination.Price * float64(req.Guests)

	booking := &models.Booking{
		UserID:        userID,
		DestinationID: req.DestinationID,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		Guests:        req.Guests,
		TotalPrice:    totalPrice,
		Status:        models.BookingStatusPending,
		PaymentStatus: models.PaymentStatusPending,
	}

	err = s.bookingRepo.Create(ctx, booking)
	if err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	// Get the full booking with joined data
	return s.bookingRepo.GetByID(ctx, booking.ID)
}

// GetByID retrieves a booking by ID
func (s *BookingService) GetByID(ctx context.Context, id int, userID int, userRole string) (*models.Booking, error) {
	booking, err := s.bookingRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	// Check if user has permission to view this booking
	if userRole != models.RoleAdmin && booking.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}

	return booking, nil
}

// Update updates a booking
func (s *BookingService) Update(ctx context.Context, id int, userID int, userRole string, req *models.UpdateBookingRequest) (*models.Booking, error) {
	// Get existing booking
	booking, err := s.bookingRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("booking not found: %w", err)
	}

	// Check permissions
	if userRole != models.RoleAdmin && booking.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}

	// Update fields if provided
	if req.StartDate != nil {
		booking.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		booking.EndDate = *req.EndDate
	}
	if req.Guests != nil {
		booking.Guests = *req.Guests
	}
	if req.Status != nil {
		booking.Status = *req.Status
	}
	if req.PaymentStatus != nil {
		booking.PaymentStatus = *req.PaymentStatus
	}

	// Validate updated booking
	if err := s.validateBooking(booking); err != nil {
		return nil, err
	}

	// Recalculate price if guests changed
	if req.Guests != nil {
		destination, err := s.destinationRepo.GetByID(ctx, booking.DestinationID)
		if err != nil {
			return nil, fmt.Errorf("destination not found: %w", err)
		}

		if booking.Guests > destination.MaxGuests {
			return nil, fmt.Errorf("number of guests (%d) exceeds maximum allowed (%d)", booking.Guests, destination.MaxGuests)
		}

		booking.TotalPrice = destination.Price * float64(booking.Guests)
	}

	err = s.bookingRepo.Update(ctx, booking)
	if err != nil {
		return nil, fmt.Errorf("failed to update booking: %w", err)
	}

	return s.bookingRepo.GetByID(ctx, booking.ID)
}

// Delete deletes a booking
func (s *BookingService) Delete(ctx context.Context, id int, userID int, userRole string) error {
	// Get existing booking
	booking, err := s.bookingRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("booking not found: %w", err)
	}

	// Check permissions
	if userRole != models.RoleAdmin && booking.UserID != userID {
		return fmt.Errorf("access denied")
	}

	// Only allow deletion of pending bookings
	if booking.Status != models.BookingStatusPending {
		return fmt.Errorf("cannot delete booking with status: %s", booking.Status)
	}

	err = s.bookingRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete booking: %w", err)
	}

	return nil
}

// List retrieves bookings with filtering
func (s *BookingService) List(ctx context.Context, userID int, userRole string, filter *models.BookingFilter) ([]*models.Booking, error) {
	// Set default values
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	// Non-admin users can only see their own bookings
	if userRole != models.RoleAdmin {
		filter.UserID = userID
	}

	bookings, err := s.bookingRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list bookings: %w", err)
	}

	return bookings, nil
}

// GetByUserID retrieves bookings for a specific user
func (s *BookingService) GetByUserID(ctx context.Context, targetUserID int, requestingUserID int, userRole string, limit, offset int) ([]*models.Booking, error) {
	// Check permissions
	if userRole != models.RoleAdmin && targetUserID != requestingUserID {
		return nil, fmt.Errorf("access denied")
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	bookings, err := s.bookingRepo.GetByUserID(ctx, targetUserID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user bookings: %w", err)
	}

	return bookings, nil
}

// validateCreateRequest validates a create booking request
func (s *BookingService) validateCreateRequest(req *models.CreateBookingRequest) error {
	if req.DestinationID <= 0 {
		return fmt.Errorf("destination ID is required")
	}
	if req.Guests <= 0 {
		return fmt.Errorf("number of guests must be greater than 0")
	}
	if req.StartDate.IsZero() {
		return fmt.Errorf("start date is required")
	}
	if req.EndDate.IsZero() {
		return fmt.Errorf("end date is required")
	}
	if req.EndDate.Before(req.StartDate) || req.EndDate.Equal(req.StartDate) {
		return fmt.Errorf("end date must be after start date")
	}
	if req.StartDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return fmt.Errorf("start date cannot be in the past")
	}

	return nil
}

// validateBooking validates a booking
func (s *BookingService) validateBooking(booking *models.Booking) error {
	if booking.Guests <= 0 {
		return fmt.Errorf("number of guests must be greater than 0")
	}
	if booking.StartDate.IsZero() {
		return fmt.Errorf("start date is required")
	}
	if booking.EndDate.IsZero() {
		return fmt.Errorf("end date is required")
	}
	if booking.EndDate.Before(booking.StartDate) || booking.EndDate.Equal(booking.StartDate) {
		return fmt.Errorf("end date must be after start date")
	}

	// Validate status
	validStatuses := []string{models.BookingStatusPending, models.BookingStatusConfirmed, models.BookingStatusCancelled, models.BookingStatusCompleted}
	validStatus := false
	for _, status := range validStatuses {
		if booking.Status == status {
			validStatus = true
			break
		}
	}
	if !validStatus {
		return fmt.Errorf("invalid booking status: %s", booking.Status)
	}

	// Validate payment status
	validPaymentStatuses := []string{models.PaymentStatusPending, models.PaymentStatusPaid, models.PaymentStatusFailed, models.PaymentStatusRefunded}
	validPaymentStatus := false
	for _, status := range validPaymentStatuses {
		if booking.PaymentStatus == status {
			validPaymentStatus = true
			break
		}
	}
	if !validPaymentStatus {
		return fmt.Errorf("invalid payment status: %s", booking.PaymentStatus)
	}

	return nil
}

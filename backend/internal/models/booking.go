package models

import (
	"time"
)

// Booking represents a travel booking
type Booking struct {
	ID            int       `json:"id" db:"id"`
	UserID        int       `json:"user_id" db:"user_id"`
	DestinationID int       `json:"destination_id" db:"destination_id"`
	StartDate     time.Time `json:"start_date" db:"start_date"`
	EndDate       time.Time `json:"end_date" db:"end_date"`
	Guests        int       `json:"guests" db:"guests"`
	TotalPrice    float64   `json:"total_price" db:"total_price"`
	Status        string    `json:"status" db:"status"`
	PaymentStatus string    `json:"payment_status" db:"payment_status"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`

	// Joined fields
	User        *User        `json:"user,omitempty"`
	Destination *Destination `json:"destination,omitempty"`
}

// Booking status constants
const (
	BookingStatusPending   = "pending"
	BookingStatusConfirmed = "confirmed"
	BookingStatusCancelled = "cancelled"
	BookingStatusCompleted = "completed"
)

// Payment status constants
const (
	PaymentStatusPending  = "pending"
	PaymentStatusPaid     = "paid"
	PaymentStatusFailed   = "failed"
	PaymentStatusRefunded = "refunded"
)

// CreateBookingRequest represents the request to create a new booking
type CreateBookingRequest struct {
	DestinationID int       `json:"destination_id" validate:"required"`
	StartDate     time.Time `json:"start_date" validate:"required"`
	EndDate       time.Time `json:"end_date" validate:"required"`
	Guests        int       `json:"guests" validate:"required,min=1"`
}

// UpdateBookingRequest represents the request to update a booking
type UpdateBookingRequest struct {
	StartDate     *time.Time `json:"start_date,omitempty"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	Guests        *int       `json:"guests,omitempty"`
	Status        *string    `json:"status,omitempty"`
	PaymentStatus *string    `json:"payment_status,omitempty"`
}

// BookingFilter represents filters for booking queries
type BookingFilter struct {
	UserID        int    `json:"user_id,omitempty"`
	DestinationID int    `json:"destination_id,omitempty"`
	Status        string `json:"status,omitempty"`
	PaymentStatus string `json:"payment_status,omitempty"`
	Limit         int    `json:"limit,omitempty"`
	Offset        int    `json:"offset,omitempty"`
}

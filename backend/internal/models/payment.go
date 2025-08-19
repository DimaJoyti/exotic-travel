package models

import (
	"time"
)

// Payment represents a payment transaction
type Payment struct {
	ID              int       `json:"id" db:"id"`
	BookingID       int       `json:"booking_id" db:"booking_id"`
	StripePaymentID string    `json:"stripe_payment_id" db:"stripe_payment_id"`
	Amount          float64   `json:"amount" db:"amount"`
	Status          string    `json:"status" db:"status"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`

	// Joined fields
	Booking *Booking `json:"booking,omitempty"`
}

// Payment status constants
const (
	PaymentStatusSucceeded = "succeeded"
	PaymentStatusCancelled = "cancelled"
)

// CreatePaymentRequest represents the request to create a new payment
type CreatePaymentRequest struct {
	BookingID       int     `json:"booking_id" validate:"required"`
	Amount          float64 `json:"amount" validate:"required,min=0"`
	StripePaymentID string  `json:"stripe_payment_id,omitempty"`
}

// UpdatePaymentRequest represents the request to update a payment
type UpdatePaymentRequest struct {
	Status          *string `json:"status,omitempty"`
	StripePaymentID *string `json:"stripe_payment_id,omitempty"`
}

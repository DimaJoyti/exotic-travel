package models

import (
	"time"
)

// Review represents a user review for a destination
type Review struct {
	ID            int       `json:"id" db:"id"`
	UserID        int       `json:"user_id" db:"user_id"`
	DestinationID int       `json:"destination_id" db:"destination_id"`
	Rating        int       `json:"rating" db:"rating"`
	Comment       string    `json:"comment" db:"comment"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	
	// Joined fields
	User        *User        `json:"user,omitempty"`
	Destination *Destination `json:"destination,omitempty"`
}

// CreateReviewRequest represents the request to create a new review
type CreateReviewRequest struct {
	DestinationID int    `json:"destination_id" validate:"required"`
	Rating        int    `json:"rating" validate:"required,min=1,max=5"`
	Comment       string `json:"comment" validate:"max=1000"`
}

// UpdateReviewRequest represents the request to update a review
type UpdateReviewRequest struct {
	Rating  *int    `json:"rating,omitempty" validate:"omitempty,min=1,max=5"`
	Comment *string `json:"comment,omitempty" validate:"omitempty,max=1000"`
}

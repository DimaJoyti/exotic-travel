package models

import (
	"time"
)

// Destination represents an exotic travel destination
type Destination struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Country     string    `json:"country" db:"country"`
	City        string    `json:"city" db:"city"`
	Price       float64   `json:"price" db:"price"`
	Duration    int       `json:"duration" db:"duration"` // Duration in days
	MaxGuests   int       `json:"max_guests" db:"max_guests"`
	Images      []string  `json:"images" db:"images"`
	Features    []string  `json:"features" db:"features"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreateDestinationRequest represents the request to create a new destination
type CreateDestinationRequest struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description" validate:"required"`
	Country     string   `json:"country" validate:"required"`
	City        string   `json:"city" validate:"required"`
	Price       float64  `json:"price" validate:"required,min=0"`
	Duration    int      `json:"duration" validate:"required,min=1"`
	MaxGuests   int      `json:"max_guests" validate:"required,min=1"`
	Images      []string `json:"images"`
	Features    []string `json:"features"`
}

// UpdateDestinationRequest represents the request to update a destination
type UpdateDestinationRequest struct {
	Name        *string   `json:"name,omitempty"`
	Description *string   `json:"description,omitempty"`
	Country     *string   `json:"country,omitempty"`
	City        *string   `json:"city,omitempty"`
	Price       *float64  `json:"price,omitempty"`
	Duration    *int      `json:"duration,omitempty"`
	MaxGuests   *int      `json:"max_guests,omitempty"`
	Images      *[]string `json:"images,omitempty"`
	Features    *[]string `json:"features,omitempty"`
}

// DestinationFilter represents filters for destination queries
type DestinationFilter struct {
	Country   string  `json:"country,omitempty"`
	City      string  `json:"city,omitempty"`
	MinPrice  float64 `json:"min_price,omitempty"`
	MaxPrice  float64 `json:"max_price,omitempty"`
	Duration  int     `json:"duration,omitempty"`
	MaxGuests int     `json:"max_guests,omitempty"`
	Search    string  `json:"search,omitempty"`
	Limit     int     `json:"limit,omitempty"`
	Offset    int     `json:"offset,omitempty"`
}

package services

import (
	"context"
	"fmt"

	"github.com/exotic-travel-booking/backend/internal/models"
	"github.com/exotic-travel-booking/backend/internal/repositories"
)

// DestinationService handles destination operations
type DestinationService struct {
	destinationRepo repositories.DestinationRepository
}

// NewDestinationService creates a new destination service
func NewDestinationService(destinationRepo repositories.DestinationRepository) *DestinationService {
	return &DestinationService{
		destinationRepo: destinationRepo,
	}
}

// Create creates a new destination
func (s *DestinationService) Create(ctx context.Context, req *models.CreateDestinationRequest) (*models.Destination, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	destination := &models.Destination{
		Name:        req.Name,
		Description: req.Description,
		Country:     req.Country,
		City:        req.City,
		Price:       req.Price,
		Duration:    req.Duration,
		MaxGuests:   req.MaxGuests,
		Images:      req.Images,
		Features:    req.Features,
	}

	if destination.Images == nil {
		destination.Images = []string{}
	}
	if destination.Features == nil {
		destination.Features = []string{}
	}

	err := s.destinationRepo.Create(ctx, destination)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination: %w", err)
	}

	return destination, nil
}

// GetByID retrieves a destination by ID
func (s *DestinationService) GetByID(ctx context.Context, id int) (*models.Destination, error) {
	destination, err := s.destinationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get destination: %w", err)
	}

	return destination, nil
}

// Update updates a destination
func (s *DestinationService) Update(ctx context.Context, id int, req *models.UpdateDestinationRequest) (*models.Destination, error) {
	// Get existing destination
	destination, err := s.destinationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("destination not found: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		destination.Name = *req.Name
	}
	if req.Description != nil {
		destination.Description = *req.Description
	}
	if req.Country != nil {
		destination.Country = *req.Country
	}
	if req.City != nil {
		destination.City = *req.City
	}
	if req.Price != nil {
		destination.Price = *req.Price
	}
	if req.Duration != nil {
		destination.Duration = *req.Duration
	}
	if req.MaxGuests != nil {
		destination.MaxGuests = *req.MaxGuests
	}
	if req.Images != nil {
		destination.Images = *req.Images
	}
	if req.Features != nil {
		destination.Features = *req.Features
	}

	// Validate updated destination
	if err := s.validateDestination(destination); err != nil {
		return nil, err
	}

	err = s.destinationRepo.Update(ctx, destination)
	if err != nil {
		return nil, fmt.Errorf("failed to update destination: %w", err)
	}

	return destination, nil
}

// Delete deletes a destination
func (s *DestinationService) Delete(ctx context.Context, id int) error {
	err := s.destinationRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete destination: %w", err)
	}

	return nil
}

// List retrieves destinations with filtering
func (s *DestinationService) List(ctx context.Context, filter *models.DestinationFilter) ([]*models.Destination, error) {
	// Set default values
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	destinations, err := s.destinationRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list destinations: %w", err)
	}

	return destinations, nil
}

// Search performs full-text search on destinations
func (s *DestinationService) Search(ctx context.Context, query string, limit, offset int) ([]*models.Destination, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	destinations, err := s.destinationRepo.Search(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search destinations: %w", err)
	}

	return destinations, nil
}

// validateCreateRequest validates a create destination request
func (s *DestinationService) validateCreateRequest(req *models.CreateDestinationRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Description == "" {
		return fmt.Errorf("description is required")
	}
	if req.Country == "" {
		return fmt.Errorf("country is required")
	}
	if req.City == "" {
		return fmt.Errorf("city is required")
	}
	if req.Price <= 0 {
		return fmt.Errorf("price must be greater than 0")
	}
	if req.Duration <= 0 {
		return fmt.Errorf("duration must be greater than 0")
	}
	if req.MaxGuests <= 0 {
		return fmt.Errorf("max guests must be greater than 0")
	}

	return nil
}

// validateDestination validates a destination
func (s *DestinationService) validateDestination(destination *models.Destination) error {
	if destination.Name == "" {
		return fmt.Errorf("name is required")
	}
	if destination.Description == "" {
		return fmt.Errorf("description is required")
	}
	if destination.Country == "" {
		return fmt.Errorf("country is required")
	}
	if destination.City == "" {
		return fmt.Errorf("city is required")
	}
	if destination.Price <= 0 {
		return fmt.Errorf("price must be greater than 0")
	}
	if destination.Duration <= 0 {
		return fmt.Errorf("duration must be greater than 0")
	}
	if destination.MaxGuests <= 0 {
		return fmt.Errorf("max guests must be greater than 0")
	}

	return nil
}

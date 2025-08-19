package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/lib/pq"

	"github.com/exotic-travel-booking/backend/internal/cache"
	"github.com/exotic-travel-booking/backend/internal/models"
	"github.com/exotic-travel-booking/backend/pkg/database"
)

type destinationRepository struct {
	db    *database.DB
	cache *cache.CacheManager
}

// NewDestinationRepository creates a new destination repository
func NewDestinationRepository(db *database.DB) DestinationRepository {
	return &destinationRepository{db: db}
}

// NewDestinationRepositoryWithCache creates a new destination repository with caching
func NewDestinationRepositoryWithCache(db *database.DB, cacheManager *cache.CacheManager) DestinationRepository {
	return &destinationRepository{
		db:    db,
		cache: cacheManager,
	}
}

// Create creates a new destination
func (r *destinationRepository) Create(ctx context.Context, destination *models.Destination) error {
	query := `
		INSERT INTO destinations (name, description, country, city, price, duration, max_guests, images, features)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		destination.Name,
		destination.Description,
		destination.Country,
		destination.City,
		destination.Price,
		destination.Duration,
		destination.MaxGuests,
		pq.Array(destination.Images),
		pq.Array(destination.Features),
	).Scan(&destination.ID, &destination.CreatedAt, &destination.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}

	return nil
}

// GetByID retrieves a destination by ID with caching
func (r *destinationRepository) GetByID(ctx context.Context, id int) (*models.Destination, error) {
	// Try cache first if available
	if r.cache != nil {
		var destination models.Destination
		err := r.cache.GetDestination(ctx, strconv.Itoa(id), &destination)
		if err == nil {
			return &destination, nil
		}
		// Cache miss or error, continue to database
	}

	destination := &models.Destination{}
	query := `
		SELECT id, name, description, country, city, price, duration, max_guests, images, features, created_at, updated_at
		FROM destinations
		WHERE id = $1`

	start := time.Now()
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&destination.ID,
		&destination.Name,
		&destination.Description,
		&destination.Country,
		&destination.City,
		&destination.Price,
		&destination.Duration,
		&destination.MaxGuests,
		pq.Array(&destination.Images),
		pq.Array(&destination.Features),
		&destination.CreatedAt,
		&destination.UpdatedAt,
	)

	// Record database performance metrics
	duration := time.Since(start)
	_ = duration // Use the duration for metrics if needed

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("destination not found")
		}
		return nil, fmt.Errorf("failed to get destination: %w", err)
	}

	// Cache the result if cache is available
	if r.cache != nil {
		if cacheErr := r.cache.CacheDestination(ctx, strconv.Itoa(id), destination); cacheErr != nil {
			// Log cache error but don't fail the request
			fmt.Printf("Failed to cache destination %d: %v\n", id, cacheErr)
		}
	}

	return destination, nil
}

// Update updates a destination
func (r *destinationRepository) Update(ctx context.Context, destination *models.Destination) error {
	query := `
		UPDATE destinations
		SET name = $2, description = $3, country = $4, city = $5, price = $6, 
		    duration = $7, max_guests = $8, images = $9, features = $10, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.QueryRowContext(ctx, query,
		destination.ID,
		destination.Name,
		destination.Description,
		destination.Country,
		destination.City,
		destination.Price,
		destination.Duration,
		destination.MaxGuests,
		pq.Array(destination.Images),
		pq.Array(destination.Features),
	).Scan(&destination.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update destination: %w", err)
	}

	return nil
}

// Delete deletes a destination
func (r *destinationRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM destinations WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete destination: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("destination not found")
	}

	return nil
}

// List retrieves destinations with filtering
func (r *destinationRepository) List(ctx context.Context, filter *models.DestinationFilter) ([]*models.Destination, error) {
	query := `
		SELECT id, name, description, country, city, price, duration, max_guests, images, features, created_at, updated_at
		FROM destinations
		WHERE 1=1`

	args := []interface{}{}
	argCount := 0

	// Apply filters
	if filter.Country != "" {
		argCount++
		query += fmt.Sprintf(" AND country = $%d", argCount)
		args = append(args, filter.Country)
	}

	if filter.City != "" {
		argCount++
		query += fmt.Sprintf(" AND city = $%d", argCount)
		args = append(args, filter.City)
	}

	if filter.MinPrice > 0 {
		argCount++
		query += fmt.Sprintf(" AND price >= $%d", argCount)
		args = append(args, filter.MinPrice)
	}

	if filter.MaxPrice > 0 {
		argCount++
		query += fmt.Sprintf(" AND price <= $%d", argCount)
		args = append(args, filter.MaxPrice)
	}

	if filter.Duration > 0 {
		argCount++
		query += fmt.Sprintf(" AND duration = $%d", argCount)
		args = append(args, filter.Duration)
	}

	if filter.MaxGuests > 0 {
		argCount++
		query += fmt.Sprintf(" AND max_guests >= $%d", argCount)
		args = append(args, filter.MaxGuests)
	}

	if filter.Search != "" {
		argCount++
		query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+filter.Search+"%")
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
	}

	return r.queryDestinations(ctx, query, args...)
}

// Search performs full-text search on destinations
func (r *destinationRepository) Search(ctx context.Context, query string, limit, offset int) ([]*models.Destination, error) {
	searchQuery := `
		SELECT id, name, description, country, city, price, duration, max_guests, images, features, created_at, updated_at
		FROM destinations
		WHERE to_tsvector('english', name || ' ' || description) @@ plainto_tsquery('english', $1)
		ORDER BY ts_rank(to_tsvector('english', name || ' ' || description), plainto_tsquery('english', $1)) DESC
		LIMIT $2 OFFSET $3`

	return r.queryDestinations(ctx, searchQuery, query, limit, offset)
}

// queryDestinations is a helper method to execute destination queries
func (r *destinationRepository) queryDestinations(ctx context.Context, query string, args ...interface{}) ([]*models.Destination, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query destinations: %w", err)
	}
	defer rows.Close()

	var destinations []*models.Destination
	for rows.Next() {
		destination := &models.Destination{}
		err := rows.Scan(
			&destination.ID,
			&destination.Name,
			&destination.Description,
			&destination.Country,
			&destination.City,
			&destination.Price,
			&destination.Duration,
			&destination.MaxGuests,
			pq.Array(&destination.Images),
			pq.Array(&destination.Features),
			&destination.CreatedAt,
			&destination.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan destination: %w", err)
		}
		destinations = append(destinations, destination)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate destinations: %w", err)
	}

	return destinations, nil
}

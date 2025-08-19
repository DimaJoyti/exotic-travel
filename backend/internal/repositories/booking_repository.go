package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/exotic-travel-booking/backend/internal/models"
	"github.com/exotic-travel-booking/backend/pkg/database"
)

type bookingRepository struct {
	db *database.DB
}

// NewBookingRepository creates a new booking repository
func NewBookingRepository(db *database.DB) BookingRepository {
	return &bookingRepository{db: db}
}

// Create creates a new booking
func (r *bookingRepository) Create(ctx context.Context, booking *models.Booking) error {
	query := `
		INSERT INTO bookings (user_id, destination_id, start_date, end_date, guests, total_price, status, payment_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		booking.UserID,
		booking.DestinationID,
		booking.StartDate,
		booking.EndDate,
		booking.Guests,
		booking.TotalPrice,
		booking.Status,
		booking.PaymentStatus,
	).Scan(&booking.ID, &booking.CreatedAt, &booking.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create booking: %w", err)
	}

	return nil
}

// GetByID retrieves a booking by ID
func (r *bookingRepository) GetByID(ctx context.Context, id int) (*models.Booking, error) {
	booking := &models.Booking{}
	query := `
		SELECT b.id, b.user_id, b.destination_id, b.start_date, b.end_date, b.guests, 
		       b.total_price, b.status, b.payment_status, b.created_at, b.updated_at,
		       u.id, u.email, u.first_name, u.last_name, u.role, u.created_at, u.updated_at,
		       d.id, d.name, d.description, d.country, d.city, d.price, d.duration, 
		       d.max_guests, d.images, d.features, d.created_at, d.updated_at
		FROM bookings b
		LEFT JOIN users u ON b.user_id = u.id
		LEFT JOIN destinations d ON b.destination_id = d.id
		WHERE b.id = $1`

	var user models.User
	var destination models.Destination

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&booking.ID, &booking.UserID, &booking.DestinationID, &booking.StartDate,
		&booking.EndDate, &booking.Guests, &booking.TotalPrice, &booking.Status,
		&booking.PaymentStatus, &booking.CreatedAt, &booking.UpdatedAt,
		&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Role,
		&user.CreatedAt, &user.UpdatedAt,
		&destination.ID, &destination.Name, &destination.Description, &destination.Country,
		&destination.City, &destination.Price, &destination.Duration, &destination.MaxGuests,
		&destination.Images, &destination.Features, &destination.CreatedAt, &destination.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("booking not found")
		}
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	booking.User = &user
	booking.Destination = &destination
	return booking, nil
}

// Update updates a booking
func (r *bookingRepository) Update(ctx context.Context, booking *models.Booking) error {
	query := `
		UPDATE bookings
		SET start_date = $2, end_date = $3, guests = $4, total_price = $5, 
		    status = $6, payment_status = $7, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.QueryRowContext(ctx, query,
		booking.ID,
		booking.StartDate,
		booking.EndDate,
		booking.Guests,
		booking.TotalPrice,
		booking.Status,
		booking.PaymentStatus,
	).Scan(&booking.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update booking: %w", err)
	}

	return nil
}

// Delete deletes a booking
func (r *bookingRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM bookings WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete booking: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("booking not found")
	}

	return nil
}

// List retrieves bookings with filtering
func (r *bookingRepository) List(ctx context.Context, filter *models.BookingFilter) ([]*models.Booking, error) {
	query := `
		SELECT b.id, b.user_id, b.destination_id, b.start_date, b.end_date, b.guests, 
		       b.total_price, b.status, b.payment_status, b.created_at, b.updated_at,
		       u.id, u.email, u.first_name, u.last_name, u.role, u.created_at, u.updated_at,
		       d.id, d.name, d.description, d.country, d.city, d.price, d.duration, 
		       d.max_guests, d.images, d.features, d.created_at, d.updated_at
		FROM bookings b
		LEFT JOIN users u ON b.user_id = u.id
		LEFT JOIN destinations d ON b.destination_id = d.id
		WHERE 1=1`

	args := []interface{}{}
	argCount := 0

	// Apply filters
	if filter.UserID > 0 {
		argCount++
		query += fmt.Sprintf(" AND b.user_id = $%d", argCount)
		args = append(args, filter.UserID)
	}

	if filter.DestinationID > 0 {
		argCount++
		query += fmt.Sprintf(" AND b.destination_id = $%d", argCount)
		args = append(args, filter.DestinationID)
	}

	if filter.Status != "" {
		argCount++
		query += fmt.Sprintf(" AND b.status = $%d", argCount)
		args = append(args, filter.Status)
	}

	if filter.PaymentStatus != "" {
		argCount++
		query += fmt.Sprintf(" AND b.payment_status = $%d", argCount)
		args = append(args, filter.PaymentStatus)
	}

	query += " ORDER BY b.created_at DESC"

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

	return r.queryBookings(ctx, query, args...)
}

// GetByUserID retrieves bookings for a specific user
func (r *bookingRepository) GetByUserID(ctx context.Context, userID int, limit, offset int) ([]*models.Booking, error) {
	query := `
		SELECT b.id, b.user_id, b.destination_id, b.start_date, b.end_date, b.guests, 
		       b.total_price, b.status, b.payment_status, b.created_at, b.updated_at,
		       u.id, u.email, u.first_name, u.last_name, u.role, u.created_at, u.updated_at,
		       d.id, d.name, d.description, d.country, d.city, d.price, d.duration, 
		       d.max_guests, d.images, d.features, d.created_at, d.updated_at
		FROM bookings b
		LEFT JOIN users u ON b.user_id = u.id
		LEFT JOIN destinations d ON b.destination_id = d.id
		WHERE b.user_id = $1
		ORDER BY b.created_at DESC
		LIMIT $2 OFFSET $3`

	return r.queryBookings(ctx, query, userID, limit, offset)
}

// queryBookings is a helper method to execute booking queries
func (r *bookingRepository) queryBookings(ctx context.Context, query string, args ...interface{}) ([]*models.Booking, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query bookings: %w", err)
	}
	defer rows.Close()

	var bookings []*models.Booking
	for rows.Next() {
		booking := &models.Booking{}
		user := &models.User{}
		destination := &models.Destination{}

		err := rows.Scan(
			&booking.ID, &booking.UserID, &booking.DestinationID, &booking.StartDate,
			&booking.EndDate, &booking.Guests, &booking.TotalPrice, &booking.Status,
			&booking.PaymentStatus, &booking.CreatedAt, &booking.UpdatedAt,
			&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Role,
			&user.CreatedAt, &user.UpdatedAt,
			&destination.ID, &destination.Name, &destination.Description, &destination.Country,
			&destination.City, &destination.Price, &destination.Duration, &destination.MaxGuests,
			&destination.Images, &destination.Features, &destination.CreatedAt, &destination.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}

		booking.User = user
		booking.Destination = destination
		bookings = append(bookings, booking)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate bookings: %w", err)
	}

	return bookings, nil
}

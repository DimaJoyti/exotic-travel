package repositories

import (
	"context"

	"github.com/exotic-travel-booking/backend/internal/models"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id int) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, limit, offset int) ([]*models.User, error)
}

// DestinationRepository defines the interface for destination data operations
type DestinationRepository interface {
	Create(ctx context.Context, destination *models.Destination) error
	GetByID(ctx context.Context, id int) (*models.Destination, error)
	Update(ctx context.Context, destination *models.Destination) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter *models.DestinationFilter) ([]*models.Destination, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*models.Destination, error)
}

// BookingRepository defines the interface for booking data operations
type BookingRepository interface {
	Create(ctx context.Context, booking *models.Booking) error
	GetByID(ctx context.Context, id int) (*models.Booking, error)
	Update(ctx context.Context, booking *models.Booking) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter *models.BookingFilter) ([]*models.Booking, error)
	GetByUserID(ctx context.Context, userID int, limit, offset int) ([]*models.Booking, error)
}

// ReviewRepository defines the interface for review data operations
type ReviewRepository interface {
	Create(ctx context.Context, review *models.Review) error
	GetByID(ctx context.Context, id int) (*models.Review, error)
	Update(ctx context.Context, review *models.Review) error
	Delete(ctx context.Context, id int) error
	GetByDestinationID(ctx context.Context, destinationID int, limit, offset int) ([]*models.Review, error)
	GetByUserID(ctx context.Context, userID int, limit, offset int) ([]*models.Review, error)
}

// PaymentRepository defines the interface for payment data operations
type PaymentRepository interface {
	Create(ctx context.Context, payment *models.Payment) error
	GetByID(ctx context.Context, id int) (*models.Payment, error)
	GetByBookingID(ctx context.Context, bookingID int) (*models.Payment, error)
	GetByStripePaymentID(ctx context.Context, stripePaymentID string) (*models.Payment, error)
	Update(ctx context.Context, payment *models.Payment) error
	List(ctx context.Context, limit, offset int) ([]*models.Payment, error)
}

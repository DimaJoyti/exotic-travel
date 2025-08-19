package services

import (
	"context"
	"fmt"
	"time"

	"github.com/exotic-travel-booking/backend/internal/models"
	"github.com/exotic-travel-booking/backend/internal/repositories"
	"github.com/exotic-travel-booking/backend/pkg/auth"
)

// AuthService handles authentication operations
type AuthService struct {
	userRepo   repositories.UserRepository
	jwtManager *auth.JWTManager
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo repositories.UserRepository, jwtSecret string) *AuthService {
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour) // 24 hour token expiry
	return &AuthService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req *models.CreateUserRequest) (*models.LoginResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         models.RoleUser,
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Clear password hash from response
	user.PasswordHash = ""

	return &models.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

// Login authenticates a user and returns a token
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check password
	err = auth.CheckPassword(req.Password, user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Clear password hash from response
	user.PasswordHash = ""

	return &models.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

// RefreshToken generates a new token from an existing token
func (s *AuthService) RefreshToken(ctx context.Context, tokenString string) (string, error) {
	return s.jwtManager.RefreshToken(tokenString)
}

// ValidateToken validates a token and returns user claims
func (s *AuthService) ValidateToken(tokenString string) (*auth.Claims, error) {
	return s.jwtManager.ValidateToken(tokenString)
}

// GetUserByID retrieves a user by ID (for middleware)
func (s *AuthService) GetUserByID(ctx context.Context, userID int) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Clear password hash
	user.PasswordHash = ""
	return user, nil
}

package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/exotic-travel-booking/backend/internal/models"
	"github.com/exotic-travel-booking/backend/pkg/auth"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func TestAuthService_Register_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo, "test-jwt-secret")
	
	request := models.CreateUserRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	
	// Mock expectations
	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, assert.AnError)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
	
	// Execute
	response, err := service.Register(context.Background(), &request)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, request.Email, response.User.Email)
	assert.Equal(t, request.FirstName, response.User.FirstName)
	assert.Equal(t, request.LastName, response.User.LastName)
	assert.Equal(t, "user", response.User.Role)
	assert.NotEmpty(t, response.Token)
	
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo, "test-jwt-secret")
	
	request := models.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	
	// Create a user with properly hashed password
	hashedPassword, err := auth.HashPassword("password123")
	assert.NoError(t, err)
	
	user := &models.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		FirstName:    "John",
		LastName:     "Doe",
		Role:         "user",
	}
	
	// Mock expectations
	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
	
	// Execute
	response, err := service.Login(context.Background(), &request)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, user.Email, response.User.Email)
	assert.NotEmpty(t, response.Token)
	
	mockRepo.AssertExpectations(t)
}
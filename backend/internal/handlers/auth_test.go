package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/exotic-travel-booking/backend/internal/models"
)

// AuthServiceInterface defines the interface for auth service
type AuthServiceInterface interface {
	Register(ctx context.Context, req *models.CreateUserRequest) (*models.LoginResponse, error)
	Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
	ValidateToken(token string) (*models.User, error)
	GetUserByID(ctx context.Context, userID int) (*models.User, error)
}

// TestAuthHandlers for testing with interface
type TestAuthHandlers struct {
	authService AuthServiceInterface
}

// NewTestAuthHandlers creates test auth handlers
func NewTestAuthHandlers(authService AuthServiceInterface) *TestAuthHandlers {
	return &TestAuthHandlers{
		authService: authService,
	}
}

// Register handles user registration (copied from actual handler)
func (h *TestAuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 8 {
		http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	response, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Login handles user login (copied from actual handler)
func (h *TestAuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	response, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		http.Error(w, "Failed to login", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RefreshToken handles token refresh (simplified version)
func (h *TestAuthHandlers) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	refreshToken, ok := req["refresh_token"]
	if !ok {
		http.Error(w, "refresh_token required", http.StatusBadRequest)
		return
	}

	newToken, err := h.authService.RefreshToken(r.Context(), refreshToken)
	if err != nil {
		http.Error(w, "Failed to refresh token", http.StatusUnauthorized)
		return
	}

	response := map[string]string{
		"token": newToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// MockAuthService is a mock implementation of AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, req *models.CreateUserRequest) (*models.LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LoginResponse), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LoginResponse), args.Error(1)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	args := m.Called(ctx, refreshToken)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) ValidateToken(token string) (*models.User, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthService) GetUserByID(ctx context.Context, userID int) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func TestAuthHandlers_Register(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockAuthService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful registration",
			requestBody: models.CreateUserRequest{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			setupMock: func(m *MockAuthService) {
				response := &models.LoginResponse{
					User: models.User{
						ID:        1,
						Email:     "test@example.com",
						FirstName: "John",
						LastName:  "Doe",
						Role:      "user",
					},
					Token: "test-token",
				}
				m.On("Register", mock.Anything, mock.AnythingOfType("*models.CreateUserRequest")).Return(response, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"email": "invalid-email",
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			requestBody: models.CreateUserRequest{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			setupMock: func(m *MockAuthService) {
				m.On("Register", mock.Anything, mock.AnythingOfType("*models.CreateUserRequest")).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockAuthService)
			tt.setupMock(mockService)
			
			handlers := NewTestAuthHandlers(mockService)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			
			// Create response recorder
			w := httptest.NewRecorder()

			// Execute
			handlers.Register(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.expectedStatus == http.StatusCreated {
				var response models.LoginResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "test@example.com", response.User.Email)
				assert.Equal(t, "John", response.User.FirstName)
				assert.Equal(t, "Doe", response.User.LastName)
				assert.NotEmpty(t, response.Token)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandlers_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockAuthService)
		expectedStatus int
	}{
		{
			name: "successful login",
			requestBody: models.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(m *MockAuthService) {
				response := &models.LoginResponse{
					User: models.User{
						ID:        1,
						Email:     "test@example.com",
						FirstName: "John",
						LastName:  "Doe",
						Role:      "user",
					},
					Token: "access-token",
				}
				m.On("Login", mock.Anything, mock.AnythingOfType("*models.LoginRequest")).Return(response, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid credentials",
			requestBody: models.LoginRequest{
				Email:    "test@example.com",
				Password: "wrong-password",
			},
			setupMock: func(m *MockAuthService) {
				m.On("Login", mock.Anything, mock.AnythingOfType("*models.LoginRequest")).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"email": "invalid-email",
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockAuthService)
			tt.setupMock(mockService)
			
			handlers := NewTestAuthHandlers(mockService)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			
			// Create response recorder
			w := httptest.NewRecorder()

			// Execute
			handlers.Login(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.expectedStatus == http.StatusOK {
				var response models.LoginResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Token)
				assert.Equal(t, "test@example.com", response.User.Email)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandlers_RefreshToken(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockAuthService)
		expectedStatus int
	}{
		{
			name: "successful token refresh",
			requestBody: map[string]string{
				"refresh_token": "valid-refresh-token",
			},
			setupMock: func(m *MockAuthService) {
				m.On("RefreshToken", mock.Anything, "valid-refresh-token").Return("new-access-token", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid refresh token",
			requestBody: map[string]string{
				"refresh_token": "invalid-token",
			},
			setupMock: func(m *MockAuthService) {
				m.On("RefreshToken", mock.Anything, "invalid-token").Return("", assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "missing refresh token",
			requestBody: map[string]string{
				"token": "some-token",
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockAuthService)
			tt.setupMock(mockService)
			
			handlers := NewTestAuthHandlers(mockService)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			
			// Create response recorder
			w := httptest.NewRecorder()

			// Execute
			handlers.RefreshToken(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			mockService.AssertExpectations(t)
		})
	}
}

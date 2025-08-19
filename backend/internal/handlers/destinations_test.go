package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/exotic-travel-booking/backend/internal/models"
)

// MockDestinationService is a mock implementation of DestinationService
type MockDestinationService struct {
	mock.Mock
}

func (m *MockDestinationService) List(ctx context.Context, filter *models.DestinationFilter) ([]*models.Destination, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*models.Destination), args.Error(1)
}

func (m *MockDestinationService) GetByID(ctx context.Context, id int) (*models.Destination, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Destination), args.Error(1)
}

func (m *MockDestinationService) Create(ctx context.Context, req *models.CreateDestinationRequest) (*models.Destination, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Destination), args.Error(1)
}

func (m *MockDestinationService) Update(ctx context.Context, id int, req *models.UpdateDestinationRequest) (*models.Destination, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Destination), args.Error(1)
}

func (m *MockDestinationService) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDestinationService) Search(ctx context.Context, query string, limit, offset int) ([]*models.Destination, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0).([]*models.Destination), args.Error(1)
}

func TestDestinationHandlers_List(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(*MockDestinationService)
		expectedStatus int
		expectedCount  int
	}{
		{
			name:        "successful list all destinations",
			queryParams: "",
			setupMock: func(m *MockDestinationService) {
				destinations := []*models.Destination{
					{
						ID:          1,
						Name:        "Maldives",
						Country:     "Maldives",
						Description: "Beautiful tropical paradise",
						Price:       500.00,
						//Rating:      4.8,
						Images: []string{"image1.jpg", "image2.jpg"},
					},
					{
						ID:          2,
						Name:        "Bali",
						Country:     "Indonesia",
						Description: "Cultural and natural beauty",
						Price:       200.00,
						//Rating:      4.6,
						Images: []string{"bali1.jpg", "bali2.jpg"},
					},
				}
				m.On("List", mock.Anything, mock.AnythingOfType("*models.DestinationFilter")).Return(destinations, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:        "list with country filter",
			queryParams: "?country=Maldives",
			setupMock: func(m *MockDestinationService) {
				destinations := []*models.Destination{
					{
						ID:          1,
						Name:        "Maldives",
						Country:     "Maldives",
						Description: "Beautiful tropical paradise",
						Price:       500.00,
						//Rating:      4.8,
						Images: []string{"image1.jpg", "image2.jpg"},
					},
				}
				m.On("List", mock.Anything, mock.AnythingOfType("*models.DestinationFilter")).Return(destinations, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:        "service error",
			queryParams: "",
			setupMock: func(m *MockDestinationService) {
				m.On("List", mock.Anything, mock.AnythingOfType("*models.DestinationFilter")).Return([]*models.Destination{}, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockDestinationService)
			tt.setupMock(mockService)

			handlers := NewDestinationHandlers(mockService)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/destinations"+tt.queryParams, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute
			handlers.List(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var destinations []*models.Destination
				err := json.Unmarshal(w.Body.Bytes(), &destinations)
				assert.NoError(t, err)
				assert.Len(t, destinations, tt.expectedCount)

				if tt.expectedCount > 0 {
					assert.NotEmpty(t, destinations[0].Name)
					assert.NotEmpty(t, destinations[0].Country)
					assert.Greater(t, destinations[0].Price, 0.0)
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestDestinationHandlers_GetByID(t *testing.T) {
	tests := []struct {
		name           string
		destinationID  string
		setupMock      func(*MockDestinationService)
		expectedStatus int
	}{
		{
			name:          "successful get destination",
			destinationID: "1",
			setupMock: func(m *MockDestinationService) {
				destination := &models.Destination{
					ID:          1,
					Name:        "Maldives",
					Country:     "Maldives",
					Description: "Beautiful tropical paradise",
					Price:       500.00,
					//Rating:      4.8,
					Images: []string{"image1.jpg", "image2.jpg"},
				}
				m.On("GetByID", mock.Anything, 1).Return(destination, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "destination not found",
			destinationID: "999",
			setupMock: func(m *MockDestinationService) {
				m.On("GetByID", mock.Anything, 999).Return(nil, errors.New("destination not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid destination ID",
			destinationID:  "invalid",
			setupMock:      func(m *MockDestinationService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockDestinationService)
			tt.setupMock(mockService)

			handlers := NewDestinationHandlers(mockService)

			// Create request with path parameter
			req := httptest.NewRequest(http.MethodGet, "/api/destinations/"+tt.destinationID, nil)
			req.SetPathValue("id", tt.destinationID)

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute
			handlers.GetByID(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var destination models.Destination
				err := json.Unmarshal(w.Body.Bytes(), &destination)
				assert.NoError(t, err)
				assert.Equal(t, "Maldives", destination.Name)
				assert.Equal(t, "Maldives", destination.Country)
				assert.Equal(t, 500.00, destination.Price)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestDestinationHandlers_Create(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockDestinationService)
		expectedStatus int
	}{
		{
			name: "successful creation",
			requestBody: models.CreateDestinationRequest{
				Name:        "New Destination",
				Country:     "Test Country",
				Description: "A test destination",
				Price:       300.00,
				Images:      []string{"test1.jpg", "test2.jpg"},
			},
			setupMock: func(m *MockDestinationService) {
				destination := &models.Destination{
					ID:          1,
					Name:        "New Destination",
					Country:     "Test Country",
					Description: "A test destination",
					Price:       300.00,
					//Rating:      0.0,
					Images: []string{"test1.jpg", "test2.jpg"},
				}
				m.On("Create", mock.Anything, mock.AnythingOfType("*models.CreateDestinationRequest")).Return(destination, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"name": "", // Empty name should be invalid
			},
			setupMock: func(m *MockDestinationService) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*models.CreateDestinationRequest")).Return(nil, errors.New("name is required"))
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			requestBody: models.CreateDestinationRequest{
				Name:        "New Destination",
				Country:     "Test Country",
				Description: "A test destination",
				Price:       300.00,
				Images:      []string{"test1.jpg", "test2.jpg"},
			},
			setupMock: func(m *MockDestinationService) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*models.CreateDestinationRequest")).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockDestinationService)
			tt.setupMock(mockService)

			handlers := NewDestinationHandlers(mockService)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/destinations", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute
			handlers.Create(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var destination models.Destination
				err := json.Unmarshal(w.Body.Bytes(), &destination)
				assert.NoError(t, err)
				assert.Equal(t, "New Destination", destination.Name)
				assert.Equal(t, "Test Country", destination.Country)
				assert.Equal(t, 300.00, destination.Price)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestDestinationHandlers_Search(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(*MockDestinationService)
		expectedStatus int
		expectedCount  int
	}{
		{
			name:        "successful search",
			queryParams: "?q=maldives",
			setupMock: func(m *MockDestinationService) {
				destinations := []*models.Destination{
					{
						ID:          1,
						Name:        "Maldives Resort",
						Country:     "Maldives",
						Description: "Beautiful tropical paradise",
						Price:       500.00,
						//Rating:      4.8,
						Images: []string{"image1.jpg", "image2.jpg"},
					},
				}
				m.On("Search", mock.Anything, "maldives", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(destinations, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:        "no results found",
			queryParams: "?q=nonexistent",
			setupMock: func(m *MockDestinationService) {
				m.On("Search", mock.Anything, "nonexistent", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return([]*models.Destination{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:           "missing search query",
			queryParams:    "",
			setupMock:      func(m *MockDestinationService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockDestinationService)
			tt.setupMock(mockService)

			handlers := NewDestinationHandlers(mockService)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/destinations/search"+tt.queryParams, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute
			handlers.Search(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var destinations []*models.Destination
				err := json.Unmarshal(w.Body.Bytes(), &destinations)
				assert.NoError(t, err)
				assert.Len(t, destinations, tt.expectedCount)
			}

			mockService.AssertExpectations(t)
		})
	}
}

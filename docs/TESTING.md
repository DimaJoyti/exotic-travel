# Testing Guide

This document covers the comprehensive testing strategy for the Exotic Travel Booking Platform, including unit tests, integration tests, security tests, and performance tests.

## Testing Philosophy

Our testing approach follows the **Testing Pyramid**:
- **Unit Tests (70%)**: Fast, isolated tests for individual components
- **Integration Tests (20%)**: Tests for component interactions
- **End-to-End Tests (10%)**: Full user journey tests

## Backend Testing

### Unit Testing

#### Test Structure
```go
// Example: User service unit test
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        user    *models.User
        mockFn  func(*MockUserRepository)
        wantErr bool
        errMsg  string
    }{
        {
            name: "successful user creation",
            user: &models.User{
                Name:  "John Doe",
                Email: "john@example.com",
            },
            mockFn: func(m *MockUserRepository) {
                m.On("Create", mock.AnythingOfType("*models.User")).Return(nil)
                m.On("GetByEmail", "john@example.com").Return(nil, errors.New("not found"))
            },
            wantErr: false,
        },
        {
            name: "duplicate email error",
            user: &models.User{
                Name:  "John Doe",
                Email: "existing@example.com",
            },
            mockFn: func(m *MockUserRepository) {
                existingUser := &models.User{ID: 1, Email: "existing@example.com"}
                m.On("GetByEmail", "existing@example.com").Return(existingUser, nil)
            },
            wantErr: true,
            errMsg:  "email already exists",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := &MockUserRepository{}
            tt.mockFn(mockRepo)
            
            service := services.NewUserService(mockRepo)
            err := service.CreateUser(tt.user)
            
            if tt.wantErr {
                assert.Error(t, err)
                if tt.errMsg != "" {
                    assert.Contains(t, err.Error(), tt.errMsg)
                }
            } else {
                assert.NoError(t, err)
            }
            
            mockRepo.AssertExpectations(t)
        })
    }
}
```

#### Mock Generation
```go
//go:generate mockery --name=UserRepository --output=mocks --outpkg=mocks

// Generate mocks
go generate ./...
```

#### Running Unit Tests
```bash
# Run all unit tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run tests with race detection
go test -race ./...

# Run specific test
go test -run TestUserService_CreateUser ./internal/services

# Verbose output
go test -v ./...
```

### Integration Testing

#### Database Integration Tests
```go
func TestUserRepository_Integration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    repo := repositories.NewUserRepository(db)
    
    t.Run("create and retrieve user", func(t *testing.T) {
        user := &models.User{
            Name:  "Integration Test User",
            Email: "integration@test.com",
        }
        
        // Create user
        err := repo.Create(user)
        assert.NoError(t, err)
        assert.NotZero(t, user.ID)
        
        // Retrieve user
        retrieved, err := repo.GetByID(user.ID)
        assert.NoError(t, err)
        assert.Equal(t, user.Name, retrieved.Name)
        assert.Equal(t, user.Email, retrieved.Email)
    })
    
    t.Run("unique email constraint", func(t *testing.T) {
        user1 := &models.User{Name: "User 1", Email: "unique@test.com"}
        user2 := &models.User{Name: "User 2", Email: "unique@test.com"}
        
        err := repo.Create(user1)
        assert.NoError(t, err)
        
        err = repo.Create(user2)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "duplicate key")
    })
}

// Test database setup
func setupTestDB(t *testing.T) *database.DB {
    config := &database.Config{
        Host:     "localhost",
        Port:     5432,
        Database: "exotic_travel_test",
        Username: "test_user",
        Password: "test_password",
    }
    
    db, err := database.Connect(config)
    require.NoError(t, err)
    
    // Run migrations
    err = runMigrations(db)
    require.NoError(t, err)
    
    return db
}

func cleanupTestDB(t *testing.T, db *database.DB) {
    // Clean up test data
    _, err := db.Exec("TRUNCATE TABLE users, destinations, bookings, reviews CASCADE")
    require.NoError(t, err)
    
    db.Close()
}
```

#### API Integration Tests
```go
func TestDestinationHandler_Integration(t *testing.T) {
    // Setup test server
    server := setupTestServer(t)
    defer server.Close()
    
    client := &http.Client{}
    
    t.Run("get destinations", func(t *testing.T) {
        req, err := http.NewRequest("GET", server.URL+"/api/destinations", nil)
        require.NoError(t, err)
        
        resp, err := client.Do(req)
        require.NoError(t, err)
        defer resp.Body.Close()
        
        assert.Equal(t, http.StatusOK, resp.StatusCode)
        
        var response struct {
            Success bool                `json:"success"`
            Data    []models.Destination `json:"data"`
        }
        
        err = json.NewDecoder(resp.Body).Decode(&response)
        require.NoError(t, err)
        assert.True(t, response.Success)
    })
    
    t.Run("create destination requires authentication", func(t *testing.T) {
        payload := `{"name":"Test Destination","country":"Test Country"}`
        req, err := http.NewRequest("POST", server.URL+"/api/destinations", strings.NewReader(payload))
        require.NoError(t, err)
        req.Header.Set("Content-Type", "application/json")
        
        resp, err := client.Do(req)
        require.NoError(t, err)
        defer resp.Body.Close()
        
        assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
    })
}
```

### Test Utilities

#### Test Fixtures
```go
// fixtures/users.go
func CreateTestUser(t *testing.T, db *database.DB) *models.User {
    user := &models.User{
        Name:  "Test User",
        Email: fmt.Sprintf("test%d@example.com", time.Now().UnixNano()),
        Role:  "user",
    }
    
    repo := repositories.NewUserRepository(db)
    err := repo.Create(user)
    require.NoError(t, err)
    
    return user
}

func CreateTestDestination(t *testing.T, db *database.DB) *models.Destination {
    destination := &models.Destination{
        Name:        "Test Destination",
        Description: "A beautiful test destination",
        Country:     "Test Country",
        City:        "Test City",
        Price:       299.99,
        Duration:    7,
        MaxGuests:   4,
    }
    
    repo := repositories.NewDestinationRepository(db)
    err := repo.Create(destination)
    require.NoError(t, err)
    
    return destination
}
```

#### Test Helpers
```go
// testutil/auth.go
func GenerateTestJWT(t *testing.T, userID int64, role string) string {
    claims := &security.TokenClaims{
        UserID: userID,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte("test-secret"))
    require.NoError(t, err)
    
    return tokenString
}

func AuthenticatedRequest(t *testing.T, method, url string, body io.Reader, userID int64, role string) *http.Request {
    req, err := http.NewRequest(method, url, body)
    require.NoError(t, err)
    
    token := GenerateTestJWT(t, userID, role)
    req.Header.Set("Authorization", "Bearer "+token)
    
    return req
}
```

## Frontend Testing

### Component Testing

#### React Testing Library
```tsx
// components/DestinationCard.test.tsx
import { render, screen, fireEvent } from '@testing-library/react';
import { DestinationCard } from './DestinationCard';

const mockDestination = {
  id: 1,
  name: 'Bali Paradise',
  country: 'Indonesia',
  city: 'Ubud',
  price: 299.99,
  rating: 4.8,
  images: ['https://example.com/image.jpg'],
};

describe('DestinationCard', () => {
  test('renders destination information', () => {
    render(<DestinationCard destination={mockDestination} />);
    
    expect(screen.getByText('Bali Paradise')).toBeInTheDocument();
    expect(screen.getByText('Indonesia, Ubud')).toBeInTheDocument();
    expect(screen.getByText('$299.99')).toBeInTheDocument();
    expect(screen.getByText('4.8')).toBeInTheDocument();
  });
  
  test('calls onBook when book button is clicked', () => {
    const onBook = jest.fn();
    render(<DestinationCard destination={mockDestination} onBook={onBook} />);
    
    fireEvent.click(screen.getByText('Book Now'));
    
    expect(onBook).toHaveBeenCalledWith(mockDestination.id);
  });
  
  test('displays placeholder when image fails to load', () => {
    render(<DestinationCard destination={mockDestination} />);
    
    const image = screen.getByRole('img');
    fireEvent.error(image);
    
    expect(screen.getByText('Image not available')).toBeInTheDocument();
  });
});
```

#### Hook Testing
```tsx
// hooks/useAuth.test.tsx
import { renderHook, act } from '@testing-library/react';
import { useAuth } from './useAuth';

// Mock API
jest.mock('../lib/api', () => ({
  auth: {
    login: jest.fn(),
    logout: jest.fn(),
    getCurrentUser: jest.fn(),
  },
}));

describe('useAuth', () => {
  test('initializes with no user', () => {
    const { result } = renderHook(() => useAuth());
    
    expect(result.current.user).toBeNull();
    expect(result.current.loading).toBe(true);
  });
  
  test('logs in user successfully', async () => {
    const mockUser = { id: 1, name: 'John Doe', email: 'john@example.com' };
    (api.auth.login as jest.Mock).mockResolvedValue({ user: mockUser, token: 'mock-token' });
    
    const { result } = renderHook(() => useAuth());
    
    await act(async () => {
      await result.current.login('john@example.com', 'password');
    });
    
    expect(result.current.user).toEqual(mockUser);
    expect(result.current.loading).toBe(false);
  });
});
```

### End-to-End Testing

#### Playwright Tests
```typescript
// e2e/booking-flow.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Booking Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'test@example.com');
    await page.fill('[data-testid="password"]', 'password123');
    await page.click('[data-testid="login-button"]');
    await expect(page).toHaveURL('/destinations');
  });
  
  test('user can complete booking flow', async ({ page }) => {
    // Browse destinations
    await page.goto('/destinations');
    await expect(page.locator('[data-testid="destination-card"]')).toHaveCount.greaterThan(0);
    
    // Select destination
    await page.click('[data-testid="destination-card"]:first-child');
    await expect(page).toHaveURL(/\/destinations\/\d+/);
    
    // Start booking
    await page.click('[data-testid="book-now-button"]');
    await expect(page).toHaveURL(/\/book\/\d+/);
    
    // Fill booking form
    await page.fill('[data-testid="start-date"]', '2024-06-01');
    await page.fill('[data-testid="end-date"]', '2024-06-08');
    await page.selectOption('[data-testid="guests"]', '2');
    await page.fill('[data-testid="special-requests"]', 'Vegetarian meals please');
    
    // Proceed to payment
    await page.click('[data-testid="proceed-to-payment"]');
    await expect(page).toHaveURL(/\/payment\/\d+/);
    
    // Fill payment form (test mode)
    await page.fill('[data-testid="card-number"]', '4242424242424242');
    await page.fill('[data-testid="card-expiry"]', '12/25');
    await page.fill('[data-testid="card-cvc"]', '123');
    
    // Complete booking
    await page.click('[data-testid="complete-booking"]');
    
    // Verify confirmation
    await expect(page).toHaveURL(/\/booking-confirmation\/\d+/);
    await expect(page.locator('[data-testid="confirmation-message"]')).toBeVisible();
    await expect(page.locator('[data-testid="booking-id"]')).toBeVisible();
  });
  
  test('handles payment failure gracefully', async ({ page }) => {
    // Navigate to payment page
    await page.goto('/destinations/1');
    await page.click('[data-testid="book-now-button"]');
    await page.fill('[data-testid="start-date"]', '2024-06-01');
    await page.fill('[data-testid="end-date"]', '2024-06-08');
    await page.click('[data-testid="proceed-to-payment"]');
    
    // Use declined card
    await page.fill('[data-testid="card-number"]', '4000000000000002');
    await page.fill('[data-testid="card-expiry"]', '12/25');
    await page.fill('[data-testid="card-cvc"]', '123');
    
    await page.click('[data-testid="complete-booking"]');
    
    // Verify error handling
    await expect(page.locator('[data-testid="payment-error"]')).toBeVisible();
    await expect(page.locator('[data-testid="payment-error"]')).toContainText('payment failed');
  });
});
```

## Security Testing

### Automated Security Tests
```bash
# Run security test suite
./scripts/security-test.sh

# Run specific security tests
./scripts/security-test.sh --auth-only
./scripts/security-test.sh --input-only
./scripts/security-test.sh --full-scan
```

### Manual Security Testing
```go
// security_test.go
func TestSecurityHeaders(t *testing.T) {
    server := setupTestServer(t)
    defer server.Close()
    
    resp, err := http.Get(server.URL + "/api/destinations")
    require.NoError(t, err)
    defer resp.Body.Close()
    
    // Check security headers
    assert.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"))
    assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
    assert.Equal(t, "1; mode=block", resp.Header.Get("X-XSS-Protection"))
    assert.Contains(t, resp.Header.Get("Content-Security-Policy"), "default-src 'self'")
}

func TestInputValidation(t *testing.T) {
    server := setupTestServer(t)
    defer server.Close()
    
    maliciousInputs := []string{
        "<script>alert('xss')</script>",
        "'; DROP TABLE users; --",
        "../../../etc/passwd",
        "javascript:alert('xss')",
    }
    
    for _, input := range maliciousInputs {
        t.Run("blocks malicious input: "+input, func(t *testing.T) {
            payload := fmt.Sprintf(`{"name":"%s"}`, input)
            req, err := http.NewRequest("POST", server.URL+"/api/destinations", strings.NewReader(payload))
            require.NoError(t, err)
            req.Header.Set("Content-Type", "application/json")
            
            resp, err := http.DefaultClient.Do(req)
            require.NoError(t, err)
            defer resp.Body.Close()
            
            // Should be blocked (400 Bad Request or 422 Unprocessable Entity)
            assert.True(t, resp.StatusCode == 400 || resp.StatusCode == 422)
        })
    }
}
```

## Performance Testing

### Load Testing
```bash
# Run performance tests
./scripts/performance-test.sh

# Run with custom parameters
./scripts/performance-test.sh --concurrent-users 50 --duration 60s
```

### Benchmark Tests
```go
// benchmark_test.go
func BenchmarkDestinationService_GetDestinations(b *testing.B) {
    service := setupBenchmarkService(b)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := service.GetDestinations(context.Background(), &GetDestinationsRequest{
            Page:  1,
            Limit: 10,
        })
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkPasswordHashing(b *testing.B) {
    hasher := security.NewPasswordHasher()
    password := "TestPassword123!"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := hasher.HashPassword(password)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## Test Configuration

### Test Environment Setup
```yaml
# docker-compose.test.yml
version: '3.8'

services:
  postgres-test:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: exotic_travel_test
      POSTGRES_USER: test_user
      POSTGRES_PASSWORD: test_password
    ports:
      - "5433:5432"
    tmpfs:
      - /var/lib/postgresql/data

  redis-test:
    image: redis:7-alpine
    ports:
      - "6380:6379"
    tmpfs:
      - /data
```

### CI/CD Test Pipeline
```yaml
# .github/workflows/test.yml
name: Test Suite

on: [push, pull_request]

jobs:
  backend-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: test_password
          POSTGRES_DB: exotic_travel_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: '1.21'
    
    - name: Run tests
      run: |
        cd backend
        go test -race -coverprofile=coverage.out ./...
        go tool cover -html=coverage.out -o coverage.html
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./backend/coverage.out

  frontend-tests:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-node@v3
      with:
        node-version: '18'
    
    - name: Install dependencies
      run: |
        cd frontend
        npm ci
    
    - name: Run tests
      run: |
        cd frontend
        npm test -- --coverage --watchAll=false
    
    - name: Run E2E tests
      run: |
        cd frontend
        npm run test:e2e

  security-tests:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Run security tests
      run: ./scripts/security-test.sh --full-scan
```

## Test Reporting

### Coverage Reports
```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Coverage threshold check
go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//' | awk '{if($1<80) exit 1}'
```

### Test Results
```bash
# Generate test report
go test -json ./... > test-results.json

# Convert to JUnit format for CI
go-junit-report < test-results.json > test-results.xml
```

This comprehensive testing guide ensures high-quality, secure, and performant code through multiple layers of automated and manual testing.

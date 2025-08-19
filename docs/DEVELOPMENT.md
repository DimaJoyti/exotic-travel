# Development Guide

This guide covers setting up the development environment and contributing to the Exotic Travel Booking Platform.

## Development Environment Setup

### Prerequisites

- **Go 1.21+** ([Download](https://golang.org/dl/))
- **Node.js 18+** ([Download](https://nodejs.org/))
- **PostgreSQL 15+** ([Download](https://postgresql.org/download/))
- **Redis 7+** ([Download](https://redis.io/download))
- **Git** ([Download](https://git-scm.com/downloads))
- **Docker** (optional, [Download](https://docker.com/get-started))

### IDE Recommendations

- **VS Code** with Go and TypeScript extensions
- **GoLand** by JetBrains
- **WebStorm** for frontend development

### Required VS Code Extensions

```json
{
  "recommendations": [
    "golang.go",
    "bradlc.vscode-tailwindcss",
    "ms-vscode.vscode-typescript-next",
    "esbenp.prettier-vscode",
    "ms-vscode.vscode-json",
    "redhat.vscode-yaml",
    "ms-vscode-remote.remote-containers"
  ]
}
```

## Quick Development Setup

### 1. Clone Repository

```bash
git clone https://github.com/your-org/exotic-travel-booking.git
cd exotic-travel-booking
```

### 2. Environment Configuration

```bash
# Backend environment
cd backend
cp .env.example .env

# Frontend environment
cd ../frontend
cp .env.local.example .env.local
```

### 3. Database Setup

#### Option A: Docker (Recommended)
```bash
# Start PostgreSQL and Redis with Docker
docker-compose up -d postgres redis

# Wait for services to be ready
sleep 10
```

#### Option B: Local Installation
```bash
# PostgreSQL setup
sudo -u postgres createdb exotic_travel_dev
sudo -u postgres psql -c "CREATE USER dev_user WITH PASSWORD 'dev_password';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE exotic_travel_dev TO dev_user;"

# Redis setup (usually runs on default port 6379)
redis-server
```

### 4. Backend Setup

```bash
cd backend

# Install dependencies
go mod download

# Run database migrations
go run cmd/migrate/main.go up

# Start the development server
go run cmd/server/main.go
```

The backend will be available at `http://localhost:8080`

### 5. Frontend Setup

```bash
cd frontend

# Install dependencies
npm install

# Start the development server
npm run dev
```

The frontend will be available at `http://localhost:3000`

## Development Workflow

### 1. Feature Development

```bash
# Create feature branch
git checkout -b feature/your-feature-name

# Make changes and commit
git add .
git commit -m "feat: add your feature description"

# Push branch
git push origin feature/your-feature-name

# Create pull request
```

### 2. Code Quality Checks

#### Backend
```bash
cd backend

# Run tests
go test ./...

# Run linter
golangci-lint run

# Format code
go fmt ./...

# Check for security issues
gosec ./...

# Generate test coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### Frontend
```bash
cd frontend

# Run tests
npm test

# Run linter
npm run lint

# Format code
npm run format

# Type checking
npm run type-check

# Build check
npm run build
```

### 3. Database Migrations

#### Creating Migrations
```bash
cd backend

# Create new migration
go run cmd/migrate/main.go create add_new_table

# This creates:
# migrations/YYYYMMDDHHMMSS_add_new_table.up.sql
# migrations/YYYYMMDDHHMMSS_add_new_table.down.sql
```

#### Migration Commands
```bash
# Apply migrations
go run cmd/migrate/main.go up

# Rollback last migration
go run cmd/migrate/main.go down

# Check migration status
go run cmd/migrate/main.go status

# Reset database (development only)
go run cmd/migrate/main.go reset
```

## Project Structure

### Backend Structure
```
backend/
├── cmd/
│   ├── server/          # Main application
│   ├── migrate/         # Database migration tool
│   └── worker/          # Background job worker
├── internal/
│   ├── handlers/        # HTTP request handlers
│   │   ├── auth.go
│   │   ├── destinations.go
│   │   ├── bookings.go
│   │   └── users.go
│   ├── services/        # Business logic
│   │   ├── auth_service.go
│   │   ├── destination_service.go
│   │   └── booking_service.go
│   ├── repositories/    # Data access layer
│   │   ├── user_repository.go
│   │   ├── destination_repository.go
│   │   └── booking_repository.go
│   ├── models/          # Domain models
│   │   ├── user.go
│   │   ├── destination.go
│   │   └── booking.go
│   ├── middleware/      # HTTP middleware
│   │   ├── auth.go
│   │   ├── cors.go
│   │   └── logging.go
│   ├── security/        # Security components
│   ├── cache/           # Caching layer
│   ├── database/        # Database utilities
│   └── config/          # Configuration
├── pkg/                 # Shared packages
├── migrations/          # Database migrations
├── docs/               # API documentation
└── scripts/            # Development scripts
```

### Frontend Structure
```
frontend/
├── src/
│   ├── app/            # Next.js App Router
│   │   ├── (auth)/     # Auth route group
│   │   ├── admin/      # Admin pages
│   │   ├── destinations/ # Destination pages
│   │   └── layout.tsx  # Root layout
│   ├── components/     # Reusable components
│   │   ├── ui/         # Base UI components
│   │   ├── forms/      # Form components
│   │   └── layout/     # Layout components
│   ├── hooks/          # Custom React hooks
│   │   ├── use-auth.ts
│   │   ├── use-api.ts
│   │   └── use-performance.ts
│   ├── lib/            # Utility functions
│   │   ├── api.ts      # API client
│   │   ├── auth.ts     # Auth utilities
│   │   └── utils.ts    # General utilities
│   ├── stores/         # State management
│   │   ├── auth-store.ts
│   │   └── booking-store.ts
│   ├── types/          # TypeScript definitions
│   └── styles/         # Global styles
├── public/             # Static assets
└── docs/              # Component documentation
```

## Coding Standards

### Go Coding Standards

#### Naming Conventions
```go
// Package names: lowercase, single word
package handlers

// Interface names: noun + "er" suffix
type UserRepository interface {
    GetUser(id int64) (*User, error)
}

// Struct names: PascalCase
type UserService struct {
    repo UserRepository
}

// Function names: PascalCase for exported, camelCase for private
func (s *UserService) CreateUser(user *User) error {
    return s.validateUser(user)
}

func (s *UserService) validateUser(user *User) error {
    // validation logic
}
```

#### Error Handling
```go
// Always handle errors explicitly
user, err := userRepo.GetUser(id)
if err != nil {
    return fmt.Errorf("failed to get user: %w", err)
}

// Use custom error types when appropriate
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation error in %s: %s", e.Field, e.Message)
}
```

#### Testing
```go
// Table-driven tests
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        user    *User
        wantErr bool
    }{
        {
            name: "valid user",
            user: &User{Name: "John", Email: "john@example.com"},
            wantErr: false,
        },
        {
            name: "invalid email",
            user: &User{Name: "John", Email: "invalid"},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            service := NewUserService(mockRepo)
            err := service.CreateUser(tt.user)
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### TypeScript/React Coding Standards

#### Component Structure
```tsx
// Use functional components with TypeScript
interface UserProfileProps {
  user: User;
  onUpdate: (user: User) => void;
}

export function UserProfile({ user, onUpdate }: UserProfileProps) {
  const [isEditing, setIsEditing] = useState(false);

  const handleSubmit = useCallback((data: UserFormData) => {
    onUpdate({ ...user, ...data });
    setIsEditing(false);
  }, [user, onUpdate]);

  return (
    <div className="user-profile">
      {/* Component JSX */}
    </div>
  );
}
```

#### Custom Hooks
```tsx
// Custom hooks for reusable logic
export function useAuth() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Auth logic
  }, []);

  const login = useCallback(async (credentials: LoginCredentials) => {
    // Login logic
  }, []);

  return { user, loading, login };
}
```

#### API Integration
```tsx
// Use React Query for server state
export function useDestinations(filters?: DestinationFilters) {
  return useQuery({
    queryKey: ['destinations', filters],
    queryFn: () => api.destinations.getAll(filters),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}
```

## Testing Strategy

### Backend Testing

#### Unit Tests
```go
// Test business logic in isolation
func TestBookingService_CreateBooking(t *testing.T) {
    mockRepo := &MockBookingRepository{}
    service := NewBookingService(mockRepo)
    
    booking := &Booking{
        UserID: 1,
        DestinationID: 1,
        StartDate: time.Now().AddDate(0, 0, 7),
        EndDate: time.Now().AddDate(0, 0, 14),
    }
    
    err := service.CreateBooking(booking)
    assert.NoError(t, err)
    assert.True(t, mockRepo.CreateCalled)
}
```

#### Integration Tests
```go
// Test with real database
func TestBookingRepository_Integration(t *testing.T) {
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    repo := NewBookingRepository(db)
    
    booking := &Booking{/* test data */}
    err := repo.Create(booking)
    assert.NoError(t, err)
    assert.NotZero(t, booking.ID)
}
```

### Frontend Testing

#### Component Tests
```tsx
// Test component behavior
import { render, screen, fireEvent } from '@testing-library/react';
import { UserProfile } from './UserProfile';

test('renders user information', () => {
  const user = { id: 1, name: 'John Doe', email: 'john@example.com' };
  render(<UserProfile user={user} onUpdate={jest.fn()} />);
  
  expect(screen.getByText('John Doe')).toBeInTheDocument();
  expect(screen.getByText('john@example.com')).toBeInTheDocument();
});

test('calls onUpdate when form is submitted', () => {
  const user = { id: 1, name: 'John Doe', email: 'john@example.com' };
  const onUpdate = jest.fn();
  
  render(<UserProfile user={user} onUpdate={onUpdate} />);
  
  fireEvent.click(screen.getByText('Edit'));
  fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Jane Doe' } });
  fireEvent.click(screen.getByText('Save'));
  
  expect(onUpdate).toHaveBeenCalledWith({ ...user, name: 'Jane Doe' });
});
```

#### E2E Tests
```tsx
// Test complete user flows
import { test, expect } from '@playwright/test';

test('user can book a destination', async ({ page }) => {
  await page.goto('/destinations');
  await page.click('[data-testid="destination-card"]:first-child');
  await page.click('[data-testid="book-now-button"]');
  
  await page.fill('[data-testid="start-date"]', '2024-06-01');
  await page.fill('[data-testid="end-date"]', '2024-06-08');
  await page.fill('[data-testid="guests"]', '2');
  
  await page.click('[data-testid="confirm-booking"]');
  
  await expect(page.locator('[data-testid="booking-confirmation"]')).toBeVisible();
});
```

## Debugging

### Backend Debugging

#### Using Delve Debugger
```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug application
dlv debug cmd/server/main.go

# Debug tests
dlv test ./internal/services
```

#### Logging
```go
// Use structured logging
import "log/slog"

logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
logger.Info("User created", "user_id", user.ID, "email", user.Email)
logger.Error("Database error", "error", err, "query", query)
```

### Frontend Debugging

#### Browser DevTools
- Use React Developer Tools extension
- Use Redux DevTools for state management
- Use Network tab for API debugging
- Use Performance tab for performance analysis

#### Debug Configuration
```json
// .vscode/launch.json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Next.js",
      "type": "node",
      "request": "launch",
      "program": "${workspaceFolder}/frontend/node_modules/.bin/next",
      "args": ["dev"],
      "cwd": "${workspaceFolder}/frontend",
      "runtimeArgs": ["--inspect"]
    }
  ]
}
```

## Performance Optimization

### Backend Performance
- Use database connection pooling
- Implement caching with Redis
- Use prepared statements
- Profile with `go tool pprof`
- Monitor with Prometheus metrics

### Frontend Performance
- Use Next.js Image optimization
- Implement code splitting
- Use React.memo for expensive components
- Optimize bundle size with webpack-bundle-analyzer
- Use performance monitoring tools

## Contributing Guidelines

### Pull Request Process

1. **Fork the repository**
2. **Create a feature branch** from `main`
3. **Make your changes** following coding standards
4. **Add tests** for new functionality
5. **Update documentation** if needed
6. **Run all tests** and ensure they pass
7. **Submit a pull request** with clear description

### Commit Message Format

```
type(scope): description

[optional body]

[optional footer]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Examples:
```
feat(auth): add JWT token refresh functionality
fix(booking): resolve date validation issue
docs(api): update authentication endpoints
```

### Code Review Checklist

- [ ] Code follows project standards
- [ ] Tests are included and passing
- [ ] Documentation is updated
- [ ] No security vulnerabilities
- [ ] Performance impact considered
- [ ] Backward compatibility maintained

This development guide provides a comprehensive foundation for contributing to the Exotic Travel Booking Platform with proper development practices and quality standards.

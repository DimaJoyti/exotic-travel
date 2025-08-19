# 🌴 Exotic Travel Booking Platform

A modern, enterprise-grade travel booking platform built with Go and Next.js, featuring comprehensive security, performance optimization, and scalability.

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Node.js Version](https://img.shields.io/badge/Node.js-18+-green.svg)](https://nodejs.org)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Security](https://img.shields.io/badge/Security-Hardened-red.svg)](SECURITY_HARDENING.md)
[![Performance](https://img.shields.io/badge/Performance-Optimized-orange.svg)](PERFORMANCE_OPTIMIZATION.md)

## 🚀 Features

### 🔐 **Security & Authentication**
- **Enterprise-grade JWT authentication** with RSA-256 signing and automatic key rotation
- **Role-based access control (RBAC)** with granular permissions
- **Advanced input validation** with SQL injection, XSS, and path traversal protection
- **Comprehensive audit logging** with real-time threat detection
- **Rate limiting and DDoS protection** with Redis-backed implementation
- **Security headers and CSRF protection** following OWASP guidelines

### 🏖️ **Core Travel Features**
- **Destination management** with advanced search, filtering, and categorization
- **Intelligent booking system** with availability checking and conflict resolution
- **Secure payment processing** with Stripe integration and PCI compliance
- **Review and rating system** with moderation and spam protection
- **Multi-language support** with internationalization (i18n)
- **Mobile-responsive design** with progressive web app (PWA) capabilities

### 📊 **Admin & Analytics**
- **Comprehensive admin dashboard** with real-time analytics and reporting
- **User management** with role assignment and activity monitoring
- **Booking analytics** with revenue tracking and performance metrics
- **Content management** for destinations, images, and promotional content
- **System monitoring** with health checks and performance metrics

### ⚡ **Performance & Scalability**
- **Database optimization** with connection pooling and query optimization
- **Redis caching** with intelligent cache invalidation strategies
- **CDN integration** for static asset delivery and image optimization
- **Horizontal scaling** support with load balancer compatibility
- **Performance monitoring** with OpenTelemetry and Prometheus metrics

## 🏗️ Architecture

### Backend Architecture (Clean Architecture)
```
backend/
├── cmd/server/           # Application entry point
├── internal/
│   ├── handlers/         # HTTP handlers (controllers)
│   ├── services/         # Business logic (use cases)
│   ├── repositories/     # Data access layer
│   ├── models/          # Domain models
│   ├── middleware/      # HTTP middleware
│   ├── security/        # Security components
│   ├── cache/           # Caching layer
│   ├── database/        # Database utilities
│   └── metrics/         # Performance monitoring
├── pkg/                 # Shared packages
├── migrations/          # Database migrations
└── docs/               # API documentation
```

### Frontend Architecture (Next.js App Router)
```
frontend/
├── src/
│   ├── app/             # App Router pages and layouts
│   ├── components/      # Reusable UI components
│   ├── hooks/           # Custom React hooks
│   ├── lib/             # Utility functions and configurations
│   ├── stores/          # State management (Zustand)
│   ├── types/           # TypeScript type definitions
│   └── styles/          # Global styles and Tailwind config
├── public/              # Static assets
└── docs/               # Component documentation
```

## 🛠️ Tech Stack

### **Backend Technologies**
- **Go 1.21+** - High-performance backend with excellent concurrency
- **PostgreSQL 15+** - Robust relational database with advanced features
- **Redis 7+** - In-memory caching and session storage
- **JWT with RSA-256** - Secure authentication with automatic key rotation
- **OpenTelemetry** - Distributed tracing and observability
- **Docker & Docker Compose** - Containerization and orchestration

### **Frontend Technologies**
- **Next.js 14** - React framework with App Router and server components
- **TypeScript** - Type-safe JavaScript with enhanced developer experience
- **Tailwind CSS** - Utility-first CSS framework with custom design system
- **React Hook Form** - Performant forms with built-in validation
- **Zustand** - Lightweight state management solution
- **React Query** - Server state management and caching

### **DevOps & Monitoring**
- **Prometheus** - Metrics collection and monitoring
- **Grafana** - Visualization and alerting dashboards
- **Jaeger** - Distributed tracing and performance analysis
- **GitHub Actions** - CI/CD pipelines and automated testing
- **Docker Swarm/Kubernetes** - Container orchestration and scaling

## 🚀 Quick Start

### Prerequisites
- **Go 1.21+** ([Download](https://golang.org/dl/))
- **Node.js 18+** ([Download](https://nodejs.org/))
- **PostgreSQL 15+** ([Download](https://postgresql.org/download/))
- **Redis 7+** ([Download](https://redis.io/download))
- **Docker** (optional, [Download](https://docker.com/get-started))

### 🐳 Docker Setup (Recommended)
```bash
# Clone the repository
git clone https://github.com/your-org/exotic-travel-booking.git
cd exotic-travel-booking

# Copy environment configuration
cp .env.example .env

# Start all services with Docker Compose
docker-compose up -d

# The application will be available at:
# Frontend: http://localhost:3000
# Backend API: http://localhost:8080
# Admin Dashboard: http://localhost:3000/admin
```

### 🔧 Manual Setup

#### Backend Setup
```bash
cd backend

# Install dependencies
go mod download

# Set up environment variables
cp .env.example .env
# Edit .env with your database and Redis configurations

# Run database migrations
go run cmd/migrate/main.go up

# Start the server
go run cmd/server/main.go
```

#### Frontend Setup
```bash
cd frontend

# Install dependencies
npm install

# Set up environment variables
cp .env.local.example .env.local
# Edit .env.local with your API endpoints

# Start the development server
npm run dev
```

If you prefer to set up manually:

1. **Backend Setup:**
   ```bash
   cd backend
   cp .env.example .env
   go mod download
   ```

2. **Frontend Setup:**
   ```bash
   cd frontend
   cp .env.example .env.local
   npm install
   ```

3. **Database Setup:**
   ```bash
   docker compose up postgres redis -d
   ```

4. **Start Services:**
   ```bash
   # Terminal 1 - Backend
   cd backend && go run cmd/server/main.go

   # Terminal 2 - Frontend
   cd frontend && npm run dev
   ```

## 📚 API Documentation

The API follows RESTful principles and includes endpoints for:
- Authentication (`/api/auth/*`)
- Destinations (`/api/destinations/*`)
- Bookings (`/api/bookings/*`)
- Users (`/api/users/*`)
- Reviews (`/api/reviews/*`)
- Payments (`/api/payments/*`)

## 🧪 Testing

### Run All Tests
```bash
./scripts/test.sh
```

### Individual Test Commands

**Backend:**
```bash
cd backend
make test              # Run tests
make test-coverage     # Run tests with coverage
make lint              # Run linting
make fmt               # Format code
```

**Frontend:**
```bash
cd frontend
npm run type-check     # TypeScript type checking
npm run lint           # ESLint
npm run lint:fix       # Fix ESLint issues
npm run format         # Format with Prettier
npm run format:check   # Check formatting
```

## 🔧 Development Commands

### Backend (Go)
```bash
cd backend
make build             # Build the application
make run               # Build and run
make dev               # Run with hot reload (requires air)
make clean             # Clean build artifacts
make deps              # Download dependencies
```

### Frontend (Next.js)
```bash
cd frontend
npm run dev            # Development server
npm run build          # Production build
npm run start          # Start production server
npm run clean          # Clean build artifacts
```

### Docker Commands
```bash
docker compose up -d                    # Start all services
docker compose up postgres redis -d    # Start only database services
docker compose down                     # Stop all services
docker compose logs <service>          # View service logs
```

## 🚀 Deployment

The application is containerized and ready for deployment with Docker.

## 📄 License

MIT License

#!/bin/bash

# Exotic Travel Booking - Initial Setup Script

set -e

echo "🏗️  Setting up Exotic Travel Booking Development Environment"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
print_status "Checking prerequisites..."

# Check Go
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go 1.22+ and try again."
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
print_success "Go $GO_VERSION is installed"

# Check Node.js
if ! command -v node &> /dev/null; then
    print_error "Node.js is not installed. Please install Node.js 18+ and try again."
    exit 1
fi

NODE_VERSION=$(node --version)
print_success "Node.js $NODE_VERSION is installed"

# Check Docker
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed. Please install Docker and try again."
    exit 1
fi

print_success "Docker is installed"

# Setup backend
print_status "Setting up backend..."
cd backend

# Copy environment file
if [ ! -f ".env" ]; then
    cp .env.example .env
    print_success "Created backend .env file"
else
    print_warning "Backend .env file already exists"
fi

# Download Go dependencies
print_status "Downloading Go dependencies..."
go mod download
go mod tidy
print_success "Go dependencies installed"

cd ..

# Setup frontend
print_status "Setting up frontend..."
cd frontend

# Copy environment file
if [ ! -f ".env.local" ]; then
    cp .env.example .env.local
    print_success "Created frontend .env.local file"
else
    print_warning "Frontend .env.local file already exists"
fi

# Install Node.js dependencies
print_status "Installing Node.js dependencies..."
npm install
print_success "Node.js dependencies installed"

cd ..

# Setup database
print_status "Setting up database..."
docker compose up postgres redis -d

# Wait for database
print_status "Waiting for database to initialize..."
sleep 10

# Check database connection
if docker compose exec postgres pg_isready -U postgres > /dev/null 2>&1; then
    print_success "Database is ready and migrations have been applied!"
else
    print_warning "Database might still be initializing. Check with 'docker compose logs postgres'"
fi

print_success "🎉 Setup complete!"
echo ""
echo "📍 Next steps:"
echo "   1. Review and update .env files with your configuration"
echo "   2. Run './scripts/dev.sh' to start the development environment"
echo "   3. Visit http://localhost:3000 to see the application"
echo ""
echo "📚 Useful commands:"
echo "   ./scripts/dev.sh     - Start development environment"
echo "   ./scripts/test.sh    - Run all tests"
echo "   docker compose down  - Stop all services"

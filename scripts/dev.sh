#!/bin/bash

# Exotic Travel Booking - Development Setup Script

set -e

echo "🚀 Starting Exotic Travel Booking Development Environment"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
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

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker and try again."
    exit 1
fi

print_status "Starting database services..."
docker compose up postgres redis -d

# Wait for database to be ready
print_status "Waiting for database to be ready..."
sleep 5

# Check if database is accessible
if docker compose exec postgres pg_isready -U postgres > /dev/null 2>&1; then
    print_success "Database is ready!"
else
    print_error "Database is not ready. Please check Docker logs."
    exit 1
fi

# Start backend in development mode
print_status "Starting backend server..."
cd backend
if [ ! -f ".env" ]; then
    print_warning "No .env file found. Copying from .env.example"
    cp .env.example .env
fi

# Start backend in background
go run cmd/server/main.go &
BACKEND_PID=$!
cd ..

print_success "Backend started with PID: $BACKEND_PID"

# Start frontend in development mode
print_status "Starting frontend server..."
cd frontend
if [ ! -f ".env.local" ]; then
    print_warning "No .env.local file found. Copying from .env.example"
    cp .env.example .env.local
fi

# Install dependencies if node_modules doesn't exist
if [ ! -d "node_modules" ]; then
    print_status "Installing frontend dependencies..."
    npm install
fi

# Start frontend
npm run dev &
FRONTEND_PID=$!
cd ..

print_success "Frontend started with PID: $FRONTEND_PID"

print_success "🎉 Development environment is ready!"
echo ""
echo "📍 Services:"
echo "   Frontend: http://localhost:3000"
echo "   Backend:  http://localhost:8080"
echo "   Database: localhost:5432"
echo "   Redis:    localhost:6379"
echo ""
echo "🛑 To stop all services, press Ctrl+C"

# Function to cleanup on exit
cleanup() {
    print_status "Shutting down services..."
    kill $BACKEND_PID 2>/dev/null || true
    kill $FRONTEND_PID 2>/dev/null || true
    docker compose down
    print_success "All services stopped."
}

# Set trap to cleanup on script exit
trap cleanup EXIT

# Wait for user to stop
wait

#!/bin/bash

# Exotic Travel Booking - Test Runner Script

set -e

echo "🧪 Running Exotic Travel Booking Tests"

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

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Test backend
print_status "Running backend tests..."
cd backend

# Run Go tests
if go test -v ./...; then
    print_success "Backend tests passed!"
else
    print_error "Backend tests failed!"
    exit 1
fi

# Run linting
if command -v golangci-lint &> /dev/null; then
    print_status "Running backend linting..."
    if golangci-lint run; then
        print_success "Backend linting passed!"
    else
        print_error "Backend linting failed!"
        exit 1
    fi
else
    print_status "golangci-lint not found, skipping backend linting"
fi

cd ..

# Test frontend
print_status "Running frontend tests..."
cd frontend

# Run type checking
print_status "Running TypeScript type checking..."
if npm run type-check; then
    print_success "TypeScript type checking passed!"
else
    print_error "TypeScript type checking failed!"
    exit 1
fi

# Run linting
print_status "Running frontend linting..."
if npm run lint; then
    print_success "Frontend linting passed!"
else
    print_error "Frontend linting failed!"
    exit 1
fi

# Check formatting
print_status "Checking code formatting..."
if npm run format:check; then
    print_success "Code formatting is correct!"
else
    print_error "Code formatting issues found! Run 'npm run format' to fix."
    exit 1
fi

# Build test
print_status "Testing frontend build..."
if npm run build; then
    print_success "Frontend build successful!"
else
    print_error "Frontend build failed!"
    exit 1
fi

cd ..

print_success "🎉 All tests passed!"

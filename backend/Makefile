# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Build parameters
BINARY_NAME=server
BINARY_UNIX=$(BINARY_NAME)_unix
MAIN_PATH=./cmd/server

# LLM Server parameters
LLM_BINARY_NAME=llm-server
LLM_BINARY_UNIX=$(LLM_BINARY_NAME)_unix
LLM_MAIN_PATH=./cmd/llm-server
LLM_PORT=8081

# Simple Server parameters
SIMPLE_BINARY_NAME=simple-server
SIMPLE_BINARY_UNIX=$(SIMPLE_BINARY_NAME)_unix
SIMPLE_MAIN_PATH=./cmd/simple-server
SIMPLE_PORT=8081

# Default target
.PHONY: all
all: test build

# Build the application
.PHONY: build
build:
	$(GOBUILD) -o bin/$(BINARY_NAME) -v $(MAIN_PATH)

# Build for Linux
.PHONY: build-linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_UNIX) -v $(MAIN_PATH)

# Run tests
.PHONY: test
test:
	$(GOTEST) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)
	rm -f bin/$(BINARY_UNIX)
	rm -f coverage.out coverage.html

# Run the application
.PHONY: run
run:
	$(GOBUILD) -o bin/$(BINARY_NAME) -v $(MAIN_PATH) && ./bin/$(BINARY_NAME)

# Download dependencies
.PHONY: deps
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
.PHONY: fmt
fmt:
	$(GOFMT) -s -w .

# Lint code
.PHONY: lint
lint:
	$(GOLINT) run

# Fix linting issues
.PHONY: lint-fix
lint-fix:
	$(GOLINT) run --fix

# Run development server with hot reload
.PHONY: dev
dev:
	air

# Database migrations
.PHONY: migrate-up
migrate-up:
	docker compose exec postgres psql -U postgres -d exotic_travel -f /docker-entrypoint-initdb.d/001_create_users_table.sql

.PHONY: migrate-down
migrate-down:
	docker compose exec postgres psql -U postgres -d exotic_travel -c "DROP TABLE IF EXISTS payments, reviews, bookings, destinations, users CASCADE;"

# LLM Server targets
.PHONY: build-llm
build-llm:
	$(GOBUILD) -o bin/$(LLM_BINARY_NAME) -v $(LLM_MAIN_PATH)

.PHONY: build-llm-linux
build-llm-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(LLM_BINARY_UNIX) -v $(LLM_MAIN_PATH)

.PHONY: run-llm
run-llm:
	$(GOBUILD) -o bin/$(LLM_BINARY_NAME) -v $(LLM_MAIN_PATH) && ./bin/$(LLM_BINARY_NAME) -port $(LLM_PORT)

.PHONY: dev-llm
dev-llm:
	$(GOCMD) run $(LLM_MAIN_PATH) -port $(LLM_PORT)

# Simple Server targets
.PHONY: build-simple
build-simple:
	$(GOBUILD) -o bin/$(SIMPLE_BINARY_NAME) -v $(SIMPLE_MAIN_PATH)

.PHONY: build-simple-linux
build-simple-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(SIMPLE_BINARY_UNIX) -v $(SIMPLE_MAIN_PATH)

.PHONY: run-simple
run-simple:
	$(GOBUILD) -o bin/$(SIMPLE_BINARY_NAME) -v $(SIMPLE_MAIN_PATH) && ./bin/$(SIMPLE_BINARY_NAME) -port $(SIMPLE_PORT)

.PHONY: dev-simple
dev-simple:
	$(GOCMD) run $(SIMPLE_MAIN_PATH) -port $(SIMPLE_PORT)

# Docker commands
.PHONY: docker-build
docker-build:
	docker build -t exotic-travel-backend .

.PHONY: docker-build-llm
docker-build-llm:
	docker build -f Dockerfile.llm -t exotic-travel/llm-backend:latest .

.PHONY: docker-run
docker-run:
	docker compose up backend

.PHONY: docker-run-llm
docker-run-llm:
	docker compose -f docker-compose.llm.yml up -d

.PHONY: docker-stop-llm
docker-stop-llm:
	docker compose -f docker-compose.llm.yml down

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  build-linux   - Build for Linux"
	@echo "  build-llm     - Build the LLM server"
	@echo "  build-llm-linux - Build LLM server for Linux"
	@echo "  build-simple  - Build the simple server"
	@echo "  build-simple-linux - Build simple server for Linux"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  clean         - Clean build artifacts"
	@echo "  run           - Build and run the application"
	@echo "  run-llm       - Build and run the LLM server"
	@echo "  run-simple    - Build and run the simple server"
	@echo "  dev           - Run development server with hot reload"
	@echo "  dev-llm       - Run LLM server in development mode"
	@echo "  dev-simple    - Run simple server in development mode"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  lint-fix      - Fix linting issues"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-build-llm - Build LLM Docker image"
	@echo "  docker-run    - Run with Docker Compose"
	@echo "  docker-run-llm - Run LLM services with Docker Compose"
	@echo "  docker-stop-llm - Stop LLM services"
	@echo "  help          - Show this help message"

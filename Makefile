# Go Chat Widget Dashboard Makefile

.PHONY: build run test clean install dev docker help

# Default target
help:
	@echo "Available commands:"
	@echo "  build     - Build the application"
	@echo "  run       - Run the application"
	@echo "  test      - Run tests"
	@echo "  clean     - Clean build artifacts"
	@echo "  install   - Install dependencies"
	@echo "  dev       - Run in development mode with hot reload"
	@echo "  docker    - Build and run with Docker"
	@echo "  help      - Show this help message"

# Build the application
build:
	@echo "Building the application..."
	go build -o bin/chat-widget-server ./cmd/server

# Run the application
run: build
	@echo "Starting the server..."
	./bin/chat-widget-server

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -f db/chat_widget.db

# Install dependencies
install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Development mode (requires air for hot reload)
dev:
	@echo "Starting development server with hot reload..."
	@if command -v air >/dev/null 2>&1; then \
		air -c .air.toml; \
	else \
		echo "Air not installed. Installing..."; \
		go install github.com/cosmtrek/air@latest; \
		air -c .air.toml; \
	fi

# Initialize the database with default admin user
init-db:
	@echo "Initializing database..."
	go run scripts/init-db.go

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Security scan
security:
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Docker build and run
docker:
	@echo "Building Docker image..."
	docker build -t chat-widget-dashboard .
	@echo "Running Docker container..."
	docker run -p 8080:8080 -v $(PWD)/uploads:/app/uploads chat-widget-dashboard

# Create directory structure
setup:
	@echo "Setting up project structure..."
	mkdir -p uploads/models
	mkdir -p bin
	mkdir -p logs

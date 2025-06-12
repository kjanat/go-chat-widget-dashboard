# Go Chat Widget Dashboard Makefile

.PHONY: help install build build-css build-js build-templ build-go dev clean test lint format run start stop logs setup deps

# Default target
help: ## Show this help message
	@echo "Go Chat Widget Dashboard - Available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Setup and Installation
setup: ## Set up the development environment
	@echo "Setting up development environment..."
	go mod tidy
	npm install
	mkdir -p dist bin tmp
	@echo "Development environment ready!"

install: setup ## Install all dependencies
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/air-verse/air@latest

deps: ## Download Go dependencies
	go mod download
	go mod tidy

# Build commands
build-templ: ## Generate Templ templates
	@echo "Generating Templ templates..."
	templ generate

build-css: ## Build Tailwind CSS
	@echo "Building CSS..."
	npx tailwindcss -i ./assets/css/main.css -o ./dist/main.css --minify

build-js: ## Copy JavaScript files
	@echo "Building JavaScript..."
	cp ./assets/js/main.js ./dist/main.js

build-go: build-templ ## Build Go application
	@echo "Building Go application..."
	go build -o bin/server .

build: build-css build-js build-templ build-go ## Build all assets and application

# Development commands
dev: ## Start development server with hot reloading
	@echo "Starting development server with hot reloading..."
	npm run dev

run: build ## Build and run the application
	@echo "Running application..."
	./bin/server

start: ## Start the application (alias for run)
	@$(MAKE) run

# Watch commands (for individual components)
watch-css: ## Watch and rebuild CSS
	npx tailwindcss -i ./assets/css/main.css -o ./dist/main.css --watch

watch-js: ## Watch and rebuild JavaScript
	npm run build:js:watch

watch-templ: ## Watch and rebuild Templ templates
	templ generate --watch

watch-go: ## Watch and rebuild Go application
	air

# Utility commands
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf dist/* bin/* tmp/*
	find . -name '*_templ.go' -delete

format: ## Format Go code and templates
	@echo "Formatting code..."
	go fmt ./...
	templ fmt .

lint: ## Lint Go code
	@echo "Linting Go code..."
	golangci-lint run

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Docker commands
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t go-chat-widget-dashboard .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 3000:3000 go-chat-widget-dashboard

# Database commands
db-reset: ## Reset database
	@echo "Resetting database..."
	rm -f db/*.db
	@echo "Database reset complete"

# Production commands
build-prod: ## Build for production
	@echo "Building for production..."
	NODE_ENV=production npm run build:css
	npm run build:js
	templ generate
	CGO_ENABLED=1 go build -ldflags="-w -s" -o bin/server .

# Health check
health: ## Check application health
	@echo "Checking application health..."
	@curl -s http://localhost:3000/health || echo "Application not running"

# Logs
logs: ## Show application logs
	@echo "Showing recent logs..."
	@tail -f tmp/app.log 2>/dev/null || echo "No log file found"

# Stop processes
stop: ## Stop all development processes
	@echo "Stopping development processes..."
	@pkill -f "air" || true
	@pkill -f "tailwindcss" || true
	@pkill -f "templ" || true
	@pkill -f "chokidar" || true
	@echo "Development processes stopped"

# Git hooks
pre-commit: format lint test ## Run pre-commit checks
	@echo "Pre-commit checks completed"

# Information
info: ## Show project information
	@echo "Go Chat Widget Dashboard"
	@echo "========================"
	@echo "Go version: $$(go version)"
	@echo "Node version: $$(node --version)"
	@echo "NPM version: $$(npm --version)"
	@echo "Templ version: $$(templ version 2>/dev/null || echo 'Not installed')"
	@echo "Air version: $$(air -v 2>/dev/null || echo 'Not installed')"
	@echo ""
	@echo "Project structure:"
	@echo "  components/ - Templ components"
	@echo "  views/      - Templ views"
	@echo "  assets/     - Source CSS and JS"
	@echo "  dist/       - Built assets"
	@echo "  internal/   - Go application code"

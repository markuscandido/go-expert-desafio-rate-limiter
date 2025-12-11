.PHONY: help build run test test-coverage docker-build docker-up docker-down clean lint fmt

help:
	@echo "Rate Limiter - Available Commands"
	@echo ""
	@echo "Build & Run:"
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the application"
	@echo "  make dev            - Run with hot reload (requires air)"
	@echo ""
	@echo "Testing:"
	@echo "  make test           - Run all tests"
	@echo "  make test-v         - Run tests with verbose output"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make test-one TEST=TestName - Run specific test"
	@echo ""
	@echo "Code Quality:"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Run linter"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-up      - Start containers with docker-compose"
	@echo "  make docker-down    - Stop containers"
	@echo "  make docker-logs    - View container logs"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make mod-tidy       - Tidy go modules"
	@echo ""

# Build & Run
build:
	@echo "Building application..."
	go build -o rate-limiter .
	@echo "Build complete!"

run: build
	@echo "Starting rate limiter..."
	./rate-limiter

dev:
	@echo "Starting with hot reload (requires 'air')..."
	air

# Testing
test:
	@echo "Running tests..."
	go test ./...

test-v:
	@echo "Running tests with verbose output..."
	go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	go test -cover ./...
	go test -coverprofile=coverage.out ./...
	@echo "Coverage report generated: coverage.out"
	@echo "View HTML report: go tool cover -html=coverage.out"

test-one:
	@echo "Running test: $(TEST)"
	go test -run $(TEST) -v ./...

# Code Quality
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Format complete!"

lint:
	@echo "Running linter (requires 'golangci-lint')..."
	golangci-lint run

# Docker
docker-build:
	@echo "Building Docker image..."
	docker build -t rate-limiter:latest .
	@echo "Build complete!"

docker-up: docker-build
	@echo "Starting containers..."
	docker-compose up -d
	@echo "Containers started! Server on http://localhost:8080"

docker-down:
	@echo "Stopping containers..."
	docker-compose down
	@echo "Containers stopped!"

docker-logs:
	docker-compose logs -f rate-limiter

# Cleanup
clean:
	@echo "Cleaning build artifacts..."
	rm -f rate-limiter
	rm -f coverage.out
	go clean
	@echo "Cleanup complete!"

mod-tidy:
	@echo "Tidying go modules..."
	go mod tidy
	@echo "Modules tidied!"

# Combined targets
all: clean test build

setup:
	@echo "Setting up development environment..."
	go mod download
	@echo "Setup complete! Run 'make docker-up' to start services"

ci: test lint
	@echo "CI checks passed!"

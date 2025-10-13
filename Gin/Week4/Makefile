# News Aggregator Makefile

.PHONY: help build run clean test docker-up docker-down logs

# Default target
help:
	@echo "Available commands:"
	@echo "  build        - Build all services"
	@echo "  run          - Run services locally"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  docker-up    - Start all services with Docker Compose"
	@echo "  docker-down  - Stop all services"
	@echo "  logs         - Show logs from all services"
	@echo "  api-logs     - Show logs from API service"
	@echo "  collector-logs - Show logs from Collector service"

# Build all services
build:
	@echo "Building News API service..."
	go build -o bin/api ./cmd/api
	@echo "Building News Collector service..."
	go build -o bin/collector ./cmd/collector
	@echo "Build completed!"

# Run services locally (requires Redis and Kafka to be running)
run: build
	@echo "Starting services locally..."
	@echo "Make sure Redis and Kafka are running first!"
	@echo "Run 'make docker-up' to start infrastructure services"
	@echo ""
	@echo "Starting News Collector on port 8081..."
	./bin/collector &
	@echo "Starting News API on port 8080..."
	./bin/api

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Docker commands
docker-up:
	@echo "Starting all services with Docker Compose..."
	docker-compose up -d
	@echo "Services started!"
	@echo "News API: http://localhost:8080"
	@echo "News Collector: http://localhost:8081"
	@echo "Kafka UI: http://localhost:8082"

docker-down:
	@echo "Stopping all services..."
	docker-compose down

# Logs
logs:
	docker-compose logs -f

api-logs:
	docker-compose logs -f news-api

collector-logs:
	docker-compose logs -f news-collector

# Development helpers
dev-setup:
	@echo "Setting up development environment..."
	go mod download
	@echo "Development setup completed!"

# Quick test
quick-test:
	@echo "Testing API endpoints..."
	@echo "Health check:"
	curl -s http://localhost:8080/health | jq .
	@echo ""
	@echo "Getting latest news:"
	curl -s "http://localhost:8080/api/v1/news?limit=3" | jq .
	@echo ""
	@echo "Manual news collection:"
	curl -s -X POST http://localhost:8081/collect | jq .

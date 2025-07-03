# Terratag Makefile

.PHONY: help build-api build-ui build dev dev-api dev-ui clean test lint

# Default target
help:
	@echo "Available targets:"
	@echo "  build     - Build both API and UI"
	@echo "  build-api - Build the Go API server"
	@echo "  build-ui  - Build the React UI"
	@echo "  dev       - Run development servers (API + UI)"
	@echo "  dev-api   - Run API development server"
	@echo "  dev-ui    - Run UI development server"
	@echo "  test      - Run tests"
	@echo "  lint      - Run linters"
	@echo "  clean     - Clean build artifacts"

# Build targets
build: build-api build-ui

build-api:
	@echo "Building API server..."
	go build -o bin/terratag-api cmd/api/main.go
	go build -o bin/terratag cmd/terratag/main.go

build-ui:
	@echo "Building UI..."
	cd web/ui && npm run build

# Development targets
dev:
	@echo "Starting development servers..."
	@echo "API will be available at http://localhost:8080"
	@echo "UI will be available at http://localhost:3000"
	@make -j2 dev-api dev-ui

dev-api:
	@echo "Starting API development server..."
	DB_PATH=./data/terratag.db PORT=8080 go run cmd/api/main.go

dev-ui:
	@echo "Starting UI development server..."
	cd web/ui && REACT_APP_API_URL=http://localhost:8080/api/v1 npm start

# Test targets
test:
	@echo "Running Go tests..."
	go test -v ./...
	@echo "Running UI tests..."
	cd web/ui && npm test -- --coverage --watchAll=false

# Lint targets
lint:
	@echo "Running Go lint..."
	golangci-lint run ./... || echo "golangci-lint not installed, skipping"
	@echo "Running UI lint..."
	cd web/ui && npm run lint || echo "ESLint failed"

# Clean targets
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf web/ui/build/
	rm -rf data/

# Setup targets
setup:
	@echo "Setting up development environment..."
	@mkdir -p data
	@mkdir -p bin
	@echo "Installing UI dependencies..."
	cd web/ui && npm install
	@echo "Setup complete!"

# Docker targets
docker-build:
	@echo "Building Docker image..."
	docker build -t terratag-ui .

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 -v $(PWD)/data:/app/data terratag-ui

# Database targets
db-migrate:
	@echo "Running database migrations..."
	DB_PATH=./data/terratag.db go run cmd/api/main.go

db-reset:
	@echo "Resetting database..."
	rm -f data/terratag.db
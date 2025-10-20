BINARY_NAME=fizzbuzz-api
DOCKER_IMAGE=fizzbuzz-api
DOCKER_TAG=latest
GO=go
GOTEST=$(GO) test
GOVET=$(GO) vet
GOFMT=gofmt

.PHONY: help build build-linux test test-short test-coverage lint fmt vet run dev clean docker-build docker-run docker-stop docker-logs docker-clean compose-up compose-down compose-logs deps deps-update all ci

help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the application binary
	@echo "Building $(BINARY_NAME)..."
	$(GO) build -o bin/$(BINARY_NAME) ./cmd/server
	@echo "Build complete: bin/$(BINARY_NAME)"

build-linux: ## Build for Linux (useful for Docker)
	@echo "Building $(BINARY_NAME) for Linux..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -ldflags="-w -s" -o bin/$(BINARY_NAME)-linux ./cmd/server
	@echo "Build complete: bin/$(BINARY_NAME)-linux"

test: ## Run all tests
	@echo "Running tests..."
	$(GOTEST) -v -race ./...

test-short: ## Run short tests (skip integration tests)
	@echo "Running short tests..."
	$(GOTEST) -v -short ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint: ## Run golangci-lint
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run ./...

fmt: ## Format Go code
	@echo "Formatting code..."
	$(GOFMT) -s -w .
	@echo "Code formatted"

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOVET) ./...

run: ## Run the application locally
	@echo "Starting server..."
	$(GO) run ./cmd/server

dev: ## Run with live reload (requires air)
	@which air > /dev/null || (echo "air not installed. Run: go install github.com/cosmtrek/air@latest" && exit 1)
	air

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "Clean complete"

docker-build: ## Build Docker image
	@echo "Building Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

docker-run: ## Run Docker container
	@echo "Starting Docker container..."
	docker run -d --name $(BINARY_NAME) -p 8080:8080 $(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "Container started. Access at http://localhost:8080"

docker-stop: ## Stop Docker container
	@echo "Stopping Docker container..."
	docker stop $(BINARY_NAME) || true
	docker rm $(BINARY_NAME) || true

docker-logs: ## View Docker container logs
	docker logs -f $(BINARY_NAME)

docker-clean: docker-stop ## Clean Docker images and containers
	@echo "Cleaning Docker artifacts..."
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) || true

compose-up: ## Start services with docker-compose
	@echo "Starting services with docker-compose..."
	docker-compose up -d
	@echo "Services started. Access at http://localhost:8080"

compose-down: ## Stop docker-compose services
	@echo "Stopping services..."
	docker-compose down

compose-logs: ## View docker-compose logs
	docker-compose logs -f

deps: ## Download Go dependencies
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod verify

deps-update: ## Update Go dependencies
	@echo "Updating dependencies..."
	$(GO) get -u ./...
	$(GO) mod tidy

all: fmt vet lint test build ## Run all checks and build
	@echo "All tasks complete"

ci: fmt vet lint test-coverage ## Run CI pipeline
	@echo "CI pipeline complete"

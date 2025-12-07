.PHONY: help build test lint clean run docker-build docker-run install

# Variables
BINARY_NAME=prometheus-storagebox-exporter
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT=$(shell git rev-parse HEAD 2>/dev/null || echo "none")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-w -s -X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildDate=$(BUILD_DATE)"

# Default target
help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "Build complete: $(BINARY_NAME)"

build-all: ## Build binaries for all platforms
	@echo "Building for all platforms..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 .
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .
	@echo "All builds complete!"

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	@echo "Tests complete!"

test-coverage: test ## Run tests and show coverage
	@go tool cover -html=coverage.txt

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run ./...
	@echo "Linting complete!"

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .
	@echo "Formatting complete!"

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...
	@echo "Vetting complete!"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -rf dist/
	@rm -f coverage.txt
	@echo "Clean complete!"

run: build ## Build and run the exporter
	@echo "Running $(BINARY_NAME)..."
	@./$(BINARY_NAME)

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(BINARY_NAME):$(VERSION) \
		-t $(BINARY_NAME):latest \
		.
	@echo "Docker build complete!"

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run --rm -p 9509:9509 -e HETZNER_TOKEN=$(HETZNER_TOKEN) $(BINARY_NAME):latest

docker-compose-up: ## Start services with docker-compose
	@docker-compose up -d

docker-compose-down: ## Stop services with docker-compose
	@docker-compose down

install: build ## Install the binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME) to $(GOPATH)/bin..."
	@cp $(BINARY_NAME) $(GOPATH)/bin/
	@echo "Installation complete!"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@echo "Dependencies downloaded!"

tidy: ## Tidy go.mod
	@echo "Tidying go.mod..."
	@go mod tidy
	@echo "Tidy complete!"

verify: fmt vet lint test ## Run all verification steps
	@echo "All verification steps passed!"

release-check: ## Check if ready for release
	@echo "Checking release readiness..."
	@git diff --quiet || (echo "ERROR: Uncommitted changes detected" && exit 1)
	@go mod tidy
	@git diff --quiet go.mod go.sum || (echo "ERROR: go.mod or go.sum has changes" && exit 1)
	@echo "Ready for release!"

.DEFAULT_GOAL := help

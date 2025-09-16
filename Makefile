.PHONY: help build test clean lint

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the sttrouter binary
	@echo "Building sttrouter..."
	@go build -o bin/sttrouter main.go
	@echo "Build complete: bin/sttrouter"

test: ## Run all tests
	@echo "Running tests..."
	@go test -race -v ./...

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/

lint: ## Run golangci-lint
	@echo "Running linter..."
	@golangci-lint run
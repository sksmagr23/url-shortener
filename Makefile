.PHONY: help test lint run clean tidy test-coverage lint-fix lint-format

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

test: ## Run all tests
	go test ./... -v

test-coverage: ## Run tests with coverage
	go test ./... -v -cover

lint: ## Run golangci-lint
	golangci-lint run

lint-fix: ## Run golangci-lint with auto-fix
	golangci-lint run --fix

lint-format: ## Run golangci-lint formatting
	golangci-lint fmt

run: ## Run the application
	go run main.go

clean: ## Clean build artifacts
	rm -f url-shortener

tidy: ## Install dependencies
	go mod tidy
	go mod download

# Setup
setup: tidy ## Setup the project
	@echo "Setting up URL Shortener project..."
	@echo "Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Setup complete!"
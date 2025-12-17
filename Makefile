.PHONY: help build test test-unit test-integration test-examples clean lint fmt vet run-example-basic run-example-advanced run-example-streaming run-example-bedrock coverage

# Variables
GO := go
GOFLAGS := -v
BINARY_NAME := revenium-middleware-anthropic-go
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

# Default target
help:
	@echo "Revenium Middleware for Anthropic Go - Build Commands"
	@echo "======================================================"
	@echo ""
	@echo "Available targets:"
	@echo ""
	@echo "Build & Dependencies:"
	@echo "  make build              - Build the project"
	@echo "  make deps               - Download dependencies"
	@echo "  make tidy               - Tidy dependencies"
	@echo "  make clean              - Clean build artifacts"
	@echo ""
	@echo "Testing (No API Keys Required):"
	@echo "  make test-unit          - Run unit tests only"
	@echo "  make test               - Run all tests"
	@echo "  make test-verbose       - Run tests with verbose output"
	@echo "  make coverage           - Run tests with coverage report"
	@echo "  make coverage-html      - Generate HTML coverage report"
	@echo ""
	@echo "Testing (Requires .env with API Keys):"
	@echo "  make test-integration   - Run integration tests"
	@echo "  make test-examples      - Test all examples"
	@echo "  make run-example-basic  - Run basic example"
	@echo "  make run-example-advanced - Run advanced example"
	@echo "  make run-example-streaming - Run streaming example"
	@echo "  make run-example-bedrock - Run Bedrock example"
	@echo ""
	@echo "Code Quality:"
	@echo "  make fmt                - Format code with gofmt"
	@echo "  make vet                - Run go vet"
	@echo "  make lint               - Run linter (golangci-lint)"
	@echo ""
	@echo "Other:"
	@echo "  make help               - Show this help message"
	@echo ""

# Build the project
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) ./...
	@echo "Build complete!"

# Run all tests
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Run unit tests only (no API keys required)
test-unit:
	@echo "Running unit tests (no API keys required)..."
	$(GO) test -v ./tests/logger_test.go ./tests/middleware_test.go ./tests/provider_test.go ./tests/bedrock_test.go

# Run integration tests (requires .env file with API keys)
test-integration:
	@echo "Running integration tests (requires .env with API keys)..."
	@echo "Make sure your .env file has ANTHROPIC_API_KEY and REVENIUM_METERING_API_KEY set"
	$(GO) test -v ./tests/integration_test.go

# Test examples (requires .env file with API keys)
test-examples:
	@echo "Testing examples (requires .env with API keys)..."
	@echo "Testing basic examples..."
	$(GO) run examples/basic/main.go basic
	$(GO) run examples/basic/main.go basic-metadata
	@echo "Testing streaming examples..."
	$(GO) run examples/streaming/main.go basic
	$(GO) run examples/streaming/main.go advanced
	@echo "Testing advanced examples..."
	$(GO) run examples/advanced/main.go

# Run tests with verbose output
test-verbose:
	@echo "Running tests (verbose)..."
	$(GO) test -v -race ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=$(COVERAGE_FILE) ./...
	$(GO) tool cover -func=$(COVERAGE_FILE)

# Generate HTML coverage report
coverage-html: coverage
	@echo "Generating HTML coverage report..."
	$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

# Run linter
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "Code formatted!"

# Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...
	@echo "Vet complete!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(GO) clean
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	rm -rf bin/ dist/ build/
	@echo "Clean complete!"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	@echo "Dependencies downloaded!"

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GO) mod tidy
	@echo "Dependencies tidied!"

# Run basic example
run-example-basic:
	@echo "Running basic example..."
	$(GO) run examples/basic/main.go

# Run advanced example
run-example-advanced:
	@echo "Running advanced example..."
	$(GO) run examples/advanced/main.go

# Run streaming example
run-example-streaming:
	@echo "Running streaming example..."
	$(GO) run examples/streaming/main.go

# Run Bedrock example
run-example-bedrock:
	@echo "Running Bedrock example..."
	$(GO) run examples/bedrock/main.go

# Full CI pipeline
ci: fmt vet lint test coverage
	@echo "CI pipeline complete!"


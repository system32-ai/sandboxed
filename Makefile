# Makefile for sandboxed Go CLI application

# Variables
BINARY_NAME=sandboxed
MAIN_PATH=./main.go
BUILD_DIR=./build
VERSION?=1.0.0
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Default target
.PHONY: all
all: clean build

# Build the application
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for multiple platforms
.PHONY: build-all
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	# Linux
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	# macOS
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	# Windows
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Multi-platform build complete"

# Run the application
.PHONY: run
run:
	@echo "Running $(BINARY_NAME)..."
	go run $(MAIN_PATH)

# Run with arguments (usage: make run-args ARGS="greet Alice")
.PHONY: run-args
run-args:
	@echo "Running $(BINARY_NAME) with args: $(ARGS)"
	go run $(MAIN_PATH) $(ARGS)

# Test the application
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Test with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies installed"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Code formatted"

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		go vet ./...; \
	fi

# Vet code
.PHONY: vet
vet:
	@echo "Vetting code..."
	go vet ./...

# Install the binary to GOPATH/bin
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) $(MAIN_PATH)
	@echo "$(BINARY_NAME) installed to $(shell go env GOPATH)/bin"

# Uninstall the binary from GOPATH/bin
.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@rm -f $(shell go env GOPATH)/bin/$(BINARY_NAME)
	@echo "$(BINARY_NAME) uninstalled"

# Development workflow
.PHONY: dev
dev: fmt vet lint test build
	@echo "Development workflow complete"

# Create a release (usage: make release VERSION=1.2.3)
.PHONY: release
release:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: make release VERSION=1.2.3"; \
		exit 1; \
	fi
	@echo "Creating release $(VERSION)..."
	@chmod +x scripts/release.sh
	./scripts/release.sh $(VERSION)
	@echo "Release $(VERSION) created successfully"

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all          - Clean and build the application"
	@echo "  build        - Build the application"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  run          - Run the application"
	@echo "  run-args     - Run with arguments (use ARGS=\"your args\")"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Install and tidy dependencies"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code (requires golangci-lint)"
	@echo "  vet          - Vet code"
	@echo "  install      - Install binary to GOPATH/bin"
	@echo "  uninstall    - Remove binary from GOPATH/bin"
	@echo "  dev          - Run development workflow (fmt, vet, lint, test, build)"
	@echo "  release      - Create a GitHub release (requires VERSION=x.y.z)"
	@echo "  help         - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make run-args ARGS=\"greet Alice --uppercase\""
	@echo "  make VERSION=2.0.0 build"
	@echo "  make release VERSION=1.2.3"
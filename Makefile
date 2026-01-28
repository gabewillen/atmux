# Makefile for amux - Phase 0 verification

.PHONY: all test lint vet build clean conformance docs-check help

# Default target
all: build test lint

# Build binaries
build:
	@echo "Building amux binaries..."
	go build -o bin/amux ./cmd/amux
	go build -o bin/amux-node ./cmd/amux-node

# Run unit tests
test:
	@echo "Running unit tests..."
	go test -v ./...

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	go test -race ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linting
lint:
	@echo "Running linters..."
	@command -v staticcheck >/dev/null 2>&1 || { echo "staticcheck not installed, skipping"; exit 0; }
	staticcheck ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Run conformance suite
conformance:
	@echo "Running conformance suite..."
	@echo "Phase 0: Conformance suite stub"

# Check documentation sync
docs-check:
	@echo "Checking documentation sync..."
	@go run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./...
	@if [ -n "$$(git status --porcelain | grep README.md)" ]; then \
		echo "ERROR: README.md files are out of sync with Go documentation"; \
		echo "Run: go run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./..."; \
		echo "Then commit the changes."; \
		git status --porcelain | grep README.md; \
		exit 1; \
	fi
	@echo "Documentation is in sync"

# Verification - runs all checks
verify: tidy vet lint test test-race
	@echo "All verification checks passed!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Help
help:
	@echo "Available targets:"
	@echo "  all          - Build and test (default)"
	@echo "  build        - Build binaries"
	@echo "  test         - Run unit tests"
	@echo "  test-race    - Run tests with race detector"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint         - Run static analysis"
	@echo "  vet          - Run go vet"
	@echo "  tidy         - Tidy dependencies"
	@echo "  conformance  - Run conformance suite"
	@echo "  docs-check   - Check documentation sync"
	@echo "  verify       - Run all verification checks"
	@echo "  clean        - Remove build artifacts"

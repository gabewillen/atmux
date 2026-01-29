# amux Makefile
# Verification entrypoints for the Agent Multiplexer

.PHONY: all build test lint vet tidy staticcheck check conformance docs clean help

# Default target
all: check build

# Build all binaries
build:
	@echo "Building amux..."
	go build -o bin/amux ./cmd/amux
	@echo "Building amux-node..."
	go build -o bin/amux-node ./cmd/amux-node

# Run unit tests
test:
	go test -race ./...

# Run unit tests with coverage
test-cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	golangci-lint run ./...

# Run staticcheck
staticcheck:
	staticcheck ./...

# Run go vet
vet:
	go vet ./...

# Run go mod tidy
tidy:
	go mod tidy

# Check spec version guard
spec-check:
	@echo "Checking spec version..."
	@test -f docs/spec-v1.22.md || (echo "ERROR: docs/spec-v1.22.md not found" && exit 1)
	@grep -q "Version.*v1.22" docs/spec-v1.22.md || (echo "ERROR: spec version mismatch" && exit 1)
	@echo "Spec version OK"

# Generate documentation
docs:
	@echo "Generating documentation..."
	go run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./...

# Check documentation is up to date
docs-check: docs
	@git diff --exit-code || (echo "ERROR: Documentation out of sync. Run 'make docs' and commit." && exit 1)

# Run all checks (tidy, vet, lint, test)
check: tidy vet lint spec-check test

# Run conformance suite
conformance:
	@echo "Running conformance suite..."
	go test -v ./internal/conformance/...

# Run amux test command
amux-test: build
	./bin/amux test

# Run amux test with regression check
amux-test-regression: build
	./bin/amux test --regression

# Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html

# Cross-compile for all platforms
cross-compile:
	@echo "Building for linux/amd64..."
	GOOS=linux GOARCH=amd64 go build -o bin/amux-linux-amd64 ./cmd/amux
	GOOS=linux GOARCH=amd64 go build -o bin/amux-node-linux-amd64 ./cmd/amux-node
	@echo "Building for linux/arm64..."
	GOOS=linux GOARCH=arm64 go build -o bin/amux-linux-arm64 ./cmd/amux
	GOOS=linux GOARCH=arm64 go build -o bin/amux-node-linux-arm64 ./cmd/amux-node
	@echo "Building for darwin/amd64..."
	GOOS=darwin GOARCH=amd64 go build -o bin/amux-darwin-amd64 ./cmd/amux
	GOOS=darwin GOARCH=amd64 go build -o bin/amux-node-darwin-amd64 ./cmd/amux-node
	@echo "Building for darwin/arm64..."
	GOOS=darwin GOARCH=arm64 go build -o bin/amux-darwin-arm64 ./cmd/amux
	GOOS=darwin GOARCH=arm64 go build -o bin/amux-node-darwin-arm64 ./cmd/amux-node

# Install golangci-lint if not present
install-lint:
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Show help
help:
	@echo "amux Makefile targets:"
	@echo ""
	@echo "  all             - Run checks and build (default)"
	@echo "  build           - Build amux and amux-node binaries"
	@echo "  test            - Run unit tests with race detector"
	@echo "  test-cover      - Run tests with coverage report"
	@echo "  lint            - Run golangci-lint"
	@echo "  vet             - Run go vet"
	@echo "  tidy            - Run go mod tidy"
	@echo "  spec-check      - Verify spec-v1.22.md is present"
	@echo "  docs            - Generate documentation with go-docmd"
	@echo "  docs-check      - Verify documentation is up to date"
	@echo "  check           - Run all checks (tidy, vet, lint, spec-check, test)"
	@echo "  conformance     - Run conformance suite"
	@echo "  amux-test       - Run amux test command"
	@echo "  amux-test-regression - Run amux test with regression check"
	@echo "  cross-compile   - Cross-compile for all platforms"
	@echo "  clean           - Remove build artifacts"
	@echo "  install-lint    - Install golangci-lint"
	@echo "  help            - Show this help"

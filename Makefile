# Makefile for amux - Agent Multiplexer

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
CLI_BINARY=amux
DAEMON_BINARY=amux-node

# Build directory
BUILD_DIR=build
BIN_DIR=$(BUILD_DIR)/bin

# Source directories
CMD_DIR=cmd
INTERNAL_DIR=internal
PKG_DIR=pkg

# Version and build info
VERSION?=dev
BUILD_TIME?=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH?=$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")

# LDFLAGS for build
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT) -X main.gitBranch=$(GIT_BRANCH)"

# Default target
.PHONY: all
all: clean test build

# Build targets
.PHONY: build
build: build-cli build-daemon

.PHONY: build-cli
build-cli:
	@echo "Building CLI binary..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(CLI_BINARY) ./$(CMD_DIR)/$(CLI_BINARY)

.PHONY: build-daemon
build-daemon:
	@echo "Building daemon binary..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(DAEMON_BINARY) ./$(CMD_DIR)/$(DAEMON_BINARY)

# Test targets
.PHONY: test
test: test-unit test-integration test-conformance

.PHONY: test-unit
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v ./$(INTERNAL_DIR)/... ./$(PKG_DIR)/...

.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -tags=integration ./$(INTERNAL_DIR)/... ./$(PKG_DIR)/...

.PHONY: test-conformance
test-conformance:
	@echo "Running conformance tests..."
	$(GOTEST) -v ./$(INTERNAL_DIR)/conformance/...

# Static analysis targets
.PHONY: lint
lint: vet staticcheck

.PHONY: vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

.PHONY: staticcheck
staticcheck:
	@echo "Running staticcheck..."
	@which staticcheck > /dev/null || $(GOGET) honnef.co/go/tools/cmd/staticcheck@latest
	staticcheck ./...

# Coverage target
.PHONY: coverage
coverage:
	@echo "Running test coverage..."
	$(GOTEST) -cover ./$(INTERNAL_DIR)/... ./$(PKG_DIR)/...

# Benchmark target
.PHONY: benchmark
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./$(INTERNAL_DIR)/... ./$(PKG_DIR)/...

# Quality gate - runs all checks in sequence (continue on failure)
.PHONY: quality
quality: tidy vet lint test

# Module management
.PHONY: tidy
tidy:
	@echo "Tidying go.mod and go.sum..."
	$(GOMOD) tidy

.PHONY: download
download:
	@echo "Downloading dependencies..."
	$(GOMOD) download

.PHONY: verify
verify:
	@echo "Verifying dependencies..."
	$(GOMOD) verify

# Clean target
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f conformance-results.json

# Install target
.PHONY: install
install: build
	@echo "Installing binaries..."
	@mkdir -p $$GOPATH/bin
	@cp $(BIN_DIR)/$(CLI_BINARY) $$GOPATH/bin/
	@cp $(BIN_DIR)/$(DAEMON_BINARY) $$GOPATH/bin/

# Development targets
.PHONY: dev
dev: build-cli
	@echo "Starting development with CLI binary..."
	./$(BIN_DIR)/$(CLI_BINARY) --help

.PHONY: dev-daemon
dev-daemon: build-daemon
	@echo "Starting development with daemon binary..."
	./$(BIN_DIR)/$(DAEMON_BINARY) --help

# Documentation target (for go-docmd)
.PHONY: docs
docs:
	@echo "Generating documentation with go-docmd..."
	@which go-docmd > /dev/null || $(GOGET) github.com/agentflare-ai/go-docmd@latest
	go run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./...
	@echo "Documentation generated. Run 'git status' to see changes."

.PHONY: docs-check
docs-check: docs
	@echo "Checking for documentation changes..."
	@if [ -n "$$(git status --porcelain docs)" ]; then \
		echo "Documentation changes detected:"; \
		git status --porcelain docs; \
		echo "Please commit the generated documentation."; \
		exit 1; \
	fi

# Cross-compilation targets
.PHONY: build-all
build-all: clean
	@echo "Building for all platforms..."
	GOOS=linux GOARCH=amd64 $(MAKE) build-cli
	GOOS=darwin GOARCH=amd64 $(MAKE) build-cli
	GOOS=windows GOARCH=amd64 $(MAKE) build-cli

.PHONY: build-release
build-release: clean test quality build-all
	@echo "Creating release build with full verification..."

# Help target
.PHONY: help
help:
	@echo "amux - Agent Multiplexer Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build          Build CLI and daemon binaries"
	@echo "  build-cli      Build CLI binary only"
	@echo "  build-daemon   Build daemon binary only"
	@echo "  test           Run all tests"
	@echo "  test-unit      Run unit tests"
	@echo "  test-integration Run integration tests"
	@echo "  test-conformance Run conformance tests"
	@echo "  quality        Run quality checks (tidy, vet, lint, test)"
	@echo "  lint           Run linting (vet + staticcheck)"
	@echo "  vet            Run go vet"
	@echo "  staticcheck    Run staticcheck"
	@echo "  coverage       Run test coverage"
	@echo "  benchmark      Run benchmarks"
	@echo "  tidy           Tidy go.mod and go.sum"
	@echo "  download       Download dependencies"
	@echo "  verify         Verify dependencies"
	@echo "  clean          Clean build artifacts"
	@echo "  install        Install binaries to GOPATH/bin"
	@echo "  dev            Build and run CLI in development mode"
	@echo "  dev-daemon     Build and run daemon in development mode"
	@echo "  docs           Generate documentation with go-docmd"
	@echo "  docs-check     Generate and check documentation"
	@echo "  build-all      Build for all platforms"
	@echo "  build-release  Create full release build"
	@echo "  help           Show this help message"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION        Build version (default: dev)"
	@echo "  BUILD_TIME    Build timestamp (auto-generated)"
	@echo "  GIT_COMMIT    Git commit hash (auto-generated)"
	@echo "  GIT_BRANCH    Git branch (auto-generated)"

# Special targets for Phase 0 compliance
.PHONY: verify-entrypoints
verify-entrypoints: build test conformance
	@echo "All verification entrypoints executed successfully"

# Phony declarations for safety
.PHONY: all build build-cli build-daemon test test-unit test-integration test-conformance
.PHONY: lint vet staticcheck coverage benchmark tidy download verify clean install
.PHONY: dev dev-daemon docs docs-check build-all build-release verify-entrypoints
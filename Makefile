# Makefile for amux project verification entrypoints

# Default target
.DEFAULT_GOAL := help

# Variables
GO ?= go
GOFMT ?= gofmt
GOLINT ?= golangci-lint
STATICCHECK ?= staticcheck
TEST_TIMEOUT ?= 30s

# Help target
.PHONY: help
help: ## Display this help message
	@echo "Available targets:"
	@echo
	@grep -E '^[a-zA-Z_0-9%-]+:.*?## .*$$' $(word 1,$(MAKEFILE_LIST)) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "%-30s %s\n", $$1, $$2}'

# Run unit tests
.PHONY: test
test: ## Run unit tests
	$(GO) test -v -timeout $(TEST_TIMEOUT) ./...

# Run unit tests with race detection
.PHONY: test-race
test-race: ## Run unit tests with race detection
	$(GO) test -v -race -timeout $(TEST_TIMEOUT) ./...

# Run lint/static analysis
.PHONY: lint
lint: ## Run lint/static analysis
	$(GOLINT) run ./...
	$(STATICCHECK) ./...

# Run go vet
.PHONY: vet
vet: ## Run go vet
	$(GO) vet ./...

# Run go fmt
.PHONY: fmt
fmt: ## Run go fmt
	$(GOFMT) -s -w .

# Run go docmd to generate documentation
.PHONY: docs
docs: ## Generate documentation with go-docmd
	$(GO) run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./...

# Run docs check to ensure docs are up to date
.PHONY: docs-check
docs-check: ## Check if docs are up to date
	@echo "Checking if docs are up to date..."
	@$(GO) run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./... 2>&1 | grep -q "nothing to commit" || (echo "Documentation is out of date. Run 'make docs' to update." && exit 1)

# Run integration tests (if any)
.PHONY: integration
integration: ## Run integration tests
	@echo "Running integration tests..."
	$(GO) test -tags=integration -v -timeout $(TEST_TIMEOUT) ./...

# Run all tests (unit + integration)
.PHONY: test-all
test-all: test integration ## Run all tests (unit + integration)

# Run benchmark tests
.PHONY: bench
bench: ## Run benchmark tests
	$(GO) test -bench=. -benchmem ./...

# Run coverage analysis
.PHONY: coverage
coverage: ## Run coverage analysis
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run amux test command (snapshot generation)
.PHONY: amux-test
amux-test: ## Run amux test to create a Go verification snapshot
	$(GO) run ./cmd/amux test

# Run amux test with regression checking
.PHONY: amux-test-regression
amux-test-regression: ## Run amux test --regression to compare against previous snapshot
	$(GO) run ./cmd/amux test --regression

# Run conformance suite
.PHONY: conformance
conformance: ## Run conformance suite
	@echo "Running conformance suite..."
	$(GO) run ./cmd/amux conformance

# Build the CLI client
.PHONY: build-cli
build-cli: ## Build the CLI client
	$(GO) build -o bin/amux ./cmd/amux

# Build the daemon
.PHONY: build-daemon
build-daemon: ## Build the daemon
	$(GO) build -o bin/amux-node ./cmd/amux-node

# Build all binaries
.PHONY: build
build: build-cli build-daemon ## Build all binaries

# Clean build artifacts
.PHONY: clean
clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

# Verify everything (lint, test, vet, docs-check)
.PHONY: verify
verify: lint vet test docs-check ## Run full verification (lint, test, vet, docs-check)
	@echo "All verification steps passed!"

# Install development tools
.PHONY: install-tools
install-tools: ## Install development tools
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install honnef.co/go/tools/cmd/staticcheck@latest
	$(GO) install github.com/agentflare-ai/go-docmd@latest
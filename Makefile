# amux Makefile - Phase 0 verification entrypoints
# Provides reproducible verification commands suitable for CI automation

.PHONY: help test lint vet tidy build clean docs-check docs-gen conformance amux-test

help: ## Show this help message
	@echo "amux verification commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'
	@echo ""

tidy: ## Run go mod tidy
	go mod tidy

vet: ## Run go vet
	go vet ./...

lint: ## Run staticcheck linter
	@which staticcheck >/dev/null || (echo "Installing staticcheck..." && go install honnef.co/go/tools/cmd/staticcheck@latest)
	staticcheck -checks=-U1000 ./...

test: ## Run unit tests
	go test ./...

test-race: ## Run unit tests with race detector
	go test -race ./...

test-coverage: ## Run tests with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

build: ## Build binaries
	go build -o bin/amux ./cmd/amux
	go build -o bin/amux-node ./cmd/amux-node

docs-gen: ## Generate per-package README files using go-docmd
	go run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./...

docs-check: docs-gen ## Check that generated docs are in sync
	@if ! git diff --exit-code --quiet; then \
		echo "❌ Generated documentation is out of sync. Run 'make docs-gen' and commit the changes."; \
		git diff; \
		exit 1; \
	else \
		echo "✅ Generated documentation is in sync."; \
	fi

conformance: ## Run conformance tests (placeholder)
	@echo "Conformance suite not yet implemented (Phase 0)"
	@exit 1

amux-test: ## Run amux test to create baseline snapshot
	@echo "amux test command not yet implemented (Phase 0)"
	@exit 1

# Combined verification target for CI
verify: tidy vet lint test test-race build docs-check ## Run all verification steps
	@echo "✅ All verification steps passed"

clean: ## Clean build artifacts
	rm -rf bin/ coverage.out coverage.html
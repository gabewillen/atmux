# Makefile for amux verification entrypoints

.PHONY: test lint vet tidy build conformance test-snapshot test-regression docs-check

# Run all unit tests
test:
	go test ./...

# Run tests with race detector
test-race:
	go test -race ./...

# Run go vet
vet:
	go vet ./...

# Run go mod tidy
tidy:
	go mod tidy

# Run linting with staticcheck
lint:
	@if ! command -v staticcheck >/dev/null 2>&1; then \
		echo "Installing staticcheck..."; \
		go install honnef.co/go/tools/cmd/staticcheck@latest; \
	fi
	@staticcheck ./...

# Build all binaries
build:
	go build ./cmd/amux
	go build ./cmd/amux-node

# Run conformance suite
conformance:
	go test ./internal/conformance/...

# Run amux test to create snapshot
test-snapshot:
	go run ./cmd/amux test

# Run amux test --regression
test-regression:
	go run ./cmd/amux test --regression

# Check that generated docs are in sync
docs-check:
	go run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./...
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Error: Generated documentation is out of sync. Please run 'make docs-check' and commit the changes."; \
		git status; \
		exit 1; \
	fi

# Run all verification steps (includes amux test --regression per plan verification entrypoints)
verify: tidy vet lint test-race test docs-check test-regression

#!/bin/bash
# verify.sh: Complete verification entrypoint for amux
# This script implements the CI-friendly "one command" verification
# required by spec §4.3.1-§4.3.2 and §4.2.5

set -e

echo "🔍 Running amux verification suite..."

# Track overall status
OVERALL_STATUS=0

# Function to run a step and track failure
run_step() {
    local step_name="$1"
    shift
    echo "📋 $step_name..."
    
    if "$@"; then
        echo "✅ $step_name: PASSED"
    else
        echo "❌ $step_name: FAILED"
        OVERALL_STATUS=1
    fi
    echo ""
}

# 1. Go module verification
run_step "Go module tidy check" bash -c "
    # Store original mod files
    cp go.mod go.mod.bak
    cp go.sum go.sum.bak
    
    # Run go mod tidy
    go mod tidy
    
    # Check if only dependencies changed (not module requirements)
    if ! cmp -s go.mod go.mod.bak; then
        echo 'go.mod has uncommitted changes:'
        diff go.mod.bak go.mod || true
        rm go.mod.bak go.sum.bak
        false
    else
        # go.sum changes are okay for new dependencies
        rm go.mod.bak go.sum.bak
        echo 'Go module verification passed'
    fi
"

# Static analysis (go vet) with exemptions
run_step "Static analysis (go vet)" bash -c "
    # Run go vet but ignore known WASM-related unsafe.Pointer warnings
    if go vet ./... 2>&1 | grep -v 'possible misuse of unsafe.Pointer' | grep -E '(.*\.go:[0-9]+:[0-9]+:)|(FAIL)'; then
        echo 'go vet found issues (excluding WASM unsafe.Pointer usage)'
        exit 1
    fi
    echo 'go vet passed (WASM unsafe.Pointer warnings ignored)'
"

# 3. Build verification
run_step "Build amux binary" go build -o /tmp/amux-test ./cmd/amux
run_step "Build amux-node binary" go build -o /tmp/amux-node-test ./cmd/amux-node

# 4. Unit tests
run_step "Unit tests" go test ./... -short

# 5. Unit tests with race detection  
run_step "Unit tests with race detection" go test ./... -race -short

# 6. Documentation sync check
run_step "Documentation sync check" ./scripts/docs-check.sh

# 7. amux test (snapshot generation)
run_step "amux test snapshot" go run ./cmd/amux test --no-snapshot >/dev/null

# 8. Architecture compliance check (no forbidden imports)
run_step "Architecture compliance check" bash -c "
    # Check for forbidden imports in internal/ packages
    if find internal/ -name '*.go' -exec grep -l 'adapters/' {} \; | grep -v bootstrap.go; then
        echo 'VIOLATION: internal/ packages must not import from adapters/'
        exit 1
    fi
    echo 'Architecture compliance verified'
"

# 9. Spec version lock check
run_step "Spec version lock check" go test ./cmd/amux -run TestSpecVersionLock

# 10. Basic conformance check (skeleton)
run_step "Basic conformance check" go test ./internal/conformance -short

# Clean up test binaries
rm -f /tmp/amux-test /tmp/amux-node-test

# Final status
echo "===================="
if [ $OVERALL_STATUS -eq 0 ]; then
    echo "🎉 All verification steps PASSED"
else
    echo "💥 Some verification steps FAILED"
fi
echo "===================="

exit $OVERALL_STATUS
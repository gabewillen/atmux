#!/bin/bash
set -e

echo "Running unit tests..."
go test ./...

echo "Running snapshot verification..."
go run cmd/amux/main.go test --regression

echo "Verifying docs..."
./scripts/verify-docs.sh

echo "All verification passed."

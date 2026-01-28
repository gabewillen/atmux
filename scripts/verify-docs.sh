#!/bin/bash
set -e

echo "Running go-docmd..."
go run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./...

if ! git diff --exit-code; then
    echo "Documentation is out of sync! Please commit the generated README.md files."
    exit 1
fi
echo "Docs verified."

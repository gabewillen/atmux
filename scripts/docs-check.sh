#!/bin/bash
# docs-check.sh: Verify that generated READMEs are in sync with Go documentation
# This script implements the automated docs sync check required by spec §4.2.6.1

set -e

echo "🔍 Running docs sync check..."

# Store original state
ORIGINAL_STATE=$(git status --porcelain 2>/dev/null || echo "")

# Run the canonical go-docmd generation command
echo "📝 Generating documentation with go-docmd..."
go run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./...

# Check if any files were changed
NEW_STATE=$(git status --porcelain 2>/dev/null || echo "")

if [ "$ORIGINAL_STATE" != "$NEW_STATE" ]; then
    echo "❌ Documentation is out of sync!"
    echo ""
    echo "The following files were changed by go-docmd:"
    git diff --name-only 2>/dev/null || echo "(unable to show diff - not a git repo)"
    echo ""
    echo "Please run the following command to fix:"
    echo "  go run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./..."
    echo "Then commit the updated README.md files."
    exit 1
fi

echo "✅ Documentation is in sync!"
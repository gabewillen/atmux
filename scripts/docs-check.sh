#!/usr/bin/env bash
set -euo pipefail

go run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./...

git diff --exit-code

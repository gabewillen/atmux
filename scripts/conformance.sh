#!/usr/bin/env bash
set -euo pipefail

go test ./internal/conformance -run TestRunnerWritesResults

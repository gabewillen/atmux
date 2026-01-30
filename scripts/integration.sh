#!/usr/bin/env bash
set -euo pipefail

go test ./... -tags=integration -run Integration

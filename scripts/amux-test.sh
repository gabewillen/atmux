#!/usr/bin/env bash
set -euo pipefail

go run ./cmd/amux test "$@"

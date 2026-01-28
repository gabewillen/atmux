#!/usr/bin/env bash
set -euo pipefail

scripts/docs-check.sh
scripts/lint.sh
scripts/unit.sh
scripts/integration.sh
scripts/conformance.sh
scripts/amux-test.sh

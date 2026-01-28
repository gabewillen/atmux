#!/usr/bin/env bash
[ -n "$BASH_VERSION" ] || exec bash "$0" "$@"
set -euo pipefail

# Repository root (one level up from this script)
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"

IMAGE="bitnami/minideb:bookworm"

# Host paths
DOCS_DIR="$ROOT_DIR/docs"
JIMMY_HOST="/Users/gabrielwillen/VSCode/amux-antigravity-gemini-3.0-pro-high"
FRANK_HOST="$ROOT_DIR/.worktrees/claude-code/opus"
GREG_HOST="$ROOT_DIR/.worktrees/codex/gpt-5.2-codex-high"
CURLY_HOST="$ROOT_DIR/.worktrees/cursor-cli/auto"
JEREMY_HOST="$ROOT_DIR/.worktrees/gemini-cli/auto"
BIG_PICKLE_HOST="$ROOT_DIR/.worktrees/opencode/big-pickle"
QUAKER_HOST="$ROOT_DIR/.worktrees/qoder/auto"
QUINTON_HOST="$ROOT_DIR/.worktrees/qwen-code/coder-model"

# Run minideb with /amux as working directory.
# /amux and its subdirectories are created on container start.
CONTAINER_ID=$(docker run --detach \
  -v "${DOCS_DIR}:/amux/docs:ro" \
  -v "${JIMMY_HOST}:/amux/worktrees/jimmy:ro" \
  -v "${FRANK_HOST}:/amux/worktrees/frank:ro" \
  -v "${GREG_HOST}:/amux/worktrees/greg:ro" \
  -v "${CURLY_HOST}:/amux/worktrees/curly:ro" \
  -v "${JEREMY_HOST}:/amux/worktrees/jeremy:ro" \
  -v "${BIG_PICKLE_HOST}:/amux/worktrees/big-pickle:ro" \
  -v "${QUAKER_HOST}:/amux/worktrees/quaker:ro" \
  -v "${QUINTON_HOST}:/amux/worktrees/quinton:ro" \
  -v "${ROOT_DIR}/journal:/amux/journal:rw" \
  -w /amux \
  "${IMAGE}" \
  bash -c "apt-get update && apt-get install -y wget procps && sleep infinity")

echo "Started container ${CONTAINER_ID}"
docker ps --filter "id=${CONTAINER_ID}"

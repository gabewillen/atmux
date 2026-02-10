#!/usr/bin/env bash

set -euo pipefail

usage() {
  cat <<'EOF'
usage: cleanup_sessions.sh --force

Destructive cleanup for multi-agent local environment:
1. Kills tmux sessions: agent-0..agent-4
2. Stops background monitor process
3. Removes git worktrees under worktrees/agent-0..agent-4
4. Deletes agent branches used by those worktrees

This command requires --force.
EOF
}

if [[ $# -ne 1 ]] || [[ "${1:-}" != "--force" ]]; then
  usage >&2
  exit 2
fi

if [[ -n "${TMUX:-}" ]]; then
  echo "cleanup_sessions.sh must be run outside tmux." >&2
  exit 1
fi

if ! command -v tmux >/dev/null 2>&1; then
  echo "tmux is required but not installed." >&2
  exit 1
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
MONITOR_SCRIPT="${ROOT_DIR}/.agents/tmux/monitor_agent_activity.sh"

SESSIONS=(agent-0 agent-1 agent-2 agent-3 agent-4)
WORKTREES=(
  "${ROOT_DIR}/worktrees/agent-0"
  "${ROOT_DIR}/worktrees/agent-1"
  "${ROOT_DIR}/worktrees/agent-2"
  "${ROOT_DIR}/worktrees/agent-3"
  "${ROOT_DIR}/worktrees/agent-4"
)
BRANCHES=(
  "agent-0/coordination"
  "agent-1/core-api-runtime"
  "agent-2/inference-session-generation"
  "agent-3/state-io-recovery"
  "agent-4/app-integration"
)

echo "[1/4] killing tmux sessions"
for session in "${SESSIONS[@]}"; do
  if tmux has-session -t "${session}" 2>/dev/null; then
    tmux kill-session -t "${session}"
    echo "killed tmux session: ${session}"
  else
    echo "tmux session not found: ${session}"
  fi
done

echo "[2/4] stopping monitor process"
if pgrep -f "${MONITOR_SCRIPT}" >/dev/null 2>&1; then
  pkill -f "${MONITOR_SCRIPT}" || true
  sleep 0.2
  echo "stopped monitor processes"
else
  echo "no monitor process found"
fi

echo "[3/4] removing worktrees"
for wt in "${WORKTREES[@]}"; do
  if git -C "${ROOT_DIR}" worktree list | rg -F "${wt}" >/dev/null 2>&1; then
    git -C "${ROOT_DIR}" worktree remove -f "${wt}"
    echo "removed worktree: ${wt}"
  elif [[ -d "${wt}" ]]; then
    rm -rf "${wt}"
    echo "removed orphan worktree dir: ${wt}"
  else
    echo "worktree not found: ${wt}"
  fi
done

echo "[4/4] deleting agent branches"
for branch in "${BRANCHES[@]}"; do
  if git -C "${ROOT_DIR}" show-ref --verify --quiet "refs/heads/${branch}"; then
    git -C "${ROOT_DIR}" branch -D "${branch}"
    echo "deleted branch: ${branch}"
  else
    echo "branch not found: ${branch}"
  fi
done

echo "cleanup complete"

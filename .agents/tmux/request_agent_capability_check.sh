#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SEND_SCRIPT="${ROOT_DIR}/.agents/tmux/send_to_agent.sh"
TIMEOUT_SECONDS="${TIMEOUT_SECONDS:-90}"

if ! command -v tmux >/dev/null 2>&1; then
  echo "tmux is required but not installed." >&2
  exit 1
fi

if ! command -v rg >/dev/null 2>&1; then
  echo "rg is required but not installed." >&2
  exit 1
fi

if [[ ! -x "${SEND_SCRIPT}" ]]; then
  echo "missing executable send script: ${SEND_SCRIPT}" >&2
  exit 1
fi

for s in agent-0 agent-1 agent-2 agent-3 agent-4; do
  if ! tmux has-session -t "${s}" 2>/dev/null; then
    echo "missing tmux session: ${s}" >&2
    exit 1
  fi
done

token="capcheck-$(date -u +%Y%m%dT%H%M%SZ)"
echo "request token: ${token}"

for agent in agent-1 agent-2 agent-3 agent-4; do
  "${SEND_SCRIPT}" "${agent}" question \
    "Capability check request ${token}: verify read/write access in your worktree, web/search availability, and ability to run commands. Reply to agent-0 with type=status including EXACT token ${token} and PASS/FAIL."
done

start_epoch="$(date +%s)"
for agent in agent-1 agent-2 agent-3 agent-4; do
  got=0
  while true; do
    now_epoch="$(date +%s)"
    if (( now_epoch - start_epoch > TIMEOUT_SECONDS )); then
      break
    fi

    pane="$(tmux capture-pane -p -S -1200 -t agent-0 || true)"
    if printf '%s\n' "${pane}" | rg -F "[MSG][from:${agent}][type:status]" >/dev/null 2>&1 && \
      printf '%s\n' "${pane}" | rg -F "${token}" >/dev/null 2>&1; then
      got=1
      break
    fi

    sleep 2
  done

  if [[ "${got}" -eq 1 ]]; then
    echo "OK: received capability reply from ${agent} (${token})"
  else
    echo "FAIL: no capability reply from ${agent} within ${TIMEOUT_SECONDS}s (${token})" >&2
    exit 1
  fi
done

echo "all agents replied to capability check (${token})"

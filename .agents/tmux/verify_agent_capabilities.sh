#!/usr/bin/env bash

set -euo pipefail

if ! command -v tmux >/dev/null 2>&1; then
  echo "tmux is required but not installed." >&2
  exit 1
fi

if ! command -v rg >/dev/null 2>&1; then
  echo "rg is required but not installed." >&2
  exit 1
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

agents=(agent-0 agent-1 agent-2 agent-3 agent-4)
fail_count=0

check_equals() {
  local actual="$1"
  local expected="$2"
  local label="$3"
  if [[ "${actual}" != "${expected}" ]]; then
    echo "FAIL: ${label} expected='${expected}' actual='${actual}'" >&2
    fail_count=$((fail_count + 1))
  else
    echo "OK: ${label}=${actual}"
  fi
}

require_line() {
  local file="$1"
  local pattern="$2"
  local label="$3"
  if rg -n --fixed-strings -- "${pattern}" "${file}" >/dev/null 2>&1; then
    echo "OK: ${label}"
  else
    echo "FAIL: ${label} missing pattern '${pattern}' in ${file}" >&2
    fail_count=$((fail_count + 1))
  fi
}

echo "[1/4] checking launcher wiring"
require_line "${ROOT_DIR}/.agents/tmux/start_sessions.sh" '--sandbox ' \
  "sandbox flag wiring"
require_line "${ROOT_DIR}/.agents/tmux/start_sessions.sh" 'codex_sandbox_mode' \
  "sandbox variable wiring"
require_line "${ROOT_DIR}/.agents/tmux/start_sessions.sh" \
  '--ask-for-approval ' "approval flag wiring"
require_line "${ROOT_DIR}/.agents/tmux/start_sessions.sh" \
  'codex_approval_policy' "approval variable wiring"
require_line "${ROOT_DIR}/.agents/tmux/start_sessions.sh" '${search_flag}' "search override wiring"

echo "[2/4] checking per-agent config"
for agent in "${agents[@]}"; do
  cfg="${ROOT_DIR}/.agents/${agent}/codex.env"
  if [[ ! -f "${cfg}" ]]; then
    echo "FAIL: missing ${cfg}" >&2
    fail_count=$((fail_count + 1))
    continue
  fi

  # shellcheck source=/dev/null
  source "${cfg}"
  check_equals "${codex_model:-}" "gpt-5.3-codex" "${agent}.codex_model"
  if [[ "${agent}" == "agent-0" ]]; then
    check_equals "${codex_reasoning_effort:-}" "low" "${agent}.codex_reasoning_effort"
  else
    check_equals "${codex_reasoning_effort:-}" "high" "${agent}.codex_reasoning_effort"
  fi
  check_equals "${codex_search:-}" "true" "${agent}.codex_search"
  check_equals "${codex_approval_policy:-}" "never" "${agent}.codex_approval_policy"
  check_equals "${codex_sandbox_mode:-}" "danger-full-access" "${agent}.codex_sandbox_mode"
done

echo "[3/4] checking live tmux sessions and Codex preamble"
for agent in "${agents[@]}"; do
  if ! tmux has-session -t "${agent}" 2>/dev/null; then
    echo "FAIL: missing tmux session ${agent}" >&2
    fail_count=$((fail_count + 1))
    continue
  fi

  preamble_ok=0
  for _ in {1..12}; do
    pane="$(tmux capture-pane -p -S -240 -t "${agent}" || true)"
    if printf '%s\n' "${pane}" | rg "model:\s+gpt-5.3-codex" >/dev/null 2>&1; then
      preamble_ok=1
      break
    fi
    sleep 1
  done
  if [[ "${preamble_ok}" -eq 1 ]]; then
    echo "OK: ${agent} preamble model detected"
  else
    echo "FAIL: ${agent} preamble missing expected model" >&2
    fail_count=$((fail_count + 1))
  fi
done

echo "[4/4] checking worktree capability scaffolding"
for agent in "${agents[@]}"; do
  w="${ROOT_DIR}/worktrees/${agent}"
  if [[ ! -d "${w}" ]]; then
    echo "FAIL: missing worktree ${w}" >&2
    fail_count=$((fail_count + 1))
    continue
  fi
  if [[ ! -L "${w}/.agents" ]]; then
    echo "FAIL: ${agent} missing .agents symlink" >&2
    fail_count=$((fail_count + 1))
  else
    echo "OK: ${agent} has .agents symlink"
  fi
  if [[ ! -d "${w}/tmp/llama.cpp" ]]; then
    echo "FAIL: ${agent} missing tmp/llama.cpp mirror" >&2
    fail_count=$((fail_count + 1))
  else
    echo "OK: ${agent} has tmp/llama.cpp mirror"
  fi
done

if [[ "${fail_count}" -ne 0 ]]; then
  echo "capability verification FAILED (${fail_count} issues)" >&2
  exit 1
fi

echo "capability verification PASSED"

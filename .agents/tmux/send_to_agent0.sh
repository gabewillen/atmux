#!/usr/bin/env bash

set -euo pipefail

if [[ $# -lt 2 ]]; then
  echo "usage: $0 [from-agent] <type> <message...>" >&2
  exit 1
fi

detect_from_agent() {
  if [[ -n "${FROM_AGENT:-}" ]]; then
    printf '%s\n' "${FROM_AGENT}"
    return 0
  fi

  if [[ -n "${TMUX:-}" ]]; then
    local detected_session
    detected_session="$(tmux display-message -p '#S' 2>/dev/null || true)"
    if [[ -n "${detected_session}" ]]; then
      printf '%s\n' "${detected_session}"
      return 0
    fi
  fi
  printf '%s\n' "user"
}

from_agent=""
msg_type=""
if [[ $# -ge 3 ]]; then
  case "$1" in
    question|blocker|handoff|ready-to-merge|status|decision|merge-approved)
      from_agent="$(detect_from_agent)"
      msg_type="$1"
      shift 1
      ;;
    *)
      # Backward-compatible form: <from-agent> <type> <message...>
      from_agent="$1"
      msg_type="$2"
      shift 2
      ;;
  esac
else
  from_agent="$(detect_from_agent)"
  msg_type="$1"
  shift 1
fi

message="$*"

ts="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
payload="[MSG][from:${from_agent}][type:${msg_type}][ts:${ts}] ${message}"
target="agent-0"
lock_name="agent-send-lock-${target}"
max_wait_seconds="${TMUX_SEND_MAX_WAIT_SECONDS:-8}"

if ! tmux has-session -t "${target}" 2>/dev/null; then
  echo "tmux session agent-0 not found. Start sessions with ./agents/tmux/start_sessions.sh" >&2
  exit 1
fi

wait_for_idle_input() {
  local waited=0
  local last_line=""
  while (( waited < max_wait_seconds * 10 )); do
    last_line="$(tmux capture-pane -p -t "${target}" | tail -n 1)"
    # Best-effort detection for an active prompt line that still has unsent typed content.
    # Example: "› partially typed message"
    if [[ ! "${last_line}" =~ ^[[:space:]]*[›\>][[:space:]]+[^[:space:]].*$ ]]; then
      return 0
    fi
    sleep 0.1
    ((waited += 1))
  done
  echo "warning: target ${target} still shows pending input after ${max_wait_seconds}s; sending anyway" >&2
}

tmux wait-for -L "${lock_name}"
trap 'tmux wait-for -U "${lock_name}"' EXIT
wait_for_idle_input

pane_in_mode="$(tmux display-message -p -t "${target}" "#{pane_in_mode}" 2>/dev/null || echo 0)"

if [[ "${pane_in_mode}" == "1" ]]; then
  tmux send-keys -t "${target}" -X cancel
  sleep 0.2
fi

tmux send-keys -t "${target}" C-u
sleep 0.2
tmux send-keys -t "${target}" -l "${payload}"
sleep 0.2
tmux send-keys -t "${target}" ENTER

printf 'sent -> agent-0: %s\n' "${payload}"

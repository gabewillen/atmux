#!/usr/bin/env bash

set -euo pipefail

if [[ $# -lt 3 ]]; then
  echo "usage: $0 <target-agent-session> <type> <message...>" >&2
  exit 1
fi

target="$1"
msg_type="$2"
shift 2
message="$*"

from_agent="${FROM_AGENT:-user}"
if [[ -n "${TMUX:-}" ]]; then
  detected_session="$(tmux display-message -p '#S' 2>/dev/null || true)"
  if [[ -n "${detected_session}" ]]; then
    from_agent="${detected_session}"
  fi
fi

ts="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
payload="[MSG][from:${from_agent}][type:${msg_type}][ts:${ts}] ${message}"
lock_name="agent-send-lock-${target}"
max_wait_seconds="${TMUX_SEND_MAX_WAIT_SECONDS:-8}"

if ! tmux has-session -t "${target}" 2>/dev/null; then
  echo "tmux session ${target} not found." >&2
  exit 1
fi

wait_for_idle_input() {
  local waited=0
  local last_line=""
  while (( waited < max_wait_seconds * 10 )); do
    last_line="$(tmux capture-pane -p -t "${target}" | tail -n 1)"
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

printf 'sent -> %s: %s\n' "${target}" "${payload}"

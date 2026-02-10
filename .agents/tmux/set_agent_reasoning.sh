#!/usr/bin/env bash

set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: $0 <agent-session> <low|medium|high|extra-high|1|2|3|4>" >&2
  exit 1
fi

target="$1"
level_input="$2"

if ! command -v tmux >/dev/null 2>&1; then
  echo "tmux is required but not installed." >&2
  exit 1
fi

if ! tmux has-session -t "${target}" 2>/dev/null; then
  echo "tmux session ${target} not found." >&2
  exit 1
fi

pane_text="$(tmux capture-pane -p -S -220 -t "${target}" 2>/dev/null || true)"
was_busy=0
if printf '%s\n' "${pane_text}" | rg -F "esc to interrupt" >/dev/null; then
  was_busy=1
fi

reasoning_key=""
case "${level_input}" in
  low|1)
    reasoning_key="1"
    ;;
  medium|2)
    reasoning_key="2"
    ;;
  high|3)
    reasoning_key="3"
    ;;
  extra-high|extra_high|4)
    reasoning_key="4"
    ;;
  *)
    echo "invalid reasoning level: ${level_input}" >&2
    echo "expected one of: low, medium, high, extra-high, 1, 2, 3, 4" >&2
    exit 2
    ;;
esac

# If pane is in copy-mode, exit it so keypresses reach Codex.
pane_in_mode="$(tmux display-message -p -t "${target}" "#{pane_in_mode}" 2>/dev/null || echo 0)"
if [[ "${pane_in_mode}" == "1" ]]; then
  tmux send-keys -t "${target}" -X cancel
  sleep 0.1
fi

if [[ "${was_busy}" -eq 1 ]]; then
  tmux send-keys -t "${target}" Escape
  sleep 0.25
fi

# Open model selector, choose gpt-5.3 option (2), then choose reasoning option.
tmux send-keys -t "${target}" C-u
sleep 0.1
tmux send-keys -t "${target}" -l "/model"
sleep 0.1
tmux send-keys -t "${target}" Enter
sleep 0.35
tmux send-keys -t "${target}" -l "2"
sleep 0.25
tmux send-keys -t "${target}" -l "${reasoning_key}"

if [[ "${was_busy}" -eq 1 ]]; then
  sleep 0.2
  tmux send-keys -t "${target}" C-u
  sleep 0.1
  tmux send-keys -t "${target}" -l "continue"
  tmux send-keys -t "${target}" Enter
  echo "sent model selection to ${target}: gpt-5.3 + reasoning ${reasoning_key} (interrupted+continued)"
else
  echo "sent model selection to ${target}: gpt-5.3 + reasoning ${reasoning_key}"
fi

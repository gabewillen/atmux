#!/usr/bin/env bash

set -euo pipefail

INTERVAL_SECONDS="${INTERVAL_SECONDS:-20}"
STALE_CYCLES="${STALE_CYCLES:-3}"
REMINDER_CYCLES="${REMINDER_CYCLES:-3}"
MAX_ITERATIONS="${MAX_ITERATIONS:-0}"
WATCH_SESSIONS=("agent-1" "agent-2" "agent-3" "agent-4")
FROM_AGENT="${FROM_AGENT:-agent-0-monitor}"

usage() {
  cat <<'EOF'
usage: monitor_agent_activity.sh [--interval N] [--stale-cycles N] [--reminder-cycles N] [--max-iterations N]

Detects stale worker panes by hashing `tmux capture-pane` output for agent-1..agent-4.
When a session hash is unchanged for N consecutive checks, sends a notification to agent-0.

Options:
  --interval N         Seconds between checks (default: 20)
  --stale-cycles N     Unchanged cycles required before notify (default: 3)
  --reminder-cycles N  Additional unchanged cycles between repeat alerts (default: 3)
  --max-iterations N   Stop after N loops; 0 runs forever (default: 0)
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --interval)
      INTERVAL_SECONDS="$2"
      shift 2
      ;;
    --stale-cycles)
      STALE_CYCLES="$2"
      shift 2
      ;;
    --reminder-cycles)
      REMINDER_CYCLES="$2"
      shift 2
      ;;
    --max-iterations)
      MAX_ITERATIONS="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if ! command -v tmux >/dev/null 2>&1; then
  echo "tmux is required but not installed." >&2
  exit 1
fi

if ! [[ "${INTERVAL_SECONDS}" =~ ^[0-9]+$ ]] || ! [[ "${STALE_CYCLES}" =~ ^[0-9]+$ ]] || \
  ! [[ "${REMINDER_CYCLES}" =~ ^[0-9]+$ ]] || ! [[ "${MAX_ITERATIONS}" =~ ^[0-9]+$ ]]; then
  echo "interval, stale-cycles, reminder-cycles, and max-iterations must be non-negative integers" >&2
  exit 2
fi

if [[ "${INTERVAL_SECONDS}" -eq 0 ]]; then
  echo "interval must be >= 1" >&2
  exit 2
fi

if [[ "${STALE_CYCLES}" -eq 0 ]]; then
  echo "stale-cycles must be >= 1" >&2
  exit 2
fi

if [[ "${REMINDER_CYCLES}" -eq 0 ]]; then
  echo "reminder-cycles must be >= 1" >&2
  exit 2
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SEND_TO_AGENT0="${SCRIPT_DIR}/send_to_agent0.sh"

if [[ ! -x "${SEND_TO_AGENT0}" ]]; then
  echo "send helper missing or not executable: ${SEND_TO_AGENT0}" >&2
  exit 1
fi

hash_text() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum | awk '{print $1}'
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 | awk '{print $1}'
  else
    openssl dgst -sha256 | awk '{print $2}'
  fi
}

declare -a last_hashes
declare -a unchanged_counts
declare -a notified_flags
declare -a reminder_progress

for i in "${!WATCH_SESSIONS[@]}"; do
  last_hashes[$i]=""
  unchanged_counts[$i]=0
  notified_flags[$i]=0
  reminder_progress[$i]=0
done

iteration=0
while :; do
  for i in "${!WATCH_SESSIONS[@]}"; do
    session="${WATCH_SESSIONS[$i]}"

    if ! tmux has-session -t "${session}" 2>/dev/null; then
      continue
    fi

    pane_data="$(tmux capture-pane -p -S -200 -t "${session}" 2>/dev/null || true)"
    current_hash="$(printf '%s' "${pane_data}" | hash_text)"
    previous_hash="${last_hashes[$i]}"

    if [[ -n "${previous_hash}" ]] && [[ "${current_hash}" == "${previous_hash}" ]]; then
      unchanged_counts[$i]=$((unchanged_counts[$i] + 1))
    else
      if [[ "${notified_flags[$i]}" -eq 1 ]]; then
        "${SEND_TO_AGENT0}" "${FROM_AGENT}" status \
          "${session} pane changed again after stale period; worker appears active"
      fi
      unchanged_counts[$i]=0
      notified_flags[$i]=0
      reminder_progress[$i]=0
    fi

    if [[ "${notified_flags[$i]}" -eq 0 ]] && [[ "${unchanged_counts[$i]}" -ge "${STALE_CYCLES}" ]]; then
      "${SEND_TO_AGENT0}" "${FROM_AGENT}" status \
        "${session} appears stuck: pane unchanged for ${unchanged_counts[$i]} cycles (${INTERVAL_SECONDS}s interval)"
      notified_flags[$i]=1
      reminder_progress[$i]=0
    elif [[ "${notified_flags[$i]}" -eq 1 ]]; then
      reminder_progress[$i]=$((reminder_progress[$i] + 1))
      if [[ "${reminder_progress[$i]}" -ge "${REMINDER_CYCLES}" ]]; then
        "${SEND_TO_AGENT0}" "${FROM_AGENT}" status \
          "${session} still unchanged (possible stall): ${unchanged_counts[$i]} cycles total"
        reminder_progress[$i]=0
      fi
    fi

    last_hashes[$i]="${current_hash}"
  done

  iteration=$((iteration + 1))
  if [[ "${MAX_ITERATIONS}" -gt 0 ]] && [[ "${iteration}" -ge "${MAX_ITERATIONS}" ]]; then
    break
  fi

  sleep "${INTERVAL_SECONDS}"
done

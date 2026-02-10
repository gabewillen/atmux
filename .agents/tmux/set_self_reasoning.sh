#!/usr/bin/env bash

set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <low|medium|high|extra-high|1|2|3|4>" >&2
  exit 1
fi

if ! command -v tmux >/dev/null 2>&1; then
  echo "tmux is required but not installed." >&2
  exit 1
fi

if [[ -z "${TMUX:-}" ]]; then
  echo "must be run from inside a tmux session to auto-detect self session" >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SET_AGENT_REASONING="${SCRIPT_DIR}/set_agent_reasoning.sh"

if [[ ! -x "${SET_AGENT_REASONING}" ]]; then
  echo "missing executable script: ${SET_AGENT_REASONING}" >&2
  exit 1
fi

current_session="$(tmux display-message -p '#S')"
if [[ -z "${current_session}" ]]; then
  echo "unable to detect current tmux session" >&2
  exit 1
fi

"${SET_AGENT_REASONING}" "${current_session}" "$1"

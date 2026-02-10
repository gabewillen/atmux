#!/usr/bin/env bash

set -euo pipefail

if ! command -v tmux >/dev/null 2>&1; then
  echo "tmux is required but not installed." >&2
  exit 1
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BASE_AGENT_RULES="${ROOT_DIR}/AGENTS.md"
BASE_REF="${BASE_REF:-main}"
WORKTREES_DIR="${ROOT_DIR}/worktrees"
MONITOR_SCRIPT="${ROOT_DIR}/.agents/tmux/monitor_agent_activity.sh"
MONITOR_LOG="${ROOT_DIR}/.agents/tmux/monitor_agent_activity.log"
LLAMA_CPP_SRC="${ROOT_DIR}/tmp/llama.cpp"

mkdir -p "${WORKTREES_DIR}"

materialize_agent_md() {
  local session="$1"
  local workdir="$2"
  local overlay="${ROOT_DIR}/.agents/${session}/AGENT.diff"
  local target="${workdir}/AGENT.md"

  if [[ ! -f "${overlay}" ]]; then
    echo "warning: ${overlay} not found; skipping AGENT.md generation for ${session}" >&2
    return 0
  fi

  cp "${BASE_AGENT_RULES}" "${target}"
  if ! patch -s -u "${target}" < "${overlay}"; then
    echo "error: failed to apply ${overlay} to ${target}" >&2
    return 1
  fi
}

ensure_worktree() {
  local session="$1"
  local workdir="$2"
  local branch="$3"

  if [[ -d "${workdir}" ]]; then
    return 0
  fi

  if git -C "${ROOT_DIR}" show-ref --verify --quiet "refs/heads/${branch}"; then
    git -C "${ROOT_DIR}" worktree add "${workdir}" "${branch}"
  else
    git -C "${ROOT_DIR}" worktree add -b "${branch}" "${workdir}" "${BASE_REF}"
  fi
}

launch_codex_for_session() {
  local session="$1"
  local workdir="$2"
  local config_file="${ROOT_DIR}/.agents/${session}/codex.env"

  local codex_model="gpt-5.3-codex"
  local codex_reasoning_effort="high"
  local codex_search="true"
  local codex_approval_policy="never"
  local codex_sandbox_mode="danger-full-access"
  local codex_extra_flags=""

  if [[ -f "${config_file}" ]]; then
    # shellcheck source=/dev/null
    source "${config_file}"
  fi

  local search_flag=""
  if [[ "${codex_search}" == "true" ]]; then
    search_flag="--search"
  fi

  tmux send-keys -t "${session}" \
    "cd \"${workdir}\" && codex --model ${codex_model} -c model_reasoning_effort=\\\"${codex_reasoning_effort}\\\" --sandbox ${codex_sandbox_mode} --ask-for-approval ${codex_approval_policy} ${search_flag} ${codex_extra_flags}" \
    C-m
}

ensure_llama_cpp_hardlinks() {
  local workdir="$1"
  local target_tmp="${workdir}/tmp"
  local target_llama="${target_tmp}/llama.cpp"

  if [[ ! -d "${LLAMA_CPP_SRC}" ]]; then
    return 0
  fi

  mkdir -p "${target_tmp}"
  rm -rf "${target_llama}"
  cp -al "${LLAMA_CPP_SRC}" "${target_llama}"
}

ensure_agents_dir_link() {
  local workdir="$1"
  local target_agents="${workdir}/.agents"

  if [[ -L "${target_agents}" ]]; then
    rm -f "${target_agents}"
  elif [[ -d "${target_agents}" ]]; then
    rm -rf "${target_agents}"
  elif [[ -e "${target_agents}" ]]; then
    rm -f "${target_agents}"
  fi

  ln -s "${ROOT_DIR}/.agents" "${target_agents}"
}

start_agent0_monitor_background() {
  if [[ ! -x "${MONITOR_SCRIPT}" ]]; then
    echo "warning: monitor script missing or not executable: ${MONITOR_SCRIPT}" >&2
    return 0
  fi

  if pgrep -f "${MONITOR_SCRIPT}" >/dev/null 2>&1; then
    pkill -f "${MONITOR_SCRIPT}" || true
    sleep 0.2
  fi

  nohup "${MONITOR_SCRIPT}" --interval 20 --stale-cycles 3 --reminder-cycles 3 \
    >"${MONITOR_LOG}" 2>&1 &
  echo "monitor running in background (pid: $!, log: ${MONITOR_LOG})"
}

for session in agent-1 agent-2 agent-3 agent-4 agent-0; do
  intended_workdir=""
  branch=""
  case "${session}" in
    agent-0)
      intended_workdir="${WORKTREES_DIR}/agent-0"
      branch="agent-0/coordination"
      ;;
    agent-1)
      intended_workdir="${WORKTREES_DIR}/agent-1"
      branch="agent-1/core-api-runtime"
      ;;
    agent-2)
      intended_workdir="${WORKTREES_DIR}/agent-2"
      branch="agent-2/inference-session-generation"
      ;;
    agent-3)
      intended_workdir="${WORKTREES_DIR}/agent-3"
      branch="agent-3/state-io-recovery"
      ;;
    agent-4)
      intended_workdir="${WORKTREES_DIR}/agent-4"
      branch="agent-4/app-integration"
      ;;
    *)
      intended_workdir="${ROOT_DIR}"
      branch="main"
      ;;
  esac

  ensure_worktree "${session}" "${intended_workdir}" "${branch}"
  workdir="${intended_workdir}"
  ensure_llama_cpp_hardlinks "${workdir}"
  ensure_agents_dir_link "${workdir}"
  materialize_agent_md "${session}" "${workdir}"

  if tmux has-session -t "${session}" 2>/dev/null; then
    tmux kill-session -t "${session}"
  fi

  tmux new-session -d -s "${session}" -c "${workdir}"
  launch_codex_for_session "${session}" "${workdir}"

done

echo "tmux sessions ready: agent-0 agent-1 agent-2 agent-3 agent-4"
echo "attach with: tmux attach -t agent-0"
start_agent0_monitor_background

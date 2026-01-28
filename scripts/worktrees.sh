#!/usr/bin/env bash
set -euo pipefail

# Manage git worktrees for development.
# Usage: scripts/worktrees.sh [rebase|delete|init] [args]

main_branch="main"
repo_root="$(git rev-parse --show-toplevel)"
vscode_root="/Users/gabrielwillen/VSCode"

# Hardcoded list of branches for worktrees
branches=(
  "antigravity-gemini-3.0-pro-high"
  "claude-code-opus"
  "codex-gpt-5.2-codex-high"
  "copilot-claude-sonnet-4"
  "cursor-cli-auto"
  "gemini-cli-auto"
  "open-code-big-pickle"
  "qoder-auto"
  "qwen-code-coder-model"
)

function list_worktrees() {
  git worktree list --porcelain | awk '
    /^worktree / { path = $2 }
    /^branch /   { gsub("refs/heads/", "", $2); print path " " $2 }
  '
}

function do_rebase() {
  local target_branch="${1:-$main_branch}"
  echo "Rebasing all worktrees onto: $target_branch"
  git fetch --all --prune

  list_worktrees | while IFS=' ' read -r path branch; do
    [ -z "$path" ] && continue
    if [ "$path" = "$repo_root" ]; then continue; fi

    printf '\n=== Rebasing worktree %s (branch %s) onto %s ===\n' "$path" "$branch" "$target_branch"
    (
      cd "$path" || exit 1
      git status --short
      git rebase "$target_branch"
    )
  done
}

function do_delete() {
  echo "Deleting all worktrees and branches (except $main_branch)..."
  
  # Delete worktrees
  list_worktrees | while IFS=' ' read -r path branch; do
    [ -z "$path" ] && continue
    if [ "$path" = "$repo_root" ]; then continue; fi
    
    echo "Removing worktree: $path"
    git worktree remove --force "$path"
  done

  # Delete specified branches
  for branch in "${branches[@]}"; do
    if git rev-parse --verify "$branch" >/dev/null 2>&1; then
      echo "Deleting branch: $branch"
      git branch -D "$branch"
    else
      echo "Branch $branch does not exist, skipping..."
    fi
  done
}

function do_init() {
  echo "Initializing worktrees for specified branches..."
  
  for branch in "${branches[@]}"; do
    local worktree_path="$vscode_root/$branch"
    echo "Creating worktree for $branch at $worktree_path"
    
    if [ -d "$worktree_path" ]; then
      echo "  Directory already exists, skipping..."
      continue
    fi

    # Check if branch exists before adding worktree
    if ! git rev-parse --verify "$branch" >/dev/null 2>&1; then
       echo "  Branch $branch does not exist, creating from $main_branch..."
       git branch "$branch" "$main_branch"
    fi

    git worktree add "$worktree_path" "$branch"
    
    echo "  Updating submodules..."
    (
      cd "$worktree_path" || exit 1
      git submodule update --init --recursive
    )
  done
}

command="${1:-help}"
shift || true

case "$command" in
  rebase)
    do_rebase "$@"
    ;;
  delete)
    do_delete
    ;;
  init)
    do_init
    ;;
  *)
    echo "Usage: $0 {rebase|delete|init}"
    echo "  rebase [branch] - Rebase all worktrees onto branch (default: $main_branch)"
    echo "  delete          - Delete all worktrees and branches (except $main_branch)"
    echo "  init            - Create worktrees for specified branches in $vscode_root and init submodules"
    exit 1
    ;;
esac

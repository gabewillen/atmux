#!/usr/bin/env bash
[ -n "$BASH_VERSION" ] || exec bash "$0" "$@"
set -euo pipefail

# Manage git worktrees for development.
# Usage: scripts/worktrees.sh [rebase|delete|init] [args]

main_branch="main"
repo_root="$(git rev-parse --show-toplevel)"
vscode_root="/shared"

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

function aggressive_rm() {
  local path="$1"
  [ -d "$path" ] || return 0
  echo "  Aggressively removing: $path"
  # Attempt to unlock and make writable
  chmod -R +w "$path" 2>/dev/null || true
  # Try to delete
  if ! rm -rf "$path" 2>/dev/null; then
    # If it still fails, try to delete individual files to reveal more info (don't fail the script)
    find "$path" -mindepth 1 -delete 2>/dev/null || true
    rmdir "$path" 2>/dev/null || echo "  Warning: Could not fully delete $path"
  fi
}

function do_delete() {
  echo "Deleting all worktrees and branches (except $main_branch)..."
  
  # 1. Delete and clean up registered worktrees
  list_worktrees | while IFS=' ' read -r path branch; do
    [ -z "$path" ] && continue
    if [ "$path" = "$repo_root" ]; then continue; fi
    
    echo "Removing worktree: $path (branch $branch)"
    git worktree remove --force "$path" 2>/dev/null || true
    aggressive_rm "$path"
  done

  # 2. Delete specified branches
  for branch in "${branches[@]}"; do
    if git rev-parse --verify "$branch" >/dev/null 2>&1; then
      echo "Deleting branch: $branch"
      git branch -D "$branch" || echo "  Failed to delete branch $branch"
    fi
  done

  # 3. Clean up orphaned directories in VSCode root
  for branch in "${branches[@]}"; do
    for path in "$vscode_root/$branch" "$vscode_root/amux-$branch"; do
      if [ -d "$path" ]; then
        echo "Cleaning up orphaned directory: $path"
        aggressive_rm "$path"
      fi
    done
  done

  # 4. Clean up the old .worktrees folder if it exists
  if [ -d "$repo_root/.worktrees" ]; then
    echo "Cleaning up legacy .worktrees folder..."
    aggressive_rm "$repo_root/.worktrees"
  fi
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

#!/usr/bin/env bash
set -euo pipefail

# Recursively rebase each worktree branch onto the main branch (default: main).
# Usage: scripts/rebase-worktrees.sh [main-branch-name]

main_branch="${1:-main}"

# Ensure we are at the git repo root
repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

echo "Using main branch: $main_branch"

git fetch --all --prune

# Collect worktree path + branch pairs using porcelain format for robustness
# Use POSIX-compatible shell constructs so this script works even when invoked via `sh`.
worktrees="$(
  git worktree list --porcelain | awk '
    /^worktree / { path = $2 }
    /^branch /   { gsub("refs/heads/", "", $2); print path " " $2 }
  '
)"

# Iterate over each "path branch" pair (one per line)
printf '%s
' "$worktrees" | while IFS=' ' read -r path branch; do
  # Skip empty lines (defensive)
  [ -z "$path" ] && continue

  # Skip the primary worktree (usually the repo root)
  if [ "$path" = "$repo_root" ]; then
    continue
  fi

  printf '\n=== Rebasing worktree %s (branch %s) onto %s ===\n' "$path" "$branch" "$main_branch"
  (
    cd "$path" || exit 1
    git status --short
    git rebase "$main_branch"
  )

done

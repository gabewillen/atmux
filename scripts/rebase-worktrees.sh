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
mapfile -t worktrees < <(git worktree list --porcelain | awk '
  /^worktree / { path = $2 }
  /^branch /   { gsub("refs/heads/", "", $2); print path " " $2 }
')

for entry in "${worktrees[@]}"; do
  path="${entry%% *}"
  branch="${entry##* }"

  # Skip the primary worktree (usually the repo root)
  if [[ "$path" == "$repo_root" ]]; then
    continue
  fi

  echo "\n=== Rebasing worktree $path (branch $branch) onto $main_branch ==="
  (
    cd "$path"
    git status --short
    git rebase "$main_branch"
  )

done

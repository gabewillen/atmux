// Package git provides git repository queries and merge-strategy helpers for amux.
// Used for repo validation, base_branch resolution, and merge strategy selection (spec §5.3.4, §5.7, §5.7.1).
package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// IsRepo reports whether dir is the root of a git repository.
// It runs "git rev-parse --is-inside-work-tree" from dir.
// Non-zero exit (e.g. not a repo, missing dir) is treated as not a repo.
func IsRepo(dir string) (bool, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return false, fmt.Errorf("absolute path: %w", err)
	}
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	s := strings.TrimSpace(string(out))
	if err != nil {
		if s == "false" {
			return false, nil
		}
		return false, nil // Not a repo or missing dir
	}
	return s == "true", nil
}

// Root returns the repository root directory containing dir, or empty string if not in a repo.
func Root(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("absolute path: %w", err)
	}
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --show-toplevel: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// BaseBranch returns the current branch name in repoRoot (spec §5.7.1).
// It runs "git symbolic-ref --quiet --short HEAD" and returns the trimmed output.
// If the command fails (detached HEAD or unborn branch), returns empty string and a non-nil error.
func BaseBranch(repoRoot string) (string, error) {
	repoRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return "", fmt.Errorf("absolute path: %w", err)
	}
	cmd := exec.Command("git", "symbolic-ref", "--quiet", "--short", "HEAD")
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git symbolic-ref: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// ValidMergeStrategies are the supported git merge strategies (spec §5.7.2).
var ValidMergeStrategies = []string{"merge-commit", "squash", "rebase", "ff-only"}

// ValidStrategy returns true if s is one of merge-commit, squash, rebase, ff-only.
func ValidStrategy(s string) bool {
	for _, v := range ValidMergeStrategies {
		if s == v {
			return true
		}
	}
	return false
}

// ResolveTargetBranch returns the branch to merge into (spec §5.7.1).
// If configuredTarget is non-empty, returns it; otherwise returns base_branch from repoRoot.
// If base_branch cannot be determined (detached HEAD), returns an error instructing the user to set git.merge.target_branch.
func ResolveTargetBranch(repoRoot, configuredTarget string) (string, error) {
	if configuredTarget != "" {
		return strings.TrimSpace(configuredTarget), nil
	}
	base, err := BaseBranch(repoRoot)
	if err != nil {
		return "", fmt.Errorf("could not determine base_branch (detached HEAD or unborn branch): set git.merge.target_branch in config: %w", err)
	}
	return base, nil
}

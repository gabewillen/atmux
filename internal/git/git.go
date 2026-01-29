// Package git implements git operations and merge strategies for the amux project
package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// MergeStrategy represents different git merge strategies
type MergeStrategy string

const (
	MergeCommit MergeStrategy = "merge-commit"
	Squash      MergeStrategy = "squash"
	Rebase      MergeStrategy = "rebase"
	FFOnly      MergeStrategy = "ff-only"
)

// MergeOptions holds options for git merge operations
type MergeOptions struct {
	Strategy    MergeStrategy
	BaseBranch  string // Source branch to merge from
	TargetBranch string // Target branch to merge into
	DryRun      bool   // If true, only show what would be done
}

// PerformMerge executes a git merge operation based on the specified strategy
func PerformMerge(repoPath string, opts MergeOptions) error {
	if opts.DryRun {
		return simulateMerge(repoPath, opts)
	}

	switch opts.Strategy {
	case MergeCommit:
		return performMergeCommit(repoPath, opts)
	case Squash:
		return performSquash(repoPath, opts)
	case Rebase:
		return performRebase(repoPath, opts)
	case FFOnly:
		return performFFOnly(repoPath, opts)
	default:
		return fmt.Errorf("unknown merge strategy: %s", opts.Strategy)
	}
}

// simulateMerge shows what would happen with the merge without actually performing it
func simulateMerge(repoPath string, opts MergeOptions) error {
	fmt.Printf("DRY RUN: Would perform %s merge from %s to %s in %s\n", 
		opts.Strategy, opts.BaseBranch, opts.TargetBranch, repoPath)
	return nil
}

// performMergeCommit performs a standard merge-commit operation
func performMergeCommit(repoPath string, opts MergeOptions) error {
	// Switch to target branch
	cmd := exec.Command("git", "checkout", opts.TargetBranch)
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout target branch %s: %w", opts.TargetBranch, err)
	}

	// Perform merge-commit
	cmd = exec.Command("git", "merge", "--no-ff", opts.BaseBranch)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("merge-commit failed: %w, output: %s", err, string(output))
	}

	return nil
}

// performSquash performs a squash merge operation
func performSquash(repoPath string, opts MergeOptions) error {
	// Switch to target branch
	cmd := exec.Command("git", "checkout", opts.TargetBranch)
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout target branch %s: %w", opts.TargetBranch, err)
	}

	// Perform squash merge
	cmd = exec.Command("git", "merge", "--squash", opts.BaseBranch)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("squash merge failed: %w, output: %s", err, string(output))
	}

	// Commit the squashed changes
	cmd = exec.Command("git", "commit", "-m", fmt.Sprintf("Squash merge from %s", opts.BaseBranch))
	cmd.Dir = repoPath
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("commit after squash failed: %w, output: %s", err, string(output))
	}

	return nil
}

// performRebase performs a rebase operation
func performRebase(repoPath string, opts MergeOptions) error {
	// Switch to base branch
	cmd := exec.Command("git", "checkout", opts.BaseBranch)
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout base branch %s: %w", opts.BaseBranch, err)
	}

	// Perform rebase onto target branch
	cmd = exec.Command("git", "rebase", opts.TargetBranch)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rebase failed: %w, output: %s", err, string(output))
	}

	// Switch back to target branch
	cmd = exec.Command("git", "checkout", opts.TargetBranch)
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout target branch %s: %w", opts.TargetBranch, err)
	}

	// Fast-forward merge to incorporate rebased changes
	cmd = exec.Command("git", "merge", opts.BaseBranch)
	cmd.Dir = repoPath
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("fast-forward merge after rebase failed: %w, output: %s", err, string(output))
	}

	return nil
}

// performFFOnly performs a fast-forward only merge
func performFFOnly(repoPath string, opts MergeOptions) error {
	// Switch to target branch
	cmd := exec.Command("git", "checkout", opts.TargetBranch)
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout target branch %s: %w", opts.TargetBranch, err)
	}

	// Perform fast-forward only merge
	cmd = exec.Command("git", "merge", "--ff-only", opts.BaseBranch)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("fast-forward only merge failed: %w, output: %s", err, string(output))
	}

	return nil
}

// GetCurrentBranch returns the current git branch name
func GetCurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetBaseBranch determines the base branch for a repository
// Following the spec: run `git symbolic-ref --quiet --short HEAD` and use the output
// If that fails, use targetBranch if provided, otherwise return an error
func GetBaseBranch(repoPath string, targetBranch string) (string, error) {
	branch, err := GetCurrentBranch(repoPath)
	if err != nil {
		if targetBranch != "" {
			return targetBranch, nil
		}
		return "", fmt.Errorf("failed to determine base branch and no target branch provided: %w", err)
	}
	return branch, nil
}
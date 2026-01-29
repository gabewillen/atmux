// Package merge provides git merge strategy implementation.
// This package handles git merge strategies per spec requirements.
package merge

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Common sentinel errors for merge operations.
var (
	// ErrMergeConflict indicates a merge conflict occurred.
	ErrMergeConflict = errors.New("merge conflict")

	// ErrInvalidStrategy indicates an invalid merge strategy.
	ErrInvalidStrategy = errors.New("invalid merge strategy")

	// ErrGitCommandFailed indicates a git command failed.
	ErrGitCommandFailed = errors.New("git command failed")
)

// Strategy represents a git merge strategy.
type Strategy string

const (
	// StrategyMergeCommit creates a merge commit.
	StrategyMergeCommit Strategy = "merge-commit"

	// StrategySquash squashes commits before merging.
	StrategySquash Strategy = "squash"

	// StrategyRebase rebases before merging.
	StrategyRebase Strategy = "rebase"

	// StrategyFastForwardOnly only allows fast-forward merges.
	StrategyFastForwardOnly Strategy = "ff-only"
)

// Config contains merge strategy configuration.
type Config struct {
	Strategy     Strategy `toml:"strategy"`
	BaseBranch   string   `toml:"base_branch"`
	TargetBranch string   `toml:"target_branch"`
}

// DefaultConfig returns the default merge configuration.
func DefaultConfig() Config {
	return Config{
		Strategy:     StrategyMergeCommit,
		BaseBranch:   "main",
		TargetBranch: "main",
	}
}

// Validate checks if the merge config is valid.
func (c Config) Validate() error {
	switch c.Strategy {
	case StrategyMergeCommit, StrategySquash, StrategyRebase, StrategyFastForwardOnly:
		// Valid strategy
	default:
		return fmt.Errorf("invalid strategy %q: %w", c.Strategy, ErrInvalidStrategy)
	}

	if c.BaseBranch == "" {
		return fmt.Errorf("base_branch required: %w", ErrInvalidStrategy)
	}

	if c.TargetBranch == "" {
		return fmt.Errorf("target_branch required: %w", ErrInvalidStrategy)
	}

	return nil
}

// ExecuteMerge performs a merge using the specified strategy.
func ExecuteMerge(repoRoot, fromBranch string, config Config) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid merge config: %w", err)
	}

	switch config.Strategy {
	case StrategyMergeCommit:
		return executeMergeCommit(repoRoot, fromBranch, config.TargetBranch)
	case StrategySquash:
		return executeSquash(repoRoot, fromBranch, config.TargetBranch)
	case StrategyRebase:
		return executeRebase(repoRoot, fromBranch, config.TargetBranch)
	case StrategyFastForwardOnly:
		return executeFastForwardOnly(repoRoot, fromBranch, config.TargetBranch)
	default:
		return fmt.Errorf("unsupported strategy %q: %w", config.Strategy, ErrInvalidStrategy)
	}
}

// executeMergeCommit performs a merge commit.
func executeMergeCommit(repoRoot, fromBranch, toBranch string) error {
	// Checkout target branch
	cmd := exec.Command("git", "checkout", toBranch)
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to checkout %s: %s: %w", toBranch, string(output), ErrGitCommandFailed)
	}

	// Merge with commit
	cmd = exec.Command("git", "merge", "--no-ff", fromBranch)
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		if strings.Contains(string(output), "CONFLICT") {
			return fmt.Errorf("merge conflict: %s: %w", string(output), ErrMergeConflict)
		}
		return fmt.Errorf("failed to merge: %s: %w", string(output), ErrGitCommandFailed)
	}

	return nil
}

// executeSquash performs a squash merge.
func executeSquash(repoRoot, fromBranch, toBranch string) error {
	// Checkout target branch
	cmd := exec.Command("git", "checkout", toBranch)
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to checkout %s: %s: %w", toBranch, string(output), ErrGitCommandFailed)
	}

	// Squash merge
	cmd = exec.Command("git", "merge", "--squash", fromBranch)
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		if strings.Contains(string(output), "CONFLICT") {
			return fmt.Errorf("merge conflict: %s: %w", string(output), ErrMergeConflict)
		}
		return fmt.Errorf("failed to squash merge: %s: %w", string(output), ErrGitCommandFailed)
	}

	// Commit the squashed changes
	cmd = exec.Command("git", "commit", "-m", fmt.Sprintf("Squash merge from %s", fromBranch))
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to commit squash: %s: %w", string(output), ErrGitCommandFailed)
	}

	return nil
}

// executeRebase performs a rebase merge.
func executeRebase(repoRoot, fromBranch, toBranch string) error {
	// Checkout source branch
	cmd := exec.Command("git", "checkout", fromBranch)
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to checkout %s: %s: %w", fromBranch, string(output), ErrGitCommandFailed)
	}

	// Rebase onto target
	cmd = exec.Command("git", "rebase", toBranch)
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		if strings.Contains(string(output), "CONFLICT") {
			return fmt.Errorf("rebase conflict: %s: %w", string(output), ErrMergeConflict)
		}
		return fmt.Errorf("failed to rebase: %s: %w", string(output), ErrGitCommandFailed)
	}

	// Checkout target and fast-forward
	cmd = exec.Command("git", "checkout", toBranch)
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to checkout %s: %s: %w", toBranch, string(output), ErrGitCommandFailed)
	}

	cmd = exec.Command("git", "merge", "--ff-only", fromBranch)
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to fast-forward: %s: %w", string(output), ErrGitCommandFailed)
	}

	return nil
}

// executeFastForwardOnly performs a fast-forward only merge.
func executeFastForwardOnly(repoRoot, fromBranch, toBranch string) error {
	// Checkout target branch
	cmd := exec.Command("git", "checkout", toBranch)
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to checkout %s: %s: %w", toBranch, string(output), ErrGitCommandFailed)
	}

	// Fast-forward only merge
	cmd = exec.Command("git", "merge", "--ff-only", fromBranch)
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		if strings.Contains(string(output), "not possible to fast-forward") {
			return fmt.Errorf("fast-forward not possible: %s: %w", string(output), ErrMergeConflict)
		}
		return fmt.Errorf("failed to fast-forward: %s: %w", string(output), ErrGitCommandFailed)
	}

	return nil
}

// DryRun simulates a merge without actually performing it.
func DryRun(repoRoot, fromBranch string, config Config) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid merge config: %w", err)
	}

	// For dry run, we just validate that the branches exist
	branches := []string{fromBranch, config.BaseBranch, config.TargetBranch}
	for _, branch := range branches {
		cmd := exec.Command("git", "rev-parse", "--verify", branch)
		cmd.Dir = repoRoot
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("branch %q does not exist: %w", branch, ErrGitCommandFailed)
		}
	}

	return nil
}
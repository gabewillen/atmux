package agent

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/pkg/api"
)

// MergeStrategy represents a git merge strategy.
type MergeStrategy string

const (
	MergeMergeCommit MergeStrategy = "merge-commit"
	MergeSquash      MergeStrategy = "squash"
	MergeRebase      MergeStrategy = "rebase"
	MergeFFOnly      MergeStrategy = "ff-only"
)

// SelectMergeStrategy determines the merge strategy and target branch.
// It checks repo_root for current HEAD if target_branch is not configured.
func SelectMergeStrategy(cfg config.GitConfig, repoRoot api.RepoRoot) (MergeStrategy, string, error) {
	strategy := MergeStrategy(cfg.Merge.Strategy)
	if strategy == "" {
		strategy = MergeSquash // Default
	}

	targetBranch := cfg.Merge.TargetBranch
	if targetBranch == "" {
		// Determine base_branch from repo HEAD
		// "run git symbolic-ref --quiet --short HEAD in repo_root"
		cmd := exec.Command("git", "symbolic-ref", "--quiet", "--short", "HEAD")
		cmd.Dir = string(repoRoot)
		out, err := cmd.Output()
		if err != nil {
			return "", "", MessageError("failed to determine base branch from HEAD (detached?): %v. Please configure git.merge.target_branch", err)
		}
		targetBranch = strings.TrimSpace(string(out))
	}

	return strategy, targetBranch, nil
}

// ExecuteMerge performs the git integration of the source branch into the target branch
// using the specified strategy. It assumes the repo is clean if allowDirty is false.
func ExecuteMerge(repoRoot api.RepoRoot, strategy MergeStrategy, targetBranch string, sourceBranch string, allowDirty bool) error {
	// 1. Check for dirty state if required
	if !allowDirty {
		if isDirty(repoRoot) {
			return MessageError("working tree is dirty and allow_dirty is false")
		}
	}

	// 2. Checkout target branch
	if err := runGit(repoRoot, "checkout", targetBranch); err != nil {
		return MessageError("failed to checkout target branch %s: %v", targetBranch, err)
	}

	// 3. Perform merge strategy
	var err error
	switch strategy {
	case MergeMergeCommit:
		// git merge --no-ff source
		err = runGit(repoRoot, "merge", "--no-ff", sourceBranch)
	case MergeSquash:
		// git merge --squash source && git commit -m "Squash merge of {source}"
		if err = runGit(repoRoot, "merge", "--squash", sourceBranch); err == nil {
			// Commit the squash
			msg := fmt.Sprintf("Squash merge of %s", sourceBranch)
			err = runGit(repoRoot, "commit", "-m", msg)
		}
	case MergeRebase:
		// Rebase source onto target, then fast-forward target
		// 1. checkout source
		// 2. rebase target
		// 3. checkout target
		// 4. merge --ff-only source
		// Note: This modifies the source branch history.
		if err = runGit(repoRoot, "checkout", sourceBranch); err != nil {
			return err
		}
		if err = runGit(repoRoot, "rebase", targetBranch); err != nil {
			// Abort rebase on failure
			_ = runGit(repoRoot, "rebase", "--abort")
			return MessageError("rebase failed: %v", err)
		}
		if err = runGit(repoRoot, "checkout", targetBranch); err != nil {
			return err
		}
		err = runGit(repoRoot, "merge", "--ff-only", sourceBranch)

	case MergeFFOnly:
		// git merge --ff-only source
		err = runGit(repoRoot, "merge", "--ff-only", sourceBranch)
	
	default:
		return MessageError("unsupported merge strategy: %s", strategy)
	}

	if err != nil {
		return MessageError("merge operation (%s) failed: %v", strategy, err)
	}

	return nil
}

func isDirty(repoRoot api.RepoRoot) bool {
	// git status --porcelain
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = string(repoRoot)
	out, err := cmd.Output()
	if err != nil {
		return true // Assume dirty on error for safety
	}
	return len(strings.TrimSpace(string(out))) > 0
}

func runGit(repoRoot api.RepoRoot, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = string(repoRoot)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %s: %s (%w)", strings.Join(args, " "), string(out), err)
	}
	return nil
}

// MessageError creates a formatted error.
func MessageError(format string, a ...any) error {
	return fmt.Errorf(format, a...)
}

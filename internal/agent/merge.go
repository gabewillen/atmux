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
	MergeSquash MergeStrategy = "squash"
	MergeRebase MergeStrategy = "rebase"
	MergeFFOnly MergeStrategy = "ff-only"
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

// MessageError creates a formatted error.
func MessageError(format string, a ...any) error {
	return fmt.Errorf(format, a...)
}

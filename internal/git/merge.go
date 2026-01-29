package git

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var (
	// ErrDirtyWorktree is returned when a worktree has uncommitted changes.
	ErrDirtyWorktree = errors.New("dirty worktree")
	// ErrBranchMissing is returned when a required branch is missing.
	ErrBranchMissing = errors.New("branch missing")
	// ErrMergeConflict is returned when a merge conflict is detected.
	ErrMergeConflict = errors.New("merge conflict")
	// ErrDetachedHead is returned when base branch detection fails.
	ErrDetachedHead = errors.New("detached head")
	// ErrInvalidStrategy is returned for unsupported strategies.
	ErrInvalidStrategy = errors.New("invalid merge strategy")
)

// MergeStrategy identifies a supported merge strategy.
type MergeStrategy string

const (
	// StrategyMergeCommit performs a merge commit.
	StrategyMergeCommit MergeStrategy = "merge-commit"
	// StrategySquash performs a squash merge.
	StrategySquash MergeStrategy = "squash"
	// StrategyRebase rebases and fast-forwards.
	StrategyRebase MergeStrategy = "rebase"
	// StrategyFFOnly performs a fast-forward only merge.
	StrategyFFOnly MergeStrategy = "ff-only"
)

// MergeOptions configures a merge operation.
type MergeOptions struct {
	RepoRoot     string
	WorktreePath string
	AgentSlug    string
	Strategy     MergeStrategy
	TargetBranch string
	BaseBranch   string
	AllowDirty   bool
}

// MergeResult describes a merge attempt.
type MergeResult struct {
	TargetBranch string
	Strategy     MergeStrategy
}

// DetectBaseBranch determines the base branch for a repository.
func (r *Runner) DetectBaseBranch(ctx context.Context, repoRoot, fallback string) (string, error) {
	result, err := r.run(ctx, repoRoot, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err == nil {
		branch := strings.TrimSpace(string(result.Output))
		if branch != "" {
			return branch, nil
		}
	}
	if strings.TrimSpace(fallback) != "" {
		return fallback, nil
	}
	return "", fmt.Errorf("detect base branch: %w", ErrDetachedHead)
}

// Merge integrates the agent branch into the target branch.
func (r *Runner) Merge(ctx context.Context, opts MergeOptions) (MergeResult, error) {
	if strings.TrimSpace(opts.RepoRoot) == "" {
		return MergeResult{}, fmt.Errorf("merge: %w", ErrRepoRequired)
	}
	if strings.TrimSpace(opts.AgentSlug) == "" {
		return MergeResult{}, fmt.Errorf("merge: %w", ErrBranchMissing)
	}
	strategy := opts.Strategy
	if strategy == "" {
		strategy = StrategySquash
	}
	target := strings.TrimSpace(opts.TargetBranch)
	if target == "" {
		target = strings.TrimSpace(opts.BaseBranch)
	}
	if target == "" {
		return MergeResult{}, fmt.Errorf("merge: %w", ErrBranchMissing)
	}
	agentBranch := "amux/" + opts.AgentSlug
	if err := r.ensureBranchExists(ctx, opts.RepoRoot, target); err != nil {
		return MergeResult{}, fmt.Errorf("merge: %w", err)
	}
	if err := r.ensureBranchExists(ctx, opts.RepoRoot, agentBranch); err != nil {
		return MergeResult{}, fmt.Errorf("merge: %w", err)
	}
	worktreePath := opts.WorktreePath
	if worktreePath == "" {
		worktreePath = filepath.Join(opts.RepoRoot, ".amux", "worktrees", opts.AgentSlug)
	}
	if err := r.ensureClean(ctx, worktreePath, opts.AllowDirty); err != nil {
		return MergeResult{}, fmt.Errorf("merge: %w", err)
	}
	switch strategy {
	case StrategyMergeCommit:
		if err := r.mergeCommit(ctx, opts.RepoRoot, agentBranch, target); err != nil {
			return MergeResult{}, fmt.Errorf("merge: %w", err)
		}
	case StrategySquash:
		if err := r.mergeSquash(ctx, opts.RepoRoot, agentBranch, target); err != nil {
			return MergeResult{}, fmt.Errorf("merge: %w", err)
		}
	case StrategyRebase:
		if err := r.mergeRebase(ctx, worktreePath, opts.RepoRoot, agentBranch, target); err != nil {
			return MergeResult{}, fmt.Errorf("merge: %w", err)
		}
	case StrategyFFOnly:
		if err := r.mergeFFOnly(ctx, opts.RepoRoot, agentBranch, target); err != nil {
			return MergeResult{}, fmt.Errorf("merge: %w", err)
		}
	default:
		return MergeResult{}, fmt.Errorf("merge: %w", ErrInvalidStrategy)
	}
	return MergeResult{TargetBranch: target, Strategy: strategy}, nil
}

func (r *Runner) ensureBranchExists(ctx context.Context, repoRoot, branch string) error {
	result, err := r.run(ctx, repoRoot, "show-ref", "--verify", "refs/heads/"+branch)
	if err != nil {
		if isMissingRef(result.ExitCode, result.Output) {
			return fmt.Errorf("branch %s: %w", branch, ErrBranchMissing)
		}
		return fmt.Errorf("branch %s: %w", branch, err)
	}
	return nil
}

func (r *Runner) ensureClean(ctx context.Context, worktreePath string, allowDirty bool) error {
	if allowDirty {
		return nil
	}
	result, err := r.run(ctx, worktreePath, "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("worktree status: %w", err)
	}
	if strings.TrimSpace(string(result.Output)) != "" {
		return ErrDirtyWorktree
	}
	return nil
}

func (r *Runner) mergeCommit(ctx context.Context, repoRoot, agentBranch, target string) error {
	if _, err := r.run(ctx, repoRoot, "checkout", target); err != nil {
		return fmt.Errorf("checkout target: %w", err)
	}
	if _, err := r.run(ctx, repoRoot, "merge", "--no-ff", agentBranch); err != nil {
		if conflict := r.hasConflicts(ctx, repoRoot); conflict != nil {
			return conflict
		}
		return fmt.Errorf("merge commit: %w", err)
	}
	return nil
}

func (r *Runner) mergeSquash(ctx context.Context, repoRoot, agentBranch, target string) error {
	if _, err := r.run(ctx, repoRoot, "checkout", target); err != nil {
		return fmt.Errorf("checkout target: %w", err)
	}
	if _, err := r.run(ctx, repoRoot, "merge", "--squash", agentBranch); err != nil {
		if conflict := r.hasConflicts(ctx, repoRoot); conflict != nil {
			return conflict
		}
		return fmt.Errorf("merge squash: %w", err)
	}
	message := fmt.Sprintf("amux: squash %s", agentBranch)
	if _, err := r.run(ctx, repoRoot, "commit", "-m", message); err != nil {
		return fmt.Errorf("merge squash commit: %w", err)
	}
	return nil
}

func (r *Runner) mergeRebase(ctx context.Context, worktreePath, repoRoot, agentBranch, target string) error {
	current, err := r.run(ctx, worktreePath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return fmt.Errorf("rebase branch: %w", err)
	}
	branchName := strings.TrimSpace(string(current.Output))
	if branchName != agentBranch && branchName != strings.TrimPrefix(agentBranch, "refs/heads/") {
		return fmt.Errorf("rebase branch: %w", ErrBranchMissing)
	}
	if _, err := r.run(ctx, worktreePath, "rebase", target); err != nil {
		if conflict := r.hasConflicts(ctx, worktreePath); conflict != nil {
			return conflict
		}
		return fmt.Errorf("rebase: %w", err)
	}
	if _, err := r.run(ctx, repoRoot, "checkout", target); err != nil {
		return fmt.Errorf("checkout target: %w", err)
	}
	if _, err := r.run(ctx, repoRoot, "merge", "--ff-only", agentBranch); err != nil {
		return fmt.Errorf("rebase ff: %w", err)
	}
	return nil
}

func (r *Runner) mergeFFOnly(ctx context.Context, repoRoot, agentBranch, target string) error {
	if _, err := r.run(ctx, repoRoot, "checkout", target); err != nil {
		return fmt.Errorf("checkout target: %w", err)
	}
	if _, err := r.run(ctx, repoRoot, "merge", "--ff-only", agentBranch); err != nil {
		return fmt.Errorf("ff-only: %w", err)
	}
	return nil
}

func (r *Runner) hasConflicts(ctx context.Context, repoRoot string) error {
	result, err := r.run(ctx, repoRoot, "diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return fmt.Errorf("conflict check: %w", err)
	}
	if strings.TrimSpace(string(result.Output)) == "" {
		return nil
	}
	return ErrMergeConflict
}

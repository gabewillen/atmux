// Package gitmerge implements git merge strategy selection and execution for amux.
//
// This package specifies how changes made in an agent worktree branch
// (amux/{agent_slug}) are integrated into a target branch within the same
// repository. Merge execution is performed by running local git commands
// in the corresponding repo_root.
//
// Supported strategies: merge-commit, squash, rebase, ff-only.
//
// See spec §5.7 for merge strategy requirements.
package gitmerge

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/stateforward/hsm-go/muid"

	amuxerrors "github.com/agentflare-ai/amux/internal/errors"
	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/internal/worktree"
)

// Strategy represents a supported merge strategy.
type Strategy string

const (
	// StrategyMergeCommit creates a non-fast-forward merge commit.
	StrategyMergeCommit Strategy = "merge-commit"

	// StrategySquash squashes all commits into a single commit on target_branch.
	StrategySquash Strategy = "squash"

	// StrategyRebase rebases the agent branch onto target_branch and fast-forwards.
	StrategyRebase Strategy = "rebase"

	// StrategyFFOnly fast-forwards target_branch only if direct descendant.
	StrategyFFOnly Strategy = "ff-only"
)

// ValidStrategies returns all supported merge strategies.
func ValidStrategies() []Strategy {
	return []Strategy{StrategyMergeCommit, StrategySquash, StrategyRebase, StrategyFFOnly}
}

// ParseStrategy parses a strategy string. Returns an error for unsupported values.
func ParseStrategy(s string) (Strategy, error) {
	switch Strategy(s) {
	case StrategyMergeCommit, StrategySquash, StrategyRebase, StrategyFFOnly:
		return Strategy(s), nil
	default:
		return "", fmt.Errorf("%w: %q (supported: %v)", amuxerrors.ErrInvalidStrategy, s, ValidStrategies())
	}
}

// Request represents a merge integration request.
type Request struct {
	// RepoRoot is the absolute path to the repository root.
	RepoRoot string

	// AgentSlug is the agent's slug (branch is amux/{agent_slug}).
	AgentSlug string

	// Strategy is the merge strategy to use.
	Strategy Strategy

	// TargetBranch is the branch to merge into. If empty, uses base_branch.
	TargetBranch string

	// BaseBranch is the branch recorded when the first agent was added.
	BaseBranch string

	// AllowDirty permits merging from a dirty worktree.
	AllowDirty bool

	// AgentID is the agent's runtime ID for event emission.
	AgentID muid.MUID
}

// Result represents the result of a merge operation.
type Result struct {
	// Strategy is the strategy that was used.
	Strategy Strategy

	// TargetBranch is the branch that was merged into.
	TargetBranch string

	// SourceBranch is the agent branch that was merged from.
	SourceBranch string

	// CommitSHA is the resulting commit hash, if applicable.
	CommitSHA string

	// Conflict indicates merge conflicts were detected.
	Conflict bool
}

// Executor executes git merge operations.
type Executor struct {
	gitPath    string
	dispatcher event.Dispatcher
}

// NewExecutor creates a new merge executor.
func NewExecutor(dispatcher event.Dispatcher) *Executor {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		gitPath = "git"
	}
	if dispatcher == nil {
		dispatcher = event.NewNoopDispatcher()
	}
	return &Executor{gitPath: gitPath, dispatcher: dispatcher}
}

// Execute performs a merge operation per the given request.
//
// It validates preconditions, executes the merge using the selected strategy,
// and emits the appropriate events.
//
// See spec §5.7.2-§5.7.5 for strategy behavior, preconditions, and events.
func (e *Executor) Execute(ctx context.Context, req Request) (*Result, error) {
	// Resolve target branch
	targetBranch := req.TargetBranch
	if targetBranch == "" {
		targetBranch = req.BaseBranch
	}
	if targetBranch == "" {
		return nil, fmt.Errorf("merge: no target branch specified and no base branch recorded; set git.merge.target_branch")
	}

	sourceBranch := worktree.BranchName(req.AgentSlug)
	evtData := map[string]any{
		"repo_root":     req.RepoRoot,
		"agent_slug":    req.AgentSlug,
		"strategy":      string(req.Strategy),
		"target_branch": targetBranch,
		"source_branch": sourceBranch,
	}

	// Emit git.merge.requested
	_ = e.dispatcher.Dispatch(ctx, event.NewEvent(event.TypeGitMergeRequested, req.AgentID, evtData))

	// Validate preconditions (spec §5.7.3)
	if err := e.validatePreconditions(req, targetBranch, sourceBranch); err != nil {
		evtData["error"] = err.Error()
		_ = e.dispatcher.Dispatch(ctx, event.NewEvent(event.TypeGitMergeFailed, req.AgentID, evtData))
		return nil, err
	}

	// Execute the merge strategy
	result, err := e.executeStrategy(req, targetBranch, sourceBranch)
	if err != nil {
		// Check if it's a conflict
		if isConflictError(err) {
			evtData["error"] = err.Error()
			_ = e.dispatcher.Dispatch(ctx, event.NewEvent(event.TypeGitMergeConflict, req.AgentID, evtData))
			return &Result{
				Strategy:     req.Strategy,
				TargetBranch: targetBranch,
				SourceBranch: sourceBranch,
				Conflict:     true,
			}, fmt.Errorf("merge: %w: %v", amuxerrors.ErrMergeConflict, err)
		}

		evtData["error"] = err.Error()
		_ = e.dispatcher.Dispatch(ctx, event.NewEvent(event.TypeGitMergeFailed, req.AgentID, evtData))
		return nil, fmt.Errorf("merge: %w: %v", amuxerrors.ErrMergeFailed, err)
	}

	// Get the resulting commit SHA
	commitSHA := e.getHeadSHA(req.RepoRoot)
	evtData["commit"] = commitSHA

	_ = e.dispatcher.Dispatch(ctx, event.NewEvent(event.TypeGitMergeCompleted, req.AgentID, evtData))

	result.CommitSHA = commitSHA
	return result, nil
}

// validatePreconditions checks merge preconditions per spec §5.7.3.
func (e *Executor) validatePreconditions(req Request, targetBranch, sourceBranch string) error {
	// Validate repo_root is a git repository
	if !e.isGitRepo(req.RepoRoot) {
		return fmt.Errorf("precondition: %q is %w", req.RepoRoot, amuxerrors.ErrNotInRepository)
	}

	// Validate target branch exists
	if !e.branchExists(req.RepoRoot, targetBranch) {
		return fmt.Errorf("precondition: target branch %q: %w", targetBranch, amuxerrors.ErrBranchNotFound)
	}

	// Validate source branch exists
	if !e.branchExists(req.RepoRoot, sourceBranch) {
		return fmt.Errorf("precondition: source branch %q: %w", sourceBranch, amuxerrors.ErrBranchNotFound)
	}

	// Check for uncommitted changes in worktree unless allow_dirty
	if !req.AllowDirty {
		wtDir := filepath.Join(req.RepoRoot, ".amux", "worktrees", req.AgentSlug)
		dirty, err := isDirtyWorktree(e.gitPath, wtDir)
		if err == nil && dirty {
			return fmt.Errorf("precondition: %w", amuxerrors.ErrDirtyWorktree)
		}
	}

	return nil
}

// executeStrategy runs the appropriate git commands for the selected strategy.
func (e *Executor) executeStrategy(req Request, targetBranch, sourceBranch string) (*Result, error) {
	result := &Result{
		Strategy:     req.Strategy,
		TargetBranch: targetBranch,
		SourceBranch: sourceBranch,
	}

	switch req.Strategy {
	case StrategyMergeCommit:
		return result, e.doMergeCommit(req.RepoRoot, targetBranch, sourceBranch)
	case StrategySquash:
		return result, e.doSquash(req.RepoRoot, targetBranch, sourceBranch)
	case StrategyRebase:
		return result, e.doRebase(req.RepoRoot, targetBranch, sourceBranch)
	case StrategyFFOnly:
		return result, e.doFFOnly(req.RepoRoot, targetBranch, sourceBranch)
	default:
		return nil, fmt.Errorf("%w: %q", amuxerrors.ErrInvalidStrategy, req.Strategy)
	}
}

// doMergeCommit performs a non-fast-forward merge commit.
func (e *Executor) doMergeCommit(repoRoot, targetBranch, sourceBranch string) error {
	// Checkout target branch
	if err := e.gitCheckout(repoRoot, targetBranch); err != nil {
		return err
	}

	// Merge with --no-ff to force a merge commit
	cmd := exec.Command(e.gitPath, "merge", "--no-ff", sourceBranch,
		"-m", fmt.Sprintf("Merge %s into %s", sourceBranch, targetBranch))
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Abort the merge on failure
		_ = e.gitMergeAbort(repoRoot)
		return fmt.Errorf("merge-commit: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

// doSquash squashes all commits from source into a single commit on target.
func (e *Executor) doSquash(repoRoot, targetBranch, sourceBranch string) error {
	// Checkout target branch
	if err := e.gitCheckout(repoRoot, targetBranch); err != nil {
		return err
	}

	// Squash merge
	cmd := exec.Command(e.gitPath, "merge", "--squash", sourceBranch)
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		_ = e.gitMergeAbort(repoRoot)
		return fmt.Errorf("squash: %s", strings.TrimSpace(string(output)))
	}

	// Commit the squashed changes
	commitCmd := exec.Command(e.gitPath, "commit", "-m",
		fmt.Sprintf("Squash merge %s into %s", sourceBranch, targetBranch))
	commitCmd.Dir = repoRoot
	commitOutput, err := commitCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("squash commit: %s", strings.TrimSpace(string(commitOutput)))
	}
	return nil
}

// doRebase rebases source onto target and fast-forwards target.
//
// The source branch is checked out in its worktree, so the rebase runs inside
// the worktree directory rather than checking out the branch in the main repo
// (git prohibits checking out a branch that is already in a worktree).
func (e *Executor) doRebase(repoRoot, targetBranch, sourceBranch string) error {
	// Compute the worktree directory from the source branch name.
	slug := strings.TrimPrefix(sourceBranch, worktree.BranchPrefix)
	wtDir := filepath.Join(repoRoot, ".amux", "worktrees", slug)

	// Rebase the source branch onto target inside the worktree.
	cmd := exec.Command(e.gitPath, "rebase", targetBranch)
	cmd.Dir = wtDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		abortCmd := exec.Command(e.gitPath, "rebase", "--abort")
		abortCmd.Dir = wtDir
		_ = abortCmd.Run()
		return fmt.Errorf("rebase: %s", strings.TrimSpace(string(output)))
	}

	// Fast-forward target to rebased source in the main repo.
	if err := e.gitCheckout(repoRoot, targetBranch); err != nil {
		return err
	}

	ffCmd := exec.Command(e.gitPath, "merge", "--ff-only", sourceBranch)
	ffCmd.Dir = repoRoot
	ffOutput, err := ffCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rebase ff: %s", strings.TrimSpace(string(ffOutput)))
	}
	return nil
}

// doFFOnly fast-forwards target to source, failing if not a direct descendant.
func (e *Executor) doFFOnly(repoRoot, targetBranch, sourceBranch string) error {
	// Checkout target branch
	if err := e.gitCheckout(repoRoot, targetBranch); err != nil {
		return err
	}

	// Attempt fast-forward only merge
	cmd := exec.Command(e.gitPath, "merge", "--ff-only", sourceBranch)
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ff-only: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

// gitCheckout checks out a branch.
func (e *Executor) gitCheckout(repoRoot, branch string) error {
	cmd := exec.Command(e.gitPath, "checkout", branch)
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("checkout %q: %s", branch, strings.TrimSpace(string(output)))
	}
	return nil
}

// gitMergeAbort aborts a merge in progress.
func (e *Executor) gitMergeAbort(repoRoot string) error {
	cmd := exec.Command(e.gitPath, "merge", "--abort")
	cmd.Dir = repoRoot
	return cmd.Run()
}

// branchExists checks if a branch exists in the repository.
func (e *Executor) branchExists(repoRoot, branch string) bool {
	cmd := exec.Command(e.gitPath, "rev-parse", "--verify", "refs/heads/"+branch)
	cmd.Dir = repoRoot
	return cmd.Run() == nil
}

// getHeadSHA returns the current HEAD commit SHA.
func (e *Executor) getHeadSHA(repoRoot string) string {
	cmd := exec.Command(e.gitPath, "rev-parse", "HEAD")
	cmd.Dir = repoRoot
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// isConflictError checks if a git error indicates merge conflicts.
func isConflictError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "CONFLICT") ||
		strings.Contains(msg, "Automatic merge failed") ||
		strings.Contains(msg, "conflict")
}

// isDirtyWorktree checks if a worktree has uncommitted changes.
func isDirtyWorktree(gitPath, dir string) (bool, error) {
	cmd := exec.Command(gitPath, "status", "--porcelain")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(output))) > 0, nil
}

// isGitRepo returns true if the directory is a git repository root.
func (e *Executor) isGitRepo(dir string) bool {
	cmd := exec.Command(e.gitPath, "rev-parse", "--git-dir")
	cmd.Dir = dir
	return cmd.Run() == nil
}

package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	// ErrWorktreeConflict is returned when a worktree path is in use.
	ErrWorktreeConflict = errors.New("worktree path conflict")
)

// Worktree describes a git worktree entry.
type Worktree struct {
	Path     string
	Branch   string
	Detached bool
	Existing bool
}

// EnsureWorktree creates or reuses a worktree for the agent slug.
func (r *Runner) EnsureWorktree(ctx context.Context, repoRoot, agentSlug string) (Worktree, error) {
	if strings.TrimSpace(agentSlug) == "" {
		return Worktree{}, fmt.Errorf("ensure worktree: %w", ErrWorktreeConflict)
	}
	worktreePath := filepath.Join(repoRoot, ".amux", "worktrees", agentSlug)
	branch := "amux/" + agentSlug
	entries, err := r.ListWorktrees(ctx, repoRoot)
	if err != nil {
		return Worktree{}, fmt.Errorf("ensure worktree: %w", err)
	}
	if entry, ok := entries[worktreePath]; ok {
		entry.Existing = true
		return entry, nil
	}
	if info, err := os.Stat(worktreePath); err == nil && info.IsDir() {
		return Worktree{}, fmt.Errorf("ensure worktree: %w", ErrWorktreeConflict)
	}
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0o755); err != nil {
		return Worktree{}, fmt.Errorf("ensure worktree: %w", err)
	}
	exists, err := r.ensureBranch(ctx, repoRoot, branch)
	if err != nil {
		return Worktree{}, fmt.Errorf("ensure worktree: %w", err)
	}
	if !exists {
		addArgs := []string{"worktree", "add", "-b", branch, worktreePath}
		if _, addErr := r.run(ctx, repoRoot, addArgs...); addErr != nil {
			return Worktree{}, fmt.Errorf("ensure worktree: %w", addErr)
		}
		return Worktree{Path: worktreePath, Branch: branch}, nil
	}
	addArgs := []string{"worktree", "add", worktreePath, branch}
	if _, err := r.run(ctx, repoRoot, addArgs...); err != nil {
		return Worktree{}, fmt.Errorf("ensure worktree: %w", err)
	}
	return Worktree{Path: worktreePath, Branch: branch}, nil
}

// RemoveWorktree removes the worktree and optionally deletes the branch.
func (r *Runner) RemoveWorktree(ctx context.Context, repoRoot, agentSlug string, deleteBranch bool) error {
	if strings.TrimSpace(agentSlug) == "" {
		return fmt.Errorf("remove worktree: %w", ErrWorktreeConflict)
	}
	worktreePath := filepath.Join(repoRoot, ".amux", "worktrees", agentSlug)
	if _, err := r.run(ctx, repoRoot, "worktree", "remove", worktreePath); err != nil {
		return fmt.Errorf("remove worktree: %w", err)
	}
	if !deleteBranch {
		return nil
	}
	branch := "amux/" + agentSlug
	if _, err := r.run(ctx, repoRoot, "branch", "-D", branch); err != nil {
		return fmt.Errorf("remove worktree: %w", err)
	}
	return nil
}

// ListWorktrees returns worktrees keyed by path.
func (r *Runner) ListWorktrees(ctx context.Context, repoRoot string) (map[string]Worktree, error) {
	result, err := r.run(ctx, repoRoot, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("list worktrees: %w", err)
	}
	return parseWorktrees(string(result.Output)), nil
}

func parseWorktrees(output string) map[string]Worktree {
	entries := make(map[string]Worktree)
	lines := strings.Split(output, "\n")
	var current Worktree
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "worktree":
			if current.Path != "" {
				entries[current.Path] = current
			}
			current = Worktree{Path: strings.TrimSpace(strings.TrimPrefix(line, "worktree"))}
			current.Path = strings.TrimSpace(current.Path)
		case "branch":
			if len(fields) >= 2 {
				current.Branch = strings.TrimPrefix(fields[1], "refs/heads/")
			}
		case "detached":
			current.Detached = true
		}
	}
	if current.Path != "" {
		entries[current.Path] = current
	}
	return entries
}

func (r *Runner) ensureBranch(ctx context.Context, repoRoot, branch string) (bool, error) {
	if strings.TrimSpace(branch) == "" {
		return false, fmt.Errorf("ensure branch: %w", ErrWorktreeConflict)
	}
	result, err := r.run(ctx, repoRoot, "show-ref", "--verify", "refs/heads/"+branch)
	if err != nil {
		if isMissingRef(result.ExitCode, result.Output) {
			return false, nil
		}
		return false, fmt.Errorf("ensure branch: %w", err)
	}
	return true, nil
}

func isMissingRef(exitCode int, output []byte) bool {
	if exitCode != 1 && exitCode != 128 {
		return false
	}
	msg := strings.ToLower(string(output))
	return strings.Contains(msg, "not a valid ref") || strings.Contains(msg, "unknown revision") || strings.Contains(msg, "ambiguous argument")
}

package git

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestDetectBaseBranch(t *testing.T) {
	runner := &Runner{Exec: func(ctx context.Context, dir string, args ...string) (ExecResult, error) {
		if args[0] == "symbolic-ref" {
			return ExecResult{Output: []byte("main\n"), ExitCode: 0}, nil
		}
		return ExecResult{}, nil
	}}
	branch, err := runner.DetectBaseBranch(context.Background(), "/repo", "")
	if err != nil || branch != "main" {
		t.Fatalf("unexpected branch: %q %v", branch, err)
	}
}

func TestDetectBaseBranchFallback(t *testing.T) {
	runner := &Runner{Exec: func(ctx context.Context, dir string, args ...string) (ExecResult, error) {
		return ExecResult{ExitCode: 1}, errors.New("no head")
	}}
	branch, err := runner.DetectBaseBranch(context.Background(), "/repo", "develop")
	if err != nil || branch != "develop" {
		t.Fatalf("unexpected fallback: %q %v", branch, err)
	}
}

func TestDetectBaseBranchDetached(t *testing.T) {
	runner := &Runner{Exec: func(ctx context.Context, dir string, args ...string) (ExecResult, error) {
		return ExecResult{ExitCode: 1}, errors.New("no head")
	}}
	if _, err := runner.DetectBaseBranch(context.Background(), "/repo", ""); err == nil {
		t.Fatalf("expected detached head error")
	}
}

func TestMergeRebaseBranchMismatch(t *testing.T) {
	runner := &Runner{Exec: func(ctx context.Context, dir string, args ...string) (ExecResult, error) {
		if args[0] == "show-ref" {
			return ExecResult{ExitCode: 0}, nil
		}
		if args[0] == "status" {
			return ExecResult{ExitCode: 0}, nil
		}
		if args[0] == "rev-parse" {
			return ExecResult{ExitCode: 0, Output: []byte("other\n")}, nil
		}
		return ExecResult{ExitCode: 0}, nil
	}}
	if _, err := runner.Merge(context.Background(), MergeOptions{
		RepoRoot:     "/repo",
		WorktreePath: "/repo/.amux/worktrees/alpha",
		AgentSlug:    "alpha",
		Strategy:     StrategyRebase,
		BaseBranch:   "main",
	}); err == nil {
		t.Fatalf("expected rebase branch error")
	}
}

func TestMergeCommitConflictDetectsConflicts(t *testing.T) {
	runner := &Runner{Exec: func(ctx context.Context, dir string, args ...string) (ExecResult, error) {
		switch args[0] {
		case "show-ref":
			return ExecResult{ExitCode: 0}, nil
		case "status":
			return ExecResult{ExitCode: 0}, nil
		case "checkout":
			return ExecResult{ExitCode: 0}, nil
		case "merge":
			return ExecResult{ExitCode: 1}, errors.New("conflict")
		case "diff":
			return ExecResult{ExitCode: 0, Output: []byte("file.txt\n")}, nil
		}
		return ExecResult{ExitCode: 0}, nil
	}}
	if _, err := runner.Merge(context.Background(), MergeOptions{
		RepoRoot:     "/repo",
		WorktreePath: "/repo/.amux/worktrees/alpha",
		AgentSlug:    "alpha",
		Strategy:     StrategyMergeCommit,
		BaseBranch:   "main",
	}); err == nil || !strings.Contains(err.Error(), ErrMergeConflict.Error()) {
		t.Fatalf("expected merge conflict error, got %v", err)
	}
}

func TestMergeFFOnlySuccessExtra(t *testing.T) {
	calls := make([]call, 0, 4)
	runner := &Runner{Exec: func(ctx context.Context, dir string, args ...string) (ExecResult, error) {
		calls = append(calls, call{dir: dir, args: args})
		return ExecResult{ExitCode: 0}, nil
	}}
	if _, err := runner.Merge(context.Background(), MergeOptions{
		RepoRoot:     "/repo",
		WorktreePath: "/repo/.amux/worktrees/alpha",
		AgentSlug:    "alpha",
		Strategy:     StrategyFFOnly,
		BaseBranch:   "main",
	}); err != nil {
		t.Fatalf("merge ff-only: %v", err)
	}
	want := []string{
		"show-ref --verify refs/heads/main",
		"show-ref --verify refs/heads/amux/alpha",
		"status --porcelain",
		"checkout main",
		"merge --ff-only amux/alpha",
	}
	if len(calls) != len(want) {
		t.Fatalf("unexpected calls: %d", len(calls))
	}
	for i, expected := range want {
		got := strings.Join(calls[i].args, " ")
		if got != expected {
			t.Fatalf("call %d: got %q want %q", i, got, expected)
		}
	}
}

func TestMergeRebaseSuccessExtra(t *testing.T) {
	calls := make([]call, 0, 8)
	runner := &Runner{Exec: func(ctx context.Context, dir string, args ...string) (ExecResult, error) {
		calls = append(calls, call{dir: dir, args: args})
		if args[0] == "rev-parse" {
			return ExecResult{ExitCode: 0, Output: []byte("amux/alpha\n")}, nil
		}
		return ExecResult{ExitCode: 0}, nil
	}}
	if _, err := runner.Merge(context.Background(), MergeOptions{
		RepoRoot:     "/repo",
		WorktreePath: "/repo/.amux/worktrees/alpha",
		AgentSlug:    "alpha",
		Strategy:     StrategyRebase,
		BaseBranch:   "main",
	}); err != nil {
		t.Fatalf("merge rebase: %v", err)
	}
	want := []string{
		"show-ref --verify refs/heads/main",
		"show-ref --verify refs/heads/amux/alpha",
		"status --porcelain",
		"rev-parse --abbrev-ref HEAD",
		"rebase main",
		"checkout main",
		"merge --ff-only amux/alpha",
	}
	if len(calls) != len(want) {
		t.Fatalf("unexpected calls: %d", len(calls))
	}
	for i, expected := range want {
		got := strings.Join(calls[i].args, " ")
		if got != expected {
			t.Fatalf("call %d: got %q want %q", i, got, expected)
		}
	}
}

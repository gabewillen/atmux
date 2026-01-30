package git

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type call struct {
	dir  string
	args []string
}

func TestMergeSquashUsesBaseBranchDefault(t *testing.T) {
	calls := make([]call, 0, 8)
	fake := func(ctx context.Context, dir string, args ...string) (ExecResult, error) {
		calls = append(calls, call{dir: dir, args: args})
		return ExecResult{Output: []byte(""), ExitCode: 0}, nil
	}
	runner := &Runner{Exec: fake}
	_, err := runner.Merge(context.Background(), MergeOptions{
		RepoRoot:     "/repo",
		WorktreePath: "/repo/.amux/worktrees/alpha",
		AgentSlug:    "alpha",
		Strategy:     StrategySquash,
		BaseBranch:   "main",
		AllowDirty:   false,
	})
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	want := []string{
		"show-ref --verify refs/heads/main",
		"show-ref --verify refs/heads/amux/alpha",
		"status --porcelain",
		"checkout main",
		"merge --squash amux/alpha",
		"commit -m amux: squash amux/alpha",
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

func TestMergeInvalidStrategy(t *testing.T) {
	runner := &Runner{Exec: func(ctx context.Context, dir string, args ...string) (ExecResult, error) {
		return ExecResult{ExitCode: 0}, nil
	}}
	if _, err := runner.Merge(context.Background(), MergeOptions{
		RepoRoot:     "/repo",
		WorktreePath: "/repo/.amux/worktrees/alpha",
		AgentSlug:    "alpha",
		Strategy:     MergeStrategy("nope"),
		BaseBranch:   "main",
	}); err == nil {
		t.Fatalf("expected invalid strategy error")
	}
}

func TestMergeBranchMissing(t *testing.T) {
	runner := &Runner{Exec: func(ctx context.Context, dir string, args ...string) (ExecResult, error) {
		if len(args) >= 3 && args[0] == "show-ref" {
			return ExecResult{ExitCode: 1, Output: []byte("not a valid ref")}, errors.New("missing")
		}
		return ExecResult{ExitCode: 0}, nil
	}}
	if _, err := runner.Merge(context.Background(), MergeOptions{
		RepoRoot:     "/repo",
		WorktreePath: "/repo/.amux/worktrees/alpha",
		AgentSlug:    "alpha",
		Strategy:     StrategySquash,
		BaseBranch:   "main",
	}); err == nil {
		t.Fatalf("expected branch missing error")
	}
}

func TestMergeDirtyWorktree(t *testing.T) {
	calledStatus := false
	runner := &Runner{Exec: func(ctx context.Context, dir string, args ...string) (ExecResult, error) {
		if len(args) >= 2 && args[0] == "show-ref" {
			return ExecResult{ExitCode: 0}, nil
		}
		if len(args) >= 2 && args[0] == "status" {
			calledStatus = true
			return ExecResult{ExitCode: 0, Output: []byte(" M file")}, nil
		}
		return ExecResult{ExitCode: 0}, nil
	}}
	if _, err := runner.Merge(context.Background(), MergeOptions{
		RepoRoot:     "/repo",
		WorktreePath: "/repo/.amux/worktrees/alpha",
		AgentSlug:    "alpha",
		Strategy:     StrategySquash,
		BaseBranch:   "main",
	}); err == nil {
		t.Fatalf("expected dirty worktree error")
	}
	if !calledStatus {
		t.Fatalf("expected status check")
	}
}

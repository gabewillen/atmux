package git

import (
	"context"
	"errors"
	"testing"
)

func TestMergeCommitConflict(t *testing.T) {
	calls := 0
	runner := &Runner{Exec: func(ctx context.Context, repoRoot string, args ...string) (ExecResult, error) {
		_ = ctx
		_ = repoRoot
		calls++
		if args[0] == "show-ref" {
			return ExecResult{ExitCode: 0}, nil
		}
		if args[0] == "status" {
			return ExecResult{ExitCode: 0}, nil
		}
		if args[0] == "checkout" {
			return ExecResult{ExitCode: 0}, nil
		}
		if args[0] == "merge" {
			return ExecResult{ExitCode: 1}, errors.New("merge failed")
		}
		if args[0] == "diff" {
			return ExecResult{Output: []byte("file.txt")}, nil
		}
		return ExecResult{}, nil
	}}
	_, err := runner.Merge(context.Background(), MergeOptions{
		RepoRoot:     "/tmp",
		AgentSlug:    "alpha",
		Strategy:     StrategyMergeCommit,
		TargetBranch: "main",
	})
	if !errors.Is(err, ErrMergeConflict) {
		t.Fatalf("expected merge conflict, got %v", err)
	}
	if calls == 0 {
		t.Fatalf("expected calls")
	}
}

func TestMergeFFOnlySuccess(t *testing.T) {
	runner := &Runner{Exec: func(ctx context.Context, repoRoot string, args ...string) (ExecResult, error) {
		if args[0] == "show-ref" {
			return ExecResult{ExitCode: 0}, nil
		}
		if args[0] == "status" {
			return ExecResult{ExitCode: 0}, nil
		}
		return ExecResult{ExitCode: 0}, nil
	}}
	_, err := runner.Merge(context.Background(), MergeOptions{
		RepoRoot:     "/tmp",
		AgentSlug:    "alpha",
		Strategy:     StrategyFFOnly,
		TargetBranch: "main",
	})
	if err != nil {
		t.Fatalf("merge ff-only: %v", err)
	}
}

func TestMergeRebaseWrongBranch(t *testing.T) {
	runner := &Runner{Exec: func(ctx context.Context, repoRoot string, args ...string) (ExecResult, error) {
		if args[0] == "show-ref" {
			return ExecResult{ExitCode: 0}, nil
		}
		if args[0] == "status" {
			return ExecResult{ExitCode: 0}, nil
		}
		if args[0] == "rev-parse" {
			return ExecResult{Output: []byte("other")}, nil
		}
		return ExecResult{}, nil
	}}
	_, err := runner.Merge(context.Background(), MergeOptions{
		RepoRoot:     "/tmp",
		AgentSlug:    "alpha",
		Strategy:     StrategyRebase,
		TargetBranch: "main",
	})
	if !errors.Is(err, ErrBranchMissing) {
		t.Fatalf("expected branch missing, got %v", err)
	}
}

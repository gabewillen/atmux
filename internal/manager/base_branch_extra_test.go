package manager

import (
	"context"
	"errors"
	"testing"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/git"
)

func TestBaseBranchCaching(t *testing.T) {
	calls := 0
	mgr := &Manager{
		git: &git.Runner{Exec: func(ctx context.Context, dir string, args ...string) (git.ExecResult, error) {
			_ = ctx
			_ = dir
			_ = args
			calls++
			return git.ExecResult{Output: []byte("main")}, nil
		}},
		bases: make(map[string]string),
	}
	branch, err := mgr.baseBranch(context.Background(), "/repo")
	if err != nil || branch != "main" {
		t.Fatalf("base branch: %v %s", err, branch)
	}
	branch, err = mgr.baseBranch(context.Background(), "/repo")
	if err != nil || branch != "main" {
		t.Fatalf("base branch cached: %v %s", err, branch)
	}
	if calls != 1 {
		t.Fatalf("expected detect base branch once")
	}
}

func TestBaseBranchDetachedHead(t *testing.T) {
	mgr := &Manager{
		git: &git.Runner{Exec: func(ctx context.Context, dir string, args ...string) (git.ExecResult, error) {
			return git.ExecResult{}, git.ErrDetachedHead
		}},
		cfg: config.Config{},
		bases: make(map[string]string),
	}
	if _, err := mgr.baseBranch(context.Background(), "/repo"); err == nil {
		t.Fatalf("expected detached head error")
	}
	mgr.cfg.Git.Merge.TargetBranch = "main"
	if branch, err := mgr.baseBranch(context.Background(), "/repo"); err != nil || branch != "main" {
		t.Fatalf("expected fallback branch, got %v %s", err, branch)
	}
	if !errors.Is(git.ErrDetachedHead, git.ErrDetachedHead) {
		t.Fatalf("expected detached head error")
	}
}

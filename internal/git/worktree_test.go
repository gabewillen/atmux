package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureWorktreeIdempotent(t *testing.T) {
	repoRoot := initRepo(t)
	runner := NewRunner()
	ctx := context.Background()
	worktree, err := runner.EnsureWorktree(ctx, repoRoot, "alpha")
	if err != nil {
		t.Fatalf("ensure worktree: %v", err)
	}
	if worktree.Path == "" {
		t.Fatalf("expected worktree path")
	}
	if _, err := os.Stat(worktree.Path); err != nil {
		t.Fatalf("worktree missing: %v", err)
	}
	again, err := runner.EnsureWorktree(ctx, repoRoot, "alpha")
	if err != nil {
		t.Fatalf("ensure worktree again: %v", err)
	}
	if !again.Existing {
		t.Fatalf("expected existing worktree")
	}
}

func TestRemoveWorktreeDeletesBranch(t *testing.T) {
	repoRoot := initRepo(t)
	runner := NewRunner()
	ctx := context.Background()
	worktree, err := runner.EnsureWorktree(ctx, repoRoot, "beta")
	if err != nil {
		t.Fatalf("ensure worktree: %v", err)
	}
	if err := runner.RemoveWorktree(ctx, repoRoot, "beta", true); err != nil {
		t.Fatalf("remove worktree: %v", err)
	}
	if _, err := os.Stat(worktree.Path); err == nil {
		t.Fatalf("expected worktree removed")
	}
}

func initRepo(t *testing.T) string {
	repoRoot := t.TempDir()
	runGit(t, repoRoot, "init")
	runGit(t, repoRoot, "config", "user.email", "test@example.com")
	runGit(t, repoRoot, "config", "user.name", "Test")
	path := filepath.Join(repoRoot, "README.md")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	runGit(t, repoRoot, "add", "README.md")
	runGit(t, repoRoot, "commit", "-m", "init")
	return repoRoot
}

func runGit(t *testing.T, dir string, args ...string) {
	runner := NewRunner()
	result, err := runner.Exec(context.Background(), dir, args...)
	if err != nil {
		t.Fatalf("git %v: %v (output: %s)", args, err, string(result.Output))
	}
}

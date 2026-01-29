// Package worktree provides git worktree create/remove for agent isolation.
package worktree

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestWorktreePath(t *testing.T) {
	got := WorktreePath("/repo", "my-agent")
	want := "/repo/.amux/worktrees/my-agent"
	if got != want {
		t.Errorf("WorktreePath = %q, want %q", got, want)
	}
}

func TestBranchName(t *testing.T) {
	got := BranchName("my-agent")
	want := "amux/my-agent"
	if got != want {
		t.Errorf("BranchName = %q, want %q", got, want)
	}
}

func TestCreate_NotRepo(t *testing.T) {
	dir := t.TempDir()
	_, err := Create(dir, "test-agent")
	if err == nil {
		t.Error("Create in non-repo: expected error")
	}
}

func TestCreate_RealRepo(t *testing.T) {
	dir := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %s: %v", out, err)
	}
	// Initial commit so we have HEAD
	cmd = exec.Command("git", "commit", "--allow-empty", "-m", "init")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=a", "GIT_AUTHOR_EMAIL=a@a.com", "GIT_COMMITTER_NAME=a", "GIT_COMMITTER_EMAIL=a@a.com")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %s: %v", out, err)
	}

	wtPath, err := Create(dir, "test-agent")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if wtPath != filepath.Join(dir, WorktreesDir, "test-agent") {
		t.Errorf("Create returned path %q", wtPath)
	}
	if !Exists(dir, "test-agent") {
		t.Error("Exists after Create: false")
	}

	// Idempotent: second Create reuses
	wtPath2, err := Create(dir, "test-agent")
	if err != nil {
		t.Fatalf("Create idempotent: %v", err)
	}
	if wtPath2 != wtPath {
		t.Errorf("Create idempotent path %q != %q", wtPath2, wtPath)
	}

	if err := Remove(dir, "test-agent"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if Exists(dir, "test-agent") {
		t.Error("Exists after Remove: true")
	}
}

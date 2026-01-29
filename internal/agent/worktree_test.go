package agent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/agentflare-ai/amux/pkg/api"
)

func TestWorktreeLifecycle(t *testing.T) {
	tmpRepo := t.TempDir()
	initGitRepo(t, tmpRepo)
	repoRoot := api.RepoRoot(tmpRepo)
	slug := api.AgentSlug("test-agent")

	// Ensure
	path, err := EnsureWorktree(repoRoot, slug, "main")
	if err != nil {
		t.Fatalf("EnsureWorktree failed: %v", err)
	}

	expectedPath := filepath.Join(tmpRepo, ".amux", "worktrees", "test-agent")
	if path != expectedPath {
		t.Errorf("Expected path %q, got %q", expectedPath, path)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Worktree dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("Worktree path is not a directory")
	}

	// Idempotency
	path2, err := EnsureWorktree(repoRoot, slug, "main")
	if err != nil {
		t.Fatalf("EnsureWorktree (reuse) failed: %v", err)
	}
	if path2 != path {
		t.Errorf("Reuse returned different path")
	}

	// Remove
	if err := RemoveWorktree(repoRoot, slug); err != nil {
		t.Fatalf("RemoveWorktree failed: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("Worktree dir still exists after removal")
	}
}

func TestEnsureWorktree_InvalidSlug(t *testing.T) {
	_, err := EnsureWorktree("/tmp", "Invalid Slug", "main")
	if err == nil {
		t.Error("Expected error for invalid slug")
	}
}

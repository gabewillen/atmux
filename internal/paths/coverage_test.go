package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolverPaths(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir git: %v", err)
	}
	resolver, err := NewResolver(repo)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	if resolver.RepoRoot() == "" {
		t.Fatalf("expected repo root")
	}
	if resolver.HomeDir() == "" {
		t.Fatalf("expected home dir")
	}
	if resolver.AmuxRoot() == "" || resolver.WorktreesDir() == "" || resolver.PTYDir() == "" {
		t.Fatalf("expected amux dirs")
	}
	if resolver.UserAdaptersDir() == "" || resolver.ProjectAdaptersDir() == "" {
		t.Fatalf("expected adapter dirs")
	}
	if resolver.UserConfigPath() == "" || resolver.ProjectConfigPath() == "" {
		t.Fatalf("expected config paths")
	}
}

func TestResolverOptionalRepo(t *testing.T) {
	resolver, err := NewResolverOptionalRepo(t.TempDir())
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	if resolver.RepoRoot() != "" {
		t.Fatalf("expected empty repo root")
	}
}

func TestWorktreeHelpers(t *testing.T) {
	if AmuxRootForRepo("") != "" {
		t.Fatalf("expected empty amux root")
	}
	if WorktreesDirForRepo("") != "" {
		t.Fatalf("expected empty worktrees")
	}
	if WorktreePathForRepo("", "slug") != "" {
		t.Fatalf("expected empty worktree path")
	}
	if PTYDirForRepo("") != "" {
		t.Fatalf("expected empty pty dir")
	}
}

func TestExpandHome(t *testing.T) {
	resolver := &Resolver{homeDir: "/home/test"}
	path := resolver.ExpandHome("~/work")
	if path != "/home/test/work" {
		t.Fatalf("unexpected path: %s", path)
	}
}

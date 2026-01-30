package paths

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindRepoRootMissing(t *testing.T) {
	if _, err := FindRepoRoot(""); err == nil {
		t.Fatalf("expected repo root error")
	}
}

func TestFindRepoRootWithGitFile(t *testing.T) {
	repo := t.TempDir()
	gitFile := filepath.Join(repo, ".git")
	if err := os.WriteFile(gitFile, []byte("gitdir: /tmp/gitdir\n"), 0o644); err != nil {
		t.Fatalf("write git file: %v", err)
	}
	root, err := FindRepoRoot(repo)
	if err != nil {
		t.Fatalf("find repo root: %v", err)
	}
	if root != repo {
		t.Fatalf("expected repo root")
	}
}

func TestResolverPathsExtra(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir git: %v", err)
	}
	resolver, err := NewResolver(repo)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	if !strings.Contains(resolver.WorktreePath("alpha"), "worktrees") {
		t.Fatalf("unexpected worktree path")
	}
	if resolver.PTYDir() == "" {
		t.Fatalf("expected pty dir")
	}
	if resolver.SocketPath() == "" {
		t.Fatalf("expected socket path")
	}
	if resolver.UserAdapterConfigPath("alpha") == "" {
		t.Fatalf("expected user adapter config path")
	}
	if resolver.ProjectAdapterConfigPath("alpha") == "" {
		t.Fatalf("expected project adapter config path")
	}
	path := resolver.ExpandHome("~/test")
	if !strings.Contains(path, "test") {
		t.Fatalf("expected expand home")
	}
	if _, err := resolver.CanonicalizeRepoRoot(repo); err != nil {
		t.Fatalf("canonicalize repo root: %v", err)
	}
}

func TestRepoPathHelpers(t *testing.T) {
	if AmuxRootForRepo("") != "" {
		t.Fatalf("expected empty repo root")
	}
	if WorktreePathForRepo("", "alpha") != "" {
		t.Fatalf("expected empty worktree path")
	}
	if PTYDirForRepo("") != "" {
		t.Fatalf("expected empty pty dir")
	}
}

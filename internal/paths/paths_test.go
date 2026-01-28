package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSlugifyAgent(t *testing.T) {
	slug := SlugifyAgent("My Agent!! 123")
	if slug != "my-agent-123" {
		t.Fatalf("unexpected slug: %s", slug)
	}
}

func TestFindRepoRootCanonicalizesSymlink(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	link := filepath.Join(tmp, "repo-link")
	if err := os.Symlink(repo, link); err != nil {
		t.Skipf("symlink not supported: %v", err)
	}
	want, err := canonicalizePath(repo)
	if err != nil {
		t.Fatalf("canonicalize: %v", err)
	}
	root, err := FindRepoRoot(link)
	if err != nil {
		t.Fatalf("find repo root: %v", err)
	}
	if root != want {
		t.Fatalf("expected %s, got %s", want, root)
	}
}

func TestWorktreePath(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	resolver, err := NewResolver(repo)
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}
	path := resolver.WorktreePath("agent-one")
	want := filepath.Join(repo, ".amux", "worktrees", "agent-one")
	if path != want {
		t.Fatalf("expected %s, got %s", want, path)
	}
}

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

func TestSlugifyAgentEmptyFallsBack(t *testing.T) {
	slug := SlugifyAgent("$$$")
	if slug != "agent" {
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
	want, err := CanonicalizeRepoRoot(repo, "")
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

func TestCanonicalizeRepoRootExpandsHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("home dir: %v", err)
	}
	path, err := CanonicalizeRepoRoot("~/", home)
	if err != nil {
		t.Fatalf("canonicalize: %v", err)
	}
	if path != home {
		t.Fatalf("expected %s, got %s", home, path)
	}
}

func TestUniqueAgentSlugCollision(t *testing.T) {
	used := map[string]struct{}{
		"agent":      {},
		"agent-2":    {},
		"agent-3":    {},
		"frontend":   {},
		"frontend-2": {},
	}
	slug := UniqueAgentSlug("Agent", used)
	if slug != "agent-4" {
		t.Fatalf("unexpected slug: %s", slug)
	}
	slug = UniqueAgentSlug("Frontend", used)
	if slug != "frontend-3" {
		t.Fatalf("unexpected slug: %s", slug)
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

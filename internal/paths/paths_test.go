package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveExpand(t *testing.T) {
	r, err := NewResolver()
	if err != nil {
		t.Fatalf("NewResolver failed: %v", err)
	}

	home := os.Getenv("HOME")
	if home == "" {
		t.Skip("HOME env var not set, skipping expansion test")
	}

	tests := []struct {
		input string
		want  string // if empty, verify prefix match with home
	}{
		{"~", home},
		{"~/foo", filepath.Join(home, "foo")},
		{"/abs/path", "/abs/path"},
	}

	for _, tc := range tests {
		got := r.Expand(tc.input)
		if got != tc.want {
			t.Errorf("Expand(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestCanonicalizeRepoRoot(t *testing.T) {
	r, err := NewResolver()
	if err != nil {
		t.Fatalf("NewResolver failed: %v", err)
	}

	tmpDir, err := os.MkdirTemp("", "amux-paths-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a real directory
	realDir := filepath.Join(tmpDir, "real")
	if err := os.Mkdir(realDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a symlink
	linkDir := filepath.Join(tmpDir, "link")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Fatal(err)
	}

	// Test 1: Real directory persistence
	canon, err := r.CanonicalizeRepoRoot(realDir)
	if err != nil {
		t.Errorf("CanonicalizeRepoRoot(real) failed: %v", err)
	}
	// Resolve symlinks on the real path just in case tmpDir itself has symlinks (common on macOS)
	wantReal, _ := filepath.EvalSymlinks(realDir)
	if canon != wantReal {
		t.Errorf("CanonicalizeRepoRoot(real) = %q, want %q", canon, wantReal)
	}

	// Test 2: Symlink resolution
	canonLink, err := r.CanonicalizeRepoRoot(linkDir)
	if err != nil {
		t.Errorf("CanonicalizeRepoRoot(link) failed: %v", err)
	}
	if canonLink != wantReal {
		t.Errorf("CanonicalizeRepoRoot(link) = %q, want %q (should resolve to real path)", canonLink, wantReal)
	}

	// Test 3: Non-existent path
	_, err = r.CanonicalizeRepoRoot(filepath.Join(tmpDir, "does-not-exist"))
	if err == nil {
		t.Error("CanonicalizeRepoRoot(non-existent) should error, got nil")
	}
}

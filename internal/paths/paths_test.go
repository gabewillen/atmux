package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Skipping home dir test: " + err.Error())
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"~/foo", filepath.Join(home, "foo")},
		{"/abs/path", "/abs/path"},
		{"rel/path", "rel/path"},
	}

	for _, tt := range tests {
		got := ExpandHome(tt.input)
		if got != tt.expected {
			t.Errorf("ExpandHome(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestCanonicalizeRepoRoot(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a real directory
	realDir := filepath.Join(tmpDir, "repo")
	if err := os.Mkdir(realDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a symlink
	linkDir := filepath.Join(tmpDir, "link")
	if err := os.Symlink(realDir, linkDir); err != nil {
		// Symlinks might not be supported on Windows or some envs
		t.Skip("Skipping symlink test")
	}

	// Test canonicalization of symlink
	canon, err := CanonicalizeRepoRoot(linkDir)
	if err != nil {
		t.Fatalf("CanonicalizeRepoRoot failed: %v", err)
	}

	// EvalSymlinks on realDir (to handle /var/folders vs /private/var on macOS)
	expected, _ := filepath.EvalSymlinks(realDir)
	if canon != expected {
		t.Errorf("CanonicalizeRepoRoot(%q) = %q, want %q", linkDir, canon, expected)
	}
}

func TestDefaultDirs(t *testing.T) {
	home, _ := os.UserHomeDir()
	
	cfg, err := DefaultConfigDir()
	if err == nil {
		expected := filepath.Join(home, ".config", "amux")
		if cfg != expected {
			t.Errorf("DefaultConfigDir = %q, want %q", cfg, expected)
		}
	}

	sock, err := DefaultSocketPath()
	if err == nil {
		expected := filepath.Join(home, ".amux", "amuxd.sock")
		if sock != expected {
			t.Errorf("DefaultSocketPath = %q, want %q", sock, expected)
		}
	}

	repoRoot := "/tmp/foo"
	wt := DefaultWorktreesDir(repoRoot)
	expected := filepath.Join(repoRoot, ".amux", "worktrees")
	if wt != expected {
		t.Errorf("DefaultWorktreesDir = %q, want %q", wt, expected)
	}
}

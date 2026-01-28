package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandHome(t *testing.T) {
	homeDir := "/home/test"
	tests := []struct {
		input    string
		expected string
	}{
		{"~/config", "/home/test/config"},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := expandHome(tt.input, homeDir)
			if got != tt.expected {
				t.Errorf("expandHome(%q, %q) = %q, want %q", tt.input, homeDir, got, tt.expected)
			}
		})
	}
}

func TestWorktreePath(t *testing.T) {
	tmpDir := t.TempDir()
	resolver, err := NewResolver("~/.config/amux", "", tmpDir)
	if err != nil {
		t.Fatalf("NewResolver failed: %v", err)
	}

	path, err := resolver.WorktreePath("test-agent")
	if err != nil {
		t.Fatalf("WorktreePath failed: %v", err)
	}

	expected := filepath.Join(tmpDir, ".amux", "worktrees", "test-agent")
	if path != expected {
		t.Errorf("WorktreePath() = %q, want %q", path, expected)
	}
}

func TestAmuxDir(t *testing.T) {
	tmpDir := t.TempDir()
	resolver, err := NewResolver("~/.config/amux", "", tmpDir)
	if err != nil {
		t.Fatalf("NewResolver failed: %v", err)
	}

	path, err := resolver.AmuxDir()
	if err != nil {
		t.Fatalf("AmuxDir failed: %v", err)
	}

	expected := filepath.Join(tmpDir, ".amux")
	if path != expected {
		t.Errorf("AmuxDir() = %q, want %q", path, expected)
	}
}

func TestRepoRootRequired(t *testing.T) {
	resolver, err := NewResolver("~/.config/amux", "", "")
	if err != nil {
		t.Fatalf("NewResolver failed: %v", err)
	}

	_, err = resolver.WorktreePath("test")
	if err == nil {
		t.Error("WorktreePath should fail when repo root is not set")
	}

	_, err = resolver.AmuxDir()
	if err == nil {
		t.Error("AmuxDir should fail when repo root is not set")
	}
}

func TestNewResolver(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir failed: %v", err)
	}

	tests := []struct {
		name      string
		configDir string
		homeDir   string
		repoRoot  string
		wantErr   bool
	}{
		{
			name:      "valid config",
			configDir: "~/.config/amux",
			homeDir:   homeDir,
			repoRoot:  t.TempDir(),
			wantErr:   false,
		},
		{
			name:      "auto home dir",
			configDir: "~/.config/amux",
			homeDir:   "",
			repoRoot:  t.TempDir(),
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewResolver(tt.configDir, tt.homeDir, tt.repoRoot)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewResolver() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

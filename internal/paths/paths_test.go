package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot get home directory: %v", err)
	}

	r := &Resolver{homeDir: home}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "expand home",
			input:    "~/test/path",
			expected: filepath.Join(home, "test/path"),
		},
		{
			name:     "no expansion needed",
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:     "relative path",
			input:    "relative/path",
			expected: "relative/path",
		},
		{
			name:     "home only",
			input:    "~/",
			expected: home,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.ExpandHome(tt.input)
			if result != tt.expected {
				t.Errorf("ExpandHome(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestWorktreeDir(t *testing.T) {
	r := &Resolver{repoRoot: "/test/repo"}

	expected := "/test/repo/.amux/worktrees/test-agent"
	result := r.WorktreeDir("test-agent")

	if result != expected {
		t.Errorf("WorktreeDir(%q) = %q, want %q", "test-agent", result, expected)
	}
}

func TestWorktreeDirNoRepo(t *testing.T) {
	r := &Resolver{repoRoot: ""}

	result := r.WorktreeDir("test-agent")

	if result != "" {
		t.Errorf("WorktreeDir with no repo = %q, want empty string", result)
	}
}

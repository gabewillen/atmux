package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeAgentSlug(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple lowercase",
			input:    "test-agent",
			expected: "test-agent",
		},
		{
			name:     "uppercase and special chars",
			input:    "Test-Agent_123!",
			expected: "test-agent-123",
		},
		{
			name:     "multiple hyphens collapsed",
			input:    "test--agent",
			expected: "test-agent",
		},
		{
			name:     "long name truncated",
			input:    "very-long-agent-name-that-exceeds-sixty-three-characters-limit-and-more",
			expected: "very-long-agent-name-that-exceeds-sixty-three-characters-limit",
		},
		{
			name:     "trimmed hyphens",
			input:    "-test-agent-",
			expected: "test-agent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeAgentSlug(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeAgentSlug(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestPaths(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir, err := os.MkdirTemp("", "amux-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize git repo
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Change to test directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to test dir: %v", err)
	}

	config := Config{
		HomeDir:      tmpDir,
		ConfigDir:    filepath.Join(tmpDir, ".config", "amux"),
		DataDir:      filepath.Join(tmpDir, ".local", "share", "amux"),
		RuntimeDir:   filepath.Join(tmpDir, ".local", "run", "amux"),
		RegistryRoot: filepath.Join(tmpDir, ".local", "share", "amux", "registry"),
		ModelsRoot:   filepath.Join(tmpDir, ".local", "share", "amux", "models"),
		HooksRoot:    filepath.Join(tmpDir, ".local", "share", "amux", "hooks"),
	}

	resolver, err := NewResolver(config)
	if err != nil {
		t.Fatalf("Failed to create resolver: %v", err)
	}

	// Test basic paths
	if resolver.RepoRoot() != tmpDir {
		t.Errorf("RepoRoot() = %q, want %q", resolver.RepoRoot(), tmpDir)
	}

	expectedAmuxRoot := filepath.Join(tmpDir, ".amux")
	if resolver.AmuxRoot() != expectedAmuxRoot {
		t.Errorf("AmuxRoot() = %q, want %q", resolver.AmuxRoot(), expectedAmuxRoot)
	}

	// Test agent worktree
	expectedWorktree := filepath.Join(expectedAmuxRoot, "worktrees", "test-agent")
	if resolver.AgentWorktree("test-agent") != expectedWorktree {
		t.Errorf("AgentWorktree() = %q, want %q", resolver.AgentWorktree("test-agent"), expectedWorktree)
	}
}

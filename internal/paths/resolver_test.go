package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewResolver(t *testing.T) {
	// Create temporary directory that looks like a git repo
	tempDir, err := os.MkdirTemp("", "amux-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	gitDir := filepath.Join(tempDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	resolver, err := NewResolver(tempDir)
	if err != nil {
		t.Fatalf("NewResolver() failed: %v", err)
	}

	expectedAmuxDir := filepath.Join(tempDir, ".amux")
	if resolver.AmuxDir() != expectedAmuxDir {
		t.Errorf("AmuxDir() = %s, want %s", resolver.AmuxDir(), expectedAmuxDir)
	}
}

func TestNormalizeAgentSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"Claude Code", "claude-code", false},
		{"cursor_123", "cursor-123", false},
		{"Test Agent!@#", "test-agent", false},
		{"", "", true},
		{"---", "", true},
		{"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", "abcdefghijklmnopqrstuvwxyz0123456789", false}, // Should be truncated
	}

	for _, tt := range tests {
		result, err := normalizeAgentSlug(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("normalizeAgentSlug(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && result != tt.expected {
			t.Errorf("normalizeAgentSlug(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
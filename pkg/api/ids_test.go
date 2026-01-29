package api

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewID(t *testing.T) {
	id := NewID()
	if !IDIsValid(id) {
		t.Fatalf("NewID() returned invalid ID: %v", id)
	}

	// Test multiple calls produce different IDs
	id2 := NewID()
	if id == id2 {
		t.Fatalf("NewID() produced duplicate IDs: %v", id)
	}
}

func TestParseID(t *testing.T) {
	tests := []struct {
		input    string
		expected ID
		hasError bool
	}{
		{"", 0, true},
		{"0", 0, true},
		{"invalid", 0, true},
		{"123", 123, false},
		{"42", 42, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			id, err := ParseID(tt.input)
			if tt.hasError {
				if err == nil {
					t.Errorf("ParseID(%q) expected error but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseID(%q) unexpected error: %v", tt.input, err)
				}
				if id != tt.expected {
					t.Errorf("ParseID(%q) = %v, want %v", tt.input, id, tt.expected)
				}
			}
		})
	}
}

func TestIDToString(t *testing.T) {
	id := ID(42)
	result := IDToString(id)
	if result != "42" {
		t.Errorf("IDToString(42) = %q, want %q", result, "42")
	}
}

func TestIDIsValid(t *testing.T) {
	if !IDIsValid(ID(1)) {
		t.Error("IDIsValid(1) should return true")
	}
	if IDIsValid(ID(0)) {
		t.Error("IDIsValid(0) should return false")
	}
}

func TestNormalizeAgentSlug(t *testing.T) {
	tests := []struct {
		name     string
		expected AgentSlug
	}{
		{"My Agent", "my-agent"},
		{"Frontend Dev", "frontend-dev"},
		{"test@example.com", "test-example-com"},
		{"UPPERCASE", "uppercase"},
		{"multiple---dashes", "multiple-dashes"},
		{"---leading-and-trailing---", "leading-and-trailing"},
		{"", "agent"},
		{"!@#$%", "agent"},
		{"very-long-name-that-exceeds-the-sixty-three-character-limit", "very-long-name-that-exceeds-the-sixty-three-charac"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeAgentSlug(tt.name)
			if result != tt.expected {
				t.Errorf("NormalizeAgentSlug(%q) = %q, want %q", tt.name, result, tt.expected)
			}
		})
	}
}

func TestMakeUniqueAgentSlug(t *testing.T) {
	existing := map[AgentSlug]bool{
		"agent": true,
		"test":  true,
	}

	// Test unique slug (no conflict)
	result := MakeUniqueAgentSlug("unique", existing)
	if result != "unique" {
		t.Errorf("MakeUniqueAgentSlug(%q) = %q, want %q", "unique", result, "unique")
	}

	// Test conflicting slug
	result = MakeUniqueAgentSlug("agent", existing)
	if result != "agent-2" {
		t.Errorf("MakeUniqueAgentSlug(%q) = %q, want %q", "agent", result, "agent-2")
	}

	// Test multiple conflicts
	existing["agent-2"] = true
	result = MakeUniqueAgentSlug("agent", existing)
	if result != "agent-3" {
		t.Errorf("MakeUniqueAgentSlug(%q) = %q, want %q", "agent", result, "agent-3")
	}
}

func TestCanonicalizeRepoRoot(t *testing.T) {
	// Test empty path
	_, err := CanonicalizeRepoRoot("")
	if err == nil {
		t.Error("CanonicalizeRepoRoot(\"\") expected error")
	}

	// Test current directory (should become absolute)

	result, err := CanonicalizeRepoRoot(".")
	if err != nil {
		t.Errorf("CanonicalizeRepoRoot(.\") unexpected error: %v", err)
	}

	absExpected, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	if string(result) != absExpected {
		t.Errorf("CanonicalizeRepoRoot(.\") = %q, want %q", result, absExpected)
	}

	// Test home directory expansion (this test may need adjustment in CI)
	home, err := os.UserHomeDir()
	if err == nil {
		result, err := CanonicalizeRepoRoot("~/test")
		if err != nil {
			t.Errorf("CanonicalizeRepoRoot(~/test\") unexpected error: %v", err)
		}

		expected := filepath.Join(home, "test")
		if string(result) != expected {
			t.Errorf("CanonicalizeRepoRoot(~/test\") = %q, want %q", result, expected)
		}
	}
}

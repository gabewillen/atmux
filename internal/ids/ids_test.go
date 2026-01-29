// Package ids implements tests for identifier utilities and normalization functions
package ids

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stateforward/hsm-go/muid"
)

// TestNew tests generating new IDs
func TestNew(t *testing.T) {
	id1 := New()
	id2 := New()

	if uint64(id1) == 0 {
		t.Error("Expected non-zero ID from New()")
	}

	if uint64(id2) == 0 {
		t.Error("Expected non-zero ID from New()")
	}

	if id1 == id2 {
		t.Error("Expected different IDs from successive New() calls")
	}
}

// TestEncodeID tests encoding an muid.MUID as base-10 string
func TestEncodeID(t *testing.T) {
	id := muid.MUID(42)
	encoded := EncodeID(id)

	if encoded != "42" {
		t.Errorf("Expected '42', got '%s'", encoded)
	}
}

// TestDecodeID tests decoding a base-10 string to muid.MUID
func TestDecodeID(t *testing.T) {
	// Test valid ID
	id, err := DecodeID("42")
	if err != nil {
		t.Fatalf("Unexpected error decoding valid ID: %v", err)
	}
	if uint64(id) != 42 {
		t.Errorf("Expected ID 42, got %d", uint64(id))
	}

	// Test invalid ID
	_, err = DecodeID("invalid")
	if err == nil {
		t.Error("Expected error decoding invalid ID")
	}
}

// TestNormalizeAgentSlug tests agent slug normalization
func TestNormalizeAgentSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"MyAgent", "myagent"},
		{"My Agent 123!", "my-agent-123"},
		{"___test___", "test"},
		{"test--multiple---hyphens", "test-multiple-hyphens"},
		{"Test_Agent_Name", "test-agent-name"},
		{"", ""},
		{"a", "a"},
		{"A-B-C", "a-b-c"},
		{"very-long-agent-name-that-exceeds-sixty-three-characters-limit-and-more", "very-long-agent-name-that-exceeds-sixty-three-characters-limit"},
	}

	for _, tt := range tests {
		result := NormalizeAgentSlug(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeAgentSlug(%q) = %q; expected %q", tt.input, result, tt.expected)
		}
	}
}

// TestCanonicalizeRepoRoot tests repository root canonicalization
func TestCanonicalizeRepoRoot(t *testing.T) {
	// Test basic functionality
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "test", "path")
	
	// Create the test directory
	err := os.MkdirAll(testPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Test relative path canonicalization
	relPath := filepath.Join("test", "path")
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(origDir) // Restore original directory
	
	canonical, err := CanonicalizeRepoRoot(relPath)
	if err != nil {
		t.Fatalf("Unexpected error canonicalizing path: %v", err)
	}
	
	expected := filepath.Join(tempDir, "test", "path")
	if canonical != expected {
		t.Errorf("CanonicalizeRepoRoot(%q) = %q; expected %q", relPath, canonical, expected)
	}

	// Test absolute path
	absPath := filepath.Join(tempDir, "test", "path")
	canonical, err = CanonicalizeRepoRoot(absPath)
	if err != nil {
		t.Fatalf("Unexpected error canonicalizing absolute path: %v", err)
	}
	if canonical != expected {
		t.Errorf("CanonicalizeRepoRoot(%q) = %q; expected %q", absPath, canonical, expected)
	}

	// Test path with dots
	dotPath := filepath.Join(tempDir, "test", "..", "test", "path")
	canonical, err = CanonicalizeRepoRoot(dotPath)
	if err != nil {
		t.Fatalf("Unexpected error canonicalizing dotted path: %v", err)
	}
	if canonical != expected {
		t.Errorf("CanonicalizeRepoRoot(%q) = %q; expected %q", dotPath, canonical, expected)
	}

	// Test home directory expansion if possible
	homeDir, err := os.UserHomeDir()
	if err == nil {
		homePath := "~/tmp"

		canonical, err = CanonicalizeRepoRoot(homePath)
		if err != nil {
			t.Logf("Could not canonicalize home path (might not exist): %v", err)
		} else {
			// Just check that it starts with the home directory
			if !strings.HasPrefix(canonical, homeDir) {
				t.Errorf("CanonicalizeRepoRoot(%q) = %q; expected to start with %q", homePath, canonical, homeDir)
			}
		}
	} else {
		t.Logf("Skipping home directory test: %v", err)
	}
}
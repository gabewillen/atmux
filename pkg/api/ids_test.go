package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stateforward/hsm-go/muid"
)

func TestNormalizeAgentSlug(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple lowercase",
			input:    "frontend-dev",
			expected: "frontend-dev",
		},
		{
			name:     "uppercase to lowercase",
			input:    "Frontend-Dev",
			expected: "frontend-dev",
		},
		{
			name:     "spaces to dashes",
			input:    "frontend dev",
			expected: "frontend-dev",
		},
		{
			name:     "special characters to dashes",
			input:    "frontend_dev@test",
			expected: "frontend-dev-test",
		},
		{
			name:     "collapse consecutive dashes",
			input:    "frontend---dev",
			expected: "frontend-dev",
		},
		{
			name:     "trim leading and trailing dashes",
			input:    "-frontend-dev-",
			expected: "frontend-dev",
		},
		{
			name:     "truncate long names",
			input:    "this-is-a-very-long-agent-name-that-exceeds-sixty-three-characters-total",
			expected: "this-is-a-very-long-agent-name-that-exceeds-sixty-three-charact",
		},
		{
			name:     "empty string becomes agent",
			input:    "",
			expected: "agent",
		},
		{
			name:     "only special characters becomes agent",
			input:    "!!!",
			expected: "agent",
		},
		{
			name:     "unicode to dashes",
			input:    "前端-开发",
			expected: "agent", // All non-ASCII becomes dashes, trim leaves empty
		},
		{
			name:     "mixed alphanumeric",
			input:    "backend123test",
			expected: "backend123test",
		},
		{
			name:     "preserve valid dashes",
			input:    "test-runner-2",
			expected: "test-runner-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeAgentSlug(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeAgentSlug(%q) = %q, want %q", tt.input, result, tt.expected)
			}

			// Verify it's filesystem-safe
			if len(result) > 63 {
				t.Errorf("Result exceeds 63 characters: %q (len=%d)", result, len(result))
			}
		})
	}
}

func TestCanonicalizeRepoRoot(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a test directory
	testDir := filepath.Join(tmpDir, "test-repo")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	tests := []struct {
		name      string
		input     string
		wantErr   bool
		checkAbs  bool
		checkPath string
	}{
		{
			name:      "absolute path",
			input:     testDir,
			wantErr:   false,
			checkAbs:  true,
			checkPath: testDir,
		},
		{
			name:     "relative path",
			input:    ".",
			wantErr:  false,
			checkAbs: true,
		},
		{
			name:     "path with ..",
			input:    testDir + "/..",
			wantErr:  false,
			checkAbs: true,
		},
		{
			name:     "non-existent path still canonicalizes",
			input:    "/nonexistent/path",
			wantErr:  false,
			checkAbs: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CanonicalizeRepoRoot(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CanonicalizeRepoRoot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// Check if result is absolute
				if tt.checkAbs && !filepath.IsAbs(result) {
					t.Errorf("CanonicalizeRepoRoot() = %q, expected absolute path", result)
				}

				// Check specific path if provided
				if tt.checkPath != "" && result != tt.checkPath {
					t.Errorf("CanonicalizeRepoRoot() = %q, want %q", result, tt.checkPath)
				}

				// Verify path is clean (no ./ or ../)
				if filepath.Clean(result) != result {
					t.Errorf("CanonicalizeRepoRoot() = %q, not clean", result)
				}
			}
		})
	}
}

func TestCanonicalizeRepoRoot_HomeExpansion(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("Cannot get home directory: %v", err)
	}

	result, err := CanonicalizeRepoRoot("~/test-repo")
	if err != nil {
		t.Fatalf("CanonicalizeRepoRoot() error = %v", err)
	}

	expected := filepath.Join(homeDir, "test-repo")
	// Clean both for comparison since symlink resolution may differ
	if filepath.Clean(result) != filepath.Clean(expected) {
		t.Errorf("CanonicalizeRepoRoot(\"~/test-repo\") = %q, want %q", result, expected)
	}

	// Verify it's absolute
	if !filepath.IsAbs(result) {
		t.Errorf("CanonicalizeRepoRoot() = %q, expected absolute path", result)
	}
}

func TestGenerateID(t *testing.T) {
	// Generate multiple IDs and verify they're all non-zero
	seen := make(map[muid.MUID]bool)
	for i := 0; i < 1000; i++ {
		id := GenerateID()
		if id == BroadcastID {
			t.Errorf("GenerateID() returned reserved value 0")
		}
		if id == 0 {
			t.Errorf("GenerateID() returned 0")
		}
		seen[id] = true
	}

	// Verify we got unique IDs (should be very high uniqueness)
	if len(seen) < 999 {
		t.Errorf("GenerateID() produced too many collisions: %d unique out of 1000", len(seen))
	}
}

func TestBroadcastID(t *testing.T) {
	// Verify BroadcastID is 0 as specified
	if BroadcastID != 0 {
		t.Errorf("BroadcastID = %v, want 0", BroadcastID)
	}
}

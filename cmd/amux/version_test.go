package main

import (
	"os"
	"strings"
	"testing"
)

// TestSpecVersionLock ensures spec-v1.22.md is present and version-locked.
func TestSpecVersionLock(t *testing.T) {
	specPath := "../../docs/spec-v1.22.md"
	
	// Check if spec file exists
	if _, err := os.Stat(specPath); err != nil {
		t.Fatalf("spec-v1.22.md not found: %v", err)
	}

	// Read first few lines to verify version
	content, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("failed to read spec file: %v", err)
	}

	// Check for version marker
	expectedVersion := "**Version:** v1.22"
	if !strings.Contains(string(content), expectedVersion) {
		t.Fatalf("spec file does not contain expected version marker: %s", expectedVersion)
	}

	t.Log("✅ spec-v1.22.md version lock verified")
}
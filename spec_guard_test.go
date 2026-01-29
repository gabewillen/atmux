// Package main implements guard tests to ensure spec compliance
package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stateforward/amux/internal/specchecker"
)

// TestSpecFileExistsAndValid verifies that spec-v1.22.md exists in the repository
// and contains the expected version, fulfilling the requirement that
// "a guard test or startup check fails fast with a clear error if the file 
// is missing or the expected version marker does not match."
func TestSpecFileExistsAndValid(t *testing.T) {
	// Determine the root directory relative to this test file
	// Since this file is in the root, we just need to check for the spec file
	specPath := filepath.Join("docs", "spec-v1.22.md")
	
	// Check if the file exists in the expected location
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Fatalf("spec file does not exist at expected location: %s", specPath)
	} else if err != nil {
		t.Fatalf("error accessing spec file: %v", err)
	}

	// Use the specchecker to verify the version
	err := specchecker.CheckSpecPresenceAndVersion(specPath)
	if err != nil {
		t.Fatalf("spec file validation failed: %v", err)
	}

	t.Logf("Successfully verified spec file exists and has correct version at: %s", specPath)
}
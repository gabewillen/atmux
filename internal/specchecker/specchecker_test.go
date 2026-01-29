// Package specchecker implements tests for the spec checker
package specchecker

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCheckSpecPresenceAndVersion tests checking spec presence and version
func TestCheckSpecPresenceAndVersion(t *testing.T) {
	// Create a temporary spec file with correct version
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, SpecFileName)

	specContent := `# Agent Multiplexer (amux) Specification

**Version:** v1.22
**Status:** Draft

This is a test spec file.
`

	err := os.WriteFile(specPath, []byte(specContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write spec file: %v", err)
	}

	// Test with correct spec file
	err = CheckSpecPresenceAndVersion(specPath)
	if err != nil {
		t.Errorf("Expected no error for valid spec file, got: %v", err)
	}

	// Test with non-existent file
	nonExistentPath := filepath.Join(tempDir, "non-existent.md")
	err = CheckSpecPresenceAndVersion(nonExistentPath)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test with wrong version
	wrongVersionContent := `# Agent Multiplexer (amux) Specification

**Version:** v1.21
**Status:** Draft

This is a test spec file with wrong version.
`

	wrongVersionPath := filepath.Join(tempDir, "wrong-version.md")
	err = os.WriteFile(wrongVersionPath, []byte(wrongVersionContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write wrong version spec file: %v", err)
	}

	err = CheckSpecPresenceAndVersion(wrongVersionPath)
	if err == nil {
		t.Error("Expected error for wrong version spec file")
	}
}

// TestGetSpecVersion tests extracting the version from the spec file
func TestGetSpecVersion(t *testing.T) {
	tempDir := t.TempDir()

	// Test with markdown format
	specPath1 := filepath.Join(tempDir, "spec1.md")
	specContent1 := `# Agent Multiplexer (amux) Specification

**Version:** v1.22
**Status:** Draft

This is a test spec file.
`

	err := os.WriteFile(specPath1, []byte(specContent1), 0644)
	if err != nil {
		t.Fatalf("Failed to write spec file: %v", err)
	}

	version, err := GetSpecVersion(specPath1)
	if err != nil {
		t.Fatalf("Unexpected error getting version: %v", err)
	}

	if version != "v1.22" {
		t.Errorf("Expected version 'v1.22', got '%s'", version)
	}

	// Test with plain text format
	specPath2 := filepath.Join(tempDir, "spec2.md")
	specContent2 := `# Agent Multiplexer (amux) Specification

Version: v1.23
Status: Draft

This is another test spec file.
`

	err = os.WriteFile(specPath2, []byte(specContent2), 0644)
	if err != nil {
		t.Fatalf("Failed to write spec file: %v", err)
	}

	version, err = GetSpecVersion(specPath2)
	if err != nil {
		t.Fatalf("Unexpected error getting version: %v", err)
	}

	if version != "v1.23" {
		t.Errorf("Expected version 'v1.23', got '%s'", version)
	}

	// Test with file that has no version
	specPath3 := filepath.Join(tempDir, "spec3.md")
	specContent3 := `# Agent Multiplexer (amux) Specification

Status: Draft

This is a test spec file with no version.
`

	err = os.WriteFile(specPath3, []byte(specContent3), 0644)
	if err != nil {
		t.Fatalf("Failed to write spec file: %v", err)
	}

	_, err = GetSpecVersion(specPath3)
	if err == nil {
		t.Error("Expected error for spec file with no version")
	}

	// Test with non-existent file
	nonExistentPath := filepath.Join(tempDir, "non-existent.md")
	_, err = GetSpecVersion(nonExistentPath)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}
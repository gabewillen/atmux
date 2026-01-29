package spec

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateSpecVersion(t *testing.T) {
	// Test with actual repo root
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	
	// Go up two levels to get to repo root
	repoRoot := filepath.Join(cwd, "..", "..")
	
	// Test that spec file exists and validates
	err = ValidateSpecVersion(repoRoot)
	if err != nil {
		t.Errorf("ValidateSpecVersion failed: %v", err)
	}
}

func TestValidateSpecVersionNotFound(t *testing.T) {
	// Test with non-existent directory
	err := ValidateSpecVersion("/tmp/nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent spec file")
	}
	
	if !IsSpecNotFound(err) {
		t.Errorf("Expected ErrSpecNotFound, got: %v", err)
	}
}

// IsSpecNotFound returns true if the error indicates the spec file was not found.
func IsSpecNotFound(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrSpecNotFound) || 
		strings.Contains(err.Error(), "spec-v1.22.md not found")
}
package main

import (
	"os"
	"testing"
)

// TestSpecVersionExists verifies spec-v1.22.md is present per plan requirement.
func TestSpecVersionExists(t *testing.T) {
	specPath := "../../docs/spec-v1.22.md"
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Fatalf("spec-v1.22.md not found - required per plan Phase 0")
	}
}

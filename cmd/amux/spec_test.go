package main

import (
	"bytes"
	"os"
	"testing"
)

// TestSpecVersionExists verifies spec-v1.22.md is present and has the expected version marker per plan Phase 0.
func TestSpecVersionExists(t *testing.T) {
	specPath := "../../docs/spec-v1.22.md"
	data, err := os.ReadFile(specPath)
	if os.IsNotExist(err) {
		t.Fatalf("spec-v1.22.md not found - required per plan Phase 0")
	}
	if err != nil {
		t.Fatalf("failed to read spec-v1.22.md: %v", err)
	}
	if !bytes.Contains(data, []byte("**Version:** v1.22")) {
		t.Fatalf("spec-v1.22.md does not contain expected version marker '**Version:** v1.22'")
	}
}

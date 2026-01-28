// Package internal provides spec version guard tests.
package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agentflare-ai/amux/pkg/api"
)

// TestSpecVersionGuard ensures the authoritative spec file exists and matches
// the expected version. This test fails fast with a clear error if the spec
// is missing or has a version mismatch.
func TestSpecVersionGuard(t *testing.T) {
	// Find repo root (search upward for go.mod)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	repoRoot := findRepoRoot(wd)
	if repoRoot == "" {
		t.Fatal("could not find repository root (go.mod)")
	}

	specPath := filepath.Join(repoRoot, "docs", "spec-v1.22.md")

	// Check spec file exists
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Fatalf("SPEC VERSION GUARD FAILED: %s does not exist\n"+
			"The authoritative specification file is required for this plan.\n"+
			"Please ensure spec-v1.22.md is present in the docs/ directory.",
			specPath)
	}

	// Read spec file and check version
	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read spec file: %v", err)
	}

	content := string(data)

	// Check for version marker
	expectedVersion := "v1.22"
	if !strings.Contains(content, "**Version:** "+expectedVersion) &&
		!strings.Contains(content, "Version: "+expectedVersion) {
		t.Fatalf("SPEC VERSION GUARD FAILED: spec version mismatch\n"+
			"Expected version: %s\n"+
			"The spec file must contain a version marker matching %s",
			expectedVersion, expectedVersion)
	}

	// Verify api.SpecVersion matches
	if api.SpecVersion != expectedVersion {
		t.Fatalf("SPEC VERSION MISMATCH: api.SpecVersion=%q but expected %q\n"+
			"Update pkg/api/types.go to match the spec version",
			api.SpecVersion, expectedVersion)
	}

	t.Logf("Spec version guard passed: %s", expectedVersion)
}

func findRepoRoot(start string) string {
	dir := start
	for {
		goMod := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goMod); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

package spec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckSpecVersion_MissingFile(t *testing.T) {
	root := t.TempDir()
	// No docs/spec-v1.22.md
	err := CheckSpecVersion(root)
	if err == nil {
		t.Fatal("CheckSpecVersion should fail when spec file is missing")
	}
	if !strings.Contains(err.Error(), "spec file not found") {
		t.Errorf("error should mention spec file not found, got: %v", err)
	}
}

func TestCheckSpecVersion_WrongVersionMarker(t *testing.T) {
	root := t.TempDir()
	docsDir := filepath.Join(root, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatalf("mkdir docs: %v", err)
	}
	specPath := filepath.Join(docsDir, SpecFileName)
	// Write file with wrong version marker
	content := "# Agent Multiplexer\n\n**Version:** v1.0.0\n"
	if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
		t.Fatalf("write spec file: %v", err)
	}
	err := CheckSpecVersion(root)
	if err == nil {
		t.Fatal("CheckSpecVersion should fail when version marker does not match")
	}
	if !strings.Contains(err.Error(), "expected version marker") {
		t.Errorf("error should mention expected version marker, got: %v", err)
	}
}

func TestCheckSpecVersion_ValidMarker(t *testing.T) {
	root := t.TempDir()
	docsDir := filepath.Join(root, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatalf("mkdir docs: %v", err)
	}
	specPath := filepath.Join(docsDir, SpecFileName)
	versionMarker := "**Version:** " + ExpectedSpecVersion
	if err := os.WriteFile(specPath, []byte("# Spec\n\n"+versionMarker+"\n"), 0644); err != nil {
		t.Fatalf("write spec file: %v", err)
	}
	if err := CheckSpecVersion(root); err != nil {
		t.Errorf("CheckSpecVersion should pass with correct marker: %v", err)
	}
}

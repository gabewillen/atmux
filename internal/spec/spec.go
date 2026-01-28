// Package spec provides spec version checking and validation.
package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// ExpectedSpecVersion is the expected spec version for this implementation.
	ExpectedSpecVersion = "v1.22"

	// SpecFileName is the name of the spec file.
	SpecFileName = "spec-v1.22.md"
)

// CheckSpecVersion verifies that the spec file exists and contains the expected version.
func CheckSpecVersion(repoRoot string) error {
	specPath := filepath.Join(repoRoot, "docs", SpecFileName)
	data, err := os.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("spec file not found at %s: %w", specPath, err)
	}

	content := string(data)
	versionMarker := fmt.Sprintf("**Version:** %s", ExpectedSpecVersion)
	if !strings.Contains(content, versionMarker) {
		return fmt.Errorf("spec file at %s does not contain expected version marker %s", specPath, versionMarker)
	}

	return nil
}

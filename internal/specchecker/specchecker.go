// Package specchecker implements verification that spec-v1.22.md is present and version-locked
package specchecker

import (
	"fmt"
	"os"
	"strings"
)

const ExpectedSpecVersion = "v1.22"
const SpecFileName = "spec-v1.22.md"

// CheckSpecPresenceAndVersion verifies that spec-v1.22.md exists and contains the expected version
func CheckSpecPresenceAndVersion(specPath string) error {
	// Check if the file exists
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		return fmt.Errorf("spec file does not exist: %s", specPath)
	} else if err != nil {
		return fmt.Errorf("error accessing spec file: %w", err)
	}

	// Read the file content
	content, err := os.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("error reading spec file: %w", err)
	}

	// Check if the expected version is present in the file
	contentStr := string(content)
	if !strings.Contains(contentStr, "Version: "+ExpectedSpecVersion) &&
	   !strings.Contains(contentStr, "**Version:** "+ExpectedSpecVersion) {
		return fmt.Errorf("expected version %s not found in spec file %s", ExpectedSpecVersion, specPath)
	}

	return nil
}

// GetSpecVersion extracts the version from the spec file
func GetSpecVersion(specPath string) (string, error) {
	content, err := os.ReadFile(specPath)
	if err != nil {
		return "", fmt.Errorf("error reading spec file: %w", err)
	}

	contentStr := string(content)
	
	// Look for the version in the format "Version: vX.YY" or "**Version:** vX.YY"
	lines := strings.Split(contentStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Version:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				version := strings.TrimSpace(parts[1])
				return version, nil
			}
		}
		if strings.Contains(line, "**Version:**") {
			parts := strings.Split(line, "**Version:**")
			if len(parts) >= 2 {
				version := strings.TrimSpace(parts[1])
				return version, nil
			}
		}
	}

	return "", fmt.Errorf("version not found in spec file %s", specPath)
}
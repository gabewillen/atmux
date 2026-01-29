// Package spec provides specification version validation.
// This package ensures the implementation matches the required spec version.
package spec

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Common sentinel errors for spec validation.
var (
	// ErrSpecNotFound indicates the specification file was not found.
	ErrSpecNotFound = errors.New("specification file not found")

	// ErrVersionMismatch indicates the spec version doesn't match expected.
	ErrVersionMismatch = errors.New("specification version mismatch")

	// ErrInvalidSpec indicates the specification file is malformed.
	ErrInvalidSpec = errors.New("invalid specification file")
)

const (
	// RequiredSpecVersion is the required specification version for this implementation.
	RequiredSpecVersion = "v1.22"
	
	// SpecFileName is the expected specification file name.
	SpecFileName = "spec-v1.22.md"
)

// ValidateSpecVersion checks that the required specification file exists and matches the expected version.
// This function implements the spec version lock guard required by Phase 0.
func ValidateSpecVersion(repoRoot string) error {
	specPath := filepath.Join(repoRoot, "docs", SpecFileName)
	
	// Check if spec file exists
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		return fmt.Errorf("%s not found in docs/: %w", SpecFileName, ErrSpecNotFound)
	}
	
	// Read and validate spec file
	file, err := os.Open(specPath)
	if err != nil {
		return fmt.Errorf("failed to open spec file: %w", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	versionFound := false
	
	// Look for version marker in the first few lines
	lineCount := 0
	for scanner.Scan() && lineCount < 20 {
		line := strings.TrimSpace(scanner.Text())
		
		// Look for version line like "**Version:** v1.22"
		if strings.Contains(line, "**Version:**") {
			versionFound = true
			if !strings.Contains(line, RequiredSpecVersion) {
				return fmt.Errorf("spec version mismatch: expected %s, found line: %q: %w", 
					RequiredSpecVersion, line, ErrVersionMismatch)
			}
			break
		}
		lineCount++
	}
	
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading spec file: %w", err)
	}
	
	if !versionFound {
		return fmt.Errorf("version marker not found in %s: %w", SpecFileName, ErrInvalidSpec)
	}
	
	return nil
}

// MustValidateSpecVersion validates the spec version and panics on failure.
// This is suitable for startup validation where the application should not continue
// if the spec version is wrong.
func MustValidateSpecVersion(repoRoot string) {
	if err := ValidateSpecVersion(repoRoot); err != nil {
		panic(fmt.Sprintf("spec version validation failed: %v", err))
	}
}
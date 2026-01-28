// Package inference provides the local inference engine interface for amux.
package inference

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// liquidgenEngine integrates with the liquidgen inference engine from third_party/liquidgen.
// Phase 0: Basic integration that validates models and extracts version info.
type liquidgenEngine struct {
	binaryPath string
	version    string
}

// NewLiquidgenEngine creates a new liquidgen-based inference engine.
// It locates the liquidgen binary and extracts version information.
func NewLiquidgenEngine() (Engine, error) {
	// Find module root to locate third_party/liquidgen
	moduleRoot, err := findModuleRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find module root: %w", err)
	}

	liquidgenDir := filepath.Join(moduleRoot, "third_party", "liquidgen")
	
	// Check if liquidgen directory exists
	if _, err := os.Stat(liquidgenDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("liquidgen directory not found at %s", liquidgenDir)
	}

	// Extract version from git commit hash
	version := getLiquidgenVersion(liquidgenDir)

	// Look for built binary (could be in build/bin or bin/)
	binaryPath := findLiquidgenBinary(liquidgenDir)
	
	// Phase 0: Allow creation even if binary not found (it may need to be built)
	// The binary will be required when actual inference is called in later phases
	if binaryPath == "" {
		// Log that binary is not found but allow engine creation for Phase 0
		_ = binaryPath // Will be used in later phases
	}

	return &liquidgenEngine{
		binaryPath: binaryPath,
		version:    version,
	}, nil
}

func (e *liquidgenEngine) Generate(ctx context.Context, req Request) (Stream, error) {
	// Validate model ID per spec §4.2.10
	if req.Model != "lfm2.5-thinking" && req.Model != "lfm2.5-VL" {
		return nil, fmt.Errorf("%w: %s", ErrUnknownModel, req.Model)
	}

	// Phase 0: Return noop stream - actual inference will be implemented in later phases
	// When implemented, this will call the liquidgen binary or library
	return &noopStream{}, nil
}

// GetLiquidgenVersion returns the liquidgen version/commit identifier for traceability.
func GetLiquidgenVersion() string {
	moduleRoot, err := findModuleRoot()
	if err != nil {
		return "unknown"
	}
	liquidgenDir := filepath.Join(moduleRoot, "third_party", "liquidgen")
	return getLiquidgenVersion(liquidgenDir)
}

// findModuleRoot finds the Go module root directory.
func findModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not in a Go module")
		}
		dir = parent
	}
}

// getLiquidgenVersion extracts the git commit hash from third_party/liquidgen.
func getLiquidgenVersion(liquidgenDir string) string {
	// Try to get git commit hash
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = liquidgenDir
	output, err := cmd.Output()
	if err != nil {
		// Fallback: check CMakeLists.txt for version
		cmakePath := filepath.Join(liquidgenDir, "CMakeLists.txt")
		if data, err := os.ReadFile(cmakePath); err == nil {
			content := string(data)
			// Look for "project(liquidgen VERSION ..."
			if strings.Contains(content, "project(liquidgen VERSION") {
				// Extract version if possible
				lines := strings.Split(content, "\n")
				for _, line := range lines {
					if strings.Contains(line, "project(liquidgen VERSION") {
						// Simple extraction - could be improved
						parts := strings.Fields(line)
						for i, part := range parts {
							if part == "VERSION" && i+1 < len(parts) {
								return fmt.Sprintf("v%s", strings.Trim(parts[i+1], ")"))
							}
						}
					}
				}
			}
		}
		return "unknown"
	}

	commitHash := strings.TrimSpace(string(output))
	if commitHash != "" {
		return fmt.Sprintf("commit-%s", commitHash)
	}
	return "unknown"
}

// findLiquidgenBinary looks for the liquidgen binary in common build locations.
func findLiquidgenBinary(liquidgenDir string) string {
	// Common build locations
	paths := []string{
		filepath.Join(liquidgenDir, "build", "bin", "liquidgen"),
		filepath.Join(liquidgenDir, "build", "bin", "liquidgen_cli"),
		filepath.Join(liquidgenDir, "bin", "liquidgen"),
		filepath.Join(liquidgenDir, "bin", "liquidgen_cli"),
		"liquidgen", // In PATH
	}

	for _, path := range paths {
		if path == "liquidgen" {
			// Check PATH
			if found, err := exec.LookPath("liquidgen"); err == nil {
				return found
			}
		} else {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	return ""
}

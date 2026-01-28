package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExpandHome expands a path starting with ~/ to the user's home directory.
func ExpandHome(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path // Fallback: return as is if home dir cannot be found
	}
	return filepath.Join(home, path[2:])
}

// CanonicalizeRepoRoot canonicalizes a repository root path.
// It expands ~/, converts to absolute path, cleans ./.., and resolves symlinks.
func CanonicalizeRepoRoot(path string) (string, error) {
	expanded := ExpandHome(path)
	abs, err := filepath.Abs(expanded)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Resolve symlinks
	realPath, err := filepath.EvalSymlinks(abs)
	if err != nil {
		// If EvalSymlinks fails (e.g. permission), fall back to Abs + Clean per spec
		return filepath.Clean(abs), nil
	}
	return realPath, nil
}

// DefaultConfigDir returns the default configuration directory.
// ~/.config/amux
func DefaultConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home dir: %w", err)
	}
	return filepath.Join(home, ".config", "amux"), nil
}

// DefaultSocketPath returns the default daemon socket path.
// ~/.amux/amuxd.sock
func DefaultSocketPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home dir: %w", err)
	}
	return filepath.Join(home, ".amux", "amuxd.sock"), nil
}

// DefaultWorktreesDir returns the default worktrees directory.
// .amux/worktrees/
func DefaultWorktreesDir(repoRoot string) string {
	return filepath.Join(repoRoot, ".amux", "worktrees")
}

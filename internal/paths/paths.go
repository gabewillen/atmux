// Package paths provides centralized filesystem path resolution for amux.
//
// All filesystem paths are resolved through this package per spec §4.2.6, §4.2.8.
// This ensures consistent handling of:
// - Home directory expansion (~/)
// - Repo-scoped .amux/ paths
// - Configuration file paths
// - Adapter/plugin registry paths
package paths

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/stateforward/amux/internal/errors"
)

// Resolver resolves filesystem paths with consistent expansion and canonicalization.
type Resolver struct {
	repoRoot string // Canonical repository root path
}

// NewResolver creates a new path resolver for the given repository root.
// repoRoot must be an absolute path to a git repository.
func NewResolver(repoRoot string) (*Resolver, error) {
	if repoRoot == "" {
		return nil, errors.Wrap(errors.ErrInvalidInput, "repo root cannot be empty")
	}
	
	canonical, err := Canonicalize(repoRoot)
	if err != nil {
		return nil, errors.Wrap(err, "canonicalize repo root")
	}
	
	return &Resolver{repoRoot: canonical}, nil
}

// RepoRoot returns the canonical repository root.
func (r *Resolver) RepoRoot() string {
	return r.repoRoot
}

// WorktreePath returns the worktree path for the given agent slug.
// Format: <repo_root>/.amux/worktrees/<agent_slug>/
func (r *Resolver) WorktreePath(agentSlug string) string {
	return filepath.Join(r.repoRoot, ".amux", "worktrees", agentSlug)
}

// AmuxDir returns the .amux directory path.
func (r *Resolver) AmuxDir() string {
	return filepath.Join(r.repoRoot, ".amux")
}

// ProjectConfig returns the project configuration file path.
func (r *Resolver) ProjectConfig() string {
	return filepath.Join(r.repoRoot, ".amux", "config.toml")
}

// ExpandHome expands ~ prefix to user's home directory.
func ExpandHome(path string) (string, error) {
	if !strings.HasPrefix(path, "~/") {
		return path, nil
	}
	
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "get home directory")
	}
	
	return filepath.Join(home, path[2:]), nil
}

// Canonicalize canonicalizes a path per spec §3.23:
// - Expands ~/ to home directory
// - Converts to absolute path
// - Cleans . and .. segments
// - Resolves symbolic links (where possible)
func Canonicalize(path string) (string, error) {
	// Expand home directory
	expanded, err := ExpandHome(path)
	if err != nil {
		return "", err
	}
	
	// Convert to absolute path
	abs, err := filepath.Abs(expanded)
	if err != nil {
		return "", errors.Wrap(err, "get absolute path")
	}
	
	// Clean path
	clean := filepath.Clean(abs)
	
	// Resolve symbolic links (best effort)
	resolved, err := filepath.EvalSymlinks(clean)
	if err != nil {
		// If symlink resolution fails, use the cleaned path
		// This matches spec §3.23 requirement for insufficient permissions/OS support
		return clean, nil
	}
	
	return resolved, nil
}

// UserConfigDir returns the user configuration directory.
// Default: ~/.config/amux/
func UserConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "get home directory")
	}
	return filepath.Join(home, ".config", "amux"), nil
}

// UserConfigFile returns the user configuration file path.
func UserConfigFile() (string, error) {
	dir, err := UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

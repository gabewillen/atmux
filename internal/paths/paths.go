// Package paths provides centralized filesystem path resolution for amux.
// All filesystem paths MUST be resolved through this package to ensure
// consistent handling of config/env overrides and repository root canonicalization.
package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Resolver provides path resolution functionality.
type Resolver struct {
	configDir string
	homeDir   string
	repoRoot  string
}

// NewResolver creates a new path resolver with the given configuration.
func NewResolver(configDir, homeDir, repoRoot string) (*Resolver, error) {
	if homeDir == "" {
		var err error
		homeDir, err = os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
	}

	// Expand ~ in paths
	configDir = expandHome(configDir, homeDir)
	if repoRoot != "" {
		repoRoot = expandHome(repoRoot, homeDir)
		var err error
		repoRoot, err = filepath.Abs(repoRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to canonicalize repo root: %w", err)
		}
	}

	return &Resolver{
		configDir: configDir,
		homeDir:   homeDir,
		repoRoot:  repoRoot,
	}, nil
}

// expandHome expands ~ to the home directory.
func expandHome(path, homeDir string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir, path[2:])
	}
	return path
}

// ConfigDir returns the user configuration directory.
func (r *Resolver) ConfigDir() string {
	return r.configDir
}

// HomeDir returns the user home directory.
func (r *Resolver) HomeDir() string {
	return r.homeDir
}

// RepoRoot returns the canonical repository root path, or empty string if not set.
func (r *Resolver) RepoRoot() string {
	return r.repoRoot
}

// WorktreePath returns the path to an agent's worktree directory.
// The path is relative to the repository root: .amux/worktrees/{agent_slug}/
func (r *Resolver) WorktreePath(agentSlug string) (string, error) {
	if r.repoRoot == "" {
		return "", fmt.Errorf("repo root not set")
	}
	return filepath.Join(r.repoRoot, ".amux", "worktrees", agentSlug), nil
}

// AmuxDir returns the path to the .amux directory in the repository root.
func (r *Resolver) AmuxDir() (string, error) {
	if r.repoRoot == "" {
		return "", fmt.Errorf("repo root not set")
	}
	return filepath.Join(r.repoRoot, ".amux"), nil
}

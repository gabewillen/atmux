// Package paths provides shared path resolution functionality.
// This package resolves all filesystem paths from config/env and repo_root,
// maintaining .amux/ directory structure invariants.
package paths

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Common sentinel errors for path operations.
var (
	// ErrRepoNotFound indicates no git repository was found.
	ErrRepoNotFound = errors.New("git repository not found")

	// ErrInvalidSlug indicates an invalid agent slug.
	ErrInvalidSlug = errors.New("invalid agent slug")

	// ErrPathResolveFailed indicates path resolution failed.
	ErrPathResolveFailed = errors.New("path resolve failed")
)

// Resolver handles filesystem path resolution with .amux/ invariants.
type Resolver struct {
	repoRoot  string
	amuxDir   string
	homeDir   string
}

// NewResolver creates a new path resolver for the given repository.
func NewResolver(repoRoot string) (*Resolver, error) {
	if repoRoot == "" {
		return nil, fmt.Errorf("repository root required: %w", ErrRepoNotFound)
	}

	// Verify it's a git repository
	gitDir := filepath.Join(repoRoot, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return nil, fmt.Errorf("not a git repository %s: %w", repoRoot, ErrRepoNotFound)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	return &Resolver{
		repoRoot: repoRoot,
		amuxDir:  filepath.Join(repoRoot, ".amux"),
		homeDir:  homeDir,
	}, nil
}

// AmuxDir returns the .amux directory path.
func (r *Resolver) AmuxDir() string {
	return r.amuxDir
}

// WorktreeDir returns the worktree directory for the given agent slug.
func (r *Resolver) WorktreeDir(agentSlug string) (string, error) {
	slug, err := normalizeAgentSlug(agentSlug)
	if err != nil {
		return "", fmt.Errorf("invalid agent slug %s: %w", agentSlug, err)
	}

	return filepath.Join(r.amuxDir, "worktrees", slug), nil
}

// SocketPath returns the daemon socket path.
func (r *Resolver) SocketPath() string {
	return filepath.Join(r.homeDir, ".amux", "amuxd.sock")
}

// ConfigDir returns the user configuration directory.
func (r *Resolver) ConfigDir() string {
	return filepath.Join(r.homeDir, ".amux")
}

// SnapshotsDir returns the snapshots directory in the repository.
func (r *Resolver) SnapshotsDir() string {
	return filepath.Join(r.repoRoot, "snapshots")
}

// normalizeAgentSlug creates a valid agent slug per spec requirements:
// lowercase, non-[a-z0-9-] → -, collapse, trim, max 63 chars
func normalizeAgentSlug(slug string) (string, error) {
	if slug == "" {
		return "", fmt.Errorf("empty slug: %w", ErrInvalidSlug)
	}

	// Convert to lowercase
	slug = strings.ToLower(slug)

	// Replace non-alphanumeric with hyphens
	re := regexp.MustCompile(`[^a-z0-9-]`)
	slug = re.ReplaceAllString(slug, "-")

	// Collapse multiple hyphens
	re = regexp.MustCompile(`-+`)
	slug = re.ReplaceAllString(slug, "-")

	// Trim hyphens from start/end
	slug = strings.Trim(slug, "-")

	// Limit to 63 characters
	if len(slug) > 63 {
		slug = slug[:63]
		slug = strings.TrimRight(slug, "-")
	}

	if slug == "" {
		return "", fmt.Errorf("slug normalized to empty: %w", ErrInvalidSlug)
	}

	return slug, nil
}
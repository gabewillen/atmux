// Package paths provides centralized filesystem path resolution for amux.
package paths

import (
	amuxerrors "github.com/agentflare-ai/amux/internal/errors"
	"os"
	"path/filepath"
	"strings"
)

// Resolver provides centralized path resolution.
type Resolver struct {
	repoRoot string
	config   Config
}

// Config holds path configuration.
type Config struct {
	HomeDir      string
	ConfigDir    string
	DataDir      string
	RuntimeDir   string
	SocketPath   string
	RegistryRoot string
	ModelsRoot   string
	HooksRoot    string
}

// NewResolver creates a new path resolver.
func NewResolver(config Config) (*Resolver, error) {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return nil, amuxerrors.Wrap("finding repository root", err)
	}

	return &Resolver{
		repoRoot: repoRoot,
		config:   config,
	}, nil
}

// findRepoRoot walks up the directory tree looking for .git.
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", amuxerrors.Wrap("getting working directory", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", amuxerrors.Wrap("finding git repository", os.ErrNotExist)
		}
		dir = parent
	}
}

// RepoRoot returns the canonical repository root path.
func (r *Resolver) RepoRoot() string {
	return r.repoRoot
}

// AmuxRoot returns the .amux directory within the repo.
func (r *Resolver) AmuxRoot() string {
	return filepath.Join(r.repoRoot, ".amux")
}

// WorktreeRoot returns the worktrees directory for agents.
func (r *Resolver) WorktreeRoot() string {
	return filepath.Join(r.AmuxRoot(), "worktrees")
}

// AgentWorktree returns the worktree path for a specific agent slug.
func (r *Resolver) AgentWorktree(agentSlug string) string {
	normalized := normalizeAgentSlug(agentSlug)
	return filepath.Join(r.WorktreeRoot(), normalized)
}

// normalizeAgentSlug converts agent names to lowercase alphanumerics with hyphens.
func normalizeAgentSlug(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace invalid characters with hyphens
	result := make([]rune, 0, len(name))
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result = append(result, r)
		} else {
			result = append(result, '-')
		}
	}

	// Collapse multiple hyphens
	slug := string(result)
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Limit to 63 chars, then trim hyphens from start and end
	if len(slug) > 63 {
		slug = slug[:63]
	}
	slug = strings.Trim(slug, "-")

	return slug
}

// SocketPath returns the Unix socket path for daemon communication.
func (r *Resolver) SocketPath() string {
	if r.config.SocketPath != "" {
		return r.config.SocketPath
	}
	return filepath.Join(r.config.RuntimeDir, "amux.sock")
}

// ConfigPath returns the path to a config file.
func (r *Resolver) ConfigPath(name string) string {
	return filepath.Join(r.config.ConfigDir, name)
}

// DataPath returns a path in the data directory.
func (r *Resolver) DataPath(parts ...string) string {
	parts = append([]string{r.config.DataDir}, parts...)
	return filepath.Join(parts...)
}

// RegistryPath returns a path in the registry directory.
func (r *Resolver) RegistryPath(parts ...string) string {
	parts = append([]string{r.config.RegistryRoot}, parts...)
	return filepath.Join(parts...)
}

// ModelsPath returns a path in the models directory.
func (r *Resolver) ModelsPath(parts ...string) string {
	parts = append([]string{r.config.ModelsRoot}, parts...)
	return filepath.Join(parts...)
}

// HooksPath returns a path in the hooks directory.
func (r *Resolver) HooksPath(parts ...string) string {
	parts = append([]string{r.config.HooksRoot}, parts...)
	return filepath.Join(parts...)
}

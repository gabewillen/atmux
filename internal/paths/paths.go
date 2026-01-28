package paths

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrRepoRootNotFound is returned when a git repository root cannot be located.
var ErrRepoRootNotFound = errors.New("repo root not found")

// Resolver resolves filesystem paths based on repo root and user home.
type Resolver struct {
	repoRoot string
	homeDir  string
}

// NewResolver creates a resolver rooted at the discovered repo and user home.
func NewResolver(start string) (*Resolver, error) {
	repoRoot, err := FindRepoRoot(start)
	if err != nil {
		return nil, fmt.Errorf("new resolver: %w", err)
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("new resolver: %w", err)
	}
	return &Resolver{repoRoot: repoRoot, homeDir: homeDir}, nil
}

// RepoRoot returns the resolved repository root.
func (r *Resolver) RepoRoot() string {
	return r.repoRoot
}

// HomeDir returns the resolved user home directory.
func (r *Resolver) HomeDir() string {
	return r.homeDir
}

// AmuxRoot returns the repo-scoped .amux directory path.
func (r *Resolver) AmuxRoot() string {
	return filepath.Join(r.repoRoot, ".amux")
}

// WorktreesDir returns the repo-scoped worktrees directory.
func (r *Resolver) WorktreesDir() string {
	return filepath.Join(r.AmuxRoot(), "worktrees")
}

// WorktreePath returns the worktree path for the given agent slug.
func (r *Resolver) WorktreePath(agentSlug string) string {
	return filepath.Join(r.WorktreesDir(), agentSlug)
}

// UserConfigPath returns the user config path (~/.config/amux/config.toml).
func (r *Resolver) UserConfigPath() string {
	return filepath.Join(r.homeDir, ".config", "amux", "config.toml")
}

// UserAdapterConfigPath returns the per-adapter user config path.
func (r *Resolver) UserAdapterConfigPath(adapter string) string {
	return filepath.Join(r.homeDir, ".config", "amux", "adapters", adapter, "config.toml")
}

// ProjectConfigPath returns the repo-scoped config path.
func (r *Resolver) ProjectConfigPath() string {
	return filepath.Join(r.AmuxRoot(), "config.toml")
}

// ProjectAdapterConfigPath returns the per-adapter repo-scoped config path.
func (r *Resolver) ProjectAdapterConfigPath(adapter string) string {
	return filepath.Join(r.AmuxRoot(), "adapters", adapter, "config.toml")
}

// SocketPath returns the default daemon socket path under ~/.amux.
func (r *Resolver) SocketPath() string {
	return filepath.Join(r.homeDir, ".amux", "amuxd.sock")
}

// ExpandHome expands a leading ~/ in the provided path.
func (r *Resolver) ExpandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(r.homeDir, path[2:])
	}
	return path
}

// FindRepoRoot searches upward from start for a git repository root.
func FindRepoRoot(start string) (string, error) {
	if start == "" {
		return "", fmt.Errorf("find repo root: %w", ErrRepoRootNotFound)
	}
	expanded, err := expandHomePath(start)
	if err != nil {
		return "", fmt.Errorf("find repo root: %w", err)
	}
	abs, err := filepath.Abs(expanded)
	if err != nil {
		return "", fmt.Errorf("find repo root: %w", err)
	}
	current := abs
	for {
		gitPath := filepath.Join(current, ".git")
		info, err := os.Stat(gitPath)
		if err == nil {
			if info.IsDir() {
					return canonicalizePath(current)
				}
				if info.Mode().IsRegular() {
					data, readErr := os.ReadFile(gitPath)
					if readErr != nil {
						return "", fmt.Errorf("find repo root: %w", readErr)
					}
					if strings.Contains(string(data), "gitdir:") {
						return canonicalizePath(current)
					}
				}
			}
		if !errors.Is(err, os.ErrNotExist) && err != nil {
			return "", fmt.Errorf("find repo root: %w", err)
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return "", fmt.Errorf("find repo root: %w", ErrRepoRootNotFound)
}

func expandHomePath(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			return home, nil
		}
		return filepath.Join(home, strings.TrimPrefix(path, "~/")), nil
	}
	return path, nil
}

func canonicalizePath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("canonicalize path: %w", err)
	}
	abs = filepath.Clean(abs)
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return abs, nil
	}
	return resolved, nil
}

// SlugifyAgent derives the agent slug per the spec rules.
func SlugifyAgent(name string) string {
	if name == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(name))
	lastDash := false
	for _, r := range strings.ToLower(name) {
		isAllowed := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-'
		if !isAllowed {
			r = '-'
		}
		if r == '-' {
			if lastDash {
				continue
			}
			lastDash = true
		} else {
			lastDash = false
		}
		b.WriteRune(r)
	}
	slug := strings.Trim(b.String(), "-")
	if len(slug) > 63 {
		slug = slug[:63]
		slug = strings.TrimRight(slug, "-")
	}
	return slug
}

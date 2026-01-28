// Package paths provides centralized filesystem path resolution for amux.
//
// This package is the single source of truth for all filesystem path resolution
// in the amux codebase. All subsystems MUST use this package for path resolution
// and MUST NOT hardcode paths.
//
// Path resolution follows these rules:
// - Paths starting with ~/ are expanded to the user's home directory
// - Paths are converted to absolute paths
// - Paths are cleaned (. and .. segments resolved)
// - Symbolic links are resolved where possible
package paths

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Resolver handles filesystem path resolution.
type Resolver struct {
	mu sync.RWMutex

	// homeDir is the user's home directory
	homeDir string

	// configDir is the user config directory (~/.config/amux)
	configDir string

	// dataDir is the user data directory (~/.amux)
	dataDir string

	// repoRoot is the current repository root (if any)
	repoRoot string
}

// DefaultResolver is the default path resolver instance.
var DefaultResolver = &Resolver{}

func init() {
	// Initialize default resolver with home directory
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}
	DefaultResolver.homeDir = home
	DefaultResolver.configDir = filepath.Join(home, ".config", "amux")
	DefaultResolver.dataDir = filepath.Join(home, ".amux")
}

// SetRepoRoot sets the repository root for the resolver.
func (r *Resolver) SetRepoRoot(root string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	canonical, err := r.canonicalize(root)
	if err != nil {
		return err
	}
	r.repoRoot = canonical
	return nil
}

// RepoRoot returns the current repository root.
func (r *Resolver) RepoRoot() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.repoRoot
}

// HomeDir returns the user's home directory.
func (r *Resolver) HomeDir() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.homeDir
}

// ConfigDir returns the user config directory (~/.config/amux).
func (r *Resolver) ConfigDir() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.configDir
}

// DataDir returns the user data directory (~/.amux).
func (r *Resolver) DataDir() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.dataDir
}

// UserConfigFile returns the path to the user config file.
func (r *Resolver) UserConfigFile() string {
	return filepath.Join(r.ConfigDir(), "config.toml")
}

// ProjectConfigFile returns the path to the project config file within the repo.
func (r *Resolver) ProjectConfigFile() string {
	r.mu.RLock()
	root := r.repoRoot
	r.mu.RUnlock()
	if root == "" {
		return ""
	}
	return filepath.Join(root, ".amux", "config.toml")
}

// WorktreeDir returns the worktree directory for an agent.
// Pattern: {repo_root}/.amux/worktrees/{agent_slug}/
func (r *Resolver) WorktreeDir(agentSlug string) string {
	r.mu.RLock()
	root := r.repoRoot
	r.mu.RUnlock()
	if root == "" {
		return ""
	}
	return filepath.Join(root, ".amux", "worktrees", agentSlug)
}

// AdapterDir returns the adapter directory for a given adapter name.
// User adapter config: ~/.config/amux/adapters/{name}/
func (r *Resolver) AdapterDir(name string) string {
	return filepath.Join(r.ConfigDir(), "adapters", name)
}

// ProjectAdapterDir returns the project adapter directory within the repo.
// Project adapter config: .amux/adapters/{name}/
func (r *Resolver) ProjectAdapterDir(name string) string {
	r.mu.RLock()
	root := r.repoRoot
	r.mu.RUnlock()
	if root == "" {
		return ""
	}
	return filepath.Join(root, ".amux", "adapters", name)
}

// PluginDir returns the plugin registry directory.
func (r *Resolver) PluginDir() string {
	return filepath.Join(r.ConfigDir(), "plugins")
}

// DaemonSocketPath returns the daemon socket path.
func (r *Resolver) DaemonSocketPath() string {
	return filepath.Join(r.DataDir(), "amuxd.sock")
}

// NATSDataDir returns the NATS/JetStream data directory.
func (r *Resolver) NATSDataDir() string {
	return filepath.Join(r.DataDir(), "nats")
}

// SnapshotsDir returns the test snapshots directory.
func (r *Resolver) SnapshotsDir() string {
	r.mu.RLock()
	root := r.repoRoot
	r.mu.RUnlock()
	if root == "" {
		return "snapshots"
	}
	return filepath.Join(root, "snapshots")
}

// Resolve resolves a path, expanding ~ and making it absolute.
func (r *Resolver) Resolve(path string) (string, error) {
	expanded := r.ExpandHome(path)
	return r.canonicalize(expanded)
}

// ExpandHome expands a path that starts with ~/ to the user's home directory.
func (r *Resolver) ExpandHome(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}
	r.mu.RLock()
	home := r.homeDir
	r.mu.RUnlock()
	if home == "" {
		return path
	}
	return filepath.Join(home, path[2:])
}

// canonicalize converts a path to its canonical form:
// 1. Expands ~/ to home directory
// 2. Converts to absolute path
// 3. Cleans . and .. segments
// 4. Resolves symbolic links where possible
func (r *Resolver) canonicalize(path string) (string, error) {
	// Expand home directory
	path = r.ExpandHome(path)

	// Convert to absolute
	if !filepath.IsAbs(path) {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		path = filepath.Join(wd, path)
	}

	// Clean the path
	path = filepath.Clean(path)

	// Attempt to resolve symlinks (best effort)
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		return resolved, nil
	}

	// If symlink resolution fails, return the cleaned path
	// This handles cases where the path doesn't exist yet
	return path, nil
}

// FindRepoRoot searches upward from the given directory to find a git repository root.
// Returns an empty string if no repository is found.
func (r *Resolver) FindRepoRoot(startDir string) (string, error) {
	dir, err := r.canonicalize(startDir)
	if err != nil {
		return "", err
	}

	for {
		gitDir := filepath.Join(dir, ".git")
		info, err := os.Stat(gitDir)
		if err == nil {
			// Found .git - could be a directory or a file (for worktrees)
			if info.IsDir() {
				return dir, nil
			}
			// .git file indicates a worktree; still a valid repo root
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			return "", nil
		}
		dir = parent
	}
}

// FindModuleRoot searches upward from the given directory to find a Go module root.
// Returns the directory containing go.mod, or an error if not found.
func (r *Resolver) FindModuleRoot(startDir string) (string, error) {
	dir, err := r.canonicalize(startDir)
	if err != nil {
		return "", err
	}

	for {
		goMod := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goMod); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

// NormalizeAgentSlug normalizes an agent name to a filesystem-safe slug.
// Rules per spec §5.3.1:
// - Convert to lowercase
// - Replace any character not in [a-z0-9-] with -
// - Collapse consecutive - characters to a single -
// - Trim leading and trailing -
// - Truncate to at most 63 characters
// - If the result is empty, use "agent"
func NormalizeAgentSlug(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace invalid characters with -
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		} else {
			result.WriteRune('-')
		}
	}
	slug := result.String()

	// Collapse consecutive hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Trim leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	// Truncate to 63 characters
	if len(slug) > 63 {
		slug = slug[:63]
		// Remove trailing hyphen after truncation
		slug = strings.TrimRight(slug, "-")
	}

	// Use "agent" if empty
	if slug == "" {
		return "agent"
	}

	return slug
}

// Package-level convenience functions that use the default resolver

// Resolve resolves a path using the default resolver.
func Resolve(path string) (string, error) {
	return DefaultResolver.Resolve(path)
}

// ExpandHome expands ~/ using the default resolver.
func ExpandHome(path string) string {
	return DefaultResolver.ExpandHome(path)
}

// FindRepoRoot finds the repository root using the default resolver.
func FindRepoRoot(startDir string) (string, error) {
	return DefaultResolver.FindRepoRoot(startDir)
}

// FindModuleRoot finds the Go module root using the default resolver.
func FindModuleRoot(startDir string) (string, error) {
	return DefaultResolver.FindModuleRoot(startDir)
}

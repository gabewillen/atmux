// Package paths implements a centralized path resolution system for the amux project
package paths

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// Resolver provides centralized path resolution
type Resolver struct {
	config Config
}

// Config holds configuration for the path resolver
type Config struct {
	// BaseDir is the base directory for relative paths
	BaseDir string
	
	// HomeDir is the user's home directory (usually auto-detected)
	HomeDir string
	
	// RepoRoot is the root of the git repository
	RepoRoot string
	
	// CacheDir is the directory for cache files
	CacheDir string
	
	// ConfigDir is the directory for configuration files
	ConfigDir string
}

// New creates a new path resolver with the given configuration
func New(config Config) *Resolver {
	if config.HomeDir == "" {
		usr, err := user.Current()
		if err == nil {
			config.HomeDir = usr.HomeDir
		}
	}
	
	if config.BaseDir == "" {
		config.BaseDir = "."
	}
	
	return &Resolver{config: config}
}

// Resolve expands a path relative to the resolver's configuration
func (r *Resolver) Resolve(path string) string {
	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		path = filepath.Join(r.config.HomeDir, path[2:])
	}
	
	// If it's a relative path, resolve relative to base dir
	if !filepath.IsAbs(path) {
		path = filepath.Join(r.config.BaseDir, path)
	}
	
	// Clean the path
	return filepath.Clean(path)
}

// AmuxDir returns the path to the .amux directory in the repo root
func (r *Resolver) AmuxDir() string {
	return r.Resolve(filepath.Join(r.config.RepoRoot, ".amux"))
}

// WorktreesDir returns the path to the worktrees directory
func (r *Resolver) WorktreesDir() string {
	return r.Resolve(filepath.Join(r.AmuxDir(), "worktrees"))
}

// WorktreeDir returns the path to a specific agent's worktree directory
func (r *Resolver) WorktreeDir(agentSlug string) string {
	return r.Resolve(filepath.Join(r.WorktreesDir(), agentSlug))
}

// AgentBranch returns the git branch name for an agent
func (r *Resolver) AgentBranch(agentSlug string) string {
	return "amux/" + agentSlug
}

// SocketPath returns the path to the amux daemon socket
func (r *Resolver) SocketPath() string {
	return r.Resolve(filepath.Join(r.AmuxDir(), "amuxd.sock"))
}

// CachePath returns a path within the cache directory
func (r *Resolver) CachePath(subpath ...string) string {
	elements := append([]string{r.config.CacheDir}, subpath...)
	return r.Resolve(filepath.Join(elements...))
}

// ConfigPath returns a path within the config directory
func (r *Resolver) ConfigPath(subpath ...string) string {
	elements := append([]string{r.config.ConfigDir}, subpath...)
	return r.Resolve(filepath.Join(elements...))
}

// RepoRoot returns the resolved repository root
func (r *Resolver) RepoRoot() string {
	return r.Resolve(r.config.RepoRoot)
}

// ExpandHome expands the ~ symbol to the user's home directory in the given path
func ExpandHome(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		usr, err := user.Current()
		if err == nil {
			return filepath.Join(usr.HomeDir, path[2:])
		}
	}
	return path
}

// EnsureDir ensures that a directory exists, creating it if necessary
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// EnsureParentDir ensures that the parent directory of a file path exists
func EnsureParentDir(filePath string) error {
	dir := filepath.Dir(filePath)
	return os.MkdirAll(dir, 0755)
}
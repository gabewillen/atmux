package paths

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/agentflare-ai/amux/internal/errors"
)

// Resolver handles path resolution for the application.
type Resolver struct {
	homeDir  string
	repoRoot string
}

// NewResolver creates a new path resolver.
func NewResolver() (*Resolver, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user home directory")
	}
	return &Resolver{
		homeDir: home,
	}, nil
}

// Expand expands a path starting with ~/ to the user's home directory.
func (r *Resolver) Expand(path string) string {
	if path == "~" {
		return r.homeDir
	}
	if len(path) >= 2 && path[0] == '~' && path[1] == os.PathSeparator {
		return filepath.Join(r.homeDir, path[2:])
	}
	return path
}

// Resolve returns the absolute path for the given path.
// It expands ~ and resolves relative paths against CWD.
func (r *Resolver) Resolve(path string) (string, error) {
	expanded := r.Expand(path)
	abs, err := filepath.Abs(expanded)
	if err != nil {
		return "", errors.Wrap(err, "failed to resolve absolute path")
	}
	return abs, nil
}

// ConfigDir returns the default configuration directory.
// Linux: ~/.config/amux
// macOS: ~/Library/Application Support/amux (or ~/.config/amux if preferred, spec says ~/.config/amux)
func (r *Resolver) ConfigDir() string {
	// Spec §4.2.8.2: ~/.config/amux/config.toml
	// adhering to XDG-style for simplicity as implied by spec examples
	return filepath.Join(r.homeDir, ".config", "amux")
}

// ProjectConfigDir returns the project-local configuration directory (.amux).
func (r *Resolver) ProjectConfigDir(root string) string {
	return filepath.Join(root, ".amux")
}

// WorktreesDir returns the worktrees directory within the project config (.amux/worktrees).
func (r *Resolver) WorktreesDir(root string) string {
	return filepath.Join(r.ProjectConfigDir(root), "worktrees")
}

// EnsureDir ensures that the directory exists.
func EnsureDir(path string) error {
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("path exists but is not a directory: %s", path)
		}
		return nil
	}
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return errors.Wrap(err, "failed to check directory")
}

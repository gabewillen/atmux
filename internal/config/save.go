package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/pelletier/go-toml/v2"
)

// SaveUserConfig saves the configuration to the user's config directory.
// It writes to ~/.config/amux/config.toml (or platform equivalent).
func SaveUserConfig(cfg *Config) error {
	dir, err := paths.DefaultConfigDir()
	if err != nil {
		return fmt.Errorf("failed to resolve config dir: %w", err)
	}
	return saveToFile(filepath.Join(dir, "config.toml"), cfg)
}

// SaveProjectConfig saves the configuration to the project's config directory.
// It writes to <repoRoot>/.amux/config.toml.
func SaveProjectConfig(cfg *Config, repoRoot string) error {
	if repoRoot == "" {
		return fmt.Errorf("repoRoot cannot be empty")
	}
	dir := filepath.Join(repoRoot, ".amux")
	return saveToFile(filepath.Join(dir, "config.toml"), cfg)
}

func saveToFile(path string, cfg *Config) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", path, err)
	}
	return nil
}

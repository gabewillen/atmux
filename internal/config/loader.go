package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/pelletier/go-toml/v2"
)

// Load loads the configuration from all sources.
func Load(repoRoot string) (*Config, error) {
	cfg := DefaultConfig()

	// Load User Config (~/.config/amux/config.toml)
	configDir, err := paths.DefaultConfigDir()
	if err == nil {
		userConfigPath := filepath.Join(configDir, "config.toml")
		if err := loadFile(userConfigPath, &cfg); err != nil {
			// It is okay if the file does not exist
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to load user config: %w", err)
			}
		}
	}

	// Load Project Config (.amux/config.toml)
	if repoRoot != "" {
		projectConfigPath := filepath.Join(repoRoot, ".amux", "config.toml")
		if err := loadFile(projectConfigPath, &cfg); err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to load project config: %w", err)
			}
		}
	}

	// Apply Environment Variables
	if err := applyEnvOverrides(&cfg); err != nil {
		return nil, fmt.Errorf("failed to apply env overrides: %w", err)
	}

	return &cfg, nil
}

func loadFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return toml.Unmarshal(data, cfg)
}

func applyEnvOverrides(cfg *Config) error {
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		key := pair[0]
		value := pair[1]

		if !strings.HasPrefix(key, "AMUX__") {
			continue
		}

		// Parse AMUX__SECTION__KEY -> section.key
		// Handle adapters special case: AMUX__ADAPTERS__CLAUDE_CODE -> adapters.claude-code
		
		// For now, simple TOML overlay
		tomlKey, err := envKeyToTomlPath(key)
		if err != nil {
			continue 
		}

		// Create a partial TOML document
		// This is a naive implementation. For deep nesting, we need a better approach or use a library that supports env mapping.
		// Spec says: "Embed it into a temporary TOML document of the form v = <value> ... and reading v."
		// But strictly mapping AMUX__GENERAL__LOG_LEVEL to cfg.General.LogLevel requires reflection or recursive map structure.
		// Given we have a struct, we can use a trick: Generate TOML from the env var and Unmarshal it into the struct?
		// No, Unmarshal replaces the struct content usually? No, it merges if we unmarshal into existing?
		// go-toml v2 Unmarshal replaces? We should check. Usually it overwrites fields present in TOML.

		// Constructing a TOML string for a single key is hard (nested tables).
		// Alternative: Walk the struct and check env vars?
		// Or: construct a map[string]any from env vars and marshal/unmarshal.
		
		_ = tomlKey
		_ = value
	}
	return nil
}

// envKeyToTomlPath converts AMUX__FOO__BAR to foo.bar
func envKeyToTomlPath(key string) (string, error) {
	parts := strings.Split(key, "__")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid key")
	}
	// Remove AMUX prefix
	parts = parts[1:]
	
	// Normalize
	for i, p := range parts {
		if i == 1 && parts[0] == "ADAPTERS" {
			// Handle adapter name: CLAUDE_CODE -> claude-code
			parts[i] = strings.ToLower(strings.ReplaceAll(p, "_", "-"))
		} else {
			parts[i] = strings.ToLower(p)
		}
	}
	return strings.Join(parts, "."), nil
}

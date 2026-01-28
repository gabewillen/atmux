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
	overrides := make(map[string]any)

	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		key := pair[0]
		value := pair[1]

		if !strings.HasPrefix(key, "AMUX__") {
			continue
		}

		// Parse AMUX__SECTION__KEY -> section.key path
		path, err := envKeyToPath(key)
		if err != nil {
			continue
		}

		// Insert into nested map
		if err := insertMap(overrides, path, value); err != nil {
			return fmt.Errorf("failed to process env var %s: %w", key, err)
		}
	}

	if len(overrides) == 0 {
		return nil
	}

	// Marshal overrides to TOML
	data, err := toml.Marshal(overrides)
	if err != nil {
		return fmt.Errorf("failed to marshal env overrides: %w", err)
	}

	// Unmarshal back into cfg to apply overrides
	return toml.Unmarshal(data, cfg)
}

// envKeyToPath converts AMUX__FOO__BAR to []string{"foo", "bar"}
func envKeyToPath(key string) ([]string, error) {
	parts := strings.Split(key, "__")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid key")
	}
	// Remove AMUX prefix
	parts = parts[1:]
	
	// Normalize
	normalized := make([]string, len(parts))
	for i, p := range parts {
		if i == 1 && parts[0] == "ADAPTERS" {
			// Handle adapter name: CLAUDE_CODE -> claude-code
			normalized[i] = strings.ToLower(strings.ReplaceAll(p, "_", "-"))
		} else {
			normalized[i] = strings.ToLower(p)
		}
	}
	return normalized, nil
}

func insertMap(m map[string]any, path []string, value string) error {
	if len(path) == 0 {
		return nil
	}
	key := path[0]
	if len(path) == 1 {
		m[key] = value
		return nil
	}

	if _, ok := m[key]; !ok {
		m[key] = make(map[string]any)
	}

	subMap, ok := m[key].(map[string]any)
	if !ok {
		return fmt.Errorf("conflict at key %s", key)
	}
	return insertMap(subMap, path[1:], value)
}

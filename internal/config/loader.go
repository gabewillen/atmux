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
	cfg.Adapters = make(map[string]map[string]any)

	// 3. Load User Config (~/.config/amux/config.toml)
	configDir, err := paths.DefaultConfigDir()
	if err == nil {
		userConfigPath := filepath.Join(configDir, "config.toml")
		if err := loadFile(userConfigPath, &cfg); err != nil {
			// It is okay if the file does not exist
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to load user config: %w", err)
			}
		}
		
		// 4. Load User Adapter Configs (~/.config/amux/adapters/*/config.toml)
		if err := loadAdapterConfigs(configDir, &cfg); err != nil {
			return nil, fmt.Errorf("failed to load user adapter configs: %w", err)
		}
	}

	// 5. Load Project Config (.amux/config.toml)
	if repoRoot != "" {
		projectConfigPath := filepath.Join(repoRoot, ".amux", "config.toml")
		if err := loadFile(projectConfigPath, &cfg); err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to load project config: %w", err)
			}
		}

		// 6. Load Project Adapter Configs (.amux/adapters/*/config.toml)
		projectAmuxDir := filepath.Join(repoRoot, ".amux")
		if err := loadAdapterConfigs(projectAmuxDir, &cfg); err != nil {
			return nil, fmt.Errorf("failed to load project adapter configs: %w", err)
		}
	}

	// 7. Apply Environment Variables
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

func loadAdapterConfigs(baseDir string, cfg *Config) error {
	adaptersDir := filepath.Join(baseDir, "adapters")
	entries, err := os.ReadDir(adaptersDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		adapterName := entry.Name()
		configPath := filepath.Join(adaptersDir, adapterName, "config.toml")
		
		data, err := os.ReadFile(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}

		var adapterCfg map[string]any
		if err := toml.Unmarshal(data, &adapterCfg); err != nil {
			return fmt.Errorf("failed to parse config for adapter %s: %w", adapterName, err)
		}

		if cfg.Adapters == nil {
			cfg.Adapters = make(map[string]map[string]any)
		}
		
		// Merge into existing adapter config if present
		if existing, ok := cfg.Adapters[adapterName]; ok {
			for k, v := range adapterCfg {
				existing[k] = v
			}
		} else {
			cfg.Adapters[adapterName] = adapterCfg
		}
	}
	return nil
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

		// Spec §4.2.8.3: Attempt to parse value as TOML
		var parsedValue any = value // Default to string
		var wrapper struct {
			V any `toml:"v"`
		}
		if err := toml.Unmarshal([]byte("v = "+value), &wrapper); err == nil {
			parsedValue = wrapper.V
		}

		// Insert into nested map
		if err := insertMap(overrides, path, parsedValue); err != nil {
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

func insertMap(m map[string]any, path []string, value any) error {
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
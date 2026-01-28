package config

// Actor implementation follows in actor.go

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"

	"github.com/agentflare-ai/amux/internal/errors"
	"github.com/agentflare-ai/amux/internal/paths"
)

// LoadConfig loads the configuration respecting the hierarchy.
// root is the repository root for project-level config.
func Load(root string) (*Config, error) {
	resolver, err := paths.NewResolver()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create path resolver")
	}

	// 1. Built-in defaults
	cfg := DefaultConfig()

	// TODO: 2. Adapter defaults (requires adapter discovery, implemented later)

	// 3. User config (~/.config/amux/config.toml)
	userConfigPath := filepath.Join(resolver.ConfigDir(), "config.toml")
	if err := loadFile(userConfigPath, &cfg); err != nil {
		return nil, errors.Wrap(err, "failed to load user config")
	}

	// TODO: 4. User adapter config (~/.config/amux/adapters/{name}/config.toml)

	// 5. Project config (.amux/config.toml)
	if root != "" {
		projectConfigPath := filepath.Join(resolver.ProjectConfigDir(root), "config.toml")
		if err := loadFile(projectConfigPath, &cfg); err != nil {
			return nil, errors.Wrap(err, "failed to load project config")
		}
	}

	// TODO: 6. Project adapter config (.amux/adapters/{name}/config.toml)

	// 7. Environment variables (AMUX__*)
	if err := loadEnv(&cfg); err != nil {
		return nil, errors.Wrap(err, "failed to load environment variables")
	}

	return &cfg, nil
}

func loadFile(path string, cfg *Config) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Optional
		}
		return errors.Wrap(err, "failed to open config file")
	}
	defer f.Close()

	if err := toml.NewDecoder(f).Decode(cfg); err != nil {
		return errors.Wrap(err, "failed to decode config file")
	}
	return nil
}

// loadEnv is implemented in env.go

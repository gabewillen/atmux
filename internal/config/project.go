// Package config provides configuration management for amux.
// project.go handles reading and writing project-scoped config (.amux/config.toml) for agent persistence.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// ProjectConfigPath returns the path to the project config file under repoRoot.
func ProjectConfigPath(repoRoot string) string {
	return filepath.Join(repoRoot, ".amux", "config.toml")
}

// LoadProjectFile reads the project config from .amux/config.toml under repoRoot.
// If the file does not exist, returns a minimal config with empty agents.
func LoadProjectFile(repoRoot string) (*Config, error) {
	path := ProjectConfigPath(repoRoot)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Agents: []AgentConfig{}, Adapters: make(map[string]interface{})}, nil
		}
		return nil, fmt.Errorf("read project config: %w", err)
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse project config: %w", err)
	}
	if cfg.Agents == nil {
		cfg.Agents = []AgentConfig{}
	}
	if cfg.Adapters == nil {
		cfg.Adapters = make(map[string]interface{})
	}
	return &cfg, nil
}

// SaveProjectFile writes cfg to .amux/config.toml under repoRoot.
// Creates .amux directory if needed.
func SaveProjectFile(repoRoot string, cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	path := ProjectConfigPath(repoRoot)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create .amux dir: %w", err)
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal project config: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write project config: %w", err)
	}
	return nil
}

// AddAgentToProject appends agent to the agents list in project config and saves (spec §5.2).
// Ensures .amux/config.toml exists and is updated.
func AddAgentToProject(repoRoot string, agent AgentConfig) error {
	cfg, err := LoadProjectFile(repoRoot)
	if err != nil {
		return err
	}
	cfg.Agents = append(cfg.Agents, agent)
	return SaveProjectFile(repoRoot, cfg)
}

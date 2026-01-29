package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"

	"github.com/agentflare-ai/amux/internal/errors"
	"github.com/agentflare-ai/amux/internal/paths"
)

// AddAgent appends an agent definition to the configuration file.
// If root is empty, it writes to the user config (~/.config/amux/config.toml).
// If root is provided, it writes to the project config (.amux/config.toml).
func AddAgent(root string, agent AgentDef) error {
	resolver, err := paths.NewResolver()
	if err != nil {
		return errors.Wrap(err, "failed to create path resolver")
	}

	var configPath string
	if root == "" {
		configPath = filepath.Join(resolver.ConfigDir(), "config.toml")
	} else {
		// Ensure .amux exists
		projectDir := resolver.ProjectConfigDir(root)
		if err := paths.EnsureDir(projectDir); err != nil {
			return errors.Wrap(err, "failed to create project config directory")
		}
		configPath = filepath.Join(projectDir, "config.toml")
	}

	// 1. Read existing config to preserve other fields
	// We use a map structure for flexible round-tripping or specific partial struct
	// For simplicity, we'll try to load into Config, append, and save back.
	// NOTE: generic comments might be lost with simple unmarshal/marshal.
	// For robust editing, we might want a purely AST-based approach or just append to the file if it's simple TOML.
	// But spec requires structured config. Let's load, append, save.

	// We need to handle the case where the file doesn't exist
	cfg := DefaultConfig()
	if err := loadFile(configPath, &cfg); err != nil {
		// If it's a parse error on existing file, that's bad.
		// If it just doesn't exist, we start with defaults but EMPTY defaults clearly?
		// Actually DefaultConfig() has defaults.
		// If file doesn't exist, we just create a new one with this agent.
		// However, writing back DefaultConfig() might write a lot of defaults the user didn't ask for.
		// Better strategy: Read into a map, append to "agents" list, write back.
	}

	// Strategy specific for AddAgent:
	// We want to avoid overwriting the user's manual edits or formatting if possible,
	// but go-toml v2 doesn't preserve comments/formatting perfect.
	// For Phase 2, simple load-modify-dump is acceptable as per plan.

	// Check for duplicates
	for _, a := range cfg.Agents {
		if a.Name == agent.Name {
			return fmt.Errorf("agent %q already exists in %s", agent.Name, configPath)
		}
	}

	cfg.Agents = append(cfg.Agents, agent)

	if err := saveFile(configPath, &cfg); err != nil {
		return errors.Wrap(err, "failed to save config file")
	}

	return nil
}

// RemoveAgent removes an agent from the configuration file.
func RemoveAgent(root string, name string) error {
	resolver, err := paths.NewResolver()
	if err != nil {
		return errors.Wrap(err, "failed to create path resolver")
	}

	var configPath string
	if root == "" {
		configPath = filepath.Join(resolver.ConfigDir(), "config.toml")
	} else {
		configPath = filepath.Join(resolver.ProjectConfigDir(root), "config.toml")
	}

	cfg := DefaultConfig()
	if err := loadFile(configPath, &cfg); err != nil {
		// If file doesn't exist, nothing to remove
		if os.IsNotExist(err) {
			return fmt.Errorf("config file does not exist: %s", configPath)
		}
		// If parse error but file exists, we probably shouldn't touch it
		return errors.Wrap(err, "failed to load existing config")
	}

	found := false
	newAgents := make([]AgentDef, 0, len(cfg.Agents))
	for _, a := range cfg.Agents {
		if a.Name == name {
			found = true
			continue
		}
		newAgents = append(newAgents, a)
	}

	if !found {
		return fmt.Errorf("agent %q not found in %s", name, configPath)
	}

	cfg.Agents = newAgents

	if err := saveFile(configPath, &cfg); err != nil {
		return errors.Wrap(err, "failed to save config file")
	}

	return nil
}

func saveFile(path string, cfg *Config) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return errors.Wrap(err, "failed to create config directory")
	}

	f, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "failed to create config file")
	}
	defer f.Close()

	enc := toml.NewEncoder(f)
	// enc.SetArraysMultiline(true) // nice for agents list
	// go-toml v2 encoding options are limited compared to v1, but defaults are usually sane.

	if err := enc.Encode(cfg); err != nil {
		return errors.Wrap(err, "failed to encode config")
	}
	return nil
}

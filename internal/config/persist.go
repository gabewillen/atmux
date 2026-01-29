package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// LoadConfigFile reads and parses a TOML config file.
func LoadConfigFile(path string) (map[string]any, error) {
	if path == "" {
		return map[string]any{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, fmt.Errorf("load config file: %w", err)
	}
	parsed, err := ParseTOML(data)
	if err != nil {
		return nil, fmt.Errorf("load config file: %w", err)
	}
	return parsed, nil
}

// WriteConfigFile writes a TOML config file.
func WriteConfigFile(path string, data map[string]any) error {
	if path == "" {
		return fmt.Errorf("write config file: path is empty")
	}
	encoded, err := EncodeTOML(data)
	if err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	if err := os.WriteFile(path, encoded, 0o644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}

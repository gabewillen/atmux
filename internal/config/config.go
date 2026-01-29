// Package config provides configuration management with TOML format support.
// This package handles the configuration hierarchy: built-in < adapter < user < project < env
// with env vars using AMUX__ prefix. Adapter configs are treated as opaque.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// Common sentinel errors for configuration operations.
var (
	// ErrConfigNotFound indicates a configuration file was not found.
	ErrConfigNotFound = errors.New("config not found")

	// ErrInvalidConfig indicates invalid configuration data.
	ErrInvalidConfig = errors.New("invalid config")

	// ErrLoadFailed indicates configuration loading failed.
	ErrLoadFailed = errors.New("config load failed")
)

// Config represents the amux configuration structure.
type Config struct {
	// Core daemon settings
	Daemon struct {
		SocketPath string `toml:"socket_path"`
		LogLevel   string `toml:"log_level"`
	} `toml:"daemon"`

	// Agent configurations (opaque to core)
	Agents map[string]interface{} `toml:"agents"`

	// Remote settings for NATS-based multi-host orchestration
	Remote struct {
		Enabled        bool   `toml:"enabled"`
		Hub            string `toml:"hub"`
		RequestTimeout string `toml:"request_timeout"` // duration string
		BufferSize     int    `toml:"buffer_size"`
		Manager        struct {
			Enabled bool `toml:"enabled"`
		} `toml:"manager"`
		NATS struct {
			URL           string `toml:"url"`
			CredsPath     string `toml:"creds_path"`
			SubjectPrefix string `toml:"subject_prefix"`
			KVBucket      string `toml:"kv_bucket"`
		} `toml:"nats"`
	} `toml:"remote"`
}

// Load reads configuration from files and environment variables.
// Implements the hierarchy: built-in < adapter < user < project < env
func Load() (*Config, error) {
	config := &Config{}

	// Apply defaults
	config.Daemon.SocketPath = filepath.Join(os.Getenv("HOME"), ".amux", "amuxd.sock")
	config.Daemon.LogLevel = "info"
	config.Remote.Enabled = false
	config.Remote.RequestTimeout = "30s"
	config.Remote.BufferSize = 16384  // 16KB default buffer
	config.Remote.Manager.Enabled = false
	config.Remote.NATS.SubjectPrefix = "amux"
	config.Remote.NATS.KVBucket = "AMUX_KV"
	config.Agents = make(map[string]interface{})

	// Load from TOML files (implementation deferred)
	if err := loadFromFiles(config); err != nil {
		return nil, fmt.Errorf("failed to load config files: %w", err)
	}

	// Override with environment variables
	if err := loadFromEnv(config); err != nil {
		return nil, fmt.Errorf("failed to load environment config: %w", err)
	}

	return config, nil
}

// loadFromFiles loads configuration from TOML files.
// Implementation deferred to Phase 0 completion.
func loadFromFiles(config *Config) error {
	// File loading not yet implemented
	return nil
}

// loadFromEnv applies environment variable overrides with AMUX__ prefix.
func loadFromEnv(config *Config) error {
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, "AMUX__") {
			continue
		}
		
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimPrefix(parts[0], "AMUX__")
		value := parts[1]

		// Simple env var mapping - will be expanded in later phases
		switch strings.ToLower(key) {
		case "daemon_socket_path":
			config.Daemon.SocketPath = value
		case "daemon_log_level":
			config.Daemon.LogLevel = value
		case "remote_enabled":
			config.Remote.Enabled = (value == "true")
		case "remote_hub":
			config.Remote.Hub = value
		case "remote_request_timeout":
			config.Remote.RequestTimeout = value
		case "remote_buffer_size":
			if size := parseInt(value); size > 0 {
				config.Remote.BufferSize = size
			}
		case "remote_manager_enabled":
			config.Remote.Manager.Enabled = (value == "true")
		case "remote_nats_url":
			config.Remote.NATS.URL = value
		case "remote_nats_creds_path":
			config.Remote.NATS.CredsPath = value
		case "remote_nats_subject_prefix":
			config.Remote.NATS.SubjectPrefix = value
		case "remote_nats_kv_bucket":
			config.Remote.NATS.KVBucket = value
		}
	}

	return nil
}

// parseInt safely parses an integer string, returning 0 on error.
func parseInt(s string) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return 0
}

// SaveToFile writes configuration to a TOML file.
func (c *Config) SaveToFile(path string) error {
	if path == "" {
		return fmt.Errorf("file path required: %w", ErrInvalidConfig)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file %s: %w", path, ErrLoadFailed)
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to encode config to %s: %w", path, ErrLoadFailed)
	}

	return nil
}
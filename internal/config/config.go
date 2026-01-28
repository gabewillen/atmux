// Package config provides configuration loading and management for amux.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	amuxerrors "github.com/agentflare-ai/amux/internal/errors"
)

// Config represents the complete amux configuration.
type Config struct {
	// Core configuration
	Core CoreConfig `toml:"core"`

	// Agent management configuration
	Agents AgentsConfig `toml:"agents"`

	// Remote configuration
	Remote RemoteConfig `toml:"remote"`

	// Event system configuration
	Events EventsConfig `toml:"events"`

	// Process tracking configuration
	Process ProcessConfig `toml:"process"`

	// PTY monitoring configuration
	Monitor MonitorConfig `toml:"monitor"`

	// Adapter configuration (opaque per-adapter blocks)
	Adapters map[string]map[string]interface{} `toml:"adapters"`

	// OpenTelemetry configuration
	OTel OTelConfig `toml:"otel"`

	// Inference configuration
	Inference InferenceConfig `toml:"inference"`
}

// CoreConfig contains core daemon configuration.
type CoreConfig struct {
	// Data directory for persistent state
	DataDir string `toml:"data_dir"`

	// Runtime directory for sockets and temporary files
	RuntimeDir string `toml:"runtime_dir"`

	// Log level (trace, debug, info, warn, error)
	LogLevel string `toml:"log_level"`

	// Whether to run in debug mode
	Debug bool `toml:"debug"`
}

// AgentsConfig contains agent management settings.
type AgentsConfig struct {
	// Default worktree strategy
	DefaultStrategy string `toml:"default_strategy"`

	// Auto-cleanup worktrees on remove
	AutoCleanup bool `toml:"auto_cleanup"`

	// Maximum concurrent agents
	MaxConcurrent int `toml:"max_concurrent"`
}

// RemoteConfig contains remote orchestration settings.
type RemoteConfig struct {
	// NATS server URL
	ServerURL string `toml:"server_url"`

	// Credentials file path
	CredsPath string `toml:"creds_path"`

	// Subject prefix for all subjects
	SubjectPrefix string `toml:"subject_prefix"`

	// Request timeout
	RequestTimeout time.Duration `toml:"request_timeout"`

	// Buffer size for PTY output replay
	BufferSize int `toml:"buffer_size"`

	// JetStream configuration
	JetStream JetStreamConfig `toml:"jetstream"`
}

// JetStreamConfig contains JetStream-specific settings.
type JetStreamConfig struct {
	// KV bucket name
	BucketName string `toml:"bucket_name"`

	// Stream name for events
	StreamName string `toml:"stream_name"`

	// Domain for JetStream
	Domain string `toml:"domain"`
}

// EventsConfig contains event system settings.
type EventsConfig struct {
	// Enable event subscriptions
	Subscriptions SubscriptionsConfig `toml:"subscriptions"`
}

// SubscriptionsConfig contains subscription settings.
type SubscriptionsConfig struct {
	// Enable MCP server for subscriptions
	Enabled bool `toml:"enabled"`

	// Socket path for MCP server
	SocketPath string `toml:"socket_path"`

	// Maximum concurrent subscribers
	MaxConcurrent int `toml:"max_concurrent"`
}

// ProcessConfig contains process tracking settings.
type ProcessConfig struct {
	// Enable process interception
	InterceptionEnabled bool `toml:"interception_enabled"`

	// Hook library path
	HookPath string `toml:"hook_path"`

	// Fallback polling interval
	PollingInterval time.Duration `toml:"polling_interval"`

	// I/O capture mode
	CaptureMode string `toml:"capture_mode"`

	// Batch size for events
	BatchSize int `toml:"batch_size"`

	// Batch timeout
	BatchTimeout time.Duration `toml:"batch_timeout"`
}

// MonitorConfig contains PTY monitoring settings.
type MonitorConfig struct {
	// Activity timeout
	ActivityTimeout time.Duration `toml:"activity_timeout"`

	// Pattern matching timeout
	PatternTimeout time.Duration `toml:"pattern_timeout"`

	// Enable TUI decoding
	TUIEnabled bool `toml:"tui_enabled"`
}

// OTelConfig contains OpenTelemetry settings.
type OTelConfig struct {
	// Enable OpenTelemetry
	Enabled bool `toml:"enabled"`

	// Service name
	ServiceName string `toml:"service_name"`

	// Service version
	ServiceVersion string `toml:"service_version"`

	// Exporter configuration
	Exporter OTelExporterConfig `toml:"exporter"`
}

// OTelExporterConfig contains OTel exporter settings.
type OTelExporterConfig struct {
	// Exporter type (stdout, otlp, etc.)
	Type string `toml:"type"`

	// Endpoint for OTLP exporter
	Endpoint string `toml:"endpoint"`

	// Headers for OTLP exporter
	Headers map[string]string `toml:"headers"`
}

// InferenceConfig contains local inference settings.
type InferenceConfig struct {
	// Enable local inference
	Enabled bool `toml:"enabled"`

	// Inference engine type
	Engine string `toml:"engine"`

	// Model configuration
	Models map[string]ModelConfig `toml:"models"`
}

// ModelConfig contains model-specific settings.
type ModelConfig struct {
	// Model type (embedding, generation, etc.)
	Type string `toml:"type"`

	// Model path or identifier
	Path string `toml:"path"`

	// Model parameters
	Parameters map[string]interface{} `toml:"parameters"`
}

// Load loads configuration from multiple sources in priority order.
func Load() (*Config, error) {
	config := &Config{}

	// Start with built-in defaults
	setDefaults(config)

	// Load adapter configurations (per-adapter config files)
	if err := loadAdapterConfigs(config); err != nil {
		return nil, amuxerrors.Wrap("loading adapter configs", err)
	}

	// Load user config
	if err := loadUserConfig(config); err != nil {
		return nil, amuxerrors.Wrap("loading user config", err)
	}

	// Load project config (in git repo)
	if err := loadProjectConfig(config); err != nil && !os.IsNotExist(err) {
		return nil, amuxerrors.Wrap("loading project config", err)
	}

	// Apply environment variable overrides
	if err := applyEnvOverrides(config); err != nil {
		return nil, amuxerrors.Wrap("applying env overrides", err)
	}

	return config, nil
}

// setDefaults sets built-in default values.
func setDefaults(config *Config) {
	config.Core.LogLevel = "info"
	config.Core.Debug = false

	config.Agents.DefaultStrategy = "merge-commit"
	config.Agents.AutoCleanup = true
	config.Agents.MaxConcurrent = 10

	config.Remote.RequestTimeout = 30 * time.Second
	config.Remote.BufferSize = 1024 * 1024 // 1MB
	config.Remote.SubjectPrefix = "amux"

	config.Remote.JetStream.BucketName = "AMUX_KV"
	config.Remote.JetStream.StreamName = "AMUX_EVENTS"

	config.Events.Subscriptions.Enabled = false
	config.Events.Subscriptions.MaxConcurrent = 100

	config.Process.InterceptionEnabled = true
	config.Process.PollingInterval = 1 * time.Second
	config.Process.CaptureMode = "full"
	config.Process.BatchSize = 100
	config.Process.BatchTimeout = 5 * time.Second

	config.Monitor.ActivityTimeout = 30 * time.Second
	config.Monitor.PatternTimeout = 10 * time.Second
	config.Monitor.TUIEnabled = true

	config.OTel.Enabled = false
	config.OTel.ServiceName = "amux"
	config.OTel.ServiceVersion = "dev"
	config.OTel.Exporter.Type = "stdout"

	config.Inference.Enabled = true
	config.Inference.Engine = "liquidgen"
}

// loadAdapterConfigs loads adapter-specific configurations.
func loadAdapterConfigs(config *Config) error {
	// For Phase 0, load from standard discovery paths
	configPaths := []string{
		"~/.config/amux/adapters",
		".amux/adapters",
	}

	loadedConfigs := make(map[string]map[string]interface{})

	for _, configPath := range configPaths {
		// Expand ~ to home directory
		if strings.HasPrefix(configPath, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				continue
			}
			configPath = filepath.Join(home, configPath[2:])
		}

		// Check if directory exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			continue
		}

		// Read adapter configs
		entries, err := os.ReadDir(configPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			adapterPath := filepath.Join(configPath, entry.Name(), "config.toml")
			if _, err := os.Stat(adapterPath); err != nil {
				continue
			}

			data, err := os.ReadFile(adapterPath)
			if err != nil {
				continue
			}

			var adapterConfig map[string]interface{}
			if err := toml.Unmarshal(data, &adapterConfig); err != nil {
				continue
			}

			loadedConfigs[entry.Name()] = adapterConfig
		}
	}

	config.Adapters = loadedConfigs
	return nil
}

// loadUserConfig loads user configuration from standard paths.
func loadUserConfig(config *Config) error {
	// For Phase 0, check standard XDG config locations
	configPaths := []string{
		"~/.config/amux/config.toml",
		"~/.amux/config.toml",
	}

	for _, configPath := range configPaths {
		// Expand ~ to home directory
		if strings.HasPrefix(configPath, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				continue
			}
			configPath = filepath.Join(home, configPath[2:])
		}

		// Check if file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			continue
		}

		// Load and merge configuration
		data, err := os.ReadFile(configPath)
		if err != nil {
			continue
		}

		if err := toml.Unmarshal(data, config); err != nil {
			return amuxerrors.Wrap("parsing user config", err)
		}

		// Found and loaded successfully
		return nil
	}

	// No user config found - use defaults
	return nil
}

// loadProjectConfig loads project-specific configuration from git repo.
func loadProjectConfig(config *Config) error {
	// For Phase 0, check for .amux/config.toml in current directory and parent directories
	projectConfig := ".amux/config.toml"

	// Check if file exists
	if _, err := os.Stat(projectConfig); os.IsNotExist(err) {
		// Try parent directories (for nested git repos)
		for i := 0; i < 5; i++ {
			projectConfig = filepath.Join("..", projectConfig)
			if _, err := os.Stat(projectConfig); os.IsNotExist(err) {
				continue
			}
			break
		}

		// Still not found, that's ok for Phase 0
		return nil
	}

	// Load and merge configuration
	data, err := os.ReadFile(projectConfig)
	if err != nil {
		return amuxerrors.Wrap("reading project config", err)
	}

	if err := toml.Unmarshal(data, config); err != nil {
		return amuxerrors.Wrap("parsing project config", err)
	}

	return nil
}

// applyEnvOverrides applies environment variable overrides.
func applyEnvOverrides(config *Config) error {
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

		if err := setEnvOverride(config, key, value); err != nil {
			return amuxerrors.Wrap("setting env override", err)
		}
	}

	return nil
}

// setEnvOverride sets a single environment variable override.
func setEnvOverride(config *Config, key, value string) error {
	// Parse AMUX__KEY format with nested hierarchy
	parts := strings.Split(key, "__")
	if len(parts) < 2 {
		return nil // Skip invalid keys
	}

	// Normalize key parts to lowercase
	for i, part := range parts {
		parts[i] = strings.ToLower(part)
	}

	// Set nested configuration based on key path
	switch parts[0] {
	case "core":
		if len(parts) >= 2 {
			switch parts[1] {
			case "log_level":
				config.Core.LogLevel = value
			case "debug":
				config.Core.Debug = value == "true"
			case "data_dir":
				config.Core.DataDir = value
			case "runtime_dir":
				config.Core.RuntimeDir = value
			}
		}
	case "remote":
		if len(parts) >= 2 {
			switch parts[1] {
			case "server_url":
				config.Remote.ServerURL = value
			case "creds_path":
				config.Remote.CredsPath = value
			case "subject_prefix":
				config.Remote.SubjectPrefix = value
			case "request_timeout":
				if dur, err := time.ParseDuration(value); err == nil {
					config.Remote.RequestTimeout = dur
				}
			case "buffer_size":
				if size, err := parseByteSize(value); err == nil {
					config.Remote.BufferSize = size
				}
			}
		}
	case "otel":
		if len(parts) >= 2 {
			switch parts[1] {
			case "enabled":
				config.OTel.Enabled = value == "true"
			case "service_name":
				config.OTel.ServiceName = value
			case "service_version":
				config.OTel.ServiceVersion = value
			case "exporter":
				if len(parts) >= 3 {
					switch parts[2] {
					case "type":
						config.OTel.Exporter.Type = value
					case "endpoint":
						config.OTel.Exporter.Endpoint = value
					}
				}
			}
		}
	case "inference":
		if len(parts) >= 2 {
			switch parts[1] {
			case "enabled":
				config.Inference.Enabled = value == "true"
			case "engine":
				config.Inference.Engine = value
			}
		}
	}

	return nil
}

// parseByteSize parses byte size strings like "1MB", "64KB" etc.
func parseByteSize(s string) (int, error) {
	// Simple implementation for Phase 0
	if len(s) == 0 {
		return 0, amuxerrors.ErrInvalidConfig
	}

	// Extract numeric part
	var numStr string
	var unit string

	for i, r := range s {
		if r >= '0' && r <= '9' || r == '.' {
			numStr += string(r)
		} else {
			unit = s[i:]
			break
		}
	}

	// Parse numeric part
	var value float64
	fmt.Sscanf(numStr, "%f", &value)

	// Apply unit multiplier
	switch strings.ToUpper(unit) {
	case "B", "":
		return int(value), nil
	case "KB":
		return int(value * 1024), nil
	case "MB":
		return int(value * 1024 * 1024), nil
	case "GB":
		return int(value * 1024 * 1024 * 1024), nil
	default:
		return 0, amuxerrors.ErrInvalidConfig
	}
}

// Save saves configuration to TOML format.
func Save(config *Config, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return amuxerrors.Wrap("creating config file", err)
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return amuxerrors.Wrap("encoding config", err)
	}

	return nil
}

// Validate validates configuration values.
func Validate(config *Config) error {
	// TODO: implement comprehensive validation
	if config.Core.LogLevel == "" {
		return amuxerrors.ErrInvalidConfig
	}

	return nil
}

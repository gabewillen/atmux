// Package config provides configuration loading and management for amux.
package config

import (
	"os"
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
	// TODO: implement adapter config discovery from registry paths
	return nil
}

// loadUserConfig loads user configuration from standard paths.
func loadUserConfig(config *Config) error {
	// TODO: implement user config loading from XDG config dirs
	return nil
}

// loadProjectConfig loads project-specific configuration from git repo.
func loadProjectConfig(config *Config) error {
	// TODO: implement project config loading from .amux/config.toml
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
	// TODO: implement proper env override parsing based on key hierarchy
	// For now, just handle common cases
	switch key {
	case "LOG_LEVEL":
		config.Core.LogLevel = value
	case "DEBUG":
		config.Core.Debug = value == "true"
	case "DATA_DIR":
		config.Core.DataDir = value
	case "RUNTIME_DIR":
		config.Core.RuntimeDir = value
	default:
		// TODO: implement nested key parsing (CORE__LOG_LEVEL, REMOTE__SERVER_URL, etc.)
	}

	return nil
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

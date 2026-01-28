// Package config provides configuration management for amux.
//
// Configuration is loaded in a hierarchical order where later sources
// override earlier ones:
//  1. Built-in defaults
//  2. Adapter defaults (from WASM or config.default.toml)
//  3. User config (~/.config/amux/config.toml)
//  4. User adapter config (~/.config/amux/adapters/{name}/config.toml)
//  5. Project config (.amux/config.toml)
//  6. Project adapter config (.amux/adapters/{name}/config.toml)
//  7. Environment variables (AMUX__* prefix)
//
// This package follows the configuration conventions in spec §4.2.8.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pelletier/go-toml/v2"

	"github.com/agentflare-ai/amux/internal/paths"
)

// Config holds the complete application configuration.
type Config struct {
	General   GeneralConfig   `toml:"general"`
	Timeouts  TimeoutsConfig  `toml:"timeouts"`
	Process   ProcessConfig   `toml:"process"`
	Git       GitConfig       `toml:"git"`
	Events    EventsConfig    `toml:"events"`
	Remote    RemoteConfig    `toml:"remote"`
	NATS      NATSConfig      `toml:"nats"`
	Node      NodeConfig      `toml:"node"`
	Daemon    DaemonConfig    `toml:"daemon"`
	Plugins   PluginsConfig   `toml:"plugins"`
	Telemetry TelemetryConfig `toml:"telemetry"`
	Adapters  map[string]any  `toml:"adapters"`
	Agents    []AgentConfig   `toml:"agents"`
}

// GeneralConfig holds general application settings.
type GeneralConfig struct {
	LogLevel  string `toml:"log_level"`
	LogFormat string `toml:"log_format"`
}

// TimeoutsConfig holds timeout settings.
type TimeoutsConfig struct {
	Idle  Duration `toml:"idle"`
	Stuck Duration `toml:"stuck"`
}

// ProcessConfig holds process tracking settings.
type ProcessConfig struct {
	CaptureMode      string   `toml:"capture_mode"`
	StreamBufferSize ByteSize `toml:"stream_buffer_size"`
	HookMode         string   `toml:"hook_mode"`
	PollInterval     Duration `toml:"poll_interval"`
	HookSocketDir    string   `toml:"hook_socket_dir"`
}

// GitConfig holds git-related settings.
type GitConfig struct {
	Merge GitMergeConfig `toml:"merge"`
}

// GitMergeConfig holds git merge settings.
type GitMergeConfig struct {
	Strategy     string `toml:"strategy"`
	AllowDirty   bool   `toml:"allow_dirty"`
	TargetBranch string `toml:"target_branch"`
}

// EventsConfig holds event system settings.
type EventsConfig struct {
	BatchWindow    Duration          `toml:"batch_window"`
	BatchMaxEvents int               `toml:"batch_max_events"`
	BatchMaxBytes  ByteSize          `toml:"batch_max_bytes"`
	BatchIdleFlush Duration          `toml:"batch_idle_flush"`
	Coalesce       CoalesceConfig    `toml:"coalesce"`
	Subscriptions  SubscriptionConfig `toml:"subscriptions"`
}

// CoalesceConfig holds event coalescing settings.
type CoalesceConfig struct {
	IOStreams bool `toml:"io_streams"`
	Presence  bool `toml:"presence"`
	Activity  bool `toml:"activity"`
}

// SubscriptionConfig holds event subscription settings.
type SubscriptionConfig struct {
	Enabled    bool   `toml:"enabled"`
	SocketPath string `toml:"socket_path"`
}

// RemoteConfig holds remote agent settings.
type RemoteConfig struct {
	Transport             string       `toml:"transport"`
	BufferSize            ByteSize     `toml:"buffer_size"`
	RequestTimeout        Duration     `toml:"request_timeout"`
	ReconnectMaxAttempts  int          `toml:"reconnect_max_attempts"`
	ReconnectBackoffBase  Duration     `toml:"reconnect_backoff_base"`
	ReconnectBackoffMax   Duration     `toml:"reconnect_backoff_max"`
	NATS                  RemoteNATSConfig `toml:"nats"`
	Manager               ManagerConfig    `toml:"manager"`
}

// RemoteNATSConfig holds NATS-specific remote settings.
type RemoteNATSConfig struct {
	URL               string   `toml:"url"`
	CredsPath         string   `toml:"creds_path"`
	SubjectPrefix     string   `toml:"subject_prefix"`
	KVBucket          string   `toml:"kv_bucket"`
	StreamEvents      string   `toml:"stream_events"`
	StreamPTY         string   `toml:"stream_pty"`
	HeartbeatInterval Duration `toml:"heartbeat_interval"`
}

// ManagerConfig holds host manager settings.
type ManagerConfig struct {
	Enabled bool   `toml:"enabled"`
	Model   string `toml:"model"`
}

// NATSConfig holds NATS server settings.
type NATSConfig struct {
	Mode         string `toml:"mode"`
	Topology     string `toml:"topology"`
	HubURL       string `toml:"hub_url"`
	Listen       string `toml:"listen"`
	AdvertiseURL string `toml:"advertise_url"`
	JetStreamDir string `toml:"jetstream_dir"`
}

// NodeConfig holds node role settings.
type NodeConfig struct {
	Role string `toml:"role"`
}

// DaemonConfig holds daemon settings.
type DaemonConfig struct {
	SocketPath string `toml:"socket_path"`
	AutoStart  bool   `toml:"autostart"`
}

// PluginsConfig holds plugin settings.
type PluginsConfig struct {
	Dir         string `toml:"dir"`
	AllowRemote bool   `toml:"allow_remote"`
}

// TelemetryConfig holds OpenTelemetry settings.
type TelemetryConfig struct {
	Enabled     bool                   `toml:"enabled"`
	ServiceName string                 `toml:"service_name"`
	Exporter    TelemetryExporterConfig `toml:"exporter"`
	Traces      TelemetryTracesConfig   `toml:"traces"`
	Metrics     TelemetryMetricsConfig  `toml:"metrics"`
	Logs        TelemetryLogsConfig     `toml:"logs"`
}

// TelemetryExporterConfig holds exporter settings.
type TelemetryExporterConfig struct {
	Endpoint string `toml:"endpoint"`
	Protocol string `toml:"protocol"`
}

// TelemetryTracesConfig holds trace settings.
type TelemetryTracesConfig struct {
	Enabled    bool    `toml:"enabled"`
	Sampler    string  `toml:"sampler"`
	SamplerArg float64 `toml:"sampler_arg"`
}

// TelemetryMetricsConfig holds metrics settings.
type TelemetryMetricsConfig struct {
	Enabled  bool     `toml:"enabled"`
	Interval Duration `toml:"interval"`
}

// TelemetryLogsConfig holds logs settings.
type TelemetryLogsConfig struct {
	Enabled bool   `toml:"enabled"`
	Level   string `toml:"level"`
}

// AgentConfig holds agent definition settings.
type AgentConfig struct {
	Name     string         `toml:"name"`
	About    string         `toml:"about"`
	Adapter  string         `toml:"adapter"`
	Location LocationConfig `toml:"location"`
}

// LocationConfig holds agent location settings.
type LocationConfig struct {
	Type     string `toml:"type"`
	Host     string `toml:"host"`
	User     string `toml:"user"`
	Port     int    `toml:"port"`
	RepoPath string `toml:"repo_path"`
}

// DefaultConfig returns the built-in default configuration.
func DefaultConfig() *Config {
	return &Config{
		General: GeneralConfig{
			LogLevel:  "info",
			LogFormat: "text",
		},
		Timeouts: TimeoutsConfig{
			Idle:  Duration{30 * time.Second},
			Stuck: Duration{5 * time.Minute},
		},
		Process: ProcessConfig{
			CaptureMode:      "all",
			StreamBufferSize: ByteSize{1024 * 1024}, // 1MB
			HookMode:         "auto",
			PollInterval:     Duration{100 * time.Millisecond},
			HookSocketDir:    "/tmp",
		},
		Git: GitConfig{
			Merge: GitMergeConfig{
				Strategy:   "squash",
				AllowDirty: false,
			},
		},
		Events: EventsConfig{
			BatchWindow:    Duration{50 * time.Millisecond},
			BatchMaxEvents: 100,
			BatchMaxBytes:  ByteSize{64 * 1024}, // 64KB
			BatchIdleFlush: Duration{10 * time.Millisecond},
			Coalesce: CoalesceConfig{
				IOStreams: true,
				Presence:  true,
				Activity:  true,
			},
		},
		Remote: RemoteConfig{
			Transport:            "nats",
			BufferSize:           ByteSize{10 * 1024 * 1024}, // 10MB
			RequestTimeout:       Duration{5 * time.Second},
			ReconnectMaxAttempts: 10,
			ReconnectBackoffBase: Duration{time.Second},
			ReconnectBackoffMax:  Duration{30 * time.Second},
			NATS: RemoteNATSConfig{
				SubjectPrefix:     "amux",
				KVBucket:          "AMUX_KV",
				StreamEvents:      "AMUX_EVENTS",
				StreamPTY:         "AMUX_PTY",
				HeartbeatInterval: Duration{5 * time.Second},
			},
			Manager: ManagerConfig{
				Enabled: true,
				Model:   "lfm2.5-thinking",
			},
		},
		NATS: NATSConfig{
			Mode:     "embedded",
			Topology: "hub",
			Listen:   "0.0.0.0:4222",
		},
		Node: NodeConfig{
			Role: "director",
		},
		Daemon: DaemonConfig{
			SocketPath: "~/.amux/amuxd.sock",
			AutoStart:  true,
		},
		Plugins: PluginsConfig{
			Dir:         "~/.config/amux/plugins",
			AllowRemote: true,
		},
		Telemetry: TelemetryConfig{
			Enabled:     false,
			ServiceName: "amux",
			Exporter: TelemetryExporterConfig{
				Endpoint: "http://localhost:4317",
				Protocol: "grpc",
			},
			Traces: TelemetryTracesConfig{
				Enabled:    true,
				Sampler:    "parentbased_traceidratio",
				SamplerArg: 0.1,
			},
			Metrics: TelemetryMetricsConfig{
				Enabled:  true,
				Interval: Duration{60 * time.Second},
			},
			Logs: TelemetryLogsConfig{
				Enabled: true,
				Level:   "info",
			},
		},
		Adapters: make(map[string]any),
		Agents:   []AgentConfig{},
	}
}

// Duration is a wrapper around time.Duration that supports TOML parsing
// with Go duration strings (e.g., "30s", "5m").
type Duration struct {
	time.Duration
}

// UnmarshalText implements encoding.TextUnmarshaler for Duration.
func (d *Duration) UnmarshalText(text []byte) error {
	parsed, err := time.ParseDuration(string(text))
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", string(text), err)
	}
	d.Duration = parsed
	return nil
}

// MarshalText implements encoding.TextMarshaler for Duration.
func (d Duration) MarshalText() ([]byte, error) {
	return []byte(d.Duration.String()), nil
}

// ByteSize is a wrapper for byte sizes that supports parsing strings like "1MB", "64KB".
// Units are binary (1KB = 1024 bytes).
type ByteSize struct {
	Bytes int64
}

// UnmarshalText implements encoding.TextUnmarshaler for ByteSize.
func (b *ByteSize) UnmarshalText(text []byte) error {
	s := strings.TrimSpace(string(text))
	if s == "" {
		return fmt.Errorf("empty byte size")
	}

	// Try parsing as a plain integer first
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		b.Bytes = n
		return nil
	}

	// Parse with unit suffix
	var multiplier int64 = 1
	var numStr string

	switch {
	case strings.HasSuffix(s, "GB"):
		multiplier = 1024 * 1024 * 1024
		numStr = strings.TrimSuffix(s, "GB")
	case strings.HasSuffix(s, "MB"):
		multiplier = 1024 * 1024
		numStr = strings.TrimSuffix(s, "MB")
	case strings.HasSuffix(s, "KB"):
		multiplier = 1024
		numStr = strings.TrimSuffix(s, "KB")
	case strings.HasSuffix(s, "B"):
		numStr = strings.TrimSuffix(s, "B")
	default:
		return fmt.Errorf("invalid byte size %q: unknown unit", s)
	}

	n, err := strconv.ParseInt(strings.TrimSpace(numStr), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid byte size %q: %w", s, err)
	}

	b.Bytes = n * multiplier
	return nil
}

// MarshalText implements encoding.TextMarshaler for ByteSize.
func (b ByteSize) MarshalText() ([]byte, error) {
	switch {
	case b.Bytes >= 1024*1024*1024 && b.Bytes%(1024*1024*1024) == 0:
		return []byte(fmt.Sprintf("%dGB", b.Bytes/(1024*1024*1024))), nil
	case b.Bytes >= 1024*1024 && b.Bytes%(1024*1024) == 0:
		return []byte(fmt.Sprintf("%dMB", b.Bytes/(1024*1024))), nil
	case b.Bytes >= 1024 && b.Bytes%1024 == 0:
		return []byte(fmt.Sprintf("%dKB", b.Bytes/1024)), nil
	default:
		return []byte(fmt.Sprintf("%dB", b.Bytes)), nil
	}
}

// Loader loads configuration from multiple sources.
type Loader struct {
	mu       sync.RWMutex
	config   *Config
	resolver *paths.Resolver
}

// NewLoader creates a new configuration loader.
func NewLoader(resolver *paths.Resolver) *Loader {
	if resolver == nil {
		resolver = paths.DefaultResolver
	}
	return &Loader{
		config:   DefaultConfig(),
		resolver: resolver,
	}
}

// Load loads configuration from all sources in order.
func (l *Loader) Load() (*Config, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Start with defaults
	l.config = DefaultConfig()

	// Load user config
	userConfigPath := l.resolver.UserConfigFile()
	if err := l.loadFile(userConfigPath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("user config: %w", err)
	}

	// Load project config
	projectConfigPath := l.resolver.ProjectConfigFile()
	if projectConfigPath != "" {
		if err := l.loadFile(projectConfigPath); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("project config: %w", err)
		}
	}

	// Apply environment variables
	if err := l.loadEnv(); err != nil {
		return nil, fmt.Errorf("environment config: %w", err)
	}

	return l.config, nil
}

// loadFile loads a TOML configuration file and merges it with current config.
func (l *Loader) loadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var fileConfig Config
	if err := toml.Unmarshal(data, &fileConfig); err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Merge the file config into current config
	l.merge(&fileConfig)
	return nil
}

// merge merges source config into the loader's config.
// Non-zero values in source override values in dest.
func (l *Loader) merge(source *Config) {
	// Simple merge - override non-zero values
	// A more sophisticated merge could be implemented if needed
	if source.General.LogLevel != "" {
		l.config.General.LogLevel = source.General.LogLevel
	}
	if source.General.LogFormat != "" {
		l.config.General.LogFormat = source.General.LogFormat
	}
	if source.Timeouts.Idle.Duration != 0 {
		l.config.Timeouts.Idle = source.Timeouts.Idle
	}
	if source.Timeouts.Stuck.Duration != 0 {
		l.config.Timeouts.Stuck = source.Timeouts.Stuck
	}
	// Add more merge rules as needed for other fields

	// Merge adapters map
	for k, v := range source.Adapters {
		l.config.Adapters[k] = v
	}

	// Merge agents list
	if len(source.Agents) > 0 {
		l.config.Agents = source.Agents
	}
}

// loadEnv loads configuration from environment variables.
// Environment variables use the prefix AMUX__ and path segments are separated by __.
func (l *Loader) loadEnv() error {
	prefix := "AMUX__"

	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, prefix) {
			continue
		}

		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0][len(prefix):]
		value := parts[1]

		// Convert key to config path
		// e.g., GENERAL__LOG_LEVEL -> general.log_level
		segments := strings.Split(key, "__")
		for i, seg := range segments {
			segments[i] = strings.ToLower(seg)
			// Handle adapter names: convert single underscore to hyphen in adapter name segment
			if i == 1 && segments[0] == "adapters" {
				segments[i] = strings.ReplaceAll(segments[i], "_", "-")
			}
		}

		// Apply the value to the config
		// This is a simplified implementation - a full implementation would
		// use reflection to set nested struct fields
		if err := l.setConfigValue(segments, value); err != nil {
			return fmt.Errorf("env %s: %w", parts[0], err)
		}
	}

	return nil
}

// setConfigValue sets a configuration value from a path and string value.
func (l *Loader) setConfigValue(path []string, value string) error {
	if len(path) == 0 {
		return fmt.Errorf("empty path")
	}

	// Handle common top-level config keys
	switch path[0] {
	case "general":
		if len(path) > 1 {
			switch path[1] {
			case "log_level":
				l.config.General.LogLevel = value
			case "log_format":
				l.config.General.LogFormat = value
			}
		}
	case "timeouts":
		if len(path) > 1 {
			switch path[1] {
			case "idle":
				if err := l.config.Timeouts.Idle.UnmarshalText([]byte(value)); err != nil {
					return err
				}
			case "stuck":
				if err := l.config.Timeouts.Stuck.UnmarshalText([]byte(value)); err != nil {
					return err
				}
			}
		}
	case "telemetry":
		if len(path) > 1 {
			switch path[1] {
			case "enabled":
				l.config.Telemetry.Enabled = value == "true"
			case "service_name":
				l.config.Telemetry.ServiceName = value
			}
		}
	// Add more cases as needed
	}

	return nil
}

// Config returns the current configuration.
func (l *Loader) Config() *Config {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.config
}

// Global functions for default loader

var defaultLoader = NewLoader(nil)

// Load loads configuration using the default loader.
func Load() (*Config, error) {
	return defaultLoader.Load()
}

// Get returns the current configuration from the default loader.
func Get() *Config {
	return defaultLoader.Config()
}

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
	Shutdown  ShutdownConfig  `toml:"shutdown"`
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

// ShutdownConfig holds graceful shutdown settings per spec §5.6.
type ShutdownConfig struct {
	DrainTimeout     Duration `toml:"drain_timeout"`
	CleanupWorktrees bool     `toml:"cleanup_worktrees"`
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
		Shutdown: ShutdownConfig{
			DrainTimeout:     Duration{30 * time.Second},
			CleanupWorktrees: false,
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

	// Expand all path fields
	l.expandPaths()

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
	// General
	if source.General.LogLevel != "" {
		l.config.General.LogLevel = source.General.LogLevel
	}
	if source.General.LogFormat != "" {
		l.config.General.LogFormat = source.General.LogFormat
	}

	// Timeouts
	if source.Timeouts.Idle.Duration != 0 {
		l.config.Timeouts.Idle = source.Timeouts.Idle
	}
	if source.Timeouts.Stuck.Duration != 0 {
		l.config.Timeouts.Stuck = source.Timeouts.Stuck
	}

	// Process
	if source.Process.CaptureMode != "" {
		l.config.Process.CaptureMode = source.Process.CaptureMode
	}
	if source.Process.StreamBufferSize.Bytes != 0 {
		l.config.Process.StreamBufferSize = source.Process.StreamBufferSize
	}
	if source.Process.HookMode != "" {
		l.config.Process.HookMode = source.Process.HookMode
	}
	if source.Process.PollInterval.Duration != 0 {
		l.config.Process.PollInterval = source.Process.PollInterval
	}
	if source.Process.HookSocketDir != "" {
		l.config.Process.HookSocketDir = source.Process.HookSocketDir
	}

	// Git
	if source.Git.Merge.Strategy != "" {
		l.config.Git.Merge.Strategy = source.Git.Merge.Strategy
	}
	if source.Git.Merge.AllowDirty {
		l.config.Git.Merge.AllowDirty = source.Git.Merge.AllowDirty
	}
	if source.Git.Merge.TargetBranch != "" {
		l.config.Git.Merge.TargetBranch = source.Git.Merge.TargetBranch
	}

	// Shutdown
	if source.Shutdown.DrainTimeout.Duration != 0 {
		l.config.Shutdown.DrainTimeout = source.Shutdown.DrainTimeout
	}
	if source.Shutdown.CleanupWorktrees {
		l.config.Shutdown.CleanupWorktrees = source.Shutdown.CleanupWorktrees
	}

	// Events
	if source.Events.BatchWindow.Duration != 0 {
		l.config.Events.BatchWindow = source.Events.BatchWindow
	}
	if source.Events.BatchMaxEvents != 0 {
		l.config.Events.BatchMaxEvents = source.Events.BatchMaxEvents
	}
	if source.Events.BatchMaxBytes.Bytes != 0 {
		l.config.Events.BatchMaxBytes = source.Events.BatchMaxBytes
	}
	if source.Events.BatchIdleFlush.Duration != 0 {
		l.config.Events.BatchIdleFlush = source.Events.BatchIdleFlush
	}
	// Events.Coalesce - bools can't be distinguished from false, so always merge
	l.config.Events.Coalesce.IOStreams = source.Events.Coalesce.IOStreams || l.config.Events.Coalesce.IOStreams
	l.config.Events.Coalesce.Presence = source.Events.Coalesce.Presence || l.config.Events.Coalesce.Presence
	l.config.Events.Coalesce.Activity = source.Events.Coalesce.Activity || l.config.Events.Coalesce.Activity
	// Events.Subscriptions
	if source.Events.Subscriptions.Enabled {
		l.config.Events.Subscriptions.Enabled = source.Events.Subscriptions.Enabled
	}
	if source.Events.Subscriptions.SocketPath != "" {
		l.config.Events.Subscriptions.SocketPath = source.Events.Subscriptions.SocketPath
	}

	// Remote
	if source.Remote.Transport != "" {
		l.config.Remote.Transport = source.Remote.Transport
	}
	if source.Remote.BufferSize.Bytes != 0 {
		l.config.Remote.BufferSize = source.Remote.BufferSize
	}
	if source.Remote.RequestTimeout.Duration != 0 {
		l.config.Remote.RequestTimeout = source.Remote.RequestTimeout
	}
	if source.Remote.ReconnectMaxAttempts != 0 {
		l.config.Remote.ReconnectMaxAttempts = source.Remote.ReconnectMaxAttempts
	}
	if source.Remote.ReconnectBackoffBase.Duration != 0 {
		l.config.Remote.ReconnectBackoffBase = source.Remote.ReconnectBackoffBase
	}
	if source.Remote.ReconnectBackoffMax.Duration != 0 {
		l.config.Remote.ReconnectBackoffMax = source.Remote.ReconnectBackoffMax
	}
	// Remote.NATS
	if source.Remote.NATS.URL != "" {
		l.config.Remote.NATS.URL = source.Remote.NATS.URL
	}
	if source.Remote.NATS.CredsPath != "" {
		l.config.Remote.NATS.CredsPath = source.Remote.NATS.CredsPath
	}
	if source.Remote.NATS.SubjectPrefix != "" {
		l.config.Remote.NATS.SubjectPrefix = source.Remote.NATS.SubjectPrefix
	}
	if source.Remote.NATS.KVBucket != "" {
		l.config.Remote.NATS.KVBucket = source.Remote.NATS.KVBucket
	}
	if source.Remote.NATS.StreamEvents != "" {
		l.config.Remote.NATS.StreamEvents = source.Remote.NATS.StreamEvents
	}
	if source.Remote.NATS.StreamPTY != "" {
		l.config.Remote.NATS.StreamPTY = source.Remote.NATS.StreamPTY
	}
	if source.Remote.NATS.HeartbeatInterval.Duration != 0 {
		l.config.Remote.NATS.HeartbeatInterval = source.Remote.NATS.HeartbeatInterval
	}
	// Remote.Manager
	if source.Remote.Manager.Enabled {
		l.config.Remote.Manager.Enabled = source.Remote.Manager.Enabled
	}
	if source.Remote.Manager.Model != "" {
		l.config.Remote.Manager.Model = source.Remote.Manager.Model
	}

	// NATS
	if source.NATS.Mode != "" {
		l.config.NATS.Mode = source.NATS.Mode
	}
	if source.NATS.Topology != "" {
		l.config.NATS.Topology = source.NATS.Topology
	}
	if source.NATS.HubURL != "" {
		l.config.NATS.HubURL = source.NATS.HubURL
	}
	if source.NATS.Listen != "" {
		l.config.NATS.Listen = source.NATS.Listen
	}
	if source.NATS.AdvertiseURL != "" {
		l.config.NATS.AdvertiseURL = source.NATS.AdvertiseURL
	}
	if source.NATS.JetStreamDir != "" {
		l.config.NATS.JetStreamDir = source.NATS.JetStreamDir
	}

	// Node
	if source.Node.Role != "" {
		l.config.Node.Role = source.Node.Role
	}

	// Daemon
	if source.Daemon.SocketPath != "" {
		l.config.Daemon.SocketPath = source.Daemon.SocketPath
	}
	if source.Daemon.AutoStart {
		l.config.Daemon.AutoStart = source.Daemon.AutoStart
	}

	// Plugins
	if source.Plugins.Dir != "" {
		l.config.Plugins.Dir = source.Plugins.Dir
	}
	if source.Plugins.AllowRemote {
		l.config.Plugins.AllowRemote = source.Plugins.AllowRemote
	}

	// Telemetry
	if source.Telemetry.Enabled {
		l.config.Telemetry.Enabled = source.Telemetry.Enabled
	}
	if source.Telemetry.ServiceName != "" {
		l.config.Telemetry.ServiceName = source.Telemetry.ServiceName
	}
	if source.Telemetry.Exporter.Endpoint != "" {
		l.config.Telemetry.Exporter.Endpoint = source.Telemetry.Exporter.Endpoint
	}
	if source.Telemetry.Exporter.Protocol != "" {
		l.config.Telemetry.Exporter.Protocol = source.Telemetry.Exporter.Protocol
	}
	if source.Telemetry.Traces.Enabled {
		l.config.Telemetry.Traces.Enabled = source.Telemetry.Traces.Enabled
	}
	if source.Telemetry.Traces.Sampler != "" {
		l.config.Telemetry.Traces.Sampler = source.Telemetry.Traces.Sampler
	}
	if source.Telemetry.Traces.SamplerArg != 0 {
		l.config.Telemetry.Traces.SamplerArg = source.Telemetry.Traces.SamplerArg
	}
	if source.Telemetry.Metrics.Enabled {
		l.config.Telemetry.Metrics.Enabled = source.Telemetry.Metrics.Enabled
	}
	if source.Telemetry.Metrics.Interval.Duration != 0 {
		l.config.Telemetry.Metrics.Interval = source.Telemetry.Metrics.Interval
	}
	if source.Telemetry.Logs.Enabled {
		l.config.Telemetry.Logs.Enabled = source.Telemetry.Logs.Enabled
	}
	if source.Telemetry.Logs.Level != "" {
		l.config.Telemetry.Logs.Level = source.Telemetry.Logs.Level
	}

	// Merge adapters map
	for k, v := range source.Adapters {
		l.config.Adapters[k] = v
	}

	// Merge agents list (replace if non-empty)
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

	// Handle top-level config keys
	switch path[0] {
	case "general":
		return l.setGeneralConfig(path[1:], value)
	case "timeouts":
		return l.setTimeoutsConfig(path[1:], value)
	case "process":
		return l.setProcessConfig(path[1:], value)
	case "git":
		return l.setGitConfig(path[1:], value)
	case "shutdown":
		return l.setShutdownConfig(path[1:], value)
	case "events":
		return l.setEventsConfig(path[1:], value)
	case "remote":
		return l.setRemoteConfig(path[1:], value)
	case "nats":
		return l.setNATSConfig(path[1:], value)
	case "node":
		return l.setNodeConfig(path[1:], value)
	case "daemon":
		return l.setDaemonConfig(path[1:], value)
	case "plugins":
		return l.setPluginsConfig(path[1:], value)
	case "telemetry":
		return l.setTelemetryConfig(path[1:], value)
	case "adapters":
		return l.setAdaptersConfig(path[1:], value)
	}

	return nil
}

func (l *Loader) setGeneralConfig(path []string, value string) error {
	if len(path) == 0 {
		return nil
	}
	switch path[0] {
	case "log_level":
		l.config.General.LogLevel = value
	case "log_format":
		l.config.General.LogFormat = value
	}
	return nil
}

func (l *Loader) setTimeoutsConfig(path []string, value string) error {
	if len(path) == 0 {
		return nil
	}
	switch path[0] {
	case "idle":
		return l.config.Timeouts.Idle.UnmarshalText([]byte(value))
	case "stuck":
		return l.config.Timeouts.Stuck.UnmarshalText([]byte(value))
	}
	return nil
}

func (l *Loader) setProcessConfig(path []string, value string) error {
	if len(path) == 0 {
		return nil
	}
	switch path[0] {
	case "capture_mode":
		l.config.Process.CaptureMode = value
	case "stream_buffer_size":
		return l.config.Process.StreamBufferSize.UnmarshalText([]byte(value))
	case "hook_mode":
		l.config.Process.HookMode = value
	case "poll_interval":
		return l.config.Process.PollInterval.UnmarshalText([]byte(value))
	case "hook_socket_dir":
		l.config.Process.HookSocketDir = value
	}
	return nil
}

func (l *Loader) setGitConfig(path []string, value string) error {
	if len(path) == 0 {
		return nil
	}
	if path[0] == "merge" && len(path) > 1 {
		switch path[1] {
		case "strategy":
			l.config.Git.Merge.Strategy = value
		case "allow_dirty":
			l.config.Git.Merge.AllowDirty = value == "true"
		case "target_branch":
			l.config.Git.Merge.TargetBranch = value
		}
	}
	return nil
}

func (l *Loader) setShutdownConfig(path []string, value string) error {
	if len(path) == 0 {
		return nil
	}
	switch path[0] {
	case "drain_timeout":
		return l.config.Shutdown.DrainTimeout.UnmarshalText([]byte(value))
	case "cleanup_worktrees":
		l.config.Shutdown.CleanupWorktrees = value == "true"
	}
	return nil
}

func (l *Loader) setEventsConfig(path []string, value string) error {
	if len(path) == 0 {
		return nil
	}
	switch path[0] {
	case "batch_window":
		return l.config.Events.BatchWindow.UnmarshalText([]byte(value))
	case "batch_max_events":
		n, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		l.config.Events.BatchMaxEvents = n
	case "batch_max_bytes":
		return l.config.Events.BatchMaxBytes.UnmarshalText([]byte(value))
	case "batch_idle_flush":
		return l.config.Events.BatchIdleFlush.UnmarshalText([]byte(value))
	case "coalesce":
		if len(path) > 1 {
			switch path[1] {
			case "io_streams":
				l.config.Events.Coalesce.IOStreams = value == "true"
			case "presence":
				l.config.Events.Coalesce.Presence = value == "true"
			case "activity":
				l.config.Events.Coalesce.Activity = value == "true"
			}
		}
	case "subscriptions":
		if len(path) > 1 {
			switch path[1] {
			case "enabled":
				l.config.Events.Subscriptions.Enabled = value == "true"
			case "socket_path":
				l.config.Events.Subscriptions.SocketPath = value
			}
		}
	}
	return nil
}

func (l *Loader) setRemoteConfig(path []string, value string) error {
	if len(path) == 0 {
		return nil
	}
	switch path[0] {
	case "transport":
		l.config.Remote.Transport = value
	case "buffer_size":
		return l.config.Remote.BufferSize.UnmarshalText([]byte(value))
	case "request_timeout":
		return l.config.Remote.RequestTimeout.UnmarshalText([]byte(value))
	case "reconnect_max_attempts":
		n, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		l.config.Remote.ReconnectMaxAttempts = n
	case "reconnect_backoff_base":
		return l.config.Remote.ReconnectBackoffBase.UnmarshalText([]byte(value))
	case "reconnect_backoff_max":
		return l.config.Remote.ReconnectBackoffMax.UnmarshalText([]byte(value))
	case "nats":
		if len(path) > 1 {
			switch path[1] {
			case "url":
				l.config.Remote.NATS.URL = value
			case "creds_path":
				l.config.Remote.NATS.CredsPath = value
			case "subject_prefix":
				l.config.Remote.NATS.SubjectPrefix = value
			case "kv_bucket":
				l.config.Remote.NATS.KVBucket = value
			case "stream_events":
				l.config.Remote.NATS.StreamEvents = value
			case "stream_pty":
				l.config.Remote.NATS.StreamPTY = value
			case "heartbeat_interval":
				return l.config.Remote.NATS.HeartbeatInterval.UnmarshalText([]byte(value))
			}
		}
	case "manager":
		if len(path) > 1 {
			switch path[1] {
			case "enabled":
				l.config.Remote.Manager.Enabled = value == "true"
			case "model":
				l.config.Remote.Manager.Model = value
			}
		}
	}
	return nil
}

func (l *Loader) setNATSConfig(path []string, value string) error {
	if len(path) == 0 {
		return nil
	}
	switch path[0] {
	case "mode":
		l.config.NATS.Mode = value
	case "topology":
		l.config.NATS.Topology = value
	case "hub_url":
		l.config.NATS.HubURL = value
	case "listen":
		l.config.NATS.Listen = value
	case "advertise_url":
		l.config.NATS.AdvertiseURL = value
	case "jetstream_dir":
		l.config.NATS.JetStreamDir = value
	}
	return nil
}

func (l *Loader) setNodeConfig(path []string, value string) error {
	if len(path) == 0 {
		return nil
	}
	if path[0] == "role" {
		l.config.Node.Role = value
	}
	return nil
}

func (l *Loader) setDaemonConfig(path []string, value string) error {
	if len(path) == 0 {
		return nil
	}
	switch path[0] {
	case "socket_path":
		l.config.Daemon.SocketPath = value
	case "autostart":
		l.config.Daemon.AutoStart = value == "true"
	}
	return nil
}

func (l *Loader) setPluginsConfig(path []string, value string) error {
	if len(path) == 0 {
		return nil
	}
	switch path[0] {
	case "dir":
		l.config.Plugins.Dir = value
	case "allow_remote":
		l.config.Plugins.AllowRemote = value == "true"
	}
	return nil
}

func (l *Loader) setTelemetryConfig(path []string, value string) error {
	if len(path) == 0 {
		return nil
	}
	switch path[0] {
	case "enabled":
		l.config.Telemetry.Enabled = value == "true"
	case "service_name":
		l.config.Telemetry.ServiceName = value
	case "exporter":
		if len(path) > 1 {
			switch path[1] {
			case "endpoint":
				l.config.Telemetry.Exporter.Endpoint = value
			case "protocol":
				l.config.Telemetry.Exporter.Protocol = value
			}
		}
	case "traces":
		if len(path) > 1 {
			switch path[1] {
			case "enabled":
				l.config.Telemetry.Traces.Enabled = value == "true"
			case "sampler":
				l.config.Telemetry.Traces.Sampler = value
			case "sampler_arg":
				f, err := strconv.ParseFloat(value, 64)
				if err != nil {
					return err
				}
				l.config.Telemetry.Traces.SamplerArg = f
			}
		}
	case "metrics":
		if len(path) > 1 {
			switch path[1] {
			case "enabled":
				l.config.Telemetry.Metrics.Enabled = value == "true"
			case "interval":
				return l.config.Telemetry.Metrics.Interval.UnmarshalText([]byte(value))
			}
		}
	case "logs":
		if len(path) > 1 {
			switch path[1] {
			case "enabled":
				l.config.Telemetry.Logs.Enabled = value == "true"
			case "level":
				l.config.Telemetry.Logs.Level = value
			}
		}
	}
	return nil
}

func (l *Loader) setAdaptersConfig(path []string, value string) error {
	if len(path) < 2 {
		return nil
	}
	adapterName := path[0]
	if l.config.Adapters == nil {
		l.config.Adapters = make(map[string]any)
	}

	// Get or create adapter config map
	adapterConfig, ok := l.config.Adapters[adapterName].(map[string]any)
	if !ok {
		adapterConfig = make(map[string]any)
		l.config.Adapters[adapterName] = adapterConfig
	}

	// Set nested value
	current := adapterConfig
	for i := 1; i < len(path)-1; i++ {
		key := path[i]
		next, ok := current[key].(map[string]any)
		if !ok {
			next = make(map[string]any)
			current[key] = next
		}
		current = next
	}
	current[path[len(path)-1]] = value
	return nil
}

// Config returns the current configuration.
func (l *Loader) Config() *Config {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.config
}

// expandPaths expands all path fields that start with ~/ to the user's home directory.
func (l *Loader) expandPaths() {
	homeDir := l.resolver.HomeDir()
	if homeDir == "" {
		return
	}

	expand := func(p string) string {
		if strings.HasPrefix(p, "~/") {
			return homeDir + p[1:]
		}
		return p
	}

	// Expand path fields
	l.config.Process.HookSocketDir = expand(l.config.Process.HookSocketDir)
	l.config.Remote.NATS.CredsPath = expand(l.config.Remote.NATS.CredsPath)
	l.config.NATS.JetStreamDir = expand(l.config.NATS.JetStreamDir)
	l.config.Daemon.SocketPath = expand(l.config.Daemon.SocketPath)
	l.config.Plugins.Dir = expand(l.config.Plugins.Dir)
	l.config.Events.Subscriptions.SocketPath = expand(l.config.Events.Subscriptions.SocketPath)

	// Expand agent location paths
	for i := range l.config.Agents {
		l.config.Agents[i].Location.RepoPath = expand(l.config.Agents[i].Location.RepoPath)
	}
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

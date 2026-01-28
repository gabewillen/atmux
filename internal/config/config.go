// Package config provides configuration management for amux.
// Configuration is loaded from multiple sources in a defined hierarchy:
// built-in defaults < adapter defaults < user config < project config < environment variables.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the complete amux configuration.
type Config struct {
	General  GeneralConfig  `toml:"general"`
	Timeouts TimeoutsConfig `toml:"timeouts"`
	Process  ProcessConfig  `toml:"process"`
	Git      GitConfig      `toml:"git"`
	Events   EventsConfig   `toml:"events"`
	Remote   RemoteConfig   `toml:"remote"`
	NATS     NATSConfig     `toml:"nats"`
	Node     NodeConfig     `toml:"node"`
	Daemon   DaemonConfig   `toml:"daemon"`
	Plugins  PluginsConfig  `toml:"plugins"`
	Telemetry TelemetryConfig `toml:"telemetry"`
	Agents   []AgentConfig  `toml:"agents"`
	Adapters map[string]interface{} `toml:"adapters"` // Opaque adapter configs
}

// GeneralConfig holds general application settings.
type GeneralConfig struct {
	LogLevel  string `toml:"log_level"`
	LogFormat string `toml:"log_format"`
}

// TimeoutsConfig holds timeout settings.
type TimeoutsConfig struct {
	Idle  string `toml:"idle"`  // Duration string
	Stuck string `toml:"stuck"` // Duration string
}

// ProcessConfig holds process tracking configuration.
type ProcessConfig struct {
	CaptureMode     string `toml:"capture_mode"`
	StreamBufferSize string `toml:"stream_buffer_size"`
	HookMode        string `toml:"hook_mode"`
	PollInterval    string `toml:"poll_interval"`
	HookSocketDir   string `toml:"hook_socket_dir"`
}

// GitConfig holds git-related configuration.
type GitConfig struct {
	Merge GitMergeConfig `toml:"merge"`
}

// GitMergeConfig holds git merge strategy configuration.
type GitMergeConfig struct {
	Strategy   string `toml:"strategy"`
	AllowDirty bool   `toml:"allow_dirty"`
}

// EventsConfig holds event batching and coalescing configuration.
type EventsConfig struct {
	BatchWindow    string `toml:"batch_window"`
	BatchMaxEvents int    `toml:"batch_max_events"`
	BatchMaxBytes  string `toml:"batch_max_bytes"`
	BatchIdleFlush string `toml:"batch_idle_flush"`
	Coalesce       EventsCoalesceConfig `toml:"coalesce"`
}

// EventsCoalesceConfig holds event coalescing settings.
type EventsCoalesceConfig struct {
	IOStreams bool `toml:"io_streams"`
	Presence  bool `toml:"presence"`
	Activity  bool `toml:"activity"`
}

// RemoteConfig holds remote agent configuration.
type RemoteConfig struct {
	Transport            string      `toml:"transport"`
	BufferSize          string       `toml:"buffer_size"`
	RequestTimeout      string       `toml:"request_timeout"`
	ReconnectMaxAttempts int        `toml:"reconnect_max_attempts"`
	ReconnectBackoffBase string     `toml:"reconnect_backoff_base"`
	ReconnectBackoffMax  string     `toml:"reconnect_backoff_max"`
	NATS                 RemoteNATSConfig `toml:"nats"`
	Manager              RemoteManagerConfig `toml:"manager"`
}

// RemoteNATSConfig holds NATS-specific remote configuration.
type RemoteNATSConfig struct {
	URL              string `toml:"url"`
	CredsPath        string `toml:"creds_path"`
	SubjectPrefix    string `toml:"subject_prefix"`
	KVBucket         string `toml:"kv_bucket"`
	StreamEvents     string `toml:"stream_events"`
	StreamPTY        string `toml:"stream_pty"`
	HeartbeatInterval string `toml:"heartbeat_interval"`
}

// RemoteManagerConfig holds remote manager configuration.
type RemoteManagerConfig struct {
	Enabled bool   `toml:"enabled"`
	Model   string `toml:"model"`
}

// NATSConfig holds NATS server configuration.
type NATSConfig struct {
	Mode         string `toml:"mode"`
	Topology     string `toml:"topology"`
	HubURL       string `toml:"hub_url"`
	Listen       string `toml:"listen"`
	AdvertiseURL string `toml:"advertise_url"`
	JetStreamDir string `toml:"jetstream_dir"`
}

// NodeConfig holds node role configuration.
type NodeConfig struct {
	Role string `toml:"role"`
}

// DaemonConfig holds daemon configuration.
type DaemonConfig struct {
	SocketPath string `toml:"socket_path"`
	Autostart  bool   `toml:"autostart"`
}

// PluginsConfig holds plugin configuration.
type PluginsConfig struct {
	Dir        string `toml:"dir"`
	AllowRemote bool `toml:"allow_remote"`
}

// TelemetryConfig holds OpenTelemetry configuration.
type TelemetryConfig struct {
	Enabled    bool                `toml:"enabled"`
	ServiceName string             `toml:"service_name"`
	Exporter   TelemetryExporterConfig `toml:"exporter"`
	Traces     TelemetryTracesConfig   `toml:"traces"`
	Metrics    TelemetryMetricsConfig  `toml:"metrics"`
	Logs       TelemetryLogsConfig     `toml:"logs"`
}

// TelemetryExporterConfig holds OTel exporter configuration.
type TelemetryExporterConfig struct {
	Endpoint string `toml:"endpoint"`
	Protocol string `toml:"protocol"`
}

// TelemetryTracesConfig holds OTel traces configuration.
type TelemetryTracesConfig struct {
	Enabled    bool    `toml:"enabled"`
	Sampler    string  `toml:"sampler"`
	SamplerArg float64 `toml:"sampler_arg"`
}

// TelemetryMetricsConfig holds OTel metrics configuration.
type TelemetryMetricsConfig struct {
	Enabled  bool   `toml:"enabled"`
	Interval string `toml:"interval"`
}

// TelemetryLogsConfig holds OTel logs configuration.
type TelemetryLogsConfig struct {
	Enabled bool   `toml:"enabled"`
	Level   string `toml:"level"`
}

// AgentConfig represents a single agent definition.
type AgentConfig struct {
	Name     string         `toml:"name"`
	About    string         `toml:"about"`
	Adapter  string         `toml:"adapter"`
	Location AgentLocationConfig `toml:"location"`
}

// AgentLocationConfig holds agent location information.
type AgentLocationConfig struct {
	Type     string `toml:"type"`
	Host     string `toml:"host"`
	User     string `toml:"user"`
	Port     int    `toml:"port"`
	RepoPath string `toml:"repo_path"`
}

// Loader loads configuration from multiple sources.
type Loader struct {
	configDir string
	homeDir   string
	repoRoot  string
}

// NewLoader creates a new configuration loader.
func NewLoader(configDir, homeDir, repoRoot string) *Loader {
	return &Loader{
		configDir: configDir,
		homeDir:   homeDir,
		repoRoot:  repoRoot,
	}
}

// Load loads configuration from all sources in the correct order.
func (l *Loader) Load() (*Config, error) {
	cfg := l.defaultConfig()

	// Load adapter defaults (placeholder for Phase 8)
	// TODO: Load adapter defaults when adapter system is implemented

	// Load user config
	userConfigPath := filepath.Join(l.expandHome(l.configDir), "config.toml")
	if err := l.loadFile(userConfigPath, cfg); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load user config: %w", err)
	}

	// Load project config
	if l.repoRoot != "" {
		projectConfigPath := filepath.Join(l.repoRoot, ".amux", "config.toml")
		if err := l.loadFile(projectConfigPath, cfg); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load project config: %w", err)
		}
	}

	// Apply environment variable overrides
	if err := l.applyEnvOverrides(cfg); err != nil {
		return nil, fmt.Errorf("failed to apply environment overrides: %w", err)
	}

	return cfg, nil
}

// defaultConfig returns the built-in default configuration.
func (l *Loader) defaultConfig() *Config {
	return &Config{
		General: GeneralConfig{
			LogLevel:  "info",
			LogFormat: "text",
		},
		Timeouts: TimeoutsConfig{
			Idle:  "30s",
			Stuck: "5m",
		},
		Process: ProcessConfig{
			CaptureMode:     "all",
			StreamBufferSize: "1MB",
			HookMode:        "auto",
			PollInterval:     "100ms",
			HookSocketDir:   "/tmp",
		},
		Git: GitConfig{
			Merge: GitMergeConfig{
				Strategy:   "squash",
				AllowDirty: false,
			},
		},
		Events: EventsConfig{
			BatchWindow:    "50ms",
			BatchMaxEvents: 100,
			BatchMaxBytes:  "64KB",
			BatchIdleFlush: "10ms",
			Coalesce: EventsCoalesceConfig{
				IOStreams: true,
				Presence:  true,
				Activity:  true,
			},
		},
		Remote: RemoteConfig{
			Transport:            "nats",
			BufferSize:          "10MB",
			RequestTimeout:      "5s",
			ReconnectMaxAttempts: 10,
			ReconnectBackoffBase: "1s",
			ReconnectBackoffMax:  "30s",
			NATS: RemoteNATSConfig{
				URL:              "nats://amux-host:4222",
				CredsPath:        "~/.config/amux/nats.creds",
				SubjectPrefix:    "amux",
				KVBucket:         "AMUX_KV",
				StreamEvents:     "AMUX_EVENTS",
				StreamPTY:        "AMUX_PTY",
				HeartbeatInterval: "5s",
			},
			Manager: RemoteManagerConfig{
				Enabled: true,
				Model:   "lfm2.5-thinking",
			},
		},
		NATS: NATSConfig{
			Mode:         "embedded",
			Topology:     "hub",
			Listen:       "0.0.0.0:4222",
			AdvertiseURL: "nats://amux-host:4222",
			JetStreamDir: "~/.amux/nats",
		},
		Node: NodeConfig{
			Role: "director",
		},
		Daemon: DaemonConfig{
			SocketPath: "~/.amux/amuxd.sock",
			Autostart:  true,
		},
		Plugins: PluginsConfig{
			Dir:        "~/.config/amux/plugins",
			AllowRemote: true,
		},
		Telemetry: TelemetryConfig{
			Enabled:     true,
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
				Interval: "60s",
			},
			Logs: TelemetryLogsConfig{
				Enabled: true,
				Level:   "info",
			},
		},
		Adapters: make(map[string]interface{}),
	}
}

// loadFile loads configuration from a TOML file and merges it into cfg.
func (l *Loader) loadFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var fileCfg Config
	if err := toml.Unmarshal(data, &fileCfg); err != nil {
		return fmt.Errorf("failed to parse TOML: %w", err)
	}

	// Merge fileCfg into cfg (simple merge for now)
	// TODO: Implement proper deep merge
	if fileCfg.General.LogLevel != "" {
		cfg.General.LogLevel = fileCfg.General.LogLevel
	}
	if fileCfg.General.LogFormat != "" {
		cfg.General.LogFormat = fileCfg.General.LogFormat
	}
	// TODO: Merge all other fields properly

	return nil
}

// applyEnvOverrides applies environment variable overrides to cfg.
func (l *Loader) applyEnvOverrides(cfg *Config) error {
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, "AMUX__") {
			continue
		}

		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		// Remove AMUX__ prefix and split by __
		key = strings.TrimPrefix(key, "AMUX__")
		segments := strings.Split(key, "__")

		// TODO: Implement full environment variable mapping per spec §4.2.8.3
		_ = segments
		_ = value
	}

	return nil
}

// expandHome expands ~ to the home directory.
func (l *Loader) expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(l.homeDir, path[2:])
	}
	return path
}

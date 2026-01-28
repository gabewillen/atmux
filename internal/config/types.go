package config

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
	Adapters map[string]map[string]any `toml:"adapters"` // Opaque adapter configs
	Agents   []AgentConfig  `toml:"agents"`
	Telemetry TelemetryConfig `toml:"telemetry"`
}

type GeneralConfig struct {
	LogLevel  string `toml:"log_level"`
	LogFormat string `toml:"log_format"`
}

type TimeoutsConfig struct {
	Idle  string `toml:"idle"` // Duration string
	Stuck string `toml:"stuck"` // Duration string
}

type ProcessConfig struct {
	CaptureMode      string `toml:"capture_mode"`
	StreamBufferSize string `toml:"stream_buffer_size"`
	HookMode         string `toml:"hook_mode"`
	PollInterval     string `toml:"poll_interval"`
	HookSocketDir    string `toml:"hook_socket_dir"`
}

type GitConfig struct {
	Merge MergeConfig `toml:"merge"`
}

type MergeConfig struct {
	Strategy    string `toml:"strategy"`
	AllowDirty  bool   `toml:"allow_dirty"`
	TargetBranch string `toml:"target_branch,omitempty"`
}

type EventsConfig struct {
	BatchWindow    string         `toml:"batch_window"`
	BatchMaxEvents int            `toml:"batch_max_events"`
	BatchMaxBytes  string         `toml:"batch_max_bytes"`
	BatchIdleFlush string         `toml:"batch_idle_flush"`
	Coalesce       CoalesceConfig `toml:"coalesce"`
}

type CoalesceConfig struct {
	IOStreams bool `toml:"io_streams"`
	Presence  bool `toml:"presence"`
	Activity  bool `toml:"activity"`
}

type RemoteConfig struct {
	Transport            string       `toml:"transport"`
	BufferSize           string       `toml:"buffer_size"`
	RequestTimeout       string       `toml:"request_timeout"`
	ReconnectMaxAttempts int          `toml:"reconnect_max_attempts"`
	ReconnectBackoffBase string       `toml:"reconnect_backoff_base"`
	ReconnectBackoffMax  string       `toml:"reconnect_backoff_max"`
	NATS                 RemoteNATSConfig `toml:"nats"`
	Manager              RemoteManagerConfig `toml:"manager"`
}

type RemoteNATSConfig struct {
	URL           string `toml:"url"`
	CredsPath     string `toml:"creds_path"`
	SubjectPrefix string `toml:"subject_prefix"`
	KVBucket      string `toml:"kv_bucket"`
	StreamEvents  string `toml:"stream_events"`
	StreamPTY     string `toml:"stream_pty"`
	HeartbeatInterval string `toml:"heartbeat_interval"`
}

type RemoteManagerConfig struct {
	Enabled bool   `toml:"enabled"`
	Model   string `toml:"model"`
}

type NATSConfig struct {
	Mode         string `toml:"mode"`
	Topology     string `toml:"topology"`
	HubURL       string `toml:"hub_url"`
	Listen       string `toml:"listen"`
	AdvertiseURL string `toml:"advertise_url"`
	JetStreamDir string `toml:"jetstream_dir"`
}

type NodeConfig struct {
	Role string `toml:"role"`
}

type DaemonConfig struct {
	SocketPath string `toml:"socket_path"`
	Autostart  bool   `toml:"autostart"`
}

type PluginsConfig struct {
	Dir         string `toml:"dir"`
	AllowRemote bool   `toml:"allow_remote"`
}

type AgentConfig struct {
	Name     string         `toml:"name"`
	About    string         `toml:"about"`
	Adapter  string         `toml:"adapter"`
	Location LocationConfig `toml:"location"`
}

type LocationConfig struct {
	Type     string `toml:"type"`
	Host     string `toml:"host,omitempty"`
	User     string `toml:"user,omitempty"`
	Port     int    `toml:"port,omitempty"`
	RepoPath string `toml:"repo_path,omitempty"`
}

type TelemetryConfig struct {
	Enabled     bool               `toml:"enabled"`
	ServiceName string             `toml:"service_name"`
	Exporter    TelemetryExporter  `toml:"exporter"`
	Traces      TelemetryTraces    `toml:"traces"`
	Metrics     TelemetryMetrics   `toml:"metrics"`
	Logs        TelemetryLogs      `toml:"logs"`
}

type TelemetryExporter struct {
	Endpoint string `toml:"endpoint"`
	Protocol string `toml:"protocol"`
}

type TelemetryTraces struct {
	Enabled    bool    `toml:"enabled"`
	Sampler    string  `toml:"sampler"`
	SamplerArg float64 `toml:"sampler_arg"`
}

type TelemetryMetrics struct {
	Enabled  bool   `toml:"enabled"`
	Interval string `toml:"interval"`
}

type TelemetryLogs struct {
	Enabled bool   `toml:"enabled"`
	Level   string `toml:"level"`
}

// DefaultConfig returns the built-in defaults.
func DefaultConfig() Config {
	return Config{
		General: GeneralConfig{
			LogLevel:  "info",
			LogFormat: "text",
		},
		Timeouts: TimeoutsConfig{
			Idle:  "30s",
			Stuck: "5m",
		},
		Process: ProcessConfig{
			CaptureMode:      "all",
			StreamBufferSize: "1MB",
			HookMode:         "auto",
			PollInterval:     "100ms",
			HookSocketDir:    "/tmp",
		},
		Git: GitConfig{
			Merge: MergeConfig{
				Strategy:   "squash",
				AllowDirty: false,
			},
		},
		Events: EventsConfig{
			BatchWindow:    "50ms",
			BatchMaxEvents: 100,
			BatchMaxBytes:  "64KB",
			BatchIdleFlush: "10ms",
			Coalesce: CoalesceConfig{
				IOStreams: true,
				Presence:  true,
				Activity:  true,
			},
		},
		Remote: RemoteConfig{
			Transport:            "nats",
			BufferSize:           "10MB",
			RequestTimeout:       "5s",
			ReconnectMaxAttempts: 10,
			ReconnectBackoffBase: "1s",
			ReconnectBackoffMax:  "30s",
			NATS: RemoteNATSConfig{
				SubjectPrefix:     "amux",
				KVBucket:          "AMUX_KV",
				StreamEvents:      "AMUX_EVENTS",
				StreamPTY:         "AMUX_PTY",
				HeartbeatInterval: "5s",
			},
		},
		NATS: NATSConfig{
			Mode:         "embedded",
			Topology:     "hub",
			Listen:       "0.0.0.0:4222",
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
			Dir:         "~/.config/amux/plugins",
			AllowRemote: true,
		},
		Telemetry: TelemetryConfig{
			Enabled:     true,
			ServiceName: "amux",
			Traces: TelemetryTraces{
				Enabled:    true,
				Sampler:    "parentbased_traceidratio",
				SamplerArg: 0.1,
			},
			Metrics: TelemetryMetrics{
				Enabled:  true,
				Interval: "60s",
			},
			Logs: TelemetryLogs{
				Enabled: true,
				Level:   "info",
			},
		},
	}
}

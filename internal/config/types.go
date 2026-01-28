package config

// Config represents the root configuration.
type Config struct {
	General   GeneralConfig            `toml:"general"`
	Timeouts  TimeoutsConfig           `toml:"timeouts"`
	Process   ProcessConfig            `toml:"process"`
	Git       GitConfig                `toml:"git"`
	Events    EventsConfig             `toml:"events"`
	Remote    RemoteConfig             `toml:"remote"`
	NATS      NATSConfig               `toml:"nats"`
	Node      NodeConfig               `toml:"node"`
	Daemon    DaemonConfig             `toml:"daemon"`
	Plugins   PluginsConfig            `toml:"plugins"`
	Telemetry TelemetryConfig          `toml:"telemetry"`
	Adapters  map[string]AdapterConfig `toml:"adapters"`
	Agents    []AgentDef               `toml:"agents"`
}

type GeneralConfig struct {
	LogLevel  string `toml:"log_level"`
	LogFormat string `toml:"log_format"`
}

type TimeoutsConfig struct {
	Idle  string `toml:"idle"`  // Duration string
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
	Strategy   string `toml:"strategy"`
	AllowDirty bool   `toml:"allow_dirty"`
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
	Transport            string           `toml:"transport"`
	BufferSize           string           `toml:"buffer_size"`
	RequestTimeout       string           `toml:"request_timeout"`
	ReconnectMaxAttempts int              `toml:"reconnect_max_attempts"`
	ReconnectBackoffBase string           `toml:"reconnect_backoff_base"`
	ReconnectBackoffMax  string           `toml:"reconnect_backoff_max"`
	NATS                 RemoteNATSConfig `toml:"nats"`
	Manager              ManagerConfig    `toml:"manager"`
}

type RemoteNATSConfig struct {
	URL               string `toml:"url"`
	CredsPath         string `toml:"creds_path"`
	SubjectPrefix     string `toml:"subject_prefix"`
	KVBucket          string `toml:"kv_bucket"`
	StreamEvents      string `toml:"stream_events"`
	StreamPTY         string `toml:"stream_pty"`
	HeartbeatInterval string `toml:"heartbeat_interval"`
}

type ManagerConfig struct {
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

type TelemetryConfig struct {
	Enabled     bool              `toml:"enabled"`
	ServiceName string            `toml:"service_name"`
	Exporter    TelemetryExporter `toml:"exporter"`
	Traces      TelemetrySignal   `toml:"traces"`
	Metrics     TelemetrySignal   `toml:"metrics"`
	Logs        TelemetrySignal   `toml:"logs"`
}

type TelemetryExporter struct {
	Endpoint string `toml:"endpoint"`
	Protocol string `toml:"protocol"`
}

type TelemetrySignal struct {
	Enabled    bool    `toml:"enabled"`
	Sampler    string  `toml:"sampler,omitempty"`
	SamplerArg float64 `toml:"sampler_arg,omitempty"`
	Interval   string  `toml:"interval,omitempty"`
	Level      string  `toml:"level,omitempty"`
}

// AdapterConfig holds opaque configuration for adapters.
// The structure is map[string]any to allow arbitrary keys.
type AdapterConfig map[string]any

type AgentDef struct {
	Name     string        `toml:"name"`
	About    string        `toml:"about"`
	Adapter  string        `toml:"adapter"`
	Location AgentLocation `toml:"location"`
}

type AgentLocation struct {
	Type     string `toml:"type"`
	Host     string `toml:"host,omitempty"`
	RepoPath string `toml:"repo_path,omitempty"`
}

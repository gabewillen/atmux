# package config

`import "github.com/agentflare-ai/amux/internal/config"`

- `func applyEnvOverrides(cfg *Config) error`
- `func envKeyToTomlPath(key string) (string, error)` — envKeyToTomlPath converts AMUX__FOO__BAR to foo.bar
- `func loadFile(path string, cfg *Config) error`
- `type AgentConfig`
- `type CoalesceConfig`
- `type ConfigActor`
- `type Config`
- `type DaemonConfig`
- `type EventsConfig`
- `type GeneralConfig`
- `type GitConfig`
- `type LocationConfig`
- `type MergeConfig`
- `type NATSConfig`
- `type NodeConfig`
- `type PluginsConfig`
- `type ProcessConfig`
- `type RemoteConfig`
- `type RemoteManagerConfig`
- `type RemoteNATSConfig`
- `type TelemetryConfig`
- `type TelemetryExporter`
- `type TelemetryLogs`
- `type TelemetryMetrics`
- `type TelemetryTraces`
- `type TimeoutsConfig`

### Functions

#### applyEnvOverrides

```go
func applyEnvOverrides(cfg *Config) error
```

#### envKeyToTomlPath

```go
func envKeyToTomlPath(key string) (string, error)
```

envKeyToTomlPath converts AMUX__FOO__BAR to foo.bar

#### loadFile

```go
func loadFile(path string, cfg *Config) error
```


## type AgentConfig

```go
type AgentConfig struct {
	Name     string         `toml:"name"`
	About    string         `toml:"about"`
	Adapter  string         `toml:"adapter"`
	Location LocationConfig `toml:"location"`
}
```

## type CoalesceConfig

```go
type CoalesceConfig struct {
	IOStreams bool `toml:"io_streams"`
	Presence  bool `toml:"presence"`
	Activity  bool `toml:"activity"`
}
```

## type Config

```go
type Config struct {
	General   GeneralConfig             `toml:"general"`
	Timeouts  TimeoutsConfig            `toml:"timeouts"`
	Process   ProcessConfig             `toml:"process"`
	Git       GitConfig                 `toml:"git"`
	Events    EventsConfig              `toml:"events"`
	Remote    RemoteConfig              `toml:"remote"`
	NATS      NATSConfig                `toml:"nats"`
	Node      NodeConfig                `toml:"node"`
	Daemon    DaemonConfig              `toml:"daemon"`
	Plugins   PluginsConfig             `toml:"plugins"`
	Adapters  map[string]map[string]any `toml:"adapters"` // Opaque adapter configs
	Agents    []AgentConfig             `toml:"agents"`
	Telemetry TelemetryConfig           `toml:"telemetry"`
}
```

### Functions returning Config

#### DefaultConfig

```go
func DefaultConfig() Config
```

DefaultConfig returns the built-in defaults.

#### Load

```go
func Load(repoRoot string) (*Config, error)
```

Load loads the configuration from all sources.


## type ConfigActor

```go
type ConfigActor struct {
	mu          sync.RWMutex
	current     *Config
	subscribers []chan<- Config
}
```

### Functions returning ConfigActor

#### NewActor

```go
func NewActor(initial *Config) *ConfigActor
```


### Methods

#### ConfigActor.Start

```go
func () Start(ctx context.Context) error
```

#### ConfigActor.Subscribe

```go
func () Subscribe() <-chan Config
```

#### ConfigActor.Update

```go
func () Update(cfg *Config)
```


## type DaemonConfig

```go
type DaemonConfig struct {
	SocketPath string `toml:"socket_path"`
	Autostart  bool   `toml:"autostart"`
}
```

## type EventsConfig

```go
type EventsConfig struct {
	BatchWindow    string         `toml:"batch_window"`
	BatchMaxEvents int            `toml:"batch_max_events"`
	BatchMaxBytes  string         `toml:"batch_max_bytes"`
	BatchIdleFlush string         `toml:"batch_idle_flush"`
	Coalesce       CoalesceConfig `toml:"coalesce"`
}
```

## type GeneralConfig

```go
type GeneralConfig struct {
	LogLevel  string `toml:"log_level"`
	LogFormat string `toml:"log_format"`
}
```

## type GitConfig

```go
type GitConfig struct {
	Merge MergeConfig `toml:"merge"`
}
```

## type LocationConfig

```go
type LocationConfig struct {
	Type     string `toml:"type"`
	Host     string `toml:"host,omitempty"`
	User     string `toml:"user,omitempty"`
	Port     int    `toml:"port,omitempty"`
	RepoPath string `toml:"repo_path,omitempty"`
}
```

## type MergeConfig

```go
type MergeConfig struct {
	Strategy     string `toml:"strategy"`
	AllowDirty   bool   `toml:"allow_dirty"`
	TargetBranch string `toml:"target_branch,omitempty"`
}
```

## type NATSConfig

```go
type NATSConfig struct {
	Mode         string `toml:"mode"`
	Topology     string `toml:"topology"`
	HubURL       string `toml:"hub_url"`
	Listen       string `toml:"listen"`
	AdvertiseURL string `toml:"advertise_url"`
	JetStreamDir string `toml:"jetstream_dir"`
}
```

## type NodeConfig

```go
type NodeConfig struct {
	Role string `toml:"role"`
}
```

## type PluginsConfig

```go
type PluginsConfig struct {
	Dir         string `toml:"dir"`
	AllowRemote bool   `toml:"allow_remote"`
}
```

## type ProcessConfig

```go
type ProcessConfig struct {
	CaptureMode      string `toml:"capture_mode"`
	StreamBufferSize string `toml:"stream_buffer_size"`
	HookMode         string `toml:"hook_mode"`
	PollInterval     string `toml:"poll_interval"`
	HookSocketDir    string `toml:"hook_socket_dir"`
}
```

## type RemoteConfig

```go
type RemoteConfig struct {
	Transport            string              `toml:"transport"`
	BufferSize           string              `toml:"buffer_size"`
	RequestTimeout       string              `toml:"request_timeout"`
	ReconnectMaxAttempts int                 `toml:"reconnect_max_attempts"`
	ReconnectBackoffBase string              `toml:"reconnect_backoff_base"`
	ReconnectBackoffMax  string              `toml:"reconnect_backoff_max"`
	NATS                 RemoteNATSConfig    `toml:"nats"`
	Manager              RemoteManagerConfig `toml:"manager"`
}
```

## type RemoteManagerConfig

```go
type RemoteManagerConfig struct {
	Enabled bool   `toml:"enabled"`
	Model   string `toml:"model"`
}
```

## type RemoteNATSConfig

```go
type RemoteNATSConfig struct {
	URL               string `toml:"url"`
	CredsPath         string `toml:"creds_path"`
	SubjectPrefix     string `toml:"subject_prefix"`
	KVBucket          string `toml:"kv_bucket"`
	StreamEvents      string `toml:"stream_events"`
	StreamPTY         string `toml:"stream_pty"`
	HeartbeatInterval string `toml:"heartbeat_interval"`
}
```

## type TelemetryConfig

```go
type TelemetryConfig struct {
	Enabled     bool              `toml:"enabled"`
	ServiceName string            `toml:"service_name"`
	Exporter    TelemetryExporter `toml:"exporter"`
	Traces      TelemetryTraces   `toml:"traces"`
	Metrics     TelemetryMetrics  `toml:"metrics"`
	Logs        TelemetryLogs     `toml:"logs"`
}
```

## type TelemetryExporter

```go
type TelemetryExporter struct {
	Endpoint string `toml:"endpoint"`
	Protocol string `toml:"protocol"`
}
```

## type TelemetryLogs

```go
type TelemetryLogs struct {
	Enabled bool   `toml:"enabled"`
	Level   string `toml:"level"`
}
```

## type TelemetryMetrics

```go
type TelemetryMetrics struct {
	Enabled  bool   `toml:"enabled"`
	Interval string `toml:"interval"`
}
```

## type TelemetryTraces

```go
type TelemetryTraces struct {
	Enabled    bool    `toml:"enabled"`
	Sampler    string  `toml:"sampler"`
	SamplerArg float64 `toml:"sampler_arg"`
}
```

## type TimeoutsConfig

```go
type TimeoutsConfig struct {
	Idle  string `toml:"idle"`  // Duration string
	Stuck string `toml:"stuck"` // Duration string
}
```


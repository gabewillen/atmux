# package config

`import "github.com/agentflare-ai/amux/internal/config"`

- `func AddAgent(root string, agent AgentDef) error` — AddAgent appends an agent definition to the configuration file.
- `func RemoveAgent(root string, name string) error` — RemoveAgent removes an agent from the configuration file.
- `func loadEnv(cfg *Config) error` — loadEnv loads environment variables starting with AMUX__ and overrides configuration.
- `func loadFile(path string, cfg *Config) error`
- `func parseEnvValue(s string) any`
- `func saveFile(path string, cfg *Config) error`
- `func setPath(m map[string]any, path []string, value any) error`
- `type AdapterConfig` — AdapterConfig holds opaque configuration for adapters.
- `type AgentDef`
- `type AgentLocation`
- `type CoalesceConfig`
- `type Config` — Config represents the root configuration.
- `type DaemonConfig`
- `type EventsConfig`
- `type GeneralConfig`
- `type GitConfig`
- `type ManagerConfig`
- `type MergeConfig`
- `type NATSConfig`
- `type NodeConfig`
- `type PluginsConfig`
- `type ProcessConfig`
- `type RemoteConfig`
- `type RemoteNATSConfig`
- `type TelemetryConfig`
- `type TelemetryExporter`
- `type TelemetrySignal`
- `type TimeoutsConfig`

### Functions

#### AddAgent

```go
func AddAgent(root string, agent AgentDef) error
```

AddAgent appends an agent definition to the configuration file.
If root is empty, it writes to the user config (~/.config/amux/config.toml).
If root is provided, it writes to the project config (.amux/config.toml).

#### RemoveAgent

```go
func RemoveAgent(root string, name string) error
```

RemoveAgent removes an agent from the configuration file.

#### loadEnv

```go
func loadEnv(cfg *Config) error
```

loadEnv loads environment variables starting with AMUX__ and overrides configuration.
Spec §4.2.8.3

#### loadFile

```go
func loadFile(path string, cfg *Config) error
```

#### parseEnvValue

```go
func parseEnvValue(s string) any
```

#### saveFile

```go
func saveFile(path string, cfg *Config) error
```

#### setPath

```go
func setPath(m map[string]any, path []string, value any) error
```


## type AdapterConfig

```go
type AdapterConfig map[string]any
```

AdapterConfig holds opaque configuration for adapters.
The structure is map[string]any to allow arbitrary keys.

## type AgentDef

```go
type AgentDef struct {
	Name     string        `toml:"name"`
	About    string        `toml:"about"`
	Adapter  string        `toml:"adapter"`
	Location AgentLocation `toml:"location"`
}
```

## type AgentLocation

```go
type AgentLocation struct {
	Type     string `toml:"type"`
	Host     string `toml:"host,omitempty"`
	RepoPath string `toml:"repo_path,omitempty"`
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
```

Config represents the root configuration.

### Functions returning Config

#### DefaultConfig

```go
func DefaultConfig() Config
```

DefaultConfig returns the built-in default configuration.
Spec §4.2.8.2 (1. Built-in defaults)

#### Load

```go
func Load(root string) (*Config, error)
```

LoadConfig loads the configuration respecting the hierarchy.
root is the repository root for project-level config.


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

## type ManagerConfig

```go
type ManagerConfig struct {
	Enabled bool   `toml:"enabled"`
	Model   string `toml:"model"`
}
```

## type MergeConfig

```go
type MergeConfig struct {
	Strategy   string `toml:"strategy"`
	AllowDirty bool   `toml:"allow_dirty"`
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
	Transport            string           `toml:"transport"`
	BufferSize           string           `toml:"buffer_size"`
	RequestTimeout       string           `toml:"request_timeout"`
	ReconnectMaxAttempts int              `toml:"reconnect_max_attempts"`
	ReconnectBackoffBase string           `toml:"reconnect_backoff_base"`
	ReconnectBackoffMax  string           `toml:"reconnect_backoff_max"`
	NATS                 RemoteNATSConfig `toml:"nats"`
	Manager              ManagerConfig    `toml:"manager"`
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
	Traces      TelemetrySignal   `toml:"traces"`
	Metrics     TelemetrySignal   `toml:"metrics"`
	Logs        TelemetrySignal   `toml:"logs"`
}
```

## type TelemetryExporter

```go
type TelemetryExporter struct {
	Endpoint string `toml:"endpoint"`
	Protocol string `toml:"protocol"`
}
```

## type TelemetrySignal

```go
type TelemetrySignal struct {
	Enabled    bool    `toml:"enabled"`
	Sampler    string  `toml:"sampler,omitempty"`
	SamplerArg float64 `toml:"sampler_arg,omitempty"`
	Interval   string  `toml:"interval,omitempty"`
	Level      string  `toml:"level,omitempty"`
}
```

## type TimeoutsConfig

```go
type TimeoutsConfig struct {
	Idle  string `toml:"idle"`  // Duration string
	Stuck string `toml:"stuck"` // Duration string
}
```


# package config

`import "github.com/agentflare-ai/amux/internal/config"`

Package config provides configuration management for amux.

Configuration is loaded in a hierarchical order where later sources
override earlier ones:
 1. Built-in defaults
 2. Adapter defaults (from WASM or config.default.toml)
 3. User config (~/.config/amux/config.toml)
 4. User adapter config (~/.config/amux/adapters/{name}/config.toml)
 5. Project config (.amux/config.toml)
 6. Project adapter config (.amux/adapters/{name}/config.toml)
 7. Environment variables (AMUX__* prefix)

This package follows the configuration conventions in spec §4.2.8.

- `defaultLoader`
- `type AgentConfig` — AgentConfig holds agent definition settings.
- `type ByteSize` — ByteSize is a wrapper for byte sizes that supports parsing strings like "1MB", "64KB".
- `type CoalesceConfig` — CoalesceConfig holds event coalescing settings.
- `type Config` — Config holds the complete application configuration.
- `type DaemonConfig` — DaemonConfig holds daemon settings.
- `type Duration` — Duration is a wrapper around time.Duration that supports TOML parsing with Go duration strings (e.g., "30s", "5m").
- `type EventsConfig` — EventsConfig holds event system settings.
- `type GeneralConfig` — GeneralConfig holds general application settings.
- `type GitConfig` — GitConfig holds git-related settings.
- `type GitMergeConfig` — GitMergeConfig holds git merge settings.
- `type Loader` — Loader loads configuration from multiple sources.
- `type LocationConfig` — LocationConfig holds agent location settings.
- `type ManagerConfig` — ManagerConfig holds host manager settings.
- `type NATSConfig` — NATSConfig holds NATS server settings.
- `type NodeConfig` — NodeConfig holds node role settings.
- `type PluginsConfig` — PluginsConfig holds plugin settings.
- `type ProcessConfig` — ProcessConfig holds process tracking settings.
- `type RemoteConfig` — RemoteConfig holds remote agent settings.
- `type RemoteNATSConfig` — RemoteNATSConfig holds NATS-specific remote settings.
- `type SubscriptionConfig` — SubscriptionConfig holds event subscription settings.
- `type TelemetryConfig` — TelemetryConfig holds OpenTelemetry settings.
- `type TelemetryExporterConfig` — TelemetryExporterConfig holds exporter settings.
- `type TelemetryLogsConfig` — TelemetryLogsConfig holds logs settings.
- `type TelemetryMetricsConfig` — TelemetryMetricsConfig holds metrics settings.
- `type TelemetryTracesConfig` — TelemetryTracesConfig holds trace settings.
- `type TimeoutsConfig` — TimeoutsConfig holds timeout settings.

### Variables

#### defaultLoader

```go
var defaultLoader = NewLoader(nil)
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

AgentConfig holds agent definition settings.

## type ByteSize

```go
type ByteSize struct {
	Bytes int64
}
```

ByteSize is a wrapper for byte sizes that supports parsing strings like "1MB", "64KB".
Units are binary (1KB = 1024 bytes).

### Methods

#### ByteSize.MarshalText

```go
func () MarshalText() ([]byte, error)
```

MarshalText implements encoding.TextMarshaler for ByteSize.

#### ByteSize.UnmarshalText

```go
func () UnmarshalText(text []byte) error
```

UnmarshalText implements encoding.TextUnmarshaler for ByteSize.


## type CoalesceConfig

```go
type CoalesceConfig struct {
	IOStreams bool `toml:"io_streams"`
	Presence  bool `toml:"presence"`
	Activity  bool `toml:"activity"`
}
```

CoalesceConfig holds event coalescing settings.

## type Config

```go
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
```

Config holds the complete application configuration.

### Functions returning Config

#### DefaultConfig

```go
func DefaultConfig() *Config
```

DefaultConfig returns the built-in default configuration.

#### Get

```go
func Get() *Config
```

Get returns the current configuration from the default loader.

#### Load

```go
func Load() (*Config, error)
```

Load loads configuration using the default loader.


## type DaemonConfig

```go
type DaemonConfig struct {
	SocketPath string `toml:"socket_path"`
	AutoStart  bool   `toml:"autostart"`
}
```

DaemonConfig holds daemon settings.

## type Duration

```go
type Duration struct {
	time.Duration
}
```

Duration is a wrapper around time.Duration that supports TOML parsing
with Go duration strings (e.g., "30s", "5m").

### Methods

#### Duration.MarshalText

```go
func () MarshalText() ([]byte, error)
```

MarshalText implements encoding.TextMarshaler for Duration.

#### Duration.UnmarshalText

```go
func () UnmarshalText(text []byte) error
```

UnmarshalText implements encoding.TextUnmarshaler for Duration.


## type EventsConfig

```go
type EventsConfig struct {
	BatchWindow    Duration           `toml:"batch_window"`
	BatchMaxEvents int                `toml:"batch_max_events"`
	BatchMaxBytes  ByteSize           `toml:"batch_max_bytes"`
	BatchIdleFlush Duration           `toml:"batch_idle_flush"`
	Coalesce       CoalesceConfig     `toml:"coalesce"`
	Subscriptions  SubscriptionConfig `toml:"subscriptions"`
}
```

EventsConfig holds event system settings.

## type GeneralConfig

```go
type GeneralConfig struct {
	LogLevel  string `toml:"log_level"`
	LogFormat string `toml:"log_format"`
}
```

GeneralConfig holds general application settings.

## type GitConfig

```go
type GitConfig struct {
	Merge GitMergeConfig `toml:"merge"`
}
```

GitConfig holds git-related settings.

## type GitMergeConfig

```go
type GitMergeConfig struct {
	Strategy     string `toml:"strategy"`
	AllowDirty   bool   `toml:"allow_dirty"`
	TargetBranch string `toml:"target_branch"`
}
```

GitMergeConfig holds git merge settings.

## type Loader

```go
type Loader struct {
	mu       sync.RWMutex
	config   *Config
	resolver *paths.Resolver
}
```

Loader loads configuration from multiple sources.

### Functions returning Loader

#### NewLoader

```go
func NewLoader(resolver *paths.Resolver) *Loader
```

NewLoader creates a new configuration loader.


### Methods

#### Loader.Config

```go
func () Config() *Config
```

Config returns the current configuration.

#### Loader.Load

```go
func () Load() (*Config, error)
```

Load loads configuration from all sources in order.

#### Loader.loadEnv

```go
func () loadEnv() error
```

loadEnv loads configuration from environment variables.
Environment variables use the prefix AMUX__ and path segments are separated by __.

#### Loader.loadFile

```go
func () loadFile(path string) error
```

loadFile loads a TOML configuration file and merges it with current config.

#### Loader.merge

```go
func () merge(source *Config)
```

merge merges source config into the loader's config.
Non-zero values in source override values in dest.

#### Loader.setConfigValue

```go
func () setConfigValue(path []string, value string) error
```

setConfigValue sets a configuration value from a path and string value.


## type LocationConfig

```go
type LocationConfig struct {
	Type     string `toml:"type"`
	Host     string `toml:"host"`
	User     string `toml:"user"`
	Port     int    `toml:"port"`
	RepoPath string `toml:"repo_path"`
}
```

LocationConfig holds agent location settings.

## type ManagerConfig

```go
type ManagerConfig struct {
	Enabled bool   `toml:"enabled"`
	Model   string `toml:"model"`
}
```

ManagerConfig holds host manager settings.

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

NATSConfig holds NATS server settings.

## type NodeConfig

```go
type NodeConfig struct {
	Role string `toml:"role"`
}
```

NodeConfig holds node role settings.

## type PluginsConfig

```go
type PluginsConfig struct {
	Dir         string `toml:"dir"`
	AllowRemote bool   `toml:"allow_remote"`
}
```

PluginsConfig holds plugin settings.

## type ProcessConfig

```go
type ProcessConfig struct {
	CaptureMode      string   `toml:"capture_mode"`
	StreamBufferSize ByteSize `toml:"stream_buffer_size"`
	HookMode         string   `toml:"hook_mode"`
	PollInterval     Duration `toml:"poll_interval"`
	HookSocketDir    string   `toml:"hook_socket_dir"`
}
```

ProcessConfig holds process tracking settings.

## type RemoteConfig

```go
type RemoteConfig struct {
	Transport            string           `toml:"transport"`
	BufferSize           ByteSize         `toml:"buffer_size"`
	RequestTimeout       Duration         `toml:"request_timeout"`
	ReconnectMaxAttempts int              `toml:"reconnect_max_attempts"`
	ReconnectBackoffBase Duration         `toml:"reconnect_backoff_base"`
	ReconnectBackoffMax  Duration         `toml:"reconnect_backoff_max"`
	NATS                 RemoteNATSConfig `toml:"nats"`
	Manager              ManagerConfig    `toml:"manager"`
}
```

RemoteConfig holds remote agent settings.

## type RemoteNATSConfig

```go
type RemoteNATSConfig struct {
	URL               string   `toml:"url"`
	CredsPath         string   `toml:"creds_path"`
	SubjectPrefix     string   `toml:"subject_prefix"`
	KVBucket          string   `toml:"kv_bucket"`
	StreamEvents      string   `toml:"stream_events"`
	StreamPTY         string   `toml:"stream_pty"`
	HeartbeatInterval Duration `toml:"heartbeat_interval"`
}
```

RemoteNATSConfig holds NATS-specific remote settings.

## type SubscriptionConfig

```go
type SubscriptionConfig struct {
	Enabled    bool   `toml:"enabled"`
	SocketPath string `toml:"socket_path"`
}
```

SubscriptionConfig holds event subscription settings.

## type TelemetryConfig

```go
type TelemetryConfig struct {
	Enabled     bool                    `toml:"enabled"`
	ServiceName string                  `toml:"service_name"`
	Exporter    TelemetryExporterConfig `toml:"exporter"`
	Traces      TelemetryTracesConfig   `toml:"traces"`
	Metrics     TelemetryMetricsConfig  `toml:"metrics"`
	Logs        TelemetryLogsConfig     `toml:"logs"`
}
```

TelemetryConfig holds OpenTelemetry settings.

## type TelemetryExporterConfig

```go
type TelemetryExporterConfig struct {
	Endpoint string `toml:"endpoint"`
	Protocol string `toml:"protocol"`
}
```

TelemetryExporterConfig holds exporter settings.

## type TelemetryLogsConfig

```go
type TelemetryLogsConfig struct {
	Enabled bool   `toml:"enabled"`
	Level   string `toml:"level"`
}
```

TelemetryLogsConfig holds logs settings.

## type TelemetryMetricsConfig

```go
type TelemetryMetricsConfig struct {
	Enabled  bool     `toml:"enabled"`
	Interval Duration `toml:"interval"`
}
```

TelemetryMetricsConfig holds metrics settings.

## type TelemetryTracesConfig

```go
type TelemetryTracesConfig struct {
	Enabled    bool    `toml:"enabled"`
	Sampler    string  `toml:"sampler"`
	SamplerArg float64 `toml:"sampler_arg"`
}
```

TelemetryTracesConfig holds trace settings.

## type TimeoutsConfig

```go
type TimeoutsConfig struct {
	Idle  Duration `toml:"idle"`
	Stuck Duration `toml:"stuck"`
}
```

TimeoutsConfig holds timeout settings.


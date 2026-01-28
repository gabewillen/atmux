# package config

`import "github.com/agentflare-ai/amux/internal/config"`

Package config provides configuration management for amux.

Package config provides configuration management for amux.

Package config provides configuration management for amux.
Configuration is loaded from multiple sources in a defined hierarchy:
built-in defaults < adapter defaults < user config < project config < environment variables.

- `ConfigFileChanged, ConfigReloaded, ConfigUpdated` — Config events
- `ConfigModel` — ConfigModel defines the HSM model for configuration management.
- `func isSensitiveKey(key string) bool` — isSensitiveKey determines if a configuration key contains sensitive information.
- `type Actor` — Actor manages configuration state and live updates.
- `type AdapterConfig` — AdapterConfig represents adapter-specific configuration.
- `type AgentConfig` — AgentConfig represents a single agent definition.
- `type AgentLocationConfig` — AgentLocationConfig holds agent location information.
- `type ConfigChange` — ConfigChange represents a configuration value change.
- `type Config` — Config represents the complete amux configuration.
- `type DaemonConfig` — DaemonConfig holds daemon configuration.
- `type EventsCoalesceConfig` — EventsCoalesceConfig holds event coalescing settings.
- `type EventsConfig` — EventsConfig holds event batching and coalescing configuration.
- `type GeneralConfig` — GeneralConfig holds general application settings.
- `type GitConfig` — GitConfig holds git-related configuration.
- `type GitMergeConfig` — GitMergeConfig holds git merge strategy configuration.
- `type Loader` — Loader loads configuration from multiple sources.
- `type NATSConfig` — NATSConfig holds NATS server configuration.
- `type NodeConfig` — NodeConfig holds node role configuration.
- `type PluginsConfig` — PluginsConfig holds plugin configuration.
- `type ProcessConfig` — ProcessConfig holds process tracking configuration.
- `type RemoteConfig` — RemoteConfig holds remote agent configuration.
- `type RemoteManagerConfig` — RemoteManagerConfig holds remote manager configuration.
- `type RemoteNATSConfig` — RemoteNATSConfig holds NATS-specific remote configuration.
- `type TelemetryConfig` — TelemetryConfig holds OpenTelemetry configuration.
- `type TelemetryExporterConfig` — TelemetryExporterConfig holds OTel exporter configuration.
- `type TelemetryLogsConfig` — TelemetryLogsConfig holds OTel logs configuration.
- `type TelemetryMetricsConfig` — TelemetryMetricsConfig holds OTel metrics configuration.
- `type TelemetryTracesConfig` — TelemetryTracesConfig holds OTel traces configuration.
- `type TimeoutsConfig` — TimeoutsConfig holds timeout settings.

### Constants

#### ConfigFileChanged, ConfigReloaded, ConfigUpdated

```go
const (
	ConfigFileChanged = "config.file_changed" // File modified on disk
	ConfigReloaded    = "config.reloaded"     // Reload complete
	ConfigUpdated     = "config.updated"      // Specific value changed
)
```

Config events


### Variables

#### ConfigModel

```go
var ConfigModel = hsm.Define("config",
	hsm.State("loading"),
	hsm.State("ready"),
	hsm.State("reloading"),

	hsm.Transition(hsm.On(hsm.Event{Name: "config.loaded"}), hsm.Source("loading"), hsm.Target("ready")),
	hsm.Transition(hsm.On(hsm.Event{Name: ConfigFileChanged}), hsm.Source("ready"), hsm.Target("reloading")),
	hsm.Transition(hsm.On(hsm.Event{Name: ConfigReloaded}), hsm.Source("reloading"), hsm.Target("ready")),
	hsm.Transition(hsm.On(hsm.Event{Name: "config.reload_failed"}), hsm.Source("reloading"), hsm.Target("ready")),

	hsm.Initial(hsm.Target("loading")),
)
```

ConfigModel defines the HSM model for configuration management.


### Functions

#### isSensitiveKey

```go
func isSensitiveKey(key string) bool
```

isSensitiveKey determines if a configuration key contains sensitive information.


## type Actor

```go
type Actor struct {
	mu           sync.RWMutex
	config       *Config
	loader       *Loader
	subs         []chan ConfigChange
	stopWatching func() // Function to stop file watching
}
```

Actor manages configuration state and live updates.

### Functions returning Actor

#### NewActor

```go
func NewActor(loader *Loader) *Actor
```

NewActor creates a new configuration actor.


### Methods

#### Actor.Get

```go
func () Get() *Config
```

Get returns the current configuration.

#### Actor.Load

```go
func () Load() error
```

Load loads the initial configuration.

#### Actor.Reload

```go
func () Reload() error
```

Reload reloads configuration from disk.

#### Actor.Start

```go
func () Start(ctx context.Context) error
```

Start initializes the config actor and starts the HSM.

#### Actor.StartWatching

```go
func () StartWatching()
```

StartWatching starts watching configuration files for changes.
Phase 0: Placeholder - actual file watching would use fsnotify or similar.

#### Actor.StopWatching

```go
func () StopWatching()
```

StopWatching stops watching configuration files.

#### Actor.Subscribe

```go
func () Subscribe() <-chan ConfigChange
```

Subscribe returns a channel that receives configuration change notifications.

#### Actor.compareConfigs

```go
func () compareConfigs(old, new *Config) []ConfigChange
```

compareConfigs compares two configs and returns a list of changes.
Phase 0: Simplified comparison - full deep diff would be implemented later.

#### Actor.notifySubscribers

```go
func () notifySubscribers(change ConfigChange)
```

notifySubscribers notifies all subscribers of a configuration change.


## type AdapterConfig

```go
type AdapterConfig map[string]interface{}
```

AdapterConfig represents adapter-specific configuration.
Adapter configs are opaque to the core system.

## type AgentConfig

```go
type AgentConfig struct {
	Name     string              `toml:"name"`
	About    string              `toml:"about"`
	Adapter  string              `toml:"adapter"`
	Location AgentLocationConfig `toml:"location"`
}
```

AgentConfig represents a single agent definition.

## type AgentLocationConfig

```go
type AgentLocationConfig struct {
	Type     string `toml:"type"`
	Host     string `toml:"host"`
	User     string `toml:"user"`
	Port     int    `toml:"port"`
	RepoPath string `toml:"repo_path"`
}
```

AgentLocationConfig holds agent location information.

## type Config

```go
type Config struct {
	General   GeneralConfig          `toml:"general"`
	Timeouts  TimeoutsConfig         `toml:"timeouts"`
	Process   ProcessConfig          `toml:"process"`
	Git       GitConfig              `toml:"git"`
	Events    EventsConfig           `toml:"events"`
	Remote    RemoteConfig           `toml:"remote"`
	NATS      NATSConfig             `toml:"nats"`
	Node      NodeConfig             `toml:"node"`
	Daemon    DaemonConfig           `toml:"daemon"`
	Plugins   PluginsConfig          `toml:"plugins"`
	Telemetry TelemetryConfig        `toml:"telemetry"`
	Agents    []AgentConfig          `toml:"agents"`
	Adapters  map[string]interface{} `toml:"adapters"` // Opaque adapter configs
}
```

Config represents the complete amux configuration.

### Methods

#### Config.GetAdapterConfig

```go
func () GetAdapterConfig(adapterName string) AdapterConfig
```

GetAdapterConfig returns the configuration for a specific adapter.

#### Config.RedactSensitiveFields

```go
func () RedactSensitiveFields() *Config
```

RedactSensitiveFields redacts sensitive configuration fields for logging.
Per spec §4.2.8.6, sensitive values should not appear in logs.


## type ConfigChange

```go
type ConfigChange struct {
	Path     string // Config key path: "coordination.interval"
	OldValue any
	NewValue any
}
```

ConfigChange represents a configuration value change.

## type DaemonConfig

```go
type DaemonConfig struct {
	SocketPath string `toml:"socket_path"`
	Autostart  bool   `toml:"autostart"`
}
```

DaemonConfig holds daemon configuration.

## type EventsCoalesceConfig

```go
type EventsCoalesceConfig struct {
	IOStreams bool `toml:"io_streams"`
	Presence  bool `toml:"presence"`
	Activity  bool `toml:"activity"`
}
```

EventsCoalesceConfig holds event coalescing settings.

## type EventsConfig

```go
type EventsConfig struct {
	BatchWindow    string               `toml:"batch_window"`
	BatchMaxEvents int                  `toml:"batch_max_events"`
	BatchMaxBytes  string               `toml:"batch_max_bytes"`
	BatchIdleFlush string               `toml:"batch_idle_flush"`
	Coalesce       EventsCoalesceConfig `toml:"coalesce"`
}
```

EventsConfig holds event batching and coalescing configuration.

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

GitConfig holds git-related configuration.

## type GitMergeConfig

```go
type GitMergeConfig struct {
	Strategy   string `toml:"strategy"`
	AllowDirty bool   `toml:"allow_dirty"`
}
```

GitMergeConfig holds git merge strategy configuration.

## type Loader

```go
type Loader struct {
	configDir string
	homeDir   string
	repoRoot  string
}
```

Loader loads configuration from multiple sources.

### Functions returning Loader

#### NewLoader

```go
func NewLoader(configDir, homeDir, repoRoot string) *Loader
```

NewLoader creates a new configuration loader.


### Methods

#### Loader.Load

```go
func () Load() (*Config, error)
```

Load loads configuration from all sources in the correct order.

#### Loader.ValidateAdapterConfig

```go
func () ValidateAdapterConfig(adapterName string, cfg map[string]interface{}) error
```

ValidateAdapterConfig validates that adapter configuration is properly scoped.

#### Loader.applyEnvOverrides

```go
func () applyEnvOverrides(cfg *Config) error
```

applyEnvOverrides applies environment variable overrides to cfg.

#### Loader.defaultConfig

```go
func () defaultConfig() *Config
```

defaultConfig returns the built-in default configuration.

#### Loader.expandHome

```go
func () expandHome(path string) string
```

expandHome expands ~ to the home directory.

#### Loader.loadFile

```go
func () loadFile(path string, cfg *Config) error
```

loadFile loads configuration from a TOML file and merges it into cfg.


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

NATSConfig holds NATS server configuration.

## type NodeConfig

```go
type NodeConfig struct {
	Role string `toml:"role"`
}
```

NodeConfig holds node role configuration.

## type PluginsConfig

```go
type PluginsConfig struct {
	Dir         string `toml:"dir"`
	AllowRemote bool   `toml:"allow_remote"`
}
```

PluginsConfig holds plugin configuration.

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

ProcessConfig holds process tracking configuration.

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

RemoteConfig holds remote agent configuration.

## type RemoteManagerConfig

```go
type RemoteManagerConfig struct {
	Enabled bool   `toml:"enabled"`
	Model   string `toml:"model"`
}
```

RemoteManagerConfig holds remote manager configuration.

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

RemoteNATSConfig holds NATS-specific remote configuration.

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

TelemetryConfig holds OpenTelemetry configuration.

## type TelemetryExporterConfig

```go
type TelemetryExporterConfig struct {
	Endpoint string `toml:"endpoint"`
	Protocol string `toml:"protocol"`
}
```

TelemetryExporterConfig holds OTel exporter configuration.

## type TelemetryLogsConfig

```go
type TelemetryLogsConfig struct {
	Enabled bool   `toml:"enabled"`
	Level   string `toml:"level"`
}
```

TelemetryLogsConfig holds OTel logs configuration.

## type TelemetryMetricsConfig

```go
type TelemetryMetricsConfig struct {
	Enabled  bool   `toml:"enabled"`
	Interval string `toml:"interval"`
}
```

TelemetryMetricsConfig holds OTel metrics configuration.

## type TelemetryTracesConfig

```go
type TelemetryTracesConfig struct {
	Enabled    bool    `toml:"enabled"`
	Sampler    string  `toml:"sampler"`
	SamplerArg float64 `toml:"sampler_arg"`
}
```

TelemetryTracesConfig holds OTel traces configuration.

## type TimeoutsConfig

```go
type TimeoutsConfig struct {
	Idle  string `toml:"idle"`  // Duration string
	Stuck string `toml:"stuck"` // Duration string
}
```

TimeoutsConfig holds timeout settings.


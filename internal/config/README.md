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

- `ErrActorClosed` — ErrActorClosed is returned when operating on a closed actor.
- `defaultLoader`
- `func HotReloadableKeys() []string` — HotReloadableKeys returns the list of config keys that can be hot-reloaded.
- `func IsHotReloadable(path string) bool` — IsHotReloadable checks if a config key path is hot-reloadable.
- `func SetDefaultActor(a *Actor)` — SetDefaultActor sets the global config actor.
- `func Subscribe(handler func(ctx context.Context, change ConfigChange)) func()` — Subscribe registers a handler for config change events using the default dispatcher.
- `func SubscribeAll(
	onFileChanged func(ctx context.Context),
	onReloaded func(ctx context.Context),
	onUpdated func(ctx context.Context, change ConfigChange),
	onReloadFailed func(ctx context.Context, err string),
) func()` — SubscribeAll registers handlers for all config events.
- `func WatchFile(path string)` — WatchFile adds a file to the watch list for the default actor.
- `type ActorClosedError` — ActorClosedError indicates the actor is closed.
- `type Actor` — Actor is the configuration actor that manages config loading and live updates.
- `type AgentConfig` — AgentConfig holds agent definition settings.
- `type ByteSize` — ByteSize is a wrapper for byte sizes that supports parsing strings like "1MB", "64KB".
- `type CoalesceConfig` — CoalesceConfig holds event coalescing settings.
- `type ConfigChange` — ConfigChange represents a change to a configuration value.
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
- `type ShutdownConfig` — ShutdownConfig holds graceful shutdown settings per spec §5.6.
- `type State` — State represents the config actor state.
- `type SubscriptionConfig` — SubscriptionConfig holds event subscription settings.
- `type TelemetryConfig` — TelemetryConfig holds OpenTelemetry settings.
- `type TelemetryExporterConfig` — TelemetryExporterConfig holds exporter settings.
- `type TelemetryLogsConfig` — TelemetryLogsConfig holds logs settings.
- `type TelemetryMetricsConfig` — TelemetryMetricsConfig holds metrics settings.
- `type TelemetryTracesConfig` — TelemetryTracesConfig holds trace settings.
- `type TimeoutsConfig` — TimeoutsConfig holds timeout settings.

### Variables

#### ErrActorClosed

```go
var ErrActorClosed = &ActorClosedError{}
```

ErrActorClosed is returned when operating on a closed actor.

#### defaultLoader

```go
var defaultLoader = NewLoader(nil)
```


### Functions

#### HotReloadableKeys

```go
func HotReloadableKeys() []string
```

HotReloadableKeys returns the list of config keys that can be hot-reloaded.
Per spec §4.2.8.9, these keys can be updated without restart.

#### IsHotReloadable

```go
func IsHotReloadable(path string) bool
```

IsHotReloadable checks if a config key path is hot-reloadable.

#### SetDefaultActor

```go
func SetDefaultActor(a *Actor)
```

SetDefaultActor sets the global config actor.

#### Subscribe

```go
func Subscribe(handler func(ctx context.Context, change ConfigChange)) func()
```

Subscribe registers a handler for config change events using the default dispatcher.
Returns an unsubscribe function.

#### SubscribeAll

```go
func SubscribeAll(
	onFileChanged func(ctx context.Context),
	onReloaded func(ctx context.Context),
	onUpdated func(ctx context.Context, change ConfigChange),
	onReloadFailed func(ctx context.Context, err string),
) func()
```

SubscribeAll registers handlers for all config events.
Returns an unsubscribe function.

#### WatchFile

```go
func WatchFile(path string)
```

WatchFile adds a file to the watch list for the default actor.


## type Actor

```go
type Actor struct {
	mu sync.RWMutex

	// id is the actor's unique identifier.
	id muid.MUID

	// state is the current actor state.
	state State

	// loader is the configuration loader.
	loader *Loader

	// config is the current configuration.
	config *Config

	// dispatcher is used to emit config events.
	dispatcher event.Dispatcher

	// watchCtx controls the file watcher goroutine.
	watchCtx    context.Context
	watchCancel context.CancelFunc

	// watchedFiles is the list of files being watched.
	watchedFiles []string

	// modTimes tracks file modification times.
	modTimes map[string]time.Time

	// pollInterval is how often to check for file changes.
	pollInterval time.Duration

	// closed indicates the actor has been closed.
	closed bool
}
```

Actor is the configuration actor that manages config loading and live updates.
It implements an HSM-like state machine per spec §4.2.8.7.

### Variables

#### defaultActor, actorMu

```go
var (
	defaultActor *Actor
	actorMu      sync.RWMutex
)
```

DefaultActor is the global config actor.


### Functions returning Actor

#### GetDefaultActor

```go
func GetDefaultActor() *Actor
```

GetDefaultActor returns the global config actor.

#### NewActor

```go
func NewActor(resolver *paths.Resolver, dispatcher event.Dispatcher) *Actor
```

NewActor creates a new configuration actor.


### Methods

#### Actor.Close

```go
func () Close() error
```

Close stops the configuration actor and file watcher.

#### Actor.Config

```go
func () Config() *Config
```

Config returns the current configuration.

#### Actor.ID

```go
func () ID() muid.MUID
```

ID returns the actor's unique identifier.

#### Actor.Reload

```go
func () Reload(ctx context.Context) error
```

Reload forces a configuration reload.

#### Actor.Start

```go
func () Start(ctx context.Context) error
```

Start loads configuration and transitions to the ready state.
This begins file watching for live updates.

#### Actor.State

```go
func () State() State
```

State returns the current state of the actor.

#### Actor.checkFileChanges

```go
func () checkFileChanges() bool
```

checkFileChanges checks if any watched files have changed.

#### Actor.collectWatchedFiles

```go
func () collectWatchedFiles() []string
```

collectWatchedFiles returns the list of config files to watch.

#### Actor.dispatchChanges

```go
func () dispatchChanges(ctx context.Context, oldConfig, newConfig *Config)
```

dispatchChanges compares old and new configs and dispatches config.updated events.

#### Actor.dispatchFileChanged

```go
func () dispatchFileChanged(ctx context.Context)
```

dispatchFileChanged dispatches a config.file_changed event.

#### Actor.dispatchReloadFailed

```go
func () dispatchReloadFailed(ctx context.Context, err error)
```

dispatchReloadFailed dispatches a config.reload_failed event.

#### Actor.dispatchReloaded

```go
func () dispatchReloaded(ctx context.Context)
```

dispatchReloaded dispatches a config.reloaded event.

#### Actor.updateModTimes

```go
func () updateModTimes()
```

updateModTimes updates the modification times for watched files.

#### Actor.watchLoop

```go
func () watchLoop()
```

watchLoop polls for file changes.


## type ActorClosedError

```go
type ActorClosedError struct{}
```

ActorClosedError indicates the actor is closed.

### Methods

#### ActorClosedError.Error

```go
func () Error() string
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


## type ConfigChange

```go
type ConfigChange struct {
	// Path is the config key path (e.g., "timeouts.idle").
	Path string `json:"path"`

	// OldValue is the previous value.
	OldValue any `json:"old_value,omitempty"`

	// NewValue is the new value.
	NewValue any `json:"new_value"`
}
```

ConfigChange represents a change to a configuration value.
Used in config.updated events per spec §4.2.8.8.

### Functions returning ConfigChange

#### compareConfigs

```go
func compareConfigs(old, new *Config) []ConfigChange
```

compareConfigs compares two configs and returns a list of changes.


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

#### Loader.expandPaths

```go
func () expandPaths()
```

expandPaths expands all path fields that start with ~/ to the user's home directory.

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

#### Loader.setAdaptersConfig

```go
func () setAdaptersConfig(path []string, value string) error
```

#### Loader.setConfigValue

```go
func () setConfigValue(path []string, value string) error
```

setConfigValue sets a configuration value from a path and string value.

#### Loader.setDaemonConfig

```go
func () setDaemonConfig(path []string, value string) error
```

#### Loader.setEventsConfig

```go
func () setEventsConfig(path []string, value string) error
```

#### Loader.setGeneralConfig

```go
func () setGeneralConfig(path []string, value string) error
```

#### Loader.setGitConfig

```go
func () setGitConfig(path []string, value string) error
```

#### Loader.setNATSConfig

```go
func () setNATSConfig(path []string, value string) error
```

#### Loader.setNodeConfig

```go
func () setNodeConfig(path []string, value string) error
```

#### Loader.setPluginsConfig

```go
func () setPluginsConfig(path []string, value string) error
```

#### Loader.setProcessConfig

```go
func () setProcessConfig(path []string, value string) error
```

#### Loader.setRemoteConfig

```go
func () setRemoteConfig(path []string, value string) error
```

#### Loader.setShutdownConfig

```go
func () setShutdownConfig(path []string, value string) error
```

#### Loader.setTelemetryConfig

```go
func () setTelemetryConfig(path []string, value string) error
```

#### Loader.setTimeoutsConfig

```go
func () setTimeoutsConfig(path []string, value string) error
```


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

## type ShutdownConfig

```go
type ShutdownConfig struct {
	DrainTimeout     Duration `toml:"drain_timeout"`
	CleanupWorktrees bool     `toml:"cleanup_worktrees"`
}
```

ShutdownConfig holds graceful shutdown settings per spec §5.6.

## type State

```go
type State string
```

State represents the config actor state.

### Constants

#### StateLoading, StateReady, StateReloading

```go
const (
	// StateLoading is the initial state while loading configuration.
	StateLoading State = "loading"

	// StateReady is the state when configuration is loaded and watching for changes.
	StateReady State = "ready"

	// StateReloading is the state while reloading configuration after a file change.
	StateReloading State = "reloading"
)
```


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


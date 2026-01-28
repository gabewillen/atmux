# package config

`import "github.com/stateforward/amux/internal/config"`

Package config provides configuration management for amux per spec §4.2.8.

Configuration is loaded from TOML files in a hierarchy:
1. Built-in defaults
2. Adapter defaults (from WASM manifests)
3. User config (~/.config/amux/config.toml)
4. Project config (.amux/config.toml)
5. Environment variables (AMUX__ prefix)

The configuration supports:
- Live reload via file watching
- Opaque adapter config blocks
- Sensitive configuration redaction
- HSM-based config actor for updates

- `func ParseByteSize(s string) (int64, error)` — ParseByteSize parses a byte size string (e.g., "1MB", "512KB").
- `func applyEnvOverrides(cfg *Config) error` — applyEnvOverrides applies environment variable overrides to the config.
- `type AgentConfig` — AgentConfig represents a single agent configuration.
- `type ByteSize` — ByteSize represents a byte size that can be parsed from TOML.
- `type CoalesceConfig` — CoalesceConfig holds event coalescing configuration.
- `type Config` — Config represents the complete amux configuration.
- `type DaemonConfig` — DaemonConfig holds daemon configuration.
- `type Duration` — Duration is a time.Duration that can be parsed from TOML.
- `type EventsConfig` — EventsConfig holds event system configuration.
- `type GeneralConfig` — GeneralConfig holds general application settings.
- `type GitConfig` — GitConfig holds git-related configuration.
- `type GitMergeConfig` — GitMergeConfig holds git merge strategy configuration.
- `type LocationConfig` — LocationConfig represents agent location configuration.
- `type NATSConfig` — NATSConfig holds NATS server configuration.
- `type NodeConfig` — NodeConfig holds node role configuration.
- `type PluginsConfig` — PluginsConfig holds plugin configuration.
- `type ProcessConfig` — ProcessConfig holds process tracking configuration.
- `type RemoteConfig` — RemoteConfig holds remote agent configuration.
- `type RemoteManagerConfig` — RemoteManagerConfig holds remote manager configuration.
- `type RemoteNATSConfig` — RemoteNATSConfig holds NATS configuration for remote agents.
- `type TimeoutsConfig` — TimeoutsConfig holds timeout durations.

### Functions

#### ParseByteSize

```go
func ParseByteSize(s string) (int64, error)
```

ParseByteSize parses a byte size string (e.g., "1MB", "512KB").

#### applyEnvOverrides

```go
func applyEnvOverrides(cfg *Config) error
```

applyEnvOverrides applies environment variable overrides to the config.
Environment variables follow the pattern: AMUX__<path>__<path>...


## type AgentConfig

```go
type AgentConfig struct {
	Name     string         `toml:"name"`     // Agent name
	About    string         `toml:"about"`    // Agent description
	Adapter  string         `toml:"adapter"`  // Adapter name (string reference)
	Location LocationConfig `toml:"location"` // Location configuration
}
```

AgentConfig represents a single agent configuration.

## type ByteSize

```go
type ByteSize int64
```

ByteSize represents a byte size that can be parsed from TOML.

### Methods

#### ByteSize.MarshalText

```go
func () MarshalText() ([]byte, error)
```

MarshalText implements the encoding.TextMarshaler interface.

#### ByteSize.UnmarshalText

```go
func () UnmarshalText(text []byte) error
```

UnmarshalText implements the encoding.TextUnmarshaler interface.


## type CoalesceConfig

```go
type CoalesceConfig struct {
	IOStreams bool `toml:"io_streams"` // Coalesce stdout/stderr/stdin per process
	Presence  bool `toml:"presence"`   // Keep only latest presence per agent
	Activity  bool `toml:"activity"`   // Deduplicate activity events
}
```

CoalesceConfig holds event coalescing configuration.

## type Config

```go
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
	Adapters map[string]any `toml:"adapters"` // Opaque adapter configs
	Agents   []AgentConfig  `toml:"agents"`
}
```

Config represents the complete amux configuration.

### Functions returning Config

#### DefaultConfig

```go
func DefaultConfig() *Config
```

DefaultConfig returns the default configuration.

#### Load

```go
func Load(paths ...string) (*Config, error)
```

Load loads configuration from the specified paths and environment variables.


## type DaemonConfig

```go
type DaemonConfig struct {
	SocketPath string `toml:"socket_path"` // Unix socket path
	Autostart  bool   `toml:"autostart"`   // Auto-start daemon
}
```

DaemonConfig holds daemon configuration.

## type Duration

```go
type Duration time.Duration
```

Duration is a time.Duration that can be parsed from TOML.

### Methods

#### Duration.MarshalText

```go
func () MarshalText() ([]byte, error)
```

MarshalText implements the encoding.TextMarshaler interface.

#### Duration.UnmarshalText

```go
func () UnmarshalText(text []byte) error
```

UnmarshalText implements the encoding.TextUnmarshaler interface.


## type EventsConfig

```go
type EventsConfig struct {
	BatchWindow    Duration       `toml:"batch_window"`     // Coalesce window
	BatchMaxEvents int            `toml:"batch_max_events"` // Maximum events per batch
	BatchMaxBytes  ByteSize       `toml:"batch_max_bytes"`  // Maximum bytes for I/O batches
	BatchIdleFlush Duration       `toml:"batch_idle_flush"` // Flush if idle
	Coalesce       CoalesceConfig `toml:"coalesce"`         // Coalescing rules
}
```

EventsConfig holds event system configuration.

## type GeneralConfig

```go
type GeneralConfig struct {
	LogLevel  string `toml:"log_level"`  // debug, info, warn, error
	LogFormat string `toml:"log_format"` // text, json
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
	Strategy     string `toml:"strategy"`      // merge-commit, squash, rebase, ff-only
	AllowDirty   bool   `toml:"allow_dirty"`   // Allow merges with uncommitted changes
	TargetBranch string `toml:"target_branch"` // Target branch for merges
}
```

GitMergeConfig holds git merge strategy configuration.

## type LocationConfig

```go
type LocationConfig struct {
	Type     string `toml:"type"`      // local, ssh
	Host     string `toml:"host"`      // SSH host (for type=ssh)
	RepoPath string `toml:"repo_path"` // Repository path
}
```

LocationConfig represents agent location configuration.

## type NATSConfig

```go
type NATSConfig struct {
	Mode         string `toml:"mode"`          // embedded, external
	Topology     string `toml:"topology"`      // hub, leaf
	HubURL       string `toml:"hub_url"`       // Hub URL (for leaf)
	Listen       string `toml:"listen"`        // Listen address
	AdvertiseURL string `toml:"advertise_url"` // Advertise URL
	JetstreamDir string `toml:"jetstream_dir"` // JetStream directory
}
```

NATSConfig holds NATS server configuration.

## type NodeConfig

```go
type NodeConfig struct {
	Role string `toml:"role"` // director, manager
}
```

NodeConfig holds node role configuration.

## type PluginsConfig

```go
type PluginsConfig struct {
	Dir         string `toml:"dir"`          // Plugin directory
	AllowRemote bool   `toml:"allow_remote"` // Allow remote plugins
}
```

PluginsConfig holds plugin configuration.

## type ProcessConfig

```go
type ProcessConfig struct {
	CaptureMode      string   `toml:"capture_mode"`       // none, stdout, stderr, stdin, all
	StreamBufferSize ByteSize `toml:"stream_buffer_size"` // Ring buffer size per stream
	HookMode         string   `toml:"hook_mode"`          // auto, preload, polling, disabled
	PollInterval     Duration `toml:"poll_interval"`      // Polling interval
	HookSocketDir    string   `toml:"hook_socket_dir"`    // Directory for hook Unix sockets
}
```

ProcessConfig holds process tracking configuration.

## type RemoteConfig

```go
type RemoteConfig struct {
	Transport            string              `toml:"transport"`              // nats, ssh_yamux
	BufferSize           ByteSize            `toml:"buffer_size"`            // Per-session PTY replay buffer
	RequestTimeout       Duration            `toml:"request_timeout"`        // NATS request-reply timeout
	ReconnectMaxAttempts int                 `toml:"reconnect_max_attempts"` // Max reconnection attempts
	ReconnectBackoffBase Duration            `toml:"reconnect_backoff_base"` // Base backoff duration
	ReconnectBackoffMax  Duration            `toml:"reconnect_backoff_max"`  // Max backoff duration
	NATS                 RemoteNATSConfig    `toml:"nats"`                   // NATS configuration
	Manager              RemoteManagerConfig `toml:"manager"`                // Manager configuration
}
```

RemoteConfig holds remote agent configuration.

## type RemoteManagerConfig

```go
type RemoteManagerConfig struct {
	Enabled bool   `toml:"enabled"` // Enable manager mode
	Model   string `toml:"model"`   // LLM model ID
}
```

RemoteManagerConfig holds remote manager configuration.

## type RemoteNATSConfig

```go
type RemoteNATSConfig struct {
	URL               string   `toml:"url"`                // NATS server URL
	CredsPath         string   `toml:"creds_path"`         // Per-host NATS credential file
	SubjectPrefix     string   `toml:"subject_prefix"`     // Root subject namespace
	KVBucket          string   `toml:"kv_bucket"`          // JetStream KV bucket
	StreamEvents      string   `toml:"stream_events"`      // JetStream events stream
	StreamPTY         string   `toml:"stream_pty"`         // JetStream PTY stream
	HeartbeatInterval Duration `toml:"heartbeat_interval"` // Heartbeat interval
}
```

RemoteNATSConfig holds NATS configuration for remote agents.

## type TimeoutsConfig

```go
type TimeoutsConfig struct {
	Idle  Duration `toml:"idle"`  // Idle timeout
	Stuck Duration `toml:"stuck"` // Stuck timeout
}
```

TimeoutsConfig holds timeout durations.


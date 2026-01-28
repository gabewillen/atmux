# package config

`import "github.com/agentflare-ai/amux/internal/config"`

Package config provides configuration management with live updates.

Package config provides configuration loading and management for amux.

- `func Save(config *Config, path string) error` — Save saves configuration to TOML format.
- `func Validate(config *Config) error` — Validate validates configuration values.
- `func ValidateAdapterConfig(config *AdapterConfig) error` — ValidateAdapterConfig validates a specific adapter configuration.
- `func applyEnvOverrides(config *Config) error` — applyEnvOverrides applies environment variable overrides.
- `func loadAdapterConfigs(config *Config) error` — loadAdapterConfigs loads adapter-specific configurations.
- `func loadProjectConfig(config *Config) error` — loadProjectConfig loads project-specific configuration from git repo.
- `func loadUserConfig(config *Config) error` — loadUserConfig loads user configuration from standard paths.
- `func parseByteSize(s string) (int, error)` — parseByteSize parses byte size strings like "1MB", "64KB" etc.
- `func setDefaults(config *Config)` — setDefaults sets built-in default values.
- `func setEnvOverride(config *Config, key, value string) error` — setEnvOverride sets a single environment variable override.
- `type Actor` — Actor manages configuration with live updates and subscriptions.
- `type AdapterConfig` — AdapterConfig represents adapter-specific configuration with sensitive field handling.
- `type AgentsConfig` — AgentsConfig contains agent management settings.
- `type Config` — Config represents the complete amux configuration.
- `type CoreConfig` — CoreConfig contains core daemon configuration.
- `type EventsConfig` — EventsConfig contains event system settings.
- `type InferenceConfig` — InferenceConfig contains local inference settings.
- `type JetStreamConfig` — JetStreamConfig contains JetStream-specific settings.
- `type ModelConfig` — ModelConfig contains model-specific settings.
- `type MonitorConfig` — MonitorConfig contains PTY monitoring settings.
- `type OTelConfig` — OTelConfig contains OpenTelemetry settings.
- `type OTelExporterConfig` — OTelExporterConfig contains OTel exporter settings.
- `type ProcessConfig` — ProcessConfig contains process tracking settings.
- `type RemoteConfig` — RemoteConfig contains remote orchestration settings.
- `type SubscriptionsConfig` — SubscriptionsConfig contains subscription settings.
- `type Watcher` — Watcher represents a configuration change subscriber.

### Functions

#### Save

```go
func Save(config *Config, path string) error
```

Save saves configuration to TOML format.

#### Validate

```go
func Validate(config *Config) error
```

Validate validates configuration values.

#### ValidateAdapterConfig

```go
func ValidateAdapterConfig(config *AdapterConfig) error
```

ValidateAdapterConfig validates a specific adapter configuration.

#### applyEnvOverrides

```go
func applyEnvOverrides(config *Config) error
```

applyEnvOverrides applies environment variable overrides.

#### loadAdapterConfigs

```go
func loadAdapterConfigs(config *Config) error
```

loadAdapterConfigs loads adapter-specific configurations.

#### loadProjectConfig

```go
func loadProjectConfig(config *Config) error
```

loadProjectConfig loads project-specific configuration from git repo.

#### loadUserConfig

```go
func loadUserConfig(config *Config) error
```

loadUserConfig loads user configuration from standard paths.

#### parseByteSize

```go
func parseByteSize(s string) (int, error)
```

parseByteSize parses byte size strings like "1MB", "64KB" etc.

#### setDefaults

```go
func setDefaults(config *Config)
```

setDefaults sets built-in default values.

#### setEnvOverride

```go
func setEnvOverride(config *Config, key, value string) error
```

setEnvOverride sets a single environment variable override.


## type Actor

```go
type Actor struct {
	mu       sync.RWMutex
	config   *Config
	watchers []chan *Config
	ctx      context.Context
	cancel   context.CancelFunc
}
```

Actor manages configuration with live updates and subscriptions.

### Functions returning Actor

#### NewActor

```go
func NewActor(config *Config) *Actor
```

NewActor creates a new configuration actor.


### Methods

#### Actor.Get

```go
func () Get() *Config
```

Get returns the current configuration.

#### Actor.RegisterWatcher

```go
func () RegisterWatcher(watcher Watcher)
```

RegisterWatcher registers a watcher that receives config change callbacks.

#### Actor.Shutdown

```go
func () Shutdown()
```

Shutdown gracefully shuts down the actor and closes all subscriptions.

#### Actor.Subscribe

```go
func () Subscribe() <-chan *Config
```

Subscribe creates a new subscription for configuration changes.

#### Actor.Unsubscribe

```go
func () Unsubscribe(ch <-chan *Config)
```

Unsubscribe removes a subscription.

#### Actor.Update

```go
func () Update(config *Config) error
```

Update updates the configuration and notifies all watchers.


## type AdapterConfig

```go
type AdapterConfig struct {
	// Adapter name
	Name string `toml:"name" json:"name"`

	// Adapter version requirements
	Version string `toml:"version" json:"version"`

	// Adapter-specific configuration (may contain sensitive data)
	Config map[string]interface{} `toml:"config" json:"config"`

	// Sensitive field names (to be redacted in logs/output)
	SensitiveFields []string `toml:"sensitive_fields" json:"sensitive_fields"`
}
```

AdapterConfig represents adapter-specific configuration with sensitive field handling.

### Methods

#### AdapterConfig.GetSensitiveValue

```go
func () GetSensitiveValue(key string) (interface{}, error)
```

GetSensitiveValue returns a sensitive field value or error if field is not found.

#### AdapterConfig.Redact

```go
func () Redact() *AdapterConfig
```

Redact returns a copy of the config with sensitive fields redacted.

#### AdapterConfig.isSensitiveField

```go
func () isSensitiveField(field string) bool
```

isSensitiveField checks if a field name is in the sensitive fields list.


## type AgentsConfig

```go
type AgentsConfig struct {
	// Default worktree strategy
	DefaultStrategy string `toml:"default_strategy"`

	// Auto-cleanup worktrees on remove
	AutoCleanup bool `toml:"auto_cleanup"`

	// Maximum concurrent agents
	MaxConcurrent int `toml:"max_concurrent"`
}
```

AgentsConfig contains agent management settings.

## type Config

```go
type Config struct {
	// Core configuration
	Core CoreConfig `toml:"core"`

	// Agent management configuration
	Agents AgentsConfig `toml:"agents"`

	// Remote configuration
	Remote RemoteConfig `toml:"remote"`

	// Event system configuration
	Events EventsConfig `toml:"events"`

	// Process tracking configuration
	Process ProcessConfig `toml:"process"`

	// PTY monitoring configuration
	Monitor MonitorConfig `toml:"monitor"`

	// Adapter configuration (opaque per-adapter blocks)
	Adapters map[string]map[string]interface{} `toml:"adapters"`

	// OpenTelemetry configuration
	OTel OTelConfig `toml:"otel"`

	// Inference configuration
	Inference InferenceConfig `toml:"inference"`
}
```

Config represents the complete amux configuration.

### Functions returning Config

#### Load

```go
func Load() (*Config, error)
```

Load loads configuration from multiple sources in priority order.


### Methods

#### Config.GetAdapterConfigs

```go
func () GetAdapterConfigs() (map[string]*AdapterConfig, error)
```

GetAdapterConfigs extracts and validates adapter configurations from the main config.


## type CoreConfig

```go
type CoreConfig struct {
	// Data directory for persistent state
	DataDir string `toml:"data_dir"`

	// Runtime directory for sockets and temporary files
	RuntimeDir string `toml:"runtime_dir"`

	// Log level (trace, debug, info, warn, error)
	LogLevel string `toml:"log_level"`

	// Whether to run in debug mode
	Debug bool `toml:"debug"`
}
```

CoreConfig contains core daemon configuration.

## type EventsConfig

```go
type EventsConfig struct {
	// Enable event subscriptions
	Subscriptions SubscriptionsConfig `toml:"subscriptions"`
}
```

EventsConfig contains event system settings.

## type InferenceConfig

```go
type InferenceConfig struct {
	// Enable local inference
	Enabled bool `toml:"enabled"`

	// Inference engine type
	Engine string `toml:"engine"`

	// Model configuration
	Models map[string]ModelConfig `toml:"models"`
}
```

InferenceConfig contains local inference settings.

## type JetStreamConfig

```go
type JetStreamConfig struct {
	// KV bucket name
	BucketName string `toml:"bucket_name"`

	// Stream name for events
	StreamName string `toml:"stream_name"`

	// Domain for JetStream
	Domain string `toml:"domain"`
}
```

JetStreamConfig contains JetStream-specific settings.

## type ModelConfig

```go
type ModelConfig struct {
	// Model type (embedding, generation, etc.)
	Type string `toml:"type"`

	// Model path or identifier
	Path string `toml:"path"`

	// Model parameters
	Parameters map[string]interface{} `toml:"parameters"`
}
```

ModelConfig contains model-specific settings.

## type MonitorConfig

```go
type MonitorConfig struct {
	// Activity timeout
	ActivityTimeout time.Duration `toml:"activity_timeout"`

	// Pattern matching timeout
	PatternTimeout time.Duration `toml:"pattern_timeout"`

	// Enable TUI decoding
	TUIEnabled bool `toml:"tui_enabled"`
}
```

MonitorConfig contains PTY monitoring settings.

## type OTelConfig

```go
type OTelConfig struct {
	// Enable OpenTelemetry
	Enabled bool `toml:"enabled"`

	// Service name
	ServiceName string `toml:"service_name"`

	// Service version
	ServiceVersion string `toml:"service_version"`

	// Exporter configuration
	Exporter OTelExporterConfig `toml:"exporter"`
}
```

OTelConfig contains OpenTelemetry settings.

## type OTelExporterConfig

```go
type OTelExporterConfig struct {
	// Exporter type (stdout, otlp, etc.)
	Type string `toml:"type"`

	// Endpoint for OTLP exporter
	Endpoint string `toml:"endpoint"`

	// Headers for OTLP exporter
	Headers map[string]string `toml:"headers"`
}
```

OTelExporterConfig contains OTel exporter settings.

## type ProcessConfig

```go
type ProcessConfig struct {
	// Enable process interception
	InterceptionEnabled bool `toml:"interception_enabled"`

	// Hook library path
	HookPath string `toml:"hook_path"`

	// Fallback polling interval
	PollingInterval time.Duration `toml:"polling_interval"`

	// I/O capture mode
	CaptureMode string `toml:"capture_mode"`

	// Batch size for events
	BatchSize int `toml:"batch_size"`

	// Batch timeout
	BatchTimeout time.Duration `toml:"batch_timeout"`
}
```

ProcessConfig contains process tracking settings.

## type RemoteConfig

```go
type RemoteConfig struct {
	// NATS server URL
	ServerURL string `toml:"server_url"`

	// Credentials file path
	CredsPath string `toml:"creds_path"`

	// Subject prefix for all subjects
	SubjectPrefix string `toml:"subject_prefix"`

	// Request timeout
	RequestTimeout time.Duration `toml:"request_timeout"`

	// Buffer size for PTY output replay
	BufferSize int `toml:"buffer_size"`

	// JetStream configuration
	JetStream JetStreamConfig `toml:"jetstream"`
}
```

RemoteConfig contains remote orchestration settings.

## type SubscriptionsConfig

```go
type SubscriptionsConfig struct {
	// Enable MCP server for subscriptions
	Enabled bool `toml:"enabled"`

	// Socket path for MCP server
	SocketPath string `toml:"socket_path"`

	// Maximum concurrent subscribers
	MaxConcurrent int `toml:"max_concurrent"`
}
```

SubscriptionsConfig contains subscription settings.

## type Watcher

```go
type Watcher interface {
	OnConfigChange(oldConfig, newConfig *Config) error
}
```

Watcher represents a configuration change subscriber.


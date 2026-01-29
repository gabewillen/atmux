# package config

`import "github.com/stateforward/amux/internal/config"`

Package config implements a configuration actor with live updates and subscriptions

Package config implements configuration management with hierarchy, environment mapping,
and parsing conventions as specified in the amux specification.

- `func DecodeSensitiveData(encoded string) ([]byte, error)` — DecodeSensitiveData decodes base64 encoded sensitive data
- `func EncodeSensitiveData(data []byte) string` — EncodeSensitiveData base64 encodes sensitive data for secure transmission/storage
- `func SecureCompare(a, b string) bool` — SecureCompare compares sensitive values in a timing-attack resistant way
- `func applyEnvOverrides(config *Config)` — applyEnvOverrides applies environment variable overrides to the config
- `func configsEqual(a, b *Config) bool` — configsEqual compares two config objects for equality
- `func expandHomeDir(path string) string` — expandHomeDir expands the ~ symbol to the user's home directory
- `func getEnvWithPrefix(key string) string` — getEnvWithPrefix gets an environment variable with the AMUX__ prefix The key should already include the full variable name (e.g., "AMUX_CORE_REPO_ROOT")
- `func isSensitiveField(fieldName string) bool` — isSensitiveField checks if a field name indicates sensitive data
- `func mergeConfig(dst *Config, src *Config)` — mergeConfig merges source config into destination config
- `func parseBytes(s string) (int64, error)` — parseBytes parses a string representation of bytes (e.g., "10MB", "1GB") into an int64
- `type Actor` — Actor manages configuration with live updates and subscriptions
- `type Config` — Config represents the main configuration structure
- `type CoreConfig` — CoreConfig holds core application settings
- `type LoggingConfig` — LoggingConfig holds logging settings
- `type ManagerConfig` — ManagerConfig holds manager-specific settings
- `type NATSConfig` — NATSConfig holds NATS connection and server settings
- `type RemoteConfig` — RemoteConfig holds remote orchestration settings
- `type ServerConfig` — ServerConfig holds server/daemon settings
- `type TelemetryConfig` — TelemetryConfig holds OpenTelemetry settings

### Functions

#### DecodeSensitiveData

```go
func DecodeSensitiveData(encoded string) ([]byte, error)
```

DecodeSensitiveData decodes base64 encoded sensitive data

#### EncodeSensitiveData

```go
func EncodeSensitiveData(data []byte) string
```

EncodeSensitiveData base64 encodes sensitive data for secure transmission/storage

#### SecureCompare

```go
func SecureCompare(a, b string) bool
```

SecureCompare compares sensitive values in a timing-attack resistant way

#### applyEnvOverrides

```go
func applyEnvOverrides(config *Config)
```

applyEnvOverrides applies environment variable overrides to the config

#### configsEqual

```go
func configsEqual(a, b *Config) bool
```

configsEqual compares two config objects for equality

#### expandHomeDir

```go
func expandHomeDir(path string) string
```

expandHomeDir expands the ~ symbol to the user's home directory

#### getEnvWithPrefix

```go
func getEnvWithPrefix(key string) string
```

getEnvWithPrefix gets an environment variable with the AMUX__ prefix
The key should already include the full variable name (e.g., "AMUX_CORE_REPO_ROOT")

#### isSensitiveField

```go
func isSensitiveField(fieldName string) bool
```

isSensitiveField checks if a field name indicates sensitive data

#### mergeConfig

```go
func mergeConfig(dst *Config, src *Config)
```

mergeConfig merges source config into destination config

#### parseBytes

```go
func parseBytes(s string) (int64, error)
```

parseBytes parses a string representation of bytes (e.g., "10MB", "1GB") into an int64


## type Actor

```go
type Actor struct {
	mu          sync.RWMutex
	config      *Config
	subscribers map[string]chan Config
	nextID      int
	ctx         context.Context
	cancel      context.CancelFunc
}
```

Actor manages configuration with live updates and subscriptions

### Functions returning Actor

#### NewActor

```go
func NewActor(initialConfig *Config) *Actor
```

NewActor creates a new configuration actor


### Methods

#### Actor.Close

```go
func () Close()
```

Close shuts down the actor and all subscriptions

#### Actor.Get

```go
func () Get() *Config
```

Get returns the current configuration

#### Actor.LoadAndWatch

```go
func () LoadAndWatch(ctx context.Context, configPath string) error
```

LoadAndWatch loads configuration from a file and watches for changes

#### Actor.Subscribe

```go
func () Subscribe() (<-chan Config, func())
```

Subscribe registers a subscriber to receive configuration updates

#### Actor.Update

```go
func () Update(newConfig *Config) error
```

Update updates the configuration and notifies subscribers


## type Config

```go
type Config struct {
	// Core settings
	Core CoreConfig `toml:"core" json:"core"`

	// Server settings for daemon
	Server ServerConfig `toml:"server" json:"server"`

	// Logging settings
	Logging LoggingConfig `toml:"logging" json:"logging"`

	// Telemetry settings
	Telemetry TelemetryConfig `toml:"telemetry" json:"telemetry"`

	// Remote settings
	Remote RemoteConfig `toml:"remote" json:"remote"`

	// Adapter-specific configurations (opaque to core)
	Adapters map[string]map[string]interface{} `toml:"adapters" json:"adapters"`
}
```

Config represents the main configuration structure

### Functions returning Config

#### LoadConfig

```go
func LoadConfig(configPath string) (*Config, error)
```

LoadConfig loads configuration from multiple sources with precedence:
1. Built-in defaults
2. Config file
3. Environment variables

#### getDefaultConfig

```go
func getDefaultConfig() Config
```

getDefaultConfig returns the default configuration values

#### loadConfigFromFile

```go
func loadConfigFromFile(path string) (*Config, error)
```

loadConfigFromFile loads configuration from a TOML file


### Methods

#### Config.GetAdapterConfig

```go
func () GetAdapterConfig(adapterName string) map[string]interface{}
```

GetAdapterConfig retrieves the configuration for a specific adapter

#### Config.RedactSensitiveFields

```go
func () RedactSensitiveFields() *Config
```

RedactSensitiveFields removes sensitive information from the config for logging/debugging

#### Config.Validate

```go
func () Validate() error
```

Validate validates the configuration

#### Config.ValidateAdapterConfig

```go
func () ValidateAdapterConfig(adapterName string, requiredFields []string) error
```

ValidateAdapterConfig validates the configuration for a specific adapter


## type CoreConfig

```go
type CoreConfig struct {
	RepoRoot string `toml:"repo_root" json:"repo_root"`
	Debug    bool   `toml:"debug" json:"debug"`
}
```

CoreConfig holds core application settings

## type LoggingConfig

```go
type LoggingConfig struct {
	Level  string `toml:"level" json:"level"`
	Format string `toml:"format" json:"format"`
	File   string `toml:"file" json:"file"`
}
```

LoggingConfig holds logging settings

## type ManagerConfig

```go
type ManagerConfig struct {
	Enabled bool `toml:"enabled" json:"enabled"` // Whether to run local supervisor loop
}
```

ManagerConfig holds manager-specific settings

## type NATSConfig

```go
type NATSConfig struct {
	URL           string `toml:"url" json:"url"`                       // NATS server URL
	CredsPath     string `toml:"creds_path" json:"creds_path"`         // Path to NATS credential file
	SubjectPrefix string `toml:"subject_prefix" json:"subject_prefix"` // Root subject namespace for all amux traffic
	KVBucket      string `toml:"kv_bucket" json:"kv_bucket"`           // JetStream KV bucket for remote state
	StreamEvents  string `toml:"stream_events" json:"stream_events"`   // JetStream stream for EventMessage envelopes
	StreamPTY     string `toml:"stream_pty" json:"stream_pty"`         // JetStream stream for PTY byte chunks
}
```

NATSConfig holds NATS connection and server settings

## type RemoteConfig

```go
type RemoteConfig struct {
	Enabled        bool          `toml:"enabled" json:"enabled"`
	Transport      string        `toml:"transport" json:"transport"`             // nats or ssh_yamux
	RequestTimeout time.Duration `toml:"request_timeout" json:"request_timeout"` // Timeout for NATS request-reply control operations
	BufferSize     int64         `toml:"buffer_size" json:"buffer_size"`         // Size of replay buffer in bytes

	// NATS-specific settings
	NATS NATSConfig `toml:"nats" json:"nats"`

	// Manager-specific settings
	Manager ManagerConfig `toml:"manager" json:"manager"`
}
```

RemoteConfig holds remote orchestration settings

## type ServerConfig

```go
type ServerConfig struct {
	SocketPath string        `toml:"socket_path" json:"socket_path"`
	RPCTimeout time.Duration `toml:"rpc_timeout" json:"rpc_timeout"`
}
```

ServerConfig holds server/daemon settings

## type TelemetryConfig

```go
type TelemetryConfig struct {
	Enabled     bool   `toml:"enabled" json:"enabled"`
	Endpoint    string `toml:"endpoint" json:"endpoint"`
	ServiceName string `toml:"service_name" json:"service_name"`
}
```

TelemetryConfig holds OpenTelemetry settings


# package config

`import "github.com/copilot-claude-sonnet-4/amux/internal/config"`

Package config provides configuration management with TOML format support.
This package handles the configuration hierarchy: built-in < adapter < user < project < env
with env vars using AMUX__ prefix. Adapter configs are treated as opaque.

- `ErrConfigNotFound, ErrInvalidConfig, ErrLoadFailed` — Common sentinel errors for configuration operations.
- `func loadFromEnv(config *Config) error` — loadFromEnv applies environment variable overrides with AMUX__ prefix.
- `func loadFromFiles(config *Config) error` — loadFromFiles loads configuration from TOML files.
- `type Config` — Config represents the amux configuration structure.

### Variables

#### ErrConfigNotFound, ErrInvalidConfig, ErrLoadFailed

```go
var (
	// ErrConfigNotFound indicates a configuration file was not found.
	ErrConfigNotFound = errors.New("config not found")

	// ErrInvalidConfig indicates invalid configuration data.
	ErrInvalidConfig = errors.New("invalid config")

	// ErrLoadFailed indicates configuration loading failed.
	ErrLoadFailed = errors.New("config load failed")
)
```

Common sentinel errors for configuration operations.


### Functions

#### loadFromEnv

```go
func loadFromEnv(config *Config) error
```

loadFromEnv applies environment variable overrides with AMUX__ prefix.

#### loadFromFiles

```go
func loadFromFiles(config *Config) error
```

loadFromFiles loads configuration from TOML files.
Implementation deferred to Phase 0 completion.


## type Config

```go
type Config struct {
	// Core daemon settings
	Daemon struct {
		SocketPath string `toml:"socket_path"`
		LogLevel   string `toml:"log_level"`
	} `toml:"daemon"`

	// Agent configurations (opaque to core)
	Agents map[string]interface{} `toml:"agents"`

	// Remote settings
	Remote struct {
		Enabled bool   `toml:"enabled"`
		Hub     string `toml:"hub"`
	} `toml:"remote"`
}
```

Config represents the amux configuration structure.

### Functions returning Config

#### Load

```go
func Load() (*Config, error)
```

Load reads configuration from files and environment variables.
Implements the hierarchy: built-in < adapter < user < project < env


### Methods

#### Config.SaveToFile

```go
func () SaveToFile(path string) error
```

SaveToFile writes configuration to a TOML file.



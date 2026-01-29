# package telemetry

`import "github.com/copilot-claude-sonnet-4/amux/internal/telemetry"`

Package telemetry provides OpenTelemetry instrumentation for amux.
This package implements the observability requirements from spec §4.2.9,
providing traces, metrics, and logs for all core components.

The telemetry package supports configuration via environment variables
following the OTel specification, or via config file as defined in the spec.

Instrumentation covers:
- Agent lifecycle state transitions
- PTY monitor pattern matching
- Process tracker lifecycle and I/O events
- Adapter WASM call spans
- Remote agent NATS/SSH operations
- HSM event queue dispatch

- `ErrInitFailed, ErrExporterFailed, ErrShutdownFailed` — Common sentinel errors for telemetry operations.
- `func getEnvBool(key string, defaultValue bool) bool`
- `func getEnvFloat(key string, defaultValue float64) float64`
- `func getEnvString(key, defaultValue string) string` — Helper functions for environment variable parsing
- `type Config` — Config represents telemetry configuration matching spec §4.2.9.2.
- `type ExporterConfig` — ExporterConfig configures OpenTelemetry exporters.
- `type LogsConfig` — LogsConfig configures log collection.
- `type MetricsConfig` — MetricsConfig configures metrics collection.
- `type Provider` — Provider manages OpenTelemetry providers and exporters.
- `type TracesConfig` — TracesConfig configures trace collection.

### Variables

#### ErrInitFailed, ErrExporterFailed, ErrShutdownFailed

```go
var (
	// ErrInitFailed indicates telemetry initialization failed.
	ErrInitFailed = errors.New("telemetry initialization failed")

	// ErrExporterFailed indicates exporter setup failed.
	ErrExporterFailed = errors.New("exporter setup failed")

	// ErrShutdownFailed indicates telemetry shutdown failed.
	ErrShutdownFailed = errors.New("telemetry shutdown failed")
)
```

Common sentinel errors for telemetry operations.


### Functions

#### getEnvBool

```go
func getEnvBool(key string, defaultValue bool) bool
```

#### getEnvFloat

```go
func getEnvFloat(key string, defaultValue float64) float64
```

#### getEnvString

```go
func getEnvString(key, defaultValue string) string
```

Helper functions for environment variable parsing


## type Config

```go
type Config struct {
	Enabled     bool           `toml:"enabled"`
	ServiceName string         `toml:"service_name"`
	Exporter    ExporterConfig `toml:"exporter"`
	Traces      TracesConfig   `toml:"traces"`
	Metrics     MetricsConfig  `toml:"metrics"`
	Logs        LogsConfig     `toml:"logs"`
}
```

Config represents telemetry configuration matching spec §4.2.9.2.

### Functions returning Config

#### DefaultConfig

```go
func DefaultConfig() Config
```

DefaultConfig returns a default telemetry configuration.
It checks environment variables first, then applies defaults.


## type ExporterConfig

```go
type ExporterConfig struct {
	Endpoint string `toml:"endpoint"`
	Protocol string `toml:"protocol"` // grpc, http/protobuf, http/json
}
```

ExporterConfig configures OpenTelemetry exporters.

## type LogsConfig

```go
type LogsConfig struct {
	Enabled bool `toml:"enabled"`
}
```

LogsConfig configures log collection.

## type MetricsConfig

```go
type MetricsConfig struct {
	Enabled bool `toml:"enabled"`
}
```

MetricsConfig configures metrics collection.

## type Provider

```go
type Provider struct {
	traceProvider  *traceSDK.TracerProvider
	metricProvider *metricSDK.MeterProvider
	config         Config
}
```

Provider manages OpenTelemetry providers and exporters.

### Functions returning Provider

#### NewProvider

```go
func NewProvider(config Config) (*Provider, error)
```

NewProvider creates a new telemetry provider with the given configuration.
This function implements the initialization requirements from spec §4.2.9.


### Methods

#### Provider.Meter

```go
func () Meter(name string) metric.Meter
```

Meter returns a meter for the given name.

#### Provider.Shutdown

```go
func () Shutdown(ctx context.Context) error
```

Shutdown gracefully shuts down the telemetry providers.

#### Provider.Tracer

```go
func () Tracer(name string) trace.Tracer
```

Tracer returns a tracer for the given name.

#### Provider.initMetricProvider

```go
func () initMetricProvider(res *resource.Resource) (*metricSDK.MeterProvider, error)
```

initMetricProvider creates and configures the metric provider.

#### Provider.initTraceProvider

```go
func () initTraceProvider(res *resource.Resource) (*traceSDK.TracerProvider, error)
```

initTraceProvider creates and configures the trace provider.


## type TracesConfig

```go
type TracesConfig struct {
	Enabled    bool    `toml:"enabled"`
	Sampler    string  `toml:"sampler"`
	SamplerArg float64 `toml:"sampler_arg"`
}
```

TracesConfig configures trace collection.


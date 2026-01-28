# package telemetry

`import "github.com/agentflare-ai/amux/internal/telemetry"`

Package telemetry provides OpenTelemetry instrumentation for amux.

This package configures and exposes tracing, metrics, and logging
using OpenTelemetry. Spans follow the naming convention:
{component}.{operation}

See spec §4.2.9 for full observability requirements.

- `ServiceName` — ServiceName is the default service name for amux.
- `SpecVersion` — SpecVersion is the spec version for telemetry resource attributes.
- `func Meter() metric.Meter` — Meter returns the global meter.
- `func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span)` — StartSpan starts a new span using the global tracer.
- `func Tracer() trace.Tracer` — Tracer returns the global tracer.
- `func getSampler(name string, arg float64) sdktrace.Sampler`
- `func setGlobalProvider(p *Provider)`
- `type Config` — Config holds telemetry configuration.
- `type Provider` — Provider holds the telemetry providers.

### Constants

#### ServiceName

```go
const ServiceName = "amux"
```

ServiceName is the default service name for amux.

#### SpecVersion

```go
const SpecVersion = "v1.22"
```

SpecVersion is the spec version for telemetry resource attributes.


### Functions

#### Meter

```go
func Meter() metric.Meter
```

Meter returns the global meter.

#### StartSpan

```go
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
```

StartSpan starts a new span using the global tracer.

#### Tracer

```go
func Tracer() trace.Tracer
```

Tracer returns the global tracer.

#### getSampler

```go
func getSampler(name string, arg float64) sdktrace.Sampler
```

#### setGlobalProvider

```go
func setGlobalProvider(p *Provider)
```


## type Config

```go
type Config struct {
	// Enabled controls whether telemetry is active.
	Enabled bool

	// ServiceName is the service name to report.
	ServiceName string

	// Endpoint is the OTLP exporter endpoint.
	Endpoint string

	// Protocol is the OTLP protocol (grpc, http/protobuf, http/json).
	Protocol string

	// TracesEnabled controls trace export.
	TracesEnabled bool

	// TraceSampler is the trace sampler name.
	TraceSampler string

	// TraceSamplerArg is the sampler argument (e.g., ratio).
	TraceSamplerArg float64

	// MetricsEnabled controls metrics export.
	MetricsEnabled bool

	// LogsEnabled controls logs export.
	LogsEnabled bool
}
```

Config holds telemetry configuration.

### Functions returning Config

#### ConfigFromEnv

```go
func ConfigFromEnv() Config
```

ConfigFromEnv creates a Config from environment variables.

#### DefaultConfig

```go
func DefaultConfig() Config
```

DefaultConfig returns the default telemetry configuration.


## type Provider

```go
type Provider struct {
	config         Config
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	tracer         trace.Tracer
	meter          metric.Meter

	// Metrics
	agentsTotal           metric.Int64Counter
	agentsActive          metric.Int64UpDownCounter
	eventsTotal           metric.Int64Counter
	processesTotal        metric.Int64Counter
	processesRunning      metric.Int64UpDownCounter
	ptyOutputBytes        metric.Int64Histogram
	adapterCallDuration   metric.Float64Histogram
	eventDispatchDuration metric.Float64Histogram

	shutdown func(context.Context) error
}
```

Provider holds the telemetry providers.

### Variables

#### globalProvider, providerMu

```go
var (
	globalProvider *Provider
	providerMu     sync.RWMutex
)
```


### Functions returning Provider

#### Global

```go
func Global() *Provider
```

Global returns the global telemetry provider.

#### Init

```go
func Init(ctx context.Context, cfg Config) (*Provider, error)
```

Init initializes the global telemetry provider.


### Methods

#### Provider.Meter

```go
func () Meter() metric.Meter
```

Meter returns the meter.

#### Provider.RecordAdapterCall

```go
func () RecordAdapterCall(ctx context.Context, adapter, function string, durationSeconds float64)
```

RecordAdapterCall records an adapter call duration.

#### Provider.RecordAgentActive

```go
func () RecordAgentActive(ctx context.Context, adapter, presence string, delta int64)
```

RecordAgentActive records agent active state change.

#### Provider.RecordAgentAdded

```go
func () RecordAgentAdded(ctx context.Context, adapter, locationType string)
```

RecordAgentAdded records an agent addition.

#### Provider.RecordEvent

```go
func () RecordEvent(ctx context.Context, eventType string)
```

RecordEvent records an event dispatch.

#### Provider.RecordPTYOutput

```go
func () RecordPTYOutput(ctx context.Context, agentID string, bytes int64)
```

RecordPTYOutput records PTY output bytes.

#### Provider.Shutdown

```go
func () Shutdown(ctx context.Context) error
```

Shutdown shuts down the telemetry provider.

#### Provider.Tracer

```go
func () Tracer() trace.Tracer
```

Tracer returns the tracer.

#### Provider.initMetrics

```go
func () initMetrics() error
```



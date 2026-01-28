# package telemetry

`import "github.com/agentflare-ai/amux/internal/telemetry"`

Package telemetry provides OpenTelemetry scaffolding for amux.

It configures tracing and metrics providers based on config and environment
variables and exposes helpers for baseline metrics and span naming.

- `func CollectMetrics(ctx context.Context, reader *sdkmetric.ManualReader) (metricdata.ResourceMetrics, error)` — CollectMetrics uses a manual reader to collect metrics into ResourceMetrics.
- `func EmitBaselineMetrics(ctx context.Context) error` — EmitBaselineMetrics records baseline metrics used in conformance testing.
- `func applyEnv(cfg config.TelemetryConfig, env map[string]string) config.TelemetryConfig`
- `func selectSampler(cfg config.TelemetryTracesConfig) sdktrace.Sampler`
- `type Option` — Option configures telemetry options.
- `type Options` — Options configures telemetry setup.
- `type ShutdownFunc` — ShutdownFunc shuts down telemetry providers.

### Functions

#### CollectMetrics

```go
func CollectMetrics(ctx context.Context, reader *sdkmetric.ManualReader) (metricdata.ResourceMetrics, error)
```

CollectMetrics uses a manual reader to collect metrics into ResourceMetrics.

#### EmitBaselineMetrics

```go
func EmitBaselineMetrics(ctx context.Context) error
```

EmitBaselineMetrics records baseline metrics used in conformance testing.

#### applyEnv

```go
func applyEnv(cfg config.TelemetryConfig, env map[string]string) config.TelemetryConfig
```

#### selectSampler

```go
func selectSampler(cfg config.TelemetryTracesConfig) sdktrace.Sampler
```


## type Option

```go
type Option func(*Options)
```

Option configures telemetry options.

### Functions returning Option

#### WithLogger

```go
func WithLogger(logger *log.Logger) Option
```

WithLogger supplies a logger for telemetry setup.

#### WithMetricReader

```go
func WithMetricReader(reader sdkmetric.Reader) Option
```

WithMetricReader supplies a metrics reader.

#### WithTraceExporter

```go
func WithTraceExporter(exporter sdktrace.SpanExporter) Option
```

WithTraceExporter supplies a trace exporter.


## type Options

```go
type Options struct {
	TraceExporter sdktrace.SpanExporter
	MetricReader  sdkmetric.Reader
	Logger        *log.Logger
}
```

Options configures telemetry setup.

## type ShutdownFunc

```go
type ShutdownFunc func(context.Context) error
```

ShutdownFunc shuts down telemetry providers.

### Functions returning ShutdownFunc

#### Setup

```go
func Setup(ctx context.Context, cfg config.TelemetryConfig, env map[string]string, optFns ...Option) (ShutdownFunc, error)
```

Setup configures OpenTelemetry providers from config and environment.



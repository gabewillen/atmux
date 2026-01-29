# package telemetry

`import "github.com/stateforward/amux/internal/telemetry"`

Package telemetry provides OpenTelemetry instrumentation for amux per spec §4.2.9.

This package provides scaffolding for traces, metrics, and logs following
the OpenTelemetry specification.

- `func Init(ctx context.Context) (func(context.Context) error, error)` — Init initializes OpenTelemetry with a basic tracer and meter provider.
- `func Meter() metric.Meter` — Meter returns the amux meter for metrics.
- `func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span)` — StartSpan starts a new span with the given name.
- `func Tracer() trace.Tracer` — Tracer returns the amux tracer.

### Functions

#### Init

```go
func Init(ctx context.Context) (func(context.Context) error, error)
```

Init initializes OpenTelemetry with a basic tracer and meter provider.

It configures a batch span processor with an in-memory exporter driven by
the global OTEL_* environment variables via the default SDK behavior and
sets the global tracer and meter providers. The returned shutdown function
MUST be called on process exit to flush spans.

#### Meter

```go
func Meter() metric.Meter
```

Meter returns the amux meter for metrics.

#### StartSpan

```go
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
```

StartSpan starts a new span with the given name.

#### Tracer

```go
func Tracer() trace.Tracer
```

Tracer returns the amux tracer.



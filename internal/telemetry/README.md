# package telemetry

`import "github.com/stateforward/amux/internal/telemetry"`

Package telemetry provides OpenTelemetry instrumentation for amux per spec §4.2.9.

This package provides scaffolding for traces, metrics, and logs following
the OpenTelemetry specification.

- `func Init(ctx context.Context) (func(context.Context) error, error)` — Init initializes OpenTelemetry with the provided configuration.
- `func Meter() metric.Meter` — Meter returns the amux meter for metrics.
- `func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span)` — StartSpan starts a new span with the given name.
- `func Tracer() trace.Tracer` — Tracer returns the amux tracer.

### Functions

#### Init

```go
func Init(ctx context.Context) (func(context.Context) error, error)
```

Init initializes OpenTelemetry with the provided configuration.
This is a placeholder for Phase 0. Full implementation will be in later phases.

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



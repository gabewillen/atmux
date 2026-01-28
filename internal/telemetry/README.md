# package telemetry

`import "github.com/agentflare-ai/amux/internal/telemetry"`

Package telemetry provides OpenTelemetry instrumentation for amux.

- `func Init(ctx context.Context, cfg *config.Config) (func() error, error)` — Init initializes OpenTelemetry based on configuration.
- `func Tracer(component string) trace.Tracer` — Tracer returns a tracer for the given component.

### Functions

#### Init

```go
func Init(ctx context.Context, cfg *config.Config) (func() error, error)
```

Init initializes OpenTelemetry based on configuration.

#### Tracer

```go
func Tracer(component string) trace.Tracer
```

Tracer returns a tracer for the given component.



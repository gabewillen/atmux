# package telemetry

`import "github.com/agentflare-ai/amux/internal/telemetry"`

Package telemetry configures OpenTelemetry for traces, metrics, and logs.

- `type ShutdownFunc` — ShutdownFunc is a function that shuts down the telemetry providers.

## type ShutdownFunc

```go
type ShutdownFunc func(context.Context) error
```

ShutdownFunc is a function that shuts down the telemetry providers.

### Functions returning ShutdownFunc

#### Setup

```go
func Setup(ctx context.Context, cfg config.TelemetryConfig) (ShutdownFunc, error)
```

Setup initializes the OpenTelemetry globals based on the configuration.



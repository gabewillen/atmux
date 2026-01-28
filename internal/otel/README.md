# package otel

`import "github.com/stateforward/amux/internal/otel"`

Package otel implements OpenTelemetry scaffolding for the amux project

- `type Provider` — Provider holds the OpenTelemetry trace and metric providers

## type Provider

```go
type Provider struct {
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	serviceName    string
}
```

Provider holds the OpenTelemetry trace and metric providers

### Functions returning Provider

#### NewProvider

```go
func NewProvider(serviceName string, resourceAttrs ...attribute.KeyValue) (*Provider, error)
```

NewProvider creates a new OpenTelemetry provider with the given service name


### Methods

#### Provider.GetMeter

```go
func () GetMeter(name string) metric.Meter
```

GetMeter returns a meter with the given name

#### Provider.GetTracer

```go
func () GetTracer(name string) trace.Tracer
```

GetTracer returns a tracer with the given name

#### Provider.MeterProvider

```go
func () MeterProvider() *sdkmetric.MeterProvider
```

MeterProvider returns the metric provider

#### Provider.Shutdown

```go
func () Shutdown(ctx context.Context) error
```

Shutdown shuts down the OpenTelemetry providers

#### Provider.TracerProvider

```go
func () TracerProvider() *sdktrace.TracerProvider
```

TracerProvider returns the trace provider



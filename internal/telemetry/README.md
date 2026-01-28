# package telemetry

`import "github.com/agentflare-ai/amux/internal/telemetry"`

Package telemetry provides OpenTelemetry instrumentation for amux.

- `func AddCounter(ctx context.Context, counter metric.Int64Counter, value int64, attrs ...attribute.KeyValue)` — AddCounter adds to a counter metric.
- `func AddUpDownCounter(ctx context.Context, counter metric.Int64UpDownCounter, value int64, attrs ...attribute.KeyValue)` — AddUpDownCounter adds to an up-down counter.
- `func Counter(name string, opts ...metric.Int64CounterOption) (metric.Int64Counter, error)` — Counter creates a new counter metric.
- `func GetServiceInfo() (name, version string)` — GetServiceInfo returns service information for resource attributes.
- `func Histogram(name string, opts ...metric.Float64HistogramOption) (metric.Float64Histogram, error)` — Histogram creates a new histogram metric.
- `func Init(otelConfig *config.OTelConfig) error` — Init initializes global OpenTelemetry.
- `func InitFromConfig(cfg *config.Config) error` — InitFromConfig initializes telemetry from a full config.
- `func LogAttribute(key, value string) attribute.KeyValue` — LogAttribute creates a string attribute for logging.
- `func LogErrorAttribute(err error) attribute.KeyValue` — LogErrorAttribute creates an error attribute for logging.
- `func Meter() metric.Meter` — Meter returns the global meter.
- `func RecordError(ctx context.Context, err error, opts ...trace.EventOption)` — RecordError records an error on the current span.
- `func RecordHistogram(ctx context.Context, histogram metric.Float64Histogram, value float64, attrs ...attribute.KeyValue)` — RecordHistogram records a histogram value.
- `func Shutdown(ctx context.Context) error` — Shutdown shuts down OpenTelemetry.
- `func SpanWithAttributes(ctx context.Context, attrs ...attribute.KeyValue)` — SpanWithAttributes adds attributes to the current span.
- `func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span)` — StartSpan starts a new span with the given name.
- `func Tracer() trace.Tracer` — Tracer returns the global tracer.
- `func UpDownCounter(name string, opts ...metric.Int64UpDownCounterOption) (metric.Int64UpDownCounter, error)` — UpDownCounter creates a new up-down counter metric.
- `func WithSpan(ctx context.Context, name string, fn func(context.Context) error) error` — WithSpan creates a span for a function call.
- `func createTraceExporter(config *config.OTelConfig) (sdktrace.SpanExporter, error)` — createTraceExporter creates a trace exporter based on configuration.
- `type Telemetry` — Telemetry provides OpenTelemetry instrumentation.

### Functions

#### AddCounter

```go
func AddCounter(ctx context.Context, counter metric.Int64Counter, value int64, attrs ...attribute.KeyValue)
```

AddCounter adds to a counter metric.

#### AddUpDownCounter

```go
func AddUpDownCounter(ctx context.Context, counter metric.Int64UpDownCounter, value int64, attrs ...attribute.KeyValue)
```

AddUpDownCounter adds to an up-down counter.

#### Counter

```go
func Counter(name string, opts ...metric.Int64CounterOption) (metric.Int64Counter, error)
```

Counter creates a new counter metric.

#### GetServiceInfo

```go
func GetServiceInfo() (name, version string)
```

GetServiceInfo returns service information for resource attributes.

#### Histogram

```go
func Histogram(name string, opts ...metric.Float64HistogramOption) (metric.Float64Histogram, error)
```

Histogram creates a new histogram metric.

#### Init

```go
func Init(otelConfig *config.OTelConfig) error
```

Init initializes global OpenTelemetry.

#### InitFromConfig

```go
func InitFromConfig(cfg *config.Config) error
```

InitFromConfig initializes telemetry from a full config.

#### LogAttribute

```go
func LogAttribute(key, value string) attribute.KeyValue
```

LogAttribute creates a string attribute for logging.

#### LogErrorAttribute

```go
func LogErrorAttribute(err error) attribute.KeyValue
```

LogErrorAttribute creates an error attribute for logging.

#### Meter

```go
func Meter() metric.Meter
```

Meter returns the global meter.

#### RecordError

```go
func RecordError(ctx context.Context, err error, opts ...trace.EventOption)
```

RecordError records an error on the current span.

#### RecordHistogram

```go
func RecordHistogram(ctx context.Context, histogram metric.Float64Histogram, value float64, attrs ...attribute.KeyValue)
```

RecordHistogram records a histogram value.

#### Shutdown

```go
func Shutdown(ctx context.Context) error
```

Shutdown shuts down OpenTelemetry.

#### SpanWithAttributes

```go
func SpanWithAttributes(ctx context.Context, attrs ...attribute.KeyValue)
```

SpanWithAttributes adds attributes to the current span.

#### StartSpan

```go
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
```

StartSpan starts a new span with the given name.

#### Tracer

```go
func Tracer() trace.Tracer
```

Tracer returns the global tracer.

#### UpDownCounter

```go
func UpDownCounter(name string, opts ...metric.Int64UpDownCounterOption) (metric.Int64UpDownCounter, error)
```

UpDownCounter creates a new up-down counter metric.

#### WithSpan

```go
func WithSpan(ctx context.Context, name string, fn func(context.Context) error) error
```

WithSpan creates a span for a function call.

#### createTraceExporter

```go
func createTraceExporter(config *config.OTelConfig) (sdktrace.SpanExporter, error)
```

createTraceExporter creates a trace exporter based on configuration.


## type Telemetry

```go
type Telemetry struct {
	config   *config.OTelConfig
	tracer   trace.Tracer
	shutdown func(context.Context) error
}
```

Telemetry provides OpenTelemetry instrumentation.

### Variables

#### globalTelemetry

```go
var globalTelemetry *Telemetry
```


### Functions returning Telemetry

#### Global

```go
func Global() *Telemetry
```

Global returns the global telemetry instance.



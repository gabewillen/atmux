// Package telemetry provides OpenTelemetry instrumentation for amux.
package telemetry

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/agentflare-ai/amux/internal/config"
	amuxerrors "github.com/agentflare-ai/amux/internal/errors"
)

// Telemetry provides OpenTelemetry instrumentation.
type Telemetry struct {
	config   *config.OTelConfig
	tracer   trace.Tracer
	shutdown func(context.Context) error
}

var globalTelemetry *Telemetry

// Init initializes global OpenTelemetry.
func Init(otelConfig *config.OTelConfig) error {
	if !otelConfig.Enabled {
		// Create no-op tracer and meter
		otel.SetTracerProvider(trace.NewNoopTracerProvider())
		return nil
	}

	// Create resource with service info
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", otelConfig.ServiceName),
			attribute.String("service.version", otelConfig.ServiceVersion),
		),
	)
	if err != nil {
		return amuxerrors.Wrap("creating resource", err)
	}

	// Create trace exporter based on config
	exporter, err := createTraceExporter(otelConfig)
	if err != nil {
		return amuxerrors.Wrap("creating trace exporter", err)
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set global providers
	otel.SetTracerProvider(tp)
	// TODO: implement meter provider when available

	// Set global propagator
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Create tracer
	tracer := tp.Tracer(otelConfig.ServiceName)
	// TODO: implement meter when available

	globalTelemetry = &Telemetry{
		config:   otelConfig,
		tracer:   tracer,
		shutdown: tp.Shutdown,
	}

	return nil
}

// createTraceExporter creates a trace exporter based on configuration.
func createTraceExporter(config *config.OTelConfig) (sdktrace.SpanExporter, error) {
	switch config.Exporter.Type {
	case "stdout":
		return stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
	case "otlp":
		// TODO: implement OTLP exporter
		return nil, amuxerrors.Wrap("creating OTLP exporter", amuxerrors.ErrNotReady)
	default:
		return nil, amuxerrors.Wrap("unknown exporter type", amuxerrors.ErrInvalidConfig)
	}
}

// Global returns the global telemetry instance.
func Global() *Telemetry {
	return globalTelemetry
}

// Tracer returns the global tracer.
func Tracer() trace.Tracer {
	if globalTelemetry != nil {
		return globalTelemetry.tracer
	}
	return otel.Tracer("")
}

// Meter returns the global meter.
func Meter() metric.Meter {
	// TODO: implement meter when available
	return otel.Meter("")
}

// Shutdown shuts down OpenTelemetry.
func Shutdown(ctx context.Context) error {
	if globalTelemetry != nil && globalTelemetry.shutdown != nil {
		return globalTelemetry.shutdown(ctx)
	}
	return nil
}

// StartSpan starts a new span with the given name.
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return Tracer().Start(ctx, name, opts...)
}

// SpanWithAttributes adds attributes to the current span.
func SpanWithAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attrs...)
	}
}

// RecordError records an error on the current span.
func RecordError(ctx context.Context, err error, opts ...trace.EventOption) {
	if err == nil {
		return
	}

	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.RecordError(err, opts...)
		span.SetAttributes(attribute.String("error.message", err.Error()))
	}
}

// WithSpan creates a span for a function call.
func WithSpan(ctx context.Context, name string, fn func(context.Context) error) error {
	ctx, span := StartSpan(ctx, name)
	defer span.End()

	if err := fn(ctx); err != nil {
		RecordError(ctx, err)
		return err
	}

	return nil
}

// Metric helpers

// Counter creates a new counter metric.
func Counter(name string, opts ...metric.Int64CounterOption) (metric.Int64Counter, error) {
	return Meter().Int64Counter(name, opts...)
}

// Histogram creates a new histogram metric.
func Histogram(name string, opts ...metric.Float64HistogramOption) (metric.Float64Histogram, error) {
	return Meter().Float64Histogram(name, opts...)
}

// UpDownCounter creates a new up-down counter metric.
func UpDownCounter(name string, opts ...metric.Int64UpDownCounterOption) (metric.Int64UpDownCounter, error) {
	return Meter().Int64UpDownCounter(name, opts...)
}

// AddCounter adds to a counter metric.
func AddCounter(ctx context.Context, counter metric.Int64Counter, value int64, attrs ...attribute.KeyValue) {
	counter.Add(ctx, value, metric.WithAttributes(attrs...))
}

// RecordHistogram records a histogram value.
func RecordHistogram(ctx context.Context, histogram metric.Float64Histogram, value float64, attrs ...attribute.KeyValue) {
	histogram.Record(ctx, value, metric.WithAttributes(attrs...))
}

// AddUpDownCounter adds to an up-down counter.
func AddUpDownCounter(ctx context.Context, counter metric.Int64UpDownCounter, value int64, attrs ...attribute.KeyValue) {
	counter.Add(ctx, value, metric.WithAttributes(attrs...))
}

// InitFromConfig initializes telemetry from a full config.
func InitFromConfig(cfg *config.Config) error {
	return Init(&cfg.OTel)
}

// GetServiceInfo returns service information for resource attributes.
func GetServiceInfo() (name, version string) {
	name = "amux"
	version = "unknown"

	if globalTelemetry != nil && globalTelemetry.config != nil {
		name = globalTelemetry.config.ServiceName
		version = globalTelemetry.config.ServiceVersion
	}

	// Override with environment variables if set
	if envName := os.Getenv("OTEL_SERVICE_NAME"); envName != "" {
		name = envName
	}
	if envVersion := os.Getenv("OTEL_SERVICE_VERSION"); envVersion != "" {
		version = envVersion
	}

	return name, version
}

// LogAttribute creates a string attribute for logging.
func LogAttribute(key, value string) attribute.KeyValue {
	return attribute.String(key, value)
}

// LogErrorAttribute creates an error attribute for logging.
func LogErrorAttribute(err error) attribute.KeyValue {
	if err == nil {
		return attribute.String("error", "")
	}
	return attribute.String("error", err.Error())
}

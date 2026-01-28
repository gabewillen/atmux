// Package telemetry provides OpenTelemetry instrumentation for amux.
//
// This package configures and exposes tracing, metrics, and logging
// using OpenTelemetry. Spans follow the naming convention:
// {component}.{operation}
//
// See spec §4.2.9 for full observability requirements.
package telemetry

import (
	"context"
	"os"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// ServiceName is the default service name for amux.
const ServiceName = "amux"

// SpecVersion is the spec version for telemetry resource attributes.
const SpecVersion = "v1.22"

// Config holds telemetry configuration.
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

// DefaultConfig returns the default telemetry configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:         false,
		ServiceName:     ServiceName,
		Endpoint:        "localhost:4317",
		Protocol:        "grpc",
		TracesEnabled:   true,
		TraceSampler:    "parentbased_traceidratio",
		TraceSamplerArg: 0.1,
		MetricsEnabled:  true,
		LogsEnabled:     true,
	}
}

// Provider holds the telemetry providers.
type Provider struct {
	config         Config
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	tracer         trace.Tracer
	meter          metric.Meter

	// Metrics
	agentsTotal        metric.Int64Counter
	agentsActive       metric.Int64UpDownCounter
	eventsTotal        metric.Int64Counter
	processesTotal     metric.Int64Counter
	processesRunning   metric.Int64UpDownCounter
	ptyOutputBytes     metric.Int64Histogram
	adapterCallDuration metric.Float64Histogram
	eventDispatchDuration metric.Float64Histogram

	shutdown func(context.Context) error
}

var (
	globalProvider *Provider
	providerMu     sync.RWMutex
)

// Init initializes the global telemetry provider.
func Init(ctx context.Context, cfg Config) (*Provider, error) {
	if !cfg.Enabled {
		// Return a noop provider
		p := &Provider{
			config: cfg,
			tracer: otel.Tracer(cfg.ServiceName),
			meter:  otel.Meter(cfg.ServiceName),
		}
		setGlobalProvider(p)
		return p, nil
	}

	// Create resource
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.ServiceName),
			attribute.String("amux.spec_version", SpecVersion),
		),
	)
	if err != nil {
		return nil, err
	}

	p := &Provider{config: cfg}
	var shutdowns []func(context.Context) error

	// Initialize trace provider
	if cfg.TracesEnabled {
		traceExporter, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(cfg.Endpoint),
			otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			return nil, err
		}
		shutdowns = append(shutdowns, traceExporter.Shutdown)

		sampler := getSampler(cfg.TraceSampler, cfg.TraceSamplerArg)
		p.tracerProvider = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(traceExporter),
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sampler),
		)
		otel.SetTracerProvider(p.tracerProvider)
		shutdowns = append(shutdowns, p.tracerProvider.Shutdown)
	}

	// Initialize meter provider
	if cfg.MetricsEnabled {
		metricExporter, err := otlpmetricgrpc.New(ctx,
			otlpmetricgrpc.WithEndpoint(cfg.Endpoint),
			otlpmetricgrpc.WithInsecure(),
		)
		if err != nil {
			return nil, err
		}
		shutdowns = append(shutdowns, metricExporter.Shutdown)

		p.meterProvider = sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
			sdkmetric.WithResource(res),
		)
		otel.SetMeterProvider(p.meterProvider)
		shutdowns = append(shutdowns, p.meterProvider.Shutdown)
	}

	// Set propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Get tracer and meter
	p.tracer = otel.Tracer(cfg.ServiceName)
	p.meter = otel.Meter(cfg.ServiceName)

	// Initialize metrics
	if err := p.initMetrics(); err != nil {
		return nil, err
	}

	// Create combined shutdown
	p.shutdown = func(ctx context.Context) error {
		var errs []error
		for i := len(shutdowns) - 1; i >= 0; i-- {
			if err := shutdowns[i](ctx); err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			return errs[0]
		}
		return nil
	}

	setGlobalProvider(p)
	return p, nil
}

func getSampler(name string, arg float64) sdktrace.Sampler {
	switch name {
	case "always_on":
		return sdktrace.AlwaysSample()
	case "always_off":
		return sdktrace.NeverSample()
	case "traceidratio":
		return sdktrace.TraceIDRatioBased(arg)
	case "parentbased_always_on":
		return sdktrace.ParentBased(sdktrace.AlwaysSample())
	case "parentbased_always_off":
		return sdktrace.ParentBased(sdktrace.NeverSample())
	case "parentbased_traceidratio":
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(arg))
	default:
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(arg))
	}
}

func (p *Provider) initMetrics() error {
	var err error

	// Counters
	p.agentsTotal, err = p.meter.Int64Counter("amux_agents_total",
		metric.WithDescription("Total agents by adapter and status"),
	)
	if err != nil {
		return err
	}

	p.eventsTotal, err = p.meter.Int64Counter("amux_events_total",
		metric.WithDescription("Total events by type"),
	)
	if err != nil {
		return err
	}

	p.processesTotal, err = p.meter.Int64Counter("amux_processes_total",
		metric.WithDescription("Total processes by agent and exit status"),
	)
	if err != nil {
		return err
	}

	// UpDownCounters (gauges)
	p.agentsActive, err = p.meter.Int64UpDownCounter("amux_agents_active",
		metric.WithDescription("Active agents by adapter and presence"),
	)
	if err != nil {
		return err
	}

	p.processesRunning, err = p.meter.Int64UpDownCounter("amux_processes_running",
		metric.WithDescription("Running processes by agent"),
	)
	if err != nil {
		return err
	}

	// Histograms
	p.ptyOutputBytes, err = p.meter.Int64Histogram("amux_pty_output_bytes",
		metric.WithDescription("PTY output bytes by agent"),
	)
	if err != nil {
		return err
	}

	p.adapterCallDuration, err = p.meter.Float64Histogram("amux_adapter_call_duration_seconds",
		metric.WithDescription("Adapter call duration by adapter and function"),
	)
	if err != nil {
		return err
	}

	p.eventDispatchDuration, err = p.meter.Float64Histogram("amux_event_dispatch_duration_seconds",
		metric.WithDescription("Event dispatch duration by type"),
	)
	if err != nil {
		return err
	}

	return nil
}

// Shutdown shuts down the telemetry provider.
func (p *Provider) Shutdown(ctx context.Context) error {
	if p.shutdown != nil {
		return p.shutdown(ctx)
	}
	return nil
}

// Tracer returns the tracer.
func (p *Provider) Tracer() trace.Tracer {
	return p.tracer
}

// Meter returns the meter.
func (p *Provider) Meter() metric.Meter {
	return p.meter
}

// RecordAgentAdded records an agent addition.
func (p *Provider) RecordAgentAdded(ctx context.Context, adapter, locationType string) {
	if p.agentsTotal != nil {
		p.agentsTotal.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("adapter", adapter),
				attribute.String("location_type", locationType),
				attribute.String("status", "added"),
			),
		)
	}
}

// RecordAgentActive records agent active state change.
func (p *Provider) RecordAgentActive(ctx context.Context, adapter, presence string, delta int64) {
	if p.agentsActive != nil {
		p.agentsActive.Add(ctx, delta,
			metric.WithAttributes(
				attribute.String("adapter", adapter),
				attribute.String("presence", presence),
			),
		)
	}
}

// RecordEvent records an event dispatch.
func (p *Provider) RecordEvent(ctx context.Context, eventType string) {
	if p.eventsTotal != nil {
		p.eventsTotal.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("type", eventType),
			),
		)
	}
}

// RecordPTYOutput records PTY output bytes.
func (p *Provider) RecordPTYOutput(ctx context.Context, agentID string, bytes int64) {
	if p.ptyOutputBytes != nil {
		p.ptyOutputBytes.Record(ctx, bytes,
			metric.WithAttributes(
				attribute.String("agent_id", agentID),
			),
		)
	}
}

// RecordAdapterCall records an adapter call duration.
func (p *Provider) RecordAdapterCall(ctx context.Context, adapter, function string, durationSeconds float64) {
	if p.adapterCallDuration != nil {
		p.adapterCallDuration.Record(ctx, durationSeconds,
			metric.WithAttributes(
				attribute.String("adapter", adapter),
				attribute.String("function", function),
			),
		)
	}
}

func setGlobalProvider(p *Provider) {
	providerMu.Lock()
	defer providerMu.Unlock()
	globalProvider = p
}

// Global returns the global telemetry provider.
func Global() *Provider {
	providerMu.RLock()
	defer providerMu.RUnlock()
	if globalProvider == nil {
		// Return a minimal provider
		return &Provider{
			tracer: otel.Tracer(ServiceName),
			meter:  otel.Meter(ServiceName),
		}
	}
	return globalProvider
}

// Tracer returns the global tracer.
func Tracer() trace.Tracer {
	return Global().Tracer()
}

// Meter returns the global meter.
func Meter() metric.Meter {
	return Global().Meter()
}

// StartSpan starts a new span using the global tracer.
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return Tracer().Start(ctx, name, opts...)
}

// ConfigFromEnv creates a Config from environment variables.
func ConfigFromEnv() Config {
	cfg := DefaultConfig()

	if endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); endpoint != "" {
		cfg.Endpoint = endpoint
	}

	if serviceName := os.Getenv("OTEL_SERVICE_NAME"); serviceName != "" {
		cfg.ServiceName = serviceName
	}

	if protocol := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL"); protocol != "" {
		cfg.Protocol = protocol
	}

	// Check if any exporter is configured
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		cfg.Enabled = true
	}

	// Check for disabled signals
	if os.Getenv("OTEL_TRACES_EXPORTER") == "none" {
		cfg.TracesEnabled = false
	}
	if os.Getenv("OTEL_METRICS_EXPORTER") == "none" {
		cfg.MetricsEnabled = false
	}
	if os.Getenv("OTEL_LOGS_EXPORTER") == "none" {
		cfg.LogsEnabled = false
	}

	return cfg
}

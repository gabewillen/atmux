// Package telemetry provides OpenTelemetry instrumentation for amux.
// This package implements the observability requirements from spec §4.2.9,
// providing traces, metrics, and logs for all core components.
//
// The telemetry package supports configuration via environment variables
// following the OTel specification, or via config file as defined in the spec.
//
// Instrumentation covers:
// - Agent lifecycle state transitions
// - PTY monitor pattern matching
// - Process tracker lifecycle and I/O events
// - Adapter WASM call spans
// - Remote agent NATS/SSH operations
// - HSM event queue dispatch
package telemetry

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
	metricSDK "go.opentelemetry.io/otel/sdk/metric"
	traceSDK "go.opentelemetry.io/otel/sdk/trace"
)

// Common sentinel errors for telemetry operations.
var (
	// ErrInitFailed indicates telemetry initialization failed.
	ErrInitFailed = errors.New("telemetry initialization failed")

	// ErrExporterFailed indicates exporter setup failed.
	ErrExporterFailed = errors.New("exporter setup failed")

	// ErrShutdownFailed indicates telemetry shutdown failed.
	ErrShutdownFailed = errors.New("telemetry shutdown failed")
)

// Config represents telemetry configuration matching spec §4.2.9.2.
type Config struct {
	Enabled     bool             `toml:"enabled"`
	ServiceName string           `toml:"service_name"`
	Exporter    ExporterConfig   `toml:"exporter"`
	Traces      TracesConfig     `toml:"traces"`
	Metrics     MetricsConfig    `toml:"metrics"`
	Logs        LogsConfig       `toml:"logs"`
}

// ExporterConfig configures OpenTelemetry exporters.
type ExporterConfig struct {
	Endpoint string `toml:"endpoint"`
	Protocol string `toml:"protocol"` // grpc, http/protobuf, http/json
}

// TracesConfig configures trace collection.
type TracesConfig struct {
	Enabled    bool    `toml:"enabled"`
	Sampler    string  `toml:"sampler"`
	SamplerArg float64 `toml:"sampler_arg"`
}

// MetricsConfig configures metrics collection.
type MetricsConfig struct {
	Enabled bool `toml:"enabled"`
}

// LogsConfig configures log collection.
type LogsConfig struct {
	Enabled bool `toml:"enabled"`
}

// Provider manages OpenTelemetry providers and exporters.
type Provider struct {
	traceProvider  *traceSDK.TracerProvider
	metricProvider *metricSDK.MeterProvider
	config         Config
}

// NewProvider creates a new telemetry provider with the given configuration.
// This function implements the initialization requirements from spec §4.2.9.
func NewProvider(config Config) (*Provider, error) {
	if !config.Enabled {
		return &Provider{config: config}, nil
	}

	// Set default service name
	if config.ServiceName == "" {
		config.ServiceName = "amux"
	}

	// Create resource
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			"",
			// Add service name and other attributes as needed
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	provider := &Provider{config: config}

	// Initialize trace provider if enabled
	if config.Traces.Enabled {
		traceProvider, err := provider.initTraceProvider(res)
		if err != nil {
			return nil, fmt.Errorf("failed to init trace provider: %w", err)
		}
		provider.traceProvider = traceProvider
		otel.SetTracerProvider(traceProvider)
	}

	// Initialize metric provider if enabled
	if config.Metrics.Enabled {
		metricProvider, err := provider.initMetricProvider(res)
		if err != nil {
			return nil, fmt.Errorf("failed to init metric provider: %w", err)
		}
		provider.metricProvider = metricProvider
		otel.SetMeterProvider(metricProvider)
	}

	// Set global text map propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return provider, nil
}

// initTraceProvider creates and configures the trace provider.
func (p *Provider) initTraceProvider(res *resource.Resource) (*traceSDK.TracerProvider, error) {
	// Create OTLP trace exporter if endpoint configured
	var exporter traceSDK.SpanExporter
	if p.config.Exporter.Endpoint != "" {
		otlpExporter, err := otlptracegrpc.New(
			context.Background(),
			otlptracegrpc.WithEndpoint(p.config.Exporter.Endpoint),
			otlptracegrpc.WithInsecure(), // Configure TLS as needed
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
		}
		exporter = otlpExporter
	} else {
		// Return no-op tracer provider if no exporter configured
		return traceSDK.NewTracerProvider(), nil
	}

	// Configure sampler
	var sampler traceSDK.Sampler
	switch p.config.Traces.Sampler {
	case "parentbased_traceidratio":
		sampler = traceSDK.ParentBased(traceSDK.TraceIDRatioBased(p.config.Traces.SamplerArg))
	case "always":
		sampler = traceSDK.AlwaysSample()
	case "never":
		sampler = traceSDK.NeverSample()
	default:
		sampler = traceSDK.AlwaysSample()
	}

	return traceSDK.NewTracerProvider(
		traceSDK.WithResource(res),
		traceSDK.WithSampler(sampler),
		traceSDK.WithSpanProcessor(traceSDK.NewBatchSpanProcessor(exporter)),
	), nil
}

// initMetricProvider creates and configures the metric provider.
func (p *Provider) initMetricProvider(res *resource.Resource) (*metricSDK.MeterProvider, error) {
	// Create OTLP metric exporter if endpoint configured
	var exporter metricSDK.Exporter
	if p.config.Exporter.Endpoint != "" {
		otlpExporter, err := otlpmetricgrpc.New(
			context.Background(),
			otlpmetricgrpc.WithEndpoint(p.config.Exporter.Endpoint),
			otlpmetricgrpc.WithInsecure(), // Configure TLS as needed
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP metric exporter: %w", err)
		}
		exporter = otlpExporter
	} else {
		// Return no-op meter provider if no endpoint configured
		return metricSDK.NewMeterProvider(), nil
	}

	return metricSDK.NewMeterProvider(
		metricSDK.WithResource(res),
		metricSDK.WithReader(metricSDK.NewPeriodicReader(
			exporter,
			metricSDK.WithInterval(30*time.Second),
		)),
	), nil
}

// Tracer returns a tracer for the given name.
func (p *Provider) Tracer(name string) trace.Tracer {
	if p.traceProvider == nil {
		return trace.NewNoopTracerProvider().Tracer(name)
	}
	return p.traceProvider.Tracer(name)
}

// Meter returns a meter for the given name.
func (p *Provider) Meter(name string) metric.Meter {
	if p.metricProvider == nil {
		return metricSDK.NewMeterProvider().Meter(name)
	}
	return p.metricProvider.Meter(name)
}

// Shutdown gracefully shuts down the telemetry providers.
func (p *Provider) Shutdown(ctx context.Context) error {
	var errs []error

	if p.traceProvider != nil {
		if err := p.traceProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("trace provider shutdown: %w", err))
		}
	}

	if p.metricProvider != nil {
		if err := p.metricProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("metric provider shutdown: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("telemetry shutdown errors: %v", errs)
	}

	return nil
}

// DefaultConfig returns a default telemetry configuration.
// It checks environment variables first, then applies defaults.
func DefaultConfig() Config {
	config := Config{
		Enabled:     getEnvBool("OTEL_SDK_DISABLED", false) == false,
		ServiceName: getEnvString("OTEL_SERVICE_NAME", "amux"),
		Exporter: ExporterConfig{
			Endpoint: getEnvString("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
			Protocol: getEnvString("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc"),
		},
		Traces: TracesConfig{
			Enabled:    getEnvString("OTEL_TRACES_EXPORTER", "otlp") != "none",
			Sampler:    getEnvString("OTEL_TRACES_SAMPLER", "parentbased_always_on"),
			SamplerArg: getEnvFloat("OTEL_TRACES_SAMPLER_ARG", 1.0),
		},
		Metrics: MetricsConfig{
			Enabled: getEnvString("OTEL_METRICS_EXPORTER", "otlp") != "none",
		},
		Logs: LogsConfig{
			Enabled: getEnvString("OTEL_LOGS_EXPORTER", "otlp") != "none",
		},
	}

	return config
}

// Helper functions for environment variable parsing
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return defaultValue
}
package telemetry

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/agentflare-ai/amux/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

// ShutdownFunc shuts down telemetry providers.
type ShutdownFunc func(context.Context) error

// Options configures telemetry setup.
type Options struct {
	TraceExporter sdktrace.SpanExporter
	MetricReader  sdkmetric.Reader
	Logger        *log.Logger
}

// Option configures telemetry options.
type Option func(*Options)

// WithTraceExporter supplies a trace exporter.
func WithTraceExporter(exporter sdktrace.SpanExporter) Option {
	return func(opts *Options) {
		opts.TraceExporter = exporter
	}
}

// WithMetricReader supplies a metrics reader.
func WithMetricReader(reader sdkmetric.Reader) Option {
	return func(opts *Options) {
		opts.MetricReader = reader
	}
}

// WithLogger supplies a logger for telemetry setup.
func WithLogger(logger *log.Logger) Option {
	return func(opts *Options) {
		opts.Logger = logger
	}
}

// Setup configures OpenTelemetry providers from config and environment.
func Setup(ctx context.Context, cfg config.TelemetryConfig, env map[string]string, optFns ...Option) (ShutdownFunc, error) {
	options := Options{}
	for _, fn := range optFns {
		fn(&options)
	}
	if options.Logger == nil {
		options.Logger = log.New(os.Stderr, "amux-otel ", log.LstdFlags)
	}
	cfg = applyEnv(cfg, env)
	if !cfg.Enabled {
		otel.SetTracerProvider(sdktrace.NewTracerProvider())
		otel.SetMeterProvider(sdkmetric.NewMeterProvider())
		return func(context.Context) error { return nil }, nil
	}
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("telemetry resource: %w", err)
	}
	traceExporter := options.TraceExporter
	if traceExporter == nil {
		traceExporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("telemetry exporter: %w", err)
		}
	}
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(selectSampler(cfg.Traces)),
		sdktrace.WithBatcher(traceExporter),
	)
	otel.SetTracerProvider(tracerProvider)
	metricReader := options.MetricReader
	if metricReader == nil {
		metricReader = sdkmetric.NewManualReader()
	}
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(metricReader),
	)
	otel.SetMeterProvider(meterProvider)
	options.Logger.Printf("telemetry enabled: service=%s protocol=%s", cfg.ServiceName, cfg.Exporter.Protocol)
	return func(shutdownCtx context.Context) error {
		var firstErr error
		if err := tracerProvider.Shutdown(shutdownCtx); err != nil {
			firstErr = err
		}
		if err := meterProvider.Shutdown(shutdownCtx); err != nil && firstErr == nil {
			firstErr = err
		}
		return firstErr
	}, nil
}

// EmitBaselineMetrics records baseline metrics used in conformance testing.
func EmitBaselineMetrics(ctx context.Context) error {
	meter := otel.Meter("amux.telemetry")
	activeAgents, err := meter.Int64Counter("amux.agent.active")
	if err != nil {
		return fmt.Errorf("baseline metrics: %w", err)
	}
	activeAgents.Add(ctx, 1, metric.WithAttributes(attribute.String("state", "running")))
	ptyBytes, err := meter.Int64Counter("amux.pty.bytes")
	if err != nil {
		return fmt.Errorf("baseline metrics: %w", err)
	}
	ptyBytes.Add(ctx, 128, metric.WithAttributes(attribute.String("stream", "stdout")))
	return nil
}

// CollectMetrics uses a manual reader to collect metrics into ResourceMetrics.
func CollectMetrics(ctx context.Context, reader *sdkmetric.ManualReader) (metricdata.ResourceMetrics, error) {
	var out metricdata.ResourceMetrics
	if reader == nil {
		return out, fmt.Errorf("metrics reader is nil")
	}
	if err := reader.Collect(ctx, &out); err != nil {
		return out, fmt.Errorf("collect metrics: %w", err)
	}
	return out, nil
}

func applyEnv(cfg config.TelemetryConfig, env map[string]string) config.TelemetryConfig {
	if env == nil {
		return cfg
	}
	if value, ok := env["OTEL_SERVICE_NAME"]; ok && value != "" {
		cfg.ServiceName = value
	}
	if value, ok := env["OTEL_EXPORTER_OTLP_ENDPOINT"]; ok && value != "" {
		cfg.Exporter.Endpoint = value
		cfg.Enabled = true
	}
	if value, ok := env["OTEL_EXPORTER_OTLP_PROTOCOL"]; ok && value != "" {
		cfg.Exporter.Protocol = value
	}
	if value, ok := env["OTEL_TRACES_SAMPLER"]; ok && value != "" {
		cfg.Traces.Sampler = value
	}
	if value, ok := env["OTEL_TRACES_SAMPLER_ARG"]; ok && value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			cfg.Traces.SamplerArg = parsed
		}
	}
	if value, ok := env["OTEL_METRICS_EXPORTER"]; ok && strings.ToLower(value) == "none" {
		cfg.Metrics.Enabled = false
	}
	if value, ok := env["OTEL_LOGS_EXPORTER"]; ok && strings.ToLower(value) == "none" {
		cfg.Logs.Enabled = false
	}
	return cfg
}

func selectSampler(cfg config.TelemetryTracesConfig) sdktrace.Sampler {
	switch strings.ToLower(cfg.Sampler) {
	case "always_on":
		return sdktrace.AlwaysSample()
	case "always_off":
		return sdktrace.NeverSample()
	case "traceidratio":
		return sdktrace.TraceIDRatioBased(cfg.SamplerArg)
	case "parentbased_traceidratio":
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SamplerArg))
	default:
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.1))
	}
}

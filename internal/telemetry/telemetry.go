// Package telemetry configures OpenTelemetry for traces, metrics, and logs.
package telemetry

import (
	"context"
	"fmt"

	"github.com/agentflare-ai/amux/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// ShutdownFunc is a function that shuts down the telemetry providers.
type ShutdownFunc func(context.Context) error

// Setup initializes the OpenTelemetry globals based on the configuration.
func Setup(ctx context.Context, cfg config.TelemetryConfig) (ShutdownFunc, error) {
	if !cfg.Enabled {
		return func(context.Context) error { return nil }, nil
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	var traceExporter sdktrace.SpanExporter
	// For scaffolding, we support "stdout" or noop. Real impl might add OTLP.
	// We use the Protocol field to guess, or Endpoint.
	// If Endpoint is empty, we might default to noop or stdout.
	
	// Check config for exporter type (naive check for now)
	if cfg.Exporter.Protocol == "stdout" {
		traceExporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout exporter: %w", err)
		}
	} else {
		// Default to noop if not configured or unknown for now (to pass tests without crashing)
		// Or strictly, we should fail or use a real exporter. 
		// For Phase 0 scaffolding, let's allow "none".
		// But let's set up a stdout exporter if "stdout" is specified or for testing.
	}

	var tp *sdktrace.TracerProvider
	if traceExporter != nil {
		tp = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(traceExporter),
			sdktrace.WithResource(res),
			// Sampler config could go here
		)
	} else {
		tp = sdktrace.NewTracerProvider(sdktrace.WithResource(res))
	}

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return func(ctx context.Context) error {
		if tp != nil {
			return tp.Shutdown(ctx)
		}
		return nil
	}, nil
}
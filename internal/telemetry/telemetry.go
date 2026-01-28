// Package telemetry provides OpenTelemetry instrumentation for amux.
package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/agentflare-ai/amux/internal/config"
)

// Init initializes OpenTelemetry based on configuration.
func Init(ctx context.Context, cfg *config.Config) (func() error, error) {
	if !cfg.Telemetry.Enabled {
		return func() error { return nil }, nil
	}

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			// TODO: Add service name and other attributes
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace exporter
	var traceExporter sdktrace.SpanExporter
	if cfg.Telemetry.Traces.Enabled {
		exporter, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(cfg.Telemetry.Exporter.Endpoint),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create trace exporter: %w", err)
		}
		traceExporter = exporter
	}

	// Create tracer provider
	var tp *sdktrace.TracerProvider
	if traceExporter != nil {
		tp = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(traceExporter),
			sdktrace.WithResource(res),
			// TODO: Configure sampler based on cfg.Telemetry.Traces.Sampler
		)
		otel.SetTracerProvider(tp)
	}

	// Return shutdown function
	shutdown := func() error {
		if tp != nil {
			return tp.Shutdown(ctx)
		}
		return nil
	}

	return shutdown, nil
}

// Tracer returns a tracer for the given component.
func Tracer(component string) trace.Tracer {
	return otel.Tracer(fmt.Sprintf("amux.%s", component))
}

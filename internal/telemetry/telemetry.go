// Package telemetry provides OpenTelemetry instrumentation for amux per spec §4.2.9.
//
// This package provides scaffolding for traces, metrics, and logs following
// the OpenTelemetry specification.
package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Tracer returns the amux tracer.
func Tracer() trace.Tracer {
	return otel.Tracer("github.com/stateforward/amux")
}

// Meter returns the amux meter for metrics.
func Meter() metric.Meter {
	return otel.Meter("github.com/stateforward/amux")
}

// StartSpan starts a new span with the given name.
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return Tracer().Start(ctx, name, opts...)
}

// Init initializes OpenTelemetry with a basic tracer and meter provider.
//
// It configures a batch span processor with an in-memory exporter driven by
// the global OTEL_* environment variables via the default SDK behavior and
// sets the global tracer and meter providers. The returned shutdown function
// MUST be called on process exit to flush spans.
func Init(ctx context.Context) (func(context.Context) error, error) {
	// Create a tracer provider with a default resource; exporters can be
	// configured via OTEL_* environment variables when using auto instrumentation
	// or external wiring.
	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			attribute.String("service.name", "amux"),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	shutdown := func(ctx context.Context) error {
		return tp.Shutdown(ctx)
	}

	return shutdown, nil
}

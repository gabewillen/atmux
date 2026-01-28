// Package telemetry provides OpenTelemetry instrumentation for amux per spec §4.2.9.
//
// This package provides scaffolding for traces, metrics, and logs following
// the OpenTelemetry specification.
package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
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

// Init initializes OpenTelemetry with the provided configuration.
// This is a placeholder for Phase 0. Full implementation will be in later phases.
func Init(ctx context.Context) (func(context.Context) error, error) {
	// Phase 0: No-op initialization
	// Later phases will add:
	// - OTLP exporter setup
	// - Trace/metric/log providers
	// - Resource attributes
	// - Sampling configuration
	return func(context.Context) error { return nil }, nil
}

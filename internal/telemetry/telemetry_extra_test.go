package telemetry

import (
	"context"
	"log"
	"os"
	"testing"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type stubExporter struct{}

func (stubExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	_ = ctx
	_ = spans
	return nil
}

func (stubExporter) Shutdown(ctx context.Context) error {
	_ = ctx
	return nil
}

func TestTelemetryOptions(t *testing.T) {
	options := Options{}
	exporter := stubExporter{}
	reader := sdkmetric.NewManualReader()
	logger := log.New(os.Stderr, "test ", log.LstdFlags)
	WithTraceExporter(exporter)(&options)
	WithMetricReader(reader)(&options)
	WithLogger(logger)(&options)
	if options.TraceExporter == nil {
		t.Fatalf("expected trace exporter")
	}
	if options.MetricReader == nil {
		t.Fatalf("expected metric reader")
	}
	if options.Logger == nil {
		t.Fatalf("expected logger")
	}
}

package telemetry

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/agentflare-ai/amux/internal/config"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type noopExporter struct{}

func (noopExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	_ = ctx
	_ = spans
	return nil
}

func (noopExporter) Shutdown(ctx context.Context) error {
	_ = ctx
	return nil
}

func TestSetupDisabledTelemetry(t *testing.T) {
	ctx := context.Background()
	cfg := config.TelemetryConfig{Enabled: false, ServiceName: "amux"}
	shutdown, err := Setup(ctx, cfg, nil)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

func TestSetupFromEnvEnablesTelemetry(t *testing.T) {
	ctx := context.Background()
	env := map[string]string{
		"OTEL_SERVICE_NAME":           "amux-test",
		"OTEL_EXPORTER_OTLP_ENDPOINT": "http://localhost:4317",
		"OTEL_EXPORTER_OTLP_PROTOCOL": "grpc",
	}
	reader := sdkmetric.NewManualReader()
	logger := log.New(os.Stderr, "test ", log.LstdFlags)
	cfg := config.TelemetryConfig{Enabled: false, ServiceName: "amux"}
	shutdown, err := Setup(ctx, cfg, env, WithTraceExporter(noopExporter{}), WithMetricReader(reader), WithLogger(logger))
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

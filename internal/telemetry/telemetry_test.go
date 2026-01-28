package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

func TestBaselineMetrics(t *testing.T) {
	ctx := context.Background()
	reader := sdkmetric.NewManualReader()
	cfg := config.TelemetryConfig{
		Enabled:     true,
		ServiceName: "amux",
		Metrics: config.TelemetryMetricsConfig{
			Enabled:  true,
			Interval: time.Second,
		},
	}
	shutdown, err := Setup(ctx, cfg, nil, WithMetricReader(reader))
	if err != nil {
		t.Fatalf("setup telemetry: %v", err)
	}
	if err := EmitBaselineMetrics(ctx); err != nil {
		t.Fatalf("emit baseline: %v", err)
	}
	if _, err := CollectMetrics(ctx, reader); err != nil {
		t.Fatalf("collect metrics: %v", err)
	}
	if err := shutdown(ctx); err != nil {
		t.Fatalf("shutdown telemetry: %v", err)
	}
}

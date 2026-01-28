package telemetry

import (
	"context"
	"testing"

	"github.com/agentflare-ai/amux/internal/config"
	"go.opentelemetry.io/otel/trace"
)

func TestInit(t *testing.T) {
	// Test disabled telemetry
	otelConfig := &config.OTelConfig{
		Enabled:        false,
		ServiceName:    "test",
		ServiceVersion: "1.0.0",
	}

	if err := Init(otelConfig); err != nil {
		t.Fatalf("Failed to initialize disabled telemetry: %v", err)
	}

	// Test enabled telemetry
	otelConfig.Enabled = true
	otelConfig.Exporter.Type = "stdout"

	if err := Init(otelConfig); err != nil {
		t.Fatalf("Failed to initialize enabled telemetry: %v", err)
	}

	// Test that tracer is available
	ctx, span := StartSpan(context.Background(), "test-span")
	defer span.End()

	if span == nil {
		t.Error("Expected non-nil span")
	}

	if ctx == nil {
		t.Error("Expected non-nil context")
	}

	// Test error recording
	testErr := &testError{"test error"}
	RecordError(ctx, testErr)

	// Cleanup
	Shutdown(context.Background())
}

func TestGetServiceInfo(t *testing.T) {
	// Reset global telemetry for clean test
	globalTelemetry = nil

	// Test with no global telemetry
	infoName, infoVersion := GetServiceInfo()
	if infoName != "amux" || infoVersion != "unknown" {
		t.Errorf("Expected 'amux/unknown', got '%s/%s'", infoName, infoVersion)
	}

	// Test with global telemetry
	otelConfig := &config.OTelConfig{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "2.0.0",
		Exporter: config.OTelExporterConfig{
			Type: "stdout",
		},
	}

	err := Init(otelConfig)
	if err != nil {
		t.Fatalf("Failed to initialize telemetry: %v", err)
	}

	if globalTelemetry == nil {
		t.Fatal("globalTelemetry is nil after Init")
	}

	infoName, infoVersion = GetServiceInfo()
	if infoName != "test-service" || infoVersion != "2.0.0" {
		t.Errorf("Expected 'test-service/2.0.0', got '%s/%s'", infoName, infoVersion)
	}

	Shutdown(context.Background())
}

func TestWithSpan(t *testing.T) {
	otelConfig := &config.OTelConfig{
		Enabled:     false, // Use no-op for testing
		ServiceName: "test",
	}

	if err := Init(otelConfig); err != nil {
		t.Fatalf("Failed to initialize telemetry: %v", err)
	}

	ctx := context.Background()
	err := WithSpan(ctx, "test-operation", func(ctx context.Context) error {
		span := trace.SpanFromContext(ctx)
		if span == nil {
			t.Error("Expected non-nil span in operation")
		}
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// Test error type
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

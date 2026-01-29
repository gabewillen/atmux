package telemetry

import (
	"context"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Enabled {
		t.Error("Enabled should be false by default")
	}
	if cfg.ServiceName != ServiceName {
		t.Errorf("ServiceName = %q, want %q", cfg.ServiceName, ServiceName)
	}
	if cfg.Endpoint != "localhost:4317" {
		t.Errorf("Endpoint = %q, want %q", cfg.Endpoint, "localhost:4317")
	}
	if cfg.Protocol != "grpc" {
		t.Errorf("Protocol = %q, want %q", cfg.Protocol, "grpc")
	}
	if !cfg.TracesEnabled {
		t.Error("TracesEnabled should be true by default")
	}
	if cfg.TraceSampler != "parentbased_traceidratio" {
		t.Errorf("TraceSampler = %q, want %q", cfg.TraceSampler, "parentbased_traceidratio")
	}
	if cfg.TraceSamplerArg != 0.1 {
		t.Errorf("TraceSamplerArg = %v, want 0.1", cfg.TraceSamplerArg)
	}
	if !cfg.MetricsEnabled {
		t.Error("MetricsEnabled should be true by default")
	}
	if !cfg.LogsEnabled {
		t.Error("LogsEnabled should be true by default")
	}
}

func TestServiceNameConstant(t *testing.T) {
	if ServiceName != "amux" {
		t.Errorf("ServiceName = %q, want %q", ServiceName, "amux")
	}
}

func TestSpecVersionConstant(t *testing.T) {
	if SpecVersion != "v1.22" {
		t.Errorf("SpecVersion = %q, want %q", SpecVersion, "v1.22")
	}
}

func TestInitNoopProvider(t *testing.T) {
	// When Enabled is false, Init should return a noop provider without error.
	cfg := Config{
		Enabled:     false,
		ServiceName: "test-service",
	}

	ctx := context.Background()
	p, err := Init(ctx, cfg)
	if err != nil {
		t.Fatalf("Init() error: %v", err)
	}
	if p == nil {
		t.Fatal("Init() returned nil provider")
	}

	// Noop provider should have a tracer and meter, but no trace/meter providers
	if p.Tracer() == nil {
		t.Error("Tracer() should not be nil for noop provider")
	}
	if p.Meter() == nil {
		t.Error("Meter() should not be nil for noop provider")
	}

	// Shutdown should succeed on noop provider
	if err := p.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown() error: %v", err)
	}
}

func TestGlobalProviderDefault(t *testing.T) {
	// Global() should never return nil, even without initialization.
	p := Global()
	if p == nil {
		t.Fatal("Global() returned nil")
	}

	// Should have tracer and meter
	if p.Tracer() == nil {
		t.Error("Global().Tracer() is nil")
	}
	if p.Meter() == nil {
		t.Error("Global().Meter() is nil")
	}
}

func TestPackageLevelTracer(t *testing.T) {
	tr := Tracer()
	if tr == nil {
		t.Error("Tracer() returned nil")
	}
}

func TestPackageLevelMeter(t *testing.T) {
	m := Meter()
	if m == nil {
		t.Error("Meter() returned nil")
	}
}

func TestStartSpan(t *testing.T) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "test.operation")
	if span == nil {
		t.Fatal("StartSpan returned nil span")
	}
	defer span.End()

	// Context should not be nil
	if ctx == nil {
		t.Error("StartSpan returned nil context")
	}
}

func TestConfigFromEnvDefaults(t *testing.T) {
	// ConfigFromEnv without any env vars should return defaults
	cfg := ConfigFromEnv()

	if cfg.ServiceName != ServiceName {
		t.Errorf("ServiceName = %q, want %q", cfg.ServiceName, ServiceName)
	}
	if cfg.Protocol != "grpc" {
		t.Errorf("Protocol = %q, want %q", cfg.Protocol, "grpc")
	}
	if cfg.TraceSampler != "parentbased_traceidratio" {
		t.Errorf("TraceSampler = %q, want %q", cfg.TraceSampler, "parentbased_traceidratio")
	}
}

func TestConfigFromEnvWithOverrides(t *testing.T) {
	// Set env vars and verify they are picked up
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel-collector:4317")
	t.Setenv("OTEL_SERVICE_NAME", "my-amux")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")

	cfg := ConfigFromEnv()

	if cfg.Endpoint != "http://otel-collector:4317" {
		t.Errorf("Endpoint = %q, want %q", cfg.Endpoint, "http://otel-collector:4317")
	}
	if cfg.ServiceName != "my-amux" {
		t.Errorf("ServiceName = %q, want %q", cfg.ServiceName, "my-amux")
	}
	if cfg.Protocol != "http/protobuf" {
		t.Errorf("Protocol = %q, want %q", cfg.Protocol, "http/protobuf")
	}
	// Should be enabled when endpoint is set
	if !cfg.Enabled {
		t.Error("Enabled should be true when OTEL_EXPORTER_OTLP_ENDPOINT is set")
	}
}

func TestConfigFromEnvDisabledSignals(t *testing.T) {
	t.Setenv("OTEL_TRACES_EXPORTER", "none")
	t.Setenv("OTEL_METRICS_EXPORTER", "none")
	t.Setenv("OTEL_LOGS_EXPORTER", "none")

	cfg := ConfigFromEnv()

	if cfg.TracesEnabled {
		t.Error("TracesEnabled should be false when OTEL_TRACES_EXPORTER=none")
	}
	if cfg.MetricsEnabled {
		t.Error("MetricsEnabled should be false when OTEL_METRICS_EXPORTER=none")
	}
	if cfg.LogsEnabled {
		t.Error("LogsEnabled should be false when OTEL_LOGS_EXPORTER=none")
	}
}

func TestProviderRecordMethodsNilMetrics(t *testing.T) {
	// A provider with nil metric instruments should not panic.
	p := &Provider{
		config: Config{},
	}

	ctx := context.Background()

	// These should be no-ops when metric counters/histograms are nil
	p.RecordAgentAdded(ctx, "test-adapter", "local")
	p.RecordAgentActive(ctx, "test-adapter", "online", 1)
	p.RecordEvent(ctx, "lifecycle.started")
	p.RecordPTYOutput(ctx, "agent-1", 1024)
	p.RecordAdapterCall(ctx, "claude-code", "on_output", 0.05)
}

func TestProviderShutdownNilHandler(t *testing.T) {
	// Provider with nil shutdown handler should return nil
	p := &Provider{}
	if err := p.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown() with nil handler should return nil, got: %v", err)
	}
}

func TestGetSampler(t *testing.T) {
	// Verify getSampler does not panic for all known sampler names
	samplerNames := []string{
		"always_on",
		"always_off",
		"traceidratio",
		"parentbased_always_on",
		"parentbased_always_off",
		"parentbased_traceidratio",
		"unknown_sampler",
	}

	for _, name := range samplerNames {
		t.Run(name, func(t *testing.T) {
			s := getSampler(name, 0.5)
			if s == nil {
				t.Errorf("getSampler(%q) returned nil", name)
			}
		})
	}
}

func TestSetGlobalProviderAndRetrieve(t *testing.T) {
	cfg := Config{
		ServiceName: "test-set-global",
	}
	p, err := Init(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Init error: %v", err)
	}

	got := Global()
	if got != p {
		t.Error("Global() did not return the provider set by Init")
	}
}

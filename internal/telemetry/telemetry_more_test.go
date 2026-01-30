package telemetry

import (
	"strings"
	"testing"

	"github.com/agentflare-ai/amux/internal/config"
)

func TestSelectSampler(t *testing.T) {
	cases := []struct {
		name   string
		cfg    config.TelemetryTracesConfig
		contains []string
	}{
		{
			name: "always_on",
			cfg:  config.TelemetryTracesConfig{Sampler: "always_on"},
			contains: []string{"AlwaysOnSampler"},
		},
		{
			name: "always_off",
			cfg:  config.TelemetryTracesConfig{Sampler: "always_off"},
			contains: []string{"AlwaysOffSampler"},
		},
		{
			name: "traceidratio",
			cfg:  config.TelemetryTracesConfig{Sampler: "traceidratio", SamplerArg: 0.25},
			contains: []string{"TraceIDRatioBased{0.25}"},
		},
		{
			name: "parentbased_traceidratio",
			cfg:  config.TelemetryTracesConfig{Sampler: "parentbased_traceidratio", SamplerArg: 0.5},
			contains: []string{"ParentBased{root:TraceIDRatioBased{0.5}"},
		},
		{
			name: "default",
			cfg:  config.TelemetryTracesConfig{Sampler: "unknown"},
			contains: []string{"ParentBased{root:TraceIDRatioBased{0.1}"},
		},
	}
	for _, tc := range cases {
		sampler := selectSampler(tc.cfg)
		desc := sampler.Description()
		for _, want := range tc.contains {
			if !strings.Contains(desc, want) {
				t.Fatalf("%s: expected description to contain %q, got %q", tc.name, want, desc)
			}
		}
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	cfg := config.TelemetryConfig{
		Enabled:     false,
		ServiceName: "base",
		Exporter:    config.TelemetryExporterConfig{Endpoint: "http://old", Protocol: "grpc"},
		Traces:      config.TelemetryTracesConfig{Sampler: "always_off", SamplerArg: 0.5},
		Metrics:     config.TelemetryMetricsConfig{Enabled: true},
		Logs:        config.TelemetryLogsConfig{Enabled: true},
	}
	env := map[string]string{
		"OTEL_SERVICE_NAME":          "svc",
		"OTEL_EXPORTER_OTLP_ENDPOINT": "http://new",
		"OTEL_EXPORTER_OTLP_PROTOCOL": "http/protobuf",
		"OTEL_TRACES_SAMPLER":        "traceidratio",
		"OTEL_TRACES_SAMPLER_ARG":    "0.25",
		"OTEL_METRICS_EXPORTER":      "none",
		"OTEL_LOGS_EXPORTER":         "none",
	}
	out := applyEnv(cfg, env)
	if out.ServiceName != "svc" || out.Exporter.Endpoint != "http://new" || out.Exporter.Protocol != "http/protobuf" {
		t.Fatalf("unexpected exporter overrides: %#v", out)
	}
	if !out.Enabled {
		t.Fatalf("expected telemetry enabled")
	}
	if out.Traces.Sampler != "traceidratio" || out.Traces.SamplerArg != 0.25 {
		t.Fatalf("unexpected trace overrides: %#v", out.Traces)
	}
	if out.Metrics.Enabled || out.Logs.Enabled {
		t.Fatalf("expected metrics/logs disabled")
	}

	env["OTEL_TRACES_SAMPLER_ARG"] = "not-a-number"
	out = applyEnv(cfg, env)
	if out.Traces.SamplerArg != cfg.Traces.SamplerArg {
		t.Fatalf("expected sampler arg to remain unchanged")
	}
}

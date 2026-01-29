package config

import (
	"testing"
)

func TestEnvOverrides(t *testing.T) {
	env := map[string]string{
		"AMUX__GENERAL__LOG_LEVEL":                     "info",
		"AMUX__ADAPTERS__CLAUDE_CODE__CLI__CONSTRAINT": ">=1.0.0",
	}
	overrides, err := EnvOverrides(env)
	if err != nil {
		t.Fatalf("env overrides: %v", err)
	}
	general := overrides["general"].(map[string]any)
	if general["log_level"].(string) != "info" {
		t.Fatalf("unexpected log_level")
	}
	adapters := overrides["adapters"].(map[string]any)
	claude := adapters["claude-code"].(map[string]any)
	cli := claude["cli"].(map[string]any)
	if cli["constraint"].(string) != ">=1.0.0" {
		t.Fatalf("adapter name normalization failed")
	}
}

func TestParseByteSize(t *testing.T) {
	value, err := ParseByteSize("1MB")
	if err != nil {
		t.Fatalf("parse bytes: %v", err)
	}
	if value.Bytes() != 1024*1024 {
		t.Fatalf("unexpected bytes: %d", value.Bytes())
	}
}

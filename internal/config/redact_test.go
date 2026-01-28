package config

import (
	"testing"
)

func TestConfig_Redacted(t *testing.T) {
	cfg := Config{
		General: GeneralConfig{LogLevel: "debug"},
		Adapters: map[string]map[string]any{
			"claude": {
				"api_key": "sk-12345",
				"model":   "claude-3",
			},
		},
	}

	redacted := cfg.Redacted()

	// Check original is untouched
	if cfg.Adapters["claude"]["api_key"] != "sk-12345" {
		t.Error("Original config was modified")
	}

	// Check redacted copy
	claude := redacted.Adapters["claude"]
	if claude["api_key"] != "[REDACTED]" {
		t.Errorf("api_key not redacted: %v", claude["api_key"])
	}
	if claude["model"] != "claude-3" {
		t.Errorf("model shouldn't be redacted: %v", claude["model"])
	}
	if redacted.General.LogLevel != "debug" {
		t.Errorf("General config lost")
	}
}

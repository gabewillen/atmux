package config_test

import (
	"testing"
	"time"

	"github.com/stateforward/amux/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	
	if cfg.General.LogLevel != "info" {
		t.Errorf("expected default log level 'info', got '%s'", cfg.General.LogLevel)
	}
	
	if time.Duration(cfg.Timeouts.Idle) != 30*time.Second {
		t.Errorf("expected idle timeout 30s, got %v", cfg.Timeouts.Idle)
	}
}

func TestParseByteSize(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		wantErr  bool
	}{
		{"1024", 1024, false},
		{"1KB", 1024, false},
		{"1MB", 1024 * 1024, false},
		{"10MB", 10 * 1024 * 1024, false},
		{"invalid", 0, true},
	}
	
	for _, tt := range tests {
		result, err := config.ParseByteSize(tt.input)
		if tt.wantErr && err == nil {
			t.Errorf("ParseByteSize(%q) expected error, got nil", tt.input)
		}
		if !tt.wantErr && err != nil {
			t.Errorf("ParseByteSize(%q) unexpected error: %v", tt.input, err)
		}
		if !tt.wantErr && result != tt.expected {
			t.Errorf("ParseByteSize(%q) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

package config

import (
	"os"
	"testing"
	"time"
)

func TestSetDefaults(t *testing.T) {
	config := &Config{}
	setDefaults(config)

	if config.Core.LogLevel != "info" {
		t.Errorf("Expected log level 'info', got %q", config.Core.LogLevel)
	}

	if config.Agents.DefaultStrategy != "merge-commit" {
		t.Errorf("Expected default strategy 'merge-commit', got %q", config.Agents.DefaultStrategy)
	}

	if config.Remote.RequestTimeout != 30*time.Second {
		t.Errorf("Expected request timeout 30s, got %v", config.Remote.RequestTimeout)
	}

	if config.OTel.ServiceName != "amux" {
		t.Errorf("Expected service name 'amux', got %q", config.OTel.ServiceName)
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	// Set test environment variables
	os.Setenv("AMUX__LOG_LEVEL", "debug")
	os.Setenv("AMUX__DEBUG", "true")
	os.Setenv("AMUX__DATA_DIR", "/tmp/test")
	defer func() {
		os.Unsetenv("AMUX__LOG_LEVEL")
		os.Unsetenv("AMUX__DEBUG")
		os.Unsetenv("AMUX__DATA_DIR")
	}()

	config := &Config{}
	setDefaults(config)

	if err := applyEnvOverrides(config); err != nil {
		t.Fatalf("Failed to apply env overrides: %v", err)
	}

	if config.Core.LogLevel != "debug" {
		t.Errorf("Expected log level 'debug', got %q", config.Core.LogLevel)
	}

	if !config.Core.Debug {
		t.Errorf("Expected debug true, got %v", config.Core.Debug)
	}

	if config.Core.DataDir != "/tmp/test" {
		t.Errorf("Expected data dir '/tmp/test', got %q", config.Core.DataDir)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  &Config{Core: CoreConfig{LogLevel: "info"}},
			wantErr: false,
		},
		{
			name:    "invalid config - empty log level",
			config:  &Config{Core: CoreConfig{LogLevel: ""}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

package config

import (
	"testing"
	"time"
)

func TestDurationUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{
			name:     "seconds",
			input:    "30s",
			expected: 30 * time.Second,
		},
		{
			name:     "minutes",
			input:    "5m",
			expected: 5 * time.Minute,
		},
		{
			name:     "hours",
			input:    "1h",
			expected: time.Hour,
		},
		{
			name:     "milliseconds",
			input:    "500ms",
			expected: 500 * time.Millisecond,
		},
		{
			name:     "complex",
			input:    "1h30m",
			expected: 90 * time.Minute,
		},
		{
			name:    "invalid",
			input:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Duration
			err := d.UnmarshalText([]byte(tt.input))

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if d.Duration != tt.expected {
				t.Errorf("Duration = %v, want %v", d.Duration, tt.expected)
			}
		})
	}
}

func TestByteSizeUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
		wantErr  bool
	}{
		{
			name:     "bytes integer",
			input:    "1024",
			expected: 1024,
		},
		{
			name:     "bytes unit",
			input:    "1024B",
			expected: 1024,
		},
		{
			name:     "kilobytes",
			input:    "1KB",
			expected: 1024,
		},
		{
			name:     "megabytes",
			input:    "1MB",
			expected: 1024 * 1024,
		},
		{
			name:     "gigabytes",
			input:    "1GB",
			expected: 1024 * 1024 * 1024,
		},
		{
			name:     "with spaces",
			input:    " 10 MB ",
			expected: 10 * 1024 * 1024,
		},
		{
			name:    "invalid unit",
			input:   "10TB",
			wantErr: true,
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b ByteSize
			err := b.UnmarshalText([]byte(tt.input))

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if b.Bytes != tt.expected {
				t.Errorf("Bytes = %d, want %d", b.Bytes, tt.expected)
			}
		})
	}
}

func TestByteSizeMarshal(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{
			name:     "gigabytes",
			bytes:    1024 * 1024 * 1024,
			expected: "1GB",
		},
		{
			name:     "megabytes",
			bytes:    10 * 1024 * 1024,
			expected: "10MB",
		},
		{
			name:     "kilobytes",
			bytes:    64 * 1024,
			expected: "64KB",
		},
		{
			name:     "bytes",
			bytes:    500,
			expected: "500B",
		},
		{
			name:     "non-round megabytes",
			bytes:    1024*1024 + 512,
			expected: "1049088B", // Falls through to bytes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := ByteSize{Bytes: tt.bytes}
			data, err := b.MarshalText()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if string(data) != tt.expected {
				t.Errorf("MarshalText() = %q, want %q", string(data), tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.General.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want %q", cfg.General.LogLevel, "info")
	}

	if cfg.Timeouts.Idle.Duration != 30*time.Second {
		t.Errorf("Idle = %v, want %v", cfg.Timeouts.Idle.Duration, 30*time.Second)
	}

	if cfg.Remote.Transport != "nats" {
		t.Errorf("Transport = %q, want %q", cfg.Remote.Transport, "nats")
	}

	if cfg.Node.Role != "director" {
		t.Errorf("Role = %q, want %q", cfg.Node.Role, "director")
	}
}

func TestLoader_setConfigValue(t *testing.T) {
	loader := NewLoader(nil)

	tests := []struct {
		name     string
		path     []string
		value    string
		checkFn  func(*Config) bool
	}{
		{
			name:  "general.log_level",
			path:  []string{"general", "log_level"},
			value: "debug",
			checkFn: func(c *Config) bool {
				return c.General.LogLevel == "debug"
			},
		},
		{
			name:  "timeouts.idle",
			path:  []string{"timeouts", "idle"},
			value: "1m",
			checkFn: func(c *Config) bool {
				return c.Timeouts.Idle.Duration == time.Minute
			},
		},
		{
			name:  "process.capture_mode",
			path:  []string{"process", "capture_mode"},
			value: "stdout",
			checkFn: func(c *Config) bool {
				return c.Process.CaptureMode == "stdout"
			},
		},
		{
			name:  "process.stream_buffer_size",
			path:  []string{"process", "stream_buffer_size"},
			value: "2MB",
			checkFn: func(c *Config) bool {
				return c.Process.StreamBufferSize.Bytes == 2*1024*1024
			},
		},
		{
			name:  "git.merge.strategy",
			path:  []string{"git", "merge", "strategy"},
			value: "rebase",
			checkFn: func(c *Config) bool {
				return c.Git.Merge.Strategy == "rebase"
			},
		},
		{
			name:  "events.batch_max_events",
			path:  []string{"events", "batch_max_events"},
			value: "500",
			checkFn: func(c *Config) bool {
				return c.Events.BatchMaxEvents == 500
			},
		},
		{
			name:  "events.coalesce.io_streams",
			path:  []string{"events", "coalesce", "io_streams"},
			value: "false",
			checkFn: func(c *Config) bool {
				return !c.Events.Coalesce.IOStreams
			},
		},
		{
			name:  "remote.transport",
			path:  []string{"remote", "transport"},
			value: "ssh_yamux",
			checkFn: func(c *Config) bool {
				return c.Remote.Transport == "ssh_yamux"
			},
		},
		{
			name:  "remote.nats.url",
			path:  []string{"remote", "nats", "url"},
			value: "nats://example.com:4222",
			checkFn: func(c *Config) bool {
				return c.Remote.NATS.URL == "nats://example.com:4222"
			},
		},
		{
			name:  "nats.mode",
			path:  []string{"nats", "mode"},
			value: "external",
			checkFn: func(c *Config) bool {
				return c.NATS.Mode == "external"
			},
		},
		{
			name:  "node.role",
			path:  []string{"node", "role"},
			value: "manager",
			checkFn: func(c *Config) bool {
				return c.Node.Role == "manager"
			},
		},
		{
			name:  "daemon.socket_path",
			path:  []string{"daemon", "socket_path"},
			value: "/tmp/test.sock",
			checkFn: func(c *Config) bool {
				return c.Daemon.SocketPath == "/tmp/test.sock"
			},
		},
		{
			name:  "plugins.allow_remote",
			path:  []string{"plugins", "allow_remote"},
			value: "false",
			checkFn: func(c *Config) bool {
				return !c.Plugins.AllowRemote
			},
		},
		{
			name:  "telemetry.enabled",
			path:  []string{"telemetry", "enabled"},
			value: "true",
			checkFn: func(c *Config) bool {
				return c.Telemetry.Enabled
			},
		},
		{
			name:  "telemetry.exporter.endpoint",
			path:  []string{"telemetry", "exporter", "endpoint"},
			value: "http://otel:4317",
			checkFn: func(c *Config) bool {
				return c.Telemetry.Exporter.Endpoint == "http://otel:4317"
			},
		},
		{
			name:  "telemetry.traces.sampler_arg",
			path:  []string{"telemetry", "traces", "sampler_arg"},
			value: "0.5",
			checkFn: func(c *Config) bool {
				return c.Telemetry.Traces.SamplerArg == 0.5
			},
		},
		{
			name:  "adapters.claude-code.cli.constraint",
			path:  []string{"adapters", "claude-code", "cli", "constraint"},
			value: ">=1.0.0",
			checkFn: func(c *Config) bool {
				adapterCfg, ok := c.Adapters["claude-code"].(map[string]any)
				if !ok {
					return false
				}
				cliCfg, ok := adapterCfg["cli"].(map[string]any)
				if !ok {
					return false
				}
				return cliCfg["constraint"] == ">=1.0.0"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset loader config
			loader.config = DefaultConfig()

			if err := loader.setConfigValue(tt.path, tt.value); err != nil {
				t.Errorf("setConfigValue() error: %v", err)
				return
			}

			if !tt.checkFn(loader.config) {
				t.Errorf("setConfigValue(%v, %q) did not set expected value", tt.path, tt.value)
			}
		})
	}
}

func TestLoader_Merge(t *testing.T) {
	loader := NewLoader(nil)
	loader.config = DefaultConfig()

	source := &Config{
		General: GeneralConfig{
			LogLevel: "debug",
		},
		Timeouts: TimeoutsConfig{
			Idle: Duration{Duration: 2 * time.Minute},
		},
		Process: ProcessConfig{
			CaptureMode: "stderr",
		},
		Adapters: map[string]any{
			"test-adapter": map[string]any{
				"key": "value",
			},
		},
	}

	loader.merge(source)

	// Check merged values
	if loader.config.General.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q", loader.config.General.LogLevel, "debug")
	}

	if loader.config.Timeouts.Idle.Duration != 2*time.Minute {
		t.Errorf("Idle = %v, want %v", loader.config.Timeouts.Idle.Duration, 2*time.Minute)
	}

	if loader.config.Process.CaptureMode != "stderr" {
		t.Errorf("CaptureMode = %q, want %q", loader.config.Process.CaptureMode, "stderr")
	}

	// Check adapter was merged
	if _, ok := loader.config.Adapters["test-adapter"]; !ok {
		t.Error("adapter config not merged")
	}

	// Check defaults preserved
	if loader.config.Remote.Transport != "nats" {
		t.Errorf("Transport = %q, want %q (should be preserved)", loader.config.Remote.Transport, "nats")
	}
}

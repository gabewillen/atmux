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

func TestIsSensitiveKey(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{"api_key", true},
		{"API_KEY", true},
		{"token", true},
		{"auth_token", true},
		{"secret", true},
		{"password", true},
		{"creds", true},
		{"seed", true},
		{"private_key", true},
		{"access_key", true},
		{"my_api_key_value", true},
		{"some_secret_data", true},
		{"name", false},
		{"endpoint", false},
		{"log_level", false},
		{"url", false},
		{"host", false},
		{"port", false},
		{"mode", false},
		{"enabled", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := isSensitiveKey(tt.key); got != tt.expected {
				t.Errorf("isSensitiveKey(%q) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestRedactedAdaptersNil(t *testing.T) {
	result := RedactedAdapters(nil)
	if result != nil {
		t.Errorf("RedactedAdapters(nil) = %v, want nil", result)
	}
}

func TestRedactedAdaptersEmpty(t *testing.T) {
	result := RedactedAdapters(map[string]any{})
	if result == nil {
		t.Error("RedactedAdapters(empty) = nil, want non-nil empty map")
	}
	if len(result) != 0 {
		t.Errorf("RedactedAdapters(empty) len = %d, want 0", len(result))
	}
}

func TestRedactedAdaptersWithSensitiveKeys(t *testing.T) {
	adapters := map[string]any{
		"claude-code": map[string]any{
			"api_key":  "sk-secret-12345",
			"endpoint": "https://api.example.com",
			"token":    "tok-abcdef",
			"name":     "my-adapter",
		},
		"cursor": map[string]any{
			"password":    "hunter2",
			"auth_token":  "at-xyz",
			"host":        "localhost",
			"private_key": "-----BEGIN-----",
		},
	}

	redacted := RedactedAdapters(adapters)

	// Check claude-code adapter
	cc, ok := redacted["claude-code"].(map[string]any)
	if !ok {
		t.Fatal("claude-code adapter not found in redacted output")
	}
	if cc["api_key"] != "[REDACTED]" {
		t.Errorf("api_key = %v, want [REDACTED]", cc["api_key"])
	}
	if cc["endpoint"] != "https://api.example.com" {
		t.Errorf("endpoint = %v, want https://api.example.com", cc["endpoint"])
	}
	if cc["token"] != "[REDACTED]" {
		t.Errorf("token = %v, want [REDACTED]", cc["token"])
	}
	if cc["name"] != "my-adapter" {
		t.Errorf("name = %v, want my-adapter", cc["name"])
	}

	// Check cursor adapter
	cur, ok := redacted["cursor"].(map[string]any)
	if !ok {
		t.Fatal("cursor adapter not found in redacted output")
	}
	if cur["password"] != "[REDACTED]" {
		t.Errorf("password = %v, want [REDACTED]", cur["password"])
	}
	if cur["auth_token"] != "[REDACTED]" {
		t.Errorf("auth_token = %v, want [REDACTED]", cur["auth_token"])
	}
	if cur["host"] != "localhost" {
		t.Errorf("host = %v, want localhost", cur["host"])
	}
	if cur["private_key"] != "[REDACTED]" {
		t.Errorf("private_key = %v, want [REDACTED]", cur["private_key"])
	}

	// Original should be unchanged
	origCC, _ := adapters["claude-code"].(map[string]any)
	if origCC["api_key"] != "sk-secret-12345" {
		t.Error("original adapters map was mutated")
	}
}

func TestRedactedAdaptersNestedMaps(t *testing.T) {
	adapters := map[string]any{
		"adapter1": map[string]any{
			"nested": map[string]any{
				"api_key": "should-be-redacted",
				"url":     "should-be-visible",
			},
		},
	}

	redacted := RedactedAdapters(adapters)
	a1, _ := redacted["adapter1"].(map[string]any)
	nested, _ := a1["nested"].(map[string]any)

	if nested["api_key"] != "[REDACTED]" {
		t.Errorf("nested api_key = %v, want [REDACTED]", nested["api_key"])
	}
	if nested["url"] != "should-be-visible" {
		t.Errorf("nested url = %v, want should-be-visible", nested["url"])
	}
}

func TestRedactedAdaptersNonStringValues(t *testing.T) {
	adapters := map[string]any{
		"adapter1": map[string]any{
			"api_key": 12345,
			"port":    8080,
		},
	}

	redacted := RedactedAdapters(adapters)
	a1, _ := redacted["adapter1"].(map[string]any)

	// Non-string sensitive keys should still be redacted
	if a1["api_key"] != "[REDACTED]" {
		t.Errorf("non-string api_key = %v, want [REDACTED]", a1["api_key"])
	}
	// Non-sensitive non-string values should be preserved
	if a1["port"] != 8080 {
		t.Errorf("port = %v, want 8080", a1["port"])
	}
}

func TestDurationMarshalText(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"zero", 0, "0s"},
		{"seconds", 30 * time.Second, "30s"},
		{"minutes", 5 * time.Minute, "5m0s"},
		{"hours", time.Hour, "1h0m0s"},
		{"milliseconds", 100 * time.Millisecond, "100ms"},
		{"complex", 90 * time.Minute, "1h30m0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Duration{Duration: tt.duration}
			data, err := d.MarshalText()
			if err != nil {
				t.Errorf("MarshalText() error: %v", err)
				return
			}
			if string(data) != tt.expected {
				t.Errorf("MarshalText() = %q, want %q", string(data), tt.expected)
			}
		})
	}
}

func TestDurationRoundTrip(t *testing.T) {
	// Verify that unmarshal then marshal produces a value that can be unmarshaled again
	// to the same duration.
	inputs := []string{"30s", "5m", "1h", "100ms", "1h30m"}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			var d Duration
			if err := d.UnmarshalText([]byte(input)); err != nil {
				t.Fatalf("UnmarshalText(%q) error: %v", input, err)
			}

			marshaled, err := d.MarshalText()
			if err != nil {
				t.Fatalf("MarshalText() error: %v", err)
			}

			var d2 Duration
			if err := d2.UnmarshalText(marshaled); err != nil {
				t.Fatalf("UnmarshalText(%q) error on round-trip: %v", string(marshaled), err)
			}

			if d.Duration != d2.Duration {
				t.Errorf("round-trip mismatch: %v != %v", d.Duration, d2.Duration)
			}
		})
	}
}

func TestByteSizeRoundTrip(t *testing.T) {
	// Verify that marshal then unmarshal produces the same value.
	inputs := []int64{
		0,
		500,
		1024,
		64 * 1024,
		10 * 1024 * 1024,
		1024 * 1024 * 1024,
	}

	for _, input := range inputs {
		t.Run("", func(t *testing.T) {
			b := ByteSize{Bytes: input}
			marshaled, err := b.MarshalText()
			if err != nil {
				t.Fatalf("MarshalText() error: %v", err)
			}

			var b2 ByteSize
			if err := b2.UnmarshalText(marshaled); err != nil {
				t.Fatalf("UnmarshalText(%q) error: %v", string(marshaled), err)
			}

			if b.Bytes != b2.Bytes {
				t.Errorf("round-trip mismatch: %d != %d (marshaled as %q)", b.Bytes, b2.Bytes, string(marshaled))
			}
		})
	}
}

func TestDefaultConfigAdditionalDefaults(t *testing.T) {
	cfg := DefaultConfig()

	// General
	if cfg.General.LogFormat != "text" {
		t.Errorf("LogFormat = %q, want %q", cfg.General.LogFormat, "text")
	}

	// Timeouts
	if cfg.Timeouts.Stuck.Duration != 5*time.Minute {
		t.Errorf("Stuck = %v, want %v", cfg.Timeouts.Stuck.Duration, 5*time.Minute)
	}

	// Process
	if cfg.Process.CaptureMode != "all" {
		t.Errorf("CaptureMode = %q, want %q", cfg.Process.CaptureMode, "all")
	}
	if cfg.Process.StreamBufferSize.Bytes != 1024*1024 {
		t.Errorf("StreamBufferSize = %d, want %d", cfg.Process.StreamBufferSize.Bytes, 1024*1024)
	}
	if cfg.Process.HookMode != "auto" {
		t.Errorf("HookMode = %q, want %q", cfg.Process.HookMode, "auto")
	}
	if cfg.Process.PollInterval.Duration != 100*time.Millisecond {
		t.Errorf("PollInterval = %v, want %v", cfg.Process.PollInterval.Duration, 100*time.Millisecond)
	}
	if cfg.Process.HookSocketDir != "/tmp" {
		t.Errorf("HookSocketDir = %q, want %q", cfg.Process.HookSocketDir, "/tmp")
	}

	// Git
	if cfg.Git.Merge.Strategy != "squash" {
		t.Errorf("Merge.Strategy = %q, want %q", cfg.Git.Merge.Strategy, "squash")
	}
	if cfg.Git.Merge.AllowDirty {
		t.Error("Merge.AllowDirty should be false by default")
	}

	// Shutdown
	if cfg.Shutdown.DrainTimeout.Duration != 30*time.Second {
		t.Errorf("DrainTimeout = %v, want %v", cfg.Shutdown.DrainTimeout.Duration, 30*time.Second)
	}
	if cfg.Shutdown.CleanupWorktrees {
		t.Error("CleanupWorktrees should be false by default")
	}

	// Events
	if cfg.Events.BatchWindow.Duration != 50*time.Millisecond {
		t.Errorf("BatchWindow = %v, want %v", cfg.Events.BatchWindow.Duration, 50*time.Millisecond)
	}
	if cfg.Events.BatchMaxEvents != 100 {
		t.Errorf("BatchMaxEvents = %d, want 100", cfg.Events.BatchMaxEvents)
	}
	if cfg.Events.BatchMaxBytes.Bytes != 64*1024 {
		t.Errorf("BatchMaxBytes = %d, want %d", cfg.Events.BatchMaxBytes.Bytes, 64*1024)
	}
	if !cfg.Events.Coalesce.IOStreams {
		t.Error("Coalesce.IOStreams should be true by default")
	}
	if !cfg.Events.Coalesce.Presence {
		t.Error("Coalesce.Presence should be true by default")
	}
	if !cfg.Events.Coalesce.Activity {
		t.Error("Coalesce.Activity should be true by default")
	}

	// Remote
	if cfg.Remote.BufferSize.Bytes != 10*1024*1024 {
		t.Errorf("BufferSize = %d, want %d", cfg.Remote.BufferSize.Bytes, 10*1024*1024)
	}
	if cfg.Remote.RequestTimeout.Duration != 5*time.Second {
		t.Errorf("RequestTimeout = %v, want %v", cfg.Remote.RequestTimeout.Duration, 5*time.Second)
	}
	if cfg.Remote.ReconnectMaxAttempts != 10 {
		t.Errorf("ReconnectMaxAttempts = %d, want 10", cfg.Remote.ReconnectMaxAttempts)
	}
	if cfg.Remote.NATS.SubjectPrefix != "amux" {
		t.Errorf("SubjectPrefix = %q, want %q", cfg.Remote.NATS.SubjectPrefix, "amux")
	}
	if cfg.Remote.NATS.KVBucket != "AMUX_KV" {
		t.Errorf("KVBucket = %q, want %q", cfg.Remote.NATS.KVBucket, "AMUX_KV")
	}
	if cfg.Remote.Manager.Model != "lfm2.5-thinking" {
		t.Errorf("Manager.Model = %q, want %q", cfg.Remote.Manager.Model, "lfm2.5-thinking")
	}

	// NATS
	if cfg.NATS.Mode != "embedded" {
		t.Errorf("NATS.Mode = %q, want %q", cfg.NATS.Mode, "embedded")
	}
	if cfg.NATS.Topology != "hub" {
		t.Errorf("NATS.Topology = %q, want %q", cfg.NATS.Topology, "hub")
	}
	if cfg.NATS.Listen != "0.0.0.0:4222" {
		t.Errorf("NATS.Listen = %q, want %q", cfg.NATS.Listen, "0.0.0.0:4222")
	}

	// Daemon
	if cfg.Daemon.SocketPath != "~/.amux/amuxd.sock" {
		t.Errorf("Daemon.SocketPath = %q, want %q", cfg.Daemon.SocketPath, "~/.amux/amuxd.sock")
	}
	if !cfg.Daemon.AutoStart {
		t.Error("Daemon.AutoStart should be true by default")
	}

	// Plugins
	if cfg.Plugins.Dir != "~/.config/amux/plugins" {
		t.Errorf("Plugins.Dir = %q, want %q", cfg.Plugins.Dir, "~/.config/amux/plugins")
	}

	// Telemetry
	if cfg.Telemetry.Enabled {
		t.Error("Telemetry.Enabled should be false by default")
	}
	if cfg.Telemetry.ServiceName != "amux" {
		t.Errorf("Telemetry.ServiceName = %q, want %q", cfg.Telemetry.ServiceName, "amux")
	}
	if cfg.Telemetry.Exporter.Protocol != "grpc" {
		t.Errorf("Telemetry.Exporter.Protocol = %q, want %q", cfg.Telemetry.Exporter.Protocol, "grpc")
	}

	// Adapters and Agents initialized
	if cfg.Adapters == nil {
		t.Error("Adapters should be non-nil empty map")
	}
	if cfg.Agents == nil {
		t.Error("Agents should be non-nil empty slice")
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

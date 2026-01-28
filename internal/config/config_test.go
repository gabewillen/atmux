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

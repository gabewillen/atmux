package remote

import (
	"testing"

	"github.com/agentflare-ai/amux/internal/config"
)

func TestNewManager(t *testing.T) {
	cfg := &config.RemoteConfig{}
	m := NewManager(cfg, "devbox")
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.hostID != "devbox" {
		t.Errorf("hostID = %q, want devbox", m.hostID)
	}
}

func TestParseBufferSize(t *testing.T) {
	tests := []struct {
		s    string
		want int
	}{
		{"", 10 * 1024 * 1024},
		{"10MB", 10 * 1024 * 1024},
		{"1MB", 1024 * 1024},
		{"64KB", 64 * 1024},
		{"1KB", 1024},
		{"100", 100},
	}
	for _, tt := range tests {
		got := parseBufferSize(tt.s)
		if got != tt.want {
			t.Errorf("parseBufferSize(%q) = %d, want %d", tt.s, got, tt.want)
		}
	}
}

func TestManager_IsHandshakeDone(t *testing.T) {
	m := NewManager(nil, "host1")
	if m.IsHandshakeDone() {
		t.Error("IsHandshakeDone want false initially")
	}
}

func TestManager_Close(t *testing.T) {
	m := NewManager(nil, "host1")
	m.Close()
	if m.nc != nil {
		t.Error("Close should clear nc")
	}
}

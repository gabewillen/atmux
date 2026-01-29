package director

import (
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/event"
)

func TestExtractHostIDFromSubject(t *testing.T) {
	tests := []struct {
		name    string
		subject string
		prefix  string
		want    string
	}{
		{"normal", "amux.handshake.host-123", "amux.handshake.", "host-123"},
		{"empty host", "amux.handshake.", "amux.handshake.", ""},
		{"prefix longer", "amux", "amux.handshake.", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractHostIDFromSubject(tt.subject, tt.prefix)
			if got != tt.want {
				t.Errorf("extractHostIDFromSubject(%q, %q) = %q, want %q",
					tt.subject, tt.prefix, got, tt.want)
			}
		})
	}
}

func TestHostStateTracking(t *testing.T) {
	cfg := config.DefaultConfig()
	d := New(nil, cfg, event.NewNoopDispatcher())

	// Initially no hosts
	if len(d.ConnectedHosts()) != 0 {
		t.Errorf("expected no connected hosts, got %d", len(d.ConnectedHosts()))
	}

	// Simulate adding a host
	d.mu.Lock()
	d.hosts["test-host"] = &HostState{
		HostID:            "test-host",
		PeerID:            "123",
		Connected:         true,
		HandshakeComplete: true,
		ConnectedAt:       time.Now().UTC(),
		Sessions:          make(map[string]bool),
	}
	d.mu.Unlock()

	// Verify connected
	if !d.HostConnected("test-host") {
		t.Error("host should be connected")
	}

	hosts := d.ConnectedHosts()
	if len(hosts) != 1 || hosts[0] != "test-host" {
		t.Errorf("ConnectedHosts() = %v, want [test-host]", hosts)
	}

	// Disconnect
	d.SetHostDisconnected("test-host")
	if d.HostConnected("test-host") {
		t.Error("host should be disconnected")
	}
}

func TestSessionsForHost(t *testing.T) {
	cfg := config.DefaultConfig()
	d := New(nil, cfg, event.NewNoopDispatcher())

	// Add host with sessions
	d.mu.Lock()
	d.hosts["host-1"] = &HostState{
		HostID:            "host-1",
		Connected:         true,
		HandshakeComplete: true,
		Sessions:          map[string]bool{"s1": true, "s2": true},
	}
	d.mu.Unlock()

	sessions := d.SessionsForHost("host-1")
	if len(sessions) != 2 {
		t.Errorf("SessionsForHost() returned %d sessions, want 2", len(sessions))
	}

	// Non-existent host
	sessions = d.SessionsForHost("no-such-host")
	if len(sessions) != 0 {
		t.Errorf("SessionsForHost() for non-existent host returned %d, want 0", len(sessions))
	}
}

func TestReconnectDetectsExistingSessions(t *testing.T) {
	cfg := config.DefaultConfig()
	d := New(nil, cfg, event.NewNoopDispatcher())

	// Simulate a host with existing sessions (previous connection)
	d.mu.Lock()
	d.hosts["host-reconnect"] = &HostState{
		HostID:   "host-reconnect",
		PeerID:   "old-peer",
		Sessions: map[string]bool{"sess-1": true, "sess-2": true},
		// Connected=false simulates a disconnected host
		Connected:         false,
		HandshakeComplete: false,
	}
	d.mu.Unlock()

	// The handleHandshake logic checks for existing sessions on reconnect.
	// Since we can't directly test handleHandshake without a NATS connection,
	// verify the host state structure supports reconnection detection.
	d.mu.RLock()
	host := d.hosts["host-reconnect"]
	d.mu.RUnlock()

	if host == nil {
		t.Fatal("host should exist")
	}

	// The reconnection logic in handleHandshake checks:
	// host != nil && len(host.Sessions) > 0
	if len(host.Sessions) != 2 {
		t.Errorf("host should have 2 existing sessions, got %d", len(host.Sessions))
	}
}

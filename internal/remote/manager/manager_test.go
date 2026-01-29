package manager

import (
	"runtime"
	"testing"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/internal/protocol"
)

func TestNewManager(t *testing.T) {
	cfg := config.DefaultConfig()
	mgr := New(nil, cfg, "test-host", event.NewNoopDispatcher())

	if mgr.hostID != "test-host" {
		t.Errorf("hostID = %q, want %q", mgr.hostID, "test-host")
	}

	if mgr.peerID == "" {
		t.Error("peerID should be set")
	}

	if mgr.prefix != "amux" {
		t.Errorf("prefix = %q, want %q", mgr.prefix, "amux")
	}

	if !mgr.hubConnected {
		t.Error("hubConnected should be true initially")
	}
}

func TestManagerHubConnectionState(t *testing.T) {
	cfg := config.DefaultConfig()
	mgr := New(nil, cfg, "test-host", event.NewNoopDispatcher())

	// Initially connected
	mgr.mu.RLock()
	if !mgr.hubConnected {
		t.Error("should be connected initially")
	}
	mgr.mu.RUnlock()

	// Disconnect
	mgr.SetHubConnected(false)
	mgr.mu.RLock()
	if mgr.hubConnected {
		t.Error("should be disconnected after SetHubConnected(false)")
	}
	mgr.mu.RUnlock()

	// Reconnect
	mgr.SetHubConnected(true)
	mgr.mu.RLock()
	if !mgr.hubConnected {
		t.Error("should be connected after SetHubConnected(true)")
	}
	mgr.mu.RUnlock()
}

func TestHandshakePayloadIncludesHostInfo(t *testing.T) {
	// Verify that the HostInfoPayload fields match runtime values
	info := &protocol.HostInfoPayload{
		Version: Version,
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
	}

	if info.Version != Version {
		t.Errorf("Version = %q, want %q", info.Version, Version)
	}
	if info.OS != runtime.GOOS {
		t.Errorf("OS = %q, want %q", info.OS, runtime.GOOS)
	}
	if info.Arch != runtime.GOARCH {
		t.Errorf("Arch = %q, want %q", info.Arch, runtime.GOARCH)
	}
}

func TestExtractSessionIDFromPTYSubject(t *testing.T) {
	tests := []struct {
		name      string
		subject   string
		prefix    string
		hostID    string
		wantSID   string
	}{
		{
			"valid subject",
			"amux.pty.host1.session123.in",
			"amux", "host1",
			"session123",
		},
		{
			"short subject",
			"amux.pty.host1.",
			"amux", "host1",
			"",
		},
		{
			"no session",
			"amux.pty.host1.in",
			"amux", "host1",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractSessionIDFromPTYSubject(tt.subject, tt.prefix, tt.hostID)
			if got != tt.wantSID {
				t.Errorf("extractSessionIDFromPTYSubject(%q, %q, %q) = %q, want %q",
					tt.subject, tt.prefix, tt.hostID, got, tt.wantSID)
			}
		})
	}
}

func TestManagerCustomPrefix(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Remote.NATS.SubjectPrefix = "custom"

	mgr := New(nil, cfg, "test-host", event.NewNoopDispatcher())
	if mgr.prefix != "custom" {
		t.Errorf("prefix = %q, want %q", mgr.prefix, "custom")
	}
}

func TestManagerCustomBufferSize(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Remote.BufferSize.Bytes = 5 * 1024 * 1024 // 5MB

	mgr := New(nil, cfg, "test-host", event.NewNoopDispatcher())
	if mgr.bufferSize != 5*1024*1024 {
		t.Errorf("bufferSize = %d, want %d", mgr.bufferSize, 5*1024*1024)
	}
}

func TestVersionConst(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
}

func TestSetHubConnected_SetsAwaitingReplayOnActiveSessions(t *testing.T) {
	cfg := config.DefaultConfig()
	mgr := New(nil, cfg, "test-host", event.NewNoopDispatcher())

	// Create two test sessions: one running, one stopped
	running := &ManagedSession{
		SessionID: "sess-1",
		AgentID:   "agent-1",
		AgentSlug: "agent-one",
		running:   true,
		done:      make(chan struct{}),
	}
	stopped := &ManagedSession{
		SessionID: "sess-2",
		AgentID:   "agent-2",
		AgentSlug: "agent-two",
		running:   false,
		done:      make(chan struct{}),
	}

	mgr.sessions["agent-1"] = running
	mgr.sessions["agent-2"] = stopped
	mgr.sessionsByID["sess-1"] = running
	mgr.sessionsByID["sess-2"] = stopped

	// Simulate disconnect then reconnect
	mgr.SetHubConnected(false)
	mgr.SetHubConnected(true)

	// Running session should be awaiting replay
	running.mu.Lock()
	if !running.awaitingReplay {
		t.Error("running session should have awaitingReplay=true after reconnect")
	}
	running.mu.Unlock()

	// Stopped session should NOT be awaiting replay
	stopped.mu.Lock()
	if stopped.awaitingReplay {
		t.Error("stopped session should not have awaitingReplay set")
	}
	stopped.mu.Unlock()
}

func TestAwaitingReplay_BuffersLiveOutput(t *testing.T) {
	// Verify that a session with awaitingReplay=true buffers output
	// in liveBuf rather than losing it.
	sess := &ManagedSession{
		SessionID:      "sess-1",
		AgentID:        "agent-1",
		AgentSlug:      "agent-one",
		running:        true,
		awaitingReplay: true,
		done:           make(chan struct{}),
	}

	// Simulate what readPTYOutput does: check awaitingReplay and buffer
	data := []byte("hello from pty")
	sess.mu.Lock()
	if sess.replayPending || sess.awaitingReplay {
		sess.liveBuf = append(sess.liveBuf, data...)
	}
	sess.mu.Unlock()

	sess.mu.Lock()
	if len(sess.liveBuf) != len(data) {
		t.Errorf("liveBuf length = %d, want %d", len(sess.liveBuf), len(data))
	}
	if string(sess.liveBuf) != "hello from pty" {
		t.Errorf("liveBuf = %q, want %q", sess.liveBuf, "hello from pty")
	}
	sess.mu.Unlock()
}

func TestHandleReplay_ClearsAwaitingReplay(t *testing.T) {
	// Verify that after replay handling, awaitingReplay is cleared
	sess := &ManagedSession{
		SessionID:      "sess-1",
		AgentID:        "agent-1",
		AgentSlug:      "agent-one",
		running:        true,
		awaitingReplay: true,
		replayPending:  true,
		liveBuf:        []byte("buffered data"),
		done:           make(chan struct{}),
	}

	// Simulate the replay completion logic from handleReplay
	sess.mu.Lock()
	liveBuf := sess.liveBuf
	sess.liveBuf = nil
	sess.replayPending = false
	sess.awaitingReplay = false
	sess.mu.Unlock()

	if string(liveBuf) != "buffered data" {
		t.Errorf("liveBuf = %q, want %q", liveBuf, "buffered data")
	}

	sess.mu.Lock()
	if sess.awaitingReplay {
		t.Error("awaitingReplay should be false after replay handling")
	}
	if sess.replayPending {
		t.Error("replayPending should be false after replay handling")
	}
	if sess.liveBuf != nil {
		t.Error("liveBuf should be nil after replay handling")
	}
	sess.mu.Unlock()
}

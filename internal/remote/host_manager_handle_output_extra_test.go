package remote

import (
	"testing"

	"github.com/agentflare-ai/amux/pkg/api"
)

func TestHandleOutputReplayGateSkipsPublish(t *testing.T) {
	dispatcher := &rawDispatcher{}
	manager := &HostManager{
		dispatcher:    dispatcher,
		subjectPrefix: "amux",
		hostID:        api.MustParseHostID("host"),
		connected:     true,
	}
	sessionID := api.NewSessionID()
	session := &remoteSession{
		sessionID: sessionID,
		buffer:    NewReplayBuffer(16),
		replayGate: true,
	}
	manager.handleOutput(session, []byte("data"))
	if len(dispatcher.rawSubjects) != 0 {
		t.Fatalf("expected no publish while gated")
	}
	if snap := session.buffer.Snapshot(); len(snap) == 0 {
		t.Fatalf("expected buffer to store data")
	}
}

func TestHandleOutputReplayPending(t *testing.T) {
	dispatcher := &rawDispatcher{}
	manager := &HostManager{
		dispatcher:    dispatcher,
		subjectPrefix: "amux",
		hostID:        api.MustParseHostID("host"),
		connected:     true,
	}
	sessionID := api.NewSessionID()
	session := &remoteSession{
		sessionID: sessionID,
		buffer:    NewReplayBuffer(16),
		replaying: true,
	}
	manager.handleOutput(session, []byte("chunk"))
	if len(dispatcher.rawSubjects) != 0 {
		t.Fatalf("expected no publish while replaying")
	}
	session.mu.Lock()
	pending := len(session.pending)
	session.mu.Unlock()
	if pending != 1 {
		t.Fatalf("expected pending chunk")
	}
}

func TestHandleOutputPublishesPTY(t *testing.T) {
	dispatcher := &rawDispatcher{}
	manager := &HostManager{
		dispatcher:    dispatcher,
		subjectPrefix: "amux",
		hostID:        api.MustParseHostID("host"),
		connected:     true,
	}
	sessionID := api.NewSessionID()
	session := &remoteSession{
		sessionID: sessionID,
		buffer:    NewReplayBuffer(16),
	}
	manager.handleOutput(session, []byte("data"))
	if len(dispatcher.rawSubjects) == 0 {
		t.Fatalf("expected PTY publish")
	}
}

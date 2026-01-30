package remote

import (
	"testing"

	"github.com/agentflare-ai/amux/pkg/api"
)

func TestHandleReplayPublishes(t *testing.T) {
	dispatcher := &rawDispatcher{}
	sessionID := api.NewSessionID()
	buffer := NewReplayBuffer(16)
	buffer.Add([]byte("hello"))
	manager := &HostManager{
		subjectPrefix: "amux",
		hostID:        api.MustParseHostID("host"),
		dispatcher:    dispatcher,
		outbox:        NewOutbox(1024),
		connected:     true,
		sessions: map[api.SessionID]*remoteSession{
			sessionID: {
				sessionID: sessionID,
				buffer:    buffer,
				pending:   [][]byte{[]byte("pending")},
			},
		},
	}
	req, err := EncodePayload("replay", ReplayRequest{SessionID: sessionID.String()})
	if err != nil {
		t.Fatalf("encode replay: %v", err)
	}
	manager.handleReplay("reply.replay", req)
	found := false
	for _, subject := range dispatcher.rawSubjects {
		if subject == PtyOutSubject("amux", manager.hostID, sessionID) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected PTY publish")
	}
}

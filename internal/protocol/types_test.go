package protocol

import (
	"encoding/json"
	"testing"

	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go/muid"
)

func TestSubjectGeneration(t *testing.T) {
	prefix := "amux"
	hostID := api.HostID("test-host")
	sessionID := api.SessionID(muid.Make())

	if s := SubjectForHandshake(prefix, hostID); s != "amux.handshake.test-host" {
		t.Errorf("Handshake subject mismatch: %s", s)
	}
	if s := SubjectForCtl(prefix, hostID); s != "amux.ctl.test-host" {
		t.Errorf("Ctl subject mismatch: %s", s)
	}
	if s := SubjectForEvents(prefix, hostID); s != "amux.events.test-host" {
		t.Errorf("Events subject mismatch: %s", s)
	}
	if s := SubjectForPTYOut(prefix, hostID, sessionID); s != "amux.pty.test-host."+sessionID.String()+".out" {
		t.Errorf("PTY Out subject mismatch: %s", s)
	}
}

func TestHandshakeRequest_JSON(t *testing.T) {
	req := HandshakeRequest{
		Protocol: 1,
		PeerID:   api.PeerID(muid.Make()),
		Role:     "daemon",
		HostID:   api.HostID("host-1"),
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var req2 HandshakeRequest
	if err := json.Unmarshal(data, &req2); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if req.HostID != req2.HostID {
		t.Errorf("HostID mismatch")
	}
	if req.PeerID != req2.PeerID {
		t.Errorf("PeerID mismatch")
	}
}

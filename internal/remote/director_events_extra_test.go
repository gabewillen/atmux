package remote

import (
	"encoding/json"
	"testing"

	"github.com/agentflare-ai/amux/pkg/api"
)

func TestDirectorHandleHostEventConnectionStates(t *testing.T) {
	host := api.MustParseHostID("host")
	state := &hostState{hostID: host, connected: false, ready: false}
	director := &Director{
		subjectPrefix: "amux",
		hosts:         map[api.HostID]*hostState{host: state},
	}
	established, err := EncodeEventMessage("connection.established", ConnectionEstablishedPayload{
		PeerID:    api.NewPeerID().String(),
		Timestamp: NowRFC3339(),
	})
	if err != nil {
		t.Fatalf("encode event: %v", err)
	}
	establishedJSON, err := json.Marshal(established)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	director.handleHostEvent(protocolMessage(EventsSubject("amux", host), establishedJSON))
	if !state.connected || !state.ready {
		t.Fatalf("expected connected/ready after established")
	}
	lost, err := EncodeEventMessage("connection.lost", ConnectionLostPayload{
		PeerID:    api.NewPeerID().String(),
		Timestamp: NowRFC3339(),
		Reason:    "test",
	})
	if err != nil {
		t.Fatalf("encode event: %v", err)
	}
	lostJSON, err := json.Marshal(lost)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	director.handleHostEvent(protocolMessage(EventsSubject("amux", host), lostJSON))
	if state.connected || state.ready {
		t.Fatalf("expected disconnected/not ready after lost")
	}
}

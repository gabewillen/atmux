package remote

import (
	"testing"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

func decodeErrorPayload(t *testing.T, data []byte) ErrorPayload {
	t.Helper()
	ctrl, err := DecodeControlMessage(data)
	if err != nil {
		t.Fatalf("decode control: %v", err)
	}
	var payload ErrorPayload
	if err := DecodePayload(ctrl, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	return payload
}

func TestDirectorHandleHandshakeUnsupportedProtocol(t *testing.T) {
	dispatcher := &recordRawDispatcher{}
	director := &Director{}
	setUnexportedField(director, "dispatcher", dispatcher)
	setUnexportedField(director, "subjectPrefix", "amux")
	host := api.MustParseHostID("host")
	payload := HandshakePayload{Protocol: 2, PeerID: api.NewPeerID().String(), HostID: host.String()}
	msg, err := EncodePayload("handshake", payload)
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}
	data, err := EncodeControlMessage(msg)
	if err != nil {
		t.Fatalf("encode control: %v", err)
	}
	director.handleHandshake(protocol.Message{Subject: HandshakeSubject("amux", host), Reply: "reply", Data: data})
	errPayload := decodeErrorPayload(t, dispatcher.lastPayload)
	if errPayload.Code != "unsupported_protocol" {
		t.Fatalf("unexpected error code: %s", errPayload.Code)
	}
}

func TestDirectorHandleHandshakePeerConflict(t *testing.T) {
	dispatcher := &recordRawDispatcher{}
	director := &Director{}
	setUnexportedField(director, "dispatcher", dispatcher)
	setUnexportedField(director, "subjectPrefix", "amux")
	host := api.MustParseHostID("host")
	peer := api.NewPeerID()
	other := api.MustParseHostID("other")
	setUnexportedField(director, "hosts", map[api.HostID]*hostState{
		other: {hostID: other, peerID: peer, connected: false},
	})
	setUnexportedField(director, "peerIndex", map[string]api.HostID{peer.String(): other})
	payload := HandshakePayload{Protocol: 1, PeerID: peer.String(), HostID: host.String()}
	msg, err := EncodePayload("handshake", payload)
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}
	data, err := EncodeControlMessage(msg)
	if err != nil {
		t.Fatalf("encode control: %v", err)
	}
	director.handleHandshake(protocol.Message{Subject: HandshakeSubject("amux", host), Reply: "reply", Data: data})
	errPayload := decodeErrorPayload(t, dispatcher.lastPayload)
	if errPayload.Code != "peer_conflict" {
		t.Fatalf("unexpected error code: %s", errPayload.Code)
	}
}

func TestDirectorHandleHandshakeHostConflict(t *testing.T) {
	dispatcher := &recordRawDispatcher{}
	director := &Director{}
	setUnexportedField(director, "dispatcher", dispatcher)
	setUnexportedField(director, "subjectPrefix", "amux")
	host := api.MustParseHostID("host")
	existingPeer := api.NewPeerID()
	newPeer := api.NewPeerID()
	setUnexportedField(director, "hosts", map[api.HostID]*hostState{
		host: {hostID: host, peerID: existingPeer, connected: true, ready: true},
	})
	setUnexportedField(director, "peerIndex", map[string]api.HostID{existingPeer.String(): host})
	payload := HandshakePayload{Protocol: 1, PeerID: newPeer.String(), HostID: host.String()}
	msg, err := EncodePayload("handshake", payload)
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}
	data, err := EncodeControlMessage(msg)
	if err != nil {
		t.Fatalf("encode control: %v", err)
	}
	director.handleHandshake(protocol.Message{Subject: HandshakeSubject("amux", host), Reply: "reply", Data: data})
	errPayload := decodeErrorPayload(t, dispatcher.lastPayload)
	if errPayload.Code != "host_conflict" {
		t.Fatalf("unexpected error code: %s", errPayload.Code)
	}
}

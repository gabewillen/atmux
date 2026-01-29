package protocol

import (
	"testing"
	"time"

	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go/muid"
)

func TestHsmNet(t *testing.T) {
	selfID := api.PeerID(muid.Make())
	net := NewHsmNet(selfID)
	
	peerID := api.PeerID(muid.Make())
	peer := &Peer{
		ID:        peerID,
		Role:      "manager",
		Connected: true,
		LastSeen:  time.Now(),
	}
	
	net.RegisterPeer(peer)
	
	p, ok := net.GetPeer(peerID)
	if !ok {
		t.Fatal("Peer not found")
	}
	if p.ID != peerID {
		t.Errorf("ID mismatch")
	}
	
	peers := net.ListPeers()
	if len(peers) != 1 {
		t.Errorf("Expected 1 peer, got %d", len(peers))
	}
	
	net.MarkDisconnected(peerID)
	p, _ = net.GetPeer(peerID)
	if p.Connected {
		t.Error("Peer should be disconnected")
	}
}

package protocol

import (
	"sync"
	"time"

	"github.com/agentflare-ai/amux/pkg/api"
)

// Peer represents a remote node in the hsmnet.
type Peer struct {
	ID        api.PeerID
	Role      string
	HostID    api.HostID
	LastSeen  time.Time
	Connected bool
}

// HsmNet manages peers and routing.
type HsmNet struct {
	mu    sync.RWMutex
	peers map[api.PeerID]*Peer
	Self  api.PeerID
}

// NewHsmNet creates a new HsmNet instance.
func NewHsmNet(self api.PeerID) *HsmNet {
	return &HsmNet{
		peers: make(map[api.PeerID]*Peer),
		Self:  self,
	}
}

// RegisterPeer adds or updates a peer.
func (n *HsmNet) RegisterPeer(p *Peer) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.peers[p.ID] = p
}

// GetPeer retrieves a peer by ID.
func (n *HsmNet) GetPeer(id api.PeerID) (*Peer, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	p, ok := n.peers[id]
	return p, ok
}

// ListPeers returns all known peers.
func (n *HsmNet) ListPeers() []*Peer {
	n.mu.RLock()
	defer n.mu.RUnlock()
	list := make([]*Peer, 0, len(n.peers))
	for _, p := range n.peers {
		list = append(list, p)
	}
	return list
}

// MarkDisconnected marks a peer as disconnected.
func (n *HsmNet) MarkDisconnected(id api.PeerID) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if p, ok := n.peers[id]; ok {
		p.Connected = false
	}
}

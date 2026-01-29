package protocol

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
)

// HsmNetDispatcher manages event routing.
type HsmNetDispatcher struct {
	mu            sync.RWMutex
	localBus      LocalBus
	natsConn      *nats.Conn
	peerID        api.PeerID
	hostID        api.HostID
	subjectPrefix string
	Network       *HsmNet
}

// LocalBus is an interface for the local event bus (e.g., internal/agent/bus.go).
type LocalBus interface {
	Publish(event interface{})
}

// NewDispatcher creates a new dispatcher.
func NewDispatcher(peerID api.PeerID, hostID api.HostID, nc *nats.Conn) *HsmNetDispatcher {
	return &HsmNetDispatcher{
		peerID:        peerID,
		hostID:        hostID,
		natsConn:      nc,
		subjectPrefix: "amux", // Default
		Network:       NewHsmNet(peerID),
	}
}

// SetLocalBus attaches the local bus.
func (d *HsmNetDispatcher) SetLocalBus(bus LocalBus) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.localBus = bus
}

// Dispatch sends an event.
func (d *HsmNetDispatcher) Dispatch(ctx context.Context, event EventMessage) error {
	// 1. Dispatch locally if applicable (broadcast or directed to self)
	// Simplified: If source is remote, always dispatch local?
	// If source is local, dispatch local + remote?
	
	// For now, always publish to local bus if set, assuming it handles filtering.
	d.mu.RLock()
	if d.localBus != nil {
		d.localBus.Publish(event)
	}
	d.mu.RUnlock()

	// 2. Dispatch remotely if NATS is connected
	if d.natsConn != nil && d.natsConn.IsConnected() {
		// Determine subject based on event type or destination?
		// Spec §9.1.5: Unicast/Multicast/Broadcast routes.
		// If event.Target is set (not in EventMessage struct yet), route there.
		// Standard EventMessage is broadcast to the host's event subject.
		
		subject := SubjectForEvents(d.subjectPrefix, d.hostID)
		
		data, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}
		
		if err := d.natsConn.Publish(subject, data); err != nil {
			return fmt.Errorf("failed to publish event: %w", err)
		}
	}

	return nil
}

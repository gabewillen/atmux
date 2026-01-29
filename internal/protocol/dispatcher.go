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
	mu          sync.RWMutex
	localBus    LocalBus // Simplified interface for local bus
	natsConn    *nats.Conn
	peerID      api.PeerID
	hostID      api.HostID
	subjectPrefix string
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
	// 1. Dispatch locally if applicable
	// For now, we assume local bus takes a different struct (BusEvent), so we might need mapping.
	// Or we dispatch the EventMessage directly if subscribers handle it.
	d.mu.RLock()
	if d.localBus != nil {
		d.localBus.Publish(event)
	}
	d.mu.RUnlock()

	// 2. Dispatch remotely if NATS is connected
	if d.natsConn != nil && d.natsConn.IsConnected() {
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

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
	subs          []*nats.Subscription
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

// Start subscribes to NATS subjects to receive remote events.
func (d *HsmNetDispatcher) Start(ctx context.Context) error {
	if d.natsConn == nil || !d.natsConn.IsConnected() {
		return nil // NATS optional or not ready
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// Subscribe to all events: P.events.>
	// Note: In a real system, we might be more selective (e.g. only events for this host, or all if director).
	// Spec says Director subscribes to P.comm.> and observes.
	// For hsmnet (EventMessage), let's subscribe to global events for now to ensure connectivity.
	// Spec §5.5.7.5: Daemon publishes to P.events.<host_id>.
	// If we are Director, we should listen to P.events.>
	// If we are Manager, maybe we don't listen to events?
	// But hsmnet implies a mesh.
	
	subject := fmt.Sprintf("%s.%s.>", d.subjectPrefix, SubjectEvents)
	sub, err := d.natsConn.Subscribe(subject, func(msg *nats.Msg) {
		var event EventMessage
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			// Log error
			return
		}
		
		// Avoid routing loops: if source is self, ignore
		// (PeerID check)
		if event.Source == d.peerID.String() {
			return
		}

		// Dispatch locally
		d.mu.RLock()
		if d.localBus != nil {
			d.localBus.Publish(event)
		}
		d.mu.RUnlock()
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to events: %w", err)
	}
	d.subs = append(d.subs, sub)
	
	return nil
}

// Stop closes subscriptions.
func (d *HsmNetDispatcher) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, sub := range d.subs {
		sub.Unsubscribe()
	}
	d.subs = nil
}

// Dispatch sends an event.
func (d *HsmNetDispatcher) Dispatch(ctx context.Context, event EventMessage) error {
	// 1. Dispatch locally if applicable
	d.mu.RLock()
	if d.localBus != nil {
		d.localBus.Publish(event)
	}
	d.mu.RUnlock()

	// 2. Dispatch remotely if NATS is connected
	if d.natsConn != nil && d.natsConn.IsConnected() {
		// Use the host's event subject
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

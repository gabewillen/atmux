// Package event - nats.go provides NATS-routed event dispatch.
//
// Per CLAUDE.md: "ALWAYS route all events through NATS subjects;
// never dispatch locally bypassing NATS."
//
// The NATSDispatcher publishes events to a configurable NATS subject
// prefix and subscribes for events on that same prefix. This ensures
// that events are visible to all nodes in the cluster.
package event

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/stateforward/hsm-go/muid"
)

// NATSDispatcher routes events through NATS subjects per spec §9.1.4.
//
// Events are published using the hsmnet subject hierarchy:
//
//	{prefix}.hsmnet.broadcast - for broadcast events (Target == 0)
//	{prefix}.hsmnet.{peer_id} - for unicast events (Target != 0)
//
// Subscribers receive events by subscribing to both the broadcast subject
// and their own peer-specific subject.
type NATSDispatcher struct {
	mu          sync.RWMutex
	nc          *nats.Conn
	prefix      string
	localPeerID string // base-10 encoded muid.MUID
	subscribers map[muid.MUID]*Subscription
	natsSubs    []*nats.Subscription
	closed      bool
}

// NATSDispatcherOption configures a NATSDispatcher.
type NATSDispatcherOption func(*NATSDispatcher)

// WithSubjectPrefix sets the NATS subject prefix (default: "amux").
func WithSubjectPrefix(prefix string) NATSDispatcherOption {
	return func(d *NATSDispatcher) {
		d.prefix = prefix
	}
}

// WithLocalPeerID sets the local peer ID for unicast event reception.
func WithLocalPeerID(peerID string) NATSDispatcherOption {
	return func(d *NATSDispatcher) {
		d.localPeerID = peerID
	}
}

// NewNATSDispatcher creates a dispatcher that routes events through NATS.
//
// The nc parameter must be an established NATS connection. Events are
// published to hsmnet subjects per spec §9.1.4:
//   - {prefix}.hsmnet.broadcast - for broadcast events
//   - {prefix}.hsmnet.{peer_id} - for unicast events
//
// The dispatcher subscribes to both broadcast and its own peer-specific subject.
func NewNATSDispatcher(nc *nats.Conn, opts ...NATSDispatcherOption) (*NATSDispatcher, error) {
	d := &NATSDispatcher{
		nc:          nc,
		prefix:      "amux",
		subscribers: make(map[muid.MUID]*Subscription),
	}

	for _, opt := range opts {
		opt(d)
	}

	// Generate a local peer ID if not provided
	if d.localPeerID == "" {
		d.localPeerID = fmt.Sprintf("%d", uint64(muid.Make()))
	}

	// Subscribe to broadcast events: {prefix}.hsmnet.broadcast
	broadcastSub, err := nc.Subscribe(d.prefix+".hsmnet.broadcast", d.handleMessage)
	if err != nil {
		return nil, fmt.Errorf("nats dispatcher subscribe broadcast: %w", err)
	}
	d.natsSubs = append(d.natsSubs, broadcastSub)

	// Subscribe to unicast events for this peer: {prefix}.hsmnet.{local_peer_id}
	unicastSub, err := nc.Subscribe(d.prefix+".hsmnet."+d.localPeerID, d.handleMessage)
	if err != nil {
		_ = broadcastSub.Unsubscribe()
		return nil, fmt.Errorf("nats dispatcher subscribe unicast: %w", err)
	}
	d.natsSubs = append(d.natsSubs, unicastSub)

	return d, nil
}

// Dispatch publishes an event to NATS per hsmnet subject hierarchy (§9.1.4).
//
// Broadcast events (Target == 0) are published to {prefix}.hsmnet.broadcast.
// Unicast events (Target != 0) are published to {prefix}.hsmnet.{target_peer_id}.
func (d *NATSDispatcher) Dispatch(ctx context.Context, evt Event) error {
	d.mu.RLock()
	if d.closed {
		d.mu.RUnlock()
		return ErrDispatcherClosed
	}
	d.mu.RUnlock()

	// Encode event as JSON
	data, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("nats dispatch marshal: %w", err)
	}

	// Determine subject based on target
	var subject string
	if evt.Target == 0 {
		// Broadcast: publish to {prefix}.hsmnet.broadcast
		subject = d.prefix + ".hsmnet.broadcast"
	} else {
		// Unicast: publish to {prefix}.hsmnet.{target_peer_id}
		subject = fmt.Sprintf("%s.hsmnet.%d", d.prefix, uint64(evt.Target))
	}

	if err := d.nc.Publish(subject, data); err != nil {
		return fmt.Errorf("nats dispatch publish: %w", err)
	}

	return nil
}

// Subscribe registers a handler for events received from NATS.
func (d *NATSDispatcher) Subscribe(sub Subscription) (unsubscribe func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if sub.ID == 0 {
		sub.ID = muid.Make()
	}

	d.subscribers[sub.ID] = &sub

	return func() {
		d.mu.Lock()
		defer d.mu.Unlock()
		delete(d.subscribers, sub.ID)
	}
}

// Close shuts down the NATS dispatcher and unsubscribes from all subjects.
func (d *NATSDispatcher) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.closed = true

	for _, sub := range d.natsSubs {
		_ = sub.Unsubscribe()
	}
	d.natsSubs = nil
	d.subscribers = nil

	return nil
}

// handleMessage processes incoming NATS event messages and dispatches
// them to matching local subscribers.
func (d *NATSDispatcher) handleMessage(msg *nats.Msg) {
	var evt Event
	if err := json.Unmarshal(msg.Data, &evt); err != nil {
		return // ignore malformed events
	}

	d.mu.RLock()
	if d.closed {
		d.mu.RUnlock()
		return
	}

	var handlers []Handler
	for _, sub := range d.subscribers {
		if d.matches(sub, evt) {
			handlers = append(handlers, sub.Handler)
		}
	}
	d.mu.RUnlock()

	for _, handler := range handlers {
		_ = handler(context.Background(), evt)
	}
}

// matches checks if a subscription matches an event.
func (d *NATSDispatcher) matches(sub *Subscription, evt Event) bool {
	if evt.Target != 0 && evt.Target != sub.ID {
		return false
	}

	if len(sub.Types) == 0 {
		return true
	}

	for _, t := range sub.Types {
		if t == evt.Type {
			return true
		}
	}

	return false
}

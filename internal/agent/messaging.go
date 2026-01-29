// Messaging provides inter-agent messaging for amux.
//
// This package implements the messaging system defined in spec §6.4,
// including message routing, ToSlug resolution, and message events.
//
// Messages flow through NATS P.comm.* subjects:
//   - P.comm.director: director channel
//   - P.comm.manager.<host_id>: host manager channel
//   - P.comm.agent.<host_id>.<agent_id>: agent channel
//   - P.comm.broadcast: broadcast to all participants
//
// See spec §6.4 for the complete messaging specification.
package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/stateforward/hsm-go/muid"

	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/internal/ids"
	"github.com/agentflare-ai/amux/pkg/api"
)

// MessageRouter handles inter-agent message routing per spec §6.4.
//
// The router resolves ToSlug to runtime IDs, enriches messages with
// sender information, and dispatches to the appropriate channels.
type MessageRouter struct {
	roster     *Roster
	dispatcher event.Dispatcher
	localID    muid.MUID // ID of local participant (manager or director)
	hostID     string    // host_id for this router (empty for director)
}

// NewMessageRouter creates a new message router.
// localID is the runtime ID of the local participant (manager or director).
// hostID is the host identifier (empty string for director).
func NewMessageRouter(roster *Roster, dispatcher event.Dispatcher, localID muid.MUID, hostID string) *MessageRouter {
	if dispatcher == nil {
		dispatcher = event.NewNoopDispatcher()
	}
	return &MessageRouter{
		roster:     roster,
		dispatcher: dispatcher,
		localID:    localID,
		hostID:     hostID,
	}
}

// ResolveToSlug resolves a ToSlug string to a recipient runtime ID.
// Returns (recipient ID, error). BroadcastID (0) is returned for broadcast targets.
//
// Resolution rules per spec §6.4.1.3:
//   - "all", "broadcast", "*" -> BroadcastID
//   - "director" -> director runtime ID
//   - "manager" -> local host manager runtime ID
//   - "manager@<host_id>" -> specific host manager runtime ID
//   - Otherwise -> agent_slug lookup
func (r *MessageRouter) ResolveToSlug(toSlug string) (muid.MUID, error) {
	// Check for broadcast targets
	if api.IsBroadcastSlug(toSlug) {
		return api.BroadcastID, nil
	}

	// Check for director
	if api.IsDirectorSlug(toSlug) {
		dirID := r.roster.DirectorID()
		if dirID == 0 {
			return 0, fmt.Errorf("resolve ToSlug: director not registered")
		}
		return dirID, nil
	}

	// Check for manager
	if api.IsManagerSlug(toSlug) {
		hostID := api.ParseManagerHostID(toSlug)
		if hostID == "" {
			// "manager" -> local host manager
			// Try to find the manager for the current host
			p := r.roster.GetBySlug(api.ManagerSlug)
			if p == nil && r.hostID != "" {
				// Try manager@hostID
				p = r.roster.GetBySlug("manager@" + r.hostID)
			}
			if p == nil {
				return 0, fmt.Errorf("resolve ToSlug: local manager not registered")
			}
			return p.ID, nil
		}
		// "manager@<host_id>" -> specific manager
		p := r.roster.GetBySlug(toSlug)
		if p == nil {
			return 0, fmt.Errorf("resolve ToSlug: manager %q not found", toSlug)
		}
		return p.ID, nil
	}

	// Agent slug lookup
	p := r.roster.GetBySlug(toSlug)
	if p == nil {
		return 0, fmt.Errorf("resolve ToSlug: participant %q not found", toSlug)
	}
	return p.ID, nil
}

// RouteMessage routes an outbound message from an agent.
// This is called by the host manager when it detects an outbound message
// from an agent's PTY output (via adapter pattern matching).
//
// The router:
//  1. Sets From to the sender runtime ID
//  2. Generates a unique message ID
//  3. Sets Timestamp to current time (UTC)
//  4. Resolves ToSlug to a recipient runtime ID
//  5. Dispatches message.outbound event
//
// Returns the enriched message or an error if resolution fails.
//
// See spec §6.4.1.
func (r *MessageRouter) RouteMessage(ctx context.Context, senderID muid.MUID, toSlug, content string) (*api.AgentMessage, error) {
	// Resolve recipient
	recipientID, err := r.ResolveToSlug(toSlug)
	if err != nil {
		// Per spec §6.4.1.3: if resolution fails, MUST NOT route the message
		return nil, fmt.Errorf("route message: %w", err)
	}

	// Create enriched message
	msg := &api.AgentMessage{
		ID:        ids.NewID(),
		From:      senderID,
		To:        recipientID,
		ToSlug:    toSlug,
		Content:   content,
		Timestamp: time.Now().UTC(),
	}

	// Dispatch message.outbound event
	evtData := map[string]any{
		"message_id": ids.EncodeID(msg.ID),
		"from":       ids.EncodeID(msg.From),
		"to":         ids.EncodeID(msg.To),
		"to_slug":    msg.ToSlug,
		"content":    msg.Content,
		"timestamp":  msg.Timestamp.Format(time.RFC3339Nano),
	}
	_ = r.dispatcher.Dispatch(ctx, event.NewEvent(event.TypeMessageOutbound, senderID, evtData))

	return msg, nil
}

// DeliverMessage delivers an inbound message to a recipient.
// This is called by the host manager when a message arrives for a local participant.
//
// Dispatches message.inbound event with the message data.
//
// See spec §6.4.1.
func (r *MessageRouter) DeliverMessage(ctx context.Context, msg *api.AgentMessage) error {
	// Dispatch message.inbound event
	evtData := map[string]any{
		"message_id": ids.EncodeID(msg.ID),
		"from":       ids.EncodeID(msg.From),
		"to":         ids.EncodeID(msg.To),
		"to_slug":    msg.ToSlug,
		"content":    msg.Content,
		"timestamp":  msg.Timestamp.Format(time.RFC3339Nano),
	}
	return r.dispatcher.Dispatch(ctx, event.NewEvent(event.TypeMessageInbound, msg.To, evtData))
}

// BroadcastMessage broadcasts a message to all participants.
// This is typically used by the director.
//
// Dispatches message.broadcast event with the message data.
//
// See spec §6.4.1.
func (r *MessageRouter) BroadcastMessage(ctx context.Context, senderID muid.MUID, content string) (*api.AgentMessage, error) {
	msg := &api.AgentMessage{
		ID:        ids.NewID(),
		From:      senderID,
		To:        api.BroadcastID,
		ToSlug:    "broadcast",
		Content:   content,
		Timestamp: time.Now().UTC(),
	}

	// Dispatch message.broadcast event
	evtData := map[string]any{
		"message_id": ids.EncodeID(msg.ID),
		"from":       ids.EncodeID(msg.From),
		"content":    msg.Content,
		"timestamp":  msg.Timestamp.Format(time.RFC3339Nano),
	}
	_ = r.dispatcher.Dispatch(ctx, event.NewEvent(event.TypeMessageBroadcast, senderID, evtData))

	return msg, nil
}

// IsBroadcast returns true if the message is a broadcast message.
func IsBroadcast(msg *api.AgentMessage) bool {
	return msg.To == api.BroadcastID
}

// MessageEnvelope wraps an AgentMessage for wire transmission.
// This is the JSON format used on NATS P.comm.* subjects.
//
// See spec §5.5.7.1 and §9.1.3.1 for wire format requirements.
type MessageEnvelope struct {
	// ID is the message ID (base-10 string per spec §9.1.3.1).
	ID string `json:"id"`

	// From is the sender runtime ID (base-10 string).
	From string `json:"from"`

	// To is the recipient runtime ID (base-10 string).
	// "0" indicates broadcast.
	To string `json:"to"`

	// ToSlug is the original recipient token from the message.
	ToSlug string `json:"to_slug"`

	// Content is the message body.
	Content string `json:"content"`

	// Timestamp is the message timestamp (RFC 3339 UTC).
	Timestamp string `json:"timestamp"`
}

// ToEnvelope converts an AgentMessage to a wire-format envelope.
func ToEnvelope(msg *api.AgentMessage) *MessageEnvelope {
	return &MessageEnvelope{
		ID:        ids.EncodeID(msg.ID),
		From:      ids.EncodeID(msg.From),
		To:        ids.EncodeID(msg.To),
		ToSlug:    msg.ToSlug,
		Content:   msg.Content,
		Timestamp: msg.Timestamp.Format(time.RFC3339Nano),
	}
}

// FromEnvelope converts a wire-format envelope to an AgentMessage.
func FromEnvelope(env *MessageEnvelope) (*api.AgentMessage, error) {
	msgID, err := ids.DecodeID(env.ID)
	if err != nil {
		return nil, fmt.Errorf("decode message ID: %w", err)
	}

	fromID, err := ids.DecodeID(env.From)
	if err != nil {
		return nil, fmt.Errorf("decode from ID: %w", err)
	}

	toID, err := ids.DecodeID(env.To)
	if err != nil {
		return nil, fmt.Errorf("decode to ID: %w", err)
	}

	ts, err := time.Parse(time.RFC3339Nano, env.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("parse timestamp: %w", err)
	}

	return &api.AgentMessage{
		ID:        msgID,
		From:      fromID,
		To:        toID,
		ToSlug:    env.ToSlug,
		Content:   env.Content,
		Timestamp: ts,
	}, nil
}

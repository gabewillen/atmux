package agent

import (
	"context"

	"github.com/stateforward/amux/internal/errors"
	"github.com/stateforward/amux/internal/event"
	"github.com/stateforward/amux/pkg/api"
)

// Message event type constants per spec §6.4 and §9.3.
const (
	EventTypeMessageOutbound  = "message.outbound"
	EventTypeMessageInbound   = "message.inbound"
	EventTypeMessageBroadcast = "message.broadcast"
)

// MessagePayload wraps an AgentMessage for event dispatch.
// Per spec §6.4, agents, host managers, and the director communicate using
// AgentMessage payloads routed over the event system and, in later phases,
// NATS participant channels.
type MessagePayload struct {
	Message api.AgentMessage
}

// EmitOutboundMessage emits a message.outbound event carrying the given message.
// Phase 4 uses the local event dispatcher; Phase 7 will route these over hsmnet
// and NATS per spec §5.5.7.1 and §6.4.
func EmitOutboundMessage(ctx context.Context, dispatcher event.Dispatcher, msg api.AgentMessage) error {
	if dispatcher == nil {
		return errors.Wrap(errors.ErrInvalidInput, "dispatcher must not be nil")
	}

	return dispatcher.Dispatch(ctx, event.BasicEvent{
		EventType: EventTypeMessageOutbound,
		Payload: MessagePayload{
			Message: msg,
		},
	})
}

// EmitInboundMessage emits a message.inbound event carrying the given message.
func EmitInboundMessage(ctx context.Context, dispatcher event.Dispatcher, msg api.AgentMessage) error {
	if dispatcher == nil {
		return errors.Wrap(errors.ErrInvalidInput, "dispatcher must not be nil")
	}

	return dispatcher.Dispatch(ctx, event.BasicEvent{
		EventType: EventTypeMessageInbound,
		Payload: MessagePayload{
			Message: msg,
		},
	})
}

// EmitBroadcastMessage emits a message.broadcast event carrying the given message.
func EmitBroadcastMessage(ctx context.Context, dispatcher event.Dispatcher, msg api.AgentMessage) error {
	if dispatcher == nil {
		return errors.Wrap(errors.ErrInvalidInput, "dispatcher must not be nil")
	}

	return dispatcher.Dispatch(ctx, event.BasicEvent{
		EventType: EventTypeMessageBroadcast,
		Payload: MessagePayload{
			Message: msg,
		},
	})
}

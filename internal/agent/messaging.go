package agent

import (
	"time"

	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go/muid"
)

// AgentMessage represents an inter-agent message.
type AgentMessage struct {
	ID        muid.MUID   `json:"id"`
	From      api.AgentID `json:"from"`
	To        api.AgentID `json:"to"`
	ToSlug    string      `json:"to_slug,omitempty"`
	Content   string      `json:"content"`
	Timestamp time.Time   `json:"timestamp"`
}

// SendMessage sends a message from one agent to another.
func SendMessage(bus *EventBus, from, to api.AgentID, content string) error {
	msg := AgentMessage{
		ID:        muid.Make(),
		From:      from,
		To:        to,
		Content:   content,
		Timestamp: time.Now().UTC(),
	}
	
	// Publish to Bus
	bus.Publish(BusEvent{
		Type:    EventMessage,
		Source:  from,
		Payload: msg,
	})
	
	return nil
}

// RouteMessage handles routing of a message.
// This is typically used by a centralized router if we have one.
// For now, the EventBus broadcasts, and agents filter.
// In Phase 7 (Remote), this will route via NATS.

// SubscribeToMessages returns a channel that receives messages for a specific agent.
func SubscribeToMessages(bus *EventBus, agentID api.AgentID) (<-chan AgentMessage, func()) {
	sub := bus.Subscribe()
	out := make(chan AgentMessage, 100)
	
	go func() {
		defer close(out)
		for event := range sub.C {
			if event.Type == EventMessage {
				if msg, ok := event.Payload.(AgentMessage); ok {
					if msg.To == agentID || msg.To == 0 { // Support broadcast
						out <- msg
					}
				}
			}
		}
	}()
	
	return out, sub.Close
}

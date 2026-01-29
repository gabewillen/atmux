package agent

import (
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go/muid"
)

// Message represents an inter-agent message.
type Message struct {
	ID        string      `json:"id"`
	From      api.AgentID `json:"from"`
	To        api.AgentID `json:"to"`
	Content   string      `json:"content"`
	Timestamp int64       `json:"timestamp"`
}

// SendMessage sends a message from one agent to another.
func SendMessage(bus *EventBus, from, to api.AgentID, content string) error {
	msg := Message{
		ID:      muid.Make().String(),
		From:    from,
		To:      to,
		Content: content,
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
func SubscribeToMessages(bus *EventBus, agentID api.AgentID) (<-chan Message, func()) {
	sub := bus.Subscribe()
	out := make(chan Message, 100)
	
	go func() {
		defer close(out)
		for event := range sub.C {
			if event.Type == EventMessage {
				if msg, ok := event.Payload.(Message); ok {
					if msg.To == agentID || msg.To == 0 { // Support broadcast
						out <- msg
					}
				}
			}
		}
	}()
	
	return out, sub.Close
}

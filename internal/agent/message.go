package agent

import (
	"github.com/stateforward/hsm-go/muid"
)

// EventAgentMessage is the event name for inter-agent messages.
const EventAgentMessage = "agent.message"

// MessagePayload represents the content of an inter-agent message.
type MessagePayload struct {
	FromID  muid.MUID `json:"from_id"`
	ToID    muid.MUID `json:"to_id"`
	Content string    `json:"content"`
}

// RouteMessage determines if a message is for this agent.
func (a *AgentActor) RouteMessage(payload MessagePayload) bool {
	return payload.ToID == a.ID()
}

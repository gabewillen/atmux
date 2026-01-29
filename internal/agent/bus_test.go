package agent

import (
	"testing"
	"time"

	"github.com/agentflare-ai/amux/pkg/api"
)

func TestEventBus(t *testing.T) {
	bus := NewEventBus()
	sub1 := bus.Subscribe()
	defer sub1.Close()

	agentID := api.AgentID(123)
	event := BusEvent{
		Type:    EventPresenceUpdate,
		Source:  agentID,
		Payload: "Online",
	}

	bus.Publish(event)

	select {
	case e := <-sub1.C:
		if e.Type != EventPresenceUpdate {
			t.Errorf("Expected event type presence.update, got %s", e.Type)
		}
		if e.Source != agentID {
			t.Errorf("Expected source %v, got %v", agentID, e.Source)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for event")
	}
}

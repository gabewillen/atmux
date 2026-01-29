package agent

import (
	"testing"
	"time"

	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go/muid"
)

func TestInterAgentMessaging(t *testing.T) {
	bus := NewEventBus()
	agentA := api.AgentID(muid.Make())
	agentB := api.AgentID(muid.Make())

	// Subscribe Agent B
	msgs, cancel := SubscribeToMessages(bus, agentB)
	defer cancel()

	// Send from A to B
	err := SendMessage(bus, agentA, agentB, "hello B")
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	// Verify reception
	select {
	case msg := <-msgs:
		if msg.From != agentA {
			t.Errorf("Expected From %s, got %s", agentA, msg.From)
		}
		if msg.Content != "hello B" {
			t.Errorf("Expected content 'hello B', got %s", msg.Content)
		}
		if msg.Timestamp.IsZero() {
			t.Error("Expected timestamp to be set")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timed out waiting for message")
	}
}

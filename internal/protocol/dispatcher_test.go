package protocol

import (
	"context"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/pkg/api"
)

// MockLocalBus captures published events.
type MockLocalBus struct {
	Events []interface{}
}

func (m *MockLocalBus) Publish(event interface{}) {
	m.Events = append(m.Events, event)
}

func TestDispatcher_Local(t *testing.T) {
	d := NewDispatcher(api.PeerID(1), "host1", nil)
	localBus := &MockLocalBus{}
	d.SetLocalBus(localBus)

	event := EventMessage{
		ID:        "evt-1",
		Type:      "test.event",
		Timestamp: time.Now(),
		Payload:   "hello",
	}

	if err := d.Dispatch(context.Background(), event); err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if len(localBus.Events) != 1 {
		t.Errorf("Expected 1 local event, got %d", len(localBus.Events))
	}
	
	got, ok := localBus.Events[0].(EventMessage)
	if !ok {
		t.Fatalf("Event type mismatch: %T", localBus.Events[0])
	}
	
	if got.ID != "evt-1" {
		t.Errorf("Event ID mismatch")
	}
}

// NOTE: Testing remote dispatch requires mocking nats.Conn or using nats-server test helper.
// Since we don't have nats-server helper easily importable without adding heavy dependencies to test,
// we rely on local dispatch test and compilation of NATS logic.

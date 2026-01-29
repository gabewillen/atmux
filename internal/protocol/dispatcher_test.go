package protocol

import (
	"context"
	"testing"

	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
	"github.com/stateforward/hsm-go/muid"
)

type mockBus struct {
	events []interface{}
}

func (b *mockBus) Publish(event interface{}) {
	b.events = append(b.events, event)
}

func TestDispatcher_Start(t *testing.T) {
	// Need a NATS server or mock.
	// Since we don't have an embedded NATS server easily available in unit tests 
	// without gnatsd binary, we might skip if not connected.
	// However, we can test the structure.
	
	peerID := api.PeerID(muid.Make())
	hostID := api.HostID("test-host")
	
	// Try connecting to default NATS?
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Skip("NATS not available, skipping integration test")
	}
	defer nc.Close()
	
	d := NewDispatcher(peerID, hostID, nc)
	bus := &mockBus{}
	d.SetLocalBus(bus)
	
	ctx := context.Background()
	if err := d.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer d.Stop()
	
	// Simulate remote event (TODO: Implement proper publish/subscribe verification)
	// evt := EventMessage{
	// 	ID:     "123",
	// 	Type:   "test.event",
	// 	Source: "remote-peer",
	// }
	
	// subject := "amux.events.remote-host"
}

func TestDispatcher_Dispatch(t *testing.T) {
	peerID := api.PeerID(muid.Make())
	hostID := api.HostID("test-host")
	
	d := NewDispatcher(peerID, hostID, nil) // No NATS
	bus := &mockBus{}
	d.SetLocalBus(bus)
	
	evt := EventMessage{ID: "1", Type: "local"}
	d.Dispatch(context.Background(), evt)
	
	if len(bus.events) != 1 {
		t.Error("Expected 1 local event")
	}
}


// NOTE: Testing remote dispatch requires mocking nats.Conn or using nats-server test helper.
// Since we don't have nats-server helper easily importable without adding heavy dependencies to test,
// we rely on local dispatch test and compilation of NATS logic.

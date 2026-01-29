package agent

import (
	"context"
	"testing"

	"github.com/stateforward/hsm-go/muid"

	"github.com/stateforward/amux/internal/event"
	"github.com/stateforward/amux/pkg/api"
)

func TestEmitOutboundMessage(t *testing.T) {
	dispatcher := event.NewLocalDispatcher()
	ctx := context.Background()

	ch, err := dispatcher.Subscribe(ctx, event.TypeFilter{Prefix: EventTypeMessageOutbound})
	if err != nil {
		t.Fatalf("Subscribe returned error: %v", err)
	}

	msg := api.AgentMessage{
		ID:      muid.Make(),
		From:    muid.Make(),
		To:      muid.Make(),
		ToSlug:  "target",
		Content: "hello",
	}

	if err := EmitOutboundMessage(ctx, dispatcher, msg); err != nil {
		t.Fatalf("EmitOutboundMessage returned error: %v", err)
	}

	select {
	case ev := <-ch:
		basic, ok := ev.(event.BasicEvent)
		if !ok {
			t.Fatalf("event type = %T, want event.BasicEvent", ev)
		}
		if basic.EventType != EventTypeMessageOutbound {
			t.Errorf("event type = %q, want %q", basic.EventType, EventTypeMessageOutbound)
		}
		payload, ok := basic.Payload.(MessagePayload)
		if !ok {
			t.Fatalf("payload type = %T, want MessagePayload", basic.Payload)
		}
		if payload.Message.Content != "hello" {
			t.Errorf("payload.Message.Content = %q, want %q", payload.Message.Content, "hello")
		}
	default:
		t.Fatal("expected message.outbound event, got none")
	}
}

func TestEmitInboundAndBroadcastMessages(t *testing.T) {
	dispatcher := event.NewLocalDispatcher()
	ctx := context.Background()

	inboundCh, err := dispatcher.Subscribe(ctx, event.TypeFilter{Prefix: EventTypeMessageInbound})
	if err != nil {
		t.Fatalf("Subscribe inbound returned error: %v", err)
	}
	broadcastCh, err := dispatcher.Subscribe(ctx, event.TypeFilter{Prefix: EventTypeMessageBroadcast})
	if err != nil {
		t.Fatalf("Subscribe broadcast returned error: %v", err)
	}

	msg := api.AgentMessage{
		ID:      muid.Make(),
		From:    muid.Make(),
		To:      api.BroadcastID,
		ToSlug:  "all",
		Content: "broadcast",
	}

	if err := EmitInboundMessage(ctx, dispatcher, msg); err != nil {
		t.Fatalf("EmitInboundMessage returned error: %v", err)
	}
	if err := EmitBroadcastMessage(ctx, dispatcher, msg); err != nil {
		t.Fatalf("EmitBroadcastMessage returned error: %v", err)
	}

	select {
	case ev := <-inboundCh:
		basic, ok := ev.(event.BasicEvent)
		if !ok {
			t.Fatalf("inbound event type = %T, want event.BasicEvent", ev)
		}
		if basic.EventType != EventTypeMessageInbound {
			t.Errorf("inbound event type = %q, want %q", basic.EventType, EventTypeMessageInbound)
		}
	default:
		t.Fatal("expected message.inbound event, got none")
	}

	select {
	case ev := <-broadcastCh:
		basic, ok := ev.(event.BasicEvent)
		if !ok {
			t.Fatalf("broadcast event type = %T, want event.BasicEvent", ev)
		}
		if basic.EventType != EventTypeMessageBroadcast {
			t.Errorf("broadcast event type = %q, want %q", basic.EventType, EventTypeMessageBroadcast)
		}
	default:
		t.Fatal("expected message.broadcast event, got none")
	}
}

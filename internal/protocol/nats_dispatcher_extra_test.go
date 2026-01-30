package protocol

import (
	"context"
	"testing"
	"time"
)

func TestDispatcherNilHelpers(t *testing.T) {
	var dispatcher *NATSDispatcher
	if dispatcher.MaxPayload() != 0 {
		t.Fatalf("expected zero max payload")
	}
	select {
	case <-dispatcher.Closed():
	default:
		t.Fatalf("expected closed channel")
	}
	if err := (&natsSubscription{}).Unsubscribe(); err != nil {
		t.Fatalf("expected nil unsubscribe")
	}
	if err := (*natsSubscription)(nil).Unsubscribe(); err != nil {
		t.Fatalf("expected nil unsubscribe")
	}
}

func TestDispatcherInvalidArgs(t *testing.T) {
	dispatcher := &NATSDispatcher{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := dispatcher.SubscribeRaw(ctx, "subject", nil); err == nil {
		t.Fatalf("expected subscribe error")
	}
	if _, err := dispatcher.Request(ctx, "subject", nil, time.Millisecond); err == nil {
		t.Fatalf("expected request error")
	}
	if err := dispatcher.Publish(ctx, "subject", Event{Name: "x"}); err == nil {
		t.Fatalf("expected publish error")
	}
}

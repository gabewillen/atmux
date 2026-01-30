package manager

import (
	"context"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/nats-io/nats.go"
)

type recordSub struct {
	unsubscribed bool
}

func (s *recordSub) Unsubscribe() error {
	s.unsubscribed = true
	return nil
}

type routingRecordDispatcher struct {
	outboundSub *recordSub
}

func (r *routingRecordDispatcher) Publish(ctx context.Context, subject string, event protocol.Event) error {
	return nil
}
func (r *routingRecordDispatcher) Subscribe(ctx context.Context, subject string, handler func(protocol.Event)) (protocol.Subscription, error) {
	r.outboundSub = &recordSub{}
	return r.outboundSub, nil
}
func (r *routingRecordDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	return nil
}
func (r *routingRecordDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(protocol.Message)) (protocol.Subscription, error) {
	return &recordSub{}, nil
}
func (r *routingRecordDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (protocol.Message, error) {
	return protocol.Message{}, nil
}
func (r *routingRecordDispatcher) MaxPayload() int { return 1024 }
func (r *routingRecordDispatcher) JetStream() nats.JetStreamContext { return nil }
func (r *routingRecordDispatcher) Closed() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func TestStartMessageRoutingMissingHost(t *testing.T) {
	dispatcher := &routingRecordDispatcher{}
	mgr := &Manager{dispatcher: dispatcher}
	if err := mgr.startMessageRouting(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
	if dispatcher.outboundSub == nil || !dispatcher.outboundSub.unsubscribed {
		t.Fatalf("expected outbound unsubscribe")
	}
}

package remote

import (
	"context"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
)

type responseDispatcher struct {
	resp protocol.Message
	err  error
}

func (r responseDispatcher) Publish(ctx context.Context, subject string, event protocol.Event) error {
	_ = ctx
	_ = subject
	_ = event
	return nil
}
func (r responseDispatcher) Subscribe(ctx context.Context, subject string, handler func(protocol.Event)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}
func (r responseDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	_ = ctx
	_ = subject
	_ = payload
	_ = reply
	return nil
}
func (r responseDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(protocol.Message)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}
func (r responseDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (protocol.Message, error) {
	_ = ctx
	_ = subject
	_ = payload
	_ = timeout
	return r.resp, r.err
}
func (r responseDispatcher) MaxPayload() int { return 0 }
func (r responseDispatcher) JetStream() nats.JetStreamContext { return nil }
func (r responseDispatcher) Closed() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func TestSendControlInvalidResponse(t *testing.T) {
	host := api.MustParseHostID("host")
	state := &hostState{hostID: host, connected: true, ready: true}
	director := &Director{
		dispatcher:    responseDispatcher{resp: protocol.Message{Data: []byte("bad")}},
		subjectPrefix: "amux",
		hosts:         map[api.HostID]*hostState{host: state},
	}
	if _, err := director.sendControl(context.Background(), host, ControlMessage{Type: "ping", Payload: []byte("{}")}); err == nil {
		t.Fatalf("expected decode error")
	}
}

func TestSendControlNotReady(t *testing.T) {
	host := api.MustParseHostID("host")
	state := &hostState{hostID: host, connected: true, ready: true}
	errMsg, err := NewErrorMessage("spawn", "not_ready", "not ready")
	if err != nil {
		t.Fatalf("new error: %v", err)
	}
	data, err := EncodeControlMessage(errMsg)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	director := &Director{
		dispatcher:    responseDispatcher{resp: protocol.Message{Data: data}},
		subjectPrefix: "amux",
		hosts:         map[api.HostID]*hostState{host: state},
	}
	if _, err := director.sendControl(context.Background(), host, ControlMessage{Type: "spawn", Payload: []byte("{}")}); err == nil {
		t.Fatalf("expected not ready error")
	}
	if state.ready {
		t.Fatalf("expected host marked not ready")
	}
}

package remote

import (
	"context"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
)

type recordRawDispatcher struct {
	lastSubject string
	lastPayload []byte
}

func (r *recordRawDispatcher) Publish(ctx context.Context, subject string, event protocol.Event) error {
	_ = ctx
	_ = subject
	_ = event
	return nil
}
func (r *recordRawDispatcher) Subscribe(ctx context.Context, subject string, handler func(protocol.Event)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}
func (r *recordRawDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	_ = ctx
	_ = reply
	r.lastSubject = subject
	r.lastPayload = append([]byte(nil), payload...)
	return nil
}
func (r *recordRawDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(protocol.Message)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}
func (r *recordRawDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (protocol.Message, error) {
	_ = ctx
	_ = subject
	_ = payload
	_ = timeout
	return protocol.Message{}, nil
}
func (r *recordRawDispatcher) MaxPayload() int { return 0 }
func (r *recordRawDispatcher) JetStream() nats.JetStreamContext { return nil }
func (r *recordRawDispatcher) Closed() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func setUnexportedField(target any, name string, value any) {
	v := reflect.ValueOf(target).Elem().FieldByName(name)
	reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem().Set(reflect.ValueOf(value))
}

func TestDirectorHandleHandshakeInvalid(t *testing.T) {
	dispatcher := &recordRawDispatcher{}
	director := &Director{}
	setUnexportedField(director, "dispatcher", dispatcher)
	setUnexportedField(director, "subjectPrefix", "amux")
	msg := protocol.Message{Reply: "reply", Subject: "amux.handshake.host", Data: []byte("bad")}
	director.handleHandshake(msg)
	if dispatcher.lastSubject != "reply" {
		t.Fatalf("expected reply subject")
	}
}

func TestEnsureConnectedErrors(t *testing.T) {
	director := &Director{}
	host := api.MustParseHostID("host")
	setUnexportedField(director, "hosts", map[api.HostID]*hostState{})
	if err := director.ensureConnected(host); err == nil {
		t.Fatalf("expected disconnected error")
	}
	state := &hostState{hostID: host, connected: true, ready: false}
	setUnexportedField(director, "hosts", map[api.HostID]*hostState{host: state})
	if err := director.ensureConnected(host); err == nil {
		t.Fatalf("expected not ready error")
	}
	state.ready = true
	if err := director.ensureConnected(host); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

type errDispatcher struct{}

func (errDispatcher) Publish(ctx context.Context, subject string, event protocol.Event) error { return nil }
func (errDispatcher) Subscribe(ctx context.Context, subject string, handler func(protocol.Event)) (protocol.Subscription, error) {
	return nil, nil
}
func (errDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error { return nil }
func (errDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(protocol.Message)) (protocol.Subscription, error) {
	return nil, nil
}
func (errDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (protocol.Message, error) {
	return protocol.Message{}, nats.ErrNoResponders
}
func (errDispatcher) MaxPayload() int { return 0 }
func (errDispatcher) JetStream() nats.JetStreamContext { return nil }
func (errDispatcher) Closed() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func TestSendControlNoRespondersMarksDisconnected(t *testing.T) {
	host := api.MustParseHostID("host")
	state := &hostState{hostID: host, connected: true, ready: true}
	director := &Director{dispatcher: errDispatcher{}, subjectPrefix: "amux", hosts: map[api.HostID]*hostState{host: state}}
	_, err := director.sendControl(context.Background(), host, ControlMessage{Type: "ping", Payload: []byte("{}")})
	if err == nil {
		t.Fatalf("expected send control error")
	}
	if state.connected || state.ready {
		t.Fatalf("expected host marked disconnected/not ready")
	}
}

package remote

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
)

type replayDispatcher struct {
	requested []string
}

func (r *replayDispatcher) Publish(ctx context.Context, subject string, event protocol.Event) error {
	_ = ctx
	_ = subject
	_ = event
	return nil
}
func (r *replayDispatcher) Subscribe(ctx context.Context, subject string, handler func(protocol.Event)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}
func (r *replayDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	_ = ctx
	_ = subject
	_ = payload
	_ = reply
	return nil
}
func (r *replayDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(protocol.Message)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}
func (r *replayDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (protocol.Message, error) {
	_ = ctx
	_ = subject
	_ = timeout
	control, err := DecodeControlMessage(payload)
	if err == nil && control.Type == "replay" {
		var req ReplayRequest
		if err := json.Unmarshal(control.Payload, &req); err == nil {
			r.requested = append(r.requested, req.SessionID)
		}
	}
	response, err := EncodePayload("replay", ReplayResponse{Accepted: true})
	if err != nil {
		return protocol.Message{}, err
	}
	data, err := EncodeControlMessage(response)
	if err != nil {
		return protocol.Message{}, err
	}
	return protocol.Message{Data: data}, nil
}
func (r *replayDispatcher) MaxPayload() int { return 0 }
func (r *replayDispatcher) JetStream() nats.JetStreamContext { return nil }
func (r *replayDispatcher) Closed() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func TestDirectorRequestReplay(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	server, err := protocol.StartHubServer(ctx, protocol.HubServerConfig{
		Listen:       "127.0.0.1:-1",
		JetStreamDir: filepath.Join(t.TempDir(), "jetstream"),
	})
	if err != nil {
		t.Fatalf("start hub: %v", err)
	}
	t.Cleanup(func() { _ = server.Close() })
	dispatcher, err := protocol.NewNATSDispatcher(ctx, server.URL(), protocol.NATSOptions{})
	if err != nil {
		t.Fatalf("dispatcher: %v", err)
	}
	t.Cleanup(func() { _ = dispatcher.Close(context.Background()) })
	kv, err := NewKVStore(dispatcher.JetStream(), "kv")
	if err != nil {
		t.Fatalf("kv store: %v", err)
	}
	host := api.MustParseHostID("host")
	if err := kv.Put(ctx, "sessions/"+host.String()+"/a", []byte(`{"state":"running"}`)); err != nil {
		t.Fatalf("put session: %v", err)
	}
	if err := kv.Put(ctx, "sessions/"+host.String()+"/b", []byte(`{"state":"completed"}`)); err != nil {
		t.Fatalf("put session: %v", err)
	}
	replayDisp := &replayDispatcher{}
	director := &Director{
		dispatcher: replayDisp,
		kv:         kv,
		hosts:      map[api.HostID]*hostState{host: {hostID: host, connected: true, ready: true}},
	}
	director.requestReplay(ctx, host)
	if len(replayDisp.requested) != 1 || replayDisp.requested[0] != "a" {
		t.Fatalf("unexpected replay requests: %#v", replayDisp.requested)
	}
}

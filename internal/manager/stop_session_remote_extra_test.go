package manager

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/remote"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
)

type errorDispatcher struct {
	err error
}

func (e errorDispatcher) Publish(ctx context.Context, subject string, event protocol.Event) error {
	_ = ctx
	_ = subject
	_ = event
	return nil
}

func (e errorDispatcher) Subscribe(ctx context.Context, subject string, handler func(protocol.Event)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}

func (e errorDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	_ = ctx
	_ = subject
	_ = payload
	_ = reply
	return nil
}

func (e errorDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(protocol.Message)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}

func (e errorDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (protocol.Message, error) {
	_ = ctx
	_ = subject
	_ = payload
	_ = timeout
	return protocol.Message{}, e.err
}

func (e errorDispatcher) MaxPayload() int { return 0 }

func (e errorDispatcher) JetStream() nats.JetStreamContext { return nil }

func (e errorDispatcher) Closed() <-chan struct{} { return make(chan struct{}) }

func TestStopSessionRemoteError(t *testing.T) {
	hostID := api.MustParseHostID("host")
	agentID := api.NewAgentID()
	sessionID := api.NewSessionID()
	manager := &Manager{
		agents: map[api.AgentID]*agentState{
			agentID: {
				remote:        true,
				remoteHost:    hostID,
				remoteSession: sessionID,
			},
		},
	}
	dir := &remote.Director{}
	setUnexportedField(dir, "dispatcher", errorDispatcher{err: errors.New("boom")})
	setUnexportedField(dir, "subjectPrefix", "amux")
	setUnexportedField(dir, "requestTimeout", time.Millisecond)
	hostsField := reflect.ValueOf(dir).Elem().FieldByName("hosts")
	hostMap := reflect.MakeMap(hostsField.Type())
	statePtr := reflect.New(hostMap.Type().Elem().Elem())
	setUnexportedField(statePtr.Interface(), "hostID", hostID)
	setUnexportedField(statePtr.Interface(), "connected", true)
	setUnexportedField(statePtr.Interface(), "ready", true)
	hostMap.SetMapIndex(reflect.ValueOf(hostID), statePtr)
	reflect.NewAt(hostsField.Type(), unsafe.Pointer(hostsField.UnsafeAddr())).Elem().Set(hostMap)
	manager.remoteDirector = dir
	if err := manager.stopSession(context.Background(), agentID); err == nil {
		t.Fatalf("expected stop session error")
	}
}

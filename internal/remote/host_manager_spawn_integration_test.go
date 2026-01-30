package remote

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestHandleSpawnSuccess(t *testing.T) {
	repoRoot := initRepo(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	jetDir := filepath.Join(t.TempDir(), "jetstream")
	server, err := protocol.StartHubServer(ctx, protocol.HubServerConfig{
		Listen:       "127.0.0.1:-1",
		JetStreamDir: jetDir,
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
	kv, err := NewKVStore(dispatcher.JetStream(), "AMUX_KV")
	if err != nil {
		t.Fatalf("kv store: %v", err)
	}
	manager := &HostManager{
		hostID:     api.MustParseHostID("host"),
		dispatcher: dispatcher,
		kv:         kv,
		registry:   stubRegistry{},
		sessions:   make(map[api.SessionID]*remoteSession),
		agentIndex: make(map[api.AgentID]*remoteSession),
	}
	reply := "reply.spawn"
	respCh := make(chan ControlMessage, 1)
	_, err = dispatcher.SubscribeRaw(ctx, reply, func(msg protocol.Message) {
		control, err := DecodeControlMessage(msg.Data)
		if err != nil {
			return
		}
		select {
		case respCh <- control:
		default:
		}
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	req := SpawnRequest{
		Name:      "alpha",
		AgentID:   api.NewAgentID().String(),
		AgentSlug: "alpha",
		RepoPath:  repoRoot,
		Adapter:   "stub",
		Command:   []string{os.Args[0], "-test.run=TestRemoteHelperProcess"},
		Env: map[string]string{
			"AMUX_HELPER": "1",
		},
	}
	payload, err := EncodePayload("spawn", req)
	if err != nil {
		t.Fatalf("encode spawn: %v", err)
	}
	manager.handleSpawn(reply, payload)
	select {
	case control := <-respCh:
		if control.Type != "spawn" {
			t.Fatalf("expected spawn response")
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timeout waiting for spawn response")
	}
}

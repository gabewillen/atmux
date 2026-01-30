package remote

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestDirectorControlRequests(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
	resolver, err := configResolver(t)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	cfg := config.DefaultConfig(resolver)
	cfg.NATS.JetStreamDir = jetDir
	cfg.Remote.NATS.KVBucket = "AMUX_KV"
	director, err := NewDirector(cfg, dispatcher, DirectorOptions{Version: "test", HostID: api.MustParseHostID("director")})
	if err != nil {
		t.Fatalf("new director: %v", err)
	}
	hostID := api.MustParseHostID("host")
	setDirectorHostState(director, hostID, true, true)
	controlSubject := ControlSubject("amux", hostID)
	_, err = dispatcher.SubscribeRaw(ctx, controlSubject, func(msg protocol.Message) {
		control, err := DecodeControlMessage(msg.Data)
		if err != nil {
			return
		}
		var response ControlMessage
		switch control.Type {
		case "spawn":
			response, err = EncodePayload("spawn", SpawnResponse{AgentID: api.NewAgentID().String(), SessionID: api.NewSessionID().String()})
		case "kill":
			response, err = EncodePayload("kill", KillResponse{SessionID: "1", Killed: true})
		case "replay":
			response, err = EncodePayload("replay", ReplayResponse{SessionID: "1", Accepted: true})
		default:
			response, err = NewErrorMessage("unknown", "invalid", "bad")
		}
		if err != nil {
			return
		}
		data, err := EncodeControlMessage(response)
		if err != nil {
			return
		}
		_ = dispatcher.PublishRaw(context.Background(), msg.Reply, data, "")
	})
	if err != nil {
		t.Fatalf("subscribe control: %v", err)
	}
	_, err = director.Spawn(ctx, hostID, SpawnRequest{
		AgentID:   api.NewAgentID().String(),
		AgentSlug: "alpha",
		RepoPath:  "/tmp",
		Adapter:   "stub",
		Command:   []string{"echo", "ok"},
	})
	if err != nil {
		t.Fatalf("spawn: %v", err)
	}
	if _, err := director.Kill(ctx, hostID, KillRequest{SessionID: "1"}); err != nil {
		t.Fatalf("kill: %v", err)
	}
	if _, err := director.Replay(ctx, hostID, ReplayRequest{SessionID: "1"}); err != nil {
		t.Fatalf("replay: %v", err)
	}
}

func configResolver(t *testing.T) (*paths.Resolver, error) {
	t.Helper()
	repoRoot := initRepo(t)
	return paths.NewResolver(repoRoot)
}

func setDirectorHostState(director *Director, hostID api.HostID, connected bool, ready bool) {
	v := reflect.ValueOf(director).Elem().FieldByName("hosts")
	hostMap := reflect.MakeMap(v.Type())
	statePtr := reflect.New(v.Type().Elem().Elem())
	setUnexportedField(statePtr.Interface(), "hostID", hostID)
	setUnexportedField(statePtr.Interface(), "connected", connected)
	setUnexportedField(statePtr.Interface(), "ready", ready)
	hostMap.SetMapIndex(reflect.ValueOf(hostID), statePtr)
	reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem().Set(hostMap)
}

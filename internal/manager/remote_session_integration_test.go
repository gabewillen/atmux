package manager

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/remote"
	"github.com/agentflare-ai/amux/pkg/api"
	"os"
)

func TestStartRemoteSession(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	repoRoot := initRepo(t)
	resolver, err := paths.NewResolver(repoRoot)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	cfg := config.DefaultConfig(resolver)
	cfg.Remote.NATS.SubjectPrefix = "amux"
	cfg.Remote.RequestTimeout = 2 * time.Second
	cfg.NATS.JetStreamDir = filepath.Join(t.TempDir(), "jetstream")
	server, err := protocol.StartHubServer(ctx, protocol.HubServerConfig{
		Listen:       "127.0.0.1:-1",
		JetStreamDir: cfg.NATS.JetStreamDir,
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
	director, err := remote.NewDirector(cfg, dispatcher, remote.DirectorOptions{Version: "test", HostID: api.MustParseHostID("director")})
	if err != nil {
		t.Fatalf("director: %v", err)
	}
	hostID := api.MustParseHostID("host")
	setDirectorHostReady(director, hostID)
	_, err = dispatcher.SubscribeRaw(ctx, remote.ControlSubject("amux", hostID), func(msg protocol.Message) {
		control, err := remote.DecodeControlMessage(msg.Data)
		if err != nil {
			return
		}
		if control.Type != "spawn" || msg.Reply == "" {
			return
		}
		resp := remote.SpawnResponse{AgentID: "1", SessionID: api.NewSessionID().String()}
		out, err := remote.EncodePayload("spawn", resp)
		if err != nil {
			return
		}
		data, err := remote.EncodeControlMessage(out)
		if err != nil {
			return
		}
		_ = dispatcher.PublishRaw(context.Background(), msg.Reply, data, "")
	})
	if err != nil {
		t.Fatalf("subscribe control: %v", err)
	}
	mgr := &Manager{
		resolver:       resolver,
		dispatcher:     dispatcher,
		cfg:            cfg,
		remoteDirector: director,
		agents:         map[api.AgentID]*agentState{},
		registries:     map[string]adapter.Registry{},
	}
	mgr.SetRegistryFactory(func(resolver *paths.Resolver) (adapter.Registry, error) {
		_ = resolver
		return &stubRegistry{cmd: []string{"env", "AMUX_HELPER=1", os.Args[0], "-test.run=TestManagerHelperProcess"}}, nil
	})
	agentID := api.NewAgentID()
	state := &agentState{
		slug:       "alpha",
		repoRoot:   repoRoot,
		worktree:   filepath.Join(repoRoot, ".amux", "worktrees", "alpha"),
		remote:     true,
		remoteHost: hostID,
		config: config.AgentConfig{
			Name:    "alpha",
			Adapter: "stub",
			Location: config.AgentLocationConfig{
				Type:     api.LocationSSH.String(),
				Host:     hostID.String(),
				RepoPath: repoRoot,
			},
		},
	}
	mgr.agents[agentID] = state
	if _, err := mgr.startSession(ctx, agentID); err != nil {
		t.Fatalf("start remote session: %v", err)
	}
	if state.remoteSession.IsZero() {
		t.Fatalf("expected remote session id")
	}
}

func setDirectorHostReady(director *remote.Director, hostID api.HostID) {
	v := reflect.ValueOf(director).Elem().FieldByName("hosts")
	hostMap := reflect.MakeMap(v.Type())
	statePtr := reflect.New(v.Type().Elem().Elem())
	setUnexportedField(statePtr.Interface(), "hostID", hostID)
	setUnexportedField(statePtr.Interface(), "connected", true)
	setUnexportedField(statePtr.Interface(), "ready", true)
	hostMap.SetMapIndex(reflect.ValueOf(hostID), statePtr)
	reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem().Set(hostMap)
}

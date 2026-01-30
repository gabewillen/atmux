package manager

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/remote"
	"github.com/agentflare-ai/amux/pkg/api"
)

type sequenceRunner struct {
	outputs map[string][][]byte
}

func (s *sequenceRunner) Run(ctx context.Context, target string, options []string, command string, stdin []byte) error {
	_ = ctx
	_ = target
	_ = options
	_ = command
	_ = stdin
	return nil
}

func (s *sequenceRunner) RunOutput(ctx context.Context, target string, options []string, command string, stdin []byte) ([]byte, error) {
	_ = ctx
	_ = target
	_ = options
	_ = stdin
	if list, ok := s.outputs[command]; ok && len(list) > 0 {
		out := list[0]
		s.outputs[command] = list[1:]
		return out, nil
	}
	return []byte(""), nil
}

func TestAddRemoteAgentSuccess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)
	repoRoot := initRepo(t)
	resolver, err := paths.NewResolver(repoRoot)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	wasmPath := resolver.ProjectAdapterWasmPath("stub")
	if err := os.MkdirAll(filepath.Dir(wasmPath), 0o755); err != nil {
		t.Fatalf("mkdir wasm: %v", err)
	}
	if err := os.WriteFile(wasmPath, []byte("wasm"), 0o644); err != nil {
		t.Fatalf("write wasm: %v", err)
	}
	cfg := config.DefaultConfig(resolver)
	cfg.Remote.NATS.SubjectPrefix = "amux"
	cfg.Remote.NATS.KVBucket = "AMUX_KV"
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
	mgr, err := NewManager(ctx, resolver, cfg, dispatcher, "test")
	if err != nil {
		t.Fatalf("manager: %v", err)
	}
	mgr.SetRegistryFactory(func(resolver *paths.Resolver) (adapter.Registry, error) {
		_ = resolver
		return &stubRegistry{cmd: []string{"env", "AMUX_HELPER=1", os.Args[0], "-test.run=TestManagerHelperProcess"}}, nil
	})
	runner := &sequenceRunner{
		outputs: map[string][][]byte{
			"uname -s": {[]byte(runtime.GOOS)},
			"uname -m": {[]byte(runtime.GOARCH)},
			"PATH=\"$HOME/.local/bin:$PATH\" amux-manager status": {
				[]byte("hub_connected=false"),
				[]byte("hub_connected=true"),
			},
		},
	}
	setUnexportedField(mgr.remoteDirector, "bootstrapper", &remote.Bootstrapper{Runner: runner})
	hostID := api.MustParseHostID("host")
	setDirectorHostReady(mgr.remoteDirector, hostID)
	_, err = dispatcher.SubscribeRaw(ctx, remote.ControlSubject("amux", hostID), func(msg protocol.Message) {
		control, err := remote.DecodeControlMessage(msg.Data)
		if err != nil || msg.Reply == "" {
			return
		}
		if control.Type != "spawn" {
			return
		}
		resp := remote.SpawnResponse{AgentID: api.NewAgentID().String(), SessionID: api.NewSessionID().String()}
		payload, err := remote.EncodePayload("spawn", resp)
		if err != nil {
			return
		}
		data, err := remote.EncodeControlMessage(payload)
		if err != nil {
			return
		}
		_ = dispatcher.PublishRaw(context.Background(), msg.Reply, data, "")
	})
	if err != nil {
		t.Fatalf("subscribe control: %v", err)
	}
	record, err := mgr.AddAgent(ctx, AddRequest{
		Name:    "remote-alpha",
		Adapter: "stub",
		Location: api.Location{
			Type:     api.LocationSSH,
			Host:     hostID.String(),
			RepoPath: repoRoot,
		},
	})
	if err != nil {
		t.Fatalf("add remote agent: %v", err)
	}
	if record.Location == nil || record.Location.Type != api.LocationSSH || record.Location.Host != hostID.String() {
		t.Fatalf("unexpected location: %#v", record.Location)
	}
}

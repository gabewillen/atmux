package remote

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestDirectorSnapshotsAndIDs(t *testing.T) {
	host := api.MustParseHostID("host")
	peer := api.NewPeerID()
	director := &Director{
		hostID: host,
		peerID: peer,
		hosts: map[api.HostID]*hostState{
			host: {hostID: host, peerID: peer, connected: true, ready: false},
		},
	}
	if director.HostID() != host {
		t.Fatalf("unexpected host id")
	}
	if director.PeerID().Value() != peer.Value() {
		t.Fatalf("unexpected peer id")
	}
	snapshot, ok := director.HostSnapshot(host)
	if !ok || snapshot.HostID != host || snapshot.PeerID.Value() != peer.Value() {
		t.Fatalf("unexpected snapshot")
	}
	hosts := director.Hosts()
	if len(hosts) != 1 {
		t.Fatalf("expected hosts")
	}
	var nilDirector *Director
	if nilDirector.HostID() != "" {
		t.Fatalf("expected empty host id")
	}
	if !nilDirector.PeerID().IsZero() {
		t.Fatalf("expected zero peer id")
	}
}

func TestDirectorEnsureHostInvalidLocation(t *testing.T) {
	director := &Director{}
	if _, _, err := director.EnsureHost(testContext(t), api.Location{}, nil); err == nil {
		t.Fatalf("expected ensure host error")
	}
}

func TestDirectorEnsureHostSuccess(t *testing.T) {
	ctx := testContext(t)
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
	cfg := config.DefaultConfig(nil)
	cfg.NATS.JetStreamDir = jetDir
	cfg.NATS.HubURL = server.URL()
	cfg.Remote.NATS.URL = server.LeafURL()
	cfg.Remote.NATS.CredsPath = filepath.Join(jetDir, "host.creds")
	cfg.Remote.NATS.KVBucket = "kv"
	cfg.Remote.Manager.Model = "model"
	bootstrapper := &Bootstrapper{Runner: &sequenceRunner{
		outputs: map[string][][]byte{
			"uname -s": {[]byte(runtime.GOOS)},
			"uname -m": {[]byte(runtime.GOARCH)},
			"PATH=\"$HOME/.local/bin:$PATH\" amux-manager status": {
				[]byte("hub_connected=false"),
				[]byte("hub_connected=true"),
			},
		},
	}}
	director, err := NewDirector(cfg, dispatcher, DirectorOptions{Version: "test", HostID: api.MustParseHostID("director"), Bootstrapper: bootstrapper})
	if err != nil {
		t.Fatalf("new director: %v", err)
	}
	location := api.Location{Type: api.LocationSSH, Host: "example.com"}
	hostID, _, err := director.EnsureHost(ctx, location, nil)
	if err != nil {
		t.Fatalf("ensure host: %v", err)
	}
	if hostID == "" {
		t.Fatalf("expected host id")
	}
}

func TestDirectorKillNotReady(t *testing.T) {
	host := api.MustParseHostID("host")
	director := &Director{
		hosts: map[api.HostID]*hostState{
			host: {hostID: host, connected: false, ready: false},
		},
	}
	if _, err := director.Kill(testContext(t), host, KillRequest{SessionID: "1"}); err == nil {
		t.Fatalf("expected kill error")
	}
}

func TestDirectorAttachPTYNotReady(t *testing.T) {
	host := api.MustParseHostID("host")
	director := &Director{
		hosts: map[api.HostID]*hostState{
			host: {hostID: host, connected: false, ready: false},
		},
	}
	if _, err := director.AttachPTY(testContext(t), host, api.NewSessionID()); err == nil {
		t.Fatalf("expected attach error")
	}
}

func TestDirectorHandleCommMessage(t *testing.T) {
	director := &Director{logger: log.New(os.Stderr, "test ", log.LstdFlags)}
	payload := api.AgentMessage{
		ID:      api.NewRuntimeID(),
		From:    api.NewRuntimeID(),
		To:      api.TargetID{},
		ToSlug:  "broadcast",
		Content: "hello",
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	director.handleCommMessage(protocolMessage("amux.comm.broadcast", data))
}

func testContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}

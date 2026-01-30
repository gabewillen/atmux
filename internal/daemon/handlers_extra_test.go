package daemon

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestHandleStop(t *testing.T) {
	d := &Daemon{}
	raw := json.RawMessage(`{"force":true}`)
	resp, err := d.handleStop(context.Background(), raw)
	if err != nil {
		t.Fatalf("handle stop: %v", err)
	}
	if resp == nil {
		t.Fatalf("expected response")
	}
}

func TestHandleGitMergeInvalidParams(t *testing.T) {
	d := &Daemon{}
	_, err := d.handleGitMerge(context.Background(), json.RawMessage("bad"))
	if err == nil {
		t.Fatalf("expected invalid params error")
	}
}

func TestResolveAgentIDByName(t *testing.T) {
	repo := t.TempDir()
	initRepo(t, repo)
	resolver, err := paths.NewResolver(repo)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	cfg := config.DefaultConfig(resolver)
	cfg.NATS.JetStreamDir = filepath.Join(t.TempDir(), "jetstream")
	cfg.NATS.Listen = "127.0.0.1:-1"
	cfg.NATS.AdvertiseURL = ""
	cfg.NATS.LeafAdvertiseURL = ""
	cfg.Daemon.SocketPath = filepath.Join(t.TempDir(), "amuxd.sock")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	daemon, err := New(ctx, resolver, cfg, log.New(os.Stderr, "test ", log.LstdFlags))
	if err != nil {
		t.Fatalf("new daemon: %v", err)
	}
	daemon.manager.SetRegistryFactory(func(resolver *paths.Resolver) (adapter.Registry, error) {
		_ = resolver
		return &stubRegistry{cmd: []string{"env", "AMUX_DAEMON_HELPER=1", os.Args[0], "-test.run=TestDaemonHelperProcess"}}, nil
	})
	addParams := agentAddParams{
		Name:    "alpha",
		Adapter: "stub",
		Location: locationParam{
			Type:     "local",
			RepoPath: repo,
		},
		Cwd: repo,
	}
	addRaw, _ := json.Marshal(addParams)
	addResp, rpcErr := daemon.handleAgentAdd(context.Background(), addRaw)
	if rpcErr != nil {
		t.Fatalf("agent add: %v", rpcErr)
	}
	addResult := addResp.(agentAddResult)
	got, err := daemon.resolveAgentID(agentRefParams{Name: "alpha"})
	if err != nil {
		t.Fatalf("resolve by name: %v", err)
	}
	if got != addResult.AgentID {
		t.Fatalf("unexpected agent id")
	}
	addResp2, rpcErr := daemon.handleAgentAdd(context.Background(), addRaw)
	if rpcErr != nil {
		t.Fatalf("agent add 2: %v", rpcErr)
	}
	addResult2 := addResp2.(agentAddResult)
	if addResult2.AgentID == addResult.AgentID {
		t.Fatalf("expected distinct agent ids")
	}
	if _, err := daemon.resolveAgentID(agentRefParams{Name: "alpha"}); err == nil {
		t.Fatalf("expected ambiguous name error")
	}
	if _, err := daemon.resolveAgentID(agentRefParams{Name: "missing"}); err == nil {
		t.Fatalf("expected missing agent error")
	}
	if _, err := daemon.resolveAgentID(agentRefParams{AgentID: "bad"}); err == nil {
		t.Fatalf("expected invalid agent id error")
	}
	if _, err := daemon.resolveAgentID(agentRefParams{}); err == nil {
		t.Fatalf("expected missing reference error")
	}
	agentID := api.NewAgentID()
	if _, err := daemon.resolveAgentID(agentRefParams{AgentID: agentID.String()}); err != nil {
		t.Fatalf("expected agent id parse: %v", err)
	}
}

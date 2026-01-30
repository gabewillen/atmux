package manager

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/git"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestLoadFromConfigLocalAndSSH(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.Mkdir(filepath.Join(repoRoot, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	resolver, err := paths.NewResolver(repoRoot)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	cfg := config.Config{
		Agents: []config.AgentConfig{
			{
				Name:    "local",
				Adapter: "stub",
				Location: config.AgentLocationConfig{
					Type: "local",
				},
			},
			{
				Name:    "remote",
				Adapter: "stub",
				Location: config.AgentLocationConfig{
					Type:     "ssh",
					Host:     "host",
					RepoPath: "/tmp/repo",
				},
			},
		},
	}
	mgr := &Manager{
		resolver:  resolver,
		dispatcher: &recordDispatcher{},
		cfg:       cfg,
		agents:    make(map[api.AgentID]*agentState),
		nameIndex: make(map[string][]api.AgentID),
		bases:     make(map[string]string),
		git: &git.Runner{Exec: func(ctx context.Context, dir string, args ...string) (git.ExecResult, error) {
			_ = ctx
			_ = dir
			_ = args
			return git.ExecResult{Output: []byte("main")}, nil
		}},
	}
	if err := mgr.loadFromConfig(context.Background()); err != nil {
		t.Fatalf("load from config: %v", err)
	}
	if len(mgr.agents) != 2 {
		t.Fatalf("expected agents loaded")
	}
}

func TestAddAgentRemoteDirectorMissing(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.Mkdir(filepath.Join(repoRoot, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	resolver, err := paths.NewResolver(repoRoot)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	mgr := &Manager{
		resolver:  resolver,
		dispatcher: &recordDispatcher{},
		cfg:       config.Config{},
		agents:    make(map[api.AgentID]*agentState),
		nameIndex: make(map[string][]api.AgentID),
	}
	_, err = mgr.AddAgent(context.Background(), AddRequest{
		Name:     "remote",
		Adapter:  "stub",
		Location: api.Location{Type: api.LocationSSH, Host: "host", RepoPath: "/tmp/repo"},
	})
	if err == nil {
		t.Fatalf("expected add agent error")
	}
}

func TestLoadFromConfigInvalidLocation(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.Mkdir(filepath.Join(repoRoot, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	resolver, err := paths.NewResolver(repoRoot)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	cfg := config.Config{
		Agents: []config.AgentConfig{
			{
				Name:    "bad",
				Adapter: "stub",
				Location: config.AgentLocationConfig{
					Type: "nope",
				},
			},
		},
	}
	mgr := &Manager{
		resolver:   resolver,
		dispatcher: &recordDispatcher{},
		cfg:        cfg,
		agents:     make(map[api.AgentID]*agentState),
		nameIndex:  make(map[string][]api.AgentID),
		bases:      make(map[string]string),
		git: &git.Runner{Exec: func(ctx context.Context, dir string, args ...string) (git.ExecResult, error) {
			return git.ExecResult{Output: []byte("main")}, nil
		}},
	}
	if err := mgr.loadFromConfig(context.Background()); err == nil {
		t.Fatalf("expected invalid location error")
	}
}

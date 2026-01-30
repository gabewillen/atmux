package manager

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/internal/session"
	"github.com/agentflare-ai/amux/pkg/api"
)

type errorRegistry struct {
	loadErr error
}

func (e errorRegistry) Load(ctx context.Context, name string) (adapter.Adapter, error) {
	_ = ctx
	_ = name
	return nil, e.loadErr
}

func TestStartSessionErrors(t *testing.T) {
	repoRoot := initRepo(t)
	resolver, err := paths.NewResolver(repoRoot)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	disp := &recordDispatcher{}
	agentID := api.NewAgentID()
	worktree := paths.WorktreePathForRepo(repoRoot, "alpha")
	location := api.Location{Type: api.LocationLocal, RepoPath: repoRoot}
	meta, err := api.NewAgentWithID(agentID, "alpha", "", "stub", repoRoot, worktree, location)
	if err != nil {
		t.Fatalf("agent meta: %v", err)
	}
	runtime, err := agent.NewAgent(meta, disp)
	if err != nil {
		t.Fatalf("runtime: %v", err)
	}
	state := &agentState{
		runtime: runtime,
		slug:    "alpha",
		repoRoot: repoRoot,
		worktree: filepath.Join(repoRoot, ".amux", "worktrees", "alpha"),
		config: config.AgentConfig{
			Name:    "alpha",
			Adapter: "stub",
			Location: config.AgentLocationConfig{
				Type:     api.LocationLocal.String(),
				RepoPath: repoRoot,
			},
		},
	}
	manager := &Manager{
		resolver:        resolver,
		dispatcher:      disp,
		cfg:             config.DefaultConfig(resolver),
		agents:          map[api.AgentID]*agentState{agentID: state},
		registries:      map[string]adapter.Registry{},
		registryFactory: func(*paths.Resolver) (adapter.Registry, error) { return &stubRegistry{cmd: nil}, nil },
	}
	if _, err := manager.startSession(context.Background(), agentID); err == nil {
		t.Fatalf("expected invalid manifest error")
	}
	manager.registryFactory = func(*paths.Resolver) (adapter.Registry, error) {
		return nil, fmt.Errorf("factory error")
	}
	if _, err := manager.startSession(context.Background(), agentID); err == nil {
		t.Fatalf("expected registry factory error")
	}
	manager.registryFactory = func(*paths.Resolver) (adapter.Registry, error) {
		return errorRegistry{loadErr: fmt.Errorf("load error")}, nil
	}
	if _, err := manager.startSession(context.Background(), agentID); err == nil {
		t.Fatalf("expected adapter load error")
	}
}

func TestStartSessionAlreadyRunning(t *testing.T) {
	repoRoot := initRepo(t)
	resolver, err := paths.NewResolver(repoRoot)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	agentID := api.NewAgentID()
	existing := &session.LocalSession{}
	manager := &Manager{
		resolver:   resolver,
		agents:    map[api.AgentID]*agentState{agentID: {session: existing}},
		registries: map[string]adapter.Registry{},
	}
	sess, err := manager.startSession(context.Background(), agentID)
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	if sess != existing {
		t.Fatalf("expected existing session")
	}
}

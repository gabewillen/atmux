package manager

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/git"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestExtractEncodeAgents(t *testing.T) {
	raw := map[string]any{
		"agents": []any{
			map[string]any{
				"name":    "alpha",
				"about":   "test",
				"adapter": "stub",
				"listen_channels": []any{
					"events",
				},
				"location": map[string]any{
					"type":      "local",
					"host":      "host",
					"repo_path": "/tmp/repo",
				},
			},
		},
	}
	agents := extractAgents(raw)
	if len(agents) != 1 || agents[0].Name != "alpha" {
		t.Fatalf("unexpected agents: %#v", agents)
	}
	encoded := encodeAgents(agents)
	if len(encoded) != 1 {
		t.Fatalf("unexpected encoded length")
	}
}

func TestSameAgentAndList(t *testing.T) {
	a := config.AgentConfig{
		Name:    "alpha",
		Adapter: "stub",
		ListenChannels: []string{
			"one",
		},
		Location: config.AgentLocationConfig{
			Type:     "local",
			Host:     "host",
			RepoPath: "/tmp/repo",
		},
	}
	b := a
	b.Location.RepoPath = filepath.Join("/tmp", "repo")
	if !sameAgent(a, b) {
		t.Fatalf("expected same agent")
	}
	b.ListenChannels = []string{"two"}
	if sameAgent(a, b) {
		t.Fatalf("expected different agent")
	}
	if !sameStringList([]string{"a"}, []string{"a"}) {
		t.Fatalf("expected same string list")
	}
	if sameStringList([]string{"a"}, []string{"b"}) {
		t.Fatalf("expected different string list")
	}
}

func TestEnsureGitRepo(t *testing.T) {
	if err := ensureGitRepo(""); err == nil {
		t.Fatalf("expected error for empty repo")
	}
	repo := t.TempDir()
	gitFile := filepath.Join(repo, ".git")
	if err := os.WriteFile(gitFile, []byte("gitdir: /tmp/gitdir\n"), 0o644); err != nil {
		t.Fatalf("write git file: %v", err)
	}
	if err := ensureGitRepo(repo); err != nil {
		t.Fatalf("expected git repo ok: %v", err)
	}
}

func TestStatePresenceHelpers(t *testing.T) {
	if statePresence(nil) != agent.PresenceOffline {
		t.Fatalf("expected offline for nil state")
	}
	state := &agentState{presence: "BUSY"}
	if statePresence(state) != "busy" {
		t.Fatalf("expected busy")
	}
	if lastStateSegment("presence/online") != "online" {
		t.Fatalf("unexpected state segment")
	}
}

func TestRemoveNameIndexLocked(t *testing.T) {
	mgr := &Manager{nameIndex: map[string][]api.AgentID{}}
	id1 := api.NewAgentID()
	id2 := api.NewAgentID()
	mgr.nameIndex["alpha"] = []api.AgentID{id1, id2}
	mgr.removeNameIndexLocked("alpha", id1)
	if len(mgr.nameIndex["alpha"]) != 1 || mgr.nameIndex["alpha"][0] != id2 {
		t.Fatalf("unexpected name index")
	}
	mgr.removeNameIndexLocked("alpha", id2)
	if _, ok := mgr.nameIndex["alpha"]; ok {
		t.Fatalf("expected name index removed")
	}
}

func TestRemoveConfigEntryLocked(t *testing.T) {
	entry := config.AgentConfig{Name: "alpha", Adapter: "stub"}
	mgr := &Manager{cfg: config.Config{Agents: []config.AgentConfig{entry}}}
	mgr.removeConfigEntryLocked(entry)
	if len(mgr.cfg.Agents) != 0 {
		t.Fatalf("expected config entry removed")
	}
}

func TestResolveLocation(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir git: %v", err)
	}
	resolver, err := paths.NewResolver(repo)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	mgr := &Manager{resolver: resolver, agents: map[api.AgentID]*agentState{}}
	loc, root, err := mgr.resolveLocation(AddRequest{
		Location: api.Location{Type: api.LocationLocal, RepoPath: repo},
	})
	if err != nil {
		t.Fatalf("resolve location: %v", err)
	}
	if root == "" || loc.RepoPath == "" {
		t.Fatalf("expected repo root")
	}
	if _, _, err := mgr.resolveLocation(AddRequest{Location: api.Location{Type: api.LocationSSH}}); err == nil {
		t.Fatalf("expected ssh host error")
	}
	if _, _, err := mgr.resolveLocation(AddRequest{Location: api.Location{Type: api.LocationSSH, Host: "host"}}); err == nil {
		t.Fatalf("expected ssh repo path error")
	}
	if _, _, err := mgr.resolveLocation(AddRequest{Location: api.Location{Type: api.LocationType(99)}}); err == nil {
		t.Fatalf("expected invalid location error")
	}
}

func TestValidateMultiRepo(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir git: %v", err)
	}
	resolver, err := paths.NewResolver(repo)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	mgr := &Manager{resolver: resolver, agents: map[api.AgentID]*agentState{}}
	if err := mgr.validateMultiRepo("", false); err == nil {
		t.Fatalf("expected empty repo error")
	}
	other := filepath.Join(repo, "other")
	if err := os.MkdirAll(filepath.Join(other, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir other git: %v", err)
	}
	mgr.agents[api.NewAgentID()] = &agentState{repoRoot: repo, explicitRepoPath: true}
	if err := mgr.validateMultiRepo(other, true); err != nil {
		t.Fatalf("expected validate ok: %v", err)
	}
	if err := mgr.validateMultiRepo(other, false); err == nil {
		t.Fatalf("expected repo path required error")
	}
}

func TestFindAgent(t *testing.T) {
	mgr := &Manager{agents: map[api.AgentID]*agentState{}, nameIndex: map[string][]api.AgentID{}}
	if _, _, err := mgr.findAgent(RemoveRequest{}); err == nil {
		t.Fatalf("expected find agent error")
	}
	id := api.NewAgentID()
	mgr.agents[id] = &agentState{}
	mgr.nameIndex["alpha"] = []api.AgentID{id}
	if _, found, err := mgr.findAgent(RemoveRequest{Name: "alpha"}); err != nil || found != id {
		t.Fatalf("expected agent found")
	}
	mgr.nameIndex["alpha"] = []api.AgentID{id, api.NewAgentID()}
	if _, _, err := mgr.findAgent(RemoveRequest{Name: "alpha"}); err == nil {
		t.Fatalf("expected ambiguous error")
	}
}

func TestBuildAdapterBundleErrors(t *testing.T) {
	if _, err := buildAdapterBundle(nil, ""); err == nil {
		t.Fatalf("expected resolver error")
	}
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir git: %v", err)
	}
	resolver, err := paths.NewResolver(repo)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	if _, err := buildAdapterBundle(resolver, ""); err == nil {
		t.Fatalf("expected name error")
	}
	if _, err := buildAdapterBundle(resolver, "missing"); err == nil {
		t.Fatalf("expected missing wasm error")
	}
}

func TestSessionControlErrors(t *testing.T) {
	mgr := &Manager{agents: map[api.AgentID]*agentState{}}
	id := api.NewAgentID()
	if err := mgr.StartAgent(context.Background(), id); err == nil {
		t.Fatalf("expected start agent error")
	}
	if err := mgr.StopAgent(context.Background(), id); err == nil {
		t.Fatalf("expected stop agent error")
	}
	if err := mgr.KillAgent(context.Background(), id); err == nil {
		t.Fatalf("expected kill agent error")
	}
	if err := mgr.RestartAgent(context.Background(), id); err == nil {
		t.Fatalf("expected restart agent error")
	}
}

func TestShutdownErrors(t *testing.T) {
	var mgr *Manager
	if err := mgr.Shutdown(context.Background(), false); err != nil {
		t.Fatalf("expected nil shutdown")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	mgr = &Manager{}
	if err := mgr.Shutdown(ctx, false); err == nil {
		t.Fatalf("expected shutdown context error")
	}
}

func TestMergeAgent(t *testing.T) {
	repo := t.TempDir()
	worktree := filepath.Join(repo, "wt")
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}
	fakeGit := &git.Runner{Exec: func(ctx context.Context, dir string, args ...string) (git.ExecResult, error) {
		if len(args) > 0 && args[0] == "show-ref" {
			return git.ExecResult{ExitCode: 0}, nil
		}
		if len(args) > 0 && args[0] == "status" {
			return git.ExecResult{ExitCode: 0}, nil
		}
		return git.ExecResult{ExitCode: 0}, nil
	}}
	disp := &recordDispatcher{}
	id := api.NewAgentID()
	state := &agentState{repoRoot: repo, worktree: worktree, slug: "alpha"}
	mgr := &Manager{
		git:        fakeGit,
		dispatcher: disp,
		agents:     map[api.AgentID]*agentState{id: state},
		bases:      map[string]string{repo: "main"},
	}
	if _, err := mgr.MergeAgent(context.Background(), id, git.StrategySquash, ""); err != nil {
		t.Fatalf("merge agent: %v", err)
	}
	if len(disp.events) == 0 {
		t.Fatalf("expected events")
	}
}

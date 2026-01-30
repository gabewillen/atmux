package remote

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/internal/session"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestHandleKillSession(t *testing.T) {
	repoRoot := initRepo(t)
	worktree := paths.WorktreePathForRepo(repoRoot, "alpha")
	location := api.Location{Type: api.LocationSSH, Host: "host", RepoPath: repoRoot}
	agentID := api.NewAgentID()
	meta, err := api.NewAgentWithID(agentID, "alpha", "", "stub", repoRoot, worktree, location)
	if err != nil {
		t.Fatalf("agent meta: %v", err)
	}
	dispatcher := &stubDispatcher{}
	runtime, err := agent.NewAgent(meta, dispatcher)
	if err != nil {
		t.Fatalf("runtime: %v", err)
	}
	sessionMeta, err := api.NewSession(agentID, repoRoot, worktree, location)
	if err != nil {
		t.Fatalf("session meta: %v", err)
	}
	cmd := session.Command{Argv: []string{os.Args[0], "-test.run=TestRemoteHelperProcess"}}
	if _, err := os.Stat(os.Args[0]); err != nil {
		t.Skip("test binary unavailable")
	}
	sess, err := session.NewLocalSession(sessionMeta, runtime, cmd, worktree, stubMatcher{}, dispatcher, session.Config{})
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	if err := sess.Start(ctx); err != nil {
		t.Skipf("start session failed: %v", err)
	}
	manager := &HostManager{
		hostID:     api.MustParseHostID("host"),
		dispatcher: dispatcher,
		sessions:   map[api.SessionID]*remoteSession{},
		agentIndex: map[api.AgentID]*remoteSession{},
	}
	remoteSess := &remoteSession{
		agentID:   agentID,
		sessionID: sessionMeta.ID,
		runtime:   sess,
		slug:      "alpha",
		repoPath:  repoRoot,
	}
	manager.sessions[sessionMeta.ID] = remoteSess
	manager.agentIndex[agentID] = remoteSess
	reply := "reply.kill"
	payload, err := EncodePayload("kill", KillRequest{SessionID: sessionMeta.ID.String()})
	if err != nil {
		t.Fatalf("encode kill: %v", err)
	}
	manager.handleKill(reply, payload)
	control, err := DecodeControlMessage(dispatcher.lastPayload)
	if err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if control.Type != "kill" {
		t.Fatalf("expected kill response")
	}
	if _, ok := manager.sessions[sessionMeta.ID]; ok {
		t.Fatalf("expected session removed")
	}
}

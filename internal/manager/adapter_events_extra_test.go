package manager

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/session"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestHandleActionSendInput(t *testing.T) {
	t.Parallel()
	mgr := &Manager{logger: log.New(os.Stderr, "", 0)}
	state := &agentState{session: &session.LocalSession{}}
	payload, err := json.Marshal(actionSendInput{DataB64: base64.StdEncoding.EncodeToString([]byte("hi"))})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	mgr.handleActionSendInput(context.Background(), state, payload)
}

func TestHandleActionUpdatePresence(t *testing.T) {
	t.Parallel()
	dispatcher := &recordDispatcher{}
	repoRoot := t.TempDir()
	worktree := filepath.Join(repoRoot, "worktree")
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}
	meta := api.Agent{
		ID:       api.NewAgentID(),
		Name:     "agent",
		Adapter:  "adapter",
		RepoRoot: repoRoot,
		Worktree: worktree,
		Location: api.Location{Type: api.LocationLocal},
	}
	runtime, err := agent.NewAgent(meta, dispatcher)
	if err != nil {
		t.Fatalf("new agent: %v", err)
	}
	state := &agentState{runtime: runtime, presence: agent.PresenceOnline}
	mgr := &Manager{logger: log.New(os.Stderr, "", 0)}
	payload, err := json.Marshal(actionUpdatePresence{Presence: agent.PresenceBusy})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	mgr.handleActionUpdatePresence(context.Background(), state, payload)
	if len(dispatcher.events) == 0 {
		t.Fatalf("expected presence events")
	}
}

func TestHandleActionEmitEvent(t *testing.T) {
	t.Parallel()
	dispatcher := &recordDispatcher{}
	mgr := &Manager{dispatcher: dispatcher}
	payload, err := json.Marshal(actionEmitEvent{Event: adapter.Event{Type: "custom"}})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	mgr.handleActionEmitEvent(context.Background(), payload)
	if len(dispatcher.events) == 0 {
		t.Fatalf("expected emitted event")
	}
}

func TestPresenceTransitionEvents(t *testing.T) {
	t.Parallel()
	if events := presenceTransitionEvents(agent.PresenceOnline, agent.PresenceOnline); events != nil {
		t.Fatalf("expected nil events for same presence")
	}
	events := presenceTransitionEvents(agent.PresenceOnline, agent.PresenceBusy)
	if len(events) != 1 || events[0] != agent.EventTaskAssigned {
		t.Fatalf("unexpected events: %#v", events)
	}
	events = presenceTransitionEvents(agent.PresenceAway, agent.PresenceOffline)
	if len(events) != 2 || events[0] != agent.EventActivity {
		t.Fatalf("unexpected away->offline events: %#v", events)
	}
}

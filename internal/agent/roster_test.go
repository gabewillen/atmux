package agent

import (
	"testing"

	"github.com/stateforward/hsm-go/muid"

	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestRoster_SetDirector(t *testing.T) {
	r := NewRoster(event.NewNoopDispatcher())
	id := muid.Make()

	r.SetDirector(id, "Director", "Main director")

	p := r.Get(id)
	if p == nil {
		t.Fatal("expected director to be registered")
	}
	if p.Type != api.ParticipantDirector {
		t.Errorf("expected type %v, got %v", api.ParticipantDirector, p.Type)
	}
	if p.Slug != api.DirectorSlug {
		t.Errorf("expected slug %q, got %q", api.DirectorSlug, p.Slug)
	}
	if r.DirectorID() != id {
		t.Errorf("expected director ID %v, got %v", id, r.DirectorID())
	}
}

func TestRoster_SetDirector_ReplacesPrevious(t *testing.T) {
	r := NewRoster(event.NewNoopDispatcher())
	id1 := muid.Make()
	id2 := muid.Make()

	r.SetDirector(id1, "Director1", "First director")
	r.SetDirector(id2, "Director2", "Second director")

	if r.Get(id1) != nil {
		t.Error("expected first director to be removed")
	}
	if r.DirectorID() != id2 {
		t.Errorf("expected director ID %v, got %v", id2, r.DirectorID())
	}
}

func TestRoster_AddManager(t *testing.T) {
	r := NewRoster(event.NewNoopDispatcher())
	id := muid.Make()

	r.AddManager(id, "Manager", "host-1", "Remote manager")

	p := r.Get(id)
	if p == nil {
		t.Fatal("expected manager to be registered")
	}
	if p.Type != api.ParticipantManager {
		t.Errorf("expected type %v, got %v", api.ParticipantManager, p.Type)
	}
	if p.Slug != "manager@host-1" {
		t.Errorf("expected slug %q, got %q", "manager@host-1", p.Slug)
	}
	if p.HostID != "host-1" {
		t.Errorf("expected host ID %q, got %q", "host-1", p.HostID)
	}
}

func TestRoster_AddManager_Local(t *testing.T) {
	r := NewRoster(event.NewNoopDispatcher())
	id := muid.Make()

	r.AddManager(id, "Manager", "", "Local manager")

	p := r.Get(id)
	if p == nil {
		t.Fatal("expected manager to be registered")
	}
	if p.Slug != api.ManagerSlug {
		t.Errorf("expected slug %q, got %q", api.ManagerSlug, p.Slug)
	}
}

func TestRoster_AddAgent(t *testing.T) {
	r := NewRoster(event.NewNoopDispatcher())
	agent := &api.Agent{
		ID:       muid.Make(),
		Name:     "Test Agent",
		Slug:     "test-agent",
		About:    "A test agent",
		Adapter:  "test",
		RepoRoot: "/repo",
	}

	r.AddAgent(agent, api.LifecycleRunning, api.PresenceOnline)

	p := r.Get(agent.ID)
	if p == nil {
		t.Fatal("expected agent to be registered")
	}
	if p.Type != api.ParticipantAgent {
		t.Errorf("expected type %v, got %v", api.ParticipantAgent, p.Type)
	}
	if p.Slug != "test-agent" {
		t.Errorf("expected slug %q, got %q", "test-agent", p.Slug)
	}
	if p.Lifecycle != api.LifecycleRunning {
		t.Errorf("expected lifecycle %v, got %v", api.LifecycleRunning, p.Lifecycle)
	}
	if p.Presence != api.PresenceOnline {
		t.Errorf("expected presence %v, got %v", api.PresenceOnline, p.Presence)
	}
}

func TestRoster_GetBySlug(t *testing.T) {
	r := NewRoster(event.NewNoopDispatcher())

	dirID := muid.Make()
	r.SetDirector(dirID, "Director", "")

	mgrID := muid.Make()
	r.AddManager(mgrID, "Manager", "host-1", "")

	agentID := muid.Make()
	agent := &api.Agent{
		ID:       agentID,
		Name:     "Test Agent",
		Slug:     "test-agent",
		Adapter:  "test",
		RepoRoot: "/repo",
	}
	r.AddAgent(agent, api.LifecycleRunning, api.PresenceOnline)

	tests := []struct {
		slug     string
		wantID   muid.MUID
		wantType api.ParticipantType
	}{
		{"director", dirID, api.ParticipantDirector},
		{"DIRECTOR", dirID, api.ParticipantDirector}, // case-insensitive
		{"manager@host-1", mgrID, api.ParticipantManager},
		{"MANAGER@HOST-1", mgrID, api.ParticipantManager}, // case-insensitive
		{"test-agent", agentID, api.ParticipantAgent},
		{"TEST-AGENT", agentID, api.ParticipantAgent}, // case-insensitive
	}

	for _, tc := range tests {
		t.Run(tc.slug, func(t *testing.T) {
			p := r.GetBySlug(tc.slug)
			if p == nil {
				t.Fatalf("expected to find participant with slug %q", tc.slug)
			}
			if p.ID != tc.wantID {
				t.Errorf("expected ID %v, got %v", tc.wantID, p.ID)
			}
			if p.Type != tc.wantType {
				t.Errorf("expected type %v, got %v", tc.wantType, p.Type)
			}
		})
	}
}

func TestRoster_RemoveParticipant(t *testing.T) {
	r := NewRoster(event.NewNoopDispatcher())
	id := muid.Make()

	r.SetDirector(id, "Director", "")
	r.RemoveParticipant(id)

	if r.Get(id) != nil {
		t.Error("expected participant to be removed")
	}
	if r.DirectorID() != 0 {
		t.Error("expected director ID to be cleared")
	}
}

func TestRoster_UpdatePresence(t *testing.T) {
	r := NewRoster(event.NewNoopDispatcher())
	id := muid.Make()
	agent := &api.Agent{
		ID:       id,
		Name:     "Test",
		Slug:     "test",
		Adapter:  "test",
		RepoRoot: "/repo",
	}
	r.AddAgent(agent, api.LifecycleRunning, api.PresenceOnline)

	r.UpdatePresence(id, api.PresenceBusy)

	p := r.Get(id)
	if p.Presence != api.PresenceBusy {
		t.Errorf("expected presence %v, got %v", api.PresenceBusy, p.Presence)
	}
}

func TestRoster_UpdateLifecycle(t *testing.T) {
	r := NewRoster(event.NewNoopDispatcher())
	id := muid.Make()
	agent := &api.Agent{
		ID:       id,
		Name:     "Test",
		Slug:     "test",
		Adapter:  "test",
		RepoRoot: "/repo",
	}
	r.AddAgent(agent, api.LifecycleRunning, api.PresenceOnline)

	r.UpdateLifecycle(id, api.LifecycleTerminated)

	p := r.Get(id)
	if p.Lifecycle != api.LifecycleTerminated {
		t.Errorf("expected lifecycle %v, got %v", api.LifecycleTerminated, p.Lifecycle)
	}
}

func TestRoster_List(t *testing.T) {
	r := NewRoster(event.NewNoopDispatcher())

	r.SetDirector(muid.Make(), "Director", "")
	r.AddManager(muid.Make(), "Manager", "host-1", "")
	r.AddAgent(&api.Agent{
		ID:       muid.Make(),
		Name:     "Agent",
		Slug:     "agent",
		Adapter:  "test",
		RepoRoot: "/repo",
	}, api.LifecycleRunning, api.PresenceOnline)

	list := r.List()
	if len(list) != 3 {
		t.Errorf("expected 3 participants, got %d", len(list))
	}
}

func TestRoster_ListAgents(t *testing.T) {
	r := NewRoster(event.NewNoopDispatcher())

	r.SetDirector(muid.Make(), "Director", "")
	r.AddAgent(&api.Agent{
		ID:       muid.Make(),
		Name:     "Agent1",
		Slug:     "agent1",
		Adapter:  "test",
		RepoRoot: "/repo",
	}, api.LifecycleRunning, api.PresenceOnline)
	r.AddAgent(&api.Agent{
		ID:       muid.Make(),
		Name:     "Agent2",
		Slug:     "agent2",
		Adapter:  "test",
		RepoRoot: "/repo",
	}, api.LifecycleRunning, api.PresenceOnline)

	agents := r.ListAgents()
	if len(agents) != 2 {
		t.Errorf("expected 2 agents, got %d", len(agents))
	}
	for _, a := range agents {
		if a.Type != api.ParticipantAgent {
			t.Errorf("expected agent type, got %v", a.Type)
		}
	}
}

func TestEqualFoldASCII(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"director", "DIRECTOR", true},
		{"Director", "director", true},
		{"test", "test", true},
		{"test", "TEST", true},
		{"test", "test1", false},
		{"", "", true},
		{"a", "b", false},
	}

	for _, tc := range tests {
		t.Run(tc.a+"_"+tc.b, func(t *testing.T) {
			got := equalFoldASCII(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("equalFoldASCII(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

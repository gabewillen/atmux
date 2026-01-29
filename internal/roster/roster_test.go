// Package roster_test provides tests for the roster implementation.
package roster

import (
	"testing"
	"time"

	"github.com/stateforward/hsm-go/muid"

	"github.com/copilot-claude-sonnet-4/amux/pkg/api"
)

func TestStore_AddAgent(t *testing.T) {
	store := NewStore()
	defer store.Close()

	// Create test agent
	agentID := muid.Make()
	agent := &api.Agent{
		ID:       agentID,
		Slug:     "test-agent",
		Name:     "Test Agent",
		Presence: api.PresenceOnline,
		State:    api.AgentStateRunning,
		Adapter:  "test-adapter",
	}

	// Add agent to roster
	err := store.AddAgent(agent, "localhost")
	if err != nil {
		t.Fatalf("Failed to add agent: %v", err)
	}

	// Verify agent was added
	entry, err := store.GetByID(agentID)
	if err != nil {
		t.Fatalf("Failed to get agent by ID: %v", err)
	}

	if entry.Type != "agent" {
		t.Errorf("Expected type 'agent', got %s", entry.Type)
	}
	if entry.Name != "Test Agent" {
		t.Errorf("Expected name 'Test Agent', got %s", entry.Name)
	}
	if entry.Presence != api.PresenceOnline {
		t.Errorf("Expected presence online, got %s", entry.Presence)
	}

	// Verify slug lookup
	entryBySlug, err := store.GetBySlug("test-agent")
	if err != nil {
		t.Fatalf("Failed to get agent by slug: %v", err)
	}
	if entryBySlug.ID != agentID {
		t.Errorf("Expected ID %s, got %s", agentID, entryBySlug.ID)
	}
}

func TestStore_PresenceChangeEvents(t *testing.T) {
	store := NewStore()
	defer store.Close()

	// Subscribe to presence changes
	eventsCh := store.Subscribe()

	// Create and add agent
	agentID := muid.Make()
	agent := &api.Agent{
		ID:       agentID,
		Slug:     "test-agent",
		Name:     "Test Agent",
		Presence: api.PresenceOnline,
		State:    api.AgentStateRunning,
	}

	err := store.AddAgent(agent, "localhost")
	if err != nil {
		t.Fatalf("Failed to add agent: %v", err)
	}

	// Update presence
	err = store.UpdatePresence(agentID, api.PresenceBusy)
	if err != nil {
		t.Fatalf("Failed to update presence: %v", err)
	}

	// Wait for presence change event
	select {
	case event := <-eventsCh:
		if event.ParticipantID != agentID {
			t.Errorf("Expected participant ID %s, got %s", agentID, event.ParticipantID)
		}
		if event.OldPresence != api.PresenceOnline {
			t.Errorf("Expected old presence online, got %s", event.OldPresence)
		}
		if event.NewPresence != api.PresenceBusy {
			t.Errorf("Expected new presence busy, got %s", event.NewPresence)
		}
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for presence change event")
	}
}

func TestStore_ListOperations(t *testing.T) {
	store := NewStore()
	defer store.Close()

	// Add multiple participants
	directorID := muid.Make()
	err := store.AddDirector(directorID, api.PresenceOnline)
	if err != nil {
		t.Fatalf("Failed to add director: %v", err)
	}

	managerID := muid.Make()
	err = store.AddManager(managerID, "localhost", api.PresenceOnline)
	if err != nil {
		t.Fatalf("Failed to add manager: %v", err)
	}

	agent1ID := muid.Make()
	agent1 := &api.Agent{
		ID:       agent1ID,
		Slug:     "agent1",
		Name:     "Agent 1",
		Presence: api.PresenceOnline,
		State:    api.AgentStateRunning,
	}
	err = store.AddAgent(agent1, "localhost")
	if err != nil {
		t.Fatalf("Failed to add agent 1: %v", err)
	}

	agent2ID := muid.Make()
	agent2 := &api.Agent{
		ID:       agent2ID,
		Slug:     "agent2",
		Name:     "Agent 2",
		Presence: api.PresenceBusy,
		State:    api.AgentStateRunning,
	}
	err = store.AddAgent(agent2, "localhost")
	if err != nil {
		t.Fatalf("Failed to add agent 2: %v", err)
	}

	// Test ListAll
	allEntries := store.ListAll()
	if len(allEntries) != 4 {
		t.Errorf("Expected 4 entries, got %d", len(allEntries))
	}

	// Test ListAgents
	agents := store.ListAgents()
	if len(agents) != 2 {
		t.Errorf("Expected 2 agents, got %d", len(agents))
	}

	// Test ListByPresence
	onlineEntries := store.ListByPresence(api.PresenceOnline)
	if len(onlineEntries) != 3 { // director + manager + agent1
		t.Errorf("Expected 3 online entries, got %d", len(onlineEntries))
	}

	busyEntries := store.ListByPresence(api.PresenceBusy)
	if len(busyEntries) != 1 { // agent2
		t.Errorf("Expected 1 busy entry, got %d", len(busyEntries))
	}
}
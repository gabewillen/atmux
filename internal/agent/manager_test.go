package agent

import (
	"testing"
	"time"

	"github.com/copilot-claude-sonnet-4/amux/pkg/api"
)

func TestNewAgentActor(t *testing.T) {
	tests := []struct {
		name     string
		agentName string
		adapter  string
		repoRoot string
		config   map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "valid agent",
			agentName: "test-agent",
			adapter:  "test-adapter",
			repoRoot: "/tmp/repo",
			config:   map[string]interface{}{"model": "test-model"},
			wantErr:  false,
		},
		{
			name:     "empty name",
			agentName: "",
			adapter:  "test-adapter",
			repoRoot: "/tmp/repo",
			config:   nil,
			wantErr:  true,
		},
		{
			name:     "invalid name with non-printable characters",
			agentName: "agent\x00name",
			adapter:  "test-adapter",
			repoRoot: "/tmp/repo",
			config:   nil,
			wantErr:  true,
		},
		{
			name:     "empty repo root",
			agentName: "valid-name",
			adapter:  "test-adapter",
			repoRoot: "",
			config:   nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actor, err := NewAgentActor(tt.agentName, tt.adapter, tt.repoRoot, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAgentActor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if actor == nil {
					t.Error("NewAgentActor() returned nil actor")
					return
				}

				// Verify initial states
				if actor.Agent.State != api.AgentStatePending {
					t.Errorf("NewAgentActor() initial state = %v, want %v", actor.Agent.State, api.AgentStatePending)
				}

				if actor.Agent.Presence != api.PresenceOffline {
					t.Errorf("NewAgentActor() initial presence = %v, want %v", actor.Agent.Presence, api.PresenceOffline)
				}

				// Verify agent fields
				if actor.Agent.Name != tt.agentName {
					t.Errorf("NewAgentActor() name = %v, want %v", actor.Agent.Name, tt.agentName)
				}

				if actor.Agent.Adapter != tt.adapter {
					t.Errorf("NewAgentActor() adapter = %v, want %v", actor.Agent.Adapter, tt.adapter)
				}

				if actor.Agent.ID == 0 {
					t.Error("NewAgentActor() generated zero ID")
				}

				if actor.Agent.Slug == "" {
					t.Error("NewAgentActor() generated empty slug")
				}
			}
		})
	}
}

func TestAgentActorLifecycleTransitions(t *testing.T) {
	actor, err := NewAgentActor("test-agent", "test-adapter", "/tmp/repo", nil)
	if err != nil {
		t.Fatalf("NewAgentActor() failed: %v", err)
	}

	// Test valid lifecycle transitions
	tests := []struct {
		name      string
		event     LifecycleEvent
		wantState api.AgentState
		wantErr   bool
	}{
		{
			name:      "pending to starting",
			event:     EventStart,
			wantState: api.AgentStateStarting,
			wantErr:   false,
		},
		{
			name:      "starting to running",
			event:     EventStartupComplete,
			wantState: api.AgentStateRunning,
			wantErr:   false,
		},
		{
			name:      "running to terminated",
			event:     EventTerminate,
			wantState: api.AgentStateTerminated,
			wantErr:   false,
		},
		{
			name:      "terminated to pending",
			event:     EventRestart,
			wantState: api.AgentStatePending,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := actor.SendLifecycleEvent(tt.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendLifecycleEvent(%v) error = %v, wantErr %v", tt.event, err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				state := actor.GetState()
				if state != tt.wantState {
					t.Errorf("SendLifecycleEvent(%v) state = %v, want %v", tt.event, state, tt.wantState)
				}
			}
		})
	}
}

func TestAgentActorPresenceTransitions(t *testing.T) {
	actor, err := NewAgentActor("test-agent", "test-adapter", "/tmp/repo", nil)
	if err != nil {
		t.Fatalf("NewAgentActor() failed: %v", err)
	}

	// Test valid presence transitions
	tests := []struct {
		name         string
		event        PresenceEvent
		wantPresence api.PresenceState
		wantErr      bool
	}{
		{
			name:         "offline to online",
			event:        EventGoOnline,
			wantPresence: api.PresenceOnline,
			wantErr:      false,
		},
		{
			name:         "online to busy",
			event:        EventGoBusy,
			wantPresence: api.PresenceBusy,
			wantErr:      false,
		},
		{
			name:         "busy to online",
			event:        EventGoOnline,
			wantPresence: api.PresenceOnline,
			wantErr:      false,
		},
		{
			name:         "online to away",
			event:        EventGoAway,
			wantPresence: api.PresenceAway,
			wantErr:      false,
		},
		{
			name:         "away to online",
			event:        EventActivity,
			wantPresence: api.PresenceOnline,
			wantErr:      false,
		},
		{
			name:         "online to offline",
			event:        EventGoOffline,
			wantPresence: api.PresenceOffline,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := actor.SendPresenceEvent(tt.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendPresenceEvent(%v) error = %v, wantErr %v", tt.event, err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				presence := actor.GetPresence()
				if presence != tt.wantPresence {
					t.Errorf("SendPresenceEvent(%v) presence = %v, want %v", tt.event, presence, tt.wantPresence)
				}
			}
		})
	}
}

func TestManager(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	// Test adding agents
	agent1, err := manager.AddAgent("agent1", "test-adapter-1", "/tmp/repo1", map[string]interface{}{"model": "test-model"})
	if err != nil {
		t.Fatalf("AddAgent() failed: %v", err)
	}

	agent2, err := manager.AddAgent("agent2", "test-adapter-2", "/tmp/repo2", nil)
	if err != nil {
		t.Fatalf("AddAgent() failed: %v", err)
	}

	// Test getting agents
	retrievedAgent, err := manager.GetAgent(agent1.ID)
	if err != nil {
		t.Errorf("GetAgent() failed: %v", err)
	}
	if retrievedAgent.ID != agent1.ID {
		t.Errorf("GetAgent() returned wrong agent")
	}

	// Test getting non-existent agent
	_, err = manager.GetAgent(agent2.ID)
	if err != nil {
		t.Errorf("GetAgent() failed: %v", err)
	}

	// Test listing agents
	agents := manager.ListAgents()
	if len(agents) != 2 {
		t.Errorf("ListAgents() returned %d agents, want 2", len(agents))
	}

	// Test starting agent
	err = manager.StartAgent(agent1.ID)
	if err != nil {
		t.Errorf("StartAgent() failed: %v", err)
	}

	// Verify state changed
	updatedAgent, _ := manager.GetAgent(agent1.ID)
	if updatedAgent.State != api.AgentStateStarting {
		t.Errorf("StartAgent() state = %v, want %v", updatedAgent.State, api.AgentStateStarting)
	}

	// Test updating presence
	err = manager.UpdatePresence(agent1.ID, EvtActivityDetected)
	if err != nil {
		t.Errorf("UpdatePresence() failed: %v", err)
	}

	// Verify presence changed
	updatedAgent, _ = manager.GetAgent(agent1.ID)
	if updatedAgent.Presence != api.PresenceOnline {
		t.Errorf("UpdatePresence() presence = %v, want %v", updatedAgent.Presence, api.PresenceOnline)
	}
}

func TestEventHandlers(t *testing.T) {
	actor, err := NewAgentActor("test-agent", "test-adapter", "/tmp/repo", nil)
	if err != nil {
		t.Fatalf("NewAgentActor() failed: %v", err)
	}

	// Register event handlers
	var lifecycleEvents []string
	var presenceEvents []string

	actor.OnEvent("lifecycle", func(agent *api.Agent, data interface{}) {
		eventData := data.(map[string]interface{})
		event := eventData["event"].(LifecycleEvent)
		lifecycleEvents = append(lifecycleEvents, string(event))
	})

	actor.OnEvent("presence", func(agent *api.Agent, data interface{}) {
		eventData := data.(map[string]interface{})
		event := eventData["event"].(PresenceEvent)
		presenceEvents = append(presenceEvents, string(event))
	})

	// Trigger events
	actor.SendLifecycleEvent(EventStart)
	actor.SendPresenceEvent(EventGoOnline)

	// Verify handlers were called
	if len(lifecycleEvents) != 1 || lifecycleEvents[0] != string(EventStart) {
		t.Errorf("Lifecycle event handler not called correctly")
	}

	if len(presenceEvents) != 1 || presenceEvents[0] != string(EventGoOnline) {
		t.Errorf("Presence event handler not called correctly")
	}
}

func TestInvalidTransitions(t *testing.T) {
	actor, err := NewAgentActor("test-agent", "test-adapter", "/tmp/repo", nil)
	if err != nil {
		t.Fatalf("NewAgentActor() failed: %v", err)
	}

	// Test invalid lifecycle transition (pending -> startup_complete without start)
	err = actor.SendLifecycleEvent(EventStartupComplete)
	if err == nil {
		t.Error("SendLifecycleEvent() should have failed for invalid transition")
	}

	// Test invalid presence transition (offline -> away without going online first)
	err = actor.SendPresenceEvent(EventGoAway)
	if err == nil {
		t.Error("SendPresenceEvent() should have failed for invalid transition")
	}
}

func TestTimestampUpdates(t *testing.T) {
	actor, err := NewAgentActor("test-agent", "test-adapter", "/tmp/repo", nil)
	if err != nil {
		t.Fatalf("NewAgentActor() failed: %v", err)
	}

	originalTime := actor.Agent.UpdatedAt
	time.Sleep(10 * time.Millisecond) // Ensure time difference

	// Trigger state change
	err = actor.SendLifecycleEvent(EventStart)
	if err != nil {
		t.Fatalf("SendLifecycleEvent() failed: %v", err)
	}

	newTime := actor.Agent.UpdatedAt
	if !newTime.After(originalTime) {
		t.Error("UpdatedAt timestamp was not updated after state change")
	}
}
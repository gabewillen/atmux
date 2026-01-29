package api

import (
	"testing"
	"time"
)

func TestNewAgent(t *testing.T) {
	repoRoot := RepoRoot("/path/to/repo")
	workDir := "/path/to/repo/.amux/worktrees/test-agent"
	command := []string{"claude-code"}

	agent := NewAgent("Test Agent", "A test agent", "claude-code", repoRoot, workDir, command)

	if !IDIsValid(agent.ID) {
		t.Error("NewAgent() should set a valid ID")
	}

	if agent.Name != "Test Agent" {
		t.Errorf("NewAgent() Name = %q, want %q", agent.Name, "Test Agent")
	}

	if agent.Description != "A test agent" {
		t.Errorf("NewAgent() Description = %q, want %q", agent.Description, "A test agent")
	}

	if agent.Adapter != "claude-code" {
		t.Errorf("NewAgent() Adapter = %q, want %q", agent.Adapter, "claude-code")
	}

	if agent.AgentSlug != "test-agent" {
		t.Errorf("NewAgent() AgentSlug = %q, want %q", agent.AgentSlug, "test-agent")
	}

	if agent.RepoRoot != repoRoot {
		t.Errorf("NewAgent() RepoRoot = %q, want %q", agent.RepoRoot, repoRoot)
	}

	if agent.WorkDir != workDir {
		t.Errorf("NewAgent() WorkDir = %q, want %q", agent.WorkDir, workDir)
	}

	if len(agent.Command) != 1 || agent.Command[0] != "claude-code" {
		t.Errorf("NewAgent() Command = %v, want %v", agent.Command, command)
	}

	if agent.CreatedAt.IsZero() {
		t.Error("NewAgent() should set CreatedAt")
	}

	if agent.UpdatedAt.IsZero() {
		t.Error("NewAgent() should set UpdatedAt")
	}
}

func TestAgentIsValid(t *testing.T) {
	tests := []struct {
		name  string
		agent *Agent
		valid bool
	}{
		{
			name:  "valid agent",
			agent: NewAgent("Test", "desc", "adapter", RepoRoot("/repo"), "/work", []string{"cmd"}),
			valid: true,
		},
		{
			name:  "nil agent",
			agent: nil,
			valid: false,
		},
		{
			name: "invalid ID",
			agent: &Agent{
				ID:        ID(0),
				Name:      "Test",
				Adapter:   "adapter",
				AgentSlug: "test",
				RepoRoot:  "/repo",
				Command:   []string{"cmd"},
			},
			valid: false,
		},
		{
			name: "empty name",
			agent: &Agent{
				ID:        NewID(),
				Name:      "",
				Adapter:   "adapter",
				AgentSlug: "test",
				RepoRoot:  "/repo",
				Command:   []string{"cmd"},
			},
			valid: false,
		},
		{
			name: "empty adapter",
			agent: &Agent{
				ID:        NewID(),
				Name:      "Test",
				Adapter:   "",
				AgentSlug: "test",
				RepoRoot:  "/repo",
				Command:   []string{"cmd"},
			},
			valid: false,
		},
		{
			name: "empty command",
			agent: &Agent{
				ID:        NewID(),
				Name:      "Test",
				Adapter:   "adapter",
				AgentSlug: "test",
				RepoRoot:  "/repo",
				Command:   []string{},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.agent.IsValid() != tt.valid {
				t.Errorf("Agent.IsValid() = %v, want %v", tt.agent.IsValid(), tt.valid)
			}
		})
	}
}

func TestNewSession(t *testing.T) {
	agentID := NewID()
	session := NewSession(agentID)

	if !IDIsValid(session.ID) {
		t.Error("NewSession() should set a valid ID")
	}

	if session.AgentID != agentID {
		t.Errorf("NewSession() AgentID = %v, want %v", session.AgentID, agentID)
	}

	if session.State != SessionStatePending {
		t.Errorf("NewSession() State = %q, want %q", session.State, SessionStatePending)
	}
}

func TestSessionIsRunning(t *testing.T) {
	tests := []struct {
		name    string
		state   SessionState
		running bool
	}{
		{"pending", SessionStatePending, false},
		{"starting", SessionStateStarting, false},
		{"running", SessionStateRunning, true},
		{"terminated", SessionStateTerminated, false},
		{"errored", SessionStateErrored, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &Session{State: tt.state}
			if session.IsRunning() != tt.running {
				t.Errorf("Session.IsRunning() with state %q = %v, want %v", tt.state, session.IsRunning(), tt.running)
			}
		})
	}
}

func TestSessionIsFinished(t *testing.T) {
	tests := []struct {
		name     string
		state    SessionState
		finished bool
	}{
		{"pending", SessionStatePending, false},
		{"starting", SessionStateStarting, false},
		{"running", SessionStateRunning, false},
		{"terminated", SessionStateTerminated, true},
		{"errored", SessionStateErrored, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &Session{State: tt.state}
			if session.IsFinished() != tt.finished {
				t.Errorf("Session.IsFinished() with state %q = %v, want %v", tt.state, session.IsFinished(), tt.finished)
			}
		})
	}
}

func TestNewPresenceInfo(t *testing.T) {
	agentID := NewID()
	presence := PresenceOnline

	info := NewPresenceInfo(agentID, presence)

	if info.AgentID != agentID {
		t.Errorf("NewPresenceInfo() AgentID = %v, want %v", info.AgentID, agentID)
	}

	if info.Presence != presence {
		t.Errorf("NewPresenceInfo() Presence = %q, want %q", info.Presence, presence)
	}

	if info.LastActivity.IsZero() {
		t.Error("NewPresenceInfo() should set LastActivity")
	}

	if info.UpdatedAt.IsZero() {
		t.Error("NewPresenceInfo() should set UpdatedAt")
	}
}

func TestUpdatePresence(t *testing.T) {
	agentID := NewID()
	info := NewPresenceInfo(agentID, PresenceOnline)

	originalTime := info.UpdatedAt
	// Wait a bit to ensure different timestamp
	time.Sleep(1 * time.Millisecond)

	info.UpdatePresence(PresenceBusy)

	if info.Presence != PresenceBusy {
		t.Errorf("UpdatePresence() Presence = %q, want %q", info.Presence, PresenceBusy)
	}

	if !info.UpdatedAt.After(originalTime) {
		t.Error("UpdatePresence() should update UpdatedAt timestamp")
	}

	if !info.LastActivity.Equal(info.UpdatedAt) {
		t.Error("UpdatePresence() should update LastActivity to match UpdatedAt")
	}
}

func TestUpdateStatus(t *testing.T) {
	agentID := NewID()
	info := NewPresenceInfo(agentID, PresenceOnline)

	originalTime := info.UpdatedAt
	// Wait a bit to ensure different timestamp
	time.Sleep(1 * time.Millisecond)

	status := "Working on task"
	info.UpdateStatus(status)

	if info.Status != status {
		t.Errorf("UpdateStatus() Status = %q, want %q", info.Status, status)
	}

	if !info.UpdatedAt.After(originalTime) {
		t.Error("UpdateStatus() should update UpdatedAt timestamp")
	}

	if !info.LastActivity.Equal(info.UpdatedAt) {
		t.Error("UpdateStatus() should update LastActivity to match UpdatedAt")
	}
}

package api

import (
	"testing"
)

func TestLocationTypeString(t *testing.T) {
	tests := []struct {
		name     string
		locType  LocationType
		expected string
	}{
		{
			name:     "local",
			locType:  LocationLocal,
			expected: "local",
		},
		{
			name:     "ssh",
			locType:  LocationSSH,
			expected: "ssh",
		},
		{
			name:     "unknown",
			locType:  LocationType(999),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.locType.String()
			if result != tt.expected {
				t.Errorf("LocationType.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseLocationType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected LocationType
		wantErr  bool
	}{
		{
			name:     "local lowercase",
			input:    "local",
			expected: LocationLocal,
			wantErr:  false,
		},
		{
			name:     "local uppercase",
			input:    "LOCAL",
			expected: LocationLocal,
			wantErr:  false,
		},
		{
			name:     "local mixed case",
			input:    "Local",
			expected: LocationLocal,
			wantErr:  false,
		},
		{
			name:     "ssh lowercase",
			input:    "ssh",
			expected: LocationSSH,
			wantErr:  false,
		},
		{
			name:     "ssh uppercase",
			input:    "SSH",
			expected: LocationSSH,
			wantErr:  false,
		},
		{
			name:     "ssh mixed case",
			input:    "Ssh",
			expected: LocationSSH,
			wantErr:  false,
		},
		{
			name:     "invalid type",
			input:    "invalid",
			expected: LocationLocal,
			wantErr:  true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: LocationLocal,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseLocationType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLocationType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("ParseLocationType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAgentStructure(t *testing.T) {
	// Verify Agent structure can be instantiated with required fields
	agent := Agent{
		ID:       GenerateID(),
		Name:     "test-agent",
		About:    "A test agent",
		Adapter:  "test-adapter",
		RepoRoot: "/path/to/repo",
		Worktree: "/path/to/repo/.amux/worktrees/test-agent",
		Location: Location{
			Type: LocationLocal,
		},
	}

	if agent.ID == BroadcastID {
		t.Error("Agent ID should not be BroadcastID")
	}
	if agent.Name != "test-agent" {
		t.Errorf("Agent.Name = %q, want %q", agent.Name, "test-agent")
	}
	if agent.Adapter != "test-adapter" {
		t.Errorf("Agent.Adapter = %q, want %q", agent.Adapter, "test-adapter")
	}
}

func TestSessionStructure(t *testing.T) {
	// Verify Session structure can be instantiated
	session := Session{
		ID:     GenerateID(),
		Agents: []*Agent{},
	}

	if session.ID == BroadcastID {
		t.Error("Session ID should not be BroadcastID")
	}
	if session.Agents == nil {
		t.Error("Session.Agents should not be nil")
	}
}

func TestAgentMessageStructure(t *testing.T) {
	// Verify AgentMessage structure can be instantiated
	msg := AgentMessage{
		ID:      GenerateID(),
		From:    GenerateID(),
		To:      GenerateID(),
		ToSlug:  "target-agent",
		Content: "Hello, world!",
	}

	if msg.ID == BroadcastID {
		t.Error("AgentMessage ID should not be BroadcastID")
	}
	if msg.Content != "Hello, world!" {
		t.Errorf("AgentMessage.Content = %q, want %q", msg.Content, "Hello, world!")
	}
}

func TestBroadcastMessage(t *testing.T) {
	// Verify broadcast message with BroadcastID
	msg := AgentMessage{
		ID:      GenerateID(),
		From:    GenerateID(),
		To:      BroadcastID,
		ToSlug:  "all",
		Content: "Broadcast message",
	}

	if msg.To != BroadcastID {
		t.Errorf("Broadcast message To = %v, want %v", msg.To, BroadcastID)
	}
	if msg.To != 0 {
		t.Error("BroadcastID should be 0")
	}
}

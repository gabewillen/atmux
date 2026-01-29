// Package messaging_test provides tests for the messaging implementation.
package messaging

import (
	"testing"

	"github.com/stateforward/hsm-go/muid"

	"github.com/copilot-claude-sonnet-4/amux/pkg/api"
	"github.com/copilot-claude-sonnet-4/amux/internal/roster"
)

func TestRouter_ProcessOutboundMessage(t *testing.T) {
	// Setup roster with test participants
	rosterStore := roster.NewStore()
	defer rosterStore.Close()

	directorID := muid.Make()
	managerID := muid.Make()
	agentID := muid.Make()

	err := rosterStore.AddDirector(directorID, api.PresenceOnline)
	if err != nil {
		t.Fatalf("Failed to add director: %v", err)
	}

	err = rosterStore.AddManager(managerID, "localhost", api.PresenceOnline)
	if err != nil {
		t.Fatalf("Failed to add manager: %v", err)
	}

	agent := &api.Agent{
		ID:       agentID,
		Slug:     "backend-dev",
		Name:     "Backend Developer",
		Presence: api.PresenceOnline,
		State:    api.AgentStateRunning,
	}
	err = rosterStore.AddAgent(agent, "localhost")
	if err != nil {
		t.Fatalf("Failed to add agent: %v", err)
	}

	// Create router
	router := NewRouter(rosterStore, "localhost", directorID, managerID)
	defer router.Close()

	tests := []struct {
		name        string
		fromID      muid.MUID
		toSlug      string
		content     string
		expectedTo  muid.MUID
		expectError bool
	}{
		{
			name:       "message to agent by slug",
			fromID:     agentID,
			toSlug:     "backend-dev",
			content:    "Hello!",
			expectedTo: agentID,
		},
		{
			name:       "message to director",
			fromID:     agentID,
			toSlug:     "director",
			content:    "Status update",
			expectedTo: directorID,
		},
		{
			name:       "message to local manager",
			fromID:     agentID,
			toSlug:     "manager",
			content:    "Help needed",
			expectedTo: managerID,
		},
		{
			name:       "broadcast message",
			fromID:     agentID,
			toSlug:     "all",
			content:    "Hello everyone!",
			expectedTo: BroadcastID,
		},
		{
			name:        "message to unknown recipient",
			fromID:      agentID,
			toSlug:      "unknown-agent",
			content:     "Test",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := router.ProcessOutboundMessage(tt.fromID, tt.toSlug, tt.content)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if msg.From != tt.fromID {
				t.Errorf("Expected From %s, got %s", tt.fromID, msg.From)
			}

			if msg.To != tt.expectedTo {
				t.Errorf("Expected To %s, got %s", tt.expectedTo, msg.To)
			}

			if msg.ToSlug != tt.toSlug {
				t.Errorf("Expected ToSlug %s, got %s", tt.toSlug, msg.ToSlug)
			}

			if msg.Content != tt.content {
				t.Errorf("Expected Content %s, got %s", tt.content, msg.Content)
			}

			if msg.ID == 0 {
				t.Error("Expected non-zero message ID")
			}
		})
	}
}

func TestRouter_GetDeliveryChannels(t *testing.T) {
	// Setup roster with test participants
	rosterStore := roster.NewStore()
	defer rosterStore.Close()

	directorID := muid.Make()
	managerID := muid.Make()
	agentID := muid.Make()

	err := rosterStore.AddDirector(directorID, api.PresenceOnline)
	if err != nil {
		t.Fatalf("Failed to add director: %v", err)
	}

	err = rosterStore.AddManager(managerID, "localhost", api.PresenceOnline)
	if err != nil {
		t.Fatalf("Failed to add manager: %v", err)
	}

	agent := &api.Agent{
		ID:       agentID,
		Slug:     "test-agent",
		Name:     "Test Agent",
		Presence: api.PresenceOnline,
		State:    api.AgentStateRunning,
	}
	err = rosterStore.AddAgent(agent, "localhost")
	if err != nil {
		t.Fatalf("Failed to add agent: %v", err)
	}

	// Create router
	router := NewRouter(rosterStore, "localhost", directorID, managerID)
	defer router.Close()

	tests := []struct {
		name             string
		msg              *AgentMessage
		expectedChannels []string
	}{
		{
			name: "broadcast message",
			msg: &AgentMessage{
				From: agentID,
				To:   BroadcastID,
			},
			expectedChannels: []string{"P.comm.broadcast"},
		},
		{
			name: "agent to director",
			msg: &AgentMessage{
				From: agentID,
				To:   directorID,
			},
			expectedChannels: []string{
				"P.comm.director",
				"P.comm.agent.localhost.test-agent",
			},
		},
		{
			name: "agent to manager",
			msg: &AgentMessage{
				From: agentID,
				To:   managerID,
			},
			expectedChannels: []string{
				"P.comm.manager.localhost",
				"P.comm.agent.localhost.test-agent",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channels, err := router.GetDeliveryChannels(tt.msg)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(channels) != len(tt.expectedChannels) {
				t.Errorf("Expected %d channels, got %d", len(tt.expectedChannels), len(channels))
			}

			// Convert to map for easier comparison
			channelMap := make(map[string]bool)
			for _, ch := range channels {
				channelMap[ch] = true
			}

			for _, expectedCh := range tt.expectedChannels {
				if !channelMap[expectedCh] {
					t.Errorf("Expected channel %s not found in result", expectedCh)
				}
			}
		})
	}
}

func TestMessageDetector(t *testing.T) {
	detector := NewMessageDetector()

	// Load test patterns
	patterns := &AdapterPatterns{
		Message: `@([a-zA-Z0-9-]+):\s*(.+)`, // Allow hyphens in slugs
	}

	err := detector.LoadPatterns("claude-code", patterns)
	if err != nil {
		t.Fatalf("Failed to load patterns: %v", err)
	}

	tests := []struct {
		name           string
		adapter        string
		output         string
		expectedToSlug string
		expectedContent string
		expectedDetected bool
	}{
		{
			name:             "claude code message",
			adapter:          "claude-code",
			output:           "@backend-dev: can you review this?",
			expectedToSlug:   "backend-dev",
			expectedContent:  "can you review this?",
			expectedDetected: true,
		},
		{
			name:             "claude code broadcast",
			adapter:          "claude-code", 
			output:           "@all: task completed successfully",
			expectedToSlug:   "all",
			expectedContent:  "task completed successfully",
			expectedDetected: true,
		},
		{
			name:             "no message pattern",
			adapter:          "claude-code",
			output:           "regular output without message",
			expectedDetected: false,
		},
		{
			name:             "unknown adapter",
			adapter:          "unknown-adapter",
			output:           "@test: message",
			expectedDetected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toSlug, content, detected := detector.DetectMessage(tt.adapter, tt.output)

			if detected != tt.expectedDetected {
				t.Errorf("Expected detected %v, got %v", tt.expectedDetected, detected)
			}

			if !detected {
				return // Skip content checks if not detected
			}

			if toSlug != tt.expectedToSlug {
				t.Errorf("Expected ToSlug %s, got %s", tt.expectedToSlug, toSlug)
			}

			if content != tt.expectedContent {
				t.Errorf("Expected Content %s, got %s", tt.expectedContent, content)
			}
		})
	}
}
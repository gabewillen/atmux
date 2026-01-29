package remote

import (
	"encoding/json"
	"testing"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestNewSpawnRequest(t *testing.T) {
	agentID := api.AgentID(12345) // Assuming uint64 cast works or valid
	// Wait, AgentID is muid.MUID (uint64). We need to create a valid one.
	// Just casting is fine for struct test.
	slug := api.AgentSlug("test-agent")
	
	req := NewSpawnRequest(agentID, slug, "~/repo", []string{"ls"}, nil)
	
	if req.Type != "spawn" {
		t.Errorf("Expected type spawn, got %s", req.Type)
	}
	
	var p protocol.SpawnPayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}
	
	if p.Slug != slug {
		t.Errorf("Slug mismatch")
	}
	if p.RepoPath != "~/repo" {
		t.Errorf("RepoPath mismatch")
	}
}

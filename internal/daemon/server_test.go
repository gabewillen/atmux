package daemon

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/config"
)

func TestServer_JSONRPC(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "amuxd.sock")
	
	cfg := config.DaemonConfig{SocketPath: socketPath}
	server := NewServer(cfg)
	// Use isolated registry
	server.Registry = agent.NewRegistry()
	agent.GlobalRegistry = server.Registry // Swap global for safety in tests?
	// NewAgent uses GlobalRegistry.
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	if err := server.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	
	// Connect
	time.Sleep(10 * time.Millisecond)
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()
	
	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)
	
	// 1. Add Agent
	// Need a valid repo path.
	repoDir := filepath.Join(tmpDir, "repo")
	os.MkdirAll(repoDir, 0755)
	
	addReq := jsonRPCRequest{
		JSONRPC: "2.0",
		Method:  "agent.add",
		Params:  json.RawMessage(`{"name":"rpc-agent", "adapter":"test", "repo_path":"` + repoDir + `"}`),
		ID:      1,
	}
	enc.Encode(addReq)
	
	var addResp jsonRPCResponse
	if err := dec.Decode(&addResp); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if addResp.Error != nil {
		t.Errorf("agent.add error: %s", addResp.Error.Message)
	}
	
	// 2. List Agents
	listReq := jsonRPCRequest{
		JSONRPC: "2.0",
		Method:  "agent.list",
		ID:      2,
	}
	enc.Encode(listReq)
	
	var listResp jsonRPCResponse
	dec.Decode(&listResp) // Ignore error
	
	if listResp.Error != nil {
		t.Errorf("agent.list error: %s", listResp.Error.Message)
	}
	
	// Verify result structure (basic check)
	// We expect a list of RosterEntry
	// Since Result is interface{}, we marshal/unmarshal to check
	bytes, _ := json.Marshal(listResp.Result)
	var roster []agent.RosterEntry
	json.Unmarshal(bytes, &roster)
	
	if len(roster) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(roster))
	} else if roster[0].Name != "rpc-agent" {
		t.Errorf("Expected name 'rpc-agent', got %s", roster[0].Name)
	}
}
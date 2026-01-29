package daemon

import (
	"context"
	"encoding/json"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
)

func TestDaemonServer(t *testing.T) {
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "amuxd.sock")
	
	cfg := config.DaemonConfig{
		SocketPath: socketPath,
	}
	
	srv := NewServer(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	if err := srv.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer srv.Stop()
	
	// Wait for socket
	time.Sleep(100 * time.Millisecond)
	
	// Connect
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()
	
	// Send Ping
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		Method:  "ping",
		ID:      1,
	}
	
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	
	// Read Response
	var resp jsonRPCResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	
	if resp.Result != "pong" {
		t.Errorf("Expected pong, got %v", resp.Result)
	}
}

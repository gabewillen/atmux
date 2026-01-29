package process

import (
	"context"
	"encoding/json"
	"net"
	"path/filepath"
	"testing"
	"time"
)

func TestStartMCPServer(t *testing.T) {
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "mcp.sock")
	
	server := NewMCPServer(socketPath)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	if err := server.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	
	// Connect client
	// Give server a moment
	time.Sleep(10 * time.Millisecond)
	
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()
	
	// Broadcast
	go func() {
		time.Sleep(50 * time.Millisecond)
		server.Broadcast("test.method", map[string]string{"foo": "bar"})
	}()
	
	// Read
	dec := json.NewDecoder(conn)
	var notif MCPNotification
	if err := dec.Decode(&notif); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	
	if notif.Method != "test.method" {
		t.Errorf("Expected method test.method, got %s", notif.Method)
	}
}
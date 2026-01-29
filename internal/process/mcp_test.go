package process

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
)

func TestStartMCPServer(t *testing.T) {
	tmpDir := t.TempDir()
	
	cfg := config.ProcessConfig{
		HookSocketDir: tmpDir,
	}
	tracker := NewTracker()
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	if err := StartMCPServer(ctx, cfg, tracker); err != nil {
		t.Fatalf("StartMCPServer failed: %v", err)
	}
	
	socketPath := filepath.Join(tmpDir, "amux-mcp.sock")
	
	// Wait for socket
	time.Sleep(100 * time.Millisecond)
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		t.Fatal("Socket file not created")
	}
	
	// Try connect
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	conn.Close()
}

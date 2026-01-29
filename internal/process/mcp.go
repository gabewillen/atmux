package process

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/agentflare-ai/amux/internal/config"
)

// StartMCPServer starts the Notification MCP server.
func StartMCPServer(ctx context.Context, cfg config.ProcessConfig, tracker *Tracker) error {
	socketPath := filepath.Join(cfg.HookSocketDir, "amux-mcp.sock") // Or configured path
	
	// Ensure dir exists
	if err := os.MkdirAll(filepath.Dir(socketPath), 0755); err != nil {
		return fmt.Errorf("failed to create mcp socket dir: %w", err)
	}

	// Remove old socket
	os.Remove(socketPath)

	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("mcp listen failed: %w", err)
	}

	server := &MCPServer{
		Tracker: tracker,
		Clients: make(map[*MCPSession]struct{}),
	}

	go server.Run(ctx, ln)
	
	return nil
}

type MCPServer struct {
	Tracker *Tracker
	mu      sync.Mutex
	Clients map[*MCPSession]struct{}
}

func (s *MCPServer) Run(ctx context.Context, ln net.Listener) {
	defer ln.Close()
	
	// Accept loop
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			session := &MCPSession{
				conn: conn,
				srv:  s,
			}
			s.addClient(session)
			go session.Serve(ctx)
		}
	}()
	
	<-ctx.Done()
}

func (s *MCPServer) addClient(c *MCPSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Clients[c] = struct{}{}
}

func (s *MCPServer) removeClient(c *MCPSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Clients, c)
}

type MCPSession struct {
	conn net.Conn
	srv  *MCPServer
}

func (s *MCPSession) Serve(ctx context.Context) {
	defer s.srv.removeClient(s)
	defer s.conn.Close()
	
	decoder := json.NewDecoder(s.conn)
	
	// Read loop (basic JSON-RPC 2.0 handling)
	for {
		var msg json.RawMessage
		if err := decoder.Decode(&msg); err != nil {
			return
		}
		// Handle request (placeholder)
	}
}

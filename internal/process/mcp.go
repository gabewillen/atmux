package process

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"sync"
)

// MCPNotification represents a notification sent to clients.
type MCPNotification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

// MCPServer handles notification subscriptions via a Unix socket.
type MCPServer struct {
	SocketPath string
	mu         sync.Mutex
	clients    map[net.Conn]struct{}
	listener   net.Listener
}

// NewMCPServer creates a new MCP server.
func NewMCPServer(socketPath string) *MCPServer {
	return &MCPServer{
		SocketPath: socketPath,
		clients:    make(map[net.Conn]struct{}),
	}
}

// Start starts the server.
func (s *MCPServer) Start(ctx context.Context) error {
	if err := os.MkdirAll(filepath.Dir(s.SocketPath), 0755); err != nil {
		return err
	}
	os.Remove(s.SocketPath)

	ln, err := net.Listen("unix", s.SocketPath)
	if err != nil {
		return err
	}
	s.listener = ln

	go s.acceptLoop(ctx)
	return nil
}

func (s *MCPServer) acceptLoop(ctx context.Context) {
	defer s.listener.Close()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				continue
			}
		}
		
		s.addClient(conn)
		go s.handleClient(ctx, conn)
	}
}

func (s *MCPServer) handleClient(ctx context.Context, conn net.Conn) {
	defer s.removeClient(conn)
	defer conn.Close()
	
	// Keep connection open until context done or read error
	// Clients listen for notifications.
	// We might also support basic requests (like Subscribe), 
	// but the plan emphasizes Server-to-Client notifications.
	
buf := make([]byte, 1024)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			return
		}
		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func (s *MCPServer) addClient(c net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[c] = struct{}{}
}

func (s *MCPServer) removeClient(c net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, c)
}

// Broadcast sends a notification to all connected clients.
func (s *MCPServer) Broadcast(method string, params interface{}) {
	notif := MCPNotification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	data, _ := json.Marshal(notif)
	data = append(data, '\n') // Newline delimited

	s.mu.Lock()
	defer s.mu.Unlock()

	for conn := range s.clients {
		go func(c net.Conn) {
			c.Write(data)
		}(conn)
	}
}
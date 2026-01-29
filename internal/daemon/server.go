package daemon

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

// Server is the JSON-RPC daemon server.
type Server struct {
	Config  config.DaemonConfig
	Listener net.Listener
	mu      sync.Mutex
	clients map[*ClientConn]struct{}
}

// ClientConn represents a connected client.
type ClientConn struct {
	conn net.Conn
	srv  *Server
}

// NewServer creates a new daemon server.
func NewServer(cfg config.DaemonConfig) *Server {
	return &Server{
		Config:  cfg,
		clients: make(map[*ClientConn]struct{}),
	}
}

// Start starts the server.
func (s *Server) Start(ctx context.Context) error {
	socketPath := s.Config.SocketPath
	// Expand home if needed (using internal/paths helper in real implementation)
	// assuming it's expanded or absolute for now.
	
	if err := os.MkdirAll(filepath.Dir(socketPath), 0755); err != nil {
		return fmt.Errorf("failed to create socket dir: %w", err)
	}
	
	os.Remove(socketPath) // Clean up old socket

	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", socketPath, err)
	}
	s.Listener = ln

	go s.serve(ctx)
	return nil
}

func (s *Server) serve(ctx context.Context) {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				// Log error?
				continue
			}
		}
		
		client := &ClientConn{
			conn: conn,
			srv:  s,
		}
		s.addClient(client)
		go client.handle(ctx)
	}
}

func (s *Server) Stop() error {
	if s.Listener != nil {
		return s.Listener.Close()
	}
	return nil
}

func (s *Server) addClient(c *ClientConn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[c] = struct{}{}
}

func (s *Server) removeClient(c *ClientConn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, c)
}

func (c *ClientConn) handle(ctx context.Context) {
	defer c.srv.removeClient(c)
	defer c.conn.Close()
	
	dec := json.NewDecoder(c.conn)
	for {
		var req jsonRPCRequest
		if err := dec.Decode(&req); err != nil {
			return
		}
		
		// Process Request
		resp := c.processRequest(req)
		
		// Send Response
		enc := json.NewEncoder(c.conn)
		if err := enc.Encode(resp); err != nil {
			return
		}
	}
}

func (c *ClientConn) processRequest(req jsonRPCRequest) jsonRPCResponse {
	// Router logic here
	res := jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}
	
	switch req.Method {
	case "ping":
		res.Result = "pong"
	case "version":
		res.Result = "v0.0.1"
	default:
		res.Error = &jsonRPCError{Code: -32601, Message: "Method not found"}
	}
	
	return res
}

// Minimal JSON-RPC types
type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

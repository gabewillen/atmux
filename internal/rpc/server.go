// Package rpc implements the JSON-RPC 2.0 control plane for amux.
package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/stateforward/hsm-go/muid"

	"github.com/copilot-claude-sonnet-4/amux/internal/agent"
	"github.com/copilot-claude-sonnet-4/amux/internal/paths"
)

// Common sentinel errors for RPC operations.
var (
	ErrMethodNotFound = errors.New("method not found")
	ErrInvalidParams  = errors.New("invalid parameters")
	ErrInternalError  = errors.New("internal error")
)

// JSON-RPC 2.0 error codes per specification.
const (
	ErrorCodeParseError     = -32700
	ErrorCodeInvalidRequest = -32600
	ErrorCodeMethodNotFound = -32601
	ErrorCodeInvalidParams  = -32602
	ErrorCodeInternalError  = -32603
)

// Request represents a JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      interface{} `json:"id,omitempty"`
}

// Response represents a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *ErrorObj   `json:"error,omitempty"`
	ID      interface{} `json:"id,omitempty"`
}

// ErrorObj represents a JSON-RPC 2.0 error object.
type ErrorObj struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error implements the error interface for ErrorObj.
func (e *ErrorObj) Error() string {
	return e.Message
}

// Server implements the JSON-RPC 2.0 server for amux daemon communication.
type Server struct {
	listener net.Listener
	manager  *agent.Manager
	resolver *paths.Resolver
	mu       sync.RWMutex
	handlers map[string]func(context.Context, json.RawMessage) (interface{}, error)
}

// NewServer creates a new JSON-RPC server with the given agent manager and resolver.
func NewServer(manager *agent.Manager, resolver *paths.Resolver) *Server {
	s := &Server{
		manager:  manager,
		resolver: resolver,
		handlers: make(map[string]func(context.Context, json.RawMessage) (interface{}, error)),
	}

	// Register standard methods
	s.handlers["agent.add"] = s.agentAdd
	s.handlers["agent.list"] = s.agentList
	s.handlers["agent.start"] = s.agentStart
	s.handlers["agent.stop"] = s.agentStop

	return s
}

// Listen starts the server listening on the configured socket path.
func (s *Server) Listen() error {
	socketPath := s.resolver.SocketPath()

	// Remove existing socket file
	if err := os.RemoveAll(socketPath); err != nil {
		return fmt.Errorf("failed to remove existing socket: %w", err)
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(socketPath), 0755); err != nil {
		return fmt.Errorf("failed to create socket directory: %w", err)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on socket: %w", err)
	}

	s.listener = listener
	return nil
}

// Serve starts accepting connections and handling requests.
func (s *Server) Serve() error {
	if s.listener == nil {
		return fmt.Errorf("server not listening")
	}

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return fmt.Errorf("accept error: %w", err)
		}

		go s.handleConnection(conn)
	}
}

// Close stops the server and closes the listener.
func (s *Server) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// handleConnection processes a single client connection.
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var req Request
		if err := decoder.Decode(&req); err != nil {
			return
		}

		resp := s.handleRequest(context.Background(), &req)
		if err := encoder.Encode(resp); err != nil {
			return
		}
	}
}

// handleRequest processes a single JSON-RPC request.
func (s *Server) handleRequest(ctx context.Context, req *Request) *Response {
	resp := &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	// Validate JSON-RPC version
	if req.JSONRPC != "2.0" {
		resp.Error = &ErrorObj{
			Code:    ErrorCodeInvalidRequest,
			Message: "invalid jsonrpc version",
		}
		return resp
	}

	// Find handler
	s.mu.RLock()
	handler, exists := s.handlers[req.Method]
	s.mu.RUnlock()

	if !exists {
		resp.Error = &ErrorObj{
			Code:    ErrorCodeMethodNotFound,
			Message: fmt.Sprintf("method not found: %s", req.Method),
		}
		return resp
	}

	// Extract parameters as raw JSON
	var params json.RawMessage
	if req.Params != nil {
		var err error
		params, err = json.Marshal(req.Params)
		if err != nil {
			resp.Error = &ErrorObj{
				Code:    ErrorCodeInvalidParams,
				Message: "invalid parameters",
			}
			return resp
		}
	}

	// Call handler
	result, err := handler(ctx, params)
	if err != nil {
		resp.Error = &ErrorObj{
			Code:    ErrorCodeInternalError,
			Message: err.Error(),
		}
		return resp
	}

	resp.Result = result
	return resp
}

// AgentAddParams represents parameters for agent.add method.
type AgentAddParams struct {
	Name     string                 `json:"name"`
	Adapter  string                 `json:"adapter"`
	RepoRoot string                 `json:"repo_root,omitempty"`
	Config   map[string]interface{} `json:"config,omitempty"`
}

// AgentAddResult represents the result of agent.add method.
type AgentAddResult struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Adapter string `json:"adapter"`
	Status  string `json:"status"`
}

// agentAdd handles the agent.add RPC method.
func (s *Server) agentAdd(ctx context.Context, data json.RawMessage) (interface{}, error) {
	var p AgentAddParams
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Validate required fields
	if p.Name == "" || p.Adapter == "" {
		return nil, fmt.Errorf("name and adapter are required")
	}

	if p.RepoRoot == "" {
		// Use current working directory as default
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		p.RepoRoot = cwd
	}

	// TODO: Add agent via manager once interface is stable
	return &AgentAddResult{
		ID:      "1",
		Name:    p.Name,
		Adapter: p.Adapter,
		Status:  "pending",
	}, nil
}

// AgentListResult represents the result of agent.list method.
type AgentListResult struct {
	Agents []AgentInfo `json:"agents"`
}

// AgentInfo represents agent information in listings.
type AgentInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Adapter string `json:"adapter"`
	Status  string `json:"status"`
}

// agentList handles the agent.list RPC method.
func (s *Server) agentList(ctx context.Context, data json.RawMessage) (interface{}, error) {
	// TODO: implement proper agent listing once manager interface is stable
	return &AgentListResult{
		Agents: []AgentInfo{},
	}, nil
}

// AgentStartParams represents parameters for agent.start method.
type AgentStartParams struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// AgentStartResult represents the result of agent.start method.
type AgentStartResult struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// agentStart handles the agent.start RPC method.
func (s *Server) agentStart(ctx context.Context, data json.RawMessage) (interface{}, error) {
	var p AgentStartParams
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Parse agent ID
	var agentID muid.MUID
	if p.ID != "" {
		// Convert string to MUID (assume it's a base-10 representation)
		if id, err := strconv.ParseUint(p.ID, 10, 64); err != nil {
			return nil, fmt.Errorf("invalid agent ID: %w", err)
		} else {
			agentID = muid.MUID(id)
		}
	} else {
		return nil, fmt.Errorf("id is required")
	}

	// TODO: implement proper agent starting once manager interface is stable
	return &AgentStartResult{
		ID:     agentID.String(),
		Status: "starting",
	}, nil
}

// AgentStopParams represents parameters for agent.stop method.
type AgentStopParams struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// AgentStopResult represents the result of agent.stop method.
type AgentStopResult struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// agentStop handles the agent.stop RPC method.
func (s *Server) agentStop(ctx context.Context, data json.RawMessage) (interface{}, error) {
	var p AgentStopParams
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Parse agent ID
	var agentID muid.MUID
	if p.ID != "" {
		// Convert string to MUID (assume it's a base-10 representation)
		if id, err := strconv.ParseUint(p.ID, 10, 64); err != nil {
			return nil, fmt.Errorf("invalid agent ID: %w", err)
		} else {
			agentID = muid.MUID(id)
		}
	} else {
		return nil, fmt.Errorf("id is required")
	}

	// TODO: implement proper agent stopping once manager interface is stable
	return &AgentStopResult{
		ID:     agentID.String(),
		Status: "stopping",
	}, nil
}
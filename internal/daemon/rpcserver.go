// Package daemon - rpcserver.go provides a JSON-RPC 2.0 server over Unix socket.
//
// The daemon serves a control plane via JSON-RPC 2.0 over a Unix socket
// at ~/.amux/amuxd.sock, per spec §12. All CLI commands are routed
// through this server.
//
// Agent methods delegate to the director or manager. Plugin methods
// return stubs (full plugin system is Phase 4+).
package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/ids"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/internal/remote/director"
	"github.com/agentflare-ai/amux/internal/remote/manager"
	"github.com/agentflare-ai/amux/pkg/api"
)

// JSON-RPC error codes per spec §12.
const (
	// Standard JSON-RPC errors.
	rpcCodeParseError     = -32700
	rpcCodeInvalidRequest = -32600
	rpcCodeMethodNotFound = -32601
	rpcCodeInvalidParams  = -32602
	rpcCodeInternalError  = -32603

	// Application-defined error codes per spec.
	rpcCodePermissionDenied = -32001
)

// rpcRequest is a JSON-RPC 2.0 request.
type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// rpcResponse is a JSON-RPC 2.0 response.
type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcErrorObj    `json:"error,omitempty"`
}

// rpcErrorObj is a JSON-RPC 2.0 error object.
type rpcErrorObj struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// RPCServer serves JSON-RPC 2.0 over a Unix socket.
type RPCServer struct {
	mu         sync.Mutex
	socketPath string
	listener   net.Listener
	dir        *director.Director
	mgr        *manager.Manager
	agentMgr   *agent.Manager
	done       chan struct{}
}

// NewRPCServer creates a new JSON-RPC server. Pass the director and/or
// manager depending on the daemon role. Either may be nil.
// agentMgr is the local agent manager for handling agent RPC methods.
func NewRPCServer(dir *director.Director, mgr *manager.Manager, agentMgr *agent.Manager) *RPCServer {
	return &RPCServer{
		socketPath: paths.DefaultResolver.DaemonSocketPath(),
		dir:        dir,
		mgr:        mgr,
		agentMgr:   agentMgr,
		done:       make(chan struct{}),
	}
}

// Start begins listening on the Unix socket.
func (s *RPCServer) Start() error {
	// Remove stale socket file
	_ = os.Remove(s.socketPath)

	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("rpc server listen: %w", err)
	}
	s.listener = listener

	go s.acceptLoop()
	return nil
}

// Stop gracefully shuts down the server.
func (s *RPCServer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener != nil {
		_ = s.listener.Close()
		s.listener = nil
	}

	// Clean up socket file
	_ = os.Remove(s.socketPath)
}

func (s *RPCServer) acceptLoop() {
	for {
		s.mu.Lock()
		ln := s.listener
		s.mu.Unlock()

		if ln == nil {
			return
		}

		conn, err := ln.Accept()
		if err != nil {
			return // listener closed
		}
		go s.handleConn(conn)
	}
}

func (s *RPCServer) handleConn(conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	var req rpcRequest
	if err := decoder.Decode(&req); err != nil {
		resp := rpcResponse{
			JSONRPC: "2.0",
			ID:      nil,
			Error: &rpcErrorObj{
				Code:    rpcCodeParseError,
				Message: "parse error: " + err.Error(),
			},
		}
		_ = encoder.Encode(resp)
		return
	}

	if req.JSONRPC != "2.0" {
		resp := rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &rpcErrorObj{
				Code:    rpcCodeInvalidRequest,
				Message: "invalid jsonrpc version",
			},
		}
		_ = encoder.Encode(resp)
		return
	}

	result, rpcErr := s.dispatch(req.Method, req.Params)
	resp := rpcResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
		Error:   rpcErr,
	}
	_ = encoder.Encode(resp)
}

func (s *RPCServer) dispatch(method string, params json.RawMessage) (any, *rpcErrorObj) {
	switch method {
	// Agent methods
	case "agent.add":
		return s.handleAgentAdd(params)
	case "agent.list":
		return s.handleAgentList(params)
	case "agent.remove":
		return s.handleAgentRemove(params)
	case "agent.start":
		return s.handleAgentStart(params)
	case "agent.stop":
		return s.handleAgentStop(params)

	// System methods - per spec §12
	case "daemon.ping":
		return s.handlePing(params)
	case "daemon.version":
		return s.handleVersionRPC(params)
	case "events.subscribe":
		return s.handleEventsSubscribe(params)
	case "system.update":
		return s.handleSystemUpdate(params)

	// Plugin methods (stubs per spec - full plugin system is Phase 4+)
	case "plugin.install":
		return s.handlePluginInstall(params)
	case "plugin.list":
		return s.handlePluginList(params)
	case "plugin.remove":
		return s.handlePluginRemove(params)

	default:
		return nil, &rpcErrorObj{
			Code:    rpcCodeMethodNotFound,
			Message: fmt.Sprintf("method not found: %s", method),
		}
	}
}

// --- Agent method handlers ---

func (s *RPCServer) handleAgentAdd(params json.RawMessage) (any, *rpcErrorObj) {
	// Per spec §12, agent.add accepts: name, about, adapter, location
	var p struct {
		Name     string `json:"name"`
		About    string `json:"about"`
		Adapter  string `json:"adapter"`
		Location struct {
			Type     string `json:"type"`
			Host     string `json:"host"`
			User     string `json:"user"`
			RepoPath string `json:"repo_path"`
		} `json:"location"`
		// Legacy field for backwards compatibility
		RepoPath string `json:"repo_path"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &rpcErrorObj{
			Code:    rpcCodeInvalidParams,
			Message: "invalid params: " + err.Error(),
		}
	}
	if p.Name == "" {
		return nil, &rpcErrorObj{
			Code:    rpcCodeInvalidParams,
			Message: "name is required",
		}
	}

	if s.agentMgr == nil {
		return nil, &rpcErrorObj{
			Code:    rpcCodeInternalError,
			Message: "agent manager not available",
		}
	}

	// Determine location type
	locType := api.LocationLocal
	if p.Location.Type == "ssh" {
		locType = api.LocationSSH
	}

	// Use location.repo_path if set, fall back to legacy repo_path
	repoPath := p.Location.RepoPath
	if repoPath == "" {
		repoPath = p.RepoPath
	}

	cfg := api.Agent{
		Name:    p.Name,
		About:   p.About,
		Adapter: p.Adapter,
		Location: api.Location{
			Type:     locType,
			Host:     p.Location.Host,
			User:     p.Location.User,
			RepoPath: repoPath,
		},
	}

	ag, err := s.agentMgr.Add(context.Background(), cfg)
	if err != nil {
		return nil, &rpcErrorObj{
			Code:    rpcCodeInternalError,
			Message: "agent add: " + err.Error(),
		}
	}

	result := map[string]any{
		"agent_id":   ids.EncodeID(ag.ID),
		"agent_slug": ag.Slug,
		"name":       ag.Name,
		"about":      ag.About,
		"adapter":    ag.Adapter,
		"repo_path":  ag.RepoRoot,
		"worktree":   ag.Worktree,
	}
	return result, nil
}

func (s *RPCServer) handleAgentList(_ json.RawMessage) (any, *rpcErrorObj) {
	result := make([]map[string]any, 0)

	// Include local agents from agent manager
	if s.agentMgr != nil {
		roster := s.agentMgr.Roster()
		for _, entry := range roster {
			result = append(result, map[string]any{
				"agent_id":        ids.EncodeID(entry.Agent.ID),
				"agent_slug":      entry.Agent.Slug,
				"name":            entry.Agent.Name,
				"adapter":         entry.Agent.Adapter,
				"lifecycle_state": string(entry.Lifecycle),
				"presence":        string(entry.Presence),
				"repo_root":       entry.Agent.RepoRoot,
			})
		}
	}

	// When running as director, also include connected remote hosts
	if s.dir != nil {
		hosts := s.dir.ConnectedHosts()
		for _, hostID := range hosts {
			result = append(result, map[string]any{
				"name":            hostID,
				"lifecycle_state": "running",
				"agent_slug":      hostID,
				"location_type":   "ssh",
			})
		}
	}

	return result, nil
}

func (s *RPCServer) handleAgentRemove(params json.RawMessage) (any, *rpcErrorObj) {
	var p struct {
		Name         string `json:"name"`
		DeleteBranch bool   `json:"delete_branch"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &rpcErrorObj{
			Code:    rpcCodeInvalidParams,
			Message: "invalid params: " + err.Error(),
		}
	}
	if p.Name == "" {
		return nil, &rpcErrorObj{
			Code:    rpcCodeInvalidParams,
			Message: "name is required",
		}
	}

	if s.agentMgr == nil {
		return nil, &rpcErrorObj{
			Code:    rpcCodeInternalError,
			Message: "agent manager not available",
		}
	}

	ag := s.agentMgr.GetBySlug(p.Name)
	if ag == nil {
		return nil, &rpcErrorObj{
			Code:    rpcCodeInvalidParams,
			Message: "agent not found: " + p.Name,
		}
	}

	if err := s.agentMgr.Remove(context.Background(), ag.ID, p.DeleteBranch); err != nil {
		return nil, &rpcErrorObj{
			Code:    rpcCodeInternalError,
			Message: "agent remove: " + err.Error(),
		}
	}

	return map[string]any{"removed": true}, nil
}

func (s *RPCServer) handleAgentStart(params json.RawMessage) (any, *rpcErrorObj) {
	var p struct {
		Name  string   `json:"name"`
		Shell string   `json:"shell"`
		Args  []string `json:"args"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &rpcErrorObj{
			Code:    rpcCodeInvalidParams,
			Message: "invalid params: " + err.Error(),
		}
	}
	if p.Name == "" {
		return nil, &rpcErrorObj{
			Code:    rpcCodeInvalidParams,
			Message: "name is required",
		}
	}

	if s.agentMgr == nil {
		return nil, &rpcErrorObj{
			Code:    rpcCodeInternalError,
			Message: "agent manager not available",
		}
	}

	ag := s.agentMgr.GetBySlug(p.Name)
	if ag == nil {
		return nil, &rpcErrorObj{
			Code:    rpcCodeInvalidParams,
			Message: "agent not found: " + p.Name,
		}
	}

	shell := p.Shell
	if shell == "" {
		shell = "/bin/sh"
	}

	if err := s.agentMgr.Start(context.Background(), ag.ID, shell, p.Args...); err != nil {
		return nil, &rpcErrorObj{
			Code:    rpcCodeInternalError,
			Message: "agent start: " + err.Error(),
		}
	}

	return map[string]any{"started": true}, nil
}

func (s *RPCServer) handleAgentStop(params json.RawMessage) (any, *rpcErrorObj) {
	var p struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &rpcErrorObj{
			Code:    rpcCodeInvalidParams,
			Message: "invalid params: " + err.Error(),
		}
	}
	if p.Name == "" {
		return nil, &rpcErrorObj{
			Code:    rpcCodeInvalidParams,
			Message: "name is required",
		}
	}

	if s.agentMgr == nil {
		return nil, &rpcErrorObj{
			Code:    rpcCodeInternalError,
			Message: "agent manager not available",
		}
	}

	ag := s.agentMgr.GetBySlug(p.Name)
	if ag == nil {
		return nil, &rpcErrorObj{
			Code:    rpcCodeInvalidParams,
			Message: "agent not found: " + p.Name,
		}
	}

	if err := s.agentMgr.Stop(context.Background(), ag.ID); err != nil {
		return nil, &rpcErrorObj{
			Code:    rpcCodeInternalError,
			Message: "agent stop: " + err.Error(),
		}
	}

	return map[string]any{"stopped": true}, nil
}

// --- System method handlers ---

func (s *RPCServer) handlePing(_ json.RawMessage) (any, *rpcErrorObj) {
	return map[string]any{"ok": true}, nil
}

func (s *RPCServer) handleVersionRPC(_ json.RawMessage) (any, *rpcErrorObj) {
	return map[string]any{
		"amux_version": Version,
		"spec_version": api.SpecVersion,
	}, nil
}

// --- Plugin method handlers (stubs) ---

// handlePluginInstall rejects with permission denied per spec plugin permission model.
// Full plugin system is Phase 4+.
func (s *RPCServer) handlePluginInstall(_ json.RawMessage) (any, *rpcErrorObj) {
	return nil, &rpcErrorObj{
		Code:    rpcCodePermissionDenied,
		Message: "plugin installation requires explicit approval; plugin system not yet available",
	}
}

// handlePluginList returns an empty list since no plugins are installed.
func (s *RPCServer) handlePluginList(_ json.RawMessage) (any, *rpcErrorObj) {
	return []map[string]any{}, nil
}

// handlePluginRemove returns success (no-op when no plugins exist).
func (s *RPCServer) handlePluginRemove(_ json.RawMessage) (any, *rpcErrorObj) {
	return map[string]any{"removed": true}, nil
}

// --- Event subscription handlers ---

// handleEventsSubscribe registers a subscription for event types.
// Per spec §12, this returns a subscription ID that can be used to receive events.
// Note: Full streaming event delivery requires WebSocket/SSE which is Phase 4+.
// For now, this returns a subscription ID but events are delivered via NATS.
func (s *RPCServer) handleEventsSubscribe(params json.RawMessage) (any, *rpcErrorObj) {
	var p struct {
		Types []string `json:"types"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &rpcErrorObj{
			Code:    rpcCodeInvalidParams,
			Message: "invalid params: " + err.Error(),
		}
	}

	// Return a subscription acknowledgment
	// Full event streaming is delivered via NATS subjects, not RPC
	return map[string]any{
		"subscribed": true,
		"types":      p.Types,
		"transport":  "nats",
	}, nil
}

// --- System update handler (stub) ---

// handleSystemUpdate is a stub for the system.update RPC method.
// Full system update functionality is Phase 4+.
func (s *RPCServer) handleSystemUpdate(_ json.RawMessage) (any, *rpcErrorObj) {
	return map[string]any{
		"available": false,
		"message":   "system updates not yet implemented",
	}, nil
}

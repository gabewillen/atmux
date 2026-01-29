package daemon

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/event"
)

// rpcTestRequest is a JSON-RPC 2.0 request for tests.
type rpcTestRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// rpcTestResponse is a JSON-RPC 2.0 response for tests.
type rpcTestResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func setupRPCServer(t *testing.T) (*RPCServer, string) {
	t.Helper()
	dir := t.TempDir()
	sockPath := filepath.Join(dir, "test.sock")

	srv := &RPCServer{
		socketPath: sockPath,
		done:       make(chan struct{}),
	}

	if err := srv.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { srv.Stop() })

	return srv, sockPath
}

// setupRPCServerWithAgentMgr creates an RPC server with a wired agent.Manager.
func setupRPCServerWithAgentMgr(t *testing.T) (*RPCServer, string) {
	t.Helper()
	dir := t.TempDir()
	sockPath := filepath.Join(dir, "test.sock")

	agentMgr := agent.NewManager(event.NewNoopDispatcher())

	srv := &RPCServer{
		socketPath: sockPath,
		agentMgr:   agentMgr,
		done:       make(chan struct{}),
	}

	if err := srv.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { srv.Stop() })

	return srv, sockPath
}

func callRPC(t *testing.T, sockPath, method string, params any) rpcTestResponse {
	t.Helper()

	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer conn.Close()

	req := rpcTestRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  method,
		Params:  params,
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(req); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	var resp rpcTestResponse
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&resp); err != nil {
		t.Fatalf("Decode: %v", err)
	}

	return resp
}

func TestRPCServer_StartStop(t *testing.T) {
	_, sockPath := setupRPCServer(t)

	// Socket file should exist
	if _, err := os.Stat(sockPath); err != nil {
		t.Fatalf("socket file should exist: %v", err)
	}
}

func TestRPCServer_MethodNotFound(t *testing.T) {
	_, sockPath := setupRPCServer(t)

	resp := callRPC(t, sockPath, "nonexistent.method", nil)
	if resp.Error == nil {
		t.Fatal("expected error for unknown method")
	}
	if resp.Error.Code != rpcCodeMethodNotFound {
		t.Errorf("error code = %d, want %d", resp.Error.Code, rpcCodeMethodNotFound)
	}
}

func TestRPCServer_AgentAdd_NoManager(t *testing.T) {
	// Without agent manager, agent.add returns internal error
	_, sockPath := setupRPCServer(t)

	params := map[string]any{
		"name":    "test-agent",
		"adapter": "claude-code",
	}
	resp := callRPC(t, sockPath, "agent.add", params)

	if resp.Error == nil {
		t.Fatal("expected error when agent manager is not available")
	}
	if resp.Error.Code != rpcCodeInternalError {
		t.Errorf("error code = %d, want %d", resp.Error.Code, rpcCodeInternalError)
	}
}

func TestRPCServer_AgentAdd_MissingName(t *testing.T) {
	_, sockPath := setupRPCServer(t)

	params := map[string]any{"adapter": "claude-code"}
	resp := callRPC(t, sockPath, "agent.add", params)

	if resp.Error == nil {
		t.Fatal("expected error for missing name")
	}
	if resp.Error.Code != rpcCodeInvalidParams {
		t.Errorf("error code = %d, want %d", resp.Error.Code, rpcCodeInvalidParams)
	}
}

func TestRPCServer_AgentList_Empty(t *testing.T) {
	_, sockPath := setupRPCServer(t)

	resp := callRPC(t, sockPath, "agent.list", nil)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	var result []map[string]any
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty list, got %d items", len(result))
	}
}

func TestRPCServer_AgentList_WithManager(t *testing.T) {
	// With agent manager but no agents added, returns empty list
	_, sockPath := setupRPCServerWithAgentMgr(t)

	resp := callRPC(t, sockPath, "agent.list", nil)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	var result []map[string]any
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty list, got %d items", len(result))
	}
}

func TestRPCServer_AgentRemove_NoManager(t *testing.T) {
	_, sockPath := setupRPCServer(t)

	params := map[string]any{"name": "test-agent"}
	resp := callRPC(t, sockPath, "agent.remove", params)

	if resp.Error == nil {
		t.Fatal("expected error when agent manager is not available")
	}
	if resp.Error.Code != rpcCodeInternalError {
		t.Errorf("error code = %d, want %d", resp.Error.Code, rpcCodeInternalError)
	}
}

func TestRPCServer_AgentStart_NoManager(t *testing.T) {
	_, sockPath := setupRPCServer(t)

	params := map[string]any{"name": "test-agent"}
	resp := callRPC(t, sockPath, "agent.start", params)

	if resp.Error == nil {
		t.Fatal("expected error when agent manager is not available")
	}
	if resp.Error.Code != rpcCodeInternalError {
		t.Errorf("error code = %d, want %d", resp.Error.Code, rpcCodeInternalError)
	}
}

func TestRPCServer_AgentStop_NoManager(t *testing.T) {
	_, sockPath := setupRPCServer(t)

	params := map[string]any{"name": "test-agent"}
	resp := callRPC(t, sockPath, "agent.stop", params)

	if resp.Error == nil {
		t.Fatal("expected error when agent manager is not available")
	}
	if resp.Error.Code != rpcCodeInternalError {
		t.Errorf("error code = %d, want %d", resp.Error.Code, rpcCodeInternalError)
	}
}

func TestRPCServer_AgentRemove_NotFound(t *testing.T) {
	_, sockPath := setupRPCServerWithAgentMgr(t)

	params := map[string]any{"name": "nonexistent"}
	resp := callRPC(t, sockPath, "agent.remove", params)

	if resp.Error == nil {
		t.Fatal("expected error for nonexistent agent")
	}
	if resp.Error.Code != rpcCodeInvalidParams {
		t.Errorf("error code = %d, want %d", resp.Error.Code, rpcCodeInvalidParams)
	}
}

func TestRPCServer_AgentStart_NotFound(t *testing.T) {
	_, sockPath := setupRPCServerWithAgentMgr(t)

	params := map[string]any{"name": "nonexistent"}
	resp := callRPC(t, sockPath, "agent.start", params)

	if resp.Error == nil {
		t.Fatal("expected error for nonexistent agent")
	}
	if resp.Error.Code != rpcCodeInvalidParams {
		t.Errorf("error code = %d, want %d", resp.Error.Code, rpcCodeInvalidParams)
	}
}

func TestRPCServer_AgentStop_NotFound(t *testing.T) {
	_, sockPath := setupRPCServerWithAgentMgr(t)

	params := map[string]any{"name": "nonexistent"}
	resp := callRPC(t, sockPath, "agent.stop", params)

	if resp.Error == nil {
		t.Fatal("expected error for nonexistent agent")
	}
	if resp.Error.Code != rpcCodeInvalidParams {
		t.Errorf("error code = %d, want %d", resp.Error.Code, rpcCodeInvalidParams)
	}
}

func TestRPCServer_Ping(t *testing.T) {
	_, sockPath := setupRPCServer(t)

	resp := callRPC(t, sockPath, "daemon.ping", nil)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	var result map[string]any
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if result["ok"] != true {
		t.Errorf("expected ok: true, got %v", result)
	}
}

func TestRPCServer_Version(t *testing.T) {
	_, sockPath := setupRPCServer(t)

	resp := callRPC(t, sockPath, "daemon.version", nil)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	var result map[string]any
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if result["amux_version"] != Version {
		t.Errorf("amux_version = %v, want %v", result["amux_version"], Version)
	}
	if _, ok := result["spec_version"]; !ok {
		t.Error("expected spec_version field in version response")
	}
}

func TestRPCServer_PluginInstall_PermissionDenied(t *testing.T) {
	_, sockPath := setupRPCServer(t)

	params := map[string]any{"source": "/path/to/plugin"}
	resp := callRPC(t, sockPath, "plugin.install", params)

	if resp.Error == nil {
		t.Fatal("expected permission denied error for plugin install")
	}
	if resp.Error.Code != rpcCodePermissionDenied {
		t.Errorf("error code = %d, want %d", resp.Error.Code, rpcCodePermissionDenied)
	}
}

func TestRPCServer_PluginList_Empty(t *testing.T) {
	_, sockPath := setupRPCServer(t)

	resp := callRPC(t, sockPath, "plugin.list", nil)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	var result []map[string]any
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty plugin list, got %d items", len(result))
	}
}

func TestRPCServer_PluginRemove_NoOp(t *testing.T) {
	_, sockPath := setupRPCServer(t)

	params := map[string]any{"name": "some-plugin"}
	resp := callRPC(t, sockPath, "plugin.remove", params)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
}

func TestRPCServer_InvalidJSONRPC(t *testing.T) {
	_, sockPath := setupRPCServer(t)

	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer conn.Close()

	// Send request with wrong version
	req := rpcTestRequest{
		JSONRPC: "1.0",
		ID:      1,
		Method:  "agent.list",
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(req); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	var resp rpcTestResponse
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&resp); err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if resp.Error == nil {
		t.Fatal("expected error for invalid jsonrpc version")
	}
	if resp.Error.Code != rpcCodeInvalidRequest {
		t.Errorf("error code = %d, want %d", resp.Error.Code, rpcCodeInvalidRequest)
	}
}

func TestRPCServer_StopCleansSocket(t *testing.T) {
	dir := t.TempDir()
	sockPath := filepath.Join(dir, "test.sock")

	srv := &RPCServer{
		socketPath: sockPath,
		done:       make(chan struct{}),
	}

	if err := srv.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Verify socket exists
	if _, err := os.Stat(sockPath); err != nil {
		t.Fatalf("socket should exist: %v", err)
	}

	srv.Stop()

	// Verify socket is cleaned up
	if _, err := os.Stat(sockPath); !os.IsNotExist(err) {
		t.Error("socket file should be removed after Stop")
	}
}

package conformance

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBuildFlowList(t *testing.T) {
	tests := []struct {
		name     string
		opts     Options
		wantLen  int
		wantLast string
	}{
		{
			name:     "base flows only",
			opts:     Options{},
			wantLen:  5,
			wantLast: "control_plane",
		},
		{
			name: "with adapter",
			opts: Options{
				AdapterName: "test-adapter",
			},
			wantLen:  6,
			wantLast: "adapter",
		},
		{
			name: "with remote",
			opts: Options{
				RemoteHost: "test-host",
			},
			wantLen:  6,
			wantLast: "remote",
		},
		{
			name: "with adapter and remote",
			opts: Options{
				AdapterName: "test-adapter",
				RemoteHost:  "test-host",
			},
			wantLen:  7,
			wantLast: "remote",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flows := buildFlowList(tt.opts)
			if len(flows) != tt.wantLen {
				t.Errorf("buildFlowList() returned %d flows, want %d", len(flows), tt.wantLen)
			}
			if len(flows) > 0 && flows[len(flows)-1].Name() != tt.wantLast {
				t.Errorf("last flow = %q, want %q", flows[len(flows)-1].Name(), tt.wantLast)
			}
		})
	}
}

func TestFlowNames(t *testing.T) {
	flows := []Flow{
		&AuthFlow{},
		&MenuFlow{},
		&StatusFlow{},
		&NotificationFlow{},
		&ControlPlaneFlow{},
		&AdapterFlow{},
		&RemoteFlow{},
	}

	expectedNames := []string{
		"auth",
		"menu",
		"status",
		"notification",
		"control_plane",
		"adapter",
		"remote",
	}

	for i, flow := range flows {
		if flow.Name() != expectedNames[i] {
			t.Errorf("flow %d Name() = %q, want %q", i, flow.Name(), expectedNames[i])
		}
	}
}

func TestResultsPassed(t *testing.T) {
	tests := []struct {
		name   string
		flows  []FlowResult
		passed bool
	}{
		{
			name:   "empty flows",
			flows:  []FlowResult{},
			passed: true,
		},
		{
			name: "all pass",
			flows: []FlowResult{
				{Name: "auth", Status: StatusPass},
				{Name: "menu", Status: StatusPass},
			},
			passed: true,
		},
		{
			name: "one fail",
			flows: []FlowResult{
				{Name: "auth", Status: StatusPass},
				{Name: "menu", Status: StatusFail},
			},
			passed: false,
		},
		{
			name: "skip counts as pass",
			flows: []FlowResult{
				{Name: "auth", Status: StatusPass},
				{Name: "remote", Status: StatusSkip},
			},
			passed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Results{Flows: tt.flows}
			if r.Passed() != tt.passed {
				t.Errorf("Passed() = %v, want %v", r.Passed(), tt.passed)
			}
		})
	}
}

func TestIsValidLifecycleState(t *testing.T) {
	tests := []struct {
		state string
		valid bool
	}{
		{"pending", true},
		{"starting", true},
		{"running", true},
		{"terminated", true},
		{"errored", true},
		{"", true}, // Empty is acceptable
		{"invalid", false},
		{"RUNNING", false}, // Case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			if isValidLifecycleState(tt.state) != tt.valid {
				t.Errorf("isValidLifecycleState(%q) = %v, want %v", tt.state, !tt.valid, tt.valid)
			}
		})
	}
}

func TestIsValidPresenceState(t *testing.T) {
	tests := []struct {
		state string
		valid bool
	}{
		{"online", true},
		{"busy", true},
		{"offline", true},
		{"away", true},
		{"", true}, // Empty is acceptable
		{"invalid", false},
		{"ONLINE", false}, // Case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			if isValidPresenceState(tt.state) != tt.valid {
				t.Errorf("isValidPresenceState(%q) = %v, want %v", tt.state, !tt.valid, tt.valid)
			}
		})
	}
}

func TestWriteResults(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "results", "test.json")

	results := &Results{
		RunID:       "test-run-123",
		SpecVersion: "v1.22",
		StartedAt:   time.Now().UTC(),
		FinishedAt:  time.Now().UTC(),
		Flows: []FlowResult{
			{
				Name:       "auth",
				Status:     StatusPass,
				StartedAt:  time.Now().UTC(),
				FinishedAt: time.Now().UTC(),
			},
		},
	}

	if err := WriteResults(results, path); err != nil {
		t.Fatalf("WriteResults() error = %v", err)
	}

	// Verify file exists and contains valid JSON
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	var loaded Results
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if loaded.RunID != results.RunID {
		t.Errorf("RunID = %q, want %q", loaded.RunID, results.RunID)
	}

	if len(loaded.Flows) != 1 {
		t.Errorf("len(Flows) = %d, want 1", len(loaded.Flows))
	}
}

func TestAuthFlowRequiresDaemonAddr(t *testing.T) {
	f := &AuthFlow{}
	err := f.Run(context.Background(), Options{})
	if err == nil {
		t.Error("expected error when daemon_addr is empty")
	}
}

func TestMenuFlowRequiresAdapterName(t *testing.T) {
	f := &MenuFlow{}
	err := f.Run(context.Background(), Options{DaemonAddr: "/tmp/test.sock"})
	if err == nil {
		t.Error("expected error when adapter_name is empty")
	}
}

func TestStatusFlowRequiresDaemonAddr(t *testing.T) {
	f := &StatusFlow{}
	err := f.Run(context.Background(), Options{})
	if err == nil {
		t.Error("expected error when daemon_addr is empty")
	}
}

func TestNotificationFlowRequiresDaemonAddr(t *testing.T) {
	f := &NotificationFlow{}
	err := f.Run(context.Background(), Options{})
	if err == nil {
		t.Error("expected error when daemon_addr is empty")
	}
}

func TestControlPlaneFlowRequiresDaemonAddr(t *testing.T) {
	f := &ControlPlaneFlow{}
	err := f.Run(context.Background(), Options{})
	if err == nil {
		t.Error("expected error when daemon_addr is empty")
	}
}

func TestAdapterFlowRequiresAdapterName(t *testing.T) {
	f := &AdapterFlow{}
	err := f.Run(context.Background(), Options{DaemonAddr: "/tmp/test.sock"})
	if err == nil {
		t.Error("expected error when adapter_name is empty")
	}
}

func TestRemoteFlowRequiresRemoteHost(t *testing.T) {
	f := &RemoteFlow{}
	err := f.Run(context.Background(), Options{DaemonAddr: "/tmp/test.sock"})
	if err == nil {
		t.Error("expected error when remote_host is empty")
	}
}

// mockDaemon creates a simple mock JSON-RPC server for testing.
// The handler is called for each request and can return different responses.
func mockDaemon(t *testing.T, handler func(req rpcRequest) rpcResponse) (string, func()) {
	t.Helper()

	tmpDir := t.TempDir()
	sockPath := filepath.Join(tmpDir, "daemon.sock")

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}

	done := make(chan struct{})
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-done:
					return
				default:
					continue
				}
			}

			go func(c net.Conn) {
				defer c.Close()
				// Handle multiple requests on the same connection
				for {
					buf := make([]byte, 65536)
					n, err := c.Read(buf)
					if err != nil {
						return
					}

					var req rpcRequest
					if err := json.Unmarshal(buf[:n], &req); err != nil {
						return
					}

					resp := handler(req)
					data, _ := json.Marshal(resp)
					_, _ = c.Write(data)
				}
			}(conn)
		}
	}()

	return sockPath, func() {
		close(done)
		listener.Close()
	}
}

func TestAuthFlowWithMockDaemon(t *testing.T) {
	sockPath, cleanup := mockDaemon(t, func(req rpcRequest) rpcResponse {
		return rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"version": "0.1.0"}`),
		}
	})
	defer cleanup()

	f := &AuthFlow{}
	err := f.Run(context.Background(), Options{
		DaemonAddr: sockPath,
		Timeout:    2 * time.Second,
	})
	if err != nil {
		t.Errorf("AuthFlow.Run() error = %v", err)
	}
}

func TestStatusFlowWithMockDaemon(t *testing.T) {
	sockPath, cleanup := mockDaemon(t, func(req rpcRequest) rpcResponse {
		switch req.Method {
		case "agent.list":
			return rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  json.RawMessage(`[{"id": "123", "name": "test", "lifecycle": "running", "presence": "online"}]`),
			}
		case "roster.list":
			return rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  json.RawMessage(`[]`),
			}
		case "event.subscribe":
			return rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  json.RawMessage(`{"subscription_id": "sub-123"}`),
			}
		default:
			return rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &rpcError{Code: -32601, Message: "method not found"},
			}
		}
	})
	defer cleanup()

	f := &StatusFlow{}
	err := f.Run(context.Background(), Options{
		DaemonAddr: sockPath,
		Timeout:    2 * time.Second,
	})
	if err != nil {
		t.Errorf("StatusFlow.Run() error = %v", err)
	}
}

func TestControlPlaneFlowWithMockDaemon(t *testing.T) {
	sockPath, cleanup := mockDaemon(t, func(req rpcRequest) rpcResponse {
		switch req.Method {
		case "agent.add", "agent.list", "agent.remove", "agent.start", "agent.stop":
			// Required methods - return success or valid error
			return rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &rpcError{Code: -32602, Message: "invalid params"},
			}
		case "nonexistent.method":
			return rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &rpcError{Code: -32601, Message: "method not found"},
			}
		default:
			return rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &rpcError{Code: -32601, Message: "method not found"},
			}
		}
	})
	defer cleanup()

	f := &ControlPlaneFlow{}
	err := f.Run(context.Background(), Options{
		DaemonAddr: sockPath,
		Timeout:    5 * time.Second,
	})
	if err != nil {
		t.Errorf("ControlPlaneFlow.Run() error = %v", err)
	}
}

func TestRunConformanceSuite(t *testing.T) {
	sockPath, cleanup := mockDaemon(t, func(req rpcRequest) rpcResponse {
		switch req.Method {
		case "agent.list":
			return rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  json.RawMessage(`[]`),
			}
		case "agent.add", "agent.remove", "agent.start", "agent.stop":
			return rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &rpcError{Code: -32602, Message: "invalid params"},
			}
		case "nonexistent.method":
			return rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &rpcError{Code: -32601, Message: "method not found"},
			}
		default:
			return rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &rpcError{Code: -32601, Message: "method not found"},
			}
		}
	})
	defer cleanup()

	ctx := context.Background()
	results, err := Run(ctx, Options{
		DaemonAddr:  sockPath,
		AdapterName: "test-adapter",
		Timeout:     2 * time.Second,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if results.RunID == "" {
		t.Error("RunID should not be empty")
	}

	if results.SpecVersion == "" {
		t.Error("SpecVersion should not be empty")
	}

	// We expect 6 flows: auth, menu, status, notification, control_plane, adapter
	if len(results.Flows) != 6 {
		t.Errorf("len(Flows) = %d, want 6", len(results.Flows))
	}
}

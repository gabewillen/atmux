// Package cli - rpc.go provides JSON-RPC 2.0 client for daemon communication.
//
// The CLI communicates with the amux daemon (amuxd) over a Unix socket
// using JSON-RPC 2.0 per spec §12.
package cli

import (
	"encoding/json"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/agentflare-ai/amux/internal/paths"
)

// RPCClient communicates with the amux daemon over Unix socket.
type RPCClient struct {
	socketPath string
	nextID     atomic.Int64
}

// rpcRequest is a JSON-RPC 2.0 request.
type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// rpcResponse is a JSON-RPC 2.0 response.
type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

// rpcError is a JSON-RPC 2.0 error object.
type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (e *rpcError) Error() string {
	return fmt.Sprintf("rpc error %d: %s", e.Code, e.Message)
}

// NewRPCClient creates a new JSON-RPC client connected to the daemon socket.
func NewRPCClient() *RPCClient {
	return &RPCClient{
		socketPath: paths.DefaultResolver.DaemonSocketPath(),
	}
}

// Call sends a JSON-RPC request and returns the result.
func (c *RPCClient) Call(method string, params any, result any) error {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return fmt.Errorf("connect to daemon at %s: %w (is amuxd running?)", c.socketPath, err)
	}
	defer conn.Close()

	// Build request
	id := c.nextID.Add(1)
	req := rpcRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	// Encode and send
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(req); err != nil {
		return fmt.Errorf("send request: %w", err)
	}

	// Read response
	decoder := json.NewDecoder(conn)
	var resp rpcResponse
	if err := decoder.Decode(&resp); err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// Check for error
	if resp.Error != nil {
		return resp.Error
	}

	// Decode result
	if result != nil && resp.Result != nil {
		if err := json.Unmarshal(resp.Result, result); err != nil {
			return fmt.Errorf("decode result: %w", err)
		}
	}

	return nil
}

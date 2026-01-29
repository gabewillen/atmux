package rpc

import "encoding/json"

// Request is a JSON-RPC request or notification.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response is a JSON-RPC response.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

// Error describes a JSON-RPC error.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

const (
	// CodeParseError indicates invalid JSON.
	CodeParseError = -32700
	// CodeInvalidRequest indicates invalid JSON-RPC.
	CodeInvalidRequest = -32600
	// CodeMethodNotFound indicates missing method.
	CodeMethodNotFound = -32601
	// CodeInvalidParams indicates invalid parameters.
	CodeInvalidParams = -32602
	// CodeInternalError indicates a server error.
	CodeInternalError = -32603
)

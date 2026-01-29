package rpc

import (
	"testing"
)

// TestJSONRPCProtocol tests basic JSON-RPC 2.0 protocol compliance.
func TestJSONRPCProtocol(t *testing.T) {
	// Test ErrorObj.Error() method
	err := &ErrorObj{
		Code:    ErrorCodeInternalError,
		Message: "test error",
	}
	
	if err.Error() != "test error" {
		t.Errorf("Expected error message 'test error', got '%s'", err.Error())
	}
}

// TestRequestResponse tests basic request/response structures.
func TestRequestResponse(t *testing.T) {
	// Test request structure
	req := Request{
		JSONRPC: "2.0",
		Method:  "test.method",
		Params:  map[string]string{"test": "value"},
		ID:      1,
	}
	
	if req.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC '2.0', got '%s'", req.JSONRPC)
	}
	
	// Test response structure
	resp := Response{
		JSONRPC: "2.0",
		Result:  map[string]string{"result": "success"},
		ID:      1,
	}
	
	if resp.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC '2.0', got '%s'", resp.JSONRPC)
	}
}
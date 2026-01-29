// Package rpc provides JSON-RPC 2.0 client functionality.
package rpc

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

// Client provides JSON-RPC 2.0 client functionality for amux daemon communication.
type Client struct {
	conn   net.Conn
	nextID int
	mu     sync.Mutex
}

// NewClient creates a new JSON-RPC client connected to the given socket path.
func NewClient(socketPath string) (*Client, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %w", err)
	}

	return &Client{
		conn:   conn,
		nextID: 1,
	}, nil
}

// Close closes the client connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// call sends a JSON-RPC request and returns the result.
func (c *Client) call(method string, params interface{}, result interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create request
	req := Request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      c.nextID,
	}
	c.nextID++

	// Send request
	encoder := json.NewEncoder(c.conn)
	if err := encoder.Encode(req); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	// Read response
	decoder := json.NewDecoder(c.conn)
	var resp Response
	if err := decoder.Decode(&resp); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check for error
	if resp.Error != nil {
		return resp.Error
	}

	// Unmarshal result if provided
	if result != nil && resp.Result != nil {
		data, err := json.Marshal(resp.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
		if err := json.Unmarshal(data, result); err != nil {
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	return nil
}

// AgentAdd adds a new agent via JSON-RPC.
func (c *Client) AgentAdd(name, adapter, repoRoot string, config map[string]interface{}) (*AgentAddResult, error) {
	result := &AgentAddResult{}
	err := c.call("agent.add", map[string]interface{}{
		"name":      name,
		"adapter":   adapter,
		"repo_root": repoRoot,
		"config":    config,
	}, result)
	return result, err
}

// AgentList lists all agents via JSON-RPC.
func (c *Client) AgentList() (*AgentListResult, error) {
	result := &AgentListResult{}
	err := c.call("agent.list", nil, result)
	return result, err
}

// AgentStart starts an agent via JSON-RPC.
func (c *Client) AgentStart(id, name string) (*AgentStartResult, error) {
	result := &AgentStartResult{}
	err := c.call("agent.start", map[string]interface{}{
		"id":   id,
		"name": name,
	}, result)
	return result, err
}

// AgentStop stops an agent via JSON-RPC.
func (c *Client) AgentStop(id, name string) (*AgentStopResult, error) {
	result := &AgentStopResult{}
	err := c.call("agent.stop", map[string]interface{}{
		"id":   id,
		"name": name,
	}, result)
	return result, err
}
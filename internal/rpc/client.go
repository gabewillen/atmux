package rpc

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// Client issues JSON-RPC requests over a Unix socket.
type Client struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	mu     sync.Mutex
	nextID uint64
}

// Dial connects to a JSON-RPC socket.
func Dial(ctx context.Context, socketPath string) (*Client, error) {
	if socketPath == "" {
		return nil, fmt.Errorf("rpc dial: socket path is empty")
	}
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("rpc dial: %w", err)
	}
	return &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}, nil
}

// Close closes the underlying connection.
func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("rpc close: %w", err)
	}
	return nil
}

// Call sends a request and decodes the response.
func (c *Client) Call(ctx context.Context, method string, params any, result any) error {
	if c == nil {
		return fmt.Errorf("rpc call: client is nil")
	}
	if ctx.Err() != nil {
		return fmt.Errorf("rpc call: %w", ctx.Err())
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	id := atomic.AddUint64(&c.nextID, 1)
	idRaw := json.RawMessage(strconv.FormatUint(id, 10))
	var paramsRaw json.RawMessage
	if params != nil {
		encoded, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("rpc call: %w", err)
		}
		paramsRaw = encoded
	}
	req := Request{JSONRPC: "2.0", ID: idRaw, Method: method, Params: paramsRaw}
	if err := writeJSON(c.writer, req); err != nil {
		return fmt.Errorf("rpc call: %w", err)
	}
	deadline := time.Now().Add(30 * time.Second)
	if dl, ok := ctx.Deadline(); ok {
		deadline = dl
	}
	if err := c.conn.SetReadDeadline(deadline); err != nil {
		return fmt.Errorf("rpc call: %w", err)
	}
	for {
		line, err := c.reader.ReadBytes('\n')
		if err != nil {
			return fmt.Errorf("rpc call: %w", err)
		}
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		var resp Response
		if err := json.Unmarshal(line, &resp); err != nil {
			return fmt.Errorf("rpc call: %w", err)
		}
		if !bytes.Equal(resp.ID, idRaw) {
			continue
		}
		if resp.Error != nil {
			return fmt.Errorf("rpc call: %s", resp.Error.Message)
		}
		if result != nil {
			encoded, err := json.Marshal(resp.Result)
			if err != nil {
				return fmt.Errorf("rpc call: %w", err)
			}
			if err := json.Unmarshal(encoded, result); err != nil {
				return fmt.Errorf("rpc call: %w", err)
			}
		}
		return nil
	}
}

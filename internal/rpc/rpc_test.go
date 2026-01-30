package rpc

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestServerClientRoundTrip(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "rpc.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()
	server := NewServer(nil)
	server.Register("ping", func(ctx context.Context, raw json.RawMessage) (any, *Error) {
		_ = ctx
		_ = raw
		return map[string]any{"ok": true}, nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = server.Serve(ctx, listener)
	}()
	client, err := Dial(context.Background(), socketPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer client.Close()
	var result map[string]any
	if err := client.Call(context.Background(), "ping", nil, &result); err != nil {
		t.Fatalf("call: %v", err)
	}
	if ok, exists := result["ok"]; !exists || ok != true {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestClientErrors(t *testing.T) {
	if _, err := Dial(context.Background(), ""); err == nil {
		t.Fatalf("expected dial error")
	}
	var client *Client
	if err := client.Call(context.Background(), "noop", nil, nil); err == nil {
		t.Fatalf("expected call error")
	}
}

func TestServerErrors(t *testing.T) {
	server := NewServer(nil)
	if err := server.Serve(context.Background(), nil); err == nil {
		t.Fatalf("expected serve error")
	}
}

func TestServerHandlesInvalidRequests(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "rpc.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()
	server := NewServer(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = server.Serve(ctx, listener)
	}()
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	_, _ = conn.Write([]byte("{invalid}\n"))
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(line, "parse error") {
		t.Fatalf("unexpected response: %s", line)
	}
	req := Request{JSONRPC: "2.0", ID: json.RawMessage("1"), Method: "missing"}
	encoded, _ := json.Marshal(req)
	_, _ = conn.Write(append(encoded, '\n'))
	line, err = reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read missing: %v", err)
	}
	if !strings.Contains(line, "method not found") {
		t.Fatalf("unexpected missing response: %s", line)
	}
	badReq := Request{JSONRPC: "1.0", ID: json.RawMessage("2"), Method: "ping"}
	encoded, _ = json.Marshal(badReq)
	_, _ = conn.Write(append(encoded, '\n'))
	line, err = reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read invalid: %v", err)
	}
	if !strings.Contains(line, "invalid request") {
		t.Fatalf("unexpected invalid response: %s", line)
	}
}

func TestServerShutsDownOnContextCancel(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "rpc.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	server := NewServer(nil)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- server.Serve(ctx, listener)
	}()
	cancel()
	_ = listener.Close()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("serve error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("server did not stop")
	}
	_ = os.Remove(socketPath)
}

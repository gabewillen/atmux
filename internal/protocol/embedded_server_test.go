package protocol

import (
	"bufio"
	"context"
	"net"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestEmbeddedServerAuthRejects(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	server, err := StartEmbeddedServer(ctx, "127.0.0.1:0", EmbeddedServerConfig{
		Auth: AuthConfig{
			Tokens: map[string]Permissions{
				"good": {
					Publish:   []string{"amux.*"},
					Subscribe: []string{"amux.*"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	defer func() {
		_ = server.Close()
	}()
	hostPort := mustHostPort(t, server.URL())
	conn, err := net.Dial("tcp", hostPort)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read info: %v", err)
	}
	if !strings.HasPrefix(line, "INFO") {
		t.Fatalf("expected INFO")
	}
	if _, err := conn.Write([]byte("CONNECT {\"auth_token\":\"bad\"}\r\n")); err != nil {
		t.Fatalf("connect: %v", err)
	}
	errLine, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read err: %v", err)
	}
	if !strings.HasPrefix(errLine, "-ERR") {
		t.Fatalf("expected -ERR, got %q", errLine)
	}

	conn2, err := net.Dial("tcp", hostPort)
	if err != nil {
		t.Fatalf("dial second: %v", err)
	}
	defer func() {
		_ = conn2.Close()
	}()
	reader2 := bufio.NewReader(conn2)
	_, err = reader2.ReadString('\n')
	if err != nil {
		t.Fatalf("read info: %v", err)
	}
	if _, err := conn2.Write([]byte("CONNECT {\"auth_token\":\"good\"}\r\n")); err != nil {
		t.Fatalf("connect: %v", err)
	}
	if _, err := conn2.Write([]byte("SUB blocked.test 1\r\n")); err != nil {
		t.Fatalf("sub: %v", err)
	}
	errLine, err = reader2.ReadString('\n')
	if err != nil {
		t.Fatalf("read err: %v", err)
	}
	if !strings.HasPrefix(errLine, "-ERR") {
		t.Fatalf("expected -ERR for sub, got %q", errLine)
	}
}

func mustHostPort(t *testing.T, rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	return parsed.Host
}

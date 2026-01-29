package remote

import (
	"context"
	"strings"
	"testing"

	"github.com/agentflare-ai/amux/internal/config"
)

func TestNewDirector(t *testing.T) {
	cfg := &config.RemoteConfig{NATS: config.RemoteNATSConfig{SubjectPrefix: "amux"}}
	d := NewDirector(cfg)
	if d == nil {
		t.Fatal("NewDirector returned nil")
	}
	if d.prefix != "amux" {
		t.Errorf("prefix = %q, want amux", d.prefix)
	}
}

func TestDirector_RequestTimeout(t *testing.T) {
	cfg := &config.RemoteConfig{RequestTimeout: "3s"}
	d := NewDirector(cfg)
	timeout := d.RequestTimeout()
	if timeout.Seconds() != 3 {
		t.Errorf("RequestTimeout = %v, want 3s", timeout)
	}
}

func TestDirector_IsReady(t *testing.T) {
	d := NewDirector(nil)
	if d.IsReady("host1") {
		t.Error("IsReady(host1) want false before handshake")
	}
}

func TestDirector_Close(t *testing.T) {
	d := NewDirector(nil)
	d.Close()
	if d.nc != nil {
		t.Error("Close should clear nc")
	}
}

func TestExtractHostIDFromSubject(t *testing.T) {
	prefix := "amux.handshake."
	subject := "amux.handshake.devbox"
	got := extractHostIDFromSubject(prefix, subject)
	if got != "devbox" {
		t.Errorf("extractHostIDFromSubject = %q, want devbox", got)
	}
}

func TestDirector_SpawnFailFastWhenNotReady(t *testing.T) {
	d := NewDirector(nil)
	ctx := context.Background()
	_, err := d.Spawn(ctx, "host1", SpawnPayloadRequest{AgentID: "1", AgentSlug: "a", RepoPath: "/repo"})
	if err == nil {
		t.Fatal("Spawn should fail when host not ready")
	}
	if !strings.Contains(err.Error(), "not ready") {
		t.Errorf("error = %v, want 'not ready'", err)
	}
}

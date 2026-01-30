package remote

import (
	"context"
	"testing"

	"github.com/agentflare-ai/amux/pkg/api"
)

func TestHostManagerKVErrors(t *testing.T) {
	manager := &HostManager{}
	if err := manager.writeSessionKV(context.Background(), nil, "running", nil); err == nil {
		t.Fatalf("expected missing session error")
	}
	session := &remoteSession{
		agentID:   api.NewAgentID(),
		sessionID: api.NewSessionID(),
		slug:      "alpha",
		repoPath:  "/tmp",
	}
	if err := manager.writeSessionKV(context.Background(), session, "running", nil); err == nil {
		t.Fatalf("expected kv unavailable error")
	}
	if err := manager.writeHostKV(context.Background()); err == nil {
		t.Fatalf("expected host kv error")
	}
}

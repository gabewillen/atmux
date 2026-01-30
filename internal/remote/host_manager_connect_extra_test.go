package remote

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/paths"
)

func TestHostManagerConnectErrors(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repoRoot, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir git: %v", err)
	}
	resolver, err := paths.NewResolver(repoRoot)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	cfg := config.DefaultConfig(resolver)
	cfg.NATS.JetStreamDir = filepath.Join(t.TempDir(), "jetstream")
	cfg.Remote.NATS.SubjectPrefix = "amux"
	cfg.Remote.NATS.CredsPath = ""
	manager, err := NewHostManager(cfg, resolver, "test")
	if err != nil {
		t.Fatalf("host manager: %v", err)
	}
	if err := manager.connect(context.Background()); err == nil {
		t.Fatalf("expected missing creds error")
	}
	cfg.Remote.NATS.CredsPath = filepath.Join(t.TempDir(), "missing.creds")
	manager, err = NewHostManager(cfg, resolver, "test")
	if err != nil {
		t.Fatalf("host manager: %v", err)
	}
	if err := manager.connect(context.Background()); err == nil {
		t.Fatalf("expected missing file error")
	}
	credsPath := filepath.Join(t.TempDir(), "host.creds")
	if err := os.WriteFile(credsPath, []byte("creds"), 0o600); err != nil {
		t.Fatalf("write creds: %v", err)
	}
	cfg.Remote.NATS.CredsPath = credsPath
	manager, err = NewHostManager(cfg, resolver, "test")
	if err != nil {
		t.Fatalf("host manager: %v", err)
	}
	if err := manager.connect(context.Background()); err == nil {
		t.Fatalf("expected leaf server error")
	}
}

func TestCredentialMarshal(t *testing.T) {
	cred, err := ParseCredential([]byte("data"))
	if err != nil {
		t.Fatalf("parse credential: %v", err)
	}
	out, err := cred.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(out) != "data" {
		t.Fatalf("unexpected marshal data")
	}
	var empty Credential
	if _, err := empty.Marshal(); err == nil {
		t.Fatalf("expected marshal error")
	}
}

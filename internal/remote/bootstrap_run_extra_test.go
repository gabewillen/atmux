package remote

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/agentflare-ai/amux/pkg/api"
)

type sequenceRunner struct {
	outputs map[string][][]byte
}

func (s *sequenceRunner) Run(ctx context.Context, target string, options []string, command string, stdin []byte) error {
	_ = ctx
	_ = target
	_ = options
	_ = command
	_ = stdin
	return nil
}

func (s *sequenceRunner) RunOutput(ctx context.Context, target string, options []string, command string, stdin []byte) ([]byte, error) {
	_ = ctx
	_ = target
	_ = options
	_ = stdin
	if list, ok := s.outputs[command]; ok && len(list) > 0 {
		out := list[0]
		s.outputs[command] = list[1:]
		return out, nil
	}
	return []byte(""), nil
}

func TestBootstrapperBootstrap(t *testing.T) {
	tmp := t.TempDir()
	credStore, err := NewCredentialStore(filepath.Join(tmp, "creds"))
	if err != nil {
		t.Fatalf("cred store: %v", err)
	}
	cred, err := credStore.DirectorCredential()
	if err != nil {
		t.Fatalf("cred: %v", err)
	}
	runner := &sequenceRunner{
		outputs: map[string][][]byte{
			"uname -s": {[]byte(runtime.GOOS)},
			"uname -m": {[]byte(runtime.GOARCH)},
			"PATH=\"$HOME/.local/bin:$PATH\" amux-manager status": {
				[]byte("hub_connected=false"),
				[]byte("hub_connected=true"),
			},
		},
	}
	req := BootstrapRequest{
		HostID:        "host",
		Location:      api.Location{Type: api.LocationSSH, Host: "example.com"},
		LeafURL:       "nats://leaf",
		HubClientURL:  "nats://hub",
		CredsPath:     filepath.Join(tmp, "host.creds"),
		SubjectPrefix: "amux",
		KVBucket:      "kv",
		ManagerModel:  "model",
	}
	bootstrapper := &Bootstrapper{Runner: runner}
	if err := bootstrapper.Bootstrap(context.Background(), req, cred); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
}

func TestBootstrapperBootstrapInvalid(t *testing.T) {
	bootstrapper := &Bootstrapper{Runner: &sequenceRunner{}}
	if err := bootstrapper.Bootstrap(context.Background(), BootstrapRequest{}, Credential{}); err == nil {
		t.Fatalf("expected host id error")
	}
	req := BootstrapRequest{
		HostID: "host",
		Location: api.Location{
			Type: api.LocationSSH,
		},
	}
	if err := bootstrapper.Bootstrap(context.Background(), req, Credential{}); err == nil {
		t.Fatalf("expected location host error")
	}
}

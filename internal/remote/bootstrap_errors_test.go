package remote

import (
	"context"
	"testing"

	"github.com/agentflare-ai/amux/pkg/api"
)

func TestBootstrapperValidationErrors(t *testing.T) {
	b := &Bootstrapper{}
	if err := b.Bootstrap(context.Background(), BootstrapRequest{}, Credential{}); err == nil {
		t.Fatalf("expected host id error")
	}
	req := BootstrapRequest{HostID: api.MustParseHostID("host"), Location: api.Location{Type: api.LocationSSH}}
	if err := b.Bootstrap(context.Background(), req, Credential{}); err == nil {
		t.Fatalf("expected location host error")
	}
}


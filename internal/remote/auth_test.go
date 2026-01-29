package remote

import (
	"strings"
	"testing"

	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/jwt/v2"
)

func TestGenerateHostCredentials(t *testing.T) {
	hostID := api.HostID("test-host")
	prefix := "amux"

	creds, token, err := GenerateHostCredentials(hostID, prefix)
	if err != nil {
		t.Fatalf("GenerateHostCredentials failed: %v", err)
	}

	if !strings.Contains(creds, "BEGIN NATS USER JWT") {
		t.Error("Creds missing JWT")
	}
	if !strings.Contains(creds, "BEGIN USER NKEY SEED") {
		t.Error("Creds missing Seed")
	}

	// Decode JWT
	claims, err := jwt.DecodeUserClaims(token)
	if err != nil {
		t.Fatalf("Failed to decode JWT: %v", err)
	}

	if claims.Name != "test-host" {
		t.Errorf("Expected name 'test-host', got %q", claims.Name)
	}

	// Check permissions
	if !claims.Pub.Allow.Contains("amux.events.test-host") {
		t.Error("Pub allow missing events subject")
	}
	if !claims.Sub.Allow.Contains("amux.ctl.test-host") {
		t.Error("Sub allow missing ctl subject")
	}
	if claims.Pub.Allow.Contains("amux.ctl.test-host") {
		t.Error("Pub allow should NOT contain ctl subject")
	}
}

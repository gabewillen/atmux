package remote

import (
	"testing"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestHostManagerStatus(t *testing.T) {
	manager := &HostManager{hostID: api.MustParseHostID("host")}
	manager.connected = true
	manager.ready = true
	status := manager.Status()
	if !status.Connected || !status.Ready || status.HostID != "host" {
		t.Fatalf("unexpected status: %#v", status)
	}
}

func TestHandlePTYInputInvalidSubject(t *testing.T) {
	manager := &HostManager{hostID: api.MustParseHostID("host"), subjectPrefix: "amux"}
	manager.handlePTYInput(protocol.Message{Subject: "bad.subject", Data: []byte("data")})
}

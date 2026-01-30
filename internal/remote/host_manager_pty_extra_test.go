package remote

import (
	"os"
	"testing"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/pty"
	"github.com/agentflare-ai/amux/internal/session"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestHandlePTYInputWrites(t *testing.T) {
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	t.Cleanup(func() {
		_ = read.Close()
		_ = write.Close()
	})
	runtime := &session.LocalSession{}
	setUnexportedField(runtime, "ptyPair", &pty.Pair{Master: write})
	sessionID := api.NewSessionID()
	manager := &HostManager{
		hostID:        api.MustParseHostID("host"),
		subjectPrefix: "amux",
		sessions: map[api.SessionID]*remoteSession{
			sessionID: {runtime: runtime},
		},
	}
	subject := PtyInSubject("amux", manager.hostID, sessionID)
	manager.handlePTYInput(protocol.Message{Subject: subject, Data: []byte("ping")})
	buf := make([]byte, 4)
	if _, err := read.Read(buf); err != nil {
		t.Fatalf("read: %v", err)
	}
}

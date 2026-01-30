package remote

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/pty"
	"github.com/agentflare-ai/amux/internal/session"
	"github.com/agentflare-ai/amux/pkg/api"
)

type staticMatcher struct {
	match adapter.PatternMatch
}

func (s staticMatcher) Match(ctx context.Context, output []byte) ([]adapter.PatternMatch, error) {
	_ = ctx
	_ = output
	return []adapter.PatternMatch{s.match}, nil
}

func TestHandleOutboundMessagesUnknownRecipient(t *testing.T) {
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
	payload, err := json.Marshal(api.OutboundMessage{ToSlug: "missing", Content: "hello"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	manager := &HostManager{
		hostID: api.MustParseHostID("host"),
	}
	session := &remoteSession{
		agentID:  api.NewAgentID(),
		slug:     "alpha",
		runtime:  runtime,
		matcher:  staticMatcher{match: adapter.PatternMatch{Pattern: "message", Text: string(payload)}},
	}
	manager.handleOutboundMessages(session, []byte("trigger"))
	buf := make([]byte, 256)
	n, err := read.Read(buf)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(buf[:n]), "unknown recipient") {
		t.Fatalf("expected unknown recipient message")
	}
}

func TestHandleOutboundMessagesBroadcastPublishes(t *testing.T) {
	dispatcher := &rawDispatcher{}
	payload, err := json.Marshal(api.OutboundMessage{ToSlug: "broadcast", Content: "hello"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	manager := &HostManager{
		subjectPrefix: "amux",
		hostID:        api.MustParseHostID("host"),
		dispatcher:    dispatcher,
		outbox:        NewOutbox(1024),
		connected:     true,
	}
	session := &remoteSession{
		agentID: api.NewAgentID(),
		matcher: staticMatcher{match: adapter.PatternMatch{Pattern: "message", Text: string(payload)}},
	}
	manager.handleOutboundMessages(session, []byte("trigger"))
	found := false
	for _, subject := range dispatcher.rawSubjects {
		if subject == BroadcastCommSubject("amux") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected broadcast subject publish")
	}
}

func TestHandleCommMessageUnicastDelivers(t *testing.T) {
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
	agentID := api.NewAgentID()
	sessionID := api.NewSessionID()
	manager := &HostManager{
		sessions: map[api.SessionID]*remoteSession{
			sessionID: {agentID: agentID, sessionID: sessionID, runtime: runtime, formatter: staticFormatter{prefix: ""}},
		},
	}
	payload := api.AgentMessage{
		ID:        api.NewRuntimeID(),
		From:      api.NewRuntimeID(),
		To:        api.TargetIDFromRuntime(agentID.RuntimeID),
		ToSlug:    "alpha",
		Content:   "hello",
		Timestamp: time.Now().UTC(),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	manager.handleCommMessage(protocolMessage("amux.comm.broadcast", data))
	buf := make([]byte, 128)
	if _, err := read.Read(buf); err != nil {
		t.Fatalf("read: %v", err)
	}
}

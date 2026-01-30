package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/pty"
	"github.com/agentflare-ai/amux/internal/session"
	"github.com/agentflare-ai/amux/pkg/api"
)

type formatterStub struct {
	prefix string
}

func (f formatterStub) Format(ctx context.Context, input string) (string, error) {
	_ = ctx
	return f.prefix + input, nil
}

func TestHandleCommMessageBroadcast(t *testing.T) {
	r1, w1, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	t.Cleanup(func() { _ = r1.Close(); _ = w1.Close() })
	r2, w2, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	t.Cleanup(func() { _ = r2.Close(); _ = w2.Close() })
	manager := &HostManager{
		sessions: map[api.SessionID]*remoteSession{},
	}
	runtime1 := &session.LocalSession{}
	runtime2 := &session.LocalSession{}
	setUnexportedField(runtime1, "ptyPair", &pty.Pair{Master: w1})
	setUnexportedField(runtime2, "ptyPair", &pty.Pair{Master: w2})
	session1 := &remoteSession{
		agentID:   api.NewAgentID(),
		runtime:   runtime1,
		formatter: formatterStub{prefix: ">>"},
	}
	session2 := &remoteSession{
		agentID:   api.NewAgentID(),
		runtime:   runtime2,
		formatter: formatterStub{prefix: ">>"},
	}
	manager.sessions[api.NewSessionID()] = session1
	manager.sessions[api.NewSessionID()] = session2
	msg := api.AgentMessage{
		ID:      api.NewRuntimeID(),
		From:    api.NewRuntimeID(),
		To:      api.TargetID{},
		ToSlug:  "broadcast",
		Content: "hello",
	}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	manager.handleCommMessage(protocol.Message{Data: data})
	buf1 := make([]byte, 32)
	n1, _ := r1.Read(buf1)
	buf2 := make([]byte, 32)
	n2, _ := r2.Read(buf2)
	if !bytes.Contains(buf1[:n1], []byte("hello")) || !bytes.Contains(buf2[:n2], []byte("hello")) {
		t.Fatalf("expected message delivered")
	}
}

func TestHandleOutboundMessageUnknownRecipient(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	t.Cleanup(func() { _ = reader.Close(); _ = writer.Close() })
	runtime := &session.LocalSession{}
	setUnexportedField(runtime, "ptyPair", &pty.Pair{Master: writer})
	manager := &HostManager{
		outbox:    NewOutbox(1024),
		connected: true,
		hostID:    api.MustParseHostID("host"),
		peerID:    api.NewPeerID(),
	}
	session := &remoteSession{
		agentID:   api.NewAgentID(),
		slug:      "alpha",
		runtime:   runtime,
		formatter: formatterStub{prefix: ""},
		matcher: adapterMatcher{matches: []adapter.PatternMatch{
			{Pattern: "message", Text: `{"to_slug":"unknown","content":"hi"}`},
		}},
	}
	manager.handleOutboundMessages(session, []byte("data"))
	buf := make([]byte, 128)
	n, _ := reader.Read(buf)
	if !bytes.Contains(buf[:n], []byte("unknown recipient")) {
		t.Fatalf("expected unknown recipient notice")
	}
}

type adapterMatcher struct {
	matches []adapter.PatternMatch
}

func (m adapterMatcher) Match(ctx context.Context, output []byte) ([]adapter.PatternMatch, error) {
	_ = ctx
	_ = output
	return m.matches, nil
}

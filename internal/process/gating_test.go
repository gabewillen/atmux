package process

import (
	"context"
	"io"
	"testing"

	"github.com/agentflare-ai/amux/internal/inference"
)

type mockEngine struct {
	response string
}

func (m *mockEngine) Generate(ctx context.Context, req inference.LiquidgenRequest) (inference.LiquidgenStream, error) {
	return &mockStream{response: m.response}, nil
}

type mockStream struct {
	response string
	sent     bool
}

func (s *mockStream) Next() (string, error) {
	if s.sent {
		return "", io.EOF
	}
	s.sent = true
	return s.response, nil
}

func (s *mockStream) Close() error { return nil }

func TestGater(t *testing.T) {
	g := &Gater{
		Engine: &mockEngine{response: "YES"},
	}
	
	if !g.ShouldNotify(context.Background(), Event{Type: "test"}) {
		t.Error("Gater should have returned true for YES")
	}
	
	g.Engine = &mockEngine{response: "NO"}
	if g.ShouldNotify(context.Background(), Event{Type: "test"}) {
		t.Error("Gater should have returned false for NO")
	}
}

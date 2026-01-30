package remote

import (
	"bytes"
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
)

type ptyDispatcher struct {
	mu         sync.Mutex
	subs       map[string]func(protocol.Message)
	published  map[string][][]byte
	maxPayload int
}

func newPTYDispatcher(maxPayload int) *ptyDispatcher {
	return &ptyDispatcher{
		subs:       make(map[string]func(protocol.Message)),
		published:  make(map[string][][]byte),
		maxPayload: maxPayload,
	}
}

func (d *ptyDispatcher) Publish(ctx context.Context, subject string, event protocol.Event) error {
	_ = ctx
	_ = subject
	_ = event
	return nil
}

func (d *ptyDispatcher) Subscribe(ctx context.Context, subject string, handler func(protocol.Event)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}

func (d *ptyDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	_ = ctx
	_ = reply
	d.mu.Lock()
	d.published[subject] = append(d.published[subject], append([]byte(nil), payload...))
	handler := d.subs[subject]
	d.mu.Unlock()
	if handler != nil {
		handler(protocol.Message{Subject: subject, Data: payload})
	}
	return nil
}

func (d *ptyDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(protocol.Message)) (protocol.Subscription, error) {
	_ = ctx
	d.mu.Lock()
	d.subs[subject] = handler
	d.mu.Unlock()
	return &ptySubscription{}, nil
}

func (d *ptyDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (protocol.Message, error) {
	_ = ctx
	_ = subject
	_ = payload
	_ = timeout
	return protocol.Message{}, nil
}

func (d *ptyDispatcher) MaxPayload() int {
	return d.maxPayload
}

func (d *ptyDispatcher) JetStream() nats.JetStreamContext {
	return nil
}

func (d *ptyDispatcher) Closed() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func (d *ptyDispatcher) send(subject string, data []byte) {
	d.mu.Lock()
	handler := d.subs[subject]
	d.mu.Unlock()
	if handler != nil {
		handler(protocol.Message{Subject: subject, Data: data})
	}
}

type ptySubscription struct{}

func (p *ptySubscription) Unsubscribe() error {
	return nil
}

func TestNewPTYConnRoundTrip(t *testing.T) {
	dispatcher := newPTYDispatcher(4)
	ctx := context.Background()
	hostID := api.HostID("host")
	sessionID := api.NewSessionID()
	conn, err := NewPTYConn(ctx, dispatcher, "prefix", hostID, sessionID)
	if err != nil {
		t.Fatalf("new pty conn: %v", err)
	}
	defer conn.Close()
	inSubject := PtyInSubject("prefix", hostID, sessionID)
	outSubject := PtyOutSubject("prefix", hostID, sessionID)

	payload := []byte("chunked")
	_ = conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	if _, err := conn.Write(payload); err != nil {
		t.Fatalf("write: %v", err)
	}
	waitForPublished(t, dispatcher, inSubject, 2)

	go dispatcher.send(outSubject, []byte("incoming"))
	readBuf := make([]byte, 16)
	_ = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	n, err := conn.Read(readBuf)
	if err != nil && err != io.EOF {
		t.Fatalf("read: %v", err)
	}
	if got := string(readBuf[:n]); got != "incoming" {
		t.Fatalf("unexpected read: %s", got)
	}
}

func TestChunkBytes(t *testing.T) {
	if got := chunkBytes(0, nil); got != nil {
		t.Fatalf("expected nil chunks for empty input")
	}
	data := []byte("abcdef")
	chunks := chunkBytes(3, data)
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	if !bytes.Equal(chunks[0], []byte("abc")) || !bytes.Equal(chunks[1], []byte("def")) {
		t.Fatalf("unexpected chunks: %#v", chunks)
	}
}

func waitForPublished(t *testing.T, dispatcher *ptyDispatcher, subject string, minCount int) {
	t.Helper()
	deadline := time.NewTimer(2 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer deadline.Stop()
	defer ticker.Stop()
	for {
		dispatcher.mu.Lock()
		count := len(dispatcher.published[subject])
		dispatcher.mu.Unlock()
		if count >= minCount {
			return
		}
		select {
		case <-ticker.C:
		case <-deadline.C:
			dispatcher.mu.Lock()
			count := len(dispatcher.published[subject])
			dispatcher.mu.Unlock()
			t.Fatalf("expected %d publishes on %s, got %d", minCount, subject, count)
		}
	}
}

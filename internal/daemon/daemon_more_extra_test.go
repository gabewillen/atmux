package daemon

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/rpc"
	"github.com/nats-io/nats.go"
)

type errListener struct {
	net.Listener
	closeErr error
}

func (l errListener) Close() error {
	if l.Listener != nil {
		_ = l.Listener.Close()
	}
	return l.closeErr
}

type closeErrDispatcher struct {
	err error
}

func (c closeErrDispatcher) Publish(ctx context.Context, subject string, event protocol.Event) error {
	_ = ctx
	_ = subject
	_ = event
	return nil
}
func (c closeErrDispatcher) Subscribe(ctx context.Context, subject string, handler func(protocol.Event)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}
func (c closeErrDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	_ = ctx
	_ = subject
	_ = payload
	_ = reply
	return nil
}
func (c closeErrDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(protocol.Message)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}
func (c closeErrDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (protocol.Message, error) {
	_ = ctx
	_ = subject
	_ = payload
	_ = timeout
	return protocol.Message{}, nil
}
func (c closeErrDispatcher) MaxPayload() int { return 0 }
func (c closeErrDispatcher) JetStream() nats.JetStreamContext { return nil }
func (c closeErrDispatcher) Closed() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}
func (c closeErrDispatcher) Close(ctx context.Context) error {
	_ = ctx
	return c.err
}

func TestDaemonServeMkdirError(t *testing.T) {
	tmp := t.TempDir()
	parent := filepath.Join(tmp, "file")
	if err := os.WriteFile(parent, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	daemon := &Daemon{
		cfg:    config.Config{Daemon: config.DaemonConfig{SocketPath: filepath.Join(parent, "amuxd.sock")}},
		server: rpc.NewServer(nil),
	}
	if err := daemon.Serve(context.Background()); err == nil || !strings.Contains(err.Error(), "daemon serve") {
		t.Fatalf("expected serve error, got %v", err)
	}
}

func TestDaemonCloseListenerAndDispatcherErrors(t *testing.T) {
	baseListener, err := net.Listen("unix", filepath.Join(t.TempDir(), "amuxd.sock"))
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	daemon := &Daemon{
		listener:   errListener{Listener: baseListener, closeErr: errors.New("close fail")},
		dispatcher: closeErrDispatcher{err: errors.New("dispatch close")},
	}
	if err := daemon.Close(context.Background(), false); err == nil {
		t.Fatalf("expected close error")
	}
}

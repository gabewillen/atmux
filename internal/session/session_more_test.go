package session

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/pkg/api"
)

func TestSessionMetaAndDone(t *testing.T) {
	meta := api.Session{}
	done := make(chan error)
	sess := &LocalSession{meta: meta, done: done}
	if sess.Meta() != meta {
		t.Fatalf("unexpected meta")
	}
	if sess.Done() == nil {
		t.Fatalf("expected done channel")
	}
	close(done)
}

func TestObserversAndRemoveOutput(t *testing.T) {
	sess := &LocalSession{outputs: make(map[uint64]net.Conn)}
	sess.AddOutputObserver(nil)
	called := false
	sess.AddOutputObserver(func([]byte) { called = true })
	sess.notifyObservers([]byte("hi"))
	if !called {
		t.Fatalf("expected observer called")
	}
	local, remote := net.Pipe()
	sess.outputs[1] = local
	sess.removeOutput(local)
	if len(sess.outputs) != 0 {
		t.Fatalf("expected output removed")
	}
	_ = remote.Close()
	_ = local.Close()
}

func TestWaitForExitErrors(t *testing.T) {
	sess := &LocalSession{done: make(chan error, 1), config: Config{DrainTimeout: time.Millisecond}}
	sess.done <- errors.New("exit")
	if err := sess.waitForExit(context.Background(), false); err == nil {
		t.Fatalf("expected wait error")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := sess.waitForExit(ctx, true); err == nil {
		t.Fatalf("expected context error")
	}
}

type errConn struct {
	net.Conn
}

func (e errConn) Write(p []byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestFanoutRemovesBadConn(t *testing.T) {
	local, remote := net.Pipe()
	sess := &LocalSession{outputs: map[uint64]net.Conn{1: errConn{Conn: local}}}
	sess.fanout([]byte("hi"))
	if len(sess.outputs) != 0 {
		t.Fatalf("expected output removed on write error")
	}
	_ = remote.Close()
	_ = local.Close()
}

func TestForwardInputRemovesOutput(t *testing.T) {
	local, remote := net.Pipe()
	sess := &LocalSession{outputs: map[uint64]net.Conn{1: local}}
	done := make(chan struct{})
	go func() {
		sess.forwardInput(local, 1)
		close(done)
	}()
	_, _ = remote.Write([]byte("ping"))
	_ = remote.Close()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("forwardInput did not exit")
	}
	if len(sess.outputs) != 0 {
		t.Fatalf("expected output removed after forwardInput")
	}
}

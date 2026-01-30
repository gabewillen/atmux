package session

import (
	"context"
	"testing"
	"time"
)

func TestSessionStopKillErrors(t *testing.T) {
	sess := &LocalSession{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := sess.Stop(ctx); err == nil {
		t.Fatalf("expected stop ctx error")
	}
	if err := sess.Kill(ctx); err == nil {
		t.Fatalf("expected kill ctx error")
	}
	if err := sess.Stop(context.Background()); err == nil {
		t.Fatalf("expected stop not running error")
	}
	if err := sess.Kill(context.Background()); err == nil {
		t.Fatalf("expected kill not running error")
	}
}

func TestSessionRestartError(t *testing.T) {
	sess := &LocalSession{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	t.Cleanup(cancel)
	if err := sess.Restart(ctx); err == nil {
		t.Fatalf("expected restart error")
	}
}

func TestSessionSendErrors(t *testing.T) {
	sess := &LocalSession{}
	if err := sess.Send(nil); err != nil {
		t.Fatalf("expected nil input ok")
	}
	if err := sess.Send([]byte("hi")); err == nil {
		t.Fatalf("expected send not running error")
	}
}

package pty

import (
	"context"
	"os/exec"
	"testing"
	"time"
)

func TestStartNilCommand(t *testing.T) {
	if _, err := Start(nil); err == nil {
		t.Fatalf("expected start error")
	}
}

func TestOpenAndClose(t *testing.T) {
	pair, err := Open()
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if pair.Master == nil || pair.Slave == nil {
		t.Fatalf("expected pair")
	}
	if err := pair.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func TestStartCommand(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", "echo ok")
	master, err := Start(cmd)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	_ = master.Close()
}

func TestCloseNilPair(t *testing.T) {
	var pair *Pair
	if err := pair.Close(); err != nil {
		t.Fatalf("expected nil close to succeed")
	}
}

func TestResize(t *testing.T) {
	pair, err := Open()
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = pair.Close() }()
	if err := Resize(pair.Master, 24, 80); err != nil {
		t.Fatalf("resize: %v", err)
	}
	if err := Resize(pair.Master, 0, 80); err == nil {
		t.Fatalf("expected invalid size error")
	}
}

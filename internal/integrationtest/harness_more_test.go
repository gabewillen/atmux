package integrationtest

import (
	"context"
	"testing"
)

func TestHarnessNilOperations(t *testing.T) {
	t.Parallel()
	if _, err := (*Harness)(nil).StartNATS(context.Background(), NATSContainerOptions{}); err == nil {
		t.Fatalf("expected StartNATS error for nil harness")
	}
	if _, err := (*Harness)(nil).StartToxiproxy(context.Background()); err == nil {
		t.Fatalf("expected StartToxiproxy error for nil harness")
	}
}

func TestNATSContainerNilOperations(t *testing.T) {
	t.Parallel()
	var n *NATSContainer
	if err := n.Stop(context.Background()); err != nil {
		t.Fatalf("expected nil stop")
	}
	if err := n.Start(context.Background()); err == nil {
		t.Fatalf("expected start error")
	}
	if err := n.WaitReady(context.Background(), 0); err == nil {
		t.Fatalf("expected wait ready error")
	}
}


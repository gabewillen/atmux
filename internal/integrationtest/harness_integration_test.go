package integrationtest

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestHarnessStartContainers(t *testing.T) {
	harness, err := NewHarness(t)
	if err != nil {
		if errors.Is(err, ErrDockerUnavailable) {
			t.Skipf("docker unavailable: %v", err)
		}
		t.Fatalf("new harness: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)
	natsContainer, err := harness.StartNATS(ctx, NATSContainerOptions{})
	if err != nil {
		t.Fatalf("start nats: %v", err)
	}
	if natsContainer.URL == "" {
		t.Fatalf("expected nats url")
	}
	if err := natsContainer.WaitReady(ctx, 5*time.Second); err != nil {
		t.Fatalf("wait ready: %v", err)
	}
	if err := natsContainer.Stop(ctx); err != nil {
		t.Fatalf("stop nats: %v", err)
	}
	if err := natsContainer.Start(ctx); err != nil {
		t.Fatalf("start nats: %v", err)
	}
	toxiproxy, err := harness.StartToxiproxy(ctx)
	if err != nil {
		t.Fatalf("start toxiproxy: %v", err)
	}
	if toxiproxy.Host == "" || toxiproxy.APIPort.Port() == "" || toxiproxy.ProxyPort.Port() == "" {
		t.Fatalf("expected toxiproxy endpoints")
	}
}

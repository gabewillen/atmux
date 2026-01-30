//go:build integration
// +build integration

package agent

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/integrationtest"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestIntegrationAgentLifecyclePresence(t *testing.T) {
	harness, err := integrationtest.NewHarness(t)
	if err != nil {
		t.Fatalf("harness: %v", err)
	}
	ctx := harness.Context()
	natsContainer, err := harness.StartNATS(ctx, integrationtest.NATSContainerOptions{})
	if err != nil {
		t.Fatalf("nats: %v", err)
	}
	dispatcher, err := protocol.NewNATSDispatcher(ctx, natsContainer.URL, protocol.NATSOptions{AllowNoJetStream: true})
	if err != nil {
		t.Fatalf("dispatcher: %v", err)
	}
	t.Cleanup(func() {
		if err := dispatcher.Close(context.Background()); err != nil {
			t.Fatalf("dispatcher close: %v", err)
		}
	})
	repoRoot := t.TempDir()
	worktree := filepath.Join(repoRoot, ".amux", "worktrees", "alpha")
	meta, err := api.NewAgent("alpha", "integration", api.AdapterRef("claude-code"), repoRoot, worktree, api.Location{Type: api.LocationLocal})
	if err != nil {
		t.Fatalf("agent meta: %v", err)
	}
	agent, err := NewAgent(meta, dispatcher)
	if err != nil {
		t.Fatalf("new agent: %v", err)
	}
	lifecycleEvents := make(chan protocol.Event, 4)
	presenceEvents := make(chan protocol.Event, 4)
	if _, err := dispatcher.Subscribe(ctx, protocol.Subject("events", "agent"), func(event protocol.Event) {
		select {
		case lifecycleEvents <- event:
		default:
		}
	}); err != nil {
		t.Fatalf("subscribe lifecycle: %v", err)
	}
	if _, err := dispatcher.Subscribe(ctx, protocol.Subject("events", "presence"), func(event protocol.Event) {
		select {
		case presenceEvents <- event:
		default:
		}
	}); err != nil {
		t.Fatalf("subscribe presence: %v", err)
	}
	agent.Start(ctx)
	waitForState(t, time.Second, agent.Lifecycle.State, "/agent.lifecycle/pending")
	waitForState(t, time.Second, agent.Presence.State, "/agent.presence/online")
	if err := agent.EmitLifecycle(ctx, EventStart, nil); err != nil {
		t.Fatalf("emit lifecycle start: %v", err)
	}
	waitForState(t, time.Second, agent.Lifecycle.State, "/agent.lifecycle/starting")
	if err := agent.EmitLifecycle(ctx, EventReady, nil); err != nil {
		t.Fatalf("emit lifecycle ready: %v", err)
	}
	waitForState(t, time.Second, agent.Lifecycle.State, "/agent.lifecycle/running")
	if err := agent.EmitPresence(ctx, EventTaskAssigned, nil); err != nil {
		t.Fatalf("emit presence assigned: %v", err)
	}
	waitForState(t, time.Second, agent.Presence.State, "/agent.presence/busy")
	if err := waitForEvent(lifecycleEvents, EventAgentStarted, time.Second); err != nil {
		t.Fatalf("lifecycle event: %v", err)
	}
	if err := waitForEvent(presenceEvents, EventPresenceChanged, time.Second); err != nil {
		t.Fatalf("presence event: %v", err)
	}
}

func waitForState(t *testing.T, timeout time.Duration, state func() string, want string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if state() == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for %s (last=%s)", want, state())
}

func waitForEvent(ch <-chan protocol.Event, name string, timeout time.Duration) error {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	for {
		select {
		case event := <-ch:
			if event.Name == name {
				return nil
			}
		case <-deadline.C:
			return fmt.Errorf("wait for event: %w", context.DeadlineExceeded)
		}
	}
}

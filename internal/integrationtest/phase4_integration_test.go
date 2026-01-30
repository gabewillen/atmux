//go:build integration
// +build integration

package integrationtest

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/manager"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

type phase4Registry struct {
	cmd []string
}

type phase4Adapter struct {
	name string
	cmd  []string
}

type phase4Matcher struct{}

type phase4Formatter struct{}

func (r *phase4Registry) Load(ctx context.Context, name string) (adapter.Adapter, error) {
	_ = ctx
	return &phase4Adapter{name: name, cmd: r.cmd}, nil
}

func (a *phase4Adapter) Name() string {
	return a.name
}

func (a *phase4Adapter) Manifest() adapter.Manifest {
	return adapter.Manifest{
		Name: a.name,
		Commands: adapter.AdapterCommands{
			Start: a.cmd,
		},
	}
}

func (a *phase4Adapter) Matcher() adapter.PatternMatcher {
	return phase4Matcher{}
}

func (a *phase4Adapter) Formatter() adapter.ActionFormatter {
	return phase4Formatter{}
}

func (a *phase4Adapter) OnEvent(ctx context.Context, event adapter.Event) ([]adapter.Action, error) {
	_ = ctx
	_ = event
	return nil, nil
}

func (phase4Matcher) Match(ctx context.Context, output []byte) ([]adapter.PatternMatch, error) {
	_ = ctx
	lines := strings.Split(string(output), "\n")
	matches := make([]adapter.PatternMatch, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "MSG:") {
			payload := strings.TrimSpace(strings.TrimPrefix(line, "MSG:"))
			if payload == "" {
				continue
			}
			matches = append(matches, adapter.PatternMatch{Pattern: "message", Text: payload})
		}
	}
	if len(matches) == 0 {
		return nil, nil
	}
	return matches, nil
}

func (phase4Formatter) Format(ctx context.Context, input string) (string, error) {
	_ = ctx
	return input, nil
}

func TestIntegrationPhase4PresenceRosterMessaging(t *testing.T) {
	harness, err := NewHarness(t)
	if err != nil {
		t.Fatalf("harness: %v", err)
	}
	ctx := harness.Context()
	natsContainer, err := harness.StartNATS(ctx, NATSContainerOptions{})
	if err != nil {
		t.Fatalf("nats: %v", err)
	}
	repoRoot := initPhase2Repo(t)
	resolver, err := paths.NewResolver(repoRoot)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	cfg := config.DefaultConfig(resolver)
	cfg.NATS.JetStreamDir = filepath.Join(t.TempDir(), "nats")
	dispatcher, err := protocol.NewNATSDispatcher(ctx, natsContainer.URL, protocol.NATSOptions{})
	if err != nil {
		t.Fatalf("dispatcher: %v", err)
	}
	t.Cleanup(func() {
		if err := dispatcher.Close(context.Background()); err != nil {
			t.Errorf("dispatcher close: %v", err)
		}
	})
	mgr, err := manager.NewManager(ctx, resolver, cfg, dispatcher, "test")
	if err != nil {
		t.Fatalf("manager: %v", err)
	}
	mgr.SetRegistryFactory(func(resolver *paths.Resolver) (adapter.Registry, error) {
		_ = resolver
		return &phase4Registry{cmd: []string{"env", "AMUX_PHASE4_HELPER=1", os.Args[0], "-test.run=TestIntegrationPhase4Helper"}}, nil
	})
	presenceEvents := make(chan protocol.Event, 64)
	presenceSub, err := dispatcher.Subscribe(ctx, protocol.Subject("events", "presence"), func(event protocol.Event) {
		select {
		case presenceEvents <- event:
		default:
		}
	})
	if err != nil {
		t.Fatalf("presence subscribe: %v", err)
	}
	t.Cleanup(func() {
		if err := presenceSub.Unsubscribe(); err != nil {
			t.Errorf("presence unsubscribe: %v", err)
		}
	})
	alpha, err := mgr.AddAgent(ctx, manager.AddRequest{
		Name:     "alpha",
		Adapter:  "stub",
		Location: api.Location{Type: api.LocationLocal},
		Cwd:      repoRoot,
	})
	if err != nil {
		t.Fatalf("add alpha: %v", err)
	}
	if alpha.AgentID == nil || alpha.AgentID.IsZero() {
		t.Fatalf("alpha id missing")
	}
	beta, err := mgr.AddAgent(ctx, manager.AddRequest{
		Name:     "beta",
		Adapter:  "stub",
		Location: api.Location{Type: api.LocationLocal},
		Cwd:      repoRoot,
	})
	if err != nil {
		t.Fatalf("add beta: %v", err)
	}
	if beta.AgentID == nil || beta.AgentID.IsZero() {
		t.Fatalf("beta id missing")
	}
	t.Cleanup(func() {
		if err := mgr.RemoveAgent(ctx, manager.RemoveRequest{AgentID: *alpha.AgentID}); err != nil {
			t.Errorf("remove alpha: %v", err)
		}
		if err := mgr.RemoveAgent(ctx, manager.RemoveRequest{AgentID: *beta.AgentID}); err != nil {
			t.Errorf("remove beta: %v", err)
		}
	})
	if err := waitForPresenceEvent(presenceEvents, *alpha.AgentID, agent.PresenceOnline, 5*time.Second); err != nil {
		t.Fatalf("alpha presence: %v", err)
	}
	if err := waitForPresenceEvent(presenceEvents, *beta.AgentID, agent.PresenceOnline, 5*time.Second); err != nil {
		t.Fatalf("beta presence: %v", err)
	}
	roster, err := mgr.ListAgents()
	if err != nil {
		t.Fatalf("roster: %v", err)
	}
	if err := assertRosterEntry(roster, *alpha.AgentID, "alpha", agent.PresenceOnline); err != nil {
		t.Fatalf("roster alpha: %v", err)
	}
	if err := assertRosterEntry(roster, *beta.AgentID, "beta", agent.PresenceOnline); err != nil {
		t.Fatalf("roster beta: %v", err)
	}
	presenceSubject := protocol.Subject("events", "agent", alpha.AgentID.String(), "presence")
	if err := dispatcher.Publish(ctx, presenceSubject, protocol.Event{
		Name:       agent.EventTaskAssigned,
		OccurredAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("presence publish: %v", err)
	}
	if err := waitForPresenceEvent(presenceEvents, *alpha.AgentID, agent.PresenceBusy, 5*time.Second); err != nil {
		t.Fatalf("alpha busy: %v", err)
	}
	if err := waitForRosterPresence(presenceEvents, *alpha.AgentID, agent.PresenceBusy, 5*time.Second); err != nil {
		t.Fatalf("roster busy: %v", err)
	}
	alphaConn, err := mgr.AttachAgent(*alpha.AgentID)
	if err != nil {
		t.Fatalf("attach alpha: %v", err)
	}
	t.Cleanup(func() {
		if err := alphaConn.Close(); err != nil {
			t.Errorf("alpha close: %v", err)
		}
	})
	betaConn, err := mgr.AttachAgent(*beta.AgentID)
	if err != nil {
		t.Fatalf("attach beta: %v", err)
	}
	t.Cleanup(func() {
		if err := betaConn.Close(); err != nil {
			t.Errorf("beta close: %v", err)
		}
	})
	alphaLines := make(chan string, 32)
	betaLines := make(chan string, 32)
	go streamLines(alphaConn, alphaLines)
	go streamLines(betaConn, betaLines)
	if _, err := alphaConn.Write([]byte("send:" + beta.Slug + "\n")); err != nil {
		t.Fatalf("send trigger: %v", err)
	}
	if err := waitForLinePrefix(betaLines, "echo:hello-beta", 5*time.Second); err != nil {
		t.Fatalf("message delivery: %v", err)
	}
}

func TestIntegrationPhase4Helper(t *testing.T) {
	if os.Getenv("AMUX_PHASE4_HELPER") != "1" {
		return
	}
	writer := bufio.NewWriter(os.Stdout)
	_, _ = writer.WriteString("ready\n")
	_ = writer.Flush()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "send:") {
			target := strings.TrimSpace(strings.TrimPrefix(line, "send:"))
			if target == "" {
				continue
			}
			payload := api.OutboundMessage{ToSlug: target, Content: "hello-beta\n"}
			encoded, err := json.Marshal(payload)
			if err != nil {
				continue
			}
			_, _ = writer.WriteString("MSG:" + string(encoded) + "\n")
			_ = writer.Flush()
			continue
		}
		_, _ = writer.WriteString("echo:" + line + "\n")
		_ = writer.Flush()
	}
}

func waitForPresenceEvent(events <-chan protocol.Event, id api.AgentID, presence string, timeout time.Duration) error {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	for {
		select {
		case event := <-events:
			if event.Name != agent.EventPresenceChanged {
				continue
			}
			var payload agent.PresenceEvent
			if err := decodeEventPayload(event.Payload, &payload); err != nil {
				continue
			}
			if payload.AgentID.Value() == id.Value() && strings.EqualFold(payload.Presence, presence) {
				return nil
			}
		case <-deadline.C:
			return fmt.Errorf("presence timeout: %s", presence)
		}
	}
}

func waitForRosterPresence(events <-chan protocol.Event, id api.AgentID, presence string, timeout time.Duration) error {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	for {
		select {
		case event := <-events:
			if event.Name != "roster.updated" {
				continue
			}
			var roster []api.RosterEntry
			if err := decodeEventPayload(event.Payload, &roster); err != nil {
				continue
			}
			for _, entry := range roster {
				if entry.AgentID == nil {
					continue
				}
				if entry.AgentID.Value() == id.Value() && strings.EqualFold(entry.Presence, presence) {
					return nil
				}
			}
		case <-deadline.C:
			return fmt.Errorf("roster timeout: %s", presence)
		}
	}
}

func assertRosterEntry(roster []api.RosterEntry, id api.AgentID, name string, presence string) error {
	for _, entry := range roster {
		if entry.AgentID == nil {
			continue
		}
		if entry.AgentID.Value() != id.Value() {
			continue
		}
		if entry.Name != name {
			return fmt.Errorf("roster name mismatch: %s", entry.Name)
		}
		if entry.Adapter == "" {
			return fmt.Errorf("roster adapter missing")
		}
		if entry.RepoRoot == "" {
			return fmt.Errorf("roster repo root missing")
		}
		if !strings.EqualFold(entry.Presence, presence) {
			return fmt.Errorf("roster presence mismatch: %s", entry.Presence)
		}
		return nil
	}
	return fmt.Errorf("roster entry missing")
}

func streamLines(conn net.Conn, out chan<- string) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		select {
		case out <- line:
		default:
		}
	}
	close(out)
}

func waitForLinePrefix(lines <-chan string, prefix string, timeout time.Duration) error {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	for {
		select {
		case line, ok := <-lines:
			if !ok {
				return fmt.Errorf("line stream closed")
			}
			if strings.HasPrefix(line, prefix) {
				return nil
			}
		case <-deadline.C:
			return fmt.Errorf("timeout waiting for %q", prefix)
		}
	}
}

func decodeEventPayload(payload any, dest any) error {
	if payload == nil {
		return fmt.Errorf("decode payload: empty")
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}
	return nil
}

package agent

import (
	"context"
	"testing"
	"time"

	"github.com/stateforward/hsm-go/muid"

	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/internal/ids"
	"github.com/agentflare-ai/amux/pkg/api"
)

func setupTestRosterWithParticipants(t *testing.T) (*Roster, muid.MUID, muid.MUID, muid.MUID) {
	t.Helper()
	r := NewRoster(event.NewNoopDispatcher())

	dirID := muid.Make()
	r.SetDirector(dirID, "Director", "")

	mgrID := muid.Make()
	r.AddManager(mgrID, "Manager", "host-1", "")

	agentID := muid.Make()
	r.AddAgent(&api.Agent{
		ID:       agentID,
		Name:     "Test Agent",
		Slug:     "test-agent",
		Adapter:  "test",
		RepoRoot: "/repo",
	}, api.LifecycleRunning, api.PresenceOnline)

	return r, dirID, mgrID, agentID
}

func TestMessageRouter_ResolveToSlug_Broadcast(t *testing.T) {
	r, _, _, _ := setupTestRosterWithParticipants(t)
	router := NewMessageRouter(r, event.NewNoopDispatcher(), muid.Make(), "")

	tests := []string{"all", "ALL", "broadcast", "BROADCAST", "*"}
	for _, slug := range tests {
		t.Run(slug, func(t *testing.T) {
			id, err := router.ResolveToSlug(slug)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != api.BroadcastID {
				t.Errorf("expected BroadcastID (0), got %v", id)
			}
		})
	}
}

func TestMessageRouter_ResolveToSlug_Director(t *testing.T) {
	r, dirID, _, _ := setupTestRosterWithParticipants(t)
	router := NewMessageRouter(r, event.NewNoopDispatcher(), muid.Make(), "")

	tests := []string{"director", "DIRECTOR", "Director"}
	for _, slug := range tests {
		t.Run(slug, func(t *testing.T) {
			id, err := router.ResolveToSlug(slug)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != dirID {
				t.Errorf("expected director ID %v, got %v", dirID, id)
			}
		})
	}
}

func TestMessageRouter_ResolveToSlug_Manager(t *testing.T) {
	r, _, mgrID, _ := setupTestRosterWithParticipants(t)
	router := NewMessageRouter(r, event.NewNoopDispatcher(), muid.Make(), "host-1")

	// Test manager@host_id
	id, err := router.ResolveToSlug("manager@host-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != mgrID {
		t.Errorf("expected manager ID %v, got %v", mgrID, id)
	}

	// Case-insensitive
	id, err = router.ResolveToSlug("MANAGER@HOST-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != mgrID {
		t.Errorf("expected manager ID %v, got %v", mgrID, id)
	}
}

func TestMessageRouter_ResolveToSlug_Agent(t *testing.T) {
	r, _, _, agentID := setupTestRosterWithParticipants(t)
	router := NewMessageRouter(r, event.NewNoopDispatcher(), muid.Make(), "")

	tests := []string{"test-agent", "TEST-AGENT", "Test-Agent"}
	for _, slug := range tests {
		t.Run(slug, func(t *testing.T) {
			id, err := router.ResolveToSlug(slug)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != agentID {
				t.Errorf("expected agent ID %v, got %v", agentID, id)
			}
		})
	}
}

func TestMessageRouter_ResolveToSlug_NotFound(t *testing.T) {
	r, _, _, _ := setupTestRosterWithParticipants(t)
	router := NewMessageRouter(r, event.NewNoopDispatcher(), muid.Make(), "")

	_, err := router.ResolveToSlug("nonexistent-agent")
	if err == nil {
		t.Error("expected error for nonexistent participant")
	}
}

func TestMessageRouter_RouteMessage(t *testing.T) {
	r, _, _, agentID := setupTestRosterWithParticipants(t)

	// Track dispatched events
	dispatched := make([]event.Event, 0)
	dispatcher := &trackingDispatcher{events: &dispatched}

	router := NewMessageRouter(r, dispatcher, muid.Make(), "")
	senderID := muid.Make()

	msg, err := router.RouteMessage(context.Background(), senderID, "test-agent", "Hello!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify message fields
	if msg.From != senderID {
		t.Errorf("expected From %v, got %v", senderID, msg.From)
	}
	if msg.To != agentID {
		t.Errorf("expected To %v, got %v", agentID, msg.To)
	}
	if msg.ToSlug != "test-agent" {
		t.Errorf("expected ToSlug %q, got %q", "test-agent", msg.ToSlug)
	}
	if msg.Content != "Hello!" {
		t.Errorf("expected Content %q, got %q", "Hello!", msg.Content)
	}
	if msg.ID == 0 {
		t.Error("expected non-zero message ID")
	}
	if msg.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}

	// Verify event was dispatched
	if len(dispatched) != 1 {
		t.Fatalf("expected 1 dispatched event, got %d", len(dispatched))
	}
	if dispatched[0].Type != event.TypeMessageOutbound {
		t.Errorf("expected event type %v, got %v", event.TypeMessageOutbound, dispatched[0].Type)
	}
}

func TestMessageRouter_RouteMessage_ResolutionFailed(t *testing.T) {
	r, _, _, _ := setupTestRosterWithParticipants(t)
	router := NewMessageRouter(r, event.NewNoopDispatcher(), muid.Make(), "")

	_, err := router.RouteMessage(context.Background(), muid.Make(), "nonexistent", "Hello!")
	if err == nil {
		t.Error("expected error for failed resolution")
	}
}

func TestMessageRouter_DeliverMessage(t *testing.T) {
	r, _, _, _ := setupTestRosterWithParticipants(t)

	dispatched := make([]event.Event, 0)
	dispatcher := &trackingDispatcher{events: &dispatched}

	router := NewMessageRouter(r, dispatcher, muid.Make(), "")

	msg := &api.AgentMessage{
		ID:        muid.Make(),
		From:      muid.Make(),
		To:        muid.Make(),
		ToSlug:    "test",
		Content:   "Hello!",
		Timestamp: time.Now().UTC(),
	}

	err := router.DeliverMessage(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(dispatched) != 1 {
		t.Fatalf("expected 1 dispatched event, got %d", len(dispatched))
	}
	if dispatched[0].Type != event.TypeMessageInbound {
		t.Errorf("expected event type %v, got %v", event.TypeMessageInbound, dispatched[0].Type)
	}
}

func TestMessageRouter_BroadcastMessage(t *testing.T) {
	r, _, _, _ := setupTestRosterWithParticipants(t)

	dispatched := make([]event.Event, 0)
	dispatcher := &trackingDispatcher{events: &dispatched}

	router := NewMessageRouter(r, dispatcher, muid.Make(), "")
	senderID := muid.Make()

	msg, err := router.BroadcastMessage(context.Background(), senderID, "Hello everyone!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msg.To != api.BroadcastID {
		t.Errorf("expected To = BroadcastID (0), got %v", msg.To)
	}
	if msg.ToSlug != "broadcast" {
		t.Errorf("expected ToSlug = %q, got %q", "broadcast", msg.ToSlug)
	}

	if len(dispatched) != 1 {
		t.Fatalf("expected 1 dispatched event, got %d", len(dispatched))
	}
	if dispatched[0].Type != event.TypeMessageBroadcast {
		t.Errorf("expected event type %v, got %v", event.TypeMessageBroadcast, dispatched[0].Type)
	}
}

func TestIsBroadcast(t *testing.T) {
	tests := []struct {
		name   string
		to     muid.MUID
		expect bool
	}{
		{"broadcast", api.BroadcastID, true},
		{"non-broadcast", muid.Make(), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := &api.AgentMessage{To: tc.to}
			if got := IsBroadcast(msg); got != tc.expect {
				t.Errorf("IsBroadcast() = %v, want %v", got, tc.expect)
			}
		})
	}
}

func TestMessageEnvelope_RoundTrip(t *testing.T) {
	original := &api.AgentMessage{
		ID:        muid.Make(),
		From:      muid.Make(),
		To:        muid.Make(),
		ToSlug:    "test-agent",
		Content:   "Test message content",
		Timestamp: time.Now().UTC().Truncate(time.Microsecond), // RFC3339Nano rounds to microseconds
	}

	env := ToEnvelope(original)

	// Verify envelope fields
	if env.ID != ids.EncodeID(original.ID) {
		t.Errorf("envelope ID mismatch")
	}
	if env.From != ids.EncodeID(original.From) {
		t.Errorf("envelope From mismatch")
	}
	if env.To != ids.EncodeID(original.To) {
		t.Errorf("envelope To mismatch")
	}
	if env.ToSlug != original.ToSlug {
		t.Errorf("envelope ToSlug mismatch")
	}
	if env.Content != original.Content {
		t.Errorf("envelope Content mismatch")
	}

	// Convert back
	decoded, err := FromEnvelope(env)
	if err != nil {
		t.Fatalf("FromEnvelope failed: %v", err)
	}

	// Verify round-trip
	if decoded.ID != original.ID {
		t.Errorf("ID round-trip failed: got %v, want %v", decoded.ID, original.ID)
	}
	if decoded.From != original.From {
		t.Errorf("From round-trip failed: got %v, want %v", decoded.From, original.From)
	}
	if decoded.To != original.To {
		t.Errorf("To round-trip failed: got %v, want %v", decoded.To, original.To)
	}
	if decoded.ToSlug != original.ToSlug {
		t.Errorf("ToSlug round-trip failed: got %q, want %q", decoded.ToSlug, original.ToSlug)
	}
	if decoded.Content != original.Content {
		t.Errorf("Content round-trip failed: got %q, want %q", decoded.Content, original.Content)
	}
	// Compare timestamps with nanosecond precision
	if !decoded.Timestamp.Equal(original.Timestamp) {
		t.Errorf("Timestamp round-trip failed: got %v, want %v", decoded.Timestamp, original.Timestamp)
	}
}

// trackingDispatcher records dispatched events for testing.
type trackingDispatcher struct {
	events *[]event.Event
}

func (d *trackingDispatcher) Dispatch(ctx context.Context, e event.Event) error {
	*d.events = append(*d.events, e)
	return nil
}

func (d *trackingDispatcher) Subscribe(sub event.Subscription) func() {
	return func() {}
}

func (d *trackingDispatcher) Close() error {
	return nil
}

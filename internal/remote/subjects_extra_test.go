package remote

import (
	"testing"

	"github.com/agentflare-ai/amux/pkg/api"
)

func TestSubjectHelpers(t *testing.T) {
	host := api.MustParseHostID("host")
	session := api.MustParseSessionID("42")
	prefix := "amux"
	if SubjectPrefix("...") != "amux" {
		t.Fatalf("expected default subject prefix")
	}
	if HandshakeSubject(prefix, host) == "" {
		t.Fatalf("expected handshake subject")
	}
	if ControlSubject(prefix, host) == "" {
		t.Fatalf("expected control subject")
	}
	if EventsSubject(prefix, host) == "" {
		t.Fatalf("expected events subject")
	}
	if DirectorCommSubject(prefix) == "" {
		t.Fatalf("expected director comm subject")
	}
	out := PtyOutSubject(prefix, host, session)
	in := PtyInSubject(prefix, host, session)
	if out == "" || in == "" {
		t.Fatalf("expected pty subjects")
	}
	if _, err := ParseHandshakeSubject(prefix, "bad.subject"); err == nil {
		t.Fatalf("expected parse handshake error")
	}
	handshake := HandshakeSubject(prefix, host)
	if _, err := ParseHandshakeSubject(prefix, handshake); err != nil {
		t.Fatalf("parse handshake: %v", err)
	}
	events := EventsSubject(prefix, host)
	if _, err := ParseEventsSubject(prefix, events); err != nil {
		t.Fatalf("parse events: %v", err)
	}
	if _, _, _, err := ParseSessionSubject(prefix, "bad"); err == nil {
		t.Fatalf("expected parse session error")
	}
	if _, _, _, err := ParseSessionSubject(prefix, out); err != nil {
		t.Fatalf("parse session: %v", err)
	}
}

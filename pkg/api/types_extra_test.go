package api

import (
	"encoding/json"
	"testing"
)

func TestRuntimeIDParsingAndJSON(t *testing.T) {
	if _, err := ParseRuntimeID(""); err == nil {
		t.Fatalf("expected empty id error")
	}
	if _, err := ParseRuntimeID("0"); err == nil {
		t.Fatalf("expected zero id error")
	}
	id := NewRuntimeID()
	raw := id.String()
	parsed, err := ParseRuntimeID(raw)
	if err != nil {
		t.Fatalf("parse runtime id: %v", err)
	}
	if parsed.Value() != id.Value() {
		t.Fatalf("expected parsed value to match")
	}
	data, err := json.Marshal(id)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded RuntimeID
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Value() != id.Value() {
		t.Fatalf("expected json round trip to match")
	}
}

func TestAgentSessionPeerParsing(t *testing.T) {
	if _, err := ParseAgentID(""); err == nil {
		t.Fatalf("expected agent id error")
	}
	if _, err := ParseSessionID("invalid"); err == nil {
		t.Fatalf("expected session id parse error")
	}
	if _, err := ParsePeerID(""); err == nil {
		t.Fatalf("expected peer id error")
	}
	agent := NewAgentID()
	parsedAgent, err := ParseAgentID(agent.String())
	if err != nil {
		t.Fatalf("parse agent: %v", err)
	}
	if parsedAgent.Value() != agent.Value() {
		t.Fatalf("expected agent id round trip")
	}
	session := NewSessionID()
	parsedSession, err := ParseSessionID(session.String())
	if err != nil {
		t.Fatalf("parse session: %v", err)
	}
	if parsedSession.Value() != session.Value() {
		t.Fatalf("expected session id round trip")
	}
	peer := NewPeerID()
	parsedPeer, err := ParsePeerID(peer.String())
	if err != nil {
		t.Fatalf("parse peer: %v", err)
	}
	if parsedPeer.Value() != peer.Value() {
		t.Fatalf("expected peer id round trip")
	}
}

func TestMustParseHelpers(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()
	_ = MustParseRuntimeID("0")
}

func TestMustParseAgentIDPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()
	_ = MustParseAgentID("")
}

func TestMustParseSessionIDPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()
	_ = MustParseSessionID("bad")
}

func TestMustParsePeerIDPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()
	_ = MustParsePeerID("")
}

func TestHostIDParsing(t *testing.T) {
	if _, err := ParseHostID(""); err == nil {
		t.Fatalf("expected host id error")
	}
	host := MustParseHostID("host")
	if host.String() != "host" {
		t.Fatalf("expected host string")
	}
}

package remote

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEncodeEventMessage(t *testing.T) {
	payload := map[string]any{"ok": true}
	msg, err := EncodeEventMessage("event.test", payload)
	if err != nil {
		t.Fatalf("encode event: %v", err)
	}
	if msg.Type != MsgBroadcast || msg.Event.Name != "event.test" {
		t.Fatalf("unexpected event message")
	}
	if len(msg.Event.Data) == 0 {
		t.Fatalf("expected data")
	}
	data, err := EncodeEventMessageJSON(msg)
	if err != nil {
		t.Fatalf("encode json: %v", err)
	}
	var decoded EventMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Event.Name != "event.test" {
		t.Fatalf("unexpected decoded event")
	}
}

func TestNowRFC3339(t *testing.T) {
	text := NowRFC3339()
	if text == "" {
		t.Fatalf("expected timestamp")
	}
	if _, err := time.Parse(time.RFC3339Nano, text); err != nil {
		t.Fatalf("parse timestamp: %v", err)
	}
}


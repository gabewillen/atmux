package remote

import (
	"encoding/json"
	"testing"
)

func TestControlMessageRoundtrip(t *testing.T) {
	cm := ControlMessage{
		Type:    ControlTypeHandshake,
		Payload: json.RawMessage(`{"protocol":1,"peer_id":"42","role":"daemon","host_id":"devbox"}`),
	}
	data, err := json.Marshal(cm)
	if err != nil {
		t.Fatal(err)
	}
	var out ControlMessage
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if out.Type != cm.Type {
		t.Errorf("Type = %q, want %q", out.Type, cm.Type)
	}
	if string(out.Payload) != string(cm.Payload) {
		t.Errorf("Payload = %s, want %s", out.Payload, cm.Payload)
	}
}

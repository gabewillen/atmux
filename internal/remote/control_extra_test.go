package remote

import (
	"encoding/json"
	"testing"
)

func TestControlEncodeDecode(t *testing.T) {
	msg := ControlMessage{Type: "ping", Payload: json.RawMessage(`{"ts":123}`)}
	data, err := EncodeControlMessage(msg)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	decoded, err := DecodeControlMessage(data)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.Type != "ping" {
		t.Fatalf("unexpected type")
	}
	if _, err := DecodeControlMessage([]byte("{}")); err == nil {
		t.Fatalf("expected invalid message error")
	}
	errMsg, err := NewErrorMessage("spawn", "invalid", "bad")
	if err != nil || errMsg.Type != "error" {
		t.Fatalf("expected error message")
	}
	encoded, err := EncodePayload("pong", PingPayload{UnixMS: 1})
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}
	var payload PingPayload
	if err := DecodePayload(encoded, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if err := DecodePayload(encoded, nil); err == nil {
		t.Fatalf("expected decode error")
	}
	if err := DecodePayload(ControlMessage{Type: "ping", Payload: json.RawMessage("bad")}, &payload); err == nil {
		t.Fatalf("expected bad payload error")
	}
}

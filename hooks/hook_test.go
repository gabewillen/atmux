package hooks

import "testing"

func TestInitUnavailable(t *testing.T) {
	if err := Init(); err != ErrHooksUnavailable {
		t.Fatalf("expected ErrHooksUnavailable, got %v", err)
	}
}

func TestPlaceholderTypes(t *testing.T) {
	msg := Message{Type: MessageType("hello"), Payload: []byte("payload")}
	if msg.Type != "hello" {
		t.Fatalf("unexpected message type: %s", msg.Type)
	}
	if string(msg.Payload) != "payload" {
		t.Fatalf("unexpected payload: %s", string(msg.Payload))
	}
	fd := FD{Value: 7}
	if fd.Value != 7 {
		t.Fatalf("unexpected fd value: %d", fd.Value)
	}
}

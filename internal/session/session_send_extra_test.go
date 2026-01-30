package session

import "testing"

func TestSendEmptyInput(t *testing.T) {
	sess := &LocalSession{}
	if err := sess.Send(nil); err != nil {
		t.Fatalf("expected nil for empty input")
	}
	if err := sess.Send([]byte{}); err != nil {
		t.Fatalf("expected nil for empty input")
	}
}

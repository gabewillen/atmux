package pty

import "testing"

func TestOpenPTY(t *testing.T) {
	pair, err := Open()
	if err != nil {
		t.Fatalf("open pty: %v", err)
	}
	if err := pair.Close(); err != nil {
		t.Fatalf("close pty: %v", err)
	}
}

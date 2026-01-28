package pty

import "testing"

func TestNewSession(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}
	defer session.Close()

	if session.Master() == nil {
		t.Fatal("Master() returned nil")
	}
	if session.Slave() == nil {
		t.Fatal("Slave() returned nil")
	}
}

func TestSetSize(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}
	defer session.Close()

	// Valid size
	err = session.SetSize(25, 80)
	if err != nil {
		t.Fatalf("SetSize(25, 80) failed: %v", err)
	}

	// Invalid size
	err = session.SetSize(0, 80)
	if err == nil {
		t.Fatal("SetSize(0, 80) should fail")
	}
}
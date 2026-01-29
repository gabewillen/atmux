package remote

import (
	"testing"
)

func TestSubjectPrefix(t *testing.T) {
	if got := SubjectPrefix(""); got != "amux" {
		t.Errorf("SubjectPrefix(\"\") = %q, want amux", got)
	}
	if got := SubjectPrefix("foo"); got != "foo" {
		t.Errorf("SubjectPrefix(\"foo\") = %q, want foo", got)
	}
}

func TestSubjectHandshake(t *testing.T) {
	got := SubjectHandshake("amux", "devbox")
	want := "amux.handshake.devbox"
	if got != want {
		t.Errorf("SubjectHandshake = %q, want %q", got, want)
	}
}

func TestSubjectCtl(t *testing.T) {
	got := SubjectCtl("amux", "host1")
	want := "amux.ctl.host1"
	if got != want {
		t.Errorf("SubjectCtl = %q, want %q", got, want)
	}
}

func TestSubjectPTYOut(t *testing.T) {
	got := SubjectPTYOut("amux", "host1", "9001")
	want := "amux.pty.host1.9001.out"
	if got != want {
		t.Errorf("SubjectPTYOut = %q, want %q", got, want)
	}
}

func TestSubjectPTYIn(t *testing.T) {
	got := SubjectPTYIn("amux", "host1", "9001")
	want := "amux.pty.host1.9001.in"
	if got != want {
		t.Errorf("SubjectPTYIn = %q, want %q", got, want)
	}
}

func TestKVKeyHostInfo(t *testing.T) {
	got := KVKeyHostInfo("devbox")
	want := "hosts/devbox/info"
	if got != want {
		t.Errorf("KVKeyHostInfo = %q, want %q", got, want)
	}
}

func TestKVKeySession(t *testing.T) {
	got := KVKeySession("host1", "9001")
	want := "sessions/host1/9001"
	if got != want {
		t.Errorf("KVKeySession = %q, want %q", got, want)
	}
}

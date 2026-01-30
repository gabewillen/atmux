package git

import "testing"

func TestIsMissingRef(t *testing.T) {
	if !isMissingRef(1, []byte("not a valid ref")) {
		t.Fatalf("expected missing ref")
	}
	if !isMissingRef(128, []byte("unknown revision")) {
		t.Fatalf("expected missing ref")
	}
	if isMissingRef(2, []byte("not a valid ref")) {
		t.Fatalf("expected false for exit code")
	}
}


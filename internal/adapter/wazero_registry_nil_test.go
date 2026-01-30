package adapter

import (
	"context"
	"testing"
)

func TestWasmAdapterNil(t *testing.T) {
	var adapter *wasmAdapter
	if _, err := adapter.OnEvent(context.Background(), Event{}); err == nil {
		t.Fatalf("expected nil adapter error")
	}
	var matcher *wasmMatcher
	if _, err := matcher.Match(context.Background(), []byte("out")); err == nil {
		t.Fatalf("expected matcher error")
	}
	var formatter *wasmFormatter
	if _, err := formatter.Format(context.Background(), "input"); err == nil {
		t.Fatalf("expected formatter error")
	}
}


package rpc

import (
	"context"
	"testing"
)

func TestDialEmptyPath(t *testing.T) {
	if _, err := Dial(context.Background(), ""); err == nil {
		t.Fatalf("expected dial error")
	}
}

func TestClientNil(t *testing.T) {
	var client *Client
	if err := client.Close(); err != nil {
		t.Fatalf("expected nil close")
	}
	if err := client.Call(context.Background(), "method", nil, nil); err == nil {
		t.Fatalf("expected call error")
	}
}

func TestCallMarshalError(t *testing.T) {
	client := &Client{}
	if err := client.Call(context.Background(), "method", func() {}, nil); err == nil {
		t.Fatalf("expected marshal error")
	}
}


package integrationtest

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestToxiproxyContainerHelpersNil(t *testing.T) {
	t.Parallel()
	var container *ToxiproxyContainer
	if container.APIURL() != "" {
		t.Fatalf("expected empty API URL")
	}
	if container.ProxyAddress() != "" {
		t.Fatalf("expected empty proxy address")
	}
	client := NewToxiproxyClient("http://127.0.0.1:1")
	if client.BaseURL == "" || client.Client == nil {
		t.Fatalf("expected client defaults")
	}
	if _, err := (*ToxiproxyClient)(nil).doJSON(context.Background(), http.MethodGet, "/", nil); err == nil {
		t.Fatalf("expected nil client error")
	}
}

func TestToxiproxyClientErrors(t *testing.T) {
	t.Parallel()
	client := &ToxiproxyClient{
		BaseURL: "http://127.0.0.1:1",
		Client:  &http.Client{Timeout: 50 * time.Millisecond},
	}
	if err := client.CreateProxy(context.Background(), "p", "127.0.0.1:1", "127.0.0.1:2"); err == nil {
		t.Fatalf("expected create proxy error")
	}
	if err := client.SetProxyEnabled(context.Background(), "p", true); err == nil {
		t.Fatalf("expected set enabled error")
	}
	if err := client.AddLatency(context.Background(), "p", 10*time.Millisecond, 0); err == nil {
		t.Fatalf("expected add latency error")
	}
	if err := client.AddTimeout(context.Background(), "p", 10*time.Millisecond); err == nil {
		t.Fatalf("expected add timeout error")
	}
	if err := client.RemoveToxic(context.Background(), "p", "latency"); err == nil {
		t.Fatalf("expected remove toxic error")
	}
}

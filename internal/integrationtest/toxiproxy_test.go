package integrationtest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestToxiproxyClientDoJSON(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("/proxies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	handler.HandleFunc("/proxies/bad", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad"))
	})
	server := httptest.NewServer(handler)
	defer server.Close()
	client := NewToxiproxyClient(server.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.CreateProxy(ctx, "proxy", "127.0.0.1:1", "127.0.0.1:2"); err != nil {
		t.Fatalf("create proxy: %v", err)
	}
	if err := client.SetProxyEnabled(ctx, "bad", false); err == nil || !strings.Contains(err.Error(), "status 400") {
		t.Fatalf("expected status error, got %v", err)
	}
	if _, err := (*ToxiproxyClient)(nil).doJSON(ctx, http.MethodGet, "/", nil); err == nil {
		t.Fatalf("expected nil client error")
	}
}

func TestToxiproxyContainerHelpers(t *testing.T) {
	var container *ToxiproxyContainer
	if container.APIURL() != "" {
		t.Fatalf("expected empty api url")
	}
	if container.ProxyAddress() != "" {
		t.Fatalf("expected empty proxy address")
	}
	if container.Client() == nil {
		t.Fatalf("expected client")
	}
}

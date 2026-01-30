package integrationtest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ToxiproxyClient configures network faults via the toxiproxy API.
type ToxiproxyClient struct {
	BaseURL string
	Client  *http.Client
}

// NewToxiproxyClient constructs a ToxiproxyClient for the given base URL.
func NewToxiproxyClient(baseURL string) *ToxiproxyClient {
	client := &http.Client{Timeout: 5 * time.Second}
	return &ToxiproxyClient{BaseURL: baseURL, Client: client}
}

// APIURL returns the toxiproxy API base URL.
func (t *ToxiproxyContainer) APIURL() string {
	if t == nil {
		return ""
	}
	return fmt.Sprintf("http://%s:%s", t.Host, t.APIPort.Port())
}

// ProxyAddress returns the host:port address for proxy traffic.
func (t *ToxiproxyContainer) ProxyAddress() string {
	if t == nil {
		return ""
	}
	return fmt.Sprintf("%s:%s", t.Host, t.ProxyPort.Port())
}

// Client returns a toxiproxy API client for this container.
func (t *ToxiproxyContainer) Client() *ToxiproxyClient {
	return NewToxiproxyClient(t.APIURL())
}

// CreateProxy registers a new proxy.
func (c *ToxiproxyClient) CreateProxy(ctx context.Context, name string, listen string, upstream string) error {
	payload := map[string]any{
		"name":     name,
		"listen":   listen,
		"upstream": upstream,
		"enabled":  true,
	}
	_, err := c.doJSON(ctx, http.MethodPost, "/proxies", payload)
	if err != nil {
		return fmt.Errorf("toxiproxy create proxy: %w", err)
	}
	return nil
}

// SetProxyEnabled toggles a proxy on or off.
func (c *ToxiproxyClient) SetProxyEnabled(ctx context.Context, name string, enabled bool) error {
	payload := map[string]any{
		"enabled": enabled,
	}
	_, err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/proxies/%s", name), payload)
	if err != nil {
		return fmt.Errorf("toxiproxy set enabled: %w", err)
	}
	return nil
}

// AddLatency adds a latency toxic in milliseconds.
func (c *ToxiproxyClient) AddLatency(ctx context.Context, name string, latency time.Duration, jitter time.Duration) error {
	payload := map[string]any{
		"type":     "latency",
		"name":     "latency",
		"stream":   "downstream",
		"toxicity": 1.0,
		"attributes": map[string]any{
			"latency": int(latency / time.Millisecond),
			"jitter":  int(jitter / time.Millisecond),
		},
	}
	_, err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/proxies/%s/toxics", name), payload)
	if err != nil {
		return fmt.Errorf("toxiproxy add latency: %w", err)
	}
	return nil
}

// AddTimeout adds a timeout toxic (useful to simulate loss).
func (c *ToxiproxyClient) AddTimeout(ctx context.Context, name string, timeout time.Duration) error {
	payload := map[string]any{
		"type":     "timeout",
		"name":     "timeout",
		"stream":   "downstream",
		"toxicity": 1.0,
		"attributes": map[string]any{
			"timeout": int(timeout / time.Millisecond),
		},
	}
	_, err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/proxies/%s/toxics", name), payload)
	if err != nil {
		return fmt.Errorf("toxiproxy add timeout: %w", err)
	}
	return nil
}

func (c *ToxiproxyClient) doJSON(ctx context.Context, method string, path string, payload any) ([]byte, error) {
	if c == nil {
		return nil, fmt.Errorf("toxiproxy client is nil")
	}
	if c.Client == nil {
		c.Client = &http.Client{Timeout: 5 * time.Second}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("toxiproxy encode payload: %w", err)
	}
	url := c.BaseURL + path
	request, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("toxiproxy request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := c.Client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("toxiproxy request: %w", err)
	}
	payloadBytes, readErr := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	closeErr := response.Body.Close()
	if readErr != nil {
		return nil, fmt.Errorf("toxiproxy response: %w", readErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("toxiproxy response: %w", closeErr)
	}
	if response.StatusCode >= 300 {
		return nil, fmt.Errorf("toxiproxy response: status %d: %s", response.StatusCode, string(payloadBytes))
	}
	return payloadBytes, nil
}

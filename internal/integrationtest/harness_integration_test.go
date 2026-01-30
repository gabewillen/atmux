//go:build integration
// +build integration

package integrationtest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/protocol"
)

func TestIntegrationHarnessFaultInjection(t *testing.T) {
	harness, err := NewHarness(t)
	if err != nil {
		t.Fatalf("harness: %v", err)
	}
	ctx := harness.Context()
	natsContainer, err := harness.StartNATS(ctx)
	if err != nil {
		t.Fatalf("nats: %v", err)
	}
	toxiproxy, err := harness.StartToxiproxy(ctx)
	if err != nil {
		t.Fatalf("toxiproxy: %v", err)
	}
	client := toxiproxy.Client()
	proxyName := "nats-proxy"
	upstream := fmt.Sprintf("%s:4222", natsContainer.Alias)
	if err := client.CreateProxy(ctx, proxyName, "0.0.0.0:8666", upstream); err != nil {
		t.Fatalf("create proxy: %v", err)
	}
	if err := client.SetProxyEnabled(ctx, proxyName, false); err != nil {
		t.Fatalf("disable proxy: %v", err)
	}
	blockedCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	if _, err := protocol.NewNATSDispatcher(blockedCtx, "nats://"+toxiproxy.ProxyAddress(), protocol.NATSOptions{AllowNoJetStream: true}); err == nil {
		cancel()
		t.Fatalf("expected connection failure when proxy disabled")
	}
	cancel()
	if err := client.SetProxyEnabled(ctx, proxyName, true); err != nil {
		t.Fatalf("enable proxy: %v", err)
	}
	dispatcher, err := waitForDispatcher(ctx, "nats://"+toxiproxy.ProxyAddress(), protocol.NATSOptions{AllowNoJetStream: true}, 10*time.Second)
	if err != nil {
		t.Fatalf("dispatcher: %v", err)
	}
	if err := dispatcher.Close(context.Background()); err != nil {
		t.Fatalf("dispatcher close: %v", err)
	}
	if err := client.AddLatency(ctx, proxyName, 200*time.Millisecond, 50*time.Millisecond); err != nil {
		t.Fatalf("add latency: %v", err)
	}
	if err := client.AddTimeout(ctx, proxyName, 150*time.Millisecond); err != nil {
		t.Fatalf("add timeout: %v", err)
	}
}

func TestIntegrationHarnessContainerRestart(t *testing.T) {
	harness, err := NewHarness(t)
	if err != nil {
		t.Fatalf("harness: %v", err)
	}
	ctx := harness.Context()
	natsContainer, err := harness.StartNATS(ctx)
	if err != nil {
		t.Fatalf("nats: %v", err)
	}
	dispatcher, err := waitForDispatcher(ctx, natsContainer.URL, protocol.NATSOptions{AllowNoJetStream: true}, 10*time.Second)
	if err != nil {
		t.Fatalf("dispatcher: %v", err)
	}
	if err := dispatcher.Close(context.Background()); err != nil {
		t.Fatalf("dispatcher close: %v", err)
	}
	if err := natsContainer.Stop(ctx); err != nil {
		t.Fatalf("stop nats: %v", err)
	}
	blockedCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	if _, err := protocol.NewNATSDispatcher(blockedCtx, natsContainer.URL, protocol.NATSOptions{AllowNoJetStream: true}); err == nil {
		cancel()
		t.Fatalf("expected connection failure while nats stopped")
	}
	cancel()
	if err := natsContainer.Start(ctx); err != nil {
		t.Fatalf("start nats: %v", err)
	}
	if err := natsContainer.WaitReady(ctx, 30*time.Second); err != nil {
		t.Fatalf("wait nats: %v", err)
	}
	dispatcher, err = waitForDispatcher(ctx, natsContainer.URL, protocol.NATSOptions{AllowNoJetStream: true}, 10*time.Second)
	if err != nil {
		t.Fatalf("dispatcher after restart: %v", err)
	}
	if err := dispatcher.Close(context.Background()); err != nil {
		t.Fatalf("dispatcher close: %v", err)
	}
}

func waitForDispatcher(ctx context.Context, url string, options protocol.NATSOptions, timeout time.Duration) (*protocol.NATSDispatcher, error) {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for {
		readyCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		dispatcher, err := protocol.NewNATSDispatcher(readyCtx, url, options)
		cancel()
		if err == nil {
			return dispatcher, nil
		}
		lastErr = err
		if time.Now().After(deadline) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	return nil, fmt.Errorf("dispatcher: %w", lastErr)
}

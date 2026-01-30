package integrationtest

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	natsImage           = "nats:2.12.4"
	toxiproxyImage      = "ghcr.io/shopify/toxiproxy:2.5.0"
	natsPort            = "4222/tcp"
	toxiproxyAPIPort    = "8474/tcp"
	toxiproxyProxyPort  = "8666/tcp"
	defaultHarnessGrace = 30 * time.Second
)

// NATSContainerOptions controls how the NATS container is started.
type NATSContainerOptions struct {
	// ExposedPorts overrides the default exposed ports.
	ExposedPorts []string
	// Cmd overrides the default NATS server arguments.
	Cmd []string
	// Files copies host files into the container before startup.
	Files []testcontainers.ContainerFile
}

// ErrDockerUnavailable indicates docker is not reachable for integration tests.
var ErrDockerUnavailable = errors.New("docker unavailable")

// Harness manages docker/testcontainers infrastructure for integration tests.
type Harness struct {
	ctx        context.Context
	cancel     context.CancelFunc
	network    *testcontainers.DockerNetwork
	containers []testcontainers.Container
}

// NewHarness creates a harness with an isolated docker network and cleanup.
func NewHarness(t testing.TB) (*Harness, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	if deadlineProvider, ok := any(t).(interface{ Deadline() (time.Time, bool) }); ok {
		if deadline, ok := deadlineProvider.Deadline(); ok {
			cancel()
			ctx, cancel = context.WithDeadline(context.Background(), deadline)
		}
	}
	if os.Getenv("TESTCONTAINERS_RYUK_DISABLED") == "" {
		if err := os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true"); err != nil {
			cancel()
			return nil, fmt.Errorf("integration harness: %w", err)
		}
	}
	network, err := safeNetworkNew(ctx)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("integration harness: create network: %w", err)
	}
	h := &Harness{
		ctx:     ctx,
		cancel:  cancel,
		network: network,
	}
	t.Cleanup(func() {
		if err := h.Close(); err != nil {
			t.Logf("integration harness cleanup: %v", err)
		}
	})
	return h, nil
}

func safeNetworkNew(ctx context.Context) (dockernet *testcontainers.DockerNetwork, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("integration harness: %w: %v", ErrDockerUnavailable, recovered)
		}
	}()
	dockernet, err = network.New(ctx, network.WithAttachable())
	if err != nil {
		return nil, err
	}
	return dockernet, nil
}

// Context returns the harness context.
func (h *Harness) Context() context.Context {
	if h == nil {
		return context.Background()
	}
	return h.ctx
}

// Close terminates all containers and removes the network.
func (h *Harness) Close() error {
	if h == nil {
		return nil
	}
	cleanupCtx, cancel := context.WithTimeout(context.Background(), defaultHarnessGrace)
	defer cancel()
	var errOut error
	for i := len(h.containers) - 1; i >= 0; i-- {
		container := h.containers[i]
		if container == nil {
			continue
		}
		if err := container.Terminate(cleanupCtx); err != nil {
			errOut = errors.Join(errOut, fmt.Errorf("integration harness: terminate container: %w", err))
		}
	}
	if h.network != nil {
		if err := h.network.Remove(cleanupCtx); err != nil {
			errOut = errors.Join(errOut, fmt.Errorf("integration harness: remove network: %w", err))
		}
	}
	h.cancel()
	return errOut
}

// NATSContainer tracks a NATS container instance.
type NATSContainer struct {
	Container testcontainers.Container
	Host      string
	Port      nat.Port
	URL       string
	Alias     string
}

// StartNATS launches a NATS container with JetStream enabled.
func (h *Harness) StartNATS(ctx context.Context, opts NATSContainerOptions) (*NATSContainer, error) {
	if h == nil {
		return nil, fmt.Errorf("integration harness: nil")
	}
	ctx = h.contextOrDefault(ctx)
	alias := "nats"
	exposedPorts := opts.ExposedPorts
	if len(exposedPorts) == 0 {
		exposedPorts = []string{natsPort}
	}
	cmd := opts.Cmd
	if len(cmd) == 0 {
		cmd = []string{"-js", "--store_dir", "/data/jetstream"}
	}
	req := testcontainers.ContainerRequest{
		Image:        natsImage,
		ExposedPorts: exposedPorts,
		Cmd:          cmd,
		WaitingFor:   wait.ForListeningPort(natsPort).WithStartupTimeout(60 * time.Second),
		Files:        opts.Files,
	}
	if h.network != nil {
		req.Networks = []string{h.network.Name}
		req.NetworkAliases = map[string][]string{h.network.Name: {alias}}
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("integration harness: start nats: %w", err)
	}
	h.containers = append(h.containers, container)
	natsContainer := &NATSContainer{
		Container: container,
		Alias:     alias,
	}
	if err := natsContainer.refreshEndpoint(ctx); err != nil {
		return nil, err
	}
	return natsContainer, nil
}

// Stop halts the NATS container.
func (n *NATSContainer) Stop(ctx context.Context) error {
	if n == nil || n.Container == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	timeout := 20 * time.Second
	if err := n.Container.Stop(ctx, &timeout); err != nil {
		return fmt.Errorf("integration harness: stop nats: %w", err)
	}
	return nil
}

// Start restarts the NATS container.
func (n *NATSContainer) Start(ctx context.Context) error {
	if n == nil || n.Container == nil {
		return fmt.Errorf("integration harness: nats container is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := n.Container.Start(ctx); err != nil {
		return fmt.Errorf("integration harness: start nats: %w", err)
	}
	if err := n.refreshEndpoint(ctx); err != nil {
		return err
	}
	return nil
}

// WaitReady waits until the NATS port is reachable.
func (n *NATSContainer) WaitReady(ctx context.Context, timeout time.Duration) error {
	if n == nil {
		return fmt.Errorf("integration harness: nats container is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	deadline := time.Now().Add(timeout)
	addr := net.JoinHostPort(n.Host, n.Port.Port())
	for {
		conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
		if err == nil {
			if err := conn.Close(); err != nil {
				return fmt.Errorf("integration harness: wait nats ready: %w", err)
			}
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("integration harness: wait nats ready: %w", ctx.Err())
		default:
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("integration harness: wait nats ready: timeout")
		}
		timer := time.NewTimer(200 * time.Millisecond)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("integration harness: wait nats ready: %w", ctx.Err())
		}
	}
}

func (n *NATSContainer) refreshEndpoint(ctx context.Context) error {
	host, err := n.Container.Host(ctx)
	if err != nil {
		return fmt.Errorf("integration harness: nats host: %w", err)
	}
	port, err := n.Container.MappedPort(ctx, nat.Port(natsPort))
	if err != nil {
		return fmt.Errorf("integration harness: nats port: %w", err)
	}
	n.Host = host
	n.Port = port
	n.URL = fmt.Sprintf("nats://%s:%s", host, port.Port())
	return nil
}

// ToxiproxyContainer tracks a toxiproxy container instance.
type ToxiproxyContainer struct {
	Container testcontainers.Container
	Host      string
	APIPort   nat.Port
	ProxyPort nat.Port
}

// StartToxiproxy launches a toxiproxy container for network fault injection.
func (h *Harness) StartToxiproxy(ctx context.Context) (*ToxiproxyContainer, error) {
	if h == nil {
		return nil, fmt.Errorf("integration harness: nil")
	}
	ctx = h.contextOrDefault(ctx)
	alias := "toxiproxy"
	req := testcontainers.ContainerRequest{
		Image:        toxiproxyImage,
		ExposedPorts: []string{toxiproxyAPIPort, toxiproxyProxyPort, "8667/tcp"},
		WaitingFor:   wait.ForListeningPort(toxiproxyAPIPort).WithStartupTimeout(60 * time.Second),
	}
	if h.network != nil {
		req.Networks = []string{h.network.Name}
		req.NetworkAliases = map[string][]string{h.network.Name: {alias}}
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("integration harness: start toxiproxy: %w", err)
	}
	h.containers = append(h.containers, container)
	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("integration harness: toxiproxy host: %w", err)
	}
	apiPort, err := container.MappedPort(ctx, nat.Port(toxiproxyAPIPort))
	if err != nil {
		return nil, fmt.Errorf("integration harness: toxiproxy api port: %w", err)
	}
	proxyPort, err := container.MappedPort(ctx, nat.Port(toxiproxyProxyPort))
	if err != nil {
		return nil, fmt.Errorf("integration harness: toxiproxy proxy port: %w", err)
	}
	return &ToxiproxyContainer{
		Container: container,
		Host:      host,
		APIPort:   apiPort,
		ProxyPort: proxyPort,
	}, nil
}

func (h *Harness) contextOrDefault(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	if h == nil {
		return context.Background()
	}
	return h.ctx
}

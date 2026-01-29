package daemon

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/manager"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/rpc"
)

const (
	// SpecVersion is the spec version implemented by the daemon.
	SpecVersion = "v1.22"
	// AmuxVersion is the daemon version string.
	AmuxVersion = "dev"
)

// Daemon hosts the JSON-RPC control plane.
type Daemon struct {
	resolver   *paths.Resolver
	cfg        config.Config
	manager    *manager.LocalManager
	dispatcher protocol.Dispatcher
	server     *rpc.Server
	listener   net.Listener
	embedded   *protocol.EmbeddedServer
	logger     *log.Logger
	closeMu    sync.Mutex
	closed     bool
}

// New constructs a daemon instance.
func New(ctx context.Context, resolver *paths.Resolver, cfg config.Config, logger *log.Logger) (*Daemon, error) {
	if resolver == nil {
		return nil, fmt.Errorf("daemon: resolver is required")
	}
	if logger == nil {
		logger = log.New(os.Stderr, "amuxd ", log.LstdFlags)
	}
	var embedded *protocol.EmbeddedServer
	var dispatcher protocol.Dispatcher
	mode := strings.TrimSpace(cfg.NATS.Mode)
	if mode == "" || mode == "embedded" {
		addr := cfg.NATS.Listen
		if strings.TrimSpace(addr) == "" {
			addr = "127.0.0.1:0"
		}
		server, err := protocol.StartEmbeddedServer(ctx, addr)
		if err != nil {
			return nil, fmt.Errorf("daemon: %w", err)
		}
		embedded = server
		dispatcher, err = protocol.NewNATSDispatcher(ctx, server.URL())
		if err != nil {
			_ = server.Close()
			return nil, fmt.Errorf("daemon: %w", err)
		}
	} else {
		url := cfg.Remote.NATS.URL
		if strings.TrimSpace(url) == "" {
			url = cfg.NATS.HubURL
		}
		d, err := protocol.NewNATSDispatcher(ctx, url)
		if err != nil {
			return nil, fmt.Errorf("daemon: %w", err)
		}
		dispatcher = d
	}
	mgr, err := manager.NewLocalManager(ctx, resolver, cfg, dispatcher)
	if err != nil {
		if embedded != nil {
			_ = embedded.Close()
		}
		return nil, fmt.Errorf("daemon: %w", err)
	}
	daemon := &Daemon{
		resolver:   resolver,
		cfg:        cfg,
		manager:    mgr,
		dispatcher: dispatcher,
		embedded:   embedded,
		logger:     logger,
		server:     rpc.NewServer(logger),
	}
	daemon.registerHandlers()
	return daemon, nil
}

// Serve starts listening on the daemon socket.
func (d *Daemon) Serve(ctx context.Context) error {
	if d == nil {
		return fmt.Errorf("daemon serve: daemon is nil")
	}
	socketPath := d.cfg.Daemon.SocketPath
	if socketPath == "" {
		return fmt.Errorf("daemon serve: socket path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(socketPath), 0o755); err != nil {
		return fmt.Errorf("daemon serve: %w", err)
	}
	if err := os.RemoveAll(socketPath); err != nil {
		return fmt.Errorf("daemon serve: %w", err)
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("daemon serve: %w", err)
	}
	if err := os.Chmod(socketPath, 0o600); err != nil {
		_ = listener.Close()
		return fmt.Errorf("daemon serve: %w", err)
	}
	d.listener = listener
	stop := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = listener.Close()
		case <-stop:
		}
	}()
	err = d.server.Serve(ctx, listener)
	close(stop)
	if ctx.Err() != nil {
		return nil
	}
	if err != nil {
		return err
	}
	return nil
}

// Close shuts down the daemon, optionally forcing termination.
func (d *Daemon) Close(ctx context.Context, force bool) error {
	if d == nil {
		return nil
	}
	var errOut error
	if d.manager != nil {
		if err := d.manager.Shutdown(ctx, force); err != nil {
			errOut = errors.Join(errOut, fmt.Errorf("daemon shutdown: %w", err))
		}
	}
	d.closeMu.Lock()
	if d.closed {
		d.closeMu.Unlock()
		return errOut
	}
	d.closed = true
	listener := d.listener
	dispatcher := d.dispatcher
	embedded := d.embedded
	d.closeMu.Unlock()
	if listener != nil {
		if err := listener.Close(); err != nil {
			errOut = errors.Join(errOut, fmt.Errorf("daemon close: %w", err))
		}
	}
	if closer, ok := dispatcher.(interface{ Close(context.Context) error }); ok {
		if err := closer.Close(ctx); err != nil {
			errOut = errors.Join(errOut, fmt.Errorf("daemon close: %w", err))
		}
	}
	if embedded != nil {
		if err := embedded.Close(); err != nil {
			errOut = errors.Join(errOut, fmt.Errorf("daemon close: %w", err))
		}
	}
	return errOut
}

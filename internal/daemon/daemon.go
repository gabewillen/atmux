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
	"github.com/agentflare-ai/amux/internal/remote"
	"github.com/agentflare-ai/amux/internal/rpc"
)

const (
	// SpecVersion is the spec version implemented by the daemon.
	SpecVersion = "v1.22"
	// AmuxVersion is the daemon version string.
	AmuxVersion = "0.0.0-dev"
)

// Daemon hosts the JSON-RPC control plane.
type Daemon struct {
	resolver   *paths.Resolver
	cfg        config.Config
	manager    *manager.Manager
	hostMgr    *remote.HostManager
	dispatcher protocol.Dispatcher
	server     *rpc.Server
	listener   net.Listener
	embedded   *protocol.NATSServer
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
	var embedded *protocol.NATSServer
	var dispatcher protocol.Dispatcher
	var mgr *manager.Manager
	var hostMgr *remote.HostManager
	role := strings.TrimSpace(cfg.Node.Role)
	if role == "" {
		role = "director"
	}
	if role != "manager" {
		mode := strings.TrimSpace(cfg.NATS.Mode)
		if mode == "" {
			mode = "embedded"
		}
		credStore, err := remote.NewCredentialStore(cfg.NATS.JetStreamDir)
		if err != nil {
			return nil, fmt.Errorf("daemon: %w", err)
		}
		if mode != "external" {
			addr := cfg.NATS.Listen
			if strings.TrimSpace(addr) == "" {
				addr = "127.0.0.1:-1"
			}
			auth, err := credStore.HubAuth()
			if err != nil {
				return nil, fmt.Errorf("daemon: %w", err)
			}
			server, err := protocol.StartHubServer(ctx, protocol.HubServerConfig{
				Listen:            addr,
				Advertise:         cfg.NATS.AdvertiseURL,
				LeafListen:        cfg.NATS.LeafListen,
				LeafAdvertiseURL:  cfg.NATS.LeafAdvertiseURL,
				JetStreamDir:      cfg.NATS.JetStreamDir,
				OperatorPublicKey: auth.OperatorPublicKey,
				SystemAccountKey:  auth.SystemAccountKey,
				SystemAccountJWT:  auth.SystemAccountJWT,
				AccountPublicKey:  auth.AccountPublicKey,
				AccountJWT:        auth.AccountJWT,
			})
			if err != nil {
				return nil, fmt.Errorf("daemon: %w", err)
			}
			embedded = server
			cfg.NATS.HubURL = embedded.URL()
			if leafURL := embedded.LeafURL(); leafURL != "" {
				cfg.Remote.NATS.URL = leafURL
			}
		}
		if _, err := credStore.DirectorCredential(); err != nil {
			if embedded != nil {
				_ = embedded.Close()
			}
			return nil, fmt.Errorf("daemon: %w", err)
		}
		hubURL := cfg.NATS.HubURL
		dispatcher, err = protocol.NewNATSDispatcher(ctx, hubURL, protocol.NATSOptions{
			CredsPath: credStore.CredentialPath("director"),
		})
		if err != nil {
			if embedded != nil {
				_ = embedded.Close()
			}
			return nil, fmt.Errorf("daemon: %w", err)
		}
		mgr, err = manager.NewManager(ctx, resolver, cfg, dispatcher, AmuxVersion)
		if err != nil {
			if embedded != nil {
				_ = embedded.Close()
			}
			return nil, fmt.Errorf("daemon: %w", err)
		}
	} else {
		remoteMgr, err := remote.NewHostManager(cfg, resolver, AmuxVersion)
		if err != nil {
			return nil, fmt.Errorf("daemon: %w", err)
		}
		hostMgr = remoteMgr
	}
	daemon := &Daemon{
		resolver:   resolver,
		cfg:        cfg,
		manager:    mgr,
		hostMgr:    hostMgr,
		dispatcher: dispatcher,
		embedded:   embedded,
		logger:     logger,
		server:     rpc.NewServer(logger),
	}
	if mgr != nil {
		daemon.registerHandlers()
	}
	return daemon, nil
}

// Serve starts listening on the daemon socket.
func (d *Daemon) Serve(ctx context.Context) error {
	if d == nil {
		return fmt.Errorf("daemon serve: daemon is nil")
	}
	if d.hostMgr != nil {
		go func() {
			_ = d.hostMgr.Start(ctx)
		}()
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

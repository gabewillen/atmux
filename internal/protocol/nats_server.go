package protocol

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nats-io/nats-server/v2/server"
)

// HubServerConfig configures the embedded hub-mode NATS server.
type HubServerConfig struct {
	Listen    string
	Advertise string
	// LeafListen is the leaf node listen address.
	LeafListen string
	// LeafAdvertiseURL is the advertised leaf node URL.
	LeafAdvertiseURL  string
	JetStreamDir      string
	OperatorPublicKey string
	SystemAccountKey  string
	SystemAccountJWT  string
	AccountPublicKey  string
	AccountJWT        string
}

// LeafServerConfig configures the embedded leaf-mode NATS server.
type LeafServerConfig struct {
	Listen    string
	HubURL    string
	CredsPath string
}

// NATSServer wraps a running NATS server instance.
type NATSServer struct {
	server  *server.Server
	leafURL string
}

// StartHubServer starts a hub-mode NATS server with JetStream enabled.
func StartHubServer(ctx context.Context, cfg HubServerConfig) (*NATSServer, error) {
	host, port, err := splitHostPort(cfg.Listen)
	if err != nil {
		return nil, fmt.Errorf("nats hub: %w", err)
	}
	leafHost, leafPort, err := resolveLeafListen(cfg, host, port)
	if err != nil {
		return nil, fmt.Errorf("nats hub: %w", err)
	}
	if cfg.JetStreamDir == "" {
		return nil, fmt.Errorf("nats hub: jetstream dir required")
	}
	if err := os.MkdirAll(cfg.JetStreamDir, 0o700); err != nil {
		return nil, fmt.Errorf("nats hub: %w", err)
	}
	opts := &server.Options{
		Host:       host,
		Port:       port,
		ServerName: "amux-hub",
		NoSigs:     true,
		JetStream:  true,
		StoreDir:   cfg.JetStreamDir,
	}
	if cfg.Advertise != "" {
		opts.ClientAdvertise = cfg.Advertise
	}
	if leafPort != 0 {
		opts.LeafNode.Host = leafHost
		opts.LeafNode.Port = leafPort
		if cfg.LeafAdvertiseURL != "" {
			opts.LeafNode.Advertise = cfg.LeafAdvertiseURL
		}
	}
	if cfg.OperatorPublicKey != "" {
		opts.TrustedKeys = []string{cfg.OperatorPublicKey}
	}
	if cfg.AccountJWT != "" && cfg.AccountPublicKey != "" && cfg.SystemAccountKey != "" && cfg.SystemAccountJWT != "" {
		resolver := &server.MemAccResolver{}
		if err := resolver.Store(cfg.SystemAccountKey, cfg.SystemAccountJWT); err != nil {
			return nil, fmt.Errorf("nats hub: %w", err)
		}
		if err := resolver.Store(cfg.AccountPublicKey, cfg.AccountJWT); err != nil {
			return nil, fmt.Errorf("nats hub: %w", err)
		}
		opts.AccountResolver = resolver
		opts.SystemAccount = cfg.SystemAccountKey
	}
	srv, err := server.NewServer(opts)
	if err != nil {
		return nil, fmt.Errorf("nats hub: %w", err)
	}
	srv.ConfigureLogger()
	go srv.Start()
	if !srv.ReadyForConnections(5 * time.Second) {
		srv.Shutdown()
		return nil, fmt.Errorf("nats hub: server not ready")
	}
	wrap := &NATSServer{
		server:  srv,
		leafURL: buildLeafURL(leafHost, leafPort, cfg.LeafAdvertiseURL, opts.LeafNode.TLSConfig != nil),
	}
	go func() {
		<-ctx.Done()
		wrap.Shutdown()
	}()
	return wrap, nil
}

// StartLeafServer starts a leaf-mode NATS server connected to the hub.
func StartLeafServer(ctx context.Context, cfg LeafServerConfig) (*NATSServer, error) {
	host, port, err := splitHostPort(cfg.Listen)
	if err != nil {
		return nil, fmt.Errorf("nats leaf: %w", err)
	}
	remoteURL, err := url.Parse(cfg.HubURL)
	if err != nil {
		return nil, fmt.Errorf("nats leaf: %w", err)
	}
	opts := &server.Options{
		Host:       host,
		Port:       port,
		ServerName: "amux-leaf",
		NoSigs:     true,
		LeafNode: server.LeafNodeOpts{
			Remotes: []*server.RemoteLeafOpts{{
				URLs:        []*url.URL{remoteURL},
				Credentials: cfg.CredsPath,
			}},
		},
	}
	srv, err := server.NewServer(opts)
	if err != nil {
		return nil, fmt.Errorf("nats leaf: %w", err)
	}
	go srv.Start()
	if !srv.ReadyForConnections(5 * time.Second) {
		srv.Shutdown()
		return nil, fmt.Errorf("nats leaf: server not ready")
	}
	wrap := &NATSServer{server: srv}
	go func() {
		<-ctx.Done()
		wrap.Shutdown()
	}()
	return wrap, nil
}

// URL returns the client connection URL for the server.
func (s *NATSServer) URL() string {
	if s == nil || s.server == nil {
		return ""
	}
	return s.server.ClientURL()
}

// LeafURL returns the leaf connection URL for the server.
func (s *NATSServer) LeafURL() string {
	if s == nil {
		return ""
	}
	return s.leafURL
}

// LeafCount reports the number of leaf connections.
func (s *NATSServer) LeafCount() int {
	if s == nil || s.server == nil {
		return 0
	}
	return s.server.NumLeafNodes()
}

// Shutdown stops the server.
func (s *NATSServer) Shutdown() {
	if s == nil || s.server == nil {
		return
	}
	s.server.Shutdown()
}

// Close stops the server.
func (s *NATSServer) Close() error {
	if s == nil || s.server == nil {
		return nil
	}
	s.server.Shutdown()
	return nil
}

func splitHostPort(addr string) (string, int, error) {
	if addr == "" {
		addr = "127.0.0.1:-1"
	}
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0, err
	}
	port, err := parsePort(portStr)
	if err != nil {
		return "", 0, err
	}
	return host, port, nil
}

func parsePort(raw string) (int, error) {
	if raw == "" {
		return 0, fmt.Errorf("invalid port")
	}
	sign := 1
	if raw[0] == '-' {
		sign = -1
		raw = raw[1:]
	}
	if raw == "" {
		return 0, fmt.Errorf("invalid port")
	}
	var port int
	for _, ch := range raw {
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("invalid port")
		}
		port = port*10 + int(ch-'0')
	}
	return port * sign, nil
}

func resolveLeafListen(cfg HubServerConfig, listenHost string, listenPort int) (string, int, error) {
	leafListen := strings.TrimSpace(cfg.LeafListen)
	if leafListen != "" {
		host, port, err := splitHostPort(leafListen)
		if err != nil {
			return "", 0, err
		}
		if port <= 0 {
			port, err = allocatePort(host)
			if err != nil {
				return "", 0, err
			}
		}
		return host, port, nil
	}
	host := strings.TrimSpace(listenHost)
	if host == "" {
		host = "0.0.0.0"
	}
	port := 0
	if listenPort > 0 {
		port = listenPort + 3200
	}
	if port <= 0 {
		var err error
		port, err = allocatePort(host)
		if err != nil {
			return "", 0, err
		}
	}
	return host, port, nil
}

func allocatePort(host string) (int, error) {
	if strings.TrimSpace(host) == "" {
		host = "127.0.0.1"
	}
	listener, err := net.Listen("tcp", net.JoinHostPort(host, "0"))
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return 0, fmt.Errorf("unexpected listener address")
	}
	return addr.Port, nil
}

func buildLeafURL(host string, port int, advertise string, tlsEnabled bool) string {
	if port == 0 && advertise == "" {
		return ""
	}
	scheme := "nats"
	if tlsEnabled {
		scheme = "tls"
	}
	if strings.TrimSpace(advertise) != "" {
		advertise = strings.TrimSpace(advertise)
		if strings.Contains(advertise, "://") {
			return advertise
		}
		return fmt.Sprintf("%s://%s", scheme, advertise)
	}
	return fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(host, strconv.Itoa(port)))
}

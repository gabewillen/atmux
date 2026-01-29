// Package hub provides an embedded NATS hub server for the director role.
//
// The director starts an embedded NATS server with JetStream enabled
// and per-host authorization rules derived from auth.HostSubjectPermissions.
//
// See spec §5.5.5 and §5.5.6 for hub server requirements.
package hub

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/paths"
)

// AuthRule holds per-host authorization permissions.
type AuthRule struct {
	PublicKey string
	Publish   []string
	Subscribe []string
}

// Server wraps an embedded NATS server.
type Server struct {
	mu        sync.Mutex
	ns        *natsserver.Server
	opts      *natsserver.Options
	authRules map[string]AuthRule
	prefix    string
	configDir string
}

// Options configures the embedded hub server.
type Options struct {
	// Listen is the address to listen on (e.g., "0.0.0.0:4222").
	Listen string

	// JetStreamDir is the directory for JetStream data.
	JetStreamDir string

	// AdvertiseURL is the URL to advertise for leaf connections.
	AdvertiseURL string
}

// OptionsFromConfig creates Options from the amux configuration.
func OptionsFromConfig(cfg *config.Config) *Options {
	opts := &Options{
		Listen:       cfg.NATS.Listen,
		JetStreamDir: cfg.NATS.JetStreamDir,
		AdvertiseURL: cfg.NATS.AdvertiseURL,
	}

	if opts.Listen == "" {
		opts.Listen = "0.0.0.0:4222"
	}

	if opts.JetStreamDir == "" {
		opts.JetStreamDir = paths.DefaultResolver.NATSDataDir()
	}

	return opts
}

// Start creates and starts an embedded NATS server with JetStream.
//
// Per spec §5.5.5: the director MUST start a NATS hub server
// with JetStream enabled and leaf node support.
func Start(opts *Options) (*Server, error) {
	// Ensure JetStream directory exists
	if err := os.MkdirAll(opts.JetStreamDir, 0700); err != nil {
		return nil, fmt.Errorf("hub: create jetstream dir: %w", err)
	}

	nsOpts := &natsserver.Options{
		Host: "0.0.0.0",
		Port: 4222,

		// JetStream configuration
		JetStream:    true,
		StoreDir:     opts.JetStreamDir,
		JetStreamMaxMemory:  256 * 1024 * 1024,  // 256MB
		JetStreamMaxStore:   1024 * 1024 * 1024,  // 1GB

		// Leaf node support for manager connections
		LeafNode: natsserver.LeafNodeOpts{
			Host: "0.0.0.0",
			Port: 7422,
		},

		// Logging
		NoLog:  false,
		NoSigs: true,
	}

	// Parse listen address
	host, port, err := parseListenAddr(opts.Listen)
	if err == nil {
		nsOpts.Host = host
		nsOpts.Port = port
	}

	ns, err := natsserver.NewServer(nsOpts)
	if err != nil {
		return nil, fmt.Errorf("hub: create server: %w", err)
	}

	// Configure logging to stderr
	ns.ConfigureLogger()

	go ns.Start()

	// Wait for server to be ready
	if !ns.ReadyForConnections(10 * time.Second) {
		ns.Shutdown()
		return nil, fmt.Errorf("hub: server failed to start within timeout")
	}

	prefix := "amux"
	return &Server{
		ns:        ns,
		opts:      nsOpts,
		authRules: make(map[string]AuthRule),
		prefix:    prefix,
		configDir: opts.JetStreamDir,
	}, nil
}

// ClientURL returns the NATS client connection URL.
func (s *Server) ClientURL() string {
	return s.ns.ClientURL()
}

// Shutdown gracefully stops the embedded server.
func (s *Server) Shutdown() {
	if s.ns != nil {
		s.ns.Shutdown()
		s.ns.WaitForShutdown()
	}
}

// AddHostAuthorization adds per-host authorization rules and reloads
// the server configuration.
//
// Per spec §5.5.6.4: the hub server MUST enforce that each host_id
// can only publish/subscribe to its own subjects.
func (s *Server) AddHostAuthorization(publicKey string, publish, subscribe []string) error {
	s.mu.Lock()
	s.authRules[publicKey] = AuthRule{
		PublicKey: publicKey,
		Publish:   publish,
		Subscribe: subscribe,
	}

	// Build NKey-to-hostID map for auth config generation
	hosts := make(map[string]string)
	for key := range s.authRules {
		hosts[key] = key
	}
	prefix := s.prefix
	configDir := s.configDir
	s.mu.Unlock()

	// Write the auth config file
	authPath, err := WriteAuthConfig(configDir, prefix, hosts)
	if err != nil {
		return fmt.Errorf("add host authorization: %w", err)
	}

	// Reload the server with updated auth configuration
	newOpts := *s.opts
	newOpts.ConfigFile = authPath
	if err := s.ns.ReloadOptions(&newOpts); err != nil {
		return fmt.Errorf("reload server options: %w", err)
	}

	return nil
}

// parseListenAddr parses a listen address like "0.0.0.0:4222".
func parseListenAddr(addr string) (host string, port int, err error) {
	// Handle common formats
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			host = addr[:i]
			var p int
			_, err = fmt.Sscanf(addr[i+1:], "%d", &p)
			if err != nil {
				return "", 0, fmt.Errorf("invalid port in %q: %w", addr, err)
			}
			return host, p, nil
		}
	}
	return addr, 4222, nil
}

// GenerateAuthConfig generates the NATS server authorization configuration
// for a set of hosts. This can be written to a file and loaded via include.
//
// Per spec §5.5.6.4: "The hub MUST enforce per-host subject permissions."
func GenerateAuthConfig(prefix string, hosts map[string]string) string {
	if len(hosts) == 0 {
		return ""
	}

	// Build authorization block
	var cfg string
	cfg += "authorization {\n"
	cfg += "  users = [\n"

	for hostID, nkeyPub := range hosts {
		cfg += "    {\n"
		cfg += fmt.Sprintf("      nkey: %s\n", nkeyPub)
		cfg += "      permissions: {\n"
		cfg += "        publish: {\n"
		cfg += "          allow: [\n"
		cfg += fmt.Sprintf("            \"%s.handshake.%s\",\n", prefix, hostID)
		cfg += fmt.Sprintf("            \"%s.events.%s\",\n", prefix, hostID)
		cfg += fmt.Sprintf("            \"%s.pty.%s.*.out\",\n", prefix, hostID)
		cfg += fmt.Sprintf("            \"%s.comm.director\",\n", prefix)
		cfg += fmt.Sprintf("            \"%s.comm.manager.*\",\n", prefix)
		cfg += fmt.Sprintf("            \"%s.comm.agent.*.>\",\n", prefix)
		cfg += fmt.Sprintf("            \"%s.comm.broadcast\"\n", prefix)
		cfg += "          ]\n"
		cfg += "        }\n"
		cfg += "        subscribe: {\n"
		cfg += "          allow: [\n"
		cfg += fmt.Sprintf("            \"%s.ctl.%s\",\n", prefix, hostID)
		cfg += fmt.Sprintf("            \"%s.pty.%s.*.in\",\n", prefix, hostID)
		cfg += fmt.Sprintf("            \"%s.comm.manager.%s\",\n", prefix, hostID)
		cfg += fmt.Sprintf("            \"%s.comm.agent.%s.>\",\n", prefix, hostID)
		cfg += fmt.Sprintf("            \"%s.comm.broadcast\",\n", prefix)
		cfg += "            \"_INBOX.>\"\n"
		cfg += "          ]\n"
		cfg += "        }\n"
		cfg += "      }\n"
		cfg += "    },\n"
	}

	cfg += "  ]\n"
	cfg += "}\n"
	return cfg
}

// WriteAuthConfig writes the authorization config to a file.
func WriteAuthConfig(dir, prefix string, hosts map[string]string) (string, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("create auth config dir: %w", err)
	}

	content := GenerateAuthConfig(prefix, hosts)
	path := filepath.Join(dir, "auth.conf")

	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return "", fmt.Errorf("write auth config: %w", err)
	}

	return path, nil
}

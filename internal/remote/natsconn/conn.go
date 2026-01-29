// Package natsconn provides NATS connection management for amux.
//
// This package handles connecting to NATS as either a director (hub)
// or manager (leaf) role, with support for credential-based authentication,
// JetStream initialization, and reconnection handling.
//
// See spec §5.5.6 for NATS connectivity requirements.
package natsconn

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/agentflare-ai/amux/internal/config"
)

// Conn wraps a NATS connection with amux-specific configuration.
type Conn struct {
	nc     *nats.Conn
	js     jetstream.JetStream
	hostID string
	role   string
}

// Options configures a NATS connection.
type Options struct {
	// URL is the NATS server URL to connect to.
	URL string

	// CredsFile is the path to a NATS credentials file (NKey seed).
	CredsFile string

	// NKeySeed is the raw NKey seed bytes (alternative to CredsFile).
	NKeySeed []byte

	// HostID is this node's host identifier.
	HostID string

	// Role is "director" or "manager".
	Role string

	// Name is a connection name for debugging.
	Name string

	// ReconnectWait is the time to wait between reconnect attempts.
	ReconnectWait time.Duration

	// MaxReconnects is the maximum number of reconnect attempts.
	// -1 means unlimited.
	MaxReconnects int

	// DisconnectHandler is called when the connection is lost.
	DisconnectHandler func(*nats.Conn, error)

	// ReconnectHandler is called when the connection is restored.
	ReconnectHandler func(*nats.Conn)

	// ClosedHandler is called when the connection is permanently closed.
	ClosedHandler func(*nats.Conn)
}

// OptionsFromConfig creates Options from the amux configuration.
func OptionsFromConfig(cfg *config.Config, hostID string) *Options {
	opts := &Options{
		HostID:        hostID,
		Role:          cfg.Node.Role,
		MaxReconnects: cfg.Remote.ReconnectMaxAttempts,
	}

	if cfg.Node.Role == "manager" {
		opts.URL = cfg.Remote.NATS.URL
		opts.CredsFile = cfg.Remote.NATS.CredsPath
		opts.Name = "amux-manager-" + hostID
	} else {
		// Director connects to its own hub server
		if cfg.Remote.NATS.URL != "" {
			opts.URL = cfg.Remote.NATS.URL
		} else {
			opts.URL = "nats://localhost:4222"
		}
		opts.Name = "amux-director"
	}

	if cfg.Remote.ReconnectBackoffBase.Duration > 0 {
		opts.ReconnectWait = cfg.Remote.ReconnectBackoffBase.Duration
	} else {
		opts.ReconnectWait = time.Second
	}

	return opts
}

// Connect establishes a NATS connection.
func Connect(ctx context.Context, opts *Options) (*Conn, error) {
	natsOpts := []nats.Option{
		nats.Name(opts.Name),
		nats.ReconnectWait(opts.ReconnectWait),
		nats.MaxReconnects(opts.MaxReconnects),
	}

	if opts.CredsFile != "" {
		natsOpts = append(natsOpts, nats.UserCredentials(opts.CredsFile))
	} else if len(opts.NKeySeed) > 0 {
		opt, err := nats.NkeyOptionFromSeed(string(opts.NKeySeed))
		if err != nil {
			return nil, fmt.Errorf("nkey option from seed: %w", err)
		}
		natsOpts = append(natsOpts, opt)
	}

	if opts.DisconnectHandler != nil {
		natsOpts = append(natsOpts, nats.DisconnectErrHandler(opts.DisconnectHandler))
	}
	if opts.ReconnectHandler != nil {
		natsOpts = append(natsOpts, nats.ReconnectHandler(opts.ReconnectHandler))
	}
	if opts.ClosedHandler != nil {
		natsOpts = append(natsOpts, nats.ClosedHandler(opts.ClosedHandler))
	}

	url := opts.URL
	if url == "" {
		url = nats.DefaultURL
	}

	nc, err := nats.Connect(url, natsOpts...)
	if err != nil {
		return nil, fmt.Errorf("nats connect to %s: %w", url, err)
	}

	conn := &Conn{
		nc:     nc,
		hostID: opts.HostID,
		role:   opts.Role,
	}

	return conn, nil
}

// NC returns the underlying NATS connection.
func (c *Conn) NC() *nats.Conn {
	return c.nc
}

// JetStream returns the JetStream context, initializing it on first call.
func (c *Conn) JetStream() (jetstream.JetStream, error) {
	if c.js != nil {
		return c.js, nil
	}
	js, err := jetstream.New(c.nc)
	if err != nil {
		return nil, fmt.Errorf("jetstream init: %w", err)
	}
	c.js = js
	return c.js, nil
}

// HostID returns the host identifier for this connection.
func (c *Conn) HostID() string {
	return c.hostID
}

// Role returns the role for this connection.
func (c *Conn) Role() string {
	return c.role
}

// IsConnected returns true if the NATS connection is currently active.
func (c *Conn) IsConnected() bool {
	return c.nc != nil && c.nc.IsConnected()
}

// Close gracefully drains and closes the NATS connection.
func (c *Conn) Close() error {
	if c.nc == nil {
		return nil
	}
	return c.nc.Drain()
}

// Publish publishes a message to a NATS subject.
func (c *Conn) Publish(subject string, data []byte) error {
	return c.nc.Publish(subject, data)
}

// Subscribe subscribes to a NATS subject.
func (c *Conn) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	return c.nc.Subscribe(subject, handler)
}

// Request sends a request and waits for a reply.
func (c *Conn) Request(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
	return c.nc.Request(subject, data, timeout)
}

// Flush flushes the connection buffer.
func (c *Conn) Flush() error {
	return c.nc.Flush()
}

// NkeyOptionFromSeed is a helper that creates a NATS NKey option from a seed.
func NkeyOptionFromSeed(seed string) (nats.Option, error) {
	return nats.NkeyOptionFromSeed(seed)
}

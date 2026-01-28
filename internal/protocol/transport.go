// Package protocol provides remote communication transport functionality.
// This package transports events generically using NATS without any
// agent-specific knowledge.
package protocol

import (
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// Common sentinel errors for protocol operations.
var (
	// ErrConnectionFailed indicates NATS connection failed.
	ErrConnectionFailed = errors.New("connection failed")

	// ErrPublishFailed indicates event publishing failed.
	ErrPublishFailed = errors.New("publish failed")

	// ErrSubscribeFailed indicates subscription setup failed.
	ErrSubscribeFailed = errors.New("subscribe failed")
)

// Transport manages NATS-based event communication.
// Transports events generically for both local and remote distribution.
type Transport struct {
	conn   *nats.Conn
	config TransportConfig
}

// TransportConfig holds NATS transport configuration.
type TransportConfig struct {
	URL             string
	ConnectionName  string
	ReconnectDelay  time.Duration
	MaxReconnects   int
	CredentialsFile string
}

// NewTransport creates a new NATS transport instance.
func NewTransport(config TransportConfig) (*Transport, error) {
	if config.URL == "" {
		config.URL = nats.DefaultURL
	}
	if config.ConnectionName == "" {
		config.ConnectionName = "amux"
	}
	if config.ReconnectDelay == 0 {
		config.ReconnectDelay = 2 * time.Second
	}
	if config.MaxReconnects == 0 {
		config.MaxReconnects = -1 // Infinite
	}

	return &Transport{
		config: config,
	}, nil
}

// Connect establishes NATS connection with configured options.
func (t *Transport) Connect() error {
	opts := []nats.Option{
		nats.Name(t.config.ConnectionName),
		nats.ReconnectWait(t.config.ReconnectDelay),
		nats.MaxReconnects(t.config.MaxReconnects),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			// Handle disconnection - logging deferred to Phase 0 completion
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			// Handle reconnection - logging deferred to Phase 0 completion
		}),
	}

	if t.config.CredentialsFile != "" {
		opts = append(opts, nats.UserCredentials(t.config.CredentialsFile))
	}

	conn, err := nats.Connect(t.config.URL, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS at %s: %w", t.config.URL, ErrConnectionFailed)
	}

	t.conn = conn
	return nil
}

// Publish sends an event to the specified NATS subject.
func (t *Transport) Publish(subject string, data []byte) error {
	if t.conn == nil {
		return fmt.Errorf("not connected: %w", ErrConnectionFailed)
	}

	if err := t.conn.Publish(subject, data); err != nil {
		return fmt.Errorf("failed to publish to %s: %w", subject, ErrPublishFailed)
	}

	return nil
}

// Subscribe sets up a subscription to the specified NATS subject.
func (t *Transport) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	if t.conn == nil {
		return nil, fmt.Errorf("not connected: %w", ErrConnectionFailed)
	}

	sub, err := t.conn.Subscribe(subject, handler)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to %s: %w", subject, ErrSubscribeFailed)
	}

	return sub, nil
}

// Close closes the NATS connection.
func (t *Transport) Close() {
	if t.conn != nil {
		t.conn.Close()
		t.conn = nil
	}
}
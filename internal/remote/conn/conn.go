package conn

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Options configuration for NATS connection.
type Options struct {
	URL           string
	Name          string
	CredsPath     string
	ReconnectWait time.Duration
	MaxReconnects int
	OnDisconnect  func(*nats.Conn, error)
	OnReconnect   func(*nats.Conn)
	OnClosed      func(*nats.Conn)
}

// Connect establishes a connection to NATS.
func Connect(opts Options) (*nats.Conn, error) {
	connectOpts := []nats.Option{
		nats.Name(opts.Name),
		nats.ReconnectWait(opts.ReconnectWait),
		nats.MaxReconnects(opts.MaxReconnects),
	}

	if opts.CredsPath != "" {
		connectOpts = append(connectOpts, nats.UserCredentials(opts.CredsPath))
	}

	if opts.OnDisconnect != nil {
		connectOpts = append(connectOpts, nats.DisconnectErrHandler(opts.OnDisconnect))
	}
	if opts.OnReconnect != nil {
		connectOpts = append(connectOpts, nats.ReconnectHandler(opts.OnReconnect))
	}
	if opts.OnClosed != nil {
		connectOpts = append(connectOpts, nats.ClosedHandler(opts.OnClosed))
	}

	nc, err := nats.Connect(opts.URL, connectOpts...)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}

	return nc, nil
}

// JetStream returns a JetStream context from a NATS connection.
func JetStream(nc *nats.Conn) (jetstream.JetStream, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("jetstream init: %w", err)
	}
	return js, nil
}

package protocol

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// NATSOptions configures NATS connection metadata and auth.
type NATSOptions struct {
	// Name sets the client name.
	Name string
	// User sets the username for auth.
	User string
	// Password sets the password for auth.
	Password string
	// Token sets the auth token.
	Token string
	// CredsPath sets the path to a NATS .creds file.
	CredsPath string
	// AllowNoJetStream permits connections without JetStream enabled.
	AllowNoJetStream bool
}

// NATSDispatcher publishes and subscribes to events over NATS.
type NATSDispatcher struct {
	conn     *nats.Conn
	js       nats.JetStreamContext
	closedCh chan struct{}
}

// NewNATSDispatcher connects to a NATS server and returns a dispatcher.
func NewNATSDispatcher(ctx context.Context, rawURL string, options NATSOptions) (*NATSDispatcher, error) {
	if rawURL == "" {
		return nil, fmt.Errorf("nats connect: empty url")
	}
	closedCh := make(chan struct{})
	opts := []nats.Option{
		nats.Name("amux"),
		nats.ClosedHandler(func(*nats.Conn) {
			closeOnce(closedCh)
		}),
	}
	if options.Name != "" {
		opts = append(opts, nats.Name(options.Name))
	}
	if options.CredsPath != "" {
		opts = append(opts, nats.UserCredentials(options.CredsPath))
	} else if options.Token != "" {
		opts = append(opts, nats.Token(options.Token))
	} else if options.User != "" || options.Password != "" {
		opts = append(opts, nats.UserInfo(options.User, options.Password))
	}
	if deadline, ok := ctx.Deadline(); ok {
		timeout := time.Until(deadline)
		if timeout > 0 {
			opts = append(opts, nats.Timeout(timeout))
		}
	}
	conn, err := nats.Connect(rawURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	js, err := conn.JetStream()
	if err != nil {
		if !options.AllowNoJetStream {
			conn.Close()
			return nil, fmt.Errorf("nats connect: %w", err)
		}
	}
	return &NATSDispatcher{conn: conn, js: js, closedCh: closedCh}, nil
}

// JetStream returns the JetStream context.
func (d *NATSDispatcher) JetStream() nats.JetStreamContext {
	if d == nil {
		return nil
	}
	return d.js
}

// Close closes the underlying NATS connection.
func (d *NATSDispatcher) Close(ctx context.Context) error {
	if d == nil || d.conn == nil {
		return nil
	}
	if ctx != nil && ctx.Err() != nil {
		return fmt.Errorf("nats close: %w", ctx.Err())
	}
	d.conn.Close()
	return nil
}

// Publish publishes an event to a subject.
func (d *NATSDispatcher) Publish(ctx context.Context, subject string, event Event) error {
	if ctx.Err() != nil {
		return fmt.Errorf("nats publish: %w", ctx.Err())
	}
	if d == nil || d.conn == nil {
		return fmt.Errorf("nats publish: dispatcher unavailable")
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("nats publish: %w", err)
	}
	return d.PublishRaw(ctx, subject, payload, "")
}

// Subscribe subscribes to a subject.
func (d *NATSDispatcher) Subscribe(ctx context.Context, subject string, handler func(Event)) (Subscription, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("nats subscribe: %w", ctx.Err())
	}
	if d == nil || d.conn == nil || handler == nil {
		return nil, fmt.Errorf("nats subscribe: invalid")
	}
	sub, err := d.conn.Subscribe(subject, func(msg *nats.Msg) {
		var event Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			return
		}
		handler(event)
	})
	if err != nil {
		return nil, fmt.Errorf("nats subscribe: %w", err)
	}
	return &natsSubscription{sub: sub}, nil
}

// PublishRaw publishes a raw payload to a subject with optional reply.
func (d *NATSDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	if ctx.Err() != nil {
		return fmt.Errorf("nats publish: %w", ctx.Err())
	}
	if d == nil || d.conn == nil {
		return fmt.Errorf("nats publish: dispatcher unavailable")
	}
	msg := &nats.Msg{Subject: subject, Reply: reply, Data: payload}
	if err := d.conn.PublishMsg(msg); err != nil {
		return fmt.Errorf("nats publish: %w", err)
	}
	return nil
}

// SubscribeRaw subscribes to a subject and receives raw NATS messages.
func (d *NATSDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(Message)) (Subscription, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("nats subscribe: %w", ctx.Err())
	}
	if d == nil || d.conn == nil || handler == nil {
		return nil, fmt.Errorf("nats subscribe: invalid")
	}
	sub, err := d.conn.Subscribe(subject, func(msg *nats.Msg) {
		handler(Message{Subject: msg.Subject, Reply: msg.Reply, Data: msg.Data})
	})
	if err != nil {
		return nil, fmt.Errorf("nats subscribe: %w", err)
	}
	return &natsSubscription{sub: sub}, nil
}

// Request sends a request and waits for a response.
func (d *NATSDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (Message, error) {
	if ctx.Err() != nil {
		return Message{}, fmt.Errorf("nats request: %w", ctx.Err())
	}
	if d == nil || d.conn == nil {
		return Message{}, fmt.Errorf("nats request: dispatcher unavailable")
	}
	reqCtx := ctx
	var cancel context.CancelFunc
	if timeout > 0 {
		reqCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	msg, err := d.conn.RequestWithContext(reqCtx, subject, payload)
	if err != nil {
		return Message{}, fmt.Errorf("nats request: %w", err)
	}
	return Message{Subject: msg.Subject, Reply: msg.Reply, Data: msg.Data}, nil
}

// MaxPayload returns the maximum payload size for the connection.
func (d *NATSDispatcher) MaxPayload() int {
	if d == nil || d.conn == nil {
		return 0
	}
	return int(d.conn.MaxPayload())
}

// Closed returns a channel that closes when the connection closes.
func (d *NATSDispatcher) Closed() <-chan struct{} {
	if d == nil {
		ch := make(chan struct{})
		close(ch)
		return ch
	}
	return d.closedCh
}

type natsSubscription struct {
	sub *nats.Subscription
}

func (n *natsSubscription) Unsubscribe() error {
	if n == nil || n.sub == nil {
		return nil
	}
	if err := n.sub.Unsubscribe(); err != nil {
		return fmt.Errorf("nats unsubscribe: %w", err)
	}
	return nil
}

func closeOnce(ch chan struct{}) {
	defer func() {
		_ = recover()
	}()
	close(ch)
}

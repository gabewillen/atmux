package protocol

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrNoopDispatcher is returned when using a noop dispatcher.
var ErrNoopDispatcher = errors.New("noop dispatcher")

// Event is the generic event envelope used for dispatch.
type Event struct {
	Name      string
	Payload   any
	OccurredAt time.Time
}

// Dispatcher publishes and subscribes to events over NATS.
type Dispatcher interface {
	Publish(ctx context.Context, subject string, event Event) error
	Subscribe(ctx context.Context, subject string, handler func(Event)) (Subscription, error)
}

// Subscription represents an active event subscription.
type Subscription interface {
	Unsubscribe() error
}

// NoopDispatcher is a placeholder dispatcher that drops all events.
type NoopDispatcher struct{}

// Publish drops the event.
func (d *NoopDispatcher) Publish(ctx context.Context, subject string, event Event) error {
	return fmt.Errorf("publish %s: %w", subject, ErrNoopDispatcher)
}

// Subscribe returns a noop subscription.
func (d *NoopDispatcher) Subscribe(ctx context.Context, subject string, handler func(Event)) (Subscription, error) {
	return &noopSub{}, nil
}

type noopSub struct{}

func (n *noopSub) Unsubscribe() error {
	return nil
}

// Subject joins subject segments for NATS routing.
func Subject(parts ...string) string {
	var subject string
	for _, part := range parts {
		if part == "" {
			continue
		}
		if subject == "" {
			subject = part
			continue
		}
		subject += "." + part
	}
	return subject
}

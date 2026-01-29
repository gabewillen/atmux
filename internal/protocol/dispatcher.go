package protocol

import (
	"context"
	"time"
)

// Event is the generic event envelope used for dispatch.
type Event struct {
	Name       string
	Payload    any
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

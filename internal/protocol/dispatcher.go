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

// Message is a raw NATS message payload with optional reply subject.
type Message struct {
	Subject string
	Reply   string
	Data    []byte
}

// Dispatcher publishes and subscribes to events over NATS.
type Dispatcher interface {
	Publish(ctx context.Context, subject string, event Event) error
	Subscribe(ctx context.Context, subject string, handler func(Event)) (Subscription, error)
	PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error
	SubscribeRaw(ctx context.Context, subject string, handler func(Message)) (Subscription, error)
	Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (Message, error)
	MaxPayload() int
	Closed() <-chan struct{}
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

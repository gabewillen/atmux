# package messaging

`import "github.com/copilot-claude-sonnet-4/amux/internal/messaging"`

Package messaging implements inter-agent messaging routes per spec §6.4.
This package handles message routing, addressing, and delivery over NATS.

- `BroadcastID` — BroadcastID is the special ID for broadcast messages per spec §6.4.
- `type AdapterPatterns` — AdapterPatterns represents adapter-specific message detection patterns per spec §6.4.3.
- `type AgentMessage` — AgentMessage represents a message sent between agents per spec §6.4.
- `type MessageDetector` — MessageDetector detects outbound messages from PTY output using adapter patterns.
- `type MessageEvent` — MessageEvent represents message routing events per spec §6.4.2.
- `type Router` — Router handles message routing and addressing per spec §6.4.1.

### Constants

#### BroadcastID

```go
const BroadcastID muid.MUID = 0
```

BroadcastID is the special ID for broadcast messages per spec §6.4.


## type AdapterPatterns

```go
type AdapterPatterns struct {
	// Prompt is the pattern for detecting prompt readiness.
	Prompt string `json:"prompt"`

	// RateLimit is the pattern for detecting rate limiting.
	RateLimit string `json:"rate_limit"`

	// Error is the pattern for detecting errors.
	Error string `json:"error"`

	// Completion is the pattern for detecting task completion.
	Completion string `json:"completion"`

	// Message is the pattern for detecting outbound messages (optional).
	Message string `json:"message,omitempty"`
}
```

AdapterPatterns represents adapter-specific message detection patterns per spec §6.4.3.

## type AgentMessage

```go
type AgentMessage struct {
	// ID is the unique identifier for this message.
	ID muid.MUID `json:"id"`

	// From is the sender runtime ID (set by publishing component).
	From muid.MUID `json:"from"`

	// To is the recipient runtime ID (set by publishing component, or BroadcastID).
	To muid.MUID `json:"to"`

	// ToSlug is the recipient token captured from text (typically agent_slug).
	ToSlug string `json:"to_slug"`

	// Content is the message content.
	Content string `json:"content"`

	// Timestamp is when this message was sent.
	Timestamp time.Time `json:"timestamp"`
}
```

AgentMessage represents a message sent between agents per spec §6.4.

## type MessageDetector

```go
type MessageDetector struct {
	// patterns holds compiled regex patterns for message detection.
	patterns map[string]*regexp.Regexp

	// mu protects concurrent access to patterns.
	mu sync.RWMutex
}
```

MessageDetector detects outbound messages from PTY output using adapter patterns.

### Functions returning MessageDetector

#### NewMessageDetector

```go
func NewMessageDetector() *MessageDetector
```

NewMessageDetector creates a new message detector with adapter patterns.


### Methods

#### MessageDetector.DetectMessage

```go
func () DetectMessage(adapterName, output string) (toSlug, content string, detected bool)
```

DetectMessage attempts to detect an outbound message from PTY output.

#### MessageDetector.LoadPatterns

```go
func () LoadPatterns(adapterName string, patterns *AdapterPatterns) error
```

LoadPatterns loads adapter patterns for message detection.


## type MessageEvent

```go
type MessageEvent struct {
	// Type is the event type: "message.outbound", "message.inbound", or "message.broadcast".
	Type string `json:"type"`

	// Message is the agent message.
	Message *AgentMessage `json:"message"`

	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`
}
```

MessageEvent represents message routing events per spec §6.4.2.

## type Router

```go
type Router struct {
	// roster is the roster store for participant lookup.
	roster *roster.Store

	// hostID is this host's identifier.
	hostID string

	// directorID is the director's runtime ID.
	directorID muid.MUID

	// localManagerID is the local host manager's runtime ID.
	localManagerID muid.MUID

	// mu protects concurrent access to router state.
	mu sync.RWMutex

	// ctx is the router context.
	ctx context.Context

	// cancel cancels the router context.
	cancel context.CancelFunc
}
```

Router handles message routing and addressing per spec §6.4.1.

### Functions returning Router

#### NewRouter

```go
func NewRouter(roster *roster.Store, hostID string, directorID, localManagerID muid.MUID) *Router
```

NewRouter creates a new message router.


### Methods

#### Router.Close

```go
func () Close() error
```

Close gracefully shuts down the router.

#### Router.GetDeliveryChannels

```go
func () GetDeliveryChannels(msg *AgentMessage) ([]string, error)
```

GetDeliveryChannels returns channels where a message should be published.

#### Router.GetParticipantChannels

```go
func () GetParticipantChannels(participantID muid.MUID) ([]string, error)
```

GetParticipantChannels returns NATS subject names for a participant per spec §5.5.7.1.

#### Router.ProcessOutboundMessage

```go
func () ProcessOutboundMessage(fromID muid.MUID, toSlug, content string) (*AgentMessage, error)
```

ProcessOutboundMessage processes an outbound message from an agent per spec §6.4.1.

#### Router.resolveToSlug

```go
func () resolveToSlug(toSlug string) (muid.MUID, error)
```

resolveToSlug resolves ToSlug to a runtime ID per spec §6.4.1.3.



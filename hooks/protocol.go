package hooks

// MessageType identifies a hook message type.
type MessageType string

// Message is a placeholder hook protocol envelope.
type Message struct {
	Type MessageType
	Payload []byte
}

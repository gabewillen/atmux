package remote

import "errors"

var (
	// ErrInvalidSubject is returned for malformed NATS subjects.
	ErrInvalidSubject = errors.New("invalid subject")
	// ErrInvalidMessage is returned for malformed protocol messages.
	ErrInvalidMessage = errors.New("invalid message")
	// ErrHostDisconnected is returned when a host is offline.
	ErrHostDisconnected = errors.New("host disconnected")
	// ErrNotReady is returned when the remote daemon has not completed handshake.
	ErrNotReady = errors.New("remote not ready")
	// ErrSessionConflict is returned when spawn conflicts with existing session metadata.
	ErrSessionConflict = errors.New("session conflict")
	// ErrSessionNotFound is returned when a session is missing.
	ErrSessionNotFound = errors.New("session not found")
	// ErrReplayDisabled is returned when replay buffering is disabled.
	ErrReplayDisabled = errors.New("replay disabled")
	// ErrBootstrapFailed is returned when SSH bootstrap fails.
	ErrBootstrapFailed = errors.New("bootstrap failed")
	// ErrMessageTargetUnknown is returned when a message recipient cannot be resolved.
	ErrMessageTargetUnknown = errors.New("message target unknown")
)

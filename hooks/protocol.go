package main

// Protocol definitions for the hook <-> tracker communication.

// EventType identifies the type of intercepted event.
type EventType uint8

const (
	EventSpawn EventType = iota
	EventExec
	EventExit
)

// Header is the fixed-size header for messages.
type Header struct {
	Type       EventType
	Pid        int32
	Ppid       int32
	PayloadLen uint32
}

// Handshake is the initial message sent by the hook library.
type Handshake struct {
	Version uint32
	Pid     int32
}

// Constants for protocol versioning and socket paths.
const (
	ProtocolVersion = 1
	EnvSocketPath   = "AMUX_HOOK_SOCKET"
)

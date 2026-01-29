- `ProtocolVersion, EnvSocketPath` — Constants for protocol versioning and socket paths.
- `func AmuxHookInit()`
- `func RecvFD(conn *net.UnixConn) (int, error)` — RecvFD receives a file descriptor from a Unix domain socket.
- `func SendFD(conn *net.UnixConn, fd int) error` — SendFD sends a file descriptor over a Unix domain socket.
- `func main()` — Main required for c-shared build.
- `type EventType` — EventType identifies the type of intercepted event.
- `type Handshake` — Handshake is the initial message sent by the hook library.
- `type Header` — Header is the fixed-size header for messages.

### Constants

#### ProtocolVersion, EnvSocketPath

```go
const (
	ProtocolVersion = 1
	EnvSocketPath   = "AMUX_HOOK_SOCKET"
)
```

Constants for protocol versioning and socket paths.


### Functions

#### AmuxHookInit

```go
func AmuxHookInit()
```

#### RecvFD

```go
func RecvFD(conn *net.UnixConn) (int, error)
```

RecvFD receives a file descriptor from a Unix domain socket.

#### SendFD

```go
func SendFD(conn *net.UnixConn, fd int) error
```

SendFD sends a file descriptor over a Unix domain socket.

#### main

```go
func main()
```

Main required for c-shared build.


## type EventType

```go
type EventType uint8
```

EventType identifies the type of intercepted event.

### Constants

#### EventSpawn, EventExec, EventExit

```go
const (
	EventSpawn EventType = iota
	EventExec
	EventExit
)
```


## type Handshake

```go
type Handshake struct {
	Version uint32
	Pid     int32
}
```

Handshake is the initial message sent by the hook library.

## type Header

```go
type Header struct {
	Type       EventType
	Pid        int32
	Ppid       int32
	PayloadLen uint32
}
```

Header is the fixed-size header for messages.


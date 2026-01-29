# package buffer

`import "github.com/agentflare-ai/amux/internal/remote/buffer"`

Package buffer provides a ring buffer for PTY output replay.

The ring buffer retains the most recent N bytes of PTY output,
dropping the oldest bytes when the capacity is exceeded.
It is used by the manager-role daemon to support the replay
request-reply protocol.

See spec §5.5.7.3 for replay buffer requirements.

- `type Ring` — Ring is a thread-safe ring buffer that retains the most recent cap bytes.

## type Ring

```go
type Ring struct {
	mu   sync.Mutex
	buf  []byte
	cap  int64
	head int64 // write position (circular)
	size int64 // current bytes stored
}
```

Ring is a thread-safe ring buffer that retains the most recent cap bytes.

When the buffer is full, writing additional bytes drops the oldest
bytes (ring-buffer semantics). The buffer supports taking a snapshot
of the current contents in oldest-to-newest byte order.

A zero-capacity Ring disables buffering: writes are discarded and
Snapshot returns nil.

### Functions returning Ring

#### NewRing

```go
func NewRing(cap int64) *Ring
```

NewRing creates a Ring with the given capacity in bytes.
If cap is 0, buffering is disabled.


### Methods

#### Ring.Cap

```go
func () Cap() int64
```

Cap returns the buffer capacity in bytes.

#### Ring.Enabled

```go
func () Enabled() bool
```

Enabled returns true if the buffer has a non-zero capacity.

#### Ring.Len

```go
func () Len() int64
```

Len returns the current number of bytes stored.

#### Ring.Reset

```go
func () Reset()
```

Reset clears the buffer contents without changing capacity.

#### Ring.Snapshot

```go
func () Snapshot() []byte
```

Snapshot returns a copy of the current buffer contents in oldest-to-newest
byte order. Returns nil if the buffer is disabled or empty.

This is used to fulfill replay requests per spec §5.5.7.3:
"The replayed bytes MUST correspond to a snapshot of the replay buffer
taken at the moment the daemon receives the replay request."

#### Ring.Write

```go
func () Write(data []byte)
```

Write appends data to the ring buffer.
If the buffer is disabled (cap == 0), the data is discarded.
If data exceeds the buffer capacity, only the last cap bytes are retained.



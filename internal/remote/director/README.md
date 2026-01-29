# package director

`import "github.com/agentflare-ai/amux/internal/remote/director"`

- `type Director` — Director orchestrates remote hosts.
- `type Options` — Options for Director.

## type Director

```go
type Director struct {
	opts Options
	nc   *nats.Conn
	js   jetstream.JetStream
	kv   jetstream.KeyValue
}
```

Director orchestrates remote hosts.

### Functions returning Director

#### New

```go
func New(opts Options) *Director
```

New creates a new Director.


### Methods

#### Director.ListHosts

```go
func () ListHosts(ctx context.Context) ([]protocol.HostInfo, error)
```

ListHosts returns known hosts from KV.

#### Director.Replay

```go
func () Replay(ctx context.Context, hostID, sessionID string, sinceSeq uint64) error
```

Replay requests replay for a session.

#### Director.Signal

```go
func () Signal(ctx context.Context, hostID, sessionID, signal string) error
```

Signal sends a signal to a session.

#### Director.Spawn

```go
func () Spawn(ctx context.Context, hostID string, payload protocol.SpawnPayload) (string, error)
```

Spawn sends a spawn request to the target host.

#### Director.Start

```go
func () Start(ctx context.Context) error
```

Start connects to NATS.

#### Director.Stop

```go
func () Stop()
```

Stop disconnects.

#### Director.request

```go
func () request(ctx context.Context, hostID string, req protocol.ControlRequest) (*protocol.ControlResponse, error)
```


## type Options

```go
type Options struct {
	Config config.Config
}
```

Options for Director.


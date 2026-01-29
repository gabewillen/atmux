# package kv

`import "github.com/agentflare-ai/amux/internal/remote/kv"`

- `func EnsureBucket(ctx context.Context, js jetstream.JetStream, bucket string) (jetstream.KeyValue, error)` — EnsureBucket ensures the KV bucket exists.
- `func GetHostInfo(ctx context.Context, kv jetstream.KeyValue, hostID string) (*protocol.HostInfo, error)` — GetHostInfo retrieves host info.
- `func PutHeartbeat(ctx context.Context, kv jetstream.KeyValue, hb protocol.Heartbeat) error` — PutHeartbeat writes a heartbeat to KV.
- `func PutHostInfo(ctx context.Context, kv jetstream.KeyValue, info protocol.HostInfo) error` — PutHostInfo writes host info to KV.
- `func PutSessionInfo(ctx context.Context, kv jetstream.KeyValue, info protocol.SessionInfo) error` — PutSessionInfo writes session info to KV.

### Functions

#### EnsureBucket

```go
func EnsureBucket(ctx context.Context, js jetstream.JetStream, bucket string) (jetstream.KeyValue, error)
```

EnsureBucket ensures the KV bucket exists.

#### GetHostInfo

```go
func GetHostInfo(ctx context.Context, kv jetstream.KeyValue, hostID string) (*protocol.HostInfo, error)
```

GetHostInfo retrieves host info.

#### PutHeartbeat

```go
func PutHeartbeat(ctx context.Context, kv jetstream.KeyValue, hb protocol.Heartbeat) error
```

PutHeartbeat writes a heartbeat to KV.

#### PutHostInfo

```go
func PutHostInfo(ctx context.Context, kv jetstream.KeyValue, info protocol.HostInfo) error
```

PutHostInfo writes host info to KV.

#### PutSessionInfo

```go
func PutSessionInfo(ctx context.Context, kv jetstream.KeyValue, info protocol.SessionInfo) error
```

PutSessionInfo writes session info to KV.



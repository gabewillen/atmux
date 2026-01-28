# package daemon

`import "github.com/agentflare-ai/amux/internal/daemon"`

Package daemon implements the amux daemon (amuxd) and manager (amux-manager).

The daemon serves a JSON-RPC 2.0 control plane over a Unix socket,
manages agent lifecycles, and coordinates with remote hosts.

The role (director vs manager) is determined by the node.role configuration:
  - director: Runs the amux director with hub-mode NATS
  - manager: Runs as a host manager with leaf-mode NATS

See spec §12 for the full daemon specification.

- `Version` — Version is the daemon version string.
- `func Run(ctx context.Context, args []string) error` — Run starts the daemon with the given arguments.
- `func runDirector(ctx context.Context, cfg *config.Config) error`
- `func runManager(ctx context.Context, cfg *config.Config) error`
- `func showHelp() error`
- `func showVersion() error`

### Constants

#### Version

```go
const Version = "0.1.0-dev"
```

Version is the daemon version string.


### Functions

#### Run

```go
func Run(ctx context.Context, args []string) error
```

Run starts the daemon with the given arguments.

#### runDirector

```go
func runDirector(ctx context.Context, cfg *config.Config) error
```

#### runManager

```go
func runManager(ctx context.Context, cfg *config.Config) error
```

#### showHelp

```go
func showHelp() error
```

#### showVersion

```go
func showVersion() error
```



# package hub

`import "github.com/agentflare-ai/amux/internal/remote/hub"`

Package hub provides an embedded NATS hub server for the director role.

The director starts an embedded NATS server with JetStream enabled
and per-host authorization rules derived from auth.HostSubjectPermissions.

See spec §5.5.5 and §5.5.6 for hub server requirements.

- `func GenerateAuthConfig(prefix string, hosts map[string]string) string` — GenerateAuthConfig generates the NATS server authorization configuration for a set of hosts.
- `func WriteAuthConfig(dir, prefix string, hosts map[string]string) (string, error)` — WriteAuthConfig writes the authorization config to a file.
- `func parseListenAddr(addr string) (host string, port int, err error)` — parseListenAddr parses a listen address like "0.0.0.0:4222".
- `type AuthRule` — AuthRule holds per-host authorization permissions.
- `type Options` — Options configures the embedded hub server.
- `type Server` — Server wraps an embedded NATS server.

### Functions

#### GenerateAuthConfig

```go
func GenerateAuthConfig(prefix string, hosts map[string]string) string
```

GenerateAuthConfig generates the NATS server authorization configuration
for a set of hosts. This can be written to a file and loaded via include.

Per spec §5.5.6.4: "The hub MUST enforce per-host subject permissions."

#### WriteAuthConfig

```go
func WriteAuthConfig(dir, prefix string, hosts map[string]string) (string, error)
```

WriteAuthConfig writes the authorization config to a file.

#### parseListenAddr

```go
func parseListenAddr(addr string) (host string, port int, err error)
```

parseListenAddr parses a listen address like "0.0.0.0:4222".


## type AuthRule

```go
type AuthRule struct {
	PublicKey string
	Publish   []string
	Subscribe []string
}
```

AuthRule holds per-host authorization permissions.

## type Options

```go
type Options struct {
	// Listen is the address to listen on (e.g., "0.0.0.0:4222").
	Listen string

	// JetStreamDir is the directory for JetStream data.
	JetStreamDir string

	// AdvertiseURL is the URL to advertise for leaf connections.
	AdvertiseURL string
}
```

Options configures the embedded hub server.

### Functions returning Options

#### OptionsFromConfig

```go
func OptionsFromConfig(cfg *config.Config) *Options
```

OptionsFromConfig creates Options from the amux configuration.


## type Server

```go
type Server struct {
	mu        sync.Mutex
	ns        *natsserver.Server
	opts      *natsserver.Options
	authRules map[string]AuthRule
	prefix    string
	configDir string
}
```

Server wraps an embedded NATS server.

### Functions returning Server

#### Start

```go
func Start(opts *Options) (*Server, error)
```

Start creates and starts an embedded NATS server with JetStream.

Per spec §5.5.5: the director MUST start a NATS hub server
with JetStream enabled and leaf node support.


### Methods

#### Server.AddHostAuthorization

```go
func () AddHostAuthorization(publicKey string, publish, subscribe []string) error
```

AddHostAuthorization adds per-host authorization rules and reloads
the server configuration.

Per spec §5.5.6.4: the hub server MUST enforce that each host_id
can only publish/subscribe to its own subjects.

#### Server.ClientURL

```go
func () ClientURL() string
```

ClientURL returns the NATS client connection URL.

#### Server.Shutdown

```go
func () Shutdown()
```

Shutdown gracefully stops the embedded server.



# package integrationtest

`import "github.com/agentflare-ai/amux/internal/integrationtest"`

Package integrationtest provides docker/testcontainers helpers for integration tests.

- `ErrDockerUnavailable` — ErrDockerUnavailable indicates docker is not reachable for integration tests.
- `func safeNetworkNew(ctx context.Context) (dockernet *testcontainers.DockerNetwork, err error)`
- `natsImage, toxiproxyImage, natsPort, toxiproxyAPIPort, toxiproxyProxyPort, defaultHarnessGrace`
- `type Harness` — Harness manages docker/testcontainers infrastructure for integration tests.
- `type NATSContainer` — NATSContainer tracks a NATS container instance.
- `type ToxiproxyClient` — ToxiproxyClient configures network faults via the toxiproxy API.
- `type ToxiproxyContainer` — ToxiproxyContainer tracks a toxiproxy container instance.

### Constants

#### natsImage, toxiproxyImage, natsPort, toxiproxyAPIPort, toxiproxyProxyPort, defaultHarnessGrace

```go
const (
	natsImage           = "nats:2.12.4"
	toxiproxyImage      = "ghcr.io/shopify/toxiproxy:2.5.0"
	natsPort            = "4222/tcp"
	toxiproxyAPIPort    = "8474/tcp"
	toxiproxyProxyPort  = "8666/tcp"
	defaultHarnessGrace = 30 * time.Second
)
```


### Variables

#### ErrDockerUnavailable

```go
var ErrDockerUnavailable = errors.New("docker unavailable")
```

ErrDockerUnavailable indicates docker is not reachable for integration tests.


### Functions

#### safeNetworkNew

```go
func safeNetworkNew(ctx context.Context) (dockernet *testcontainers.DockerNetwork, err error)
```


## type Harness

```go
type Harness struct {
	ctx        context.Context
	cancel     context.CancelFunc
	network    *testcontainers.DockerNetwork
	containers []testcontainers.Container
}
```

Harness manages docker/testcontainers infrastructure for integration tests.

### Functions returning Harness

#### NewHarness

```go
func NewHarness(t testing.TB) (*Harness, error)
```

NewHarness creates a harness with an isolated docker network and cleanup.


### Methods

#### Harness.Close

```go
func () Close() error
```

Close terminates all containers and removes the network.

#### Harness.Context

```go
func () Context() context.Context
```

Context returns the harness context.

#### Harness.StartNATS

```go
func () StartNATS(ctx context.Context) (*NATSContainer, error)
```

StartNATS launches a NATS container with JetStream enabled.

#### Harness.StartToxiproxy

```go
func () StartToxiproxy(ctx context.Context) (*ToxiproxyContainer, error)
```

StartToxiproxy launches a toxiproxy container for network fault injection.

#### Harness.contextOrDefault

```go
func () contextOrDefault(ctx context.Context) context.Context
```


## type NATSContainer

```go
type NATSContainer struct {
	Container testcontainers.Container
	Host      string
	Port      nat.Port
	URL       string
	Alias     string
}
```

NATSContainer tracks a NATS container instance.

### Methods

#### NATSContainer.Start

```go
func () Start(ctx context.Context) error
```

Start restarts the NATS container.

#### NATSContainer.Stop

```go
func () Stop(ctx context.Context) error
```

Stop halts the NATS container.

#### NATSContainer.WaitReady

```go
func () WaitReady(ctx context.Context, timeout time.Duration) error
```

WaitReady waits until the NATS port is reachable.

#### NATSContainer.refreshEndpoint

```go
func () refreshEndpoint(ctx context.Context) error
```


## type ToxiproxyClient

```go
type ToxiproxyClient struct {
	BaseURL string
	Client  *http.Client
}
```

ToxiproxyClient configures network faults via the toxiproxy API.

### Functions returning ToxiproxyClient

#### NewToxiproxyClient

```go
func NewToxiproxyClient(baseURL string) *ToxiproxyClient
```

NewToxiproxyClient constructs a ToxiproxyClient for the given base URL.


### Methods

#### ToxiproxyClient.AddLatency

```go
func () AddLatency(ctx context.Context, name string, latency time.Duration, jitter time.Duration) error
```

AddLatency adds a latency toxic in milliseconds.

#### ToxiproxyClient.AddTimeout

```go
func () AddTimeout(ctx context.Context, name string, timeout time.Duration) error
```

AddTimeout adds a timeout toxic (useful to simulate loss).

#### ToxiproxyClient.CreateProxy

```go
func () CreateProxy(ctx context.Context, name string, listen string, upstream string) error
```

CreateProxy registers a new proxy.

#### ToxiproxyClient.SetProxyEnabled

```go
func () SetProxyEnabled(ctx context.Context, name string, enabled bool) error
```

SetProxyEnabled toggles a proxy on or off.

#### ToxiproxyClient.doJSON

```go
func () doJSON(ctx context.Context, method string, path string, payload any) ([]byte, error)
```


## type ToxiproxyContainer

```go
type ToxiproxyContainer struct {
	Container testcontainers.Container
	Host      string
	APIPort   nat.Port
	ProxyPort nat.Port
}
```

ToxiproxyContainer tracks a toxiproxy container instance.

### Methods

#### ToxiproxyContainer.APIURL

```go
func () APIURL() string
```

APIURL returns the toxiproxy API base URL.

#### ToxiproxyContainer.Client

```go
func () Client() *ToxiproxyClient
```

Client returns a toxiproxy API client for this container.

#### ToxiproxyContainer.ProxyAddress

```go
func () ProxyAddress() string
```

ProxyAddress returns the host:port address for proxy traffic.



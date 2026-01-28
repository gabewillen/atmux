Package main is the entry point for the amux CLI client.

The amux CLI communicates with the amux daemon (amuxd) over JSON-RPC 2.0
via a Unix socket. It provides commands for managing agents, plugins,
and running the test suite.

See spec §12 for the full CLI specification.

- `func SpecVersion() string` — SpecVersion returns the spec version this implementation targets.
- `func Version() string` — Version returns the amux version string.
- `func main()`

### Functions

#### SpecVersion

```go
func SpecVersion() string
```

SpecVersion returns the spec version this implementation targets.

#### Version

```go
func Version() string
```

Version returns the amux version string.

#### main

```go
func main()
```



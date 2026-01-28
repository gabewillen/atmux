Package main is the entry point for the amux unified node binary.

This binary serves as both the daemon (amuxd) and manager (amux-manager)
depending on configuration. The role is determined by the node.role
configuration option:
  - director: Runs the amux director with hub-mode NATS
  - manager: Runs as a host manager with leaf-mode NATS

See spec §3.44-§3.46 and §12 for the full specification.

- `func SpecVersion() string` — SpecVersion returns the spec version this implementation targets.
- `func Version() string` — Version returns the amux-node version string.
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

Version returns the amux-node version string.

#### main

```go
func main()
```



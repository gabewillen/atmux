Package main implements the Claude Code adapter for amux.
This adapter will be compiled to WASM using TinyGo.

Required exports per spec §8.2:
- amux_alloc
- amux_free
- manifest
- on_output
- format_input
- on_event

- `func amux_alloc(size uint32) uint32` — amux_alloc allocates memory for host-to-WASM communication.
- `func amux_free(ptr uint32)` — amux_free frees allocated memory.
- `func format_input(ptr, len uint32) uint64` — format_input formats input for the agent.
- `func main()`
- `func manifest() *byte` — manifest returns adapter metadata as JSON.
- `func on_event(ptr, len uint32) uint64` — on_event handles system events.
- `func on_output(ptr, len uint32) uint64` — on_output processes PTY output and returns events/actions.

### Functions

#### amux_alloc

```go
func amux_alloc(size uint32) uint32
```

amux_alloc allocates memory for host-to-WASM communication.

#### amux_free

```go
func amux_free(ptr uint32)
```

amux_free frees allocated memory.

#### format_input

```go
func format_input(ptr, len uint32) uint64
```

format_input formats input for the agent.

#### main

```go
func main()
```

#### manifest

```go
func manifest() *byte
```

manifest returns adapter metadata as JSON.

#### on_event

```go
func on_event(ptr, len uint32) uint64
```

on_event handles system events.

#### on_output

```go
func on_output(ptr, len uint32) uint64
```

on_output processes PTY output and returns events/actions.



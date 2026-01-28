Package main implements the Cursor adapter for amux.
This adapter will be compiled to WASM using TinyGo.

- `func amux_alloc(size uint32) uint32`
- `func amux_free(ptr uint32)`
- `func format_input(ptr, len uint32) uint64`
- `func main()`
- `func manifest() *byte`
- `func on_event(ptr, len uint32) uint64`
- `func on_output(ptr, len uint32) uint64`

### Functions

#### amux_alloc

```go
func amux_alloc(size uint32) uint32
```

#### amux_free

```go
func amux_free(ptr uint32)
```

#### format_input

```go
func format_input(ptr, len uint32) uint64
```

#### main

```go
func main()
```

#### manifest

```go
func manifest() *byte
```

#### on_event

```go
func on_event(ptr, len uint32) uint64
```

#### on_output

```go
func on_output(ptr, len uint32) uint64
```



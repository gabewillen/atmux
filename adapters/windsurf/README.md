Package main implements the Windsurf adapter for amux

- `func amux_alloc(size uint32) *byte`
- `func amux_free(ptr unsafe.Pointer, size uint32)`
- `func contains(s, substr string) bool` — Helper function to check if a string contains a substring
- `func find(s, substr string) bool` — Helper function to find a substring
- `func format_input(input_ptr uintptr, input_len uint32) uint64`
- `func main()`
- `func manifest() uint64`
- `func on_event(event_ptr uintptr, event_len uint32) uint64`
- `func on_output(output_ptr uintptr, output_len uint32) uint64`

### Functions

#### amux_alloc

```go
func amux_alloc(size uint32) *byte
```

#### amux_free

```go
func amux_free(ptr unsafe.Pointer, size uint32)
```

#### contains

```go
func contains(s, substr string) bool
```

Helper function to check if a string contains a substring

#### find

```go
func find(s, substr string) bool
```

Helper function to find a substring

#### format_input

```go
func format_input(input_ptr uintptr, input_len uint32) uint64
```

#### main

```go
func main()
```

#### manifest

```go
func manifest() uint64
```

#### on_event

```go
func on_event(event_ptr uintptr, event_len uint32) uint64
```

#### on_output

```go
func on_output(output_ptr uintptr, output_len uint32) uint64
```



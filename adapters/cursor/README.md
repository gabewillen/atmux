Package main implements the Cursor adapter for amux.
This adapter implements the WASM ABI per spec §10.4 for Cursor agent integration.

- `allocatedBlocks, nextPtr` — Memory management for the WASM ABI
- `func amux_alloc(size uint32) uint32`
- `func amux_free(ptr, size uint32)`
- `func containsPattern(output, pattern string) bool`
- `func findSubstring(s, substr string) int`
- `func format_input(ptr, len uint32) uint64`
- `func main()`
- `func manifest() uint64`
- `func on_event(ptr, len uint32) uint64`
- `func on_output(ptr, len uint32) uint64`
- `func packPtr(data []byte) uint64` — Helper functions
- `func readInput(ptr, len uint32) []byte`
- `type AdapterCommands`
- `type AdapterManifest` — AdapterManifest describes this adapter's capabilities per spec §10.2
- `type AdapterPatterns`
- `type CLIRequirement`

### Variables

#### allocatedBlocks, nextPtr

```go
var (
	allocatedBlocks = make(map[uint32][]byte)
	nextPtr         = uint32(1000)
)
```

Memory management for the WASM ABI


### Functions

#### amux_alloc

```go
func amux_alloc(size uint32) uint32
```

#### amux_free

```go
func amux_free(ptr, size uint32)
```

#### containsPattern

```go
func containsPattern(output, pattern string) bool
```

#### findSubstring

```go
func findSubstring(s, substr string) int
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
func manifest() uint64
```

#### on_event

```go
func on_event(ptr, len uint32) uint64
```

#### on_output

```go
func on_output(ptr, len uint32) uint64
```

#### packPtr

```go
func packPtr(data []byte) uint64
```

Helper functions

#### readInput

```go
func readInput(ptr, len uint32) []byte
```


## type AdapterCommands

```go
type AdapterCommands struct {
	Start       []string `json:"start"`
	SendMessage string   `json:"send_message"`
}
```

## type AdapterManifest

```go
type AdapterManifest struct {
	Name        string          `json:"name"`
	Version     string          `json:"version"`
	Description string          `json:"description,omitempty"`
	CLI         CLIRequirement  `json:"cli"`
	Patterns    AdapterPatterns `json:"patterns"`
	Commands    AdapterCommands `json:"commands"`
}
```

AdapterManifest describes this adapter's capabilities per spec §10.2

## type AdapterPatterns

```go
type AdapterPatterns struct {
	Ready    string `json:"ready"`
	Error    string `json:"error"`
	Complete string `json:"complete"`
}
```

## type CLIRequirement

```go
type CLIRequirement struct {
	Binary     string `json:"binary"`
	VersionCmd string `json:"version_cmd"`
	VersionRe  string `json:"version_re"`
	Constraint string `json:"constraint"`
}
```


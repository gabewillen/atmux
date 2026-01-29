Package main implements the Claude Code adapter for amux.
This adapter implements the WASM ABI per spec §10.4 for Claude Code agent integration.

Required exports per spec §10.4.2:
- amux_alloc
- amux_free
- manifest
- on_output
- format_input
- on_event

- `allocatedBlocks, nextPtr` — Memory management for the WASM ABI
- `func amux_alloc(size uint32) uint32` — amux_alloc allocates memory for host-to-WASM communication per spec §10.4.1
- `func amux_free(ptr, size uint32)` — amux_free frees allocated memory per spec §10.4.1
- `func containsPattern(output, pattern string) bool` — containsPattern checks if output contains a pattern (simple substring match)
- `func findSubstring(s, substr string) int` — findSubstring finds substring without using strings package (WASM compatibility)
- `func format_input(ptr, len uint32) uint64` — format_input formats input for the Claude Code agent
- `func main()` — main is required for Go WASM modules but not called in the WASM context
- `func manifest() uint64` — manifest returns adapter metadata as JSON per spec §10.2
- `func on_event(ptr, len uint32) uint64` — on_event handles system events
- `func on_output(ptr, len uint32) uint64` — on_output processes PTY output and returns events/actions per spec §10.4.3
- `func packPtr(data []byte) uint64` — packPtr packs data into memory and returns packed (ptr << 32 | len) per spec §10.4.1
- `func readInput(ptr, len uint32) []byte` — readInput reads input data from WASM memory
- `type AdapterCommands`
- `type AdapterManifest` — AdapterManifest describes this adapter's capabilities per spec §10.2
- `type AdapterPatterns`
- `type CLIRequirement`

### Variables

#### allocatedBlocks, nextPtr

```go
var (
	allocatedBlocks = make(map[uint32][]byte)
	nextPtr         = uint32(1000) // Start at a safe offset
)
```

Memory management for the WASM ABI


### Functions

#### amux_alloc

```go
func amux_alloc(size uint32) uint32
```

amux_alloc allocates memory for host-to-WASM communication per spec §10.4.1

#### amux_free

```go
func amux_free(ptr, size uint32)
```

amux_free frees allocated memory per spec §10.4.1

#### containsPattern

```go
func containsPattern(output, pattern string) bool
```

containsPattern checks if output contains a pattern (simple substring match)

#### findSubstring

```go
func findSubstring(s, substr string) int
```

findSubstring finds substring without using strings package (WASM compatibility)

#### format_input

```go
func format_input(ptr, len uint32) uint64
```

format_input formats input for the Claude Code agent

#### main

```go
func main()
```

main is required for Go WASM modules but not called in the WASM context

#### manifest

```go
func manifest() uint64
```

manifest returns adapter metadata as JSON per spec §10.2

#### on_event

```go
func on_event(ptr, len uint32) uint64
```

on_event handles system events

#### on_output

```go
func on_output(ptr, len uint32) uint64
```

on_output processes PTY output and returns events/actions per spec §10.4.3

#### packPtr

```go
func packPtr(data []byte) uint64
```

packPtr packs data into memory and returns packed (ptr << 32 | len) per spec §10.4.1

#### readInput

```go
func readInput(ptr, len uint32) []byte
```

readInput reads input data from WASM memory


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


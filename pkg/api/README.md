# package api

`import "github.com/agentflare-ai/amux/pkg/api"`

Package api defines public types shared between amux clients and the daemon.

The API types are stable, JSON-serializable, and enforce the wire conventions
required by the amux specification.

- `defaultIDGenerator`
- `type AdapterRef` — AdapterRef is the string name of an adapter loaded from the WASM registry.
- `type ID` — ID is a JSON-safe wrapper around muid.MUID that encodes as base-10 strings.

### Variables

#### defaultIDGenerator

```go
var defaultIDGenerator = muid.NewGenerator(muid.DefaultConfig(), 0, 0)
```


## type AdapterRef

```go
type AdapterRef string
```

AdapterRef is the string name of an adapter loaded from the WASM registry.

## type ID

```go
type ID struct {
	value muid.MUID
}
```

ID is a JSON-safe wrapper around muid.MUID that encodes as base-10 strings.

### Functions returning ID

#### MustParseID

```go
func MustParseID(raw string) ID
```

MustParseID parses a base-10 encoded ID string and panics on failure.

#### NewID

```go
func NewID() ID
```

NewID creates a new non-zero ID suitable for runtime use.

#### ParseID

```go
func ParseID(raw string) (ID, error)
```

ParseID parses a base-10 encoded ID string.


### Methods

#### ID.MarshalJSON

```go
func () MarshalJSON() ([]byte, error)
```

MarshalJSON encodes the ID as a JSON string containing a base-10 integer.

#### ID.String

```go
func () String() string
```

String returns the base-10 string form of the ID.

#### ID.UnmarshalJSON

```go
func () UnmarshalJSON(data []byte) error
```

UnmarshalJSON decodes a JSON string containing a base-10 integer ID.

#### ID.Value

```go
func () Value() muid.MUID
```

Value returns the underlying muid.MUID.



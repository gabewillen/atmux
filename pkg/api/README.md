# package api

`import "github.com/stateforward/amux/pkg/api"`

Package api contains public API types (Agent.Adapter is a string)

- `type Agent` — Agent represents an agent instance
- `type ID` — ID represents a unique identifier for entities
- `type Session` — Session represents an amux session

## type Agent

```go
type Agent struct {
	ID       ID     `json:"id"`
	Name     string `json:"name"`
	Adapter  string `json:"adapter"`
	Location string `json:"location,omitempty"`
}
```

Agent represents an agent instance

## type ID

```go
type ID uint64
```

ID represents a unique identifier for entities

### Methods

#### ID.MarshalJSON

```go
func () MarshalJSON() ([]byte, error)
```

MarshalJSON implements json.Marshaler for ID

#### ID.UnmarshalJSON

```go
func () UnmarshalJSON(data []byte) error
```

UnmarshalJSON implements json.Unmarshaler for ID


## type Session

```go
type Session struct {
	ID     ID      `json:"id"`
	Agents []Agent `json:"agents"`
}
```

Session represents an amux session


# package api

`import "github.com/stateforward/amux/pkg/api"`

Package api contains public API types (Agent.Adapter is a string)

- `type Agent` — Agent represents an agent instance
- `type Session` — Session represents an amux session

## type Agent

```go
type Agent struct {
	ID       muid.ID `json:"id"`
	Name     string  `json:"name"`
	Adapter  string  `json:"adapter"`
	Location string  `json:"location,omitempty"`
}
```

Agent represents an agent instance

## type Session

```go
type Session struct {
	ID     muid.ID `json:"id"`
	Agents []Agent `json:"agents"`
}
```

Session represents an amux session


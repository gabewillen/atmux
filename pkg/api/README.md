# package api

`import "github.com/stateforward/amux/pkg/api"`

Package api contains public API types (Agent.Adapter is a string)

- `type Agent` — Agent represents an agent instance
- `type Location` — Location represents where an agent runs
- `type Session` — Session represents an amux session

## type Agent

```go
type Agent struct {
	ID       muid.MUID `json:"id"`
	Name     string    `json:"name"`
	Adapter  string    `json:"adapter"`
	Location string    `json:"location,omitempty"`
}
```

Agent represents an agent instance

## type Location

```go
type Location struct {
	Type     string `json:"type"`                // "local" or "ssh"
	RepoPath string `json:"repo_path,omitempty"` // Path to the repository on the host
	Host     string `json:"host,omitempty"`      // For SSH locations
	User     string `json:"user,omitempty"`      // For SSH locations
	KeyPath  string `json:"key_path,omitempty"`  // For SSH locations
}
```

Location represents where an agent runs

## type Session

```go
type Session struct {
	ID     muid.MUID `json:"id"`
	Agents []Agent   `json:"agents"`
}
```

Session represents an amux session


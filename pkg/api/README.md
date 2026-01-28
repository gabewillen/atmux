# package api

`import "github.com/agentflare-ai/amux/pkg/api"`

- `type Agent` — Agent represents a managed agent instance.
- `type Location` — Location defines where the agent runs.
- `type Session` — Session represents an active PTY session.

## type Agent

```go
type Agent struct {
	ID       muid.MUID `json:"id"`
	Name     string    `json:"name"`
	About    string    `json:"about,omitempty"`
	Adapter  string    `json:"adapter"` // String reference to adapter name
	RepoPath string    `json:"repo_path,omitempty"`
	Location Location  `json:"location"`
	Presence string    `json:"presence"` // Online, Busy, Offline, Away
	Status   string    `json:"status"`   // Pending, Starting, Running, Terminated, Errored
}
```

Agent represents a managed agent instance.

## type Location

```go
type Location struct {
	Type string `json:"type"` // "local", "ssh"
	Host string `json:"host,omitempty"`
}
```

Location defines where the agent runs.

## type Session

```go
type Session struct {
	ID        muid.MUID `json:"id"`
	AgentID   muid.MUID `json:"agent_id"`
	HostID    string    `json:"host_id"` // Node ID where session runs
	StartedAt time.Time `json:"started_at"`
}
```

Session represents an active PTY session.


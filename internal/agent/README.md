# package agent

`import "github.com/agentflare-ai/amux/internal/agent"`

Package agent manages agent lifecycle and presence state machines.

- `type Agent` — Agent represents a runtime agent instance.

## type Agent

```go
type Agent struct {
	hsm.HSM
	ID      api.ID
	Adapter adapter.Adapter
}
```

Agent represents a runtime agent instance.

### Functions returning Agent

#### NewAgent

```go
func NewAgent(adapter adapter.Adapter) *Agent
```

NewAgent constructs a new agent with a fresh ID.


### Methods

#### Agent.Start

```go
func () Start(ctx context.Context, model *hsm.Model)
```

Start starts the agent state machine.



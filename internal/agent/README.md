# package agent

`import "github.com/agentflare-ai/amux/internal/agent"`

- `Model` — Model defines the agent state machine.

### Variables

#### Model

```go
var Model = hsm.Define("agent",
	hsm.State("pending"),
	hsm.Initial(hsm.Target("pending")),
)
```

Model defines the agent state machine.
Phase 1 will implement the full logic.



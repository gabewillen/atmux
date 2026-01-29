package coordination

import (
	"context"
	"fmt"

	"github.com/agentflare-ai/amux/internal/agent"
)

// Executor handles action execution.
type Executor struct {
	Agent *agent.Agent
}

// NewExecutor creates a new executor.
func NewExecutor(agent *agent.Agent) *Executor {
	return &Executor{Agent: agent}
}

// Execute performs the action.
func (e *Executor) Execute(ctx context.Context, action Action) error {
	switch action.Type {
	case "type":
		text := action.Payload["text"]
		// Inject into PTY
		// We need access to PTY.
		// e.Agent.Sessions...
		return e.injectInput(text)
	case "exec":
		cmd := action.Payload["command"]
		// Exec separate command? Or in PTY?
		// "Tool invocation"
		fmt.Printf("Executing tool: %s\n", cmd)
		return nil
	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}

func (e *Executor) injectInput(text string) error {
	// Find active session
	for _, s := range e.Agent.Sessions {
		if s.PTY != nil {
			_, err := s.PTY.Write([]byte(text))
			return err
		}
	}
	return fmt.Errorf("no active session")
}

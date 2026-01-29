package agent

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/creack/pty"
	"github.com/stateforward/hsm-go"
	"github.com/stateforward/hsm-go/muid"
)

// SpawnAgent starts the agent process in a new PTY session.
func SpawnAgent(ctx context.Context, a *Agent) error {
	// Dispatch Spawn event to lifecycle HSM
	// Note: hsm.Dispatch returns a channel that closes when processing is done.
	<-hsm.Dispatch(ctx, a.Lifecycle, hsm.Event{Name: EventSpawn})

	// Check if state transitioned to Starting (or verify logic via HSM state check if possible)
	// For now we assume if no error, we proceed.

	// Ensure Worktree
	worktreePath, err := EnsureWorktree(a.RepoRoot, a.Slug, "main") // TODO: get target branch from config
	if err != nil {
		hsm.Dispatch(ctx, a.Lifecycle, hsm.Event{Name: EventError})
		return fmt.Errorf("failed to ensure worktree: %w", err)
	}

	// Prepare command
	// Ideally, the command comes from the adapter or config.
	// For Phase 2, we just spawn a shell or a placeholder since we don't have the adapter runtime yet.
	// Let's use /bin/sh for local agent smoke testing if no command specified.
	cmdName := "/bin/sh" 
	// In real implementation, we'd query the adapter for the command.
	
	cmd := exec.Command(cmdName)
	cmd.Dir = worktreePath
	// Inherit or set env
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	// Start PTY
	ptmx, err := pty.Start(cmd)
	if err != nil {
		hsm.Dispatch(ctx, a.Lifecycle, hsm.Event{Name: EventError})
		return fmt.Errorf("failed to start pty: %w", err)
	}

	// Create Session
	sessionID := api.SessionID(muid.Make())
	session := &Session{
		ID:        sessionID,
		AgentID:   a.ID,
		StartedAt: time.Now(),
		Cmd:       cmd,
		PTY:       ptmx,
	}
	
	// Store session
	a.Sessions[sessionID] = session

	hsm.Dispatch(ctx, a.Lifecycle, hsm.Event{Name: EventStarted})

	// Cleanup goroutine
	go func() {
		// Wait for process
		cmd.Wait()
		
		// Close PTY
		ptmx.Close()
		
		// Remove session? Or keep as history?
		// For now, keep it but mark agent as Terminated.
		hsm.Dispatch(context.Background(), a.Lifecycle, hsm.Event{Name: EventExited})
	}()

	return nil
}

// StopAgent stops the agent process.
func StopAgent(ctx context.Context, a *Agent) error {
	// Dispatch Stop
	<-hsm.Dispatch(ctx, a.Lifecycle, hsm.Event{Name: EventStop})
	
	for _, s := range a.Sessions {
		if s.Cmd != nil && s.Cmd.Process != nil {
			// Graceful kill
			s.Cmd.Process.Signal(os.Interrupt)
			
			// Wait a bit then force kill?
			// For simplicity in Phase 2, just Kill if not waiting.
			// Ideally we use a select with timeout.
			go func(proc *os.Process) {
				time.Sleep(1 * time.Second)
				proc.Kill()
			}(s.Cmd.Process)
		}
	}
	
	return nil
}

package agent

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
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

	// Determine target branch
	// Ideally this comes from config or default logic.
	// For now, let's use the repo's HEAD as base, which EnsureWorktree defaults to if empty.
	targetBranch, _ := GetAgentTargetBranch(a) 

	// Ensure Worktree
	worktreePath, err := EnsureWorktree(a.RepoRoot, a.Slug, targetBranch)
	if err != nil {
		hsm.Dispatch(ctx, a.Lifecycle, hsm.Event{Name: EventError})
		return fmt.Errorf("failed to ensure worktree: %w", err)
	}
	a.Worktree = worktreePath

	// Prepare command
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
		
		// Dispatch Exited
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
			pid := s.Cmd.Process.Pid
			
			// Try to kill process group
			// We use negative PID to signal the process group.
			// pty.Start creates a session leader, so the PGID is the PID.
			if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil {
				// Fallback to process only if group fails
				_ = s.Cmd.Process.Signal(syscall.SIGTERM)
			}
			
			// Schedule force kill
			go func(p int) {
				time.Sleep(5 * time.Second)
				// Force kill group
				_ = syscall.Kill(-p, syscall.SIGKILL)
			}(pid)
		}
	}
	
	return nil
}

// GetAgentTargetBranch helps resolve the branch to use for the worktree.
// This duplicates logic from SelectMergeStrategy a bit but for worktree creation.
func GetAgentTargetBranch(a *Agent) (string, error) {
	// Assuming GitConfig is available globally or we can peek at project config?
	// But 'a.Config' is AgentConfig.
	// We might need to look at repo root.
	// For now return empty to let EnsureWorktree use HEAD.
	return "", nil
}
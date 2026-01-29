package agent

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/creack/pty"
	"github.com/stateforward/hsm-go"
	"github.com/stateforward/hsm-go/muid"
)

// SpawnAgent starts the agent process in a new PTY session.
func SpawnAgent(ctx context.Context, a *Agent) error {
	// Dispatch Spawn event to lifecycle HSM
	<-hsm.Dispatch(ctx, a.Lifecycle, hsm.Event{Name: EventSpawn})

	// Check if state transitioned to Starting (or verify logic via HSM state check if possible)
	// For now we assume if no error, we proceed.

	// Determine target branch
	targetBranch, err := GetAgentTargetBranch(a)
	if err != nil {
		hsm.Dispatch(ctx, a.Lifecycle, hsm.Event{Name: EventError})
		return fmt.Errorf("failed to resolve target branch: %w", err)
	}

	// Ensure Worktree
	worktreePath, err := EnsureWorktree(a.RepoRoot, a.Slug, targetBranch)
	if err != nil {
		hsm.Dispatch(ctx, a.Lifecycle, hsm.Event{Name: EventError})
		return fmt.Errorf("failed to ensure worktree: %w", err)
	}
	a.Worktree = worktreePath

	// Prepare command
	cmdArgs := getAgentCommand(a)
	if len(cmdArgs) == 0 {
		hsm.Dispatch(ctx, a.Lifecycle, hsm.Event{Name: EventError})
		return fmt.Errorf("agent command not resolved for adapter %s", a.Adapter)
	}
	
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
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
func GetAgentTargetBranch(a *Agent) (string, error) {
	// Spec §5.7.1: determine base_branch by running git symbolic-ref --quiet --short HEAD in repo_root
	// We run this in the RepoRoot.
	cmd := exec.Command("git", "symbolic-ref", "--quiet", "--short", "HEAD")
	cmd.Dir = string(a.RepoRoot)
	out, err := cmd.Output()
	if err != nil {
		// Spec says: "If this command fails... base_branch MUST be set to the configured git.merge.target_branch... otherwise fail"
		// We assume `a.Config` might have this info if we mapped it, but `AgentConfig` doesn't strictly have `MergeConfig`.
		// If we can't determine it, we fail.
		return "", fmt.Errorf("failed to determine base branch (HEAD detached?): %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func getAgentCommand(a *Agent) []string {
	// In a real implementation, this would come from the adapter manifest (CLIConfig)
	// or configuration. For now, we default to the adapter name as the binary.
	return []string{a.Adapter}
}

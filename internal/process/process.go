// Package process provides process tracking and interception for amux.
//
// The process tracker monitors child processes spawned within agent PTYs,
// capturing spawn/exit events and optionally intercepting I/O.
//
// See spec §8 for process tracking requirements.
package process

import (
	"context"
	"sync"
	"time"

	"github.com/agentflare-ai/amux/internal/event"
	"github.com/stateforward/hsm-go/muid"
)

// Process represents a tracked process.
type Process struct {
	// PID is the process ID.
	PID int

	// PPID is the parent process ID.
	PPID int

	// Command is the command that started the process.
	Command string

	// Args are the command arguments.
	Args []string

	// AgentID is the agent that spawned the process.
	AgentID muid.MUID

	// StartedAt is when the process started.
	StartedAt time.Time

	// ExitCode is the exit code (nil if still running).
	ExitCode *int

	// ExitedAt is when the process exited (zero if still running).
	ExitedAt time.Time
}

// Tracker tracks processes for an agent.
type Tracker struct {
	mu         sync.RWMutex
	agentID    muid.MUID
	processes  map[int]*Process
	dispatcher event.Dispatcher
}

// NewTracker creates a new process tracker.
func NewTracker(agentID muid.MUID, dispatcher event.Dispatcher) *Tracker {
	if dispatcher == nil {
		dispatcher = event.NewNoopDispatcher()
	}

	return &Tracker{
		agentID:    agentID,
		processes:  make(map[int]*Process),
		dispatcher: dispatcher,
	}
}

// Add adds a process to tracking.
func (t *Tracker) Add(ctx context.Context, p *Process) {
	t.mu.Lock()
	t.processes[p.PID] = p
	t.mu.Unlock()

	// Emit spawn event
	_ = t.dispatcher.Dispatch(ctx, event.NewEvent(event.TypeProcessSpawned, t.agentID, p))
}

// Remove removes a process from tracking.
func (t *Tracker) Remove(ctx context.Context, pid int, exitCode int) {
	t.mu.Lock()
	p, ok := t.processes[pid]
	if ok {
		p.ExitCode = &exitCode
		p.ExitedAt = time.Now()
		delete(t.processes, pid)
	}
	t.mu.Unlock()

	if ok {
		// Emit completed event
		_ = t.dispatcher.Dispatch(ctx, event.NewEvent(event.TypeProcessCompleted, t.agentID, p))
	}
}

// Get returns a process by PID.
func (t *Tracker) Get(pid int) *Process {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.processes[pid]
}

// List returns all tracked processes.
func (t *Tracker) List() []*Process {
	t.mu.RLock()
	defer t.mu.RUnlock()

	processes := make([]*Process, 0, len(t.processes))
	for _, p := range t.processes {
		processes = append(processes, p)
	}
	return processes
}

// Count returns the number of tracked processes.
func (t *Tracker) Count() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.processes)
}

// Clear removes all tracked processes.
func (t *Tracker) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.processes = make(map[int]*Process)
}

// CaptureMode represents the I/O capture mode.
type CaptureMode string

const (
	// CaptureModeNone disables I/O capture.
	CaptureModeNone CaptureMode = "none"

	// CaptureModeStdout captures stdout only.
	CaptureModeStdout CaptureMode = "stdout"

	// CaptureModeStderr captures stderr only.
	CaptureModeStderr CaptureMode = "stderr"

	// CaptureModeStdin captures stdin only.
	CaptureModeStdin CaptureMode = "stdin"

	// CaptureModeAll captures all streams.
	CaptureModeAll CaptureMode = "all"
)

// HookMode represents the process hook mode.
type HookMode string

const (
	// HookModeAuto automatically selects preload or polling.
	HookModeAuto HookMode = "auto"

	// HookModePreload uses LD_PRELOAD/DYLD_INSERT_LIBRARIES.
	HookModePreload HookMode = "preload"

	// HookModePolling uses periodic polling.
	HookModePolling HookMode = "polling"

	// HookModeDisabled disables process tracking.
	HookModeDisabled HookMode = "disabled"
)

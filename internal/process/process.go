package process

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/agentflare-ai/amux/pkg/api"
)

// EventType for process events.
type EventType string

const (
	EventSpawned   EventType = "process.spawned"
	EventExited    EventType = "process.exited"
	EventIO        EventType = "process.io"
)

// Event represents a process event.
type Event struct {
	Type      EventType   `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

// Process represents a tracked process.
type Process struct {
	PID       int           `json:"pid"`
	AgentID   api.AgentID   `json:"agent_id"`
	ProcessID api.ProcessID `json:"process_id"`
	Command   string        `json:"command"`
	Args      []string      `json:"args"`
	WorkDir   string        `json:"work_dir"`
	ParentPID int           `json:"parent_pid"`
	StartedAt time.Time     `json:"started_at"`
	EndedAt   time.Time     `json:"ended_at,omitempty"`
	ExitCode  int           `json:"exit_code,omitempty"`
	Running   bool          `json:"running"`
}

// Tracker manages the process tree.
type Tracker struct {
	mu         sync.RWMutex
	processes  map[int]*Process
	Events     chan Event
	Gater      *Gater
	SocketPath string
}

// NewTracker creates a new process tracker.
func NewTracker() *Tracker {
	return &Tracker{
		processes: make(map[int]*Process),
		Events:    make(chan Event, 1000), // Buffer events
	}
}

// TrackSpawn records a new process start.
func (t *Tracker) TrackSpawn(proc *Process) {
	t.mu.Lock()
	defer t.mu.Unlock()

	proc.Running = true
	t.processes[proc.PID] = proc

	evt := Event{
		Type:      EventSpawned,
		Timestamp: time.Now(),
		Payload:   proc,
	}

	if t.Gater == nil || t.Gater.ShouldNotify(context.Background(), evt) {
		t.Events <- evt
	}
}

// TrackExit records a process exit.
func (t *Tracker) TrackExit(pid int, exitCode int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	proc, ok := t.processes[pid]
	if !ok {
		return fmt.Errorf("process %d not found", pid)
	}

	proc.Running = false
	proc.ExitCode = exitCode
	proc.EndedAt = time.Now()

	evt := Event{
		Type:      EventExited,
		Timestamp: time.Now(),
		Payload:   proc,
	}

	if t.Gater == nil || t.Gater.ShouldNotify(context.Background(), evt) {
		t.Events <- evt
	}
	return nil
}

// GetProcess returns a copy of the process info.
func (t *Tracker) GetProcess(pid int) (*Process, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	p, ok := t.processes[pid]
	if !ok {
		return nil, false
	}
	// Copy
	cp := *p
	return &cp, true
}

// Start initiates the hook server.
func (t *Tracker) Start(ctx context.Context, socketDir string) error {
	if err := t.StartHookServer(ctx, socketDir); err != nil {
		return err
	}
	// SocketPath is set by StartHookServer? 
	// StartHookServer needs to update t.SocketPath.
	// We'll update StartHookServer logic to do so if not already.
	// Re-reading hook_server.go: It calculates socketPath but doesn't set t.SocketPath.
	// We need to fix hook_server.go to set t.SocketPath or return it.
	// For now, let's assume StartHookServer sets it (I will fix hook_server.go next).
	return nil
}
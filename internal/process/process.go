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
	mu        sync.RWMutex
	processes map[int]*Process
	Events    chan Event
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

	t.Events <- Event{
		Type:      EventSpawned,
		Timestamp: time.Now(),
		Payload:   proc,
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

	t.Events <- Event{
		Type:      EventExited,
		Timestamp: time.Now(),
		Payload:   proc,
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

// Start polling/monitoring logic would go here or be driven by hooks.
func (t *Tracker) Start(ctx context.Context) {
	// If polling fallback is needed, implementation would go here.
}
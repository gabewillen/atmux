// Package process implements process tracking and interception (generic)
package process

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/stateforward/hsm-go/muid"
)

// Process represents a tracked process
type Process struct {
	ID          muid.MUID           `json:"id"`
	PID         int                 `json:"pid"`
	Command     string              `json:"command"`
	Args        []string            `json:"args"`
	StartTime   time.Time           `json:"start_time"`
	EndTime     *time.Time          `json:"end_time,omitempty"`
	Status      ProcessStatus       `json:"status"`
	ExitCode    *int                `json:"exit_code,omitempty"`
	ParentID    *muid.MUID          `json:"parent_id,omitempty"`
	Children    []muid.MUID         `json:"children,omitempty"`
	Environment map[string]string   `json:"environment,omitempty"`
}

// ProcessStatus represents the status of a process
type ProcessStatus string

const (
	ProcessRunning   ProcessStatus = "running"
	ProcessCompleted ProcessStatus = "completed"
	ProcessErrored   ProcessStatus = "errored"
	ProcessKilled    ProcessStatus = "killed"
)

// Tracker tracks processes and their relationships
type Tracker struct {
	mu        sync.RWMutex
	processes map[muid.MUID]*Process
	commands  map[int]muid.MUID // PID -> Process ID mapping
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewTracker creates a new process tracker
func NewTracker() *Tracker {
	ctx, cancel := context.WithCancel(context.Background())
	tracker := &Tracker{
		processes: make(map[muid.MUID]*Process),
		commands:  make(map[int]muid.MUID),
		ctx:       ctx,
		cancel:    cancel,
	}
	
	// Start background cleanup routine
	go tracker.cleanupRoutine()
	
	return tracker
}

// TrackCommand starts tracking a command
func (t *Tracker) TrackCommand(cmd *exec.Cmd) (muid.MUID, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Generate a unique ID for this process
	id := muid.Make()
	
	// Ensure the ID is not zero (reserved sentinel value)
	for uint64(id) == 0 {
		id = muid.Make()
	}

	// Create a new process record
	process := &Process{
		ID:        id,
		PID:       cmd.Process.Pid,
		Command:   cmd.Path,
		Args:      cmd.Args,
		StartTime: time.Now(),
		Status:    ProcessRunning,
		Environment: make(map[string]string),
	}

	// Capture environment variables if available
	for _, envVar := range cmd.Env {
		parts := splitEnvVar(envVar)
		if len(parts) == 2 {
			process.Environment[parts[0]] = parts[1]
		}
	}

	// Store the process
	t.processes[id] = process
	t.commands[cmd.Process.Pid] = id

	return id, nil
}

// GetProcess retrieves a process by ID
func (t *Tracker) GetProcess(id muid.MUID) (*Process, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	process, exists := t.processes[id]
	if !exists {
		return nil, fmt.Errorf("process with ID %d not found", uint64(id))
	}

	return process, nil
}

// GetProcessByPID retrieves a process by PID
func (t *Tracker) GetProcessByPID(pid int) (*Process, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	id, exists := t.commands[pid]
	if !exists {
		return nil, fmt.Errorf("process with PID %d not found", pid)
	}

	process, exists := t.processes[id]
	if !exists {
		return nil, fmt.Errorf("process with ID %d not found", uint64(id))
	}

	return process, nil
}

// UpdateProcessStatus updates the status of a process
func (t *Tracker) UpdateProcessStatus(id muid.MUID, status ProcessStatus) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	process, exists := t.processes[id]
	if !exists {
		return fmt.Errorf("process with ID %d not found", uint64(id))
	}

	now := time.Now()
	process.Status = status

	// Update end time if the process is no longer running
	if status != ProcessRunning {
		process.EndTime = &now
	}

	return nil
}

// RecordProcessExit records the exit of a process
func (t *Tracker) RecordProcessExit(id muid.MUID, exitCode int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	process, exists := t.processes[id]
	if !exists {
		return fmt.Errorf("process with ID %d not found", uint64(id))
	}

	now := time.Now()
	process.Status = ProcessCompleted
	process.EndTime = &now
	process.ExitCode = &exitCode

	return nil
}

// ListProcesses returns a list of all tracked processes
func (t *Tracker) ListProcesses() []*Process {
	t.mu.RLock()
	defer t.mu.RUnlock()

	processes := make([]*Process, 0, len(t.processes))
	for _, process := range t.processes {
		// Create a copy to prevent concurrent modification
		pCopy := *process
		processes = append(processes, &pCopy)
	}

	return processes
}

// cleanupRoutine periodically cleans up completed processes
func (t *Tracker) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute) // Clean up every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-t.ctx.Done():
			return
		case <-ticker.C:
			t.cleanupCompletedProcesses()
		}
	}
}

// cleanupCompletedProcesses removes completed processes older than 10 minutes
func (t *Tracker) cleanupCompletedProcesses() {
	t.mu.Lock()
	defer t.mu.Unlock()

	cutoffTime := time.Now().Add(-10 * time.Minute)
	
	for id, process := range t.processes {
		if process.Status != ProcessRunning && 
		   process.EndTime != nil && 
		   process.EndTime.Before(cutoffTime) {
			delete(t.processes, id)
			if process.PID != 0 {
				delete(t.commands, process.PID)
			}
		}
	}
}

// Stop stops the tracker and cleans up resources
func (t *Tracker) Stop() {
	t.cancel()
}

// splitEnvVar splits an environment variable string into key and value
func splitEnvVar(envVar string) []string {
	parts := make([]string, 2)
	eqIndex := -1
	for i, char := range envVar {
		if char == '=' {
			eqIndex = i
			break
		}
	}
	
	if eqIndex == -1 {
		parts[0] = envVar
		parts[1] = ""
	} else {
		parts[0] = envVar[:eqIndex]
		parts[1] = envVar[eqIndex+1:]
	}
	
	return parts
}
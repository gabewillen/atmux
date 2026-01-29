package process

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go/muid"
)

// HookMessage matches the JSON sent by hook.c
type HookMessage struct {
	PID  int    `json:"pid"`
	PPID int    `json:"ppid"`
	Cmd  string `json:"cmd"`
}

// StartHookServer starts listening on a Unix socket for exec notifications.
func (t *Tracker) StartHookServer(ctx context.Context, socketDir string) error {
	if err := os.MkdirAll(socketDir, 0755); err != nil {
		return fmt.Errorf("failed to create socket dir: %w", err)
	}

	// Unique socket name per tracker instance
	socketPath := filepath.Join(socketDir, fmt.Sprintf("amux-hook-%d.sock", os.Getpid()))
	
	// Remove if exists
	os.Remove(socketPath)

	l, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on unix socket: %w", err)
	}

	t.mu.Lock()
	t.SocketPath = socketPath
	t.mu.Unlock()

	go func() {
		<-ctx.Done()
		l.Close()
		os.Remove(socketPath)
	}()

	go t.acceptLoop(ctx, l)
	
	return nil
}

func (t *Tracker) acceptLoop(ctx context.Context, l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			// Log error?
			continue
		}
		go t.handleConnection(conn)
	}
}

func (t *Tracker) handleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var msg HookMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue
		}
		
		t.TrackSpawn(&Process{
			PID:       msg.PID,
			ParentPID: msg.PPID,
			Command:   msg.Cmd,
			// AgentID/ProcessID need to be inferred or passed? 
			// Hook doesn't know AgentID.
			// We might need to map PID to AgentID via process tree.
			// For now, generate new IDs.
			AgentID:   0, // Placeholder, resolved by tree?
			ProcessID: api.ProcessID(muid.Make()),
			StartedAt: time.Now(),
			Running:   true,
		})
	}
}

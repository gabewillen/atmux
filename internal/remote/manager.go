// Package remote implements the main remote host manager and director functionality.
package remote

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RemoteManager coordinates all remote operations.
type RemoteManager struct {
	config     *RemoteConfig
	nats       *NATSManager
	control    *ControlOperations
	director   *DirectorOperations
	ptyStreamer *PTYStreamer
	
	// State
	role   string // "director" or "manager"
	hostID string
	
	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mutex  sync.RWMutex
}

// RemoteConfig holds configuration for remote operations.
type RemoteConfig struct {
	Role           string        // "director" or "manager"
	HostID         string        // Host identifier
	NATSURL        string        // NATS server URL
	CredsPath      string        // NATS credentials file
	SubjectPrefix  string        // NATS subject prefix
	KVBucket       string        // JetStream KV bucket
	RequestTimeout time.Duration // Request timeout
	BufferSize     int           // PTY buffer size
}

// NewRemoteManager creates a new remote manager.
func NewRemoteManager(config *RemoteConfig) (*RemoteManager, error) {
	if config.HostID == "" {
		config.HostID = GenerateHostID()
	}
	
	if config.Role != "director" && config.Role != "manager" {
		return nil, fmt.Errorf("role must be 'director' or 'manager'")
	}
	
	natsConfig := NATSConfig{
		URL:           config.NATSURL,
		CredsFile:     config.CredsPath,
		SubjectPrefix: config.SubjectPrefix,
		KVBucket:      config.KVBucket,
		Timeout:       config.RequestTimeout,
	}
	
	nm, err := NewNATSManager(config.HostID, config.Role, natsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create NATS manager: %w", err)
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	rm := &RemoteManager{
		config: config,
		nats:   nm,
		role:   config.Role,
		hostID: config.HostID,
		ctx:    ctx,
		cancel: cancel,
	}
	
	// Initialize role-specific components
	if config.Role == "manager" {
		rm.control = NewControlOperations(nm)
		rm.ptyStreamer = NewPTYStreamer(nm, config.BufferSize)
	} else {
		rm.director = NewDirectorOperations(nm)
		rm.ptyStreamer = NewPTYStreamer(nm, 0) // Director doesn't need buffering
	}
	
	return rm, nil
}

// Start starts the remote manager.
func (rm *RemoteManager) Start() error {
	// Connect to NATS
	if err := rm.nats.Connect(); err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}
	
	// Perform handshake
	if err := rm.nats.Handshake(); err != nil {
		rm.nats.Close()
		return fmt.Errorf("handshake failed: %w", err)
	}
	
	// Start role-specific operations
	if rm.role == "manager" {
		if err := rm.control.StartControlSubscriptions(); err != nil {
			rm.nats.Close()
			return fmt.Errorf("failed to start control subscriptions: %w", err)
		}
	}
	
	// Start heartbeat (both roles)
	rm.wg.Add(1)
	go rm.heartbeatLoop()
	
	return nil
}

// Stop stops the remote manager.
func (rm *RemoteManager) Stop() error {
	rm.cancel()
	rm.wg.Wait()
	
	if rm.control != nil {
		rm.control.Close()
	}
	
	if rm.ptyStreamer != nil {
		rm.ptyStreamer.Close()
	}
	
	if rm.nats != nil {
		rm.nats.Close()
	}
	
	return nil
}

// heartbeatLoop sends periodic heartbeats to maintain host presence.
func (rm *RemoteManager) heartbeatLoop() {
	defer rm.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.sendHeartbeat()
		}
	}
}

// sendHeartbeat sends a heartbeat to the KV store.
func (rm *RemoteManager) sendHeartbeat() {
	// TODO: Implement heartbeat to KV store
}

// GetRole returns the role of this manager.
func (rm *RemoteManager) GetRole() string {
	return rm.role
}

// GetHostID returns the host ID.
func (rm *RemoteManager) GetHostID() string {
	return rm.hostID
}

// IsReady returns true if the manager is ready for operations.
func (rm *RemoteManager) IsReady() bool {
	return rm.nats != nil && rm.nats.IsReady()
}

// Director operations (only available when role == "director")

// BootstrapRemoteHost bootstraps a remote host via SSH.
func (rm *RemoteManager) BootstrapRemoteHost(ctx context.Context, hostID string, config BootstrapConfig) error {
	if rm.role != "director" {
		return fmt.Errorf("bootstrap only available in director role")
	}
	
	return Bootstrap(ctx, hostID, config)
}

// SpawnRemoteAgent spawns an agent on a remote host.
func (rm *RemoteManager) SpawnRemoteAgent(ctx context.Context, hostID string, req SpawnPayload) (*SpawnPayload, error) {
	if rm.role != "director" {
		return nil, fmt.Errorf("spawn only available in director role")
	}
	
	if rm.director == nil {
		return nil, fmt.Errorf("director operations not initialized")
	}
	
	return rm.director.SpawnAgent(ctx, hostID, req)
}

// KillRemoteAgent kills an agent on a remote host.
func (rm *RemoteManager) KillRemoteAgent(ctx context.Context, hostID, agentID string) error {
	if rm.role != "director" {
		return fmt.Errorf("kill only available in director role")
	}
	
	if rm.director == nil {
		return fmt.Errorf("director operations not initialized")
	}
	
	return rm.director.KillAgent(ctx, hostID, agentID)
}

// SubscribeToRemotePTY subscribes to PTY output from a remote session.
func (rm *RemoteManager) SubscribeToRemotePTY(hostID, sessionID string, handler func(PTYData)) error {
	if rm.role != "director" {
		return fmt.Errorf("PTY subscription only available in director role")
	}
	
	if rm.ptyStreamer == nil {
		return fmt.Errorf("PTY streamer not initialized")
	}
	
	return rm.ptyStreamer.SubscribeToPTYOutput(hostID, sessionID, handler)
}

// SendRemotePTYInput sends input to a remote PTY session.
func (rm *RemoteManager) SendRemotePTYInput(hostID, sessionID string, data []byte) error {
	if rm.role != "director" {
		return fmt.Errorf("PTY input only available in director role")
	}
	
	if rm.ptyStreamer == nil {
		return fmt.Errorf("PTY streamer not initialized")
	}
	
	return rm.ptyStreamer.SendPTYInput(hostID, sessionID, data)
}

// Manager operations (only available when role == "manager")

// StartPTYSession starts PTY streaming for a local session.
func (rm *RemoteManager) StartPTYSession(sessionID string) error {
	if rm.role != "manager" {
		return fmt.Errorf("PTY sessions only available in manager role")
	}
	
	if rm.ptyStreamer == nil {
		return fmt.Errorf("PTY streamer not initialized")
	}
	
	return rm.ptyStreamer.StartPTYStreaming(sessionID)
}

// StopPTYSession stops PTY streaming for a local session.
func (rm *RemoteManager) StopPTYSession(sessionID string) {
	if rm.role != "manager" {
		return
	}
	
	if rm.ptyStreamer != nil {
		rm.ptyStreamer.StopPTYStreaming(sessionID)
	}
}

// PublishPTYOutput publishes PTY output to the director.
func (rm *RemoteManager) PublishPTYOutput(sessionID string, data []byte) error {
	if rm.role != "manager" {
		return fmt.Errorf("PTY output publishing only available in manager role")
	}
	
	if rm.ptyStreamer == nil {
		return fmt.Errorf("PTY streamer not initialized")
	}
	
	return rm.ptyStreamer.PublishPTYOutput(sessionID, data)
}
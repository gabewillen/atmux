// Package protocol implements remote communication protocol (transports events)
package protocol

import (
	"context"
	"errors"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/stateforward/amux/internal/config"
)

// ErrProtocol is returned when protocol operations fail
var ErrProtocol = errors.New("protocol operation failed")

// RemoteProtocol manages the complete remote protocol implementation
type RemoteProtocol struct {
	cfg             *config.Config
	sshBootstrap    *SSHBootstrap
	natsServer      *NATSServer
	kvStore         *KVStore
	natsAuth        *NATSAuth
	subjectBuilder  *SubjectBuilder
	controlOps      *ControlOperations
	handshakeHandler *HandshakeHandler
	reconnectionMgr *ReconnectionManager
}

// NewRemoteProtocol creates a new RemoteProtocol instance
func NewRemoteProtocol(cfg *config.Config) *RemoteProtocol {
	return &RemoteProtocol{
		cfg: cfg,
		sshBootstrap: NewSSHBootstrap(cfg),
	}
}

// Initialize initializes the remote protocol components
func (rp *RemoteProtocol) Initialize(ctx context.Context) error {
	// Initialize NATS server
	rp.natsServer = NewNATSServer(&rp.cfg.Remote.NATS)

	// Start NATS server in hub mode (director role)
	if err := rp.natsServer.StartHubServer(ctx); err != nil {
		return fmt.Errorf("failed to start NATS hub server: %w", err)
	}

	// Initialize KV store
	nc := rp.natsServer.GetClient()
	var err error
	rp.kvStore, err = NewKVStore(nc, rp.cfg.Remote.NATS.KVBucket)
	if err != nil {
		return fmt.Errorf("failed to initialize KV store: %w", err)
	}

	// Initialize other components
	rp.natsAuth = NewNATSAuth(nc)
	rp.subjectBuilder = NewSubjectBuilder(rp.cfg.Remote.NATS.SubjectPrefix)
	rp.controlOps = NewControlOperations(nc, rp.subjectBuilder, rp.cfg.Remote.RequestTimeout)
	rp.handshakeHandler = NewHandshakeHandler(nc, rp.subjectBuilder)
	rp.reconnectionMgr = NewReconnectionManager(nc, rp.subjectBuilder, rp.controlOps)

	return nil
}

// InitializeLeafNode initializes the protocol for a leaf node (manager role)
func (rp *RemoteProtocol) InitializeLeafNode(ctx context.Context, hubURL, credsPath string) error {
	// Start NATS server in leaf mode (manager role)
	rp.natsServer = NewNATSServer(&rp.cfg.Remote.NATS)
	if err := rp.natsServer.StartLeafServer(ctx, hubURL, credsPath); err != nil {
		return fmt.Errorf("failed to start NATS leaf server: %w", err)
	}

	// Initialize other components
	nc := rp.natsServer.GetClient()
	var err error
	rp.kvStore, err = NewKVStore(nc, rp.cfg.Remote.NATS.KVBucket)
	if err != nil {
		return fmt.Errorf("failed to initialize KV store: %w", err)
	}

	// Initialize other components
	rp.natsAuth = NewNATSAuth(nc)
	rp.subjectBuilder = NewSubjectBuilder(rp.cfg.Remote.NATS.SubjectPrefix)
	rp.controlOps = NewControlOperations(nc, rp.subjectBuilder, rp.cfg.Remote.RequestTimeout)
	rp.handshakeHandler = NewHandshakeHandler(nc, rp.subjectBuilder)
	rp.reconnectionMgr = NewReconnectionManager(nc, rp.subjectBuilder, rp.controlOps)

	return nil
}

// GetNATSConnection returns the NATS connection
func (rp *RemoteProtocol) GetNATSConnection() *nats.Conn {
	if rp.natsServer != nil {
		return rp.natsServer.GetClient()
	}
	return nil
}

// BootstrapRemoteHost performs the complete SSH bootstrap for a remote host
func (rp *RemoteProtocol) BootstrapRemoteHost(ctx context.Context, location Location) error {
	return rp.sshBootstrap.Bootstrap(ctx, location)
}

// PerformHandshake performs the handshake with a remote host
func (rp *RemoteProtocol) PerformHandshake(hostID, peerID string) error {
	return rp.handshakeHandler.PerformHandshake(hostID, peerID)
}

// Spawn starts a new agent session on a remote host
func (rp *RemoteProtocol) Spawn(ctx context.Context, hostID string, payload SpawnPayload) (*SpawnResponsePayload, error) {
	return rp.controlOps.Spawn(ctx, hostID, payload)
}

// Kill terminates an agent session on a remote host
func (rp *RemoteProtocol) Kill(ctx context.Context, hostID string, payload KillPayload) (*KillResponsePayload, error) {
	return rp.controlOps.Kill(ctx, hostID, payload)
}

// Replay requests replay of PTY output for a session
func (rp *RemoteProtocol) Replay(ctx context.Context, hostID string, payload ReplayPayload) (*ReplayResponsePayload, error) {
	return rp.controlOps.Replay(ctx, hostID, payload)
}

// AddSessionToReconnectionManager adds a session to the reconnection manager
func (rp *RemoteProtocol) AddSessionToReconnectionManager(sessionID string) {
	if rp.reconnectionMgr != nil {
		rp.reconnectionMgr.AddSession(sessionID, rp.cfg.Remote.BufferSize)
	}
}

// Stop shuts down the remote protocol
func (rp *RemoteProtocol) Stop() {
	if rp.natsServer != nil {
		rp.natsServer.Stop()
	}
}
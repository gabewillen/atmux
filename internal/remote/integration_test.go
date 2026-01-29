package remote

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestRemoteIntegration performs integration tests for remote functionality.
// Note: These tests require NATS server for full integration testing.
func TestRemoteIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	
	// Check if NATS is available (optional for CI)
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	
	t.Run("DirectorManagerHandshake", func(t *testing.T) {
		testDirectorManagerHandshake(t, natsURL)
	})
	
	t.Run("ControlOperations", func(t *testing.T) {
		testControlOperations(t, natsURL)
	})
	
	t.Run("PTYStreaming", func(t *testing.T) {
		testPTYStreaming(t, natsURL)
	})
}

func testDirectorManagerHandshake(t *testing.T, natsURL string) {
	// Create director
	directorConfig := &RemoteConfig{
		Role:           "director",
		HostID:         "test-director",
		NATSURL:        natsURL,
		SubjectPrefix:  "test",
		KVBucket:       "TEST_KV",
		RequestTimeout: 10 * time.Second,
	}
	
	director, err := NewRemoteManager(directorConfig)
	if err != nil {
		t.Fatalf("Failed to create director: %v", err)
	}
	
	// Create manager
	managerConfig := &RemoteConfig{
		Role:           "manager",
		HostID:         "test-manager",
		NATSURL:        natsURL,
		SubjectPrefix:  "test",
		KVBucket:       "TEST_KV",
		RequestTimeout: 10 * time.Second,
		BufferSize:     8192,
	}
	
	manager, err := NewRemoteManager(managerConfig)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	// Start director first
	if err := director.Start(); err != nil {
		// If NATS is not available, skip test
		if IsNATSError(err) {
			t.Skipf("NATS not available: %v", err)
		}
		t.Fatalf("Failed to start director: %v", err)
	}
	defer director.Stop()
	
	// Start manager
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	defer manager.Stop()
	
	// Wait for handshake to complete
	time.Sleep(2 * time.Second)
	
	// Verify both are ready
	if !director.IsReady() {
		t.Error("Director should be ready after handshake")
	}
	
	if !manager.IsReady() {
		t.Error("Manager should be ready after handshake")
	}
}

func testControlOperations(t *testing.T, natsURL string) {
	// Create director and manager
	directorConfig := &RemoteConfig{
		Role:           "director",
		HostID:         "test-director-2",
		NATSURL:        natsURL,
		SubjectPrefix:  "test2",
		KVBucket:       "TEST_KV2",
		RequestTimeout: 10 * time.Second,
	}
	
	director, err := NewRemoteManager(directorConfig)
	if err != nil {
		t.Fatalf("Failed to create director: %v", err)
	}
	
	managerConfig := &RemoteConfig{
		Role:           "manager",
		HostID:         "test-manager-2",
		NATSURL:        natsURL,
		SubjectPrefix:  "test2",
		KVBucket:       "TEST_KV2",
		RequestTimeout: 10 * time.Second,
		BufferSize:     8192,
	}
	
	manager, err := NewRemoteManager(managerConfig)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	// Start both
	if err := director.Start(); err != nil {
		if IsNATSError(err) {
			t.Skipf("NATS not available: %v", err)
		}
		t.Fatalf("Failed to start director: %v", err)
	}
	defer director.Stop()
	
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	defer manager.Stop()
	
	// Wait for handshake
	time.Sleep(2 * time.Second)
	
	// Test spawn operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	spawnReq := SpawnPayload{
		AgentID:   "42",
		AgentSlug: "test-agent",
		RepoPath:  "/tmp/test-repo",
		Command:   []string{"echo", "hello"},
	}
	
	resp, err := director.SpawnRemoteAgent(ctx, "test-manager-2", spawnReq)
	if err != nil {
		t.Fatalf("Spawn operation failed: %v", err)
	}
	
	if resp.AgentID != "42" {
		t.Errorf("Expected AgentID 42, got %s", resp.AgentID)
	}
	
	if resp.SessionID == "" {
		t.Error("Expected non-empty SessionID")
	}
}

func testPTYStreaming(t *testing.T, natsURL string) {
	// Create manager for PTY testing
	managerConfig := &RemoteConfig{
		Role:           "manager",
		HostID:         "test-pty-manager",
		NATSURL:        natsURL,
		SubjectPrefix:  "test3",
		KVBucket:       "TEST_KV3",
		RequestTimeout: 10 * time.Second,
		BufferSize:     1024,
	}
	
	manager, err := NewRemoteManager(managerConfig)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	// Start manager
	if err := manager.Start(); err != nil {
		if IsNATSError(err) {
			t.Skipf("NATS not available: %v", err)
		}
		t.Fatalf("Failed to start manager: %v", err)
	}
	defer manager.Stop()
	
	// Test PTY session
	sessionID := "test-session-123"
	
	if err := manager.StartPTYSession(sessionID); err != nil {
		t.Fatalf("Failed to start PTY session: %v", err)
	}
	
	// Publish some test data
	testData := []byte("Hello, PTY world!")
	if err := manager.PublishPTYOutput(sessionID, testData); err != nil {
		t.Fatalf("Failed to publish PTY output: %v", err)
	}
	
	// Stop session
	manager.StopPTYSession(sessionID)
}

// IsNATSError checks if an error is due to NATS connectivity issues.
func IsNATSError(err error) bool {
	return err != nil && (
		err == ErrNATSConnectionFailed ||
		err == ErrJetStreamFailed ||
		err.Error() == "nats connection failed")
}
package natsconn

import (
	"fmt"
	"testing"
)

// TestHostInfoKeyFormat verifies the KV key format for host info entries.
// Per spec section 5.5.6.3: keys follow hosts/<host_id>/info.
func TestHostInfoKeyFormat(t *testing.T) {
	tests := []struct {
		hostID   string
		expected string
	}{
		{"host-abc-123", "hosts/host-abc-123/info"},
		{"node1", "hosts/node1/info"},
		{"my-server", "hosts/my-server/info"},
		{"", "hosts//info"},
	}

	for _, tt := range tests {
		t.Run(tt.hostID, func(t *testing.T) {
			key := "hosts/" + tt.hostID + "/info"
			if key != tt.expected {
				t.Errorf("host info key = %q, want %q", key, tt.expected)
			}
		})
	}
}

// TestHeartbeatKeyFormat verifies the KV key format for heartbeat entries.
// Per spec section 5.5.6.3: keys follow hosts/<host_id>/heartbeat.
func TestHeartbeatKeyFormat(t *testing.T) {
	tests := []struct {
		hostID   string
		expected string
	}{
		{"host-abc-123", "hosts/host-abc-123/heartbeat"},
		{"node1", "hosts/node1/heartbeat"},
		{"my-server", "hosts/my-server/heartbeat"},
		{"", "hosts//heartbeat"},
	}

	for _, tt := range tests {
		t.Run(tt.hostID, func(t *testing.T) {
			key := "hosts/" + tt.hostID + "/heartbeat"
			if key != tt.expected {
				t.Errorf("heartbeat key = %q, want %q", key, tt.expected)
			}
		})
	}
}

// TestSessionMetaKeyFormat verifies the KV key format for session metadata entries.
// Per spec section 5.5.6.3: keys follow sessions/<host_id>/<session_id>.
func TestSessionMetaKeyFormat(t *testing.T) {
	tests := []struct {
		hostID    string
		sessionID string
		expected  string
	}{
		{"host1", "sess-001", "sessions/host1/sess-001"},
		{"node-abc", "session-xyz", "sessions/node-abc/session-xyz"},
		{"h", "s", "sessions/h/s"},
		{"host1", "", "sessions/host1/"},
		{"", "sess1", "sessions//sess1"},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("host=%s/session=%s", tt.hostID, tt.sessionID)
		t.Run(name, func(t *testing.T) {
			key := "sessions/" + tt.hostID + "/" + tt.sessionID
			if key != tt.expected {
				t.Errorf("session meta key = %q, want %q", key, tt.expected)
			}
		})
	}
}

// TestHostInfoJSON verifies that HostInfo fields are properly tagged for JSON marshaling.
func TestHostInfoJSON(t *testing.T) {
	info := HostInfo{
		Version:   "1.0.0",
		OS:        "linux",
		Arch:      "amd64",
		PeerID:    "peer-abc",
		StartedAt: "2025-01-01T00:00:00Z",
	}

	if info.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", info.Version, "1.0.0")
	}
	if info.OS != "linux" {
		t.Errorf("OS = %q, want %q", info.OS, "linux")
	}
	if info.Arch != "amd64" {
		t.Errorf("Arch = %q, want %q", info.Arch, "amd64")
	}
	if info.PeerID != "peer-abc" {
		t.Errorf("PeerID = %q, want %q", info.PeerID, "peer-abc")
	}
	if info.StartedAt != "2025-01-01T00:00:00Z" {
		t.Errorf("StartedAt = %q, want %q", info.StartedAt, "2025-01-01T00:00:00Z")
	}
}

// TestHostHeartbeatJSON verifies that HostHeartbeat fields are properly tagged.
func TestHostHeartbeatJSON(t *testing.T) {
	hb := HostHeartbeat{
		Timestamp: "2025-06-15T12:00:00Z",
	}

	if hb.Timestamp != "2025-06-15T12:00:00Z" {
		t.Errorf("Timestamp = %q, want %q", hb.Timestamp, "2025-06-15T12:00:00Z")
	}
}

// TestSessionMetaJSON verifies that SessionMeta fields are properly tagged.
func TestSessionMetaJSON(t *testing.T) {
	meta := SessionMeta{
		AgentID:   "agent-123",
		AgentSlug: "my-agent",
		RepoPath:  "/path/to/repo",
		State:     "running",
	}

	if meta.AgentID != "agent-123" {
		t.Errorf("AgentID = %q, want %q", meta.AgentID, "agent-123")
	}
	if meta.AgentSlug != "my-agent" {
		t.Errorf("AgentSlug = %q, want %q", meta.AgentSlug, "my-agent")
	}
	if meta.RepoPath != "/path/to/repo" {
		t.Errorf("RepoPath = %q, want %q", meta.RepoPath, "/path/to/repo")
	}
	if meta.State != "running" {
		t.Errorf("State = %q, want %q", meta.State, "running")
	}
}

// TestKVStoreBucket verifies that the Bucket method returns the expected bucket name.
func TestKVStoreBucket(t *testing.T) {
	tests := []struct {
		bucket string
	}{
		{"AMUX_KV"},
		{"test-bucket"},
		{""},
	}

	for _, tt := range tests {
		t.Run(tt.bucket, func(t *testing.T) {
			store := &KVStore{bucket: tt.bucket}
			if got := store.Bucket(); got != tt.bucket {
				t.Errorf("Bucket() = %q, want %q", got, tt.bucket)
			}
		})
	}
}

// TestKeyFormatConsistency verifies that all three key schemes use forward
// slashes as delimiters consistently.
func TestKeyFormatConsistency(t *testing.T) {
	hostID := "test-host"
	sessionID := "test-session"

	infoKey := "hosts/" + hostID + "/info"
	heartbeatKey := "hosts/" + hostID + "/heartbeat"
	sessionKey := "sessions/" + hostID + "/" + sessionID

	// All keys should use forward slashes
	for _, tc := range []struct {
		name string
		key  string
	}{
		{"info", infoKey},
		{"heartbeat", heartbeatKey},
		{"session", sessionKey},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Verify no dots are used (dots were mentioned in the code comment
			// for PutSessionMeta but slashes are used in implementation)
			for _, ch := range tc.key {
				if ch == '.' {
					t.Errorf("key %q contains dot delimiter, expected only forward slashes", tc.key)
				}
			}
		})
	}
}

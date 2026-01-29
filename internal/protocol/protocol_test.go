package protocol

import (
	"encoding/json"
	"testing"
)

func TestControlMessageRoundTrip(t *testing.T) {
	payload := &SpawnRequest{
		AgentID:   "42",
		AgentSlug: "backend-dev",
		RepoPath:  "~/projects/my-repo",
		Command:   []string{"claude-code"},
		Env:       map[string]string{"TERM": "xterm-256color"},
	}

	msg, err := NewControlMessage(TypeSpawn, payload)
	if err != nil {
		t.Fatalf("NewControlMessage: %v", err)
	}

	if msg.Type != TypeSpawn {
		t.Fatalf("Type = %q, want %q", msg.Type, TypeSpawn)
	}

	// Marshal to JSON and back
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded ControlMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Type != TypeSpawn {
		t.Fatalf("decoded Type = %q, want %q", decoded.Type, TypeSpawn)
	}

	var decodedPayload SpawnRequest
	if err := decoded.DecodePayload(&decodedPayload); err != nil {
		t.Fatalf("DecodePayload: %v", err)
	}

	if decodedPayload.AgentID != "42" {
		t.Fatalf("AgentID = %q, want %q", decodedPayload.AgentID, "42")
	}
	if decodedPayload.AgentSlug != "backend-dev" {
		t.Fatalf("AgentSlug = %q, want %q", decodedPayload.AgentSlug, "backend-dev")
	}
	if decodedPayload.RepoPath != "~/projects/my-repo" {
		t.Fatalf("RepoPath = %q, want %q", decodedPayload.RepoPath, "~/projects/my-repo")
	}
	if len(decodedPayload.Command) != 1 || decodedPayload.Command[0] != "claude-code" {
		t.Fatalf("Command = %v, want [claude-code]", decodedPayload.Command)
	}
}

func TestHandshakePayloadJSON(t *testing.T) {
	// Test the exact JSON format from the spec
	expected := `{"type":"handshake","payload":{"protocol":1,"peer_id":"5678","role":"daemon","host_id":"devbox"}}`

	payload := &HandshakePayload{
		Protocol: 1,
		PeerID:   "5678",
		Role:     "daemon",
		HostID:   "devbox",
	}

	msg, err := NewControlMessage(TypeHandshake, payload)
	if err != nil {
		t.Fatalf("NewControlMessage: %v", err)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if string(data) != expected {
		t.Fatalf("JSON = %s\nwant = %s", data, expected)
	}
}

func TestErrorPayloadJSON(t *testing.T) {
	expected := `{"type":"error","payload":{"request_type":"spawn","code":"invalid_repo","message":"repo_path is not a git repository"}}`

	msg, err := NewErrorMessage("spawn", "invalid_repo", "repo_path is not a git repository")
	if err != nil {
		t.Fatalf("NewErrorMessage: %v", err)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if string(data) != expected {
		t.Fatalf("JSON = %s\nwant = %s", data, expected)
	}
}

func TestSpawnResponseJSON(t *testing.T) {
	payload := &SpawnResponse{
		AgentID:   "42",
		SessionID: "9001",
	}

	msg, err := NewControlMessage(TypeSpawn, payload)
	if err != nil {
		t.Fatalf("NewControlMessage: %v", err)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// Decode and verify
	var decoded ControlMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	var resp SpawnResponse
	if err := decoded.DecodePayload(&resp); err != nil {
		t.Fatalf("DecodePayload: %v", err)
	}

	if resp.AgentID != "42" || resp.SessionID != "9001" {
		t.Fatalf("SpawnResponse = %+v, want AgentID=42 SessionID=9001", resp)
	}
}

func TestKillPayloadJSON(t *testing.T) {
	req := &KillRequest{SessionID: "9001"}
	msg, err := NewControlMessage(TypeKill, req)
	if err != nil {
		t.Fatalf("NewControlMessage: %v", err)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded ControlMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Type != TypeKill {
		t.Fatalf("Type = %q, want %q", decoded.Type, TypeKill)
	}

	resp := &KillResponse{SessionID: "9001", Killed: true}
	respMsg, _ := NewControlMessage(TypeKill, resp)
	respData, _ := json.Marshal(respMsg)

	var decodedResp KillResponse
	var respCtl ControlMessage
	_ = json.Unmarshal(respData, &respCtl)
	_ = respCtl.DecodePayload(&decodedResp)

	if !decodedResp.Killed || decodedResp.SessionID != "9001" {
		t.Fatalf("KillResponse = %+v, want Killed=true SessionID=9001", decodedResp)
	}
}

func TestReplayPayloadJSON(t *testing.T) {
	req := &ReplayRequest{SessionID: "9001"}
	msg, _ := NewControlMessage(TypeReplay, req)
	data, _ := json.Marshal(msg)

	var decoded ControlMessage
	_ = json.Unmarshal(data, &decoded)

	if decoded.Type != TypeReplay {
		t.Fatalf("Type = %q, want %q", decoded.Type, TypeReplay)
	}

	resp := &ReplayResponse{SessionID: "9001", Accepted: true}
	respMsg, _ := NewControlMessage(TypeReplay, resp)
	respData, _ := json.Marshal(respMsg)

	var decodedResp ReplayResponse
	var respCtl ControlMessage
	_ = json.Unmarshal(respData, &respCtl)
	_ = respCtl.DecodePayload(&decodedResp)

	if !decodedResp.Accepted || decodedResp.SessionID != "9001" {
		t.Fatalf("ReplayResponse = %+v, want Accepted=true SessionID=9001", decodedResp)
	}
}

func TestPingPongJSON(t *testing.T) {
	ping := &PingPayload{TSUnixMs: 1700000000000}
	msg, _ := NewControlMessage(TypePing, ping)
	data, _ := json.Marshal(msg)

	expected := `{"type":"ping","payload":{"ts_unix_ms":1700000000000}}`
	if string(data) != expected {
		t.Fatalf("Ping JSON = %s\nwant = %s", data, expected)
	}

	pong := &PongPayload{TSUnixMs: 1700000000000}
	pongMsg, _ := NewControlMessage(TypePong, pong)
	pongData, _ := json.Marshal(pongMsg)

	expectedPong := `{"type":"pong","payload":{"ts_unix_ms":1700000000000}}`
	if string(pongData) != expectedPong {
		t.Fatalf("Pong JSON = %s\nwant = %s", pongData, expectedPong)
	}
}

// --- Subject tests ---

func TestHandshakeSubject(t *testing.T) {
	got := HandshakeSubject("amux", "devbox")
	want := "amux.handshake.devbox"
	if got != want {
		t.Fatalf("HandshakeSubject = %q, want %q", got, want)
	}
}

func TestControlSubject(t *testing.T) {
	got := ControlSubject("amux", "devbox")
	want := "amux.ctl.devbox"
	if got != want {
		t.Fatalf("ControlSubject = %q, want %q", got, want)
	}
}

func TestEventsSubject(t *testing.T) {
	got := EventsSubject("amux", "devbox")
	want := "amux.events.devbox"
	if got != want {
		t.Fatalf("EventsSubject = %q, want %q", got, want)
	}
}

func TestPTYOutputSubject(t *testing.T) {
	got := PTYOutputSubject("amux", "devbox", "9001")
	want := "amux.pty.devbox.9001.out"
	if got != want {
		t.Fatalf("PTYOutputSubject = %q, want %q", got, want)
	}
}

func TestPTYInputSubject(t *testing.T) {
	got := PTYInputSubject("amux", "devbox", "9001")
	want := "amux.pty.devbox.9001.in"
	if got != want {
		t.Fatalf("PTYInputSubject = %q, want %q", got, want)
	}
}

func TestPTYInputWildcard(t *testing.T) {
	got := PTYInputWildcard("amux", "devbox")
	want := "amux.pty.devbox.*.in"
	if got != want {
		t.Fatalf("PTYInputWildcard = %q, want %q", got, want)
	}
}

func TestDirectorChannelSubject(t *testing.T) {
	got := DirectorChannelSubject("amux")
	want := "amux.comm.director"
	if got != want {
		t.Fatalf("DirectorChannelSubject = %q, want %q", got, want)
	}
}

func TestManagerChannelSubject(t *testing.T) {
	got := ManagerChannelSubject("amux", "devbox")
	want := "amux.comm.manager.devbox"
	if got != want {
		t.Fatalf("ManagerChannelSubject = %q, want %q", got, want)
	}
}

func TestAgentChannelSubject(t *testing.T) {
	got := AgentChannelSubject("amux", "devbox", "42")
	want := "amux.comm.agent.devbox.42"
	if got != want {
		t.Fatalf("AgentChannelSubject = %q, want %q", got, want)
	}
}

func TestBroadcastChannelSubject(t *testing.T) {
	got := BroadcastChannelSubject("amux")
	want := "amux.comm.broadcast"
	if got != want {
		t.Fatalf("BroadcastChannelSubject = %q, want %q", got, want)
	}
}

func TestCustomPrefix(t *testing.T) {
	prefix := "myorg.amux"
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"handshake", HandshakeSubject(prefix, "h1"), "myorg.amux.handshake.h1"},
		{"control", ControlSubject(prefix, "h1"), "myorg.amux.ctl.h1"},
		{"events", EventsSubject(prefix, "h1"), "myorg.amux.events.h1"},
		{"pty_out", PTYOutputSubject(prefix, "h1", "s1"), "myorg.amux.pty.h1.s1.out"},
		{"pty_in", PTYInputSubject(prefix, "h1", "s1"), "myorg.amux.pty.h1.s1.in"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

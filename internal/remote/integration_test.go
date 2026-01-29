package remote

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/remote/natsconn"
)

// startTestNATS starts an in-memory NATS server for testing.
func startTestNATS(t *testing.T) (*natsserver.Server, string) {
	t.Helper()
	opts := &natsserver.Options{
		Host:           "127.0.0.1",
		Port:           -1, // Random port
		NoLog:          true,
		NoSigs:         true,
		MaxControlLine: 4096,
		JetStream:      true,
		StoreDir:       t.TempDir(),
	}
	s, err := natsserver.NewServer(opts)
	if err != nil {
		t.Fatalf("start test nats: %v", err)
	}
	s.Start()
	if !s.ReadyForConnections(5 * time.Second) {
		t.Fatal("nats server not ready")
	}
	return s, s.ClientURL()
}

func testConfig(natsURL string) *config.Config {
	cfg := config.DefaultConfig()
	cfg.Remote.NATS.URL = natsURL
	cfg.Remote.NATS.SubjectPrefix = "amux"
	cfg.Remote.NATS.KVBucket = "AMUX_KV_TEST"
	cfg.Remote.RequestTimeout = config.Duration{Duration: 5 * time.Second}
	cfg.Remote.BufferSize = config.ByteSize{Bytes: 10 * 1024 * 1024}
	return cfg
}

func connectTest(t *testing.T, url, name string) *natsconn.Conn {
	t.Helper()
	conn, err := natsconn.Connect(context.Background(), &natsconn.Options{
		URL:  url,
		Name: name,
	})
	if err != nil {
		t.Fatalf("connect %s: %v", name, err)
	}
	return conn
}

func TestHandshakeExchange(t *testing.T) {
	s, url := startTestNATS(t)
	defer s.Shutdown()

	prefix := "amux"
	cfg := testConfig(url)

	// Connect two clients: director and daemon
	dirConn := connectTest(t, url, "director")
	defer dirConn.Close()

	daemonConn := connectTest(t, url, "daemon")
	defer daemonConn.Close()

	// Director subscribes to handshake requests
	handshakeReplies := make(chan *nats.Msg, 1)
	_, err := dirConn.NC().Subscribe(prefix+".handshake.*", func(msg *nats.Msg) {
		// Process handshake
		var ctlMsg protocol.ControlMessage
		if err := json.Unmarshal(msg.Data, &ctlMsg); err != nil {
			t.Errorf("unmarshal handshake: %v", err)
			return
		}

		var payload protocol.HandshakePayload
		if err := ctlMsg.DecodePayload(&payload); err != nil {
			t.Errorf("decode handshake payload: %v", err)
			return
		}

		// Send response
		resp, _ := protocol.NewControlMessage(protocol.TypeHandshake, &protocol.HandshakePayload{
			Protocol: protocol.ProtocolVersion,
			PeerID:   "1234",
			Role:     "director",
			HostID:   "director-host",
		})
		data, _ := json.Marshal(resp)
		_ = msg.Respond(data)
		handshakeReplies <- msg
	})
	if err != nil {
		t.Fatalf("subscribe handshake: %v", err)
	}
	_ = dirConn.Flush()

	// Daemon sends handshake request
	hsPayload := &protocol.HandshakePayload{
		Protocol: protocol.ProtocolVersion,
		PeerID:   "5678",
		Role:     "daemon",
		HostID:   "devbox",
	}
	hsMsg, _ := protocol.NewControlMessage(protocol.TypeHandshake, hsPayload)
	hsData, _ := json.Marshal(hsMsg)

	reply, err := daemonConn.Request(
		protocol.HandshakeSubject(prefix, "devbox"),
		hsData,
		cfg.Remote.RequestTimeout.Duration,
	)
	if err != nil {
		t.Fatalf("handshake request: %v", err)
	}

	// Verify response
	var respMsg protocol.ControlMessage
	if err := json.Unmarshal(reply.Data, &respMsg); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if respMsg.Type != protocol.TypeHandshake {
		t.Fatalf("response type = %q, want %q", respMsg.Type, protocol.TypeHandshake)
	}

	var respPayload protocol.HandshakePayload
	if err := respMsg.DecodePayload(&respPayload); err != nil {
		t.Fatalf("decode response payload: %v", err)
	}

	if respPayload.Role != "director" {
		t.Fatalf("response role = %q, want %q", respPayload.Role, "director")
	}

	// Wait for the director to have processed the handshake
	select {
	case <-handshakeReplies:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for handshake")
	}
}

func TestSpawnReplyExchange(t *testing.T) {
	s, url := startTestNATS(t)
	defer s.Shutdown()

	prefix := "amux"
	cfg := testConfig(url)

	dirConn := connectTest(t, url, "director")
	defer dirConn.Close()

	daemonConn := connectTest(t, url, "daemon")
	defer daemonConn.Close()

	// Daemon subscribes to control requests and handles spawn
	_, err := daemonConn.NC().Subscribe(
		protocol.ControlSubject(prefix, "devbox"),
		func(msg *nats.Msg) {
			var ctlMsg protocol.ControlMessage
			if err := json.Unmarshal(msg.Data, &ctlMsg); err != nil {
				return
			}

			if ctlMsg.Type == protocol.TypeSpawn {
				var req protocol.SpawnRequest
				_ = ctlMsg.DecodePayload(&req)

				resp, _ := protocol.NewControlMessage(protocol.TypeSpawn, &protocol.SpawnResponse{
					AgentID:   req.AgentID,
					SessionID: "9001",
				})
				data, _ := json.Marshal(resp)
				_ = msg.Respond(data)
			}
		},
	)
	if err != nil {
		t.Fatalf("subscribe control: %v", err)
	}
	_ = daemonConn.Flush()

	// Director sends spawn request
	spawnReq := &protocol.SpawnRequest{
		AgentID:   "42",
		AgentSlug: "backend-dev",
		RepoPath:  "~/projects/my-repo",
		Command:   []string{"claude-code"},
	}
	spawnMsg, _ := protocol.NewControlMessage(protocol.TypeSpawn, spawnReq)
	spawnData, _ := json.Marshal(spawnMsg)

	reply, err := dirConn.Request(
		protocol.ControlSubject(prefix, "devbox"),
		spawnData,
		cfg.Remote.RequestTimeout.Duration,
	)
	if err != nil {
		t.Fatalf("spawn request: %v", err)
	}

	var respMsg protocol.ControlMessage
	if err := json.Unmarshal(reply.Data, &respMsg); err != nil {
		t.Fatalf("unmarshal spawn response: %v", err)
	}

	if respMsg.Type != protocol.TypeSpawn {
		t.Fatalf("response type = %q, want %q", respMsg.Type, protocol.TypeSpawn)
	}

	var resp protocol.SpawnResponse
	if err := respMsg.DecodePayload(&resp); err != nil {
		t.Fatalf("decode spawn response: %v", err)
	}

	if resp.AgentID != "42" {
		t.Fatalf("AgentID = %q, want %q", resp.AgentID, "42")
	}
	if resp.SessionID != "9001" {
		t.Fatalf("SessionID = %q, want %q", resp.SessionID, "9001")
	}
}

func TestNotReadyBeforeHandshake(t *testing.T) {
	s, url := startTestNATS(t)
	defer s.Shutdown()

	prefix := "amux"
	cfg := testConfig(url)

	dirConn := connectTest(t, url, "director")
	defer dirConn.Close()

	daemonConn := connectTest(t, url, "daemon")
	defer daemonConn.Close()

	// Daemon subscribes but rejects pre-handshake requests with not_ready
	_, err := daemonConn.NC().Subscribe(
		protocol.ControlSubject(prefix, "devbox"),
		func(msg *nats.Msg) {
			var ctlMsg protocol.ControlMessage
			if err := json.Unmarshal(msg.Data, &ctlMsg); err != nil {
				return
			}

			// Reject with not_ready (simulating pre-handshake state)
			errMsg, _ := protocol.NewErrorMessage(ctlMsg.Type, protocol.CodeNotReady,
				"handshake not yet complete")
			data, _ := json.Marshal(errMsg)
			_ = msg.Respond(data)
		},
	)
	if err != nil {
		t.Fatalf("subscribe control: %v", err)
	}
	_ = daemonConn.Flush()

	// Director sends spawn before handshake
	spawnReq := &protocol.SpawnRequest{AgentID: "42", AgentSlug: "test"}
	spawnMsg, _ := protocol.NewControlMessage(protocol.TypeSpawn, spawnReq)
	spawnData, _ := json.Marshal(spawnMsg)

	reply, err := dirConn.Request(
		protocol.ControlSubject(prefix, "devbox"),
		spawnData,
		cfg.Remote.RequestTimeout.Duration,
	)
	if err != nil {
		t.Fatalf("spawn request: %v", err)
	}

	var respMsg protocol.ControlMessage
	if err := json.Unmarshal(reply.Data, &respMsg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if respMsg.Type != protocol.TypeError {
		t.Fatalf("response type = %q, want %q", respMsg.Type, protocol.TypeError)
	}

	var errPayload protocol.ErrorPayload
	if err := respMsg.DecodePayload(&errPayload); err != nil {
		t.Fatalf("decode error payload: %v", err)
	}

	if errPayload.Code != protocol.CodeNotReady {
		t.Fatalf("error code = %q, want %q", errPayload.Code, protocol.CodeNotReady)
	}
	if errPayload.RequestType != protocol.TypeSpawn {
		t.Fatalf("request_type = %q, want %q", errPayload.RequestType, protocol.TypeSpawn)
	}
}

func TestPTYIOTransport(t *testing.T) {
	s, url := startTestNATS(t)
	defer s.Shutdown()

	prefix := "amux"

	dirConn := connectTest(t, url, "director")
	defer dirConn.Close()

	daemonConn := connectTest(t, url, "daemon")
	defer daemonConn.Close()

	// Director subscribes to PTY output
	received := make(chan []byte, 10)
	_, err := dirConn.NC().Subscribe(
		protocol.PTYOutputSubject(prefix, "devbox", "9001"),
		func(msg *nats.Msg) {
			received <- msg.Data
		},
	)
	if err != nil {
		t.Fatalf("subscribe pty output: %v", err)
	}

	// Daemon subscribes to PTY input
	inputReceived := make(chan []byte, 10)
	_, err = daemonConn.NC().Subscribe(
		protocol.PTYInputSubject(prefix, "devbox", "9001"),
		func(msg *nats.Msg) {
			inputReceived <- msg.Data
		},
	)
	if err != nil {
		t.Fatalf("subscribe pty input: %v", err)
	}

	// Flush subscriptions to ensure they propagate to the server
	_ = dirConn.Flush()
	_ = daemonConn.Flush()

	// Daemon publishes PTY output
	_ = daemonConn.Publish(
		protocol.PTYOutputSubject(prefix, "devbox", "9001"),
		[]byte("$ cargo test\n"),
	)
	_ = daemonConn.Flush()

	// Director should receive output
	select {
	case data := <-received:
		if string(data) != "$ cargo test\n" {
			t.Fatalf("received = %q, want %q", data, "$ cargo test\n")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for pty output")
	}

	// Director publishes PTY input
	_ = dirConn.Publish(
		protocol.PTYInputSubject(prefix, "devbox", "9001"),
		[]byte("ls -la\n"),
	)
	_ = dirConn.Flush()

	// Daemon should receive input
	select {
	case data := <-inputReceived:
		if string(data) != "ls -la\n" {
			t.Fatalf("received input = %q, want %q", data, "ls -la\n")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for pty input")
	}
}

func TestHostEventTransport(t *testing.T) {
	s, url := startTestNATS(t)
	defer s.Shutdown()

	prefix := "amux"

	dirConn := connectTest(t, url, "director")
	defer dirConn.Close()

	daemonConn := connectTest(t, url, "daemon")
	defer daemonConn.Close()

	// Director subscribes to events from all hosts
	received := make(chan *protocol.EventMessage, 10)
	_, err := dirConn.NC().Subscribe(prefix+".events.*", func(msg *nats.Msg) {
		var evtMsg protocol.EventMessage
		if err := json.Unmarshal(msg.Data, &evtMsg); err != nil {
			return
		}
		received <- &evtMsg
	})
	if err != nil {
		t.Fatalf("subscribe events: %v", err)
	}
	_ = dirConn.Flush()

	// Daemon publishes a connection.established event
	evt, _ := protocol.NewBroadcastEvent("connection.established", &protocol.ConnectionEstablishedEvent{
		PeerID:    "5678",
		Timestamp: "2026-01-18T10:30:00Z",
	})
	evtData, _ := json.Marshal(evt)
	_ = daemonConn.Publish(protocol.EventsSubject(prefix, "devbox"), evtData)
	_ = daemonConn.Flush()

	select {
	case evtMsg := <-received:
		if evtMsg.Event.Name != "connection.established" {
			t.Fatalf("event name = %q, want %q", evtMsg.Event.Name, "connection.established")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestJetStreamKV(t *testing.T) {
	s, url := startTestNATS(t)
	defer s.Shutdown()

	ctx := context.Background()
	conn := connectTest(t, url, "test")
	defer conn.Close()

	// Initialize KV bucket
	kv, err := natsconn.InitKV(ctx, conn, "AMUX_KV_TEST")
	if err != nil {
		t.Fatalf("InitKV: %v", err)
	}

	// Put and get host info
	info := &natsconn.HostInfo{
		Version:   "0.1.0",
		OS:        "linux",
		Arch:      "amd64",
		PeerID:    "5678",
		StartedAt: "2026-01-18T10:30:00Z",
	}
	if err := kv.PutHostInfo(ctx, "devbox", info); err != nil {
		t.Fatalf("PutHostInfo: %v", err)
	}

	got, err := kv.GetHostInfo(ctx, "devbox")
	if err != nil {
		t.Fatalf("GetHostInfo: %v", err)
	}

	if got.PeerID != "5678" {
		t.Fatalf("PeerID = %q, want %q", got.PeerID, "5678")
	}
	if got.OS != "linux" {
		t.Fatalf("OS = %q, want %q", got.OS, "linux")
	}

	// Put heartbeat
	if err := kv.PutHeartbeat(ctx, "devbox"); err != nil {
		t.Fatalf("PutHeartbeat: %v", err)
	}

	hb, err := kv.GetHeartbeat(ctx, "devbox")
	if err != nil {
		t.Fatalf("GetHeartbeat: %v", err)
	}
	if hb.Timestamp == "" {
		t.Fatal("heartbeat timestamp is empty")
	}

	// Put and get session meta
	meta := &natsconn.SessionMeta{
		AgentID:   "42",
		AgentSlug: "backend-dev",
		RepoPath:  "~/projects/my-repo",
		State:     "running",
	}
	if err := kv.PutSessionMeta(ctx, "devbox", "9001", meta); err != nil {
		t.Fatalf("PutSessionMeta: %v", err)
	}

	gotMeta, err := kv.GetSessionMeta(ctx, "devbox", "9001")
	if err != nil {
		t.Fatalf("GetSessionMeta: %v", err)
	}

	if gotMeta.AgentID != "42" {
		t.Fatalf("AgentID = %q, want %q", gotMeta.AgentID, "42")
	}
	if gotMeta.State != "running" {
		t.Fatalf("State = %q, want %q", gotMeta.State, "running")
	}

	// Delete session meta
	if err := kv.DeleteSessionMeta(ctx, "devbox", "9001"); err != nil {
		t.Fatalf("DeleteSessionMeta: %v", err)
	}
}

func TestKillReplyExchange(t *testing.T) {
	s, url := startTestNATS(t)
	defer s.Shutdown()

	prefix := "amux"
	cfg := testConfig(url)

	dirConn := connectTest(t, url, "director")
	defer dirConn.Close()

	daemonConn := connectTest(t, url, "daemon")
	defer daemonConn.Close()

	// Daemon handles kill
	_, _ = daemonConn.NC().Subscribe(
		protocol.ControlSubject(prefix, "devbox"),
		func(msg *nats.Msg) {
			var ctlMsg protocol.ControlMessage
			_ = json.Unmarshal(msg.Data, &ctlMsg)

			if ctlMsg.Type == protocol.TypeKill {
				var req protocol.KillRequest
				_ = ctlMsg.DecodePayload(&req)
				resp, _ := protocol.NewControlMessage(protocol.TypeKill, &protocol.KillResponse{
					SessionID: req.SessionID,
					Killed:    true,
				})
				data, _ := json.Marshal(resp)
				_ = msg.Respond(data)
			}
		},
	)
	_ = daemonConn.Flush()

	// Director sends kill
	killReq := &protocol.KillRequest{SessionID: "9001"}
	killMsg, _ := protocol.NewControlMessage(protocol.TypeKill, killReq)
	killData, _ := json.Marshal(killMsg)

	reply, err := dirConn.Request(
		protocol.ControlSubject(prefix, "devbox"),
		killData,
		cfg.Remote.RequestTimeout.Duration,
	)
	if err != nil {
		t.Fatalf("kill request: %v", err)
	}

	var respMsg protocol.ControlMessage
	_ = json.Unmarshal(reply.Data, &respMsg)

	var resp protocol.KillResponse
	_ = respMsg.DecodePayload(&resp)

	if !resp.Killed {
		t.Fatal("kill response: Killed = false, want true")
	}
	if resp.SessionID != "9001" {
		t.Fatalf("SessionID = %q, want %q", resp.SessionID, "9001")
	}
}

func TestReplayReplyExchange(t *testing.T) {
	s, url := startTestNATS(t)
	defer s.Shutdown()

	prefix := "amux"
	cfg := testConfig(url)

	dirConn := connectTest(t, url, "director")
	defer dirConn.Close()

	daemonConn := connectTest(t, url, "daemon")
	defer daemonConn.Close()

	// Daemon handles replay
	_, _ = daemonConn.NC().Subscribe(
		protocol.ControlSubject(prefix, "devbox"),
		func(msg *nats.Msg) {
			var ctlMsg protocol.ControlMessage
			_ = json.Unmarshal(msg.Data, &ctlMsg)

			if ctlMsg.Type == protocol.TypeReplay {
				var req protocol.ReplayRequest
				_ = ctlMsg.DecodePayload(&req)
				resp, _ := protocol.NewControlMessage(protocol.TypeReplay, &protocol.ReplayResponse{
					SessionID: req.SessionID,
					Accepted:  true,
				})
				data, _ := json.Marshal(resp)
				_ = msg.Respond(data)

				// Publish replay bytes
				_ = daemonConn.Publish(
					protocol.PTYOutputSubject(prefix, "devbox", req.SessionID),
					[]byte("replayed output"),
				)
			}
		},
	)

	// Director subscribes to PTY output
	ptyReceived := make(chan []byte, 10)
	_, _ = dirConn.NC().Subscribe(
		protocol.PTYOutputSubject(prefix, "devbox", "9001"),
		func(msg *nats.Msg) {
			ptyReceived <- msg.Data
		},
	)
	_ = daemonConn.Flush()
	_ = dirConn.Flush()

	// Director sends replay request
	replayReq := &protocol.ReplayRequest{SessionID: "9001"}
	replayMsg, _ := protocol.NewControlMessage(protocol.TypeReplay, replayReq)
	replayData, _ := json.Marshal(replayMsg)

	reply, err := dirConn.Request(
		protocol.ControlSubject(prefix, "devbox"),
		replayData,
		cfg.Remote.RequestTimeout.Duration,
	)
	if err != nil {
		t.Fatalf("replay request: %v", err)
	}

	var respMsg protocol.ControlMessage
	_ = json.Unmarshal(reply.Data, &respMsg)

	var resp protocol.ReplayResponse
	_ = respMsg.DecodePayload(&resp)

	if !resp.Accepted {
		t.Fatal("replay response: Accepted = false, want true")
	}

	// Should receive replayed PTY output
	select {
	case data := <-ptyReceived:
		if string(data) != "replayed output" {
			t.Fatalf("replayed data = %q, want %q", data, "replayed output")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for replay output")
	}
}

func TestDirectorDirectFlow(t *testing.T) {
	s, url := startTestNATS(t)
	defer s.Shutdown()

	ctx := context.Background()
	cfg := testConfig(url)
	dispatcher := event.NewLocalDispatcher()

	// Start director
	dirConn := connectTest(t, url, "director")
	defer dirConn.Close()

	dir := &directorHelper{
		conn:       dirConn,
		cfg:        cfg,
		dispatcher: dispatcher,
	}

	// Simulate a daemon doing handshake and responding to spawn
	daemonConn := connectTest(t, url, "daemon")
	defer daemonConn.Close()

	// Daemon subscribes to handshake and control
	hsComplete := make(chan struct{})
	_, _ = dirConn.NC().Subscribe("amux.handshake.*", func(msg *nats.Msg) {
		resp, _ := protocol.NewControlMessage(protocol.TypeHandshake, &protocol.HandshakePayload{
			Protocol: 1,
			PeerID:   "1234",
			Role:     "director",
			HostID:   "dir-host",
		})
		data, _ := json.Marshal(resp)
		_ = msg.Respond(data)
		close(hsComplete)
	})
	_ = dirConn.Flush()

	// Daemon sends handshake
	hs, _ := protocol.NewControlMessage(protocol.TypeHandshake, &protocol.HandshakePayload{
		Protocol: 1,
		PeerID:   "5678",
		Role:     "daemon",
		HostID:   "devbox",
	})
	hsData, _ := json.Marshal(hs)
	_, err := daemonConn.Request("amux.handshake.devbox", hsData, 5*time.Second)
	if err != nil {
		t.Fatalf("handshake: %v", err)
	}
	_ = ctx
	_ = dir

	select {
	case <-hsComplete:
	case <-time.After(2 * time.Second):
		t.Fatal("handshake not received")
	}
}

// directorHelper is a minimal wrapper for testing.
type directorHelper struct {
	conn       *natsconn.Conn
	cfg        *config.Config
	dispatcher event.Dispatcher
}

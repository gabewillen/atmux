//go:build integration
// +build integration

package integrationtest

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/remote"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/docker/go-connections/nat"
	"github.com/nats-io/nats.go"
	testcontainers "github.com/testcontainers/testcontainers-go"
)

type outputEvent struct {
	line   string
	isTick bool
	tick   int
}

func TestIntegrationPhase3RemoteOrchestration(t *testing.T) {
	harness, err := NewHarness(t)
	if err != nil {
		t.Fatalf("harness: %v", err)
	}
	ctx := harness.Context()
	repoRoot := initPhase2Repo(t)
	resolver, err := paths.NewResolver(repoRoot)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	credStore, err := remote.NewCredentialStore(filepath.Join(t.TempDir(), "creds"))
	if err != nil {
		t.Fatalf("cred store: %v", err)
	}
	if _, err := credStore.DirectorCredential(); err != nil {
		t.Fatalf("director creds: %v", err)
	}
	hostID := api.MustParseHostID("phase3-host")
	if _, err := credStore.GetOrCreate(hostID.String(), "amux", "AMUX_KV"); err != nil {
		t.Fatalf("host creds: %v", err)
	}
	auth, err := credStore.HubAuth()
	if err != nil {
		t.Fatalf("hub auth: %v", err)
	}
	configPath, operatorPath := writeNATSConfig(t, auth)
	natsContainer, err := harness.StartNATS(ctx, NATSContainerOptions{
		ExposedPorts: []string{natsPort, "7422/tcp"},
		Cmd:          []string{"-c", "/etc/nats/nats.conf"},
		Files: []testcontainers.ContainerFile{
			{HostFilePath: configPath, ContainerFilePath: "/etc/nats/nats.conf", FileMode: 0o644},
			{HostFilePath: operatorPath, ContainerFilePath: "/etc/nats/operator.jwt", FileMode: 0o600},
		},
	})
	if err != nil {
		t.Fatalf("nats: %v", err)
	}
	toxiproxy, err := harness.StartToxiproxy(ctx)
	if err != nil {
		t.Fatalf("toxiproxy: %v", err)
	}
	client := toxiproxy.Client()
	hubProxyName := "hub-proxy"
	leafProxyName := "leaf-proxy"
	if err := client.CreateProxy(ctx, hubProxyName, "0.0.0.0:8666", fmt.Sprintf("%s:4222", natsContainer.Alias)); err != nil {
		t.Fatalf("create proxy: %v", err)
	}
	if err := client.CreateProxy(ctx, leafProxyName, "0.0.0.0:8667", fmt.Sprintf("%s:7422", natsContainer.Alias)); err != nil {
		t.Fatalf("create leaf proxy: %v", err)
	}
	leafProxyPort, err := toxiproxy.Container.MappedPort(ctx, nat.Port("8667/tcp"))
	if err != nil {
		t.Fatalf("leaf proxy port: %v", err)
	}
	hubProxyURL := "nats://" + toxiproxy.ProxyAddress()
	leafProxyURL := fmt.Sprintf("nats://%s:%s", toxiproxy.Host, leafProxyPort.Port())
	if err := client.AddLatency(ctx, leafProxyName, 100*time.Millisecond, 25*time.Millisecond); err != nil {
		t.Fatalf("add latency: %v", err)
	}
	if err := client.AddTimeout(ctx, leafProxyName, 150*time.Millisecond); err != nil {
		t.Fatalf("add timeout: %v", err)
	}
	select {
	case <-time.After(250 * time.Millisecond):
	}
	if err := client.RemoveToxic(ctx, leafProxyName, "timeout"); err != nil {
		t.Fatalf("remove timeout: %v", err)
	}
	dirDisp, err := protocol.NewNATSDispatcher(ctx, natsContainer.URL, protocol.NATSOptions{CredsPath: credStore.CredentialPath("director")})
	if err != nil {
		t.Fatalf("director dispatcher: %v", err)
	}
	defer func() {
		_ = dirDisp.Close(context.Background())
	}()
	dirCfg := config.DefaultConfig(resolver)
	dirCfg.Remote.NATS.SubjectPrefix = "amux"
	dirCfg.NATS.JetStreamDir = filepath.Join(t.TempDir(), "director")
	director, err := remote.NewDirector(dirCfg, dirDisp, remote.DirectorOptions{Version: "test", HostID: api.MustParseHostID("director")})
	if err != nil {
		t.Fatalf("director: %v", err)
	}
	if err := director.Start(ctx); err != nil {
		t.Fatalf("director start: %v", err)
	}
	mgrCfg := config.DefaultConfig(resolver)
	mgrCfg.Node.Role = "manager"
	mgrCfg.Remote.NATS.URL = leafProxyURL
	mgrCfg.Remote.NATS.CredsPath = credStore.CredentialPath(hostID.String())
	mgrCfg.Remote.NATS.SubjectPrefix = "amux"
	mgrCfg.Remote.Manager.HostID = hostID.String()
	mgrCfg.Remote.BufferSize = config.ByteSize(4096)
	mgrCfg.Remote.RequestTimeout = 2 * time.Second
	mgrCfg.NATS.HubURL = hubProxyURL
	mgrCfg.NATS.JetStreamDir = filepath.Join(t.TempDir(), "manager")
	mgrCfg.NATS.Listen = "127.0.0.1:-1"
	hostMgr, err := remote.NewHostManager(mgrCfg, resolver, "test")
	if err != nil {
		t.Fatalf("host manager: %v", err)
	}
	hostMgr.SetRegistry(&phase2Registry{cmd: []string{os.Args[0], "-test.run=TestIntegrationPhase3Helper"}}, nil)
	startErr := make(chan error, 1)
	go func() {
		startErr <- hostMgr.Start(ctx)
	}()
	waitForHostReady(t, director, hostID, startErr, 10*time.Second)
	time.Sleep(250 * time.Millisecond)
	spawnReq := remote.SpawnRequest{
		AgentID:   api.NewAgentID().String(),
		AgentSlug: "phase3-agent",
		RepoPath:  repoRoot,
		Adapter:   "stub",
		Command:   []string{os.Args[0], "-test.run=TestIntegrationPhase3Helper"},
		Env: map[string]string{
			"AMUX_REMOTE_HELPER": "1",
		},
	}
	spawnResp, err := retrySpawn(ctx, director, hostID, spawnReq, 5*time.Second)
	if err != nil {
		t.Fatalf("spawn: %v", err)
	}
	sessionID, err := api.ParseSessionID(spawnResp.SessionID)
	if err != nil {
		t.Fatalf("session id: %v", err)
	}
	conn, err := director.AttachPTY(ctx, hostID, sessionID)
	if err != nil {
		t.Fatalf("attach pty: %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()
	outputCh := make(chan outputEvent, 256)
	go streamOutput(conn, outputCh)
	if _, err := conn.Write([]byte("ping\n")); err != nil {
		t.Fatalf("send input: %v", err)
	}
	waitForPrefix(t, outputCh, "echo:ping", 5*time.Second)
	lastTick := waitForTick(t, outputCh, 3, 5*time.Second)
	if err := client.SetProxyEnabled(ctx, hubProxyName, false); err != nil {
		t.Fatalf("disable hub proxy: %v", err)
	}
	if err := client.SetProxyEnabled(ctx, leafProxyName, false); err != nil {
		t.Fatalf("disable leaf proxy: %v", err)
	}
	time.Sleep(750 * time.Millisecond)
	if err := client.SetProxyEnabled(ctx, hubProxyName, true); err != nil {
		t.Fatalf("enable hub proxy: %v", err)
	}
	if err := client.SetProxyEnabled(ctx, leafProxyName, true); err != nil {
		t.Fatalf("enable leaf proxy: %v", err)
	}
	waitForHostReady(t, director, hostID, startErr, 15*time.Second)
	time.Sleep(250 * time.Millisecond)
	drainOutput(outputCh)
	replayResp, err := retryReplay(ctx, director, hostID, sessionID, 10*time.Second)
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if !replayResp.Accepted {
		t.Fatalf("expected replay accepted")
	}
	postTicks := collectTicks(outputCh, 3*time.Second)
	if len(postTicks) == 0 {
		t.Fatalf("expected ticks after replay")
	}
	if !isNonDecreasing(postTicks) {
		t.Fatalf("expected replay output ordered, got %v", postTicks)
	}
	if !containsGreater(postTicks, lastTick) {
		t.Fatalf("expected tick greater than %d, got %v", lastTick, postTicks)
	}
	killResp, err := director.Kill(ctx, hostID, remote.KillRequest{SessionID: sessionID.String()})
	if err != nil {
		t.Fatalf("kill: %v", err)
	}
	if !killResp.Killed {
		t.Fatalf("expected kill to report killed")
	}
}

func TestIntegrationPhase3Helper(t *testing.T) {
	if os.Getenv("AMUX_REMOTE_HELPER") != "1" {
		return
	}
	writer := bufio.NewWriter(os.Stdout)
	flush := func() {
		_ = writer.Flush()
	}
	inputCh := make(chan string, 8)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			inputCh <- scanner.Text()
		}
		close(inputCh)
	}()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	count := 0
	for {
		select {
		case line, ok := <-inputCh:
			if !ok {
				return
			}
			_, _ = fmt.Fprintf(writer, "echo:%s\n", line)
			flush()
		case <-ticker.C:
			count++
			_, _ = fmt.Fprintf(writer, "tick:%d\n", count)
			flush()
		}
	}
}

func streamOutput(conn net.Conn, out chan<- outputEvent) {
	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			close(out)
			return
		}
		trimmed := strings.TrimSpace(line)
		event := outputEvent{line: trimmed}
		if strings.HasPrefix(trimmed, "tick:") {
			value := strings.TrimPrefix(trimmed, "tick:")
			if tick, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
				event.isTick = true
				event.tick = tick
			}
		}
		out <- event
	}
}

func waitForPrefix(t *testing.T, out <-chan outputEvent, prefix string, timeout time.Duration) {
	t.Helper()
	deadline := time.After(timeout)
	for {
		select {
		case event, ok := <-out:
			if !ok {
				t.Fatalf("output closed while waiting for %q", prefix)
			}
			if strings.HasPrefix(event.line, prefix) {
				return
			}
		case <-deadline:
			t.Fatalf("timeout waiting for %q", prefix)
		}
	}
}

func waitForTick(t *testing.T, out <-chan outputEvent, min int, timeout time.Duration) int {
	t.Helper()
	deadline := time.After(timeout)
	for {
		select {
		case event, ok := <-out:
			if !ok {
				t.Fatalf("output closed while waiting for tick")
			}
			if event.isTick && event.tick >= min {
				return event.tick
			}
		case <-deadline:
			t.Fatalf("timeout waiting for tick >= %d", min)
		}
	}
}

func drainOutput(out <-chan outputEvent) {
	for {
		select {
		case _, ok := <-out:
			if !ok {
				return
			}
		default:
			return
		}
	}
}

func collectTicks(out <-chan outputEvent, duration time.Duration) []int {
	deadline := time.After(duration)
	var ticks []int
	for {
		select {
		case event, ok := <-out:
			if !ok {
				return ticks
			}
			if event.isTick {
				ticks = append(ticks, event.tick)
			}
		case <-deadline:
			return ticks
		}
	}
}

func isNonDecreasing(values []int) bool {
	for i := 1; i < len(values); i++ {
		if values[i] < values[i-1] {
			return false
		}
	}
	return true
}

func containsGreater(values []int, min int) bool {
	for _, value := range values {
		if value > min {
			return true
		}
	}
	return false
}

func waitForHostReady(t *testing.T, director *remote.Director, hostID api.HostID, startErr <-chan error, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		select {
		case err := <-startErr:
			if err != nil {
				t.Fatalf("host manager: %v", err)
			}
		case <-time.After(100 * time.Millisecond):
		}
		if director.HostReady(hostID) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timeout waiting for host ready")
		}
	}
}

func writeNATSConfig(t *testing.T, auth remote.HubAuth) (string, string) {
	t.Helper()
	if auth.OperatorJWT == "" {
		t.Fatalf("operator jwt missing")
	}
	base := t.TempDir()
	operatorPath := filepath.Join(base, "operator.jwt")
	if err := os.WriteFile(operatorPath, []byte(auth.OperatorJWT), 0o600); err != nil {
		t.Fatalf("write operator jwt: %v", err)
	}
	configPath := filepath.Join(base, "nats.conf")
	config := fmt.Sprintf(`port: 4222
operator: "/etc/nats/operator.jwt"
system_account: %s
resolver = MEMORY
resolver_preload = {
  %s: %q
  %s: %q
}
jetstream {
  store_dir: "/data/jetstream"
}
leafnodes {
  listen: "0.0.0.0:7422"
}
`, auth.SystemAccountKey, auth.SystemAccountKey, auth.SystemAccountJWT, auth.AccountPublicKey, auth.AccountJWT)
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("write nats config: %v", err)
	}
	return configPath, operatorPath
}

func retrySpawn(ctx context.Context, director *remote.Director, hostID api.HostID, req remote.SpawnRequest, timeout time.Duration) (remote.SpawnResponse, error) {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for {
		resp, err := director.Spawn(ctx, hostID, req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if !isTransientControlErr(err) || time.Now().After(deadline) {
			return remote.SpawnResponse{}, lastErr
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func retryReplay(ctx context.Context, director *remote.Director, hostID api.HostID, sessionID api.SessionID, timeout time.Duration) (remote.ReplayResponse, error) {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for {
		resp, err := director.Replay(ctx, hostID, remote.ReplayRequest{SessionID: sessionID.String()})
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if !isTransientControlErr(err) || time.Now().After(deadline) {
			return remote.ReplayResponse{}, lastErr
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func isTransientControlErr(err error) bool {
	return errors.Is(err, remote.ErrHostDisconnected) ||
		errors.Is(err, remote.ErrNotReady) ||
		errors.Is(err, nats.ErrNoResponders) ||
		errors.Is(err, context.DeadlineExceeded)
}

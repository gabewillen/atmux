package remote

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/git"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestRemoteHandshakeAndSpawn(t *testing.T) {
	repoRoot := initRepo(t)
	resolver, err := paths.NewResolver(repoRoot)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	hostID := api.MustParseHostID("testhost")
	jetstreamDir := filepath.Join(t.TempDir(), "nats")
	credStore, err := NewCredentialStore(jetstreamDir)
	if err != nil {
		t.Fatalf("cred store: %v", err)
	}
	auth, err := credStore.HubAuth()
	if err != nil {
		t.Fatalf("hub auth: %v", err)
	}
	server, err := protocol.StartHubServer(ctx, protocol.HubServerConfig{
		Listen:            "127.0.0.1:-1",
		JetStreamDir:      jetstreamDir,
		OperatorPublicKey: auth.OperatorPublicKey,
		SystemAccountKey:  auth.SystemAccountKey,
		SystemAccountJWT:  auth.SystemAccountJWT,
		AccountPublicKey:  auth.AccountPublicKey,
		AccountJWT:        auth.AccountJWT,
	})
	if err != nil {
		t.Fatalf("start nats: %v", err)
	}
	defer func() {
		_ = server.Close()
	}()
	if _, err := credStore.DirectorCredential(); err != nil {
		t.Fatalf("director creds: %v", err)
	}
	dirDisp, err := protocol.NewNATSDispatcher(ctx, server.URL(), protocol.NATSOptions{CredsPath: credStore.CredentialPath("director")})
	if err != nil {
		t.Fatalf("director dispatcher: %v", err)
	}
	defer func() {
		_ = dirDisp.Close(context.Background())
	}()
	cfg := config.DefaultConfig(resolver)
	cfg.Remote.NATS.URL = server.LeafURL()
	cfg.NATS.JetStreamDir = jetstreamDir
	director, err := NewDirector(cfg, dirDisp, DirectorOptions{Version: "test", HostID: api.MustParseHostID("director")})
	if err != nil {
		t.Fatalf("director: %v", err)
	}
	if err := director.Start(ctx); err != nil {
		t.Fatalf("director start: %v", err)
	}
	handshakeSeen := make(chan struct{}, 1)
	_, err = dirDisp.SubscribeRaw(ctx, "amux.handshake.*", func(msg protocol.Message) {
		select {
		case handshakeSeen <- struct{}{}:
		default:
		}
	})
	if err != nil {
		t.Fatalf("handshake subscribe: %v", err)
	}
	if _, err := credStore.GetOrCreate(hostID.String(), "amux", "AMUX_KV"); err != nil {
		t.Fatalf("host creds: %v", err)
	}
	mgrCfg := config.DefaultConfig(resolver)
	mgrCfg.Node.Role = "manager"
	mgrCfg.Remote.NATS.URL = server.LeafURL()
	mgrCfg.Remote.NATS.CredsPath = credStore.CredentialPath(hostID.String())
	mgrCfg.Remote.NATS.SubjectPrefix = "amux"
	mgrCfg.Remote.Manager.HostID = hostID.String()
	mgrCfg.NATS.JetStreamDir = jetstreamDir
	mgrCfg.NATS.HubURL = server.URL()
	mgrCfg.NATS.Listen = "127.0.0.1:-1"
	hostMgr, err := NewHostManager(mgrCfg, resolver, "test")
	if err != nil {
		t.Fatalf("host manager: %v", err)
	}
	hostMgr.SetRegistry(stubRegistry{}, nil)
	startErr := make(chan error, 1)
	go func() {
		startErr <- hostMgr.Start(ctx)
	}()
	select {
	case <-handshakeSeen:
	case <-time.After(2 * time.Second):
		t.Fatalf("handshake not observed")
	}
	deadline := time.Now().Add(5 * time.Second)
	for !director.HostReady(hostID) {
		select {
		case err := <-startErr:
			if err != nil {
				t.Fatalf("host manager: %v", err)
			}
		default:
		}
		if time.Now().After(deadline) {
			t.Fatalf("host not ready")
		}
		time.Sleep(50 * time.Millisecond)
	}
	spawnReq := SpawnRequest{
		AgentID:   api.NewAgentID().String(),
		AgentSlug: "test-agent",
		RepoPath:  repoRoot,
		Adapter:   "stub",
		Command:   []string{os.Args[0], "-test.run=TestRemoteHelperProcess"},
		Env: map[string]string{
			"AMUX_HELPER": "1",
		},
	}
	resp, err := director.Spawn(ctx, hostID, spawnReq)
	if err != nil {
		t.Fatalf("spawn: %v", err)
	}
	sessionID, err := api.ParseSessionID(resp.SessionID)
	if err != nil {
		t.Fatalf("parse session: %v", err)
	}
	outSubject := PtyOutSubject("amux", hostID, sessionID)
	output := make(chan string, 1)
	_, err = dirDisp.SubscribeRaw(ctx, outSubject, func(msg protocol.Message) {
		text := string(msg.Data)
		if strings.Contains(text, "hello") {
			select {
			case output <- text:
			default:
			}
		}
	})
	if err != nil {
		t.Fatalf("subscribe output: %v", err)
	}
	select {
	case <-output:
	case <-time.After(5 * time.Second):
		t.Fatalf("timeout waiting for output")
	}
	replayResp, err := director.Replay(ctx, hostID, ReplayRequest{SessionID: sessionID.String()})
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if !replayResp.Accepted {
		t.Fatalf("expected replay accepted")
	}
}

func TestRemoteHelperProcess(t *testing.T) {
	if os.Getenv("AMUX_HELPER") != "1" {
		return
	}
	writer := bufio.NewWriter(os.Stdout)
	_, _ = writer.WriteString("hello\n")
	_ = writer.Flush()
	time.Sleep(200 * time.Millisecond)
}

func initRepo(t *testing.T) string {
	repoRoot := t.TempDir()
	runGit(t, repoRoot, "init")
	runGit(t, repoRoot, "config", "user.email", "test@example.com")
	runGit(t, repoRoot, "config", "user.name", "Test")
	path := filepath.Join(repoRoot, "README.md")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	runGit(t, repoRoot, "add", "README.md")
	runGit(t, repoRoot, "commit", "-m", "init")
	return repoRoot
}

func runGit(t *testing.T, dir string, args ...string) {
	result, err := execGit(context.Background(), dir, args...)
	if err != nil {
		t.Fatalf("git %v: %v (output: %s)", args, err, string(result.Output))
	}
}

func execGit(ctx context.Context, dir string, args ...string) (git.ExecResult, error) {
	runner := git.NewRunner()
	return runner.Exec(ctx, dir, args...)
}

type stubRegistry struct{}

func (stubRegistry) Load(ctx context.Context, name string) (adapter.Adapter, error) {
	return stubAdapter{name: name}, nil
}

type stubAdapter struct {
	name string
}

func (s stubAdapter) Name() string {
	return s.name
}

func (s stubAdapter) Manifest() adapter.Manifest {
	return adapter.Manifest{Name: s.name}
}

func (s stubAdapter) Matcher() adapter.PatternMatcher {
	return stubMatcher{}
}

func (s stubAdapter) Formatter() adapter.ActionFormatter {
	return stubFormatter{}
}

type stubMatcher struct{}

func (stubMatcher) Match(ctx context.Context, output []byte) ([]adapter.PatternMatch, error) {
	return nil, nil
}

type stubFormatter struct{}

func (stubFormatter) Format(ctx context.Context, input string) (string, error) {
	return input, nil
}

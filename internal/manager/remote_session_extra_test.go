package manager

import (
	"context"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/remote"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestAddRemoteAgentNoDirector(t *testing.T) {
	mgr := &Manager{}
	_, err := mgr.addRemoteAgent(context.Background(), AddRequest{}, api.Location{Type: api.LocationSSH, Host: "host"}, "/repo", true)
	if err == nil {
		t.Fatalf("expected add remote agent error")
	}
}

func TestStartRemoteSessionNilState(t *testing.T) {
	mgr := &Manager{}
	if err := mgr.startRemoteSession(context.Background(), api.NewAgentID(), nil); err == nil {
		t.Fatalf("expected start remote session error")
	}
}

func TestStartRemoteSessionAlreadyRunning(t *testing.T) {
	id := api.NewAgentID()
	state := &agentState{remoteSession: api.NewSessionID()}
	mgr := &Manager{}
	if err := mgr.startRemoteSession(context.Background(), id, state); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSpawnRemoteTimeout(t *testing.T) {
	mgr := &Manager{
		cfg: config.Config{Remote: config.RemoteConfig{RequestTimeout: time.Millisecond}},
	}
	dir := &remote.Director{}
	setUnexportedField(dir, "subjectPrefix", "amux")
	setUnexportedField(dir, "requestTimeout", time.Millisecond)
	mgr.remoteDirector = dir
	_, err := mgr.spawnRemote(context.Background(), api.MustParseHostID("host"), remote.SpawnRequest{})
	if err == nil {
		t.Fatalf("expected spawn remote error")
	}
}

func TestSpawnRemoteContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	mgr := &Manager{cfg: config.Config{Remote: config.RemoteConfig{RequestTimeout: time.Second}}}
	dir := &remote.Director{}
	hostID := api.MustParseHostID("host")
	v := reflect.ValueOf(dir).Elem().FieldByName("hosts")
	hostMap := reflect.MakeMap(v.Type())
	statePtr := reflect.New(v.Type().Elem().Elem())
	setUnexportedField(statePtr.Interface(), "hostID", hostID)
	setUnexportedField(statePtr.Interface(), "connected", true)
	setUnexportedField(statePtr.Interface(), "ready", false)
	hostMap.SetMapIndex(reflect.ValueOf(hostID), statePtr)
	reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem().Set(hostMap)
	mgr.remoteDirector = dir
	if _, err := mgr.spawnRemote(ctx, hostID, remote.SpawnRequest{}); err == nil {
		t.Fatalf("expected context error")
	}
}

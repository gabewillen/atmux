package manager

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/agentflare-ai/amux/internal/remote"
	"github.com/agentflare-ai/amux/pkg/api"
)

func setUnexportedField(target any, name string, value any) {
	v := reflect.ValueOf(target).Elem().FieldByName(name)
	reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem().Set(reflect.ValueOf(value))
}

func TestCommSubjectForTargetAndResolveToID(t *testing.T) {
	hostID := api.MustParseHostID("host")
	managerPeer := api.NewPeerID()
	directorPeer := api.NewPeerID()
	director := &remote.Director{}
	setUnexportedField(director, "hostID", hostID)
	setUnexportedField(director, "peerID", directorPeer)
	mgr := &Manager{
		remoteDirector: director,
		managerID:      managerPeer,
		agents:         make(map[api.AgentID]*agentState),
	}
	agentID := api.NewAgentID()
	mgr.agents[agentID] = &agentState{slug: "alpha", remote: true, remoteHost: hostID}
	if subject := mgr.commSubjectForTarget(api.TargetIDFromRuntime(managerPeer.RuntimeID)); subject == "" {
		t.Fatalf("expected manager comm subject")
	}
	if subject := mgr.commSubjectForTarget(api.TargetIDFromRuntime(directorPeer.RuntimeID)); subject == "" {
		t.Fatalf("expected director comm subject")
	}
	if subject := mgr.commSubjectForTarget(api.TargetIDFromRuntime(agentID.RuntimeID)); subject == "" {
		t.Fatalf("expected agent comm subject")
	}
	if _, ok := mgr.resolveToID("director"); !ok {
		t.Fatalf("expected director resolve")
	}
	if _, ok := mgr.resolveToID("manager"); !ok {
		t.Fatalf("expected manager resolve")
	}
	if _, ok := mgr.resolveToID("manager@host"); !ok {
		t.Fatalf("expected manager@host resolve")
	}
	if _, ok := mgr.resolveToID("alpha"); !ok {
		t.Fatalf("expected agent resolve")
	}
}

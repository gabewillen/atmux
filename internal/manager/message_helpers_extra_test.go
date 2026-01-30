package manager

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/agentflare-ai/amux/internal/remote"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestPeerHelpers(t *testing.T) {
	hostID := api.MustParseHostID("host")
	peerID := api.NewPeerID()
	director := &remote.Director{}
	setUnexportedField(director, "hostID", hostID)
	setUnexportedField(director, "peerID", peerID)
	hostsField := reflect.ValueOf(director).Elem().FieldByName("hosts")
	hostMap := reflect.MakeMap(hostsField.Type())
	statePtr := reflect.New(hostMap.Type().Elem().Elem())
	setUnexportedField(statePtr.Interface(), "hostID", hostID)
	setUnexportedField(statePtr.Interface(), "peerID", peerID)
	setUnexportedField(statePtr.Interface(), "connected", true)
	setUnexportedField(statePtr.Interface(), "ready", true)
	hostMap.SetMapIndex(reflect.ValueOf(hostID), statePtr)
	reflect.NewAt(hostsField.Type(), unsafe.Pointer(hostsField.UnsafeAddr())).Elem().Set(hostMap)
	manager := &Manager{remoteDirector: director, managerID: api.NewPeerID()}
	if got := manager.localHostID(); got != hostID {
		t.Fatalf("unexpected local host id")
	}
	if got := manager.directorPeerID(); got.Value() != peerID.Value() {
		t.Fatalf("unexpected director peer id")
	}
	if _, ok := manager.peerForHost(""); ok {
		t.Fatalf("expected empty host false")
	}
	manager.managerID = api.PeerID{}
	if _, ok := manager.peerForHost(hostID); ok {
		t.Fatalf("expected zero manager id false")
	}
	manager.managerID = api.NewPeerID()
	if got, ok := manager.peerForHost(hostID); !ok || got.IsZero() {
		t.Fatalf("expected peer for host")
	}
}

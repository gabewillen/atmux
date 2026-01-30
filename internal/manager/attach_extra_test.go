package manager

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/agentflare-ai/amux/internal/remote"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestAttachAgentErrors(t *testing.T) {
	manager := &Manager{agents: map[api.AgentID]*agentState{}}
	if _, err := manager.AttachAgent(api.NewAgentID()); err == nil {
		t.Fatalf("expected attach error for missing agent")
	}
	agentID := api.NewAgentID()
	manager.agents[agentID] = &agentState{}
	if _, err := manager.AttachAgent(agentID); err == nil {
		t.Fatalf("expected attach error for missing session")
	}
	hostID := api.MustParseHostID("host")
	manager.agents[agentID] = &agentState{remote: true, remoteHost: hostID}
	if _, err := manager.AttachAgent(agentID); err == nil {
		t.Fatalf("expected remote attach error")
	}
	manager.agents[agentID] = &agentState{
		remote:        true,
		remoteHost:    hostID,
		remoteSession: api.NewSessionID(),
	}
	director := &remote.Director{}
	hostsField := reflect.ValueOf(director).Elem().FieldByName("hosts")
	hostMap := reflect.MakeMap(hostsField.Type())
	statePtr := reflect.New(hostMap.Type().Elem().Elem())
	setUnexportedField(statePtr.Interface(), "hostID", hostID)
	setUnexportedField(statePtr.Interface(), "connected", false)
	setUnexportedField(statePtr.Interface(), "ready", false)
	hostMap.SetMapIndex(reflect.ValueOf(hostID), statePtr)
	reflect.NewAt(hostsField.Type(), unsafe.Pointer(hostsField.UnsafeAddr())).Elem().Set(hostMap)
	manager.remoteDirector = director
	if _, err := manager.AttachAgent(agentID); err == nil {
		t.Fatalf("expected attach error for disconnected host")
	}
}

package daemon

import (
	"context"
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

func TestHandleStatusWithHostManager(t *testing.T) {
	hostMgr := &remote.HostManager{}
	setUnexportedField(hostMgr, "connected", true)
	setUnexportedField(hostMgr, "ready", true)
	setUnexportedField(hostMgr, "hostID", api.HostID("host"))
	d := &Daemon{hostMgr: hostMgr}
	resp, err := d.handleStatus(context.Background(), nil)
	if err != nil {
		t.Fatalf("handle status: %v", err)
	}
	status := resp.(daemonStatusResult)
	if !status.HubConnected || !status.Ready || status.HostID != "host" {
		t.Fatalf("unexpected status: %#v", status)
	}
}

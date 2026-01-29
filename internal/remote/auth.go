package remote

import (
	"fmt"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

// HostPermissions returns the per-host subject permissions.
func HostPermissions(prefix string, hostID api.HostID) protocol.Permissions {
	prefix = SubjectPrefix(prefix)
	return protocol.Permissions{
		Publish: []string{
			fmt.Sprintf("%s.handshake.%s", prefix, hostID.String()),
			fmt.Sprintf("%s.events.%s", prefix, hostID.String()),
			fmt.Sprintf("%s.pty.%s.*.out", prefix, hostID.String()),
			fmt.Sprintf("%s.comm.director", prefix),
			fmt.Sprintf("%s.comm.manager.*", prefix),
			fmt.Sprintf("%s.comm.agent.*.>", prefix),
			fmt.Sprintf("%s.comm.broadcast", prefix),
			"_INBOX.>",
		},
		Subscribe: []string{
			fmt.Sprintf("%s.ctl.%s", prefix, hostID.String()),
			fmt.Sprintf("%s.pty.%s.*.in", prefix, hostID.String()),
			fmt.Sprintf("%s.comm.manager.%s", prefix, hostID.String()),
			fmt.Sprintf("%s.comm.agent.%s.>", prefix, hostID.String()),
			fmt.Sprintf("%s.comm.broadcast", prefix),
			"_INBOX.>",
		},
	}
}

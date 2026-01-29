package remote

import (
	"fmt"
	"strings"

	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/jwt/v2"
)

// HostPermissions returns the per-host subject permissions.
func HostPermissions(prefix string, hostID api.HostID, kvBucket string) jwt.Permissions {
	prefix = SubjectPrefix(prefix)
	kvBucket = strings.TrimSpace(kvBucket)
	if kvBucket == "" {
		kvBucket = "AMUX_KV"
	}
	return jwt.Permissions{
		Pub: jwt.Permission{
			Allow: []string{
				"$JS.API.>",
				fmt.Sprintf("$KV.%s.>", kvBucket),
				"_INBOX.>",
				fmt.Sprintf("%s.handshake.%s", prefix, hostID.String()),
				fmt.Sprintf("%s.events.%s", prefix, hostID.String()),
				fmt.Sprintf("%s.pty.%s.*.out", prefix, hostID.String()),
				fmt.Sprintf("%s.comm.director", prefix),
				fmt.Sprintf("%s.comm.manager.*", prefix),
				fmt.Sprintf("%s.comm.agent.*.>", prefix),
				fmt.Sprintf("%s.comm.broadcast", prefix),
			},
		},
		Sub: jwt.Permission{
			Allow: []string{
				fmt.Sprintf("%s.ctl.%s", prefix, hostID.String()),
				fmt.Sprintf("%s.pty.%s.*.in", prefix, hostID.String()),
				fmt.Sprintf("%s.comm.manager.%s", prefix, hostID.String()),
				fmt.Sprintf("%s.comm.agent.%s.>", prefix, hostID.String()),
				fmt.Sprintf("%s.comm.broadcast", prefix),
				"_INBOX.>",
			},
		},
	}
}

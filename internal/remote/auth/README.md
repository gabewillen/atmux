# package auth

`import "github.com/agentflare-ai/amux/internal/remote/auth"`

Package auth provides per-host NATS credential generation and subject
authorization for amux remote orchestration.

Each remote host receives a unique NKey credential pair. The director
generates the credential during SSH bootstrap and provisions it to the
remote host. Subject authorization rules restrict each host to its
own control, events, and PTY subjects.

See spec §5.5.6.4 for authentication and authorization requirements.

- `func HostSubjectPermissions(prefix, hostID string) (publish, subscribe []string)` — HostSubjectPermissions returns the publish and subscribe subject permissions for a given host_id and subject prefix per spec §5.5.6.4.
- `func WriteCredsFile(cred *HostCredential, dir string) (string, error)` — WriteCredsFile writes a NATS credentials file for a host.
- `type HostCredential` — HostCredential holds the NATS authentication material for a single host.

### Functions

#### HostSubjectPermissions

```go
func HostSubjectPermissions(prefix, hostID string) (publish, subscribe []string)
```

HostSubjectPermissions returns the publish and subscribe subject permissions
for a given host_id and subject prefix per spec §5.5.6.4.

These rules MUST be enforced by the NATS server for traffic attributable
to the given host_id.

#### WriteCredsFile

```go
func WriteCredsFile(cred *HostCredential, dir string) (string, error)
```

WriteCredsFile writes a NATS credentials file for a host.
The file contains the NKey seed and is written with mode 0600
per spec §5.5.6.4 ("file permissions no more permissive than 0600").


## type HostCredential

```go
type HostCredential struct {
	// HostID is the host identifier this credential is bound to.
	HostID string

	// Seed is the NKey private seed (starts with "S").
	Seed []byte

	// PublicKey is the NKey public key (starts with "U" for user keys).
	PublicKey string
}
```

HostCredential holds the NATS authentication material for a single host.

### Functions returning HostCredential

#### GenerateHostCredential

```go
func GenerateHostCredential(hostID string) (*HostCredential, error)
```

GenerateHostCredential creates a unique NKey credential for the given host.

Per spec §5.5.6.4: "For each host_id, the director MUST create a unique
NATS credential [...] and MUST associate it with exactly one host_id."



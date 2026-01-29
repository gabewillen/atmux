# package bootstrap

`import "github.com/agentflare-ai/amux/internal/remote/bootstrap"`

Package bootstrap implements SSH-based bootstrapping of remote hosts.

The bootstrap process provisions a remote host to run the amux manager daemon:
 1. Resolve the SSH target host
 2. Construct a bootstrap ZIP payload (binary + adapter WASMs)
 3. Copy the ZIP to the remote host
 4. Unpack and install on the remote host
 5. Provision NATS leaf→hub credentials
 6. Start the daemon if not already running
 7. Verify connectivity

See spec §5.5.2 for bootstrap requirements.

- `func Bootstrap(ctx context.Context, cfg *BootstrapConfig) error` — Bootstrap performs the full bootstrap sequence for a remote host.
- `func addFileToZip(w *zip.Writer, srcPath, zipName string) error` — addFileToZip adds a file to a ZIP archive.
- `func buildBootstrapZIP(payload *BootstrapPayload) ([]byte, error)` — buildBootstrapZIP creates an in-memory ZIP containing the binary and adapters.
- `func scpFile(ctx context.Context, target *HostTarget, localPath, remotePath string) error` — scpFile copies a local file to a remote path via scp.
- `func scpTo(ctx context.Context, target *HostTarget, data []byte, remotePath string) error` — scpTo copies data to a remote path via SSH (using stdin piping).
- `func sshRun(ctx context.Context, target *HostTarget, command string) error` — sshRun executes a command on the remote host via SSH.
- `func waitForDaemonReady(ctx context.Context, cfg *BootstrapConfig) error` — waitForDaemonReady polls the remote daemon status until it reports connected.
- `type BootstrapConfig` — BootstrapConfig holds configuration for the bootstrap process.
- `type BootstrapPayload` — BootstrapPayload describes what to include in the bootstrap ZIP.
- `type HostTarget` — HostTarget describes the SSH target for bootstrapping.

### Functions

#### Bootstrap

```go
func Bootstrap(ctx context.Context, cfg *BootstrapConfig) error
```

Bootstrap performs the full bootstrap sequence for a remote host.

Per spec §5.5.2: the director MUST NOT require a persistent SSH connection
after bootstrap completes successfully.

#### addFileToZip

```go
func addFileToZip(w *zip.Writer, srcPath, zipName string) error
```

addFileToZip adds a file to a ZIP archive.

#### buildBootstrapZIP

```go
func buildBootstrapZIP(payload *BootstrapPayload) ([]byte, error)
```

buildBootstrapZIP creates an in-memory ZIP containing the binary and adapters.

#### scpFile

```go
func scpFile(ctx context.Context, target *HostTarget, localPath, remotePath string) error
```

scpFile copies a local file to a remote path via scp.

#### scpTo

```go
func scpTo(ctx context.Context, target *HostTarget, data []byte, remotePath string) error
```

scpTo copies data to a remote path via SSH (using stdin piping).

#### sshRun

```go
func sshRun(ctx context.Context, target *HostTarget, command string) error
```

sshRun executes a command on the remote host via SSH.

#### waitForDaemonReady

```go
func waitForDaemonReady(ctx context.Context, cfg *BootstrapConfig) error
```

waitForDaemonReady polls the remote daemon status until it reports connected.


## type BootstrapConfig

```go
type BootstrapConfig struct {
	// Target is the SSH target host.
	Target *HostTarget

	// Payload describes the files to include.
	Payload *BootstrapPayload

	// HubNATSURL is the director's hub NATS URL for the manager to connect to.
	HubNATSURL string

	// Credential is the per-host NATS credential material.
	Credential *auth.HostCredential

	// RemoteInstallDir is where to install the binary on the remote host.
	// Default: ~/.local/bin
	RemoteInstallDir string

	// RemoteCredsDir is where to place the credential file on the remote host.
	// Default: ~/.amux/creds
	RemoteCredsDir string

	// RemoteBootstrapDir is the temp directory for the bootstrap ZIP.
	// Default: ~/.amux/bootstrap
	RemoteBootstrapDir string
}
```

BootstrapConfig holds configuration for the bootstrap process.

## type BootstrapPayload

```go
type BootstrapPayload struct {
	// BinaryPath is the path to the amux-manager binary for the remote OS/arch.
	BinaryPath string

	// AdapterPaths is the list of adapter WASM module paths to include.
	AdapterPaths []string
}
```

BootstrapPayload describes what to include in the bootstrap ZIP.

## type HostTarget

```go
type HostTarget struct {
	// HostID is the unique host identifier.
	HostID string

	// Host is the SSH host or alias from ~/.ssh/config.
	Host string

	// User is the SSH user (optional if defined in ssh config).
	User string

	// Port is the SSH port (optional, 0 means default).
	Port int
}
```

HostTarget describes the SSH target for bootstrapping.

### Methods

#### HostTarget.sshArgs

```go
func () sshArgs() []string
```

sshArgs returns the base SSH command arguments.

#### HostTarget.sshTarget

```go
func () sshTarget() string
```

sshTarget returns the SSH target string (user@host or just host).



# package bootstrap

`import "github.com/agentflare-ai/amux/internal/remote/bootstrap"`

- `type Client` — Client wraps an SSH client for bootstrapping.

## type Client

```go
type Client struct {
	client *ssh.Client
}
```

Client wraps an SSH client for bootstrapping.

### Functions returning Client

#### Dial

```go
func Dial(host, user, keyPath string) (*Client, error)
```

Dial connects to the remote host via SSH.
keyInfo is path to private key file. If empty, tries agent.


### Methods

#### Client.Close

```go
func () Close() error
```

Close closes the connection.

#### Client.Exec

```go
func () Exec(cmd string) (string, error)
```

Exec runs a command and returns output.

#### Client.Upload

```go
func () Upload(localPath, remotePath string, mode os.FileMode) error
```

Upload copies a local file to the remote path with permissions.

#### Client.writeViaCat

```go
func () writeViaCat(r io.Reader, remotePath string, mode os.FileMode) error
```



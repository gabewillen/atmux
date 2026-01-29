# package session

`import "github.com/agentflare-ai/amux/internal/session"`

Package session manages owned PTY sessions for local agents.

- `ErrSessionRunning, ErrSessionNotRunning, ErrSessionInvalid`
- `func dupFile(file *os.File) (*os.File, error)`
- `func sendTerminate(proc *os.Process) error`
- `type Command` — Command describes the command used to start an agent.
- `type Config` — Config configures session behavior.
- `type LocalSession` — LocalSession owns a PTY and process for a local agent.

### Variables

#### ErrSessionRunning, ErrSessionNotRunning, ErrSessionInvalid

```go
var (
	// ErrSessionRunning is returned when a session is already running.
	ErrSessionRunning = errors.New("session already running")
	// ErrSessionNotRunning is returned when a session is not running.
	ErrSessionNotRunning = errors.New("session not running")
	// ErrSessionInvalid is returned when session configuration is invalid.
	ErrSessionInvalid = errors.New("session invalid")
)
```


### Functions

#### dupFile

```go
func dupFile(file *os.File) (*os.File, error)
```

#### sendTerminate

```go
func sendTerminate(proc *os.Process) error
```


## type Command

```go
type Command struct {
	// Argv is the command argv.
	Argv []string
	// Env holds additional environment variables.
	Env []string
}
```

Command describes the command used to start an agent.

## type Config

```go
type Config struct {
	// DrainTimeout controls graceful shutdown duration.
	DrainTimeout time.Duration
}
```

Config configures session behavior.

## type LocalSession

```go
type LocalSession struct {
	mu            sync.Mutex
	agent         *agent.Agent
	meta          api.Session
	command       Command
	worktree      string
	dispatcher    protocol.Dispatcher
	monitor       *monitor.Monitor
	tracker       *process.Tracker
	ptyPair       *pty.Pair
	cmd           *exec.Cmd
	done          chan error
	stopRequested bool
	forcedKill    bool
	config        Config
	outputMu      sync.Mutex
	outputs       map[uint64]net.Conn
	nextOutputID  uint64
	writeMu       sync.Mutex
}
```

LocalSession owns a PTY and process for a local agent.

### Functions returning LocalSession

#### NewLocalSession

```go
func NewLocalSession(meta api.Session, runtime *agent.Agent, command Command, worktree string, matcher adapter.PatternMatcher, dispatcher protocol.Dispatcher, cfg Config) (*LocalSession, error)
```

NewLocalSession constructs a LocalSession for an agent.


### Methods

#### LocalSession.Attach

```go
func () Attach() (net.Conn, error)
```

Attach returns a stream for interactive use.

#### LocalSession.Kill

```go
func () Kill(ctx context.Context) error
```

Kill forces session termination.

#### LocalSession.Meta

```go
func () Meta() api.Session
```

Meta returns the session metadata.

#### LocalSession.Restart

```go
func () Restart(ctx context.Context) error
```

Restart stops and starts the session.

#### LocalSession.Send

```go
func () Send(input []byte) error
```

Send writes input bytes to the PTY.

#### LocalSession.Start

```go
func () Start(ctx context.Context) error
```

Start launches the PTY session.

#### LocalSession.Stop

```go
func () Stop(ctx context.Context) error
```

Stop requests graceful termination of the session.

#### LocalSession.fanout

```go
func () fanout(chunk []byte)
```

#### LocalSession.forwardInput

```go
func () forwardInput(conn net.Conn, id uint64)
```

#### LocalSession.handleOutput

```go
func () handleOutput(ctx context.Context, chunk []byte)
```

#### LocalSession.readOutput

```go
func () readOutput(ctx context.Context, master *os.File)
```

#### LocalSession.removeOutput

```go
func () removeOutput(target net.Conn)
```

#### LocalSession.wait

```go
func () wait(ctx context.Context)
```

#### LocalSession.waitForExit

```go
func () waitForExit(ctx context.Context, allowExitError bool) error
```



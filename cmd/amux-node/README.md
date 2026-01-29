Package main provides the unified daemon binary.
This binary serves as both amuxd (daemon) and amux-manager roles,
with the active role determined by configuration and/or flags.

- `func handleDaemon()`
- `func handleStatus()`
- `func handleStop()`
- `func main()`
- `showVersion, roleFlag, hostIDFlag, natsURLFlag, natsCredsFlag` — Command line flags
- `version`

### Constants

#### version

```go
const version = "v1.22.0-phase3"
```


### Variables

#### showVersion, roleFlag, hostIDFlag, natsURLFlag, natsCredsFlag

```go
var (
	showVersion   = flag.Bool("version", false, "Show version and exit")
	roleFlag      = flag.String("role", "", "Role to run as (director|manager)")
	hostIDFlag    = flag.String("host-id", "", "Host identifier")
	natsURLFlag   = flag.String("nats-url", "", "NATS server URL")
	natsCredsFlag = flag.String("nats-creds", "", "NATS credentials file")
)
```

Command line flags


### Functions

#### handleDaemon

```go
func handleDaemon()
```

#### handleStatus

```go
func handleStatus()
```

#### handleStop

```go
func handleStop()
```

#### main

```go
func main()
```



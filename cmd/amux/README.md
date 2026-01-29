Package main provides the amux CLI client.
The CLI communicates with the amux daemon (amuxd) over JSON-RPC.

- `func handleAgentAdd()`
- `func handleAgentAttach()`
- `func handleAgentList()`
- `func handleAgentRemove()`
- `func handleAgentStart()`
- `func handleAgentStop()`
- `func handleTest()`
- `func handleTestRegression()`
- `func loadAgentsFromConfig(resolver *paths.Resolver) ([]*api.Agent, error)` — loadAgentsFromConfig loads agents from .amux/config.toml
- `func main()`
- `func persistAgentConfig(resolver *paths.Resolver, agentData *api.Agent) error` — persistAgentConfig saves agent configuration to .amux/config.toml
- `type TestResult` — TestResult represents a single test result
- `type TestSnapshot` — TestSnapshot represents the structure of amux test snapshots
- `version`

### Constants

#### version

```go
const version = "v1.22.0-phase2"
```


### Functions

#### handleAgentAdd

```go
func handleAgentAdd()
```

#### handleAgentAttach

```go
func handleAgentAttach()
```

#### handleAgentList

```go
func handleAgentList()
```

#### handleAgentRemove

```go
func handleAgentRemove()
```

#### handleAgentStart

```go
func handleAgentStart()
```

#### handleAgentStop

```go
func handleAgentStop()
```

#### handleTest

```go
func handleTest()
```

#### handleTestRegression

```go
func handleTestRegression()
```

#### loadAgentsFromConfig

```go
func loadAgentsFromConfig(resolver *paths.Resolver) ([]*api.Agent, error)
```

loadAgentsFromConfig loads agents from .amux/config.toml

#### main

```go
func main()
```

#### persistAgentConfig

```go
func persistAgentConfig(resolver *paths.Resolver, agentData *api.Agent) error
```

persistAgentConfig saves agent configuration to .amux/config.toml


## type TestResult

```go
type TestResult struct {
	Name     string `toml:"name"`
	Status   string `toml:"status"` // pass|fail|skip
	Error    string `toml:"error,omitempty"`
	Duration string `toml:"duration,omitempty"`
}
```

TestResult represents a single test result

## type TestSnapshot

```go
type TestSnapshot struct {
	RunID       string       `toml:"run_id"`
	SpecVersion string       `toml:"spec_version"`
	StartedAt   time.Time    `toml:"started_at"`
	FinishedAt  time.Time    `toml:"finished_at"`
	ModuleRoot  string       `toml:"module_root"`
	GitCommit   string       `toml:"git_commit,omitempty"`
	TestResults []TestResult `toml:"test_results"`
}
```

TestSnapshot represents the structure of amux test snapshots


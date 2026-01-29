Package main provides the amux CLI client.
The CLI communicates with the amux daemon (amuxd) over JSON-RPC.

- `func handleTest()`
- `func handleTestRegression()`
- `func main()`
- `type TestResult` — TestResult represents a single test result
- `type TestSnapshot` — TestSnapshot represents the structure of amux test snapshots
- `version`

### Constants

#### version

```go
const version = "v1.22.0-phase1"
```


### Functions

#### handleTest

```go
func handleTest()
```

#### handleTestRegression

```go
func handleTestRegression()
```

#### main

```go
func main()
```


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


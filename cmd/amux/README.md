Package main implements the amux CLI client per spec §12.

- `func discoverGitRepoRoot() (string, error)` — discoverGitRepoRoot determines the git repository root for the current working directory.
- `func findModuleRoot() (string, error)` — findModuleRoot searches from the current working directory upward for a go.mod file and returns its directory.
- `func handleAgentAddCommand()`
- `func handleAgentCommand()`
- `func handleTestCommand()`
- `func main()`

### Functions

#### discoverGitRepoRoot

```go
func discoverGitRepoRoot() (string, error)
```

discoverGitRepoRoot determines the git repository root for the current working directory.
It runs `git rev-parse --show-toplevel` and canonicalizes the result per spec §3.23.

#### findModuleRoot

```go
func findModuleRoot() (string, error)
```

findModuleRoot searches from the current working directory upward for a
go.mod file and returns its directory. If no go.mod is found, it returns
an error and amux test must exit non-zero per spec §12.6.1.

#### handleAgentAddCommand

```go
func handleAgentAddCommand()
```

#### handleAgentCommand

```go
func handleAgentCommand()
```

#### handleTestCommand

```go
func handleTestCommand()
```

#### main

```go
func main()
```



Command amux is the main CLI client for amux.

Package main implements the amux CLI client.

- `func checkRegression(snapshotPath string) error` — checkRegression compares the current snapshot with the previous one.
- `func findModuleRoot() (string, error)` — findModuleRoot finds the Go module root directory.
- `func getStatus(result interface{}) string` — getStatus extracts the status from a test result.
- `func main()`
- `func run() error`
- `func runCommand(name string, args ...string) error` — runCommand runs a command and returns an error if it fails.
- `func runTest(args []string) error` — runTest implements the `amux test` command.
- `type TestSnapshot` — TestSnapshot represents a test snapshot.

### Functions

#### checkRegression

```go
func checkRegression(snapshotPath string) error
```

checkRegression compares the current snapshot with the previous one.

#### findModuleRoot

```go
func findModuleRoot() (string, error)
```

findModuleRoot finds the Go module root directory.

#### getStatus

```go
func getStatus(result interface{}) string
```

getStatus extracts the status from a test result.

#### main

```go
func main()
```

#### run

```go
func run() error
```

#### runCommand

```go
func runCommand(name string, args ...string) error
```

runCommand runs a command and returns an error if it fails.

#### runTest

```go
func runTest(args []string) error
```

runTest implements the `amux test` command.


## type TestSnapshot

```go
type TestSnapshot struct {
	Timestamp time.Time
	Results   map[string]interface{}
}
```

TestSnapshot represents a test snapshot.


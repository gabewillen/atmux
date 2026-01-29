# package snapshot

`import "github.com/stateforward/amux/internal/snapshot"`

Package snapshot implements the amux test snapshot functionality per spec §12.6.

- `func Compare(baseline, current *Snapshot) (bool, string)` — Compare compares two snapshots and returns a regression report.
- `func FindLatestSnapshot(moduleRoot string) (string, error)` — FindLatestSnapshot finds the most recent snapshot in the snapshots directory.
- `func GenerateSnapshotPath(moduleRoot string) string` — GenerateSnapshotPath generates a snapshot file path.
- `func Write(snapshot *Snapshot, path string) error` — Write writes a snapshot to a TOML file.
- `type Snapshot` — Snapshot represents a test snapshot per spec §12.6.

### Functions

#### Compare

```go
func Compare(baseline, current *Snapshot) (bool, string)
```

Compare compares two snapshots and returns a regression report.

#### FindLatestSnapshot

```go
func FindLatestSnapshot(moduleRoot string) (string, error)
```

FindLatestSnapshot finds the most recent snapshot in the snapshots directory.

#### GenerateSnapshotPath

```go
func GenerateSnapshotPath(moduleRoot string) string
```

GenerateSnapshotPath generates a snapshot file path.

#### Write

```go
func Write(snapshot *Snapshot, path string) error
```

Write writes a snapshot to a TOML file.


## type Snapshot

```go
type Snapshot struct {
	Timestamp    time.Time         `toml:"timestamp"`
	GoVersion    string            `toml:"go_version"`
	Module       string            `toml:"module"`
	TidyStatus   string            `toml:"tidy_status"`
	VetStatus    string            `toml:"vet_status"`
	LintStatus   string            `toml:"lint_status"`
	TestStatus   string            `toml:"test_status"`
	RaceStatus   string            `toml:"race_status"`
	Coverage     float64           `toml:"coverage"`
	BenchResults map[string]string `toml:"bench_results,omitempty"`
}
```

Snapshot represents a test snapshot per spec §12.6.

### Functions returning Snapshot

#### Create

```go
func Create(moduleRoot string) (*Snapshot, error)
```

Create creates a new snapshot by running the verification sequence per spec §12.6.

#### Read

```go
func Read(path string) (*Snapshot, error)
```

Read reads a snapshot from a TOML file.



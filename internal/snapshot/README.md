# package snapshot

`import "github.com/stateforward/amux/internal/snapshot"`

Package snapshot implements the snapshot functionality for amux test

- `type ComparisonResult` — ComparisonResult holds the result of comparing two snapshots
- `type FileSnapshotPair` — FileSnapshotPair holds two file snapshots for comparison
- `type FileSnapshot` — FileSnapshot represents a snapshot of a file
- `type Snapshot` — Snapshot represents a snapshot of the system state for testing

## type ComparisonResult

```go
type ComparisonResult struct {
	Added    map[string]FileSnapshot     `toml:"added"`
	Removed  map[string]FileSnapshot     `toml:"removed"`
	Modified map[string]FileSnapshotPair `toml:"modified"`
}
```

ComparisonResult holds the result of comparing two snapshots

## type FileSnapshot

```go
type FileSnapshot struct {
	Path    string `toml:"path"`
	Size    int64  `toml:"size"`
	ModTime int64  `toml:"mod_time"`
	Hash    string `toml:"hash"` // SHA256 hash of the file content
}
```

FileSnapshot represents a snapshot of a file

## type FileSnapshotPair

```go
type FileSnapshotPair struct {
	Before FileSnapshot `toml:"before"`
	After  FileSnapshot `toml:"after"`
}
```

FileSnapshotPair holds two file snapshots for comparison

## type Snapshot

```go
type Snapshot struct {
	Timestamp    time.Time               `toml:"timestamp"`
	Version      string                  `toml:"version"`
	GoVersion    string                  `toml:"go_version"`
	Module       string                  `toml:"module"`
	Dependencies map[string]string       `toml:"dependencies"`
	BuildInfo    map[string]interface{}  `toml:"build_info"`
	SystemInfo   map[string]interface{}  `toml:"system_info"`
	Files        map[string]FileSnapshot `toml:"files"`
}
```

Snapshot represents a snapshot of the system state for testing

### Functions returning Snapshot

#### Create

```go
func Create(moduleRoot string) (*Snapshot, error)
```

Create creates a new snapshot of the current system state

#### Load

```go
func Load(path string) (*Snapshot, error)
```

Load loads a snapshot from a TOML file


### Methods

#### Snapshot.Compare

```go
func () Compare(other *Snapshot) *ComparisonResult
```

Compare compares this snapshot with another and returns differences

#### Snapshot.Save

```go
func () Save(path string) error
```

Save saves the snapshot to a TOML file



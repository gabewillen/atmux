# package snapshot

`import "github.com/agentflare-ai/amux/internal/snapshot"`

- `func Compare(old, new *Snapshot) error`
- `func Save(path string, snap *Snapshot) error`
- `type MetaSnapshot`
- `type Snapshot`
- `type SystemSnapshot`

### Functions

#### Compare

```go
func Compare(old, new *Snapshot) error
```

#### Save

```go
func Save(path string, snap *Snapshot) error
```


## type MetaSnapshot

```go
type MetaSnapshot struct {
	Timestamp string `toml:"timestamp"`
	GoVersion string `toml:"go_version"`
}
```

## type Snapshot

```go
type Snapshot struct {
	Meta   MetaSnapshot   `toml:"meta"`
	System SystemSnapshot `toml:"system"`
	Config config.Config  `toml:"config"`
}
```

### Functions returning Snapshot

#### Capture

```go
func Capture() (*Snapshot, error)
```

#### Load

```go
func Load(path string) (*Snapshot, error)
```


## type SystemSnapshot

```go
type SystemSnapshot struct {
	OS   string `toml:"os"`
	Arch string `toml:"arch"`
}
```


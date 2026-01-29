# package plugin

`import "github.com/agentflare-ai/amux/internal/plugin"`

- `type Manager` — Manager manages installed plugins.
- `type Manifest` — Manifest describes a CLI plugin.
- `type Plugin` — Plugin represents an installed plugin.

## type Manager

```go
type Manager struct {
	mu       sync.RWMutex
	registry map[string]*Plugin
}
```

Manager manages installed plugins.

### Functions returning Manager

#### NewManager

```go
func NewManager() *Manager
```

NewManager creates a new plugin manager.


### Methods

#### Manager.Disable

```go
func () Disable(name string) error
```

Disable disables a plugin.

#### Manager.Enable

```go
func () Enable(name string) error
```

Enable enables a plugin.

#### Manager.Install

```go
func () Install(manifest Manifest, path string) error
```

Install registers a plugin (simulated installation).

#### Manager.List

```go
func () List() []*Plugin
```

List returns all installed plugins.

#### Manager.Remove

```go
func () Remove(name string) error
```

Remove removes a plugin.


## type Manifest

```go
type Manifest struct {
	Name        string   `toml:"name"`
	Version     string   `toml:"version"`
	Description string   `toml:"description"`
	Permissions []string `toml:"permissions"`
	Entrypoint  string   `toml:"entrypoint"` // Path to WASM or executable
}
```

Manifest describes a CLI plugin.

### Functions returning Manifest

#### ParseManifest

```go
func ParseManifest(data []byte) (*Manifest, error)
```

ParseManifest parses a plugin manifest.


### Methods

#### Manifest.Validate

```go
func () Validate() error
```

Validate checks required fields.


## type Plugin

```go
type Plugin struct {
	Manifest Manifest
	Enabled  bool
	Path     string // Installation path
}
```

Plugin represents an installed plugin.


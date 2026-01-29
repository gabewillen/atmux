# package plugin

`import "github.com/agentflare-ai/amux/internal/plugin"`

- `type DefaultGater` — DefaultGater implements default permission checking.
- `type Manager` — Manager manages installed plugins.
- `type Manifest` — Manifest describes a CLI plugin.
- `type PermissionGater` — PermissionGater checks if a plugin has the required permissions.
- `type Plugin` — Plugin represents an installed plugin.

## type DefaultGater

```go
type DefaultGater struct{}
```

DefaultGater implements default permission checking.

### Methods

#### DefaultGater.HasPermission

```go
func () HasPermission(p *Plugin, permission string) bool
```

HasPermission checks if the plugin's manifest declares the permission.
Supports basic globbing (e.g. "agent.*").


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


## type PermissionGater

```go
type PermissionGater interface {
	HasPermission(plugin *Plugin, permission string) bool
}
```

PermissionGater checks if a plugin has the required permissions.

## type Plugin

```go
type Plugin struct {
	Manifest Manifest
	Enabled  bool
	Path     string // Installation path
}
```

Plugin represents an installed plugin.


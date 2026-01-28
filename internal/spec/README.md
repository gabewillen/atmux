# package spec

`import "github.com/agentflare-ai/amux/internal/spec"`

- `expectedVersion` — expectedVersion is the authoritative version required by this plan/implementation.
- `func Verify(root string) error` — Verify checks that spec-v1.22.md exists and matches the expected version.

### Constants

#### expectedVersion

```go
const expectedVersion = "Version: v1.22"
```

expectedVersion is the authoritative version required by this plan/implementation.


### Functions

#### Verify

```go
func Verify(root string) error
```

Verify checks that spec-v1.22.md exists and matches the expected version.
Spec §4.3.1



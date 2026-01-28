# package specchecker

`import "github.com/stateforward/amux/internal/specchecker"`

Package specchecker implements verification that spec-v1.22.md is present and version-locked

- `ExpectedSpecVersion`
- `SpecFileName`
- `func CheckSpecPresenceAndVersion(specPath string) error` — CheckSpecPresenceAndVersion verifies that spec-v1.22.md exists and contains the expected version
- `func GetSpecVersion(specPath string) (string, error)` — GetSpecVersion extracts the version from the spec file

### Constants

#### ExpectedSpecVersion

```go
const ExpectedSpecVersion = "v1.22"
```

#### SpecFileName

```go
const SpecFileName = "spec-v1.22.md"
```


### Functions

#### CheckSpecPresenceAndVersion

```go
func CheckSpecPresenceAndVersion(specPath string) error
```

CheckSpecPresenceAndVersion verifies that spec-v1.22.md exists and contains the expected version

#### GetSpecVersion

```go
func GetSpecVersion(specPath string) (string, error)
```

GetSpecVersion extracts the version from the spec file



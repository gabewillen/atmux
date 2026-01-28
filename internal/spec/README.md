# package spec

`import "github.com/agentflare-ai/amux/internal/spec"`

Package spec provides spec version checking and validation.

- `ExpectedSpecVersion, SpecFileName`
- `func CheckSpecVersion(repoRoot string) error` — CheckSpecVersion verifies that the spec file exists and contains the expected version.

### Constants

#### ExpectedSpecVersion, SpecFileName

```go
const (
	// ExpectedSpecVersion is the expected spec version for this implementation.
	ExpectedSpecVersion = "v1.22"

	// SpecFileName is the name of the spec file.
	SpecFileName = "spec-v1.22.md"
)
```


### Functions

#### CheckSpecVersion

```go
func CheckSpecVersion(repoRoot string) error
```

CheckSpecVersion verifies that the spec file exists and contains the expected version.



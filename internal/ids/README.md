# package ids

`import "github.com/stateforward/amux/internal/ids"`

Package ids implements identifier utilities and normalization functions for the amux project

- `func CanonicalizeRepoRoot(path string) (string, error)` — CanonicalizeRepoRoot canonicalizes a repository root path according to spec rules: - expand ~/ to target host's home directory - convert to absolute path - clean .
- `func DecodeID(s string) (muid.MUID, error)` — DecodeID decodes a base-10 string to an muid.MUID
- `func EncodeID(id muid.MUID) string` — EncodeID encodes an muid.MUID as a base-10 string
- `func New() muid.MUID` — New generates a new unique ID using muid
- `func NormalizeAgentSlug(name string) string` — NormalizeAgentSlug normalizes an agent name to a filesystem-safe slug according to the spec rules: - lowercase - non-[a-z0-9-] → '-' - collapse multiple consecutive '-' into single '-' - trim leading/trailing '-' - max 63 chars

### Functions

#### CanonicalizeRepoRoot

```go
func CanonicalizeRepoRoot(path string) (string, error)
```

CanonicalizeRepoRoot canonicalizes a repository root path according to spec rules:
- expand ~/ to target host's home directory
- convert to absolute path
- clean . and .. segments
- resolve symbolic links to their target path where possible

#### DecodeID

```go
func DecodeID(s string) (muid.MUID, error)
```

DecodeID decodes a base-10 string to an muid.MUID

#### EncodeID

```go
func EncodeID(id muid.MUID) string
```

EncodeID encodes an muid.MUID as a base-10 string

#### New

```go
func New() muid.MUID
```

New generates a new unique ID using muid

#### NormalizeAgentSlug

```go
func NormalizeAgentSlug(name string) string
```

NormalizeAgentSlug normalizes an agent name to a filesystem-safe slug
according to the spec rules:
- lowercase
- non-[a-z0-9-] → '-'
- collapse multiple consecutive '-' into single '-'
- trim leading/trailing '-'
- max 63 chars



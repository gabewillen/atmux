# package ids

`import "github.com/agentflare-ai/amux/internal/ids"`

Package ids provides identifier utilities for amux.

This package implements ID generation, validation, and normalization
following the spec requirements for agent_slug (§5.3.1), repo_key (§3.24),
and runtime IDs (§3.21, §3.22).

All runtime IDs use muid.MUID from hsm-go. The value 0 is reserved
as a sentinel (e.g., BroadcastID) and SHALL NOT be assigned as a
runtime ID for agents, processes, sessions, or peers.

- `BroadcastID` — BroadcastID is the sentinel ID value (0) used for broadcast messages.
- `DefaultAgentSlug` — DefaultAgentSlug is used when normalization produces an empty string.
- `MaxAgentSlugLength` — MaxAgentSlugLength is the maximum length for agent_slug values.
- `allowedChars` — allowedChars matches characters that don't need replacement in agent_slug.
- `consecutiveDashes` — consecutiveDashes matches runs of consecutive dashes.
- `func DecodeID(s string) (muid.MUID, error)` — DecodeID decodes a base-10 string to a muid.MUID.
- `func DecodeIDs(strs []string) ([]muid.MUID, error)` — DecodeIDs decodes a slice of base-10 strings to muid.MUID values.
- `func EncodeID(id muid.MUID) string` — EncodeID encodes a muid.MUID as a base-10 string for JSON wire format.
- `func EncodeIDs(ids []muid.MUID) []string` — EncodeIDs encodes a slice of muid.MUID values as base-10 strings.
- `func IsValidRuntimeID(id muid.MUID) bool` — IsValidRuntimeID returns true if the ID is valid for runtime use.
- `func NewID() muid.MUID` — NewID generates a new globally unique runtime ID.
- `func NormalizeAgentSlug(name string) string` — NormalizeAgentSlug normalizes an agent name into a filesystem-safe slug.
- `func RepoKey(location api.Location, repoRoot string) string` — RepoKey computes the stable repository key from location and repo_root.
- `func RepoKeyHash(location api.Location, repoRoot string) string` — RepoKeyHash returns a truncated SHA-256 hash of the repo_key.
- `func UniqueAgentSlug(name string, exists func(slug string) bool) string` — UniqueAgentSlug returns a unique agent_slug by appending numeric suffixes if needed.

### Constants

#### BroadcastID

```go
const BroadcastID muid.MUID = 0
```

BroadcastID is the sentinel ID value (0) used for broadcast messages.
See spec §3.22.

#### DefaultAgentSlug

```go
const DefaultAgentSlug = "agent"
```

DefaultAgentSlug is used when normalization produces an empty string.

#### MaxAgentSlugLength

```go
const MaxAgentSlugLength = 63
```

MaxAgentSlugLength is the maximum length for agent_slug values.
See spec §5.3.1.


### Variables

#### allowedChars

```go
var allowedChars = regexp.MustCompile(`[^a-z0-9-]`)
```

allowedChars matches characters that don't need replacement in agent_slug.

#### consecutiveDashes

```go
var consecutiveDashes = regexp.MustCompile(`-+`)
```

consecutiveDashes matches runs of consecutive dashes.


### Functions

#### DecodeID

```go
func DecodeID(s string) (muid.MUID, error)
```

DecodeID decodes a base-10 string to a muid.MUID.
Returns an error if the string is not a valid base-10 unsigned integer.

#### DecodeIDs

```go
func DecodeIDs(strs []string) ([]muid.MUID, error)
```

DecodeIDs decodes a slice of base-10 strings to muid.MUID values.
Returns an error if any string is invalid.

#### EncodeID

```go
func EncodeID(id muid.MUID) string
```

EncodeID encodes a muid.MUID as a base-10 string for JSON wire format.
See spec §9.1.3.1.

#### EncodeIDs

```go
func EncodeIDs(ids []muid.MUID) []string
```

EncodeIDs encodes a slice of muid.MUID values as base-10 strings.

#### IsValidRuntimeID

```go
func IsValidRuntimeID(id muid.MUID) bool
```

IsValidRuntimeID returns true if the ID is valid for runtime use.
Runtime IDs must be non-zero per spec §3.22.

#### NewID

```go
func NewID() muid.MUID
```

NewID generates a new globally unique runtime ID.
It will never return 0 (the reserved sentinel value).

#### NormalizeAgentSlug

```go
func NormalizeAgentSlug(name string) string
```

NormalizeAgentSlug normalizes an agent name into a filesystem-safe slug.
The normalization follows spec §5.3.1:

  - Convert to lowercase
  - Replace any character not in [a-z0-9-] with "-"
  - Collapse consecutive "-" characters to a single "-"
  - Trim leading and trailing "-"
  - Truncate to at most 63 characters
  - If the result is empty, use "agent"

#### RepoKey

```go
func RepoKey(location api.Location, repoRoot string) string
```

RepoKey computes the stable repository key from location and repo_root.
See spec §3.24: repo_key is derived from (location.type, location.host, repo_root).

For local agents, the key is "local:<repo_root>".
For SSH agents, the key is "ssh:<host>:<repo_root>".

The repo_root should already be canonicalized per spec §3.23.

#### RepoKeyHash

```go
func RepoKeyHash(location api.Location, repoRoot string) string
```

RepoKeyHash returns a truncated SHA-256 hash of the repo_key.
This can be used when a shorter identifier is needed.

#### UniqueAgentSlug

```go
func UniqueAgentSlug(name string, exists func(slug string) bool) string
```

UniqueAgentSlug returns a unique agent_slug by appending numeric suffixes
if needed. The exists function should return true if a slug is already in use.
See spec §5.3.1 for collision handling.



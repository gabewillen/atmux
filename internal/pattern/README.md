# package pattern

`import "github.com/copilot-claude-sonnet-4/amux/internal/pattern"`

Package pattern provides pattern matching and action interfaces.
This package provides stable interfaces for pattern matching that adapters
will implement, with noop implementations to unblock phased work.

- `ErrNoMatches, ErrMatcherNotAvailable, ErrActionFailed` — Common sentinel errors for pattern operations.
- `type Action` — Action represents an action to be taken based on a pattern match.
- `type Match` — Match represents a pattern match result.
- `type Matcher` — Matcher provides pattern matching functionality.

### Variables

#### ErrNoMatches, ErrMatcherNotAvailable, ErrActionFailed

```go
var (
	// ErrNoMatches indicates no patterns matched the input.
	ErrNoMatches = errors.New("no pattern matches")

	// ErrMatcherNotAvailable indicates pattern matching is not available.
	ErrMatcherNotAvailable = errors.New("pattern matcher not available")

	// ErrActionFailed indicates a pattern action failed to execute.
	ErrActionFailed = errors.New("action execution failed")
)
```

Common sentinel errors for pattern operations.


## type Action

```go
type Action struct {
	// Type is the action type (e.g., "respond", "notify", "execute").
	Type string

	// Payload contains action-specific data.
	Payload map[string]interface{}
}
```

Action represents an action to be taken based on a pattern match.

## type Match

```go
type Match struct {
	// Pattern is the pattern that matched.
	Pattern string

	// Confidence is the match confidence score (0.0-1.0).
	Confidence float64

	// Data contains match-specific data.
	Data map[string]interface{}
}
```

Match represents a pattern match result.

## type Matcher

```go
type Matcher struct {
	available bool
}
```

Matcher provides pattern matching functionality.
Phase 0 provides a noop implementation that returns no matches.

### Functions returning Matcher

#### NewMatcher

```go
func NewMatcher() *Matcher
```

NewMatcher creates a new pattern matcher.


### Methods

#### Matcher.ExecuteAction

```go
func () ExecuteAction(action Action) error
```

ExecuteAction executes the specified action.
Phase 0: Noop implementation.

#### Matcher.IsAvailable

```go
func () IsAvailable() bool
```

IsAvailable returns whether pattern matching is available.

#### Matcher.Match

```go
func () Match(input string) ([]Match, error)
```

Match attempts to match the given input against available patterns.
Phase 0: Returns no matches to unblock later development.



# package hsm

`import "github.com/agentflare-ai/amux/internal/hsm"`

Package hsm provides a minimal hierarchical state machine implementation for Phase 0.

- `type Event` — Event represents a state machine event.
- `type ID` — ID represents a state machine identifier (using muid-like interface for now).
- `type Machine` — Machine represents a hierarchical state machine.
- `type State` — State represents a state in the HSM.
- `type Transition` — Transition represents a state transition.
- `type simpleState` — simpleState implements State interface for simple structs.

## type Event

```go
type Event struct {
	Type   string      `json:"type"`
	Source ID          `json:"source"`
	Data   interface{} `json:"data,omitempty"`
}
```

Event represents a state machine event.

## type ID

```go
type ID uint64
```

ID represents a state machine identifier (using muid-like interface for now).

### Functions returning ID

#### GenerateID

```go
func GenerateID() ID
```

GenerateID creates a new unique ID (simplified muid for Phase 0).


## type Machine

```go
type Machine struct {
	mu       sync.RWMutex
	id       ID
	current  State
	states   map[string]State
	transits []Transition
	handlers map[string][]func(context.Context, Event)
}
```

Machine represents a hierarchical state machine.

### Functions returning Machine

#### NewMachine

```go
func NewMachine(id ID, initialState State) *Machine
```

NewMachine creates a new state machine.


### Methods

#### Machine.AddState

```go
func () AddState(state State)
```

AddState adds a state to the machine.

#### Machine.AddTransition

```go
func () AddTransition(transit Transition)
```

AddTransition adds a transition to the machine.

#### Machine.CurrentState

```go
func () CurrentState() State
```

CurrentState returns the current state.

#### Machine.Dispatch

```go
func () Dispatch(ctx context.Context, ev Event)
```

Dispatch dispatches an event to the state machine.

#### Machine.ID

```go
func () ID() ID
```

ID returns the machine ID.

#### Machine.Subscribe

```go
func () Subscribe(eventType string, handler func(context.Context, Event))
```

Subscribe subscribes to events of a specific type.


## type State

```go
type State interface {
	Name() string
	Entry(ctx context.Context, ev Event)
	Exit(ctx context.Context, ev Event)
}
```

State represents a state in the HSM.

### Functions returning State

#### StateWrapper

```go
func StateWrapper(s interface{}) State
```

StateWrapper creates a State from a simple struct.


## type Transition

```go
type Transition struct {
	On     string                                   `json:"on"`
	Source string                                   `json:"source"`
	Target string                                   `json:"target"`
	Effect func(ctx context.Context, ev Event)      `json:"-"`
	Guard  func(ctx context.Context, ev Event) bool `json:"-"`
}
```

Transition represents a state transition.

## type simpleState

```go
type simpleState struct {
	name string
}
```

simpleState implements State interface for simple structs.

### Methods

#### simpleState.Entry

```go
func () Entry(ctx context.Context, ev Event)
```

#### simpleState.Exit

```go
func () Exit(ctx context.Context, ev Event)
```

#### simpleState.Name

```go
func () Name() string
```



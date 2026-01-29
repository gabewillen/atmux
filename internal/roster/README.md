# package roster

`import "github.com/copilot-claude-sonnet-4/amux/internal/roster"`

Package roster implements the roster data model and listing outputs per spec §6.2.
The roster maintains all agents, host managers, and the director with real-time updates.

- `type Entry` — Entry represents a single entry in the roster per spec §6.2.
- `type PresenceChangeEvent` — PresenceChangeEvent is emitted when a participant's presence changes.
- `type Store` — Store manages the roster of all participants per spec §6.2.

## type Entry

```go
type Entry struct {
	// ID is the runtime ID of the participant.
	ID muid.MUID `json:"id"`

	// Type indicates the entry type: "agent", "manager", or "director".
	Type string `json:"type"`

	// Name is the human-readable name.
	Name string `json:"name"`

	// Slug is the normalized slug for addressing.
	Slug string `json:"slug"`

	// Presence is the current presence state.
	Presence api.PresenceState `json:"presence"`

	// State is the current lifecycle state (for agents).
	State api.AgentState `json:"state,omitempty"`

	// HostID is the host identifier (for agents and managers).
	HostID string `json:"host_id,omitempty"`

	// About is a description of the participant's role.
	About string `json:"about,omitempty"`

	// CurrentTask describes what the participant is currently working on (if busy).
	CurrentTask string `json:"current_task,omitempty"`

	// LastSeen is when this participant was last active.
	LastSeen time.Time `json:"last_seen"`

	// Metadata contains additional participant-specific data.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
```

Entry represents a single entry in the roster per spec §6.2.

### Functions returning Entry

#### copyEntry

```go
func copyEntry(src *Entry) *Entry
```

copyEntry creates a deep copy of a roster entry.


## type PresenceChangeEvent

```go
type PresenceChangeEvent struct {
	// ParticipantID is the runtime ID of the participant.
	ParticipantID muid.MUID `json:"participant_id"`

	// OldPresence is the previous presence state.
	OldPresence api.PresenceState `json:"old_presence"`

	// NewPresence is the new presence state.
	NewPresence api.PresenceState `json:"new_presence"`

	// Timestamp is when the change occurred.
	Timestamp time.Time `json:"timestamp"`
}
```

PresenceChangeEvent is emitted when a participant's presence changes.

## type Store

```go
type Store struct {
	// entries maps runtime ID to roster entry.
	entries map[muid.MUID]*Entry

	// slugIndex maps slug to runtime ID for lookup.
	slugIndex map[string]muid.MUID

	// mu protects concurrent access to the roster.
	mu sync.RWMutex

	// ctx is the store context.
	ctx context.Context

	// cancel cancels the store context.
	cancel context.CancelFunc

	// subscribers holds channels for presence change notifications.
	subscribers []chan<- PresenceChangeEvent

	// subMu protects the subscribers slice.
	subMu sync.RWMutex
}
```

Store manages the roster of all participants per spec §6.2.

### Functions returning Store

#### NewStore

```go
func NewStore() *Store
```

NewStore creates a new roster store.


### Methods

#### Store.AddAgent

```go
func () AddAgent(agent *api.Agent, hostID string) error
```

AddAgent adds or updates an agent in the roster per spec requirements.

#### Store.AddDirector

```go
func () AddDirector(id muid.MUID, presence api.PresenceState) error
```

AddDirector adds or updates the director in the roster.

#### Store.AddManager

```go
func () AddManager(id muid.MUID, hostID string, presence api.PresenceState) error
```

AddManager adds or updates a host manager in the roster.

#### Store.Close

```go
func () Close() error
```

Close gracefully shuts down the roster store.

#### Store.GetByID

```go
func () GetByID(id muid.MUID) (*Entry, error)
```

GetByID retrieves a roster entry by runtime ID.

#### Store.GetBySlug

```go
func () GetBySlug(slug string) (*Entry, error)
```

GetBySlug retrieves a roster entry by slug.

#### Store.ListAgents

```go
func () ListAgents() []*Entry
```

ListAgents returns all agent entries.

#### Store.ListAll

```go
func () ListAll() []*Entry
```

ListAll returns all roster entries per spec §6.2 requirements.

#### Store.ListByPresence

```go
func () ListByPresence(presence api.PresenceState) []*Entry
```

ListByPresence returns all entries with the specified presence.

#### Store.Remove

```go
func () Remove(id muid.MUID) error
```

Remove removes a participant from the roster.

#### Store.Subscribe

```go
func () Subscribe() <-chan PresenceChangeEvent
```

Subscribe adds a subscriber for presence change events per spec §6.3.

#### Store.UpdatePresence

```go
func () UpdatePresence(id muid.MUID, presence api.PresenceState) error
```

UpdatePresence updates the presence state of a participant.

#### Store.UpdateTask

```go
func () UpdateTask(id muid.MUID, task string) error
```

UpdateTask updates the current task for a participant.

#### Store.notifySubscribers

```go
func () notifySubscribers(event PresenceChangeEvent)
```

notifySubscribers sends presence change events to all subscribers.



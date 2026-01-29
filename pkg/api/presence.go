package api

// PresenceState is re-exported/defined here for public API usage.
type PresenceState string

const (
	PresenceOnline  PresenceState = "Online"
	PresenceBusy    PresenceState = "Busy"
	PresenceOffline PresenceState = "Offline"
	PresenceAway    PresenceState = "Away"
)

package config

// ConfigChange describes an individual configuration change.
type ConfigChange struct {
	Path     string
	OldValue any
	NewValue any
}

// Config event names.
const (
	ConfigFileChanged = "config.file_changed"
	ConfigLoaded      = "config.loaded"
	ConfigReloaded    = "config.reloaded"
	ConfigReloadFailed = "config.reload_failed"
	ConfigUpdated     = "config.updated"
)

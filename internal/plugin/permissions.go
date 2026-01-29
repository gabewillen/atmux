package plugin

import "strings"

// PermissionGater checks if a plugin has the required permissions.
type PermissionGater interface {
	HasPermission(plugin *Plugin, permission string) bool
}

// DefaultGater implements default permission checking.
type DefaultGater struct{}

// HasPermission checks if the plugin's manifest declares the permission.
// Supports basic globbing (e.g. "agent.*").
func (g *DefaultGater) HasPermission(p *Plugin, permission string) bool {
	if p == nil {
		return false
	}
	for _, allowed := range p.Manifest.Permissions {
		if allowed == permission {
			return true
		}
		if strings.HasSuffix(allowed, ".*") {
			prefix := strings.TrimSuffix(allowed, ".*")
			if strings.HasPrefix(permission, prefix+".") {
				return true
			}
		}
	}
	return false
}

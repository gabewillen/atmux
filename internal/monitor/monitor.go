// Package monitor implements PTY output monitoring (delegates to adapters)
package monitor

import "errors"

// ErrMonitorFailure is returned when monitoring fails
var ErrMonitorFailure = errors.New("monitor operation failed")
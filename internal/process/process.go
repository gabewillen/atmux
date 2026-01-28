// Package process implements process tracking and interception (generic)
package process

import "errors"

// ErrProcessTracking is returned when process tracking fails
var ErrProcessTracking = errors.New("process tracking failed")
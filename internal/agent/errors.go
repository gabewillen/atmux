package agent

import "errors"

// ErrDispatcherRequired is returned when a dispatcher is required but missing.
var ErrDispatcherRequired = errors.New("dispatcher required")

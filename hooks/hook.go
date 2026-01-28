package hooks

import "errors"

// ErrHooksUnavailable is returned when hooks are not built for the platform.
var ErrHooksUnavailable = errors.New("hooks unavailable")

// Init initializes the exec hook library.
func Init() error {
	return ErrHooksUnavailable
}

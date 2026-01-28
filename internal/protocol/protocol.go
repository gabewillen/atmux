// Package protocol implements remote communication protocol (transports events)
package protocol

import "errors"

// ErrProtocol is returned when protocol operations fail
var ErrProtocol = errors.New("protocol operation failed")
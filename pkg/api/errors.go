package api

import "errors"

var (
	// ErrInvalidLocationType is returned when parsing an invalid location type string.
	ErrInvalidLocationType = errors.New("invalid location type: must be 'local' or 'ssh'")

	// ErrReservedID is returned when attempting to use the reserved ID value 0.
	ErrReservedID = errors.New("cannot use reserved ID value 0")

	// ErrInvalidAgent is returned when an agent structure fails validation.
	ErrInvalidAgent = errors.New("invalid agent configuration")
)

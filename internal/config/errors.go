package config

import (
	"errors"
)

// ErrInvalidUsage indicates an error in CLI arguments or execution parameters.
var ErrInvalidUsage = errors.New("invalid usage")

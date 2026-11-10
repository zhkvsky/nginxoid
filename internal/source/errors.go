package source

import "errors"

// ErrNotFound indicates that a file, glob pattern, or remote URL was not found or is unavailable.
var ErrNotFound = errors.New("source not found")

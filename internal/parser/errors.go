package parser

import "errors"

// ErrMalformedLine indicates a malformed NGINX log line format.
var ErrMalformedLine = errors.New("malformed log line")

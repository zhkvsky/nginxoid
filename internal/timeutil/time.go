// Package timeutil provides shared constants and helper functions for working with time.
package timeutil

import "time"

const (
	// DateLayout defines the date format used for CLI arguments (--from, --to).
	DateLayout = "2006-01-02"

	// NginxTimeLayout defines the format for the $time_local field in NGINX logs.
	NginxTimeLayout = "02/Jan/2006:15:04:05 -0700"
)

// EndOfDay returns the end-of-day timestamp (23:59:59.999999999) for the given date.
func EndOfDay(t time.Time) time.Time {
	return t.Add(24*time.Hour - time.Nanosecond)
}

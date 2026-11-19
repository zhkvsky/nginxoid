// Package analyzer handles log filtering and aggregated statistics calculation.
package analyzer

import (
	"time"
)

// MatchesDateRange checks whether the request timestamp falls within the inclusive [from, to] range.
func MatchesDateRange(recordTime, from, to time.Time) bool {
	if !from.IsZero() && recordTime.Before(from) {
		return false
	}

	if !to.IsZero() && recordTime.After(to) {
		return false
	}

	return true
}

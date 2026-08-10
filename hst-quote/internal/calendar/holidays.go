// Package calendar decides whether a symbol is shut for a holiday.
//
// This is a twin of hstcore/internal/settings/calendar.go: the two services are separate Go
// modules, so the matcher cannot be shared and has to be kept in step by hand. Change one,
// change the other.
package calendar

import (
	"strings"
	"time"

	"hstquote/model"
)

// Closed reports whether trading is shut for this symbol at this moment.
//
// From and To are the work time, the window the server stays open, as the holiday dialog
// labels them. So a matched day is closed except inside its windows, and a record with no
// work time at all closes the day outright. Several records can match one date and their
// windows add up.
func Closed(days []model.Holiday, path, symbol string, at time.Time) bool {
	// #nosec G115 -- a minute of the day is 0..1439
	minutes := int32(at.Hour()*60 + at.Minute())

	matched := false

	for i := range days {
		d := &days[i]

		// #nosec G115 -- a calendar year, month and day all fit
		if d.Year != 0 && d.Year != int32(at.Year()) {
			continue
		}
		// #nosec G115 -- a month is 1..12 and a day 1..31
		if d.Month != int16(at.Month()) || d.Day != int16(at.Day()) {
			continue
		}
		if !covers(d.Symbols, path, symbol) {
			continue
		}

		matched = true

		// the prohibiting kind: it opens nothing, but another record still may
		if d.From == 0 && d.To == 0 {
			continue
		}
		// To is the last open minute, inclusive
		if minutes >= d.From && minutes <= d.To {
			return false
		}
	}

	return matched
}

func covers(masks []string, path, symbol string) bool {
	if len(masks) == 0 {
		return true
	}

	for _, target := range masks {
		if maskHits(strings.TrimSpace(target), path, symbol) {
			return true
		}
	}

	return false
}

func maskHits(target, path, symbol string) bool {
	switch {
	case target == "" || target == "*":
		return true
	case strings.HasSuffix(target, "*"):
		prefix := strings.TrimSuffix(target, "*")
		return strings.HasPrefix(path, prefix) || strings.HasPrefix(symbol, prefix)
	default:
		return target == symbol || target == path
	}
}

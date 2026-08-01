package session

import "time"

// Window is one quote session interval for a symbol weekday.
type Window struct {
	SymbolID int64
	Day      int16
	Open     int32
	Close    int32
}

// IsQuoteOpen reports whether the quote session is open for symbolID at now.
// Missing session rows mean the symbol is always open for quotes.
func IsQuoteOpen(symbolID int64, sessions []Window, now time.Time) bool {
	// #nosec G115 -- a weekday is 0..6 and a minute of the day 0..1439
	day := int16(now.Weekday())
	// #nosec G115 -- a minute of the day is 0..1439
	minute := int32(now.Hour()*60 + now.Minute())

	hasQuoteSessions := false
	for _, sess := range sessions {
		if sess.SymbolID != symbolID {
			continue
		}
		hasQuoteSessions = true
		if sess.Day != day {
			continue
		}
		if minute >= sess.Open && minute < sess.Close {
			return true
		}
	}
	// an instrument that names no quote session is quoted around the clock
	return !hasQuoteSessions
}

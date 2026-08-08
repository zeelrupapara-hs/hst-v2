package model

import (
	"fmt"
)

// Every subject this feed uses. The tick stream is named by the shared contract so the engine and
// the api cannot drift from it; the rest are the feed's own.
const (
	SubjectDatafeedCreated = "system.datafeeds.created"
	SubjectDatafeedUpdated = "system.datafeeds.updated"
	SubjectDatafeedDeleted = "system.datafeeds.deleted"
	SubjectDatafeedConfig  = "system.datafeeds.config.>"

	SubjectSnapshot = "hstquote.snapshot"

	GroupDatafeedConfig = "hstquote-datafeed"
)

// SubjectTick is one instrument's price stream.
func SubjectTick(symbol string) string { return fmt.Sprintf("hstquote.tick.%s", symbol) }

// SubjectJournal is one datafeed's operating journal: connects, logons, errors.
func SubjectJournal(datafeedId int64) string {
	return fmt.Sprintf("hstquote.journal.%d", datafeedId)
}

// SubjectStatus is one datafeed's runtime telemetry.
func SubjectStatus(datafeedId int64) string {
	return fmt.Sprintf("hstquote.status.%d", datafeedId)
}

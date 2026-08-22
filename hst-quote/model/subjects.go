package model

import (
	"fmt"
	"strings"
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

// tickSubjectEscaper keeps the subject one token: "." splits tokens and " " is invalid in NATS.
var tickSubjectEscaper = strings.NewReplacer(".", "_", " ", "_")

// SubjectTick is one instrument's price stream; the symbol is escaped, consumers read it from the payload.
func SubjectTick(symbol string) string {
	return "hstquote.tick." + tickSubjectEscaper.Replace(symbol)
}

// SubjectJournal is one datafeed's operating journal: connects, logons, errors.
func SubjectJournal(datafeedId int64) string {
	return fmt.Sprintf("hstquote.journal.%d", datafeedId)
}

// SubjectStatus is one datafeed's runtime telemetry.
func SubjectStatus(datafeedId int64) string {
	return fmt.Sprintf("hstquote.status.%d", datafeedId)
}

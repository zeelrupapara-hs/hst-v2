package model

import (
	"fmt"

	wire "hstmodel"
)

// Every subject this feed uses. The tick stream is named by the shared contract so the engine and
// the api cannot drift from it; the rest are the feed's own.
const (
	SubjectDatafeedCreated = "system.datafeeds.created"
	SubjectDatafeedUpdated = "system.datafeeds.updated"
	SubjectDatafeedDeleted = "system.datafeeds.deleted"
	SubjectDatafeedConfig  = "system.datafeeds.config.>"

	SubjectSnapshot = wire.SubjectQuoteSnapshot

	GroupDatafeedConfig = "hstquote-datafeed"
)

// SubjectTick is one instrument's price stream.
func SubjectTick(symbol string) string { return wire.SubjectQuoteTick(symbol) }

// SubjectStatus is one datafeed's runtime telemetry.
func SubjectStatus(datafeedId int64) string {
	return fmt.Sprintf("hstquote.status.%d", datafeedId)
}

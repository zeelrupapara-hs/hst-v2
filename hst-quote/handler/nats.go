package handler

const (
	SubjectTick = "hstquote.tick.%s"

	SubjectDatafeedCreated = "hstserver.datafeed.created"
	SubjectDatafeedUpdated = "hstserver.datafeed.updated"
	SubjectDatafeedDeleted = "hstserver.datafeed.deleted"
	SubjectDatafeedConfig  = "hstserver.datafeed.config.>"

	GroupDatafeedConfig = "hstquote-datafeed"
)

// DatafeedEvent is published by hst-server when feed configuration changes.
type DatafeedEvent struct {
	DatafeedID int64 `json:"datafeed_id"`
	Mode       int32 `json:"mode"`
	Enable     int16 `json:"enable"`
}

// HasQuoteFlag reports whether the feed is configured for quote ingestion.
func (e DatafeedEvent) HasQuoteFlag() bool {
	return (e.Mode&1) != 0 && e.Enable == 1
}

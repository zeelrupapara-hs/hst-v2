package handler

// Every subject this service publishes or consumes belongs here, as a
// constant. A subject typed inline in two files is a subject that will be
// renamed in one of them.
//
// Convention: <service>.<entity>.<event>
const (
	SubjectNewsItem = "hstnews.item.%d"

	SubjectDatafeedCreated = "hstserver.datafeed.created"
	SubjectDatafeedUpdated = "hstserver.datafeed.updated"
	SubjectDatafeedDeleted = "hstserver.datafeed.deleted"
	SubjectDatafeedConfig  = "hstserver.datafeed.config.>"

	GroupDatafeedConfig = "hstnews-datafeed"
)

// DatafeedEvent is published by hst-server when feed configuration changes.
type DatafeedEvent struct {
	DatafeedID int64 `json:"datafeed_id"`
	Mode       int32 `json:"mode"`
	Enable     int16 `json:"enable"`
}

// HasNewsFlag reports whether the feed is configured for news ingestion.
func (e DatafeedEvent) HasNewsFlag() bool {
	return (e.Mode & 2) != 0 && e.Enable == 1
}

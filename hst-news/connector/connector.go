package connector

import (
	"context"

	"hstnews/model"
)

// ConnectorType identifies the wire protocol implementation.
type ConnectorType string

const (
	TypeRSSPoll ConnectorType = "rss_poll"
)

// RawItem is a provider-specific news record before normalization.
type RawItem struct {
	ExternalID  string
	Subject     string
	BodyHTML    string
	Link        string
	PublishedAt string
	Category    string
	Language    string
	Priority    bool
	BytesRead   int64
}

// Connector fetches news from an external source.
type Connector interface {
	Type() ConnectorType
	Fetch(ctx context.Context, feed model.NewsFeed) ([]RawItem, int64, error)
}

// ForModule returns the connector for an MT5 module name.
func ForModule(module string) (Connector, bool) {
	switch module {
	case "RSSNewsFeeder", "rss", "RSS":
		return &RSS{}, true
	default:
		return nil, false
	}
}

package model

import "time"

// NewsItem is the normalized news record published by hst-news.
type NewsItem struct {
	ID          string            `json:"id"`
	DatafeedID  int64             `json:"datafeed_id"`
	SourceName  string            `json:"source_name"`
	PublishedAt time.Time         `json:"published_at"`
	ReceivedAt  time.Time         `json:"received_at"`
	Subject     string            `json:"subject"`
	BodyHTML    string            `json:"body_html"`
	Category    string            `json:"category"`
	Language    string            `json:"language"`
	Link        string            `json:"link"`
	Priority    bool              `json:"priority"`
	Meta        map[string]string `json:"meta,omitempty"`
}

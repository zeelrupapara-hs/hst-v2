package normalize

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"hstnews/connector"
	"hstnews/model"
)

const (
	paramNewsCategory     = "News Category"
	paramLanguage         = "Language"
	defaultNewsCategory   = "News"
	defaultRSSPollSeconds = 300
)

var pubDateLayouts = []string{
	time.RFC1123Z,
	time.RFC1123,
	time.RFC822Z,
	time.RFC822,
	time.RFC3339,
	"Mon, 2 Jan 2006 15:04:05 -0700",
	"Mon, 02 Jan 2006 15:04:05 MST",
	"2006-01-02T15:04:05Z07:00",
}

// PollInterval returns the poll period for a feed in seconds.
func PollInterval(feed model.NewsFeed) time.Duration {
	raw := strings.TrimSpace(feed.Param("News Request Period"))
	if raw == "" {
		return defaultRSSPollSeconds * time.Second
	}
	sec, err := time.ParseDuration(raw + "s")
	if err != nil || sec < time.Second {
		return defaultRSSPollSeconds * time.Second
	}
	return sec
}

// Items converts raw provider records into normalized news items.
func Items(feed model.NewsFeed, raw []connector.RawItem, receivedAt time.Time) []model.NewsItem {
	prefix := strings.TrimSpace(feed.Param(paramNewsCategory))
	if prefix == "" {
		prefix = defaultNewsCategory
	}

	lang := strings.TrimSpace(feed.Param(paramLanguage))
	if lang == "" {
		lang = "English"
	}

	out := make([]model.NewsItem, 0, len(raw))
	for _, item := range raw {
		if item.Subject == "" && item.BodyHTML == "" {
			continue
		}

		category := prefix
		if item.Category != "" {
			category = prefix + `\` + item.Category
		}

		pubAt := parsePubDate(item.PublishedAt, receivedAt)
		id := itemID(feed.Datafeed.DatafeedID, item.ExternalID, item.Subject, item.Link, item.PublishedAt)

		out = append(out, model.NewsItem{
			ID:          id,
			DatafeedID:  feed.Datafeed.DatafeedID,
			SourceName:  feed.Datafeed.Name,
			PublishedAt: pubAt,
			ReceivedAt:  receivedAt,
			Subject:     item.Subject,
			BodyHTML:    item.BodyHTML,
			Category:    category,
			Language:    lang,
			Link:        item.Link,
			Priority:    item.Priority,
		})
	}
	return out
}

func parsePubDate(raw string, fallback time.Time) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	for _, layout := range pubDateLayouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.UTC()
		}
	}
	return fallback
}

func itemID(datafeedID int64, parts ...string) string {
	h := sha256.New()
	h.Write([]byte(strings.Join(parts, "|")))
	h.Write([]byte("|"))
	h.Write([]byte(strconv.FormatInt(datafeedID, 10)))
	return hex.EncodeToString(h.Sum(nil))[:32]
}

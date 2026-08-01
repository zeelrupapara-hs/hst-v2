package normalize_test

import (
	"testing"
	"time"

	"hstnews/connector"
	"hstnews/internal/normalize"
	"hstnews/model"
)

func TestItemsCategoryPrefix(t *testing.T) {
	feed := model.NewsFeed{
		Datafeed: model.Datafeed{DatafeedID: 1, Name: "forexlive"},
		Params: map[string]string{
			"News Category": "Forex",
			"Language":      "English",
		},
	}

	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	items := normalize.Items(feed, []connector.RawItem{{
		ExternalID:  "abc",
		Subject:     "EUR/USD rises",
		BodyHTML:    "<p>content</p>",
		Link:        "https://example.com/1",
		PublishedAt: now.Format(time.RFC1123),
		Category:    "Central Bank",
	}}, now)

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Category != `Forex\Central Bank` {
		t.Fatalf("unexpected category: %q", items[0].Category)
	}
	if items[0].Language != "English" {
		t.Fatalf("unexpected language: %q", items[0].Language)
	}
}

func TestPollIntervalDefault(t *testing.T) {
	feed := model.NewsFeed{Params: map[string]string{}}
	if got := normalize.PollInterval(feed); got != 300*time.Second {
		t.Fatalf("expected 300s default, got %s", got)
	}
}

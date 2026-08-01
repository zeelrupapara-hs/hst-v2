package connector_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"hstnews/connector"
	"hstnews/model"
)

const sampleRSS = `<?xml version="1.0"?>
<rss version="2.0">
  <channel>
    <item>
      <title>Test headline</title>
      <description><![CDATA[<p>Body</p>]]></description>
      <link>https://example.com/news/1</link>
      <pubDate>Thu, 30 Jul 2026 12:00:00 GMT</pubDate>
      <category>Macro</category>
      <guid>news-1</guid>
    </item>
  </channel>
</rss>`

func TestRSSFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(sampleRSS))
	}))
	defer srv.Close()

	rss := &connector.RSS{Client: srv.Client()}
	items, n, err := rss.Fetch(context.Background(), model.NewsFeed{
		Datafeed: model.Datafeed{DatafeedID: 7, FeedServer: srv.URL},
	})
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}
	if n <= 0 {
		t.Fatalf("expected bytes read")
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Subject != "Test headline" {
		t.Fatalf("unexpected subject: %q", items[0].Subject)
	}
}

func TestForModule(t *testing.T) {
	if _, ok := connector.ForModule("RSSNewsFeeder"); !ok {
		t.Fatal("expected RSSNewsFeeder support")
	}
	if _, ok := connector.ForModule("UnknownFeeder"); ok {
		t.Fatal("did not expect UnknownFeeder support")
	}
}

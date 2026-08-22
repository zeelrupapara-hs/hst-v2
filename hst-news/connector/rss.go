package connector

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"hstnews/model"
)

const (
	defaultHTTPTimeout = 30 * time.Second
	// maxFeedBytes caps the body read from an admin-configured URL.
	maxFeedBytes = 16 << 20
)

// RSS polls RSS 2.0 XML feeds over HTTP(S); Atom is rejected with a clear error.
type RSS struct {
	Client *http.Client
}

func (r *RSS) Type() ConnectorType { return TypeRSSPoll }

type rssFeed struct {
	XMLName xml.Name
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
	Category    string `xml:"category"`
	GUID        string `xml:"guid"`
}

func (r *RSS) Fetch(ctx context.Context, feed model.NewsFeed) ([]RawItem, int64, error) {
	url := strings.TrimSpace(feed.Datafeed.FeedServer)
	if url == "" {
		return nil, 0, fmt.Errorf("feed %d: empty feed_server", feed.Datafeed.DatafeedID)
	}

	client := r.Client
	if client == nil {
		client = &http.Client{Timeout: defaultHTTPTimeout}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}

	if pwd := strings.TrimSpace(feed.Datafeed.FeedPassword); pwd != "" {
		user := feed.Param("Feed login")
		if user == "" && feed.Datafeed.FeedLogin != "" {
			user = feed.Datafeed.FeedLogin
		}
		if user != "" {
			req.SetBasicAuth(user, pwd)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, 0, fmt.Errorf("feed %d: http %d from %s", feed.Datafeed.DatafeedID, resp.StatusCode, url)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxFeedBytes+1))
	if err != nil {
		return nil, 0, err
	}
	if len(body) > maxFeedBytes {
		return nil, int64(len(body)), fmt.Errorf("feed %d: body over %d bytes", feed.Datafeed.DatafeedID, maxFeedBytes)
	}

	var parsed rssFeed
	if err := xml.Unmarshal(body, &parsed); err != nil {
		return nil, int64(len(body)), fmt.Errorf("feed %d: parse rss: %w", feed.Datafeed.DatafeedID, err)
	}
	if parsed.XMLName.Local != "rss" {
		return nil, int64(len(body)), fmt.Errorf("feed %d: not an RSS document (<%s>), Atom is not supported", feed.Datafeed.DatafeedID, parsed.XMLName.Local)
	}

	items := make([]RawItem, 0, len(parsed.Channel.Items))
	for _, item := range parsed.Channel.Items {
		extID := strings.TrimSpace(item.GUID)
		if extID == "" {
			extID = strings.TrimSpace(item.Link)
		}
		if extID == "" {
			extID = strings.TrimSpace(item.Title)
		}

		items = append(items, RawItem{
			ExternalID:  extID,
			Subject:     strings.TrimSpace(item.Title),
			BodyHTML:    strings.TrimSpace(item.Description),
			Link:        strings.TrimSpace(item.Link),
			PublishedAt: strings.TrimSpace(item.PubDate),
			Category:    strings.TrimSpace(item.Category),
			BytesRead:   int64(len(body)),
		})
	}

	return items, int64(len(body)), nil
}

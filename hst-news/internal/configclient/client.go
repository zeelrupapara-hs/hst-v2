package configclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"errors"

	"hstnews/model"
)

var errNotFound = errors.New("datafeed not found")

const newsMode = int32(model.FeederFlags_news)

// Snapshot matches hst-server WorkerDatafeedConfig JSON.
type Snapshot struct {
	DatafeedID       int64                `json:"datafeed_id"`
	Name             string               `json:"name"`
	Module           string               `json:"module"`
	Enable           model.DatafeedEnable `json:"enable"`
	Mode             model.FeederFlags    `json:"mode"`
	FeedServer       string               `json:"feed_server"`
	FeedLogin        int64                `json:"feed_login"`
	FeedPassword     string               `json:"feed_password"`
	TimeoutReconnect int32                `json:"timeout_reconnect"`
	Params           []Param              `json:"params"`
}

type Param struct {
	ParamKey string `json:"param_key"`
	Value    string `json:"value"`
}

// Client loads datafeed config from hst-server internal API.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func New(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		token:   token,
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) ListNewsFeeds(ctx context.Context) ([]model.NewsFeed, error) {
	u, err := url.Parse(c.baseURL + "/internal/v1/datafeeds")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("mode", strconv.Itoa(int(newsMode)))
	u.RawQuery = q.Encode()

	var snaps []Snapshot
	if err := c.getJSON(ctx, u.String(), &snaps); err != nil {
		return nil, err
	}
	out := make([]model.NewsFeed, 0, len(snaps))
	for _, s := range snaps {
		out = append(out, toNewsFeed(s))
	}
	return out, nil
}

func (c *Client) GetNewsFeed(ctx context.Context, datafeedID int64) (*model.NewsFeed, error) {
	var snap Snapshot
	if err := c.getJSON(ctx, fmt.Sprintf("%s/internal/v1/datafeeds/%d", c.baseURL, datafeedID), &snap); err != nil {
		if err == errNotFound {
			return nil, nil
		}
		return nil, err
	}
	feed := toNewsFeed(snap)
	return &feed, nil
}

func (c *Client) getJSON(ctx context.Context, url string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Service-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("config request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNotFound {
		return errNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("config api status %d: %s", resp.StatusCode, string(body))
	}

	var envelope struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("decode envelope: %w", err)
	}
	if !envelope.Success {
		return fmt.Errorf("config api error response")
	}
	if err := json.Unmarshal(envelope.Data, dest); err != nil {
		return fmt.Errorf("decode config data: %w", err)
	}
	return nil
}

func toNewsFeed(s Snapshot) model.NewsFeed {
	params := make(map[string]string, len(s.Params))
	for _, p := range s.Params {
		params[p.ParamKey] = p.Value
	}
	return model.NewsFeed{
		Datafeed: model.Datafeed{
			DatafeedID:   s.DatafeedID,
			Name:         s.Name,
			Module:       s.Module,
			Enable:       s.Enable,
			Mode:         s.Mode,
			FeedServer:   s.FeedServer,
			FeedLogin:    s.FeedLogin,
			FeedPassword: s.FeedPassword,
		},
		Params: params,
	}
}

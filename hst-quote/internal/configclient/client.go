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

	"hstquote/model"
)

var errNotFound = errors.New("datafeed not found")

const quoteMode = int32(model.FeederFlags_quotes)

// Snapshot matches hst-server WorkerDatafeedConfig JSON.
type Snapshot struct {
	DatafeedID       int64                `json:"datafeed_id"`
	Name             string               `json:"name"`
	Module           string               `json:"module"`
	Enable           model.DatafeedEnable `json:"enable"`
	Mode             model.FeederFlags    `json:"mode"`
	FeedServer       string               `json:"feed_server"`
	FeedLogin        string               `json:"feed_login"`
	FeedPassword     string               `json:"feed_password"`
	TimeoutReconnect int32                `json:"timeout_reconnect"`
	Params           []Param              `json:"params"`
	Translates       []Translate          `json:"translates"`
	Sessions         []SnapshotSession    `json:"sessions"`
	Settings         []SymbolSettings     `json:"settings"`
}

type SymbolSettings struct {
	SymbolID        int64   `json:"symbol_id"`
	Digits          int16   `json:"digits"`
	Point           float64 `json:"point"`
	TickFlags       int32   `json:"tick_flags"`
	TickBookDepth   int32   `json:"tick_book_depth"`
	CalcMode        int16   `json:"calc_mode"`
	TickChartMode   int16   `json:"tick_chart_mode"`
	SpliceType      int16   `json:"splice_type"`
	FilterSoft      int32   `json:"filter_soft"`
	FilterSoftTicks int32   `json:"filter_soft_ticks"`
	FilterHard      int32   `json:"filter_hard"`
	FilterHardTicks int32   `json:"filter_hard_ticks"`
	FilterDiscard   int32   `json:"filter_discard"`
	FilterSpreadMin int32   `json:"filter_spread_min"`
	FilterSpreadMax int32   `json:"filter_spread_max"`
	FilterGap       int32   `json:"filter_gap"`
	FilterGapTicks  int32   `json:"filter_gap_ticks"`
	Spread          int32   `json:"spread"`
	SpreadBalance   int32   `json:"spread_balance"`
}

type SnapshotSession struct {
	SymbolID int64 `json:"symbol_id"`
	Day      int16 `json:"day"`
	Open     int32 `json:"open"`
	Close    int32 `json:"close"`
}

type Param struct {
	ParamKey string `json:"param_key"`
	Value    string `json:"value"`
}

type Translate struct {
	TranslateID int64  `json:"translate_id"`
	SymbolID    int64  `json:"symbol_id"`
	Symbol      string `json:"symbol"`
	Source      string `json:"source"`
	BidMarkup   int32  `json:"bid_markup"`
	AskMarkup   int32  `json:"ask_markup"`
	Digits      int16  `json:"digits"`
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

func (c *Client) ListQuoteFeeds(ctx context.Context) ([]model.QuoteFeed, error) {
	u, err := url.Parse(c.baseURL + "/internal/v1/datafeeds")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("mode", strconv.Itoa(int(quoteMode)))
	u.RawQuery = q.Encode()

	var snaps []Snapshot
	if err := c.getJSON(ctx, u.String(), &snaps); err != nil {
		return nil, err
	}
	out := make([]model.QuoteFeed, 0, len(snaps))
	for _, s := range snaps {
		out = append(out, toQuoteFeed(s))
	}
	return out, nil
}

func (c *Client) GetQuoteFeed(ctx context.Context, datafeedID int64) (*model.QuoteFeed, error) {
	var snap Snapshot
	if err := c.getJSON(ctx, fmt.Sprintf("%s/internal/v1/datafeeds/%d", c.baseURL, datafeedID), &snap); err != nil {
		if err == errNotFound {
			return nil, nil
		}
		return nil, err
	}
	feed := toQuoteFeed(snap)
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

func toQuoteFeed(s Snapshot) model.QuoteFeed {
	params := make(map[string]string, len(s.Params))
	for _, p := range s.Params {
		params[p.ParamKey] = p.Value
	}
	translates := make([]model.DatafeedTranslate, 0, len(s.Translates))
	for _, t := range s.Translates {
		translates = append(translates, model.DatafeedTranslate{
			TranslateID: t.TranslateID,
			DatafeedID:  s.DatafeedID,
			SymbolID:    t.SymbolID,
			Symbol:      t.Symbol,
			Source:      t.Source,
			BidMarkup:   t.BidMarkup,
			AskMarkup:   t.AskMarkup,
			Digits:      t.Digits,
		})
	}
	sessions := make([]model.SymbolSession, 0, len(s.Sessions))
	for _, sess := range s.Sessions {
		sessions = append(sessions, model.SymbolSession{
			SymbolID: sess.SymbolID,
			Day:      sess.Day,
			Open:     sess.Open,
			Close:    sess.Close,
		})
	}
	settings := make(map[int64]model.SymbolSettings, len(s.Settings))
	for _, st := range s.Settings {
		settings[st.SymbolID] = model.SymbolSettings{
			SymbolID:        st.SymbolID,
			Digits:          st.Digits,
			Point:           st.Point,
			TickFlags:       st.TickFlags,
			TickBookDepth:   st.TickBookDepth,
			CalcMode:        st.CalcMode,
			TickChartMode:   st.TickChartMode,
			SpliceType:      st.SpliceType,
			FilterSoft:      st.FilterSoft,
			FilterSoftTicks: st.FilterSoftTicks,
			FilterHard:      st.FilterHard,
			FilterHardTicks: st.FilterHardTicks,
			FilterDiscard:   st.FilterDiscard,
			FilterSpreadMin: st.FilterSpreadMin,
			FilterSpreadMax: st.FilterSpreadMax,
			FilterGap:       st.FilterGap,
			FilterGapTicks:  st.FilterGapTicks,
			Spread:          st.Spread,
			SpreadBalance:   st.SpreadBalance,
		}
	}
	return model.QuoteFeed{
		Datafeed: model.Datafeed{
			DatafeedID:       s.DatafeedID,
			Name:             s.Name,
			Module:           s.Module,
			Enable:           s.Enable,
			Mode:             s.Mode,
			FeedServer:       s.FeedServer,
			FeedLogin:        s.FeedLogin,
			FeedPassword:     s.FeedPassword,
			TimeoutReconnect: s.TimeoutReconnect,
		},
		Params:     params,
		Translates: translates,
		Sessions:   sessions,
		Settings:   settings,
	}
}

// FromSnapshot converts a NATS config payload into a quote feed model.
func FromSnapshot(s Snapshot) model.QuoteFeed {
	return toQuoteFeed(s)
}

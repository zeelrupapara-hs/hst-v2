package repository

import (
	"context"
	"fmt"

	"hstquote/model"
	"hstquote/pkg/db"

	"github.com/jackc/pgx/v5"
)

// DatafeedRepo reads quote feed configuration from postgres.
type DatafeedRepo struct {
	DB *db.PostgresDB
}

func NewDatafeedRepo(database *db.PostgresDB) *DatafeedRepo {
	return &DatafeedRepo{DB: database}
}

const listQuoteFeedsSQL = `
SELECT datafeed_id, name, module, enable, mode, feed_server, feed_login,
       feed_password, timeout_reconnect, sys_connection, sys_last_time,
       ticks_count, bytes_received
FROM hst.datafeeds
WHERE enable = 1 AND (mode & 1) = 1
ORDER BY datafeed_id`

const listParamsSQL = `
SELECT param_id, datafeed_id, param_key, value
FROM hst.datafeed_params
WHERE datafeed_id = ANY($1)
ORDER BY datafeed_id, param_id`

const listTranslatesSQL = `
SELECT t.translate_id, t.datafeed_id, COALESCE(s.symbol_id, 0), t.symbol, t.source,
       t.bid_markup, t.ask_markup, t.digits
FROM hst.datafeed_translates t
LEFT JOIN hst.symbols s ON s.symbol = t.symbol
WHERE t.datafeed_id = ANY($1)
ORDER BY t.datafeed_id, t.translate_id`

const getQuoteFeedSQL = `
SELECT datafeed_id, name, module, enable, mode, feed_server, feed_login,
       feed_password, timeout_reconnect, sys_connection, sys_last_time,
       ticks_count, bytes_received
FROM hst.datafeeds
WHERE datafeed_id = $1`

const touchConnectedSQL = `
UPDATE hst.datafeeds
SET sys_connection = 1, sys_last_time = $2
WHERE datafeed_id = $1`

const touchDisconnectedSQL = `
UPDATE hst.datafeeds SET sys_connection = 0 WHERE datafeed_id = $1`

const recordTickSQL = `
UPDATE hst.datafeeds
SET sys_connection = 1,
    sys_last_time = $2,
    ticks_count = ticks_count + $3,
    bytes_received = bytes_received + $4
WHERE datafeed_id = $1`

// ListQuoteFeeds returns enabled feeds with the quotes flag set.
func (r *DatafeedRepo) ListQuoteFeeds(ctx context.Context) ([]model.QuoteFeed, error) {
	rows, err := r.DB.DB.Query(ctx, listQuoteFeedsSQL)
	if err != nil {
		return nil, fmt.Errorf("list quote feeds: %w", err)
	}
	defer rows.Close()

	feeds, ids, err := scanFeeds(rows)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}

	params, err := r.loadParams(ctx, ids)
	if err != nil {
		return nil, err
	}
	translates, err := r.loadTranslates(ctx, ids)
	if err != nil {
		return nil, err
	}

	for i := range feeds {
		id := feeds[i].Datafeed.DatafeedID
		feeds[i].Params = params[id]
		feeds[i].Translates = translates[id]
	}
	return feeds, nil
}

// GetQuoteFeed loads one feed with params and translates.
func (r *DatafeedRepo) GetQuoteFeed(ctx context.Context, datafeedID int64) (*model.QuoteFeed, error) {
	row := r.DB.DB.QueryRow(ctx, getQuoteFeedSQL, datafeedID)

	var feed model.QuoteFeed
	if err := scanFeed(row, &feed.Datafeed); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get quote feed %d: %w", datafeedID, err)
	}

	params, err := r.loadParams(ctx, []int64{datafeedID})
	if err != nil {
		return nil, err
	}
	translates, err := r.loadTranslates(ctx, []int64{datafeedID})
	if err != nil {
		return nil, err
	}
	feed.Params = params[datafeedID]
	feed.Translates = translates[datafeedID]
	return &feed, nil
}

func (r *DatafeedRepo) loadParams(ctx context.Context, ids []int64) (map[int64]map[string]string, error) {
	rows, err := r.DB.DB.Query(ctx, listParamsSQL, ids)
	if err != nil {
		return nil, fmt.Errorf("list feed params: %w", err)
	}
	defer rows.Close()

	out := make(map[int64]map[string]string, len(ids))
	for rows.Next() {
		var p model.DatafeedParam
		if err := rows.Scan(&p.ParamID, &p.DatafeedID, &p.ParamKey, &p.Value); err != nil {
			return nil, fmt.Errorf("scan feed param: %w", err)
		}
		if out[p.DatafeedID] == nil {
			out[p.DatafeedID] = make(map[string]string)
		}
		out[p.DatafeedID][p.ParamKey] = p.Value
	}
	return out, rows.Err()
}

func (r *DatafeedRepo) loadTranslates(ctx context.Context, ids []int64) (map[int64][]model.DatafeedTranslate, error) {
	rows, err := r.DB.DB.Query(ctx, listTranslatesSQL, ids)
	if err != nil {
		return nil, fmt.Errorf("list feed translates: %w", err)
	}
	defer rows.Close()

	out := make(map[int64][]model.DatafeedTranslate, len(ids))
	for rows.Next() {
		var t model.DatafeedTranslate
		if err := rows.Scan(&t.TranslateID, &t.DatafeedID, &t.SymbolID, &t.Symbol, &t.Source,
			&t.BidMarkup, &t.AskMarkup, &t.Digits); err != nil {
			return nil, fmt.Errorf("scan feed translate: %w", err)
		}
		out[t.DatafeedID] = append(out[t.DatafeedID], t)
	}
	return out, rows.Err()
}

func scanFeeds(rows pgx.Rows) ([]model.QuoteFeed, []int64, error) {
	var feeds []model.QuoteFeed
	var ids []int64

	for rows.Next() {
		var feed model.QuoteFeed
		if err := scanFeed(rows, &feed.Datafeed); err != nil {
			return nil, nil, fmt.Errorf("scan quote feed: %w", err)
		}
		feeds = append(feeds, feed)
		ids = append(ids, feed.Datafeed.DatafeedID)
	}
	return feeds, ids, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanFeed(row rowScanner, df *model.Datafeed) error {
	return row.Scan(
		&df.DatafeedID, &df.Name, &df.Module, &df.Enable, &df.Mode,
		&df.FeedServer, &df.FeedLogin, &df.FeedPassword, &df.TimeoutReconnect,
		&df.SysConnection, &df.SysLastTime, &df.TicksCount, &df.BytesReceived,
	)
}

func (r *DatafeedRepo) SetConnected(ctx context.Context, datafeedID int64, at int64) error {
	_, err := r.DB.DB.Exec(ctx, touchConnectedSQL, datafeedID, at)
	return err
}

func (r *DatafeedRepo) SetDisconnected(ctx context.Context, datafeedID int64) error {
	_, err := r.DB.DB.Exec(ctx, touchDisconnectedSQL, datafeedID)
	return err
}

func (r *DatafeedRepo) RecordTick(ctx context.Context, datafeedID int64, at int64, count int64, bytes int64) error {
	_, err := r.DB.DB.Exec(ctx, recordTickSQL, datafeedID, at, count, bytes)
	return err
}

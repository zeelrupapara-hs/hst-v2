package repository

import (
	"context"
	"fmt"

	"hstnews/model"
	"hstnews/pkg/db"

	"github.com/jackc/pgx/v5"
)

// DatafeedRepo reads news feed configuration from postgres.
type DatafeedRepo struct {
	DB *db.PostgresDB
}

func NewDatafeedRepo(database *db.PostgresDB) *DatafeedRepo {
	return &DatafeedRepo{DB: database}
}

const listNewsFeedsSQL = `
SELECT datafeed_id, name, module, enable, mode, feed_server, feed_login,
       feed_password, sys_connection, sys_last_time, news_count, bytes_received
FROM hst.datafeeds
WHERE enable = 1 AND (mode & 2) = 2
ORDER BY datafeed_id`

const listParamsSQL = `
SELECT param_id, datafeed_id, param_key, value
FROM hst.datafeed_params
WHERE datafeed_id = ANY($1)
ORDER BY datafeed_id, param_id`

const getNewsFeedSQL = `
SELECT datafeed_id, name, module, enable, mode, feed_server, feed_login,
       feed_password, sys_connection, sys_last_time, news_count, bytes_received
FROM hst.datafeeds
WHERE datafeed_id = $1`

const touchConnectedSQL = `
UPDATE hst.datafeeds
SET sys_connection = 1, sys_last_time = $2
WHERE datafeed_id = $1`

const touchDisconnectedSQL = `
UPDATE hst.datafeeds
SET sys_connection = 0
WHERE datafeed_id = $1`

const recordNewsSQL = `
UPDATE hst.datafeeds
SET sys_connection = 1,
    sys_last_time = $2,
    news_count = news_count + $3,
    bytes_received = bytes_received + $4
WHERE datafeed_id = $1`

// ListNewsFeeds returns enabled feeds with the news flag set.
func (r *DatafeedRepo) ListNewsFeeds(ctx context.Context) ([]model.NewsFeed, error) {
	rows, err := r.DB.DB.Query(ctx, listNewsFeedsSQL)
	if err != nil {
		return nil, fmt.Errorf("list news feeds: %w", err)
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

	for i := range feeds {
		feeds[i].Params = params[feeds[i].Datafeed.DatafeedID]
	}
	return feeds, nil
}

// GetNewsFeed loads one feed and its params.
func (r *DatafeedRepo) GetNewsFeed(ctx context.Context, datafeedID int64) (*model.NewsFeed, error) {
	row := r.DB.DB.QueryRow(ctx, getNewsFeedSQL, datafeedID)

	var feed model.NewsFeed
	if err := scanFeed(row, &feed.Datafeed); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get news feed %d: %w", datafeedID, err)
	}

	params, err := r.loadParams(ctx, []int64{datafeedID})
	if err != nil {
		return nil, err
	}
	feed.Params = params[datafeedID]
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
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate feed params: %w", err)
	}
	return out, nil
}

func scanFeeds(rows pgx.Rows) ([]model.NewsFeed, []int64, error) {
	var feeds []model.NewsFeed
	var ids []int64

	for rows.Next() {
		var feed model.NewsFeed
		if err := scanFeed(rows, &feed.Datafeed); err != nil {
			return nil, nil, fmt.Errorf("scan news feed: %w", err)
		}
		feeds = append(feeds, feed)
		ids = append(ids, feed.Datafeed.DatafeedID)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterate news feeds: %w", err)
	}
	return feeds, ids, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanFeed(row rowScanner, df *model.Datafeed) error {
	return row.Scan(
		&df.DatafeedID, &df.Name, &df.Module, &df.Enable, &df.Mode,
		&df.FeedServer, &df.FeedLogin, &df.FeedPassword,
		&df.SysConnection, &df.SysLastTime, &df.NewsCount, &df.BytesReceived,
	)
}

// SetConnected marks a feed as connected at the given unix nanoseconds.
func (r *DatafeedRepo) SetConnected(ctx context.Context, datafeedID int64, at int64) error {
	_, err := r.DB.DB.Exec(ctx, touchConnectedSQL, datafeedID, at)
	return err
}

// SetDisconnected marks a feed as disconnected.
func (r *DatafeedRepo) SetDisconnected(ctx context.Context, datafeedID int64) error {
	_, err := r.DB.DB.Exec(ctx, touchDisconnectedSQL, datafeedID)
	return err
}

// RecordNews increments counters after items were ingested.
func (r *DatafeedRepo) RecordNews(ctx context.Context, datafeedID int64, at int64, count int64, bytes int64) error {
	_, err := r.DB.DB.Exec(ctx, recordNewsSQL, datafeedID, at, count, bytes)
	return err
}

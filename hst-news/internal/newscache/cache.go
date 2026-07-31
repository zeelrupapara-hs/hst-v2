package newscache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"hstnews/model"
	"hstnews/pkg/redis"
)

const (
	keySeenPrefix = "hstnews:seen:"
	keyFeedPrefix = "hstnews:feed:"
	seenTTL       = 7 * 24 * time.Hour
	feedTTL       = 24 * time.Hour
	maxFeedItems  = 100
)

// Store handles deduplication and hot caching in redis.
type Store struct {
	Redis *redis.Redis
}

func New(rds *redis.Redis) *Store {
	return &Store{Redis: rds}
}

func seenKey(datafeedID int64, itemID string) string {
	return fmt.Sprintf("%s%d:%s", keySeenPrefix, datafeedID, itemID)
}

func feedKey(datafeedID int64) string {
	return fmt.Sprintf("%s%d", keyFeedPrefix, datafeedID)
}

// FilterNew drops items already seen for this feed.
func (s *Store) FilterNew(ctx context.Context, items []model.NewsItem) ([]model.NewsItem, error) {
	if len(items) == 0 {
		return nil, nil
	}

	out := make([]model.NewsItem, 0, len(items))
	for _, item := range items {
		key := seenKey(item.DatafeedID, item.ID)
		n, err := s.Redis.Client.Exists(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		if n == 0 {
			out = append(out, item)
		}
	}
	return out, nil
}

// MarkSeen records item ids so duplicates are skipped on later polls.
func (s *Store) MarkSeen(ctx context.Context, items []model.NewsItem) error {
	for _, item := range items {
		key := seenKey(item.DatafeedID, item.ID)
		if err := s.Redis.Client.Set(ctx, key, "1", seenTTL).Err(); err != nil {
			return err
		}
	}
	return nil
}

// PrependFeed stores the latest items at the front of the feed cache list.
func (s *Store) PrependFeed(ctx context.Context, items []model.NewsItem) error {
	if len(items) == 0 {
		return nil
	}

	byFeed := make(map[int64][]model.NewsItem)
	for _, item := range items {
		byFeed[item.DatafeedID] = append(byFeed[item.DatafeedID], item)
	}

	for datafeedID, batch := range byFeed {
		key := feedKey(datafeedID)

		var existing []model.NewsItem
		raw, err := s.Redis.Client.Get(ctx, key).Bytes()
		if err == nil {
			_ = json.Unmarshal(raw, &existing)
		}

		merged := append(batch, existing...)
		if len(merged) > maxFeedItems {
			merged = merged[:maxFeedItems]
		}

		encoded, err := json.Marshal(merged)
		if err != nil {
			return err
		}
		if err := s.Redis.Client.Set(ctx, key, encoded, feedTTL).Err(); err != nil {
			return err
		}
	}
	return nil
}

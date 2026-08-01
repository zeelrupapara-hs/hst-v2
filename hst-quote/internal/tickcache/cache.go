package tickcache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"hstquote/model"
	"hstquote/pkg/redis"
)

const (
	keyTickPrefix = "hstquote:tick:"
	// keyLastPrefix is the last price of a symbol, under the name the platform knows it by.
	keyLastPrefix = "hstquote:last:"
	tickTTL       = 5 * time.Minute
)

// Store caches the latest tick per symbol in redis.
type Store struct {
	Redis *redis.Redis
}

func New(rds *redis.Redis) *Store {
	return &Store{Redis: rds}
}

func tickKey(datafeedID int64, symbol string) string {
	return fmt.Sprintf("%s%d:%s", keyTickPrefix, datafeedID, symbol)
}

// LastKey is where a symbol's last price is kept, for whoever needs it after a restart.
func LastKey(symbol string) string {
	return keyLastPrefix + symbol
}

// Put stores the latest tick for a symbol.
//
// Twice: once under the feed that produced it, which expires because it is only a view of what a
// feed is doing, and once under the symbol alone with no expiry, because that one is the price the
// engine reads back when it starts and no feed is running to tell it.
func (s *Store) Put(ctx context.Context, tick model.Tick) error {
	payload, err := json.Marshal(tick)
	if err != nil {
		return err
	}

	if err := s.Redis.Client.Set(ctx, tickKey(tick.DatafeedID, tick.Symbol), payload, tickTTL).Err(); err != nil {
		return err
	}

	return s.Redis.Client.Set(ctx, LastKey(tick.Symbol), payload, 0).Err()
}

// Last reads back the price of every named symbol, skipping the ones nothing has quoted yet.
func (s *Store) Last(ctx context.Context, symbols []string) ([]model.Tick, error) {
	if len(symbols) == 0 {
		return nil, nil
	}

	keys := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		keys = append(keys, LastKey(symbol))
	}

	values, err := s.Redis.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	out := make([]model.Tick, 0, len(values))

	for _, v := range values {
		text, ok := v.(string)
		if !ok {
			continue
		}

		var tick model.Tick
		if err := json.Unmarshal([]byte(text), &tick); err != nil {
			continue
		}

		out = append(out, tick)
	}

	return out, nil
}

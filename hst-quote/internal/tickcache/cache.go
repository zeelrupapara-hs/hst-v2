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

// Put stores the latest tick for a symbol.
func (s *Store) Put(ctx context.Context, tick model.Tick) error {
	payload, err := json.Marshal(tick)
	if err != nil {
		return err
	}
	return s.Redis.Client.Set(ctx, tickKey(tick.DatafeedID, tick.Symbol), payload, tickTTL).Err()
}

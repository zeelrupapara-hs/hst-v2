package arbiter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"hstquote/pkg/logger"
	"hstquote/pkg/redis"

	goredis "github.com/redis/go-redis/v9"
)

// Selector picks one active quote source per symbol across every hst-quote
// instance: a symbol's prices are accepted from a single feed at a time, the
// feed list order (feed_index) is the priority, a higher-priority feed takes
// over on its first tick, and a silent active feed loses the symbol to a
// lower-priority one after the datafeeds timeout.
type Selector struct {
	rds     *redis.Redis
	log     *logger.Logger
	timeout time.Duration
	script  *goredis.Script

	mu sync.Mutex
	// active is each local feed's view of the symbols it currently owns, kept
	// only to journal activations on transitions instead of on every claim.
	active map[int64]map[int64]struct{}
	// errLogAt throttles redis failure logging on the tick hot path.
	errLogAt time.Time
}

const keySourcePrefix = "hstquote:src:"

// claimScript arbitrates one tick: refresh if already the owner, claim if the
// symbol is free, take over if strictly higher priority, otherwise ignore.
// Returns 1 = still owner, 2 = became owner, 0 = ignored.
var claimScript = goredis.NewScript(`
local cur = redis.call('GET', KEYS[1])
if cur == ARGV[1] then
  redis.call('PEXPIRE', KEYS[1], ARGV[2])
  return 1
end
if not cur then
  redis.call('SET', KEYS[1], ARGV[1], 'PX', ARGV[2])
  return 2
end
local curidx, curid = string.match(cur, '^(-?%d+):(%d+)$')
local newidx, newid = string.match(ARGV[1], '^(-?%d+):(%d+)$')
if curidx == nil then
  redis.call('SET', KEYS[1], ARGV[1], 'PX', ARGV[2])
  return 2
end
curidx = tonumber(curidx); newidx = tonumber(newidx)
curid = tonumber(curid); newid = tonumber(newid)
if newidx < curidx or (newidx == curidx and newid < curid) then
  redis.call('SET', KEYS[1], ARGV[1], 'PX', ARGV[2])
  return 2
end
return 0
`)

func New(rds *redis.Redis, log *logger.Logger, timeout time.Duration) *Selector {
	return &Selector{
		rds:     rds,
		log:     log,
		timeout: timeout,
		script:  claimScript,
		active:  make(map[int64]map[int64]struct{}),
	}
}

func sourceKey(symbolID int64) string {
	return fmt.Sprintf("%s%d", keySourcePrefix, symbolID)
}

// Claim decides whether this feed's tick for a symbol is the accepted stream.
// activated and lost report a transition in or out of ownership, each worth a journal line.
func (s *Selector) Claim(ctx context.Context, datafeedID int64, feedIndex int32, symbolID int64) (accepted, activated, lost bool) {
	owner := fmt.Sprintf("%d:%d", feedIndex, datafeedID)
	res, err := s.script.Run(ctx, s.rds.Client, []string{sourceKey(symbolID)},
		owner, s.timeout.Milliseconds()).Int()
	if err != nil {
		// a redis hiccup must not silence the quote stream, so every feed passes
		s.mu.Lock()
		if time.Since(s.errLogAt) > time.Minute {
			s.errLogAt = time.Now()
			s.mu.Unlock()
			s.log.Log(logger.TypeNet, logger.CodeErr, "feed arbitration unavailable, accepting all sources",
				"error", err.Error())
		} else {
			s.mu.Unlock()
		}
		return true, false, false
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	owned := s.active[datafeedID]
	_, wasActive := owned[symbolID]
	switch res {
	case 0:
		if wasActive {
			delete(owned, symbolID)
		}
		return false, false, wasActive
	default:
		if !wasActive {
			if owned == nil {
				owned = make(map[int64]struct{})
				s.active[datafeedID] = owned
			}
			owned[symbolID] = struct{}{}
		}
		// res 2 with wasActive is a quiet symbol re-claiming after expiry, not a switch
		return true, res == 2 && !wasActive, false
	}
}

// Forget drops a stopped feed's local ownership view; the redis claims simply expire.
func (s *Selector) Forget(datafeedID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.active, datafeedID)
}

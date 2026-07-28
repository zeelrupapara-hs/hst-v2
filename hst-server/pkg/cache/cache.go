package cache

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"

	"hstserver/model"

	"github.com/cespare/xxhash/v2"
)

// Snapshot is everything the middleware needs to authorise a request. It is
// built once at login and read on every request after that.
type Snapshot struct {
	SessionId      string `json:"sid"`
	Login          int64  `json:"login"`
	ClientId       int64  `json:"cid"`
	Group          string `json:"grp"`
	Rights         int64  `json:"rights"`
	Scope          int32  `json:"scope"`
	ConnectionType int32  `json:"ct"`
	Restricted     bool   `json:"rst"`
	IsManager      bool   `json:"mgr"`
	// ManagerRights is the 77 right_ columns packed into two words, so a
	// permission check is a shift and an AND instead of a join.
	ManagerRights model.ManagerRights `json:"mrights"`
	Version       int64               `json:"ver"`
	CreatedAt     int64               `json:"created_at"`
	ExpiresAt     int64               `json:"expires_at"`
}

type entry struct {
	snap    *Snapshot
	expires time.Time
	elem    *list.Element
}

type bucket struct {
	mu    sync.RWMutex
	items map[string]*entry
	lru   *list.List
	cap   int
}

// Stats is what the monitor endpoint reports, so MaxAccounts is tuned from
// data rather than guessed.
type Stats struct {
	Entries     int     `json:"entries"`
	MaxAccounts int     `json:"max_accounts"`
	ShardId     uint64  `json:"shard_id"`
	ShardCount  uint64  `json:"shard_count"`
	Hits        uint64  `json:"hits"`
	Misses      uint64  `json:"misses"`
	HitRate     float64 `json:"hit_rate"`
	Evictions   uint64  `json:"evictions"`
	Foreign     uint64  `json:"foreign"`
}

// DistributeCache holds the snapshots this instance owns. Redis stays the
// source of truth, so a miss is only slower, never wrong.
type DistributeCache struct {
	shardId     uint64
	shardCount  uint64
	maxAccounts int
	ttl         time.Duration

	buckets []*bucket

	// login -> set of session ids, so revoking a user reaches every session
	idxMu sync.RWMutex
	idx   map[int64]map[string]struct{}

	hits      atomic.Uint64
	misses    atomic.Uint64
	evictions atomic.Uint64
	foreign   atomic.Uint64

	stop chan struct{}
	once sync.Once
}

// New builds the cache. bucketCount shards the lock, shardCount shards the data
// across instances; they are unrelated.
func New(shardId, shardCount, maxAccounts, bucketCount int, ttl time.Duration) *DistributeCache {
	if shardCount < 1 {
		shardCount = 1
	}
	if bucketCount < 1 {
		bucketCount = 1
	}

	// spread the capacity so no single bucket becomes the limit
	perBucket := maxAccounts / bucketCount
	if perBucket < 1 {
		perBucket = 1
	}

	d := &DistributeCache{
		// both are non negative, config.validate rejects anything else
		shardId:     uint64(shardId),    // #nosec G115
		shardCount:  uint64(shardCount), // #nosec G115
		maxAccounts: maxAccounts,
		ttl:         ttl,
		buckets:     make([]*bucket, bucketCount),
		idx:         make(map[int64]map[string]struct{}),
		stop:        make(chan struct{}),
	}

	for i := range d.buckets {
		d.buckets[i] = &bucket{
			items: make(map[string]*entry, perBucket),
			lru:   list.New(),
			cap:   perBucket,
		}
	}

	go d.janitor()

	return d
}

// Owns reports whether this instance is the cache owner of the session.
// Refusing foreign sessions is what bounds memory to this instance's share
// rather than to fleet wide traffic.
func (d *DistributeCache) Owns(sid string) bool {
	return xxhash.Sum64String(sid)%d.shardCount == d.shardId
}

// Get returns the cached snapshot. Roughly 60ns, no I/O.
func (d *DistributeCache) Get(sid string) (*Snapshot, bool) {
	b := d.bucketFor(sid)

	b.mu.RLock()
	e, ok := b.items[sid]
	if !ok {
		b.mu.RUnlock()
		d.misses.Add(1)
		return nil, false
	}
	snap, expires := e.snap, e.expires
	b.mu.RUnlock()

	if time.Now().After(expires) {
		d.Invalidate(sid)
		d.misses.Add(1)
		return nil, false
	}

	d.hits.Add(1)
	return snap, true
}

// Put stores a snapshot, unless the session belongs to another shard.
func (d *DistributeCache) Put(snap *Snapshot) {
	if snap == nil || !d.Owns(snap.SessionId) {
		d.foreign.Add(1)
		return
	}

	sid := snap.SessionId
	b := d.bucketFor(sid)

	b.mu.Lock()
	if e, ok := b.items[sid]; ok {
		e.snap = snap
		e.expires = time.Now().Add(d.ttl)
		b.lru.MoveToFront(e.elem)
		b.mu.Unlock()
		d.index(snap.Login, sid)
		return
	}

	// LRU is the safety valve: shard ownership decides what to cache, this
	// decides what to drop when MaxAccounts was set too low. Without it a bad
	// estimate is an OOM.
	for b.lru.Len() >= b.cap {
		oldest := b.lru.Back()
		if oldest == nil {
			break
		}
		victim := oldest.Value.(string)
		if ve, ok := b.items[victim]; ok {
			d.unindex(ve.snap.Login, victim)
			delete(b.items, victim)
		}
		b.lru.Remove(oldest)
		d.evictions.Add(1)
	}

	b.items[sid] = &entry{
		snap:    snap,
		expires: time.Now().Add(d.ttl),
		elem:    b.lru.PushFront(sid),
	}
	b.mu.Unlock()

	d.index(snap.Login, sid)
}

// Invalidate drops one session. Called for every instance regardless of
// ownership, since the pub/sub message is broadcast.
func (d *DistributeCache) Invalidate(sid string) {
	b := d.bucketFor(sid)

	b.mu.Lock()
	e, ok := b.items[sid]
	if ok {
		b.lru.Remove(e.elem)
		delete(b.items, sid)
	}
	b.mu.Unlock()

	if ok {
		d.unindex(e.snap.Login, sid)
	}
}

// InvalidateLogin drops every session of a login, for a rights or password change.
func (d *DistributeCache) InvalidateLogin(login int64) {
	d.idxMu.RLock()
	sids := make([]string, 0, len(d.idx[login]))
	for sid := range d.idx[login] {
		sids = append(sids, sid)
	}
	d.idxMu.RUnlock()

	for _, sid := range sids {
		d.Invalidate(sid)
	}
}

// Stats reports the numbers behind the memory dial.
func (d *DistributeCache) Stats() Stats {
	entries := 0
	for _, b := range d.buckets {
		b.mu.RLock()
		entries += len(b.items)
		b.mu.RUnlock()
	}

	hits, misses := d.hits.Load(), d.misses.Load()
	rate := 0.0
	if total := hits + misses; total > 0 {
		rate = float64(hits) / float64(total)
	}

	return Stats{
		Entries:     entries,
		MaxAccounts: d.maxAccounts,
		ShardId:     d.shardId,
		ShardCount:  d.shardCount,
		Hits:        hits,
		Misses:      misses,
		HitRate:     rate,
		Evictions:   d.evictions.Load(),
		Foreign:     d.foreign.Load(),
	}
}

// Close stops the janitor.
func (d *DistributeCache) Close() {
	d.once.Do(func() { close(d.stop) })
}

func (d *DistributeCache) bucketFor(sid string) *bucket {
	return d.buckets[xxhash.Sum64String(sid)%uint64(len(d.buckets))]
}

func (d *DistributeCache) index(login int64, sid string) {
	d.idxMu.Lock()
	defer d.idxMu.Unlock()

	if d.idx[login] == nil {
		d.idx[login] = make(map[string]struct{}, 1)
	}
	d.idx[login][sid] = struct{}{}
}

func (d *DistributeCache) unindex(login int64, sid string) {
	d.idxMu.Lock()
	defer d.idxMu.Unlock()

	if set, ok := d.idx[login]; ok {
		delete(set, sid)
		if len(set) == 0 {
			delete(d.idx, login)
		}
	}
}

// janitor reclaims expired entries so an idle session does not hold memory
// until it happens to be read again.
func (d *DistributeCache) janitor() {
	ticker := time.NewTicker(d.ttl)
	defer ticker.Stop()

	for {
		select {
		case <-d.stop:
			return
		case <-ticker.C:
			now := time.Now()
			for _, b := range d.buckets {
				b.mu.Lock()
				for sid, e := range b.items {
					if now.After(e.expires) {
						b.lru.Remove(e.elem)
						delete(b.items, sid)
						d.unindex(e.snap.Login, sid)
					}
				}
				b.mu.Unlock()
			}
		}
	}
}

package cache

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"

	"hstserver/model"

	"github.com/cespare/xxhash/v2"
)

// Session is an authenticated login: who they are, and what they may reach.
type Session struct {
	SessionId      string `json:"sid"`
	Login          int64  `json:"login"`
	ClientId       int64  `json:"cid"`
	Group          string `json:"grp"`
	Rights         int64  `json:"rights"`
	ConnectionType int32  `json:"ct"`
	// Scope is the password slot the session authenticated with: investor is read-only.
	Scope      int32 `json:"scope"`
	Restricted bool  `json:"rst"`
	IsManager  bool  `json:"mgr"`
	// ManagerRights is the 77 right columns packed into two words.
	ManagerRights model.ManagerRights `json:"mrights"`
	// ManagerGroups is the group access, a list of masks like demo\*.
	ManagerGroups []string `json:"mgroups,omitempty"`
	Version       int64    `json:"ver"`
	ExpiresAt     int64    `json:"expires_at"`
}

type cached struct {
	sess    *Session
	expires time.Time
	elem    *list.Element
}

type stripe struct {
	mu    sync.RWMutex
	items map[string]*cached
	lru   *list.List
	cap   int
}

// Stats is what the monitor endpoint reports.
type Stats struct {
	Entries       int     `json:"entries"`
	MaxAccounts   int     `json:"max_accounts"`
	InstanceId    uint64  `json:"shard_id"`
	InstanceCount uint64  `json:"shard_count"`
	Hits          uint64  `json:"hits"`
	Misses        uint64  `json:"misses"`
	HitRate       float64 `json:"hit_rate"`
	Evictions     uint64  `json:"evictions"`
	Foreign       uint64  `json:"foreign"`
}

// SessionCache holds the snapshots this instance owns.
type SessionCache struct {
	instanceId    uint64
	instanceCount uint64
	maxAccounts   int
	ttl           time.Duration

	stripes []*stripe

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

// New builds the cache.
func New(instanceId, instanceCount, maxAccounts, stripeCount int, ttl time.Duration) *SessionCache {
	if instanceCount < 1 {
		instanceCount = 1
	}
	if stripeCount < 1 {
		stripeCount = 1
	}

	// spread the capacity so no single bucket becomes the limit
	perBucket := maxAccounts / stripeCount
	if perBucket < 1 {
		perBucket = 1
	}

	d := &SessionCache{
		// both are non negative, config.validate rejects anything else
		instanceId:    uint64(instanceId),    // #nosec G115
		instanceCount: uint64(instanceCount), // #nosec G115
		maxAccounts:   maxAccounts,
		ttl:           ttl,
		stripes:       make([]*stripe, stripeCount),
		idx:           make(map[int64]map[string]struct{}),
		stop:          make(chan struct{}),
	}

	for i := range d.stripes {
		d.stripes[i] = &stripe{
			items: make(map[string]*cached, perBucket),
			lru:   list.New(),
			cap:   perBucket,
		}
	}

	go d.janitor()

	return d
}

// Owns reports whether this instance is the cache owner of the session.
func (d *SessionCache) Owns(sid string) bool {
	return xxhash.Sum64String(sid)%d.instanceCount == d.instanceId
}

// Get returns the cached snapshot. Roughly 60ns, no I/O.
func (d *SessionCache) Get(sid string) (*Session, bool) {
	b := d.stripeFor(sid)

	b.mu.RLock()
	e, ok := b.items[sid]
	if !ok {
		b.mu.RUnlock()
		d.misses.Add(1)
		return nil, false
	}
	sess, expires := e.sess, e.expires
	b.mu.RUnlock()

	if time.Now().After(expires) {
		d.Invalidate(sid)
		d.misses.Add(1)
		return nil, false
	}

	d.hits.Add(1)
	return sess, true
}

// Put stores a snapshot, unless the session belongs to another shard.
func (d *SessionCache) Put(sess *Session) {
	if sess == nil || !d.Owns(sess.SessionId) {
		d.foreign.Add(1)
		return
	}

	sid := sess.SessionId
	b := d.stripeFor(sid)

	b.mu.Lock()
	if e, ok := b.items[sid]; ok {
		e.sess = sess
		e.expires = time.Now().Add(d.ttl)
		b.lru.MoveToFront(e.elem)
		b.mu.Unlock()
		d.index(sess.Login, sid)
		return
	}

	// LRU is the safety valve:
	for b.lru.Len() >= b.cap {
		oldest := b.lru.Back()
		if oldest == nil {
			break
		}
		victim := oldest.Value.(string)
		if ve, ok := b.items[victim]; ok {
			d.unindex(ve.sess.Login, victim)
			delete(b.items, victim)
		}
		b.lru.Remove(oldest)
		d.evictions.Add(1)
	}

	b.items[sid] = &cached{
		sess:    sess,
		expires: time.Now().Add(d.ttl),
		elem:    b.lru.PushFront(sid),
	}
	b.mu.Unlock()

	d.index(sess.Login, sid)
}

// Invalidate drops one session.
func (d *SessionCache) Invalidate(sid string) {
	b := d.stripeFor(sid)

	b.mu.Lock()
	e, ok := b.items[sid]
	if ok {
		b.lru.Remove(e.elem)
		delete(b.items, sid)
	}
	b.mu.Unlock()

	if ok {
		d.unindex(e.sess.Login, sid)
	}
}

// InvalidateLogin drops every session of a login, for a rights or password change.
func (d *SessionCache) InvalidateLogin(login int64) {
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
func (d *SessionCache) Stats() Stats {
	entries := 0
	for _, b := range d.stripes {
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
		Entries:       entries,
		MaxAccounts:   d.maxAccounts,
		InstanceId:    d.instanceId,
		InstanceCount: d.instanceCount,
		Hits:          hits,
		Misses:        misses,
		HitRate:       rate,
		Evictions:     d.evictions.Load(),
		Foreign:       d.foreign.Load(),
	}
}

// Close stops the janitor.
func (d *SessionCache) Close() {
	d.once.Do(func() { close(d.stop) })
}

func (d *SessionCache) stripeFor(sid string) *stripe {
	return d.stripes[xxhash.Sum64String(sid)%uint64(len(d.stripes))]
}

func (d *SessionCache) index(login int64, sid string) {
	d.idxMu.Lock()
	defer d.idxMu.Unlock()

	if d.idx[login] == nil {
		d.idx[login] = make(map[string]struct{}, 1)
	}
	d.idx[login][sid] = struct{}{}
}

func (d *SessionCache) unindex(login int64, sid string) {
	d.idxMu.Lock()
	defer d.idxMu.Unlock()

	if set, ok := d.idx[login]; ok {
		delete(set, sid)
		if len(set) == 0 {
			delete(d.idx, login)
		}
	}
}

// janitor reclaims expired entries so idle sessions free their memory.
func (d *SessionCache) janitor() {
	ticker := time.NewTicker(d.ttl)
	defer ticker.Stop()

	for {
		select {
		case <-d.stop:
			return
		case <-ticker.C:
			now := time.Now()
			for _, b := range d.stripes {
				b.mu.Lock()
				for sid, e := range b.items {
					if now.After(e.expires) {
						b.lru.Remove(e.elem)
						delete(b.items, sid)
						d.unindex(e.sess.Login, sid)
					}
				}
				b.mu.Unlock()
			}
		}
	}
}

package handler

import (
	"context"
	"sort"
	"strconv"
	"time"

	"hstcore/internal/book"
	"hstcore/internal/shardmap"
	"hstcore/model"
	"hstcore/pkg/logger"
)

// Every pod claims itself in redis, and every shard it serves is leased there under its name.

const (
	// podKeyPrefix is where a pod says it is alive.
	podKeyPrefix = "core:pods:"
	// shardKeyPrefix is where a pod says it serves a shard; the gainer waits until the loser's lease is gone.
	shardKeyPrefix = "core:shard:"
	// podTTL is how long that claim stands without a refresh
	podTTL = 15 * time.Second
	// podRefresh is how often the claim is renewed
	podRefresh = 5 * time.Second
	// leaseRetry is how often a gainer asks again for a shard still leased to another pod
	leaseRetry = 250 * time.Millisecond
)

// WatchPodMembership keeps this pod registered and reacts to the others.
func (h *Handler) WatchPodMembership(ctx context.Context) {
	ticker := time.NewTicker(podRefresh)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.unregister()
			return
		case <-ticker.C:
			h.register(ctx)
			h.ReassignShards(ctx)
		}
	}
}

// register says this pod is alive for another podTTL, and renews the lease on every shard it serves.
func (h *Handler) register(ctx context.Context) {
	if err := h.Redis.Client.Set(ctx, podKeyPrefix+h.name, "1", podTTL).Err(); err != nil {
		h.Log.Log(logger.TypeSys, logger.CodeWarn, "could not register this pod",
			"pod", h.name, "error", err.Error())
	}

	m := h.shards()
	if m == nil {
		return
	}

	pipe := h.Redis.Client.Pipeline()
	for _, s := range m.Mine() {
		pipe.Expire(ctx, shardKey(s), podTTL)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		h.Log.Log(logger.TypeSys, logger.CodeWarn, "could not renew the shard leases",
			"pod", h.name, "error", err.Error())
	}
}

// unregister gives up this pod's claim at shutdown, so its accounts move sooner than the expiry.
func (h *Handler) unregister() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := h.Redis.Client.Del(ctx, podKeyPrefix+h.name).Err(); err != nil {
		h.Log.Log(logger.TypeSys, logger.CodeWarn, "could not unregister this pod",
			"pod", h.name, "error", err.Error())
	}
}

// shards is the current map, read under the lock that guards its swap.
func (h *Handler) shards() *shardmap.Map {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.Shards
}

func shardKey(shard uint32) string {
	return shardKeyPrefix + strconv.FormatUint(uint64(shard), 10)
}

// livePods lists the pods currently claiming to be alive, or nil when redis could not say.
func (h *Handler) livePods(ctx context.Context) []string {
	keys, err := h.Redis.Client.Keys(ctx, podKeyPrefix+"*").Result()
	if err != nil {
		h.Log.Log(logger.TypeSys, logger.CodeWarn, "could not read the pod list",
			"error", err.Error())
		// nothing is known, so nothing moves: the map that stands keeps standing
		return nil
	}

	pods := make([]string, 0, len(keys))
	for _, k := range keys {
		pods = append(pods, k[len(podKeyPrefix):])
	}
	sort.Strings(pods)

	if len(pods) == 0 {
		return []string{h.name}
	}

	return pods
}

// acquireLeases claims each shard, waiting for whoever served it last to let go or expire.
func (h *Handler) acquireLeases(ctx context.Context, shards []uint32) {
	for _, s := range shards {
		key := shardKey(s)

		for {
			ok, err := h.Redis.Client.SetNX(ctx, key, h.name, podTTL).Result()
			if err != nil {
				h.Log.Log(logger.TypeSys, logger.CodeWarn, "could not lease a shard",
					"shard", s, "error", err.Error())
			} else if ok {
				break
			} else if holder, _ := h.Redis.Client.Get(ctx, key).Result(); holder == h.name {
				// a lease this pod left behind when it last went down
				break
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(leaseRetry):
			}

			// the wait must not cost this pod its own registration
			if err := h.Redis.Client.Set(ctx, podKeyPrefix+h.name, "1", podTTL).Err(); err != nil {
				h.Log.Log(logger.TypeSys, logger.CodeWarn, "could not register this pod",
					"pod", h.name, "error", err.Error())
			}
		}
	}
}

// releaseLeases lets go of shards this pod flushed, so the gainer can load them.
func (h *Handler) releaseLeases(ctx context.Context, shards []uint32) {
	if len(shards) == 0 {
		return
	}

	pipe := h.Redis.Client.Pipeline()
	for _, s := range shards {
		pipe.Del(ctx, shardKey(s))
	}
	if _, err := pipe.Exec(ctx); err != nil {
		h.Log.Log(logger.TypeSys, logger.CodeWarn, "could not release the shard leases",
			"shards", len(shards), "error", err.Error())
	}
}

// ReassignShards rebuilds the map and moves accounts if the membership changed.
func (h *Handler) ReassignShards(ctx context.Context) {
	pods := h.livePods(ctx)
	if pods == nil {
		return
	}
	next := shardmap.New(h.name, pods)

	current := h.shards()

	if current != nil && sameShards(current.Mine(), next.Mine()) {
		return
	}

	gained, lost := shardDiff(current, next)

	h.mu.Lock()
	h.Shards = next
	h.mu.Unlock()

	h.Log.Log(logger.TypeSys, logger.CodeOK, "shards moved",
		"pod", h.name, "pods", len(pods),
		"holding", len(next.Mine()), "gained", len(gained), "lost", len(lost))

	// stop listening before letting go, so nothing arrives for an account this pod no longer has,
	// and start listening only once the accounts are here to answer for
	h.CloseLostInboxes(lost)
	h.ReleaseLostAccounts(ctx, lost)
	h.acquireLeases(ctx, gained)
	h.LoadGainedAccounts(ctx, gained)
	h.OpenGainedInboxes(gained)
}

func sameShards(a, b []uint32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func shardDiff(before, after *shardmap.Map) (gained, lost []uint32) {
	if before == nil {
		return after.Mine(), nil
	}

	for _, s := range after.Mine() {
		if !before.Holds(s) {
			gained = append(gained, s)
		}
	}
	for _, s := range before.Mine() {
		if !after.Holds(s) {
			lost = append(lost, s)
		}
	}

	return gained, lost
}

// ReleaseLostAccounts writes out and forgets the accounts that now belong elsewhere.
func (h *Handler) ReleaseLostAccounts(ctx context.Context, lost []uint32) {
	if len(lost) == 0 {
		return
	}

	set := make(map[uint32]bool, len(lost))
	for _, s := range lost {
		set[s] = true
	}

	var dropped []int64

	h.Accounts.Each(func(e *book.Entry) {
		if set[shardmap.ShardOf(e.Account.Login)] {
			dropped = append(dropped, e.Account.Login)
		}
	})

	// the dealer queue is in memory too, so the requests of moved accounts go back on the desk of the new owner
	h.dropDealingFor(set)

	// the margin and profit worked out since the last write only exist here, so they go to the
	// database before the account does, or the pod taking it over reads a stale balance
	for _, login := range dropped {
		if e, ok := h.Accounts.Get(login); ok {
			h.flushAccount(ctx, e)
		}

		h.Accounts.Remove(login)
	}

	h.releaseLeases(ctx, lost)

	h.Log.Log(logger.TypeSys, logger.CodeOK, "accounts released",
		"shards", len(lost), "accounts", len(dropped))
}

// flushAccount marks an account released and writes its money if this pod ever changed it.
func (h *Handler) flushAccount(ctx context.Context, e *book.Entry) {
	e.Lock()
	e.Released = true
	dirty := e.Dirty
	e.Dirty = false
	account := *e.Account
	e.Unlock()

	if !dirty {
		return
	}

	if err := h.SaveAccount(ctx, &account); err != nil {
		h.Log.Log(logger.TypeSys, logger.CodeErr, "could not write an account being released",
			"login", account.Login, "error", err.Error())
	}
}

// LoadGainedAccounts loads the accounts that now belong here, and only those.
func (h *Handler) LoadGainedAccounts(ctx context.Context, gained []uint32) {
	if len(gained) == 0 {
		return
	}

	before := h.Accounts.Len()

	set := make(map[uint32]bool, len(gained))
	for _, shard := range gained {
		set[shard] = true
	}

	// only the shards just gained, rather than every account this pod could ever hold
	if err := h.LoadAccountsIn(ctx, set); err != nil {
		h.Log.Log(logger.TypeSys, logger.CodeErr, "could not load the accounts this pod gained",
			"shards", len(gained), "error", err.Error())
		return
	}

	// requests the old owner had on a desk are answerable only here now
	h.RecoverDealingRequests(set)

	h.Log.Log(logger.TypeSys, logger.CodeOK, "accounts taken on",
		"shards", len(gained), "accounts", h.Accounts.Len()-before)
}

// CloseLostInboxes stops this pod answering for the shards it no longer holds.
func (h *Handler) CloseLostInboxes(lost []uint32) {
	for _, shard := range lost {
		// every inbox the shard opened, or this pod keeps answering for accounts it let go
		dropped := make(map[string]bool, len(commands))
		for _, c := range commands {
			dropped[model.SubjectOwner(shard, c.topic)] = true
		}

		h.mu.Lock()
		kept := h.subs[:0]
		for _, sub := range h.subs {
			if dropped[sub.Subject] {
				if err := sub.Unsubscribe(); err != nil {
					h.Log.Log(logger.TypeNet, logger.CodeWarn, "could not close a shard inbox",
						"subject", sub.Subject, "error", err.Error())
				}
				continue
			}
			kept = append(kept, sub)
		}
		h.subs = kept
		h.mu.Unlock()
	}

}

// OpenGainedInboxes starts this pod answering for the shards it now holds.
func (h *Handler) OpenGainedInboxes(gained []uint32) {
	for _, shard := range gained {
		if err := h.subscribeOwned(shard); err != nil {
			h.Log.Log(logger.TypeNet, logger.CodeErr, "could not open the inbox of a shard gained",
				"shard", shard, "error", err.Error())
		}
	}
}

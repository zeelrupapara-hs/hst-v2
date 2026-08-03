package handler

import (
	"context"
	"sort"
	"time"

	"hstcore/internal/book"
	"hstcore/internal/shardmap"
	"hstcore/model"
	"hstcore/pkg/logger"
)

// Every pod claims itself in redis.

const (
	// podKeyPrefix is where a pod says it is alive.
	podKeyPrefix = "core:pods:"
	// podTTL is how long that claim stands without a refresh
	podTTL = 15 * time.Second
	// podRefresh is how often the claim is renewed
	podRefresh = 5 * time.Second
)

// WatchPodMembership keeps this pod registered and reacts to the others.
func (h *Handler) WatchPodMembership(ctx context.Context) {
	// register before the first look, so this pod is in its own first map
	h.register(ctx)
	h.ReassignShards(ctx)

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

// register says this pod is alive for another podTTL.
func (h *Handler) register(ctx context.Context) {
	if err := h.Redis.Client.Set(ctx, podKeyPrefix+h.name, "1", podTTL).Err(); err != nil {
		h.Log.Log(logger.TypeSys, logger.CodeWarn, "could not register this pod",
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

// livePods lists the pods currently claiming to be alive.
func (h *Handler) livePods(ctx context.Context) []string {
	keys, err := h.Redis.Client.Keys(ctx, podKeyPrefix+"*").Result()
	if err != nil {
		h.Log.Log(logger.TypeSys, logger.CodeWarn, "could not read the pod list",
			"error", err.Error())
		// alone is the safe assumption: keep what we have rather than drop unclaimed accounts
		return []string{h.name}
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

// ReassignShards rebuilds the map and moves accounts if the membership changed.
func (h *Handler) ReassignShards(ctx context.Context) {
	pods := h.livePods(ctx)
	next := shardmap.New(h.name, pods)

	h.mu.Lock()
	current := h.Shards
	h.mu.Unlock()

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

	// the margin and profit worked out since the last write only exist here, so they go to the
	// database before the account does, or the pod taking it over reads a stale balance
	for _, login := range dropped {
		if e, ok := h.Accounts.Get(login); ok {
			h.flushAccount(ctx, e)
		}

		h.Accounts.Remove(login)
	}

	h.Log.Log(logger.TypeSys, logger.CodeOK, "accounts released",
		"shards", len(lost), "accounts", len(dropped))
}

// flushAccount writes one account's money before this pod stops holding it.
func (h *Handler) flushAccount(ctx context.Context, e *book.Entry) {
	e.Lock()
	account := *e.Account
	e.Unlock()

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

const ShardCount = shardmap.Count

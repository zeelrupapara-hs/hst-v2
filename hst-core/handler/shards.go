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

// runMembership keeps this pod registered and reacts to the others.
func (h *Handler) runMembership(ctx context.Context) {
	// register before the first look, so this pod is in its own first map
	h.register(ctx)
	h.rebalance(ctx)

	ticker := time.NewTicker(podRefresh)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.unregister()
			return
		case <-ticker.C:
			h.register(ctx)
			h.rebalance(ctx)
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

// rebalance rebuilds the map and moves accounts if the membership changed.
func (h *Handler) rebalance(ctx context.Context) {
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

	// drop first, so an account is never held by two pods at once even for a moment
	h.dropShards(lost)
	h.takeShards(ctx, gained)
	h.resubscribe(gained, lost)
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

// dropShards forgets the accounts that now belong elsewhere; nothing is written on the way out.
func (h *Handler) dropShards(lost []uint32) {
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

	for _, login := range dropped {
		h.Accounts.Remove(login)
	}

	h.Log.Log(logger.TypeSys, logger.CodeOK, "accounts released",
		"shards", len(lost), "accounts", len(dropped))
}

// takeShards loads the accounts that now belong here.
func (h *Handler) takeShards(ctx context.Context, gained []uint32) {
	if len(gained) == 0 {
		return
	}

	before := h.Accounts.Len()

	// the loaders already skip anything outside this pod's shards, so the new map is enough
	if err := h.LoadAccount(ctx); err != nil {
		h.Log.Log(logger.TypeSys, logger.CodeErr, "could not load the accounts this pod gained",
			"shards", len(gained), "error", err.Error())
		return
	}

	h.Log.Log(logger.TypeSys, logger.CodeOK, "accounts taken on",
		"shards", len(gained), "accounts", h.Accounts.Len()-before)
}

// resubscribe listens to the shards this pod gained and stops listening to the ones it lost.
func (h *Handler) resubscribe(gained, lost []uint32) {
	for _, shard := range lost {
		dropped := map[string]bool{
			model.SubjectShardOrders(shard):    true,
			model.SubjectShardPositions(shard): true,
			model.SubjectShardDealing(shard):   true,
		}

		h.mu.Lock()
		kept := h.subs[:0]
		for _, sub := range h.subs {
			if dropped[sub.Subject] {
				if err := sub.Unsubscribe(); err != nil {
					h.Log.Log(logger.TypeNet, logger.CodeWarn, "could not stop listening to a shard",
						"subject", sub.Subject, "error", err.Error())
				}
				continue
			}
			kept = append(kept, sub)
		}
		h.subs = kept
		h.mu.Unlock()
	}

	for _, shard := range gained {
		if err := h.subscribeShard(shard); err != nil {
			h.Log.Log(logger.TypeNet, logger.CodeErr, "could not listen to a shard this pod gained",
				"shard", shard, "error", err.Error())
		}
	}
}

const ShardCount = shardmap.Count

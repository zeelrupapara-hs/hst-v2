package handler

import (
	"context"
	"strconv"
	"sync"
	"time"

	"hstcore/pkg/logger"
)

// Who is on the dealing desk right now.
//
// The api server writes a leased key when a manager connects as a dealer and drops it when they
// leave, so a terminal that dies stops being offered work on its own. The engine reads that set,
// and holds each answer briefly: routing asks on every request that reaches a dealer rule, and a
// round trip to redis on the trade path would cost more than the freshness is worth.

const (
	dealerOnlineKeyPrefix = "dealers:online:"
	dealerOnlineFresh     = 2 * time.Second
)

type deskEntry struct {
	online bool
	at     time.Time
}

type deskState struct {
	mu    sync.RWMutex
	known map[int64]deskEntry
}

// onlineOf keeps only the dealers who have connected as such.
func (h *Handler) onlineOf(dealers []int64) []int64 {
	if len(dealers) == 0 {
		return nil
	}

	stale := h.staleDealers(dealers)
	if len(stale) > 0 {
		h.refreshDesk(stale)
	}

	h.desk.mu.RLock()
	defer h.desk.mu.RUnlock()

	out := make([]int64, 0, len(dealers))
	for _, login := range dealers {
		if h.desk.known[login].online {
			out = append(out, login)
		}
	}

	return out
}

// staleDealers is who we have no fresh answer for.
func (h *Handler) staleDealers(dealers []int64) []int64 {
	h.desk.mu.RLock()
	defer h.desk.mu.RUnlock()

	var stale []int64

	for _, login := range dealers {
		if e, ok := h.desk.known[login]; !ok || time.Since(e.at) >= dealerOnlineFresh {
			stale = append(stale, login)
		}
	}

	return stale
}

// refreshDesk asks redis about the logins we are unsure of, and remembers the answer.
func (h *Handler) refreshDesk(dealers []int64) {
	if h.Redis == nil || h.Redis.Client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	keys := make([]string, 0, len(dealers))
	for _, login := range dealers {
		keys = append(keys, dealerOnlineKeyPrefix+strconv.FormatInt(login, 10))
	}

	values, err := h.Redis.Client.MGet(ctx, keys...).Result()
	if err != nil {
		h.Log.Log(logger.TypeNet, logger.CodeWarn, "could not read who is on the desk",
			"error", err.Error())
		return
	}

	now := time.Now()

	h.desk.mu.Lock()
	defer h.desk.mu.Unlock()

	if h.desk.known == nil {
		h.desk.known = make(map[int64]deskEntry, len(dealers))
	}

	for i, login := range dealers {
		online := i < len(values) && values[i] != nil
		h.desk.known[login] = deskEntry{online: online, at: now}
	}
}

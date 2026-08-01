package handler

import (
	"context"
	"time"

	"hstcore/pkg/logger"
)

// The end of the trading day.
//
// Swaps are charged once, at the same moment every day. Every pod runs this over its own
// accounts, so no coordination is needed: an account belongs to one pod, and that pod charges
// it. If a pod is down at rollover its accounts are charged by whoever picks up the shard,
// which is why the last charge date is recorded per account rather than assumed.

// rolloverHour is when the trading day turns over, in the server's clock.
//
// ponytail: fixed hour, make it a group setting if brokers need different ones
const rolloverHour = 0

// runDaily waits for the rollover and charges the day.
func (h *Handler) runDaily(ctx context.Context) {
	for {
		wait := untilNextRollover(time.Now())

		h.Log.Log(logger.TypeSys, logger.CodeOK, "next rollover",
			"in", wait.Round(time.Minute).String())

		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}

		h.ChargeSwaps(ctx, time.Now())
	}
}

// untilNextRollover is how long until the next rollover hour.
func untilNextRollover(now time.Time) time.Duration {
	next := time.Date(now.Year(), now.Month(), now.Day(), rolloverHour, 0, 0, 0, now.Location())
	if !next.After(now) {
		next = next.AddDate(0, 0, 1)
	}

	return next.Sub(now)
}

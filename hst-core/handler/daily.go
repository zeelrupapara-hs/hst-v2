package handler

import (
	"context"
	"time"

	"hstcore/pkg/logger"
)

// Every pod charges its own accounts at rollover, so no coordination is needed.

// ponytail: fixed rollover hour, make it a group setting if brokers need different ones
const rolloverHour = 0

// dealingSweep is how often the dealer queue is checked for requests that ran out of time.
const dealingSweep = time.Second

// runDaily sweeps the dealer queue every second and charges the day at rollover.
func (h *Handler) runDaily(ctx context.Context) {
	sweep := time.NewTicker(dealingSweep)
	defer sweep.Stop()

	for {
		wait := untilNextRollover(time.Now())

		h.Log.Log(logger.TypeSys, logger.CodeOK, "next rollover",
			"in", wait.Round(time.Minute).String())

		rollover := time.After(wait)

		for turned := false; !turned; {
			select {
			case <-ctx.Done():
				return
			case <-sweep.C:
				h.CheckDealingRequests()
			case <-rollover:
				turned = true
			}
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

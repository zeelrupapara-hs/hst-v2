package handler

import (
	"math"
	"testing"
	"time"

	"hstcore/internal/settings"
	"hstcore/model"
)

func swapRules(mode int32, long, short float64) *settings.Rules {
	return &settings.Rules{
		SwapMode: mode, SwapLong: long, SwapShort: short,
		ContractSize: 100000, Point: 0.00001, TickSize: 0.00001, TickValue: 1,
	}
}

// A swap in points is worth a tick each, times the size of the position.
func TestSwapInPoints(t *testing.T) {
	h := newEngine()
	p := &model.Position{Volume: model.Volume(1), Action: 0}

	// -5 points on one lot, one day
	if got := h.swapFor(swapRules(swapPoints, -5, 3), p, 1); math.Abs(got+5) > 1e-6 {
		t.Fatalf("long swap = %v, want -5", got)
	}
}

// A short position pays the short rate, not the long one.
func TestSwapUsesTheSideRate(t *testing.T) {
	h := newEngine()
	short := &model.Position{Volume: model.Volume(1), Action: 1}

	if got := h.swapFor(swapRules(swapPoints, -5, 3), short, 1); math.Abs(got-3) > 1e-6 {
		t.Fatalf("short swap = %v, want 3", got)
	}
}

// Wednesday carries three days, because the position is really being held over the weekend.
func TestWednesdayChargesThreeDays(t *testing.T) {
	h := newEngine()
	p := &model.Position{Volume: model.Volume(1)}

	if got := h.swapFor(swapRules(swapPoints, -5, 0), p, 3); math.Abs(got+15) > 1e-6 {
		t.Fatalf("triple swap = %v, want -15", got)
	}
}

// A disabled swap costs nothing however big the rate looks.
func TestSwapDisabled(t *testing.T) {
	h := newEngine()
	p := &model.Position{Volume: model.Volume(10)}

	if got := h.swapFor(swapRules(swapDisabled, -5, -5), p, 1); got != 0 {
		t.Fatalf("swap = %v, want 0 when the mode is disabled", got)
	}
}

// A yearly percentage is spread over the days it is held.
func TestSwapAsYearlyPercent(t *testing.T) {
	h := newEngine()
	p := &model.Position{Volume: model.Volume(1), PriceCurrent: 1.1, PriceOpen: 1.1}

	// 3.6% a year on a 110000 position is 11 a day on a 360 day count
	got := h.swapFor(swapRules(swapPercentCurrent, -3.6, 0), p, 1)
	if math.Abs(got+11) > 1e-6 {
		t.Fatalf("percent swap = %v, want -11", got)
	}
}

// The rollover is always ahead of now, never in the past.
func TestRolloverIsAlwaysAhead(t *testing.T) {
	for _, hour := range []int{0, 5, 12, 23} {
		now := time.Date(2026, 8, 1, hour, 30, 0, 0, time.UTC)
		if d := untilNextRollover(now); d <= 0 || d > 24*time.Hour {
			t.Fatalf("at %02d:30 the next rollover is in %v", hour, d)
		}
	}
}

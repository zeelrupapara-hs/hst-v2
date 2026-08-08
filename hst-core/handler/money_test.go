package handler

import (
	"testing"

	"hstcore/internal/settings"
	"hstcore/model"
)

// The reference defines the split with three worked examples on D=4; a one-point drift here is
// a wrong fill price, so they are pinned exactly.
func TestCalculateAccountSpread(t *testing.T) {
	tick := model.Tick{Bid: 1.10000, Ask: 1.10000}

	cases := []struct {
		d, b     int32
		bid, ask float64
	}{
		{4, 0, 1.09998, 1.10002},
		{4, -1, 1.09997, 1.10001},
		{4, 1, 1.09999, 1.10003},
		{0, 0, 1.10000, 1.10000},
		{0, 2, 1.10002, 1.10002},
	}

	for _, c := range cases {
		r := &settings.Rules{SpreadDiff: c.d, SpreadDiffBalance: c.b, Point: 0.00001, Digits: 5}
		got := CalculateAccountSpread(r, tick)
		if got.Bid != c.bid || got.Ask != c.ask {
			t.Errorf("D=%d B=%d: got %.5f/%.5f, want %.5f/%.5f", c.d, c.b, got.Bid, got.Ask, c.bid, c.ask)
		}
	}
}

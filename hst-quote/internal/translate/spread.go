package translate

import (
	"math"

	"hstquote/model"
)

// ApplySpread transforms Bid/Ask per the symbol's fixed spread and spread balance.
func ApplySpread(set model.SymbolSettings, t *model.Tick) {
	// market depth symbols keep the book's own spread
	if set.HasDOM() {
		return
	}
	if set.Spread == 0 && set.SpreadBalance == 0 {
		return
	}

	pt := set.PointValue()
	digits := set.Digits
	if digits <= 0 {
		digits = 5
	}
	b := float64(set.SpreadBalance)

	switch {
	case set.Spread == 0:
		t.Bid = round(t.Bid+pt*b, digits)
		t.Ask = round(t.Ask+pt*b, digits)
	case math.Round((t.Ask-t.Bid)/pt) == float64(set.Spread):
		bid := t.Bid + pt*b
		t.Bid = round(bid, digits)
		t.Ask = round(bid+pt*float64(set.Spread), digits)
	default:
		bid := (t.Ask+t.Bid)/2 - pt*(float64(set.Spread)/2-b)
		t.Bid = round(bid, digits)
		t.Ask = round(bid+pt*float64(set.Spread), digits)
	}
}

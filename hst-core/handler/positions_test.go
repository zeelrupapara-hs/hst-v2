package handler

import (
	"math"
	"testing"

	"hstcore/internal/book"
	"hstcore/internal/settings"
	"hstcore/model"
	"hstcore/pkg/logger"
)

func rules(netting bool) *settings.Rules {
	mode := int32(model.MarginRetailNetting)
	if !netting {
		mode = int32(model.MarginRetailHedging)
	}

	return &settings.Rules{
		CalcMode:     model.CalcForex,
		ContractSize: 100000,
		Digits:       5,
		Point:        0.00001,
		Group:        &model.Group{MarginMode: mode},
	}
}

func account() *book.Entry {
	return &book.Entry{
		Account:   &model.Account{Login: 1001, Group: `demo\forex`, Balance: 10000, Leverage: 100},
		Orders:    map[int64]*model.Order{},
		Positions: map[int64]*model.Position{},
	}
}

func order(buy bool, lots float64) *model.Order {
	t := int32(model.OrderBuy)
	if !buy {
		t = int32(model.OrderSell)
	}
	return &model.Order{Login: 1001, Symbol: "EURUSD", Type: t, VolumeCurrent: model.Volume(lots)}
}

// place puts the fill's result onto the account, the way the order handler does.
func place(e *book.Entry, f *Fill, nextId int64) {
	for _, p := range f.Closed {
		delete(e.Positions, p.PositionId)
	}
	if f.Opened != nil {
		f.Opened.PositionId = nextId
		e.Positions[nextId] = f.Opened
	}
}

func newEngine() *Handler { return &Handler{Log: logger.NewNop()} }

// Under netting, a second buy on the same symbol grows the one position and the open price
// becomes the weighted average of the two fills.
func TestNettingSecondBuyGrowsAndAverages(t *testing.T) {
	h, e, r := newEngine(), account(), rules(true)

	place(e, h.Execute(e, order(true, 1), r, 1.1000, 1), 1)
	h.Execute(e, order(true, 1), r, 1.1010, 2)

	if len(e.Positions) != 1 {
		t.Fatalf("netting kept %d positions, want 1", len(e.Positions))
	}

	p := e.Positions[1]
	if p.Volume != model.Volume(2) {
		t.Fatalf("volume = %v lots, want 2", p.Lots())
	}
	if math.Abs(p.PriceOpen-1.1005) > 1e-9 {
		t.Fatalf("open price = %v, want the average 1.10050", p.PriceOpen)
	}
}

// Under hedging the same two buys are two separate positions, each at its own price.
func TestHedgingKeepsPositionsApart(t *testing.T) {
	h, e, r := newEngine(), account(), rules(false)

	place(e, h.Execute(e, order(true, 1), r, 1.1000, 1), 1)
	place(e, h.Execute(e, order(true, 1), r, 1.1010, 2), 2)

	if len(e.Positions) != 2 {
		t.Fatalf("hedging kept %d positions, want 2", len(e.Positions))
	}
	if e.Positions[1].PriceOpen == e.Positions[2].PriceOpen {
		t.Fatal("hedged positions should each keep their own open price")
	}
}

// An opposite fill for less than the position shrinks it and books the profit on the part that
// closed.
func TestNettingOppositeFillReduces(t *testing.T) {
	h, e, r := newEngine(), account(), rules(true)

	place(e, h.Execute(e, order(true, 2), r, 1.1000, 1), 1)
	f := h.Execute(e, order(false, 1), r, 1.1010, 2)
	place(e, f, 2)

	p := e.Positions[1]
	if p.Volume != model.Volume(1) {
		t.Fatalf("volume left = %v lots, want 1", p.Lots())
	}
	// one lot closed 10 points in profit: 0.0010 * 100000 = 100
	if math.Abs(f.Profit-100) > 1e-6 {
		t.Fatalf("profit = %v, want 100", f.Profit)
	}
	if len(f.Deals) != 1 || model.DealEntry(f.Deals[0].Entry) != model.EntryOut {
		t.Fatal("a reducing fill should book one deal marked out")
	}
}

// An opposite fill for exactly the position closes it.
func TestNettingOppositeFillCloses(t *testing.T) {
	h, e, r := newEngine(), account(), rules(true)

	place(e, h.Execute(e, order(true, 1), r, 1.1000, 1), 1)
	f := h.Execute(e, order(false, 1), r, 1.1010, 2)
	place(e, f, 2)

	if len(e.Positions) != 0 {
		t.Fatalf("%d positions left after closing out, want none", len(e.Positions))
	}
	if math.Abs(f.Profit-100) > 1e-6 {
		t.Fatalf("profit = %v, want 100", f.Profit)
	}
}

// An opposite fill for more than the position turns it around: the old one closes and the
// remainder opens the other way.
func TestNettingOppositeFillReverses(t *testing.T) {
	h, e, r := newEngine(), account(), rules(true)

	place(e, h.Execute(e, order(true, 1), r, 1.1000, 1), 1)
	f := h.Execute(e, order(false, 3), r, 1.1010, 2)
	place(e, f, 2)

	if len(e.Positions) != 1 {
		t.Fatalf("%d positions after a reversal, want 1", len(e.Positions))
	}

	p := e.Positions[2]
	if p.Buy() {
		t.Fatal("the position should have turned around to a sell")
	}
	if p.Volume != model.Volume(2) {
		t.Fatalf("remainder = %v lots, want 2", p.Lots())
	}
	if model.DealEntry(f.Deals[0].Entry) != model.EntryInOut {
		t.Fatal("a reversal should book one deal marked in and out")
	}
}

// Hedging never reverses: an opposite fill just opens another position beside the first.
func TestHedgingNeverReverses(t *testing.T) {
	h, e, r := newEngine(), account(), rules(false)

	place(e, h.Execute(e, order(true, 1), r, 1.1000, 1), 1)
	place(e, h.Execute(e, order(false, 3), r, 1.1010, 2), 2)

	if len(e.Positions) != 2 {
		t.Fatalf("%d positions, want 2 sitting side by side", len(e.Positions))
	}
	if e.Positions[1].Volume != model.Volume(1) {
		t.Fatal("the first position should be untouched by an opposite fill")
	}
}

// Opening reserves margin: one lot of EURUSD at 1:100 takes 1000.
func TestOpeningReservesMargin(t *testing.T) {
	h, e, r := newEngine(), account(), rules(true)

	f := h.Execute(e, order(true, 1), r, 1.1000, 1)

	if math.Abs(f.Opened.Margin-1000) > 1e-6 {
		t.Fatalf("margin = %v, want 1000", f.Opened.Margin)
	}
}

// A sell makes money when the price falls.
func TestSellProfitsWhenPriceFalls(t *testing.T) {
	h, e, r := newEngine(), account(), rules(true)

	place(e, h.Execute(e, order(false, 1), r, 1.1010, 1), 1)
	f := h.Execute(e, order(true, 1), r, 1.1000, 2)

	if math.Abs(f.Profit-100) > 1e-6 {
		t.Fatalf("profit = %v, want 100", f.Profit)
	}
}

// Closing a position must credit the profit once. The position carries the floating value the
// last tick worked out, and the fill carries the same money as realised profit; adding both
// would pay the client twice.
func TestClosingCreditsProfitOnce(t *testing.T) {
	h, e, r := newEngine(), account(), rules(true)
	opening := e.Account.Balance

	place(e, h.Execute(e, order(true, 1), r, 1.1000, 1), 1)

	// a tick has since revalued the position, exactly as the live engine does
	e.Positions[1].Profit = 190

	f := h.Execute(e, order(false, 1), r, 1.1019, 2)
	h.settle(e, order(false, 1), f, r)

	// 19 points on one lot is 190, and it should land once
	if got := e.Account.Balance - opening; math.Abs(got-190) > 1e-6 {
		t.Fatalf("balance moved by %v, want 190 — the profit was counted %.1f times", got, got/190)
	}
}

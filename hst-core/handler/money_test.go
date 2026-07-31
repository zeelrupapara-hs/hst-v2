package handler

import (
	"math"
	"testing"

	"hstcore/internal/settings"
	"hstcore/model"
)

func forex() *settings.Rules {
	return &settings.Rules{
		CalcMode:     model.CalcForex,
		ContractSize: 100000,
		Digits:       5,
		Point:        0.00001,
	}
}

// The worked example from margin_formula.htm: one lot of EURUSD at 1:100 with a contract size
// of 100 000 reserves 1 000 of the margin currency.
func TestMarginForexMatchesTheDocumentedExample(t *testing.T) {
	got := MarginFor(forex(), 1, 1.1, 100, 1)

	if math.Abs(got-1000) > 0.0001 {
		t.Fatalf("margin = %v, want 1000", got)
	}
}

// Without leverage the whole notional is reserved.
func TestMarginForexNoLeverage(t *testing.T) {
	r := forex()
	r.CalcMode = model.CalcForexNoLeverage

	if got := MarginFor(r, 1, 1.1, 100, 1); math.Abs(got-100000) > 0.0001 {
		t.Fatalf("margin = %v, want 100000", got)
	}
}

// A CFD reserves volume times contract size times the market price. The doc's example: one lot
// of oil, contract size 100, ask 80, reserves 8 000.
func TestMarginCFDUsesThePrice(t *testing.T) {
	r := forex()
	r.CalcMode = model.CalcCFD
	r.ContractSize = 100

	if got := MarginFor(r, 1, 80, 100, 1); math.Abs(got-8000) > 0.0001 {
		t.Fatalf("margin = %v, want 8000", got)
	}
}

// The margin currency is converted to the deposit currency at the given rate.
func TestMarginConvertsToDepositCurrency(t *testing.T) {
	if got := MarginFor(forex(), 1, 1.1, 100, 1.2790); math.Abs(got-1279) > 0.0001 {
		t.Fatalf("margin = %v, want 1279", got)
	}
}

// A buy makes money when the price rises; a sell when it falls. One lot of EURUSD moving 100
// points is 100 of the profit currency.
func TestProfitFollowsTheSide(t *testing.T) {
	r := forex()

	if got := ProfitFor(r, true, 1, 1.1000, 1.1010, 1); math.Abs(got-100) > 0.0001 {
		t.Fatalf("buy profit = %v, want 100", got)
	}
	if got := ProfitFor(r, false, 1, 1.1000, 1.1010, 1); math.Abs(got+100) > 0.0001 {
		t.Fatalf("sell profit = %v, want -100", got)
	}
}

// Equity is the balance plus what is floating; the margin level is the ratio the stop out reads.
func TestSettleAddsUpTheAccount(t *testing.T) {
	a := &model.Account{Balance: 10000}
	positions := map[int64]*model.Position{
		1: {Profit: -500, Margin: 1000},
		2: {Profit: 250, Margin: 1000},
	}

	m := Settle(a, positions, false)

	if math.Abs(m.Equity-9750) > 0.0001 {
		t.Fatalf("equity = %v, want 9750", m.Equity)
	}
	if math.Abs(m.Margin-2000) > 0.0001 {
		t.Fatalf("margin = %v, want 2000", m.Margin)
	}
	if math.Abs(m.MarginLevel-487.5) > 0.0001 {
		t.Fatalf("margin level = %v, want 487.5", m.MarginLevel)
	}
	if math.Abs(m.FreeMargin-7750) > 0.0001 {
		t.Fatalf("free margin = %v, want 7750", m.FreeMargin)
	}
}

// With nothing open there is no margin, so there is no level to compare against a stop out.
func TestSettleWithNothingOpen(t *testing.T) {
	m := Settle(&model.Account{Balance: 500}, nil, false)

	if m.MarginLevel != 0 {
		t.Fatalf("margin level = %v, want 0 when no margin is used", m.MarginLevel)
	}
	if math.Abs(m.Equity-500) > 0.0001 {
		t.Fatalf("equity = %v, want 500", m.Equity)
	}
}

// A group that forbids using unrealised profit as margin must ignore a gain but still feel a
// loss, otherwise a floating profit would let the account open more than it can afford.
func TestFreeMarginCanIgnoreFloatingProfit(t *testing.T) {
	a := &model.Account{Balance: 10000}
	gain := map[int64]*model.Position{1: {Profit: 500, Margin: 1000}}
	loss := map[int64]*model.Position{1: {Profit: -500, Margin: 1000}}

	if got := Settle(a, gain, true).FreeMargin; math.Abs(got-9000) > 0.0001 {
		t.Fatalf("free margin with an ignored gain = %v, want 9000", got)
	}
	if got := Settle(a, loss, true).FreeMargin; math.Abs(got-8500) > 0.0001 {
		t.Fatalf("free margin with a loss = %v, want 8500", got)
	}
}

func TestNormalisePriceRoundsToSymbolDigits(t *testing.T) {
	if got := NormalisePrice(1.234567, 5); math.Abs(got-1.23457) > 1e-9 {
		t.Fatalf("price = %v, want 1.23457", got)
	}
	if got := NormalisePrice(1.5, 0); math.Abs(got-2) > 1e-9 {
		t.Fatalf("price = %v, want 2", got)
	}
}

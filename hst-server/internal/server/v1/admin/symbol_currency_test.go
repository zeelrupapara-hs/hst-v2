package admin

import (
	"testing"

	"hstserver/model"
)

func TestParseForexPair(t *testing.T) {
	for _, tc := range []struct {
		symbol string
		ok     bool
		base   string
		profit string
	}{
		{"EURUSD", true, "EUR", "USD"},
		{"USDJPY", true, "USD", "JPY"},
		{"USDHUF.1", true, "USD", "HUF"},
		{"XAUUSD", true, "XAU", "USD"},
		{"EU", false, "", ""},
		{"TOOLONG", false, "", ""},
	} {
		pair, ok := parseForexPair(tc.symbol)
		if ok != tc.ok {
			t.Fatalf("parseForexPair(%q) ok=%v want %v", tc.symbol, ok, tc.ok)
		}
		if !ok {
			continue
		}
		if pair.Base != tc.base || pair.Profit != tc.profit {
			t.Fatalf("parseForexPair(%q) = %+v want base=%s profit=%s", tc.symbol, pair, tc.base, tc.profit)
		}
	}
}

func TestCurrencyDigits(t *testing.T) {
	if CurrencyDigits("JPY") != 0 {
		t.Fatalf("JPY digits want 0")
	}
	if CurrencyDigits("USD") != 2 {
		t.Fatalf("USD digits want 2")
	}
	if CurrencyDigits("BHD") != 3 {
		t.Fatalf("BHD digits want 3")
	}
}

func TestDeriveSymbolCurrenciesForex(t *testing.T) {
	d := deriveSymbolCurrencies("EURUSD", model.CalcMode_forex, model.Symbol{})
	if !d.OK {
		t.Fatal("expected forex derivation")
	}
	if d.CurrencyBase != "EUR" || d.CurrencyProfit != "USD" || d.CurrencyMargin != "EUR" {
		t.Fatalf("currencies = base=%s profit=%s margin=%s", d.CurrencyBase, d.CurrencyProfit, d.CurrencyMargin)
	}
	if d.Digits != 5 {
		t.Fatalf("digits = %d want 5", d.Digits)
	}
	if d.CurrencyProfitDigits != 2 || d.CurrencyBaseDigits != 2 {
		t.Fatalf("digit fields = base %d profit %d", d.CurrencyBaseDigits, d.CurrencyProfitDigits)
	}
}

func TestDeriveSymbolCurrenciesForexJPY(t *testing.T) {
	d := deriveSymbolCurrencies("USDJPY", model.CalcMode_forex, model.Symbol{})
	if !d.OK || d.Digits != 3 || d.CurrencyProfitDigits != 0 {
		t.Fatalf("USDJPY derivation = %+v", d)
	}
}

func TestDeriveSymbolCurrenciesForexSuffix(t *testing.T) {
	d := deriveSymbolCurrencies("USDHUF.1", model.CalcMode_forex_no_leverage, model.Symbol{})
	if !d.OK || d.CurrencyProfit != "HUF" || d.Digits != 3 {
		t.Fatalf("USDHUF.1 derivation = %+v", d)
	}
}

func TestDeriveSymbolCurrenciesCFD(t *testing.T) {
	d := deriveSymbolCurrencies("XAUUSD", model.CalcMode_cfd, model.Symbol{CurrencyProfit: "USD"})
	if !d.OK {
		t.Fatal("expected CFD derivation")
	}
	if d.CurrencyBase != "XAU" || d.CurrencyMargin != "USD" || d.CurrencyProfit != "USD" {
		t.Fatalf("CFD currencies = %+v", d)
	}
}

func TestDeriveSymbolCurrenciesFuturesUnchanged(t *testing.T) {
	d := deriveSymbolCurrencies("CLZ5", model.CalcMode_futures, model.Symbol{})
	if d.OK {
		t.Fatalf("futures should not auto-derive, got %+v", d)
	}
}

func TestApplyDerivedCurrencies(t *testing.T) {
	sym := model.Symbol{
		Symbol:   "GBPUSD",
		CalcMode: model.CalcMode_forex,
	}
	applyDerivedCurrencies(&sym)
	if sym.CurrencyBase != "GBP" || sym.CurrencyProfit != "USD" || sym.CurrencyMargin != "GBP" {
		t.Fatalf("applyDerivedCurrencies = %+v", sym)
	}
	if sym.Digits != 5 {
		t.Fatalf("digits = %d", sym.Digits)
	}
}

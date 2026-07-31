package translate_test

import (
	"testing"
	"time"

	"hstquote/internal/provider"
	"hstquote/internal/translate"
	"hstquote/model"
)

func TestApplyMarkup(t *testing.T) {
	feed := model.QuoteFeed{
		Datafeed: model.Datafeed{DatafeedID: 1},
		Translates: []model.DatafeedTranslate{{
			SymbolID: 42, Symbol: "EURUSD", Source: "EUR/USD", Digits: 5, BidMarkup: 1, AskMarkup: -1,
		}},
	}

	raw := provider.RawTick{
		SourceSymbol: "EUR/USD",
		Bid:          1.10000,
		Ask:          1.10020,
		Time:         time.Now().UTC(),
	}

	tick, ok := translate.ApplyMarkup(feed, raw)
	if !ok {
		t.Fatal("expected match")
	}
	if tick.SymbolID != 42 {
		t.Fatalf("symbol_id: %d", tick.SymbolID)
	}
	if tick.Symbol != "EURUSD" {
		t.Fatalf("symbol: %q", tick.Symbol)
	}
	if tick.Bid != 1.10001 {
		t.Fatalf("bid: %f", tick.Bid)
	}
	if tick.Ask != 1.10019 {
		t.Fatalf("ask: %f", tick.Ask)
	}
}

func TestExternalSymbolFallback(t *testing.T) {
	tr := model.DatafeedTranslate{Symbol: "GBPUSD", Source: ""}
	if tr.ExternalSymbol() != "GBPUSD" {
		t.Fatalf("got %q", tr.ExternalSymbol())
	}
}

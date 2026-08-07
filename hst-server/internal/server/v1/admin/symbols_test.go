package admin

import (
	"encoding/json"
	"hstserver/model"
	"testing"
)

func TestValidateSessionsOvernightWindows(t *testing.T) {
	sessions := []CrtSymbolSession{
		{Type: 0, Day: 2, Open: 710, Close: 1440},
		{Type: 0, Day: 2, Open: 10, Close: 470},
		{Type: 1, Day: 2, Open: 5, Close: 1440},
	}
	if err := validateSessions(sessions); err != nil {
		t.Fatalf("expected overnight windows to pass validation, got: %v", err)
	}
}

func TestValidateSessionsOverlapRejected(t *testing.T) {
	sessions := []CrtSymbolSession{
		{Type: 0, Day: 1, Open: 100, Close: 500},
		{Type: 0, Day: 1, Open: 400, Close: 800},
	}
	if err := validateSessions(sessions); err == nil {
		t.Fatal("expected overlapping sessions to fail validation")
	}
}

func TestValidateSessionsOpenMustBeLessThanClose(t *testing.T) {
	sessions := []CrtSymbolSession{
		{Type: 0, Day: 0, Open: 600, Close: 600},
	}
	if err := validateSessions(sessions); err == nil {
		t.Fatal("expected open >= close to fail validation")
	}
}

func TestPrepareCreateSymbolDefaults(t *testing.T) {
	body := CrtSymbol{
		Symbol:         "EURUSD",
		Path:           "Forex",
		CurrencyBase:   "EUR",
		CurrencyProfit: "USD",
		CurrencyMargin: "USD",
		TimeStart:      1_700_000_000,
	}
	sym := prepareCreateSymbol(body, `Forex\EURUSD`, 999)
	if sym.TradeMode != model.TradeMode_full {
		t.Fatalf("trade_mode default: got %v want full", sym.TradeMode)
	}
	if sym.ExecMode != model.ExecMode_market {
		t.Fatalf("exec_mode default: got %v want market", sym.ExecMode)
	}
	if sym.TimeStart != 1_700_000_000 {
		t.Fatalf("time_start: got %d want seconds passthrough", sym.TimeStart)
	}
	if sym.MarginInitialBuy != 1 || sym.MarginInitialSell != 1 {
		t.Fatalf("margin buy/sell defaults: got %v %v", sym.MarginInitialBuy, sym.MarginInitialSell)
	}
	if sym.SwapYearDay != 360 {
		t.Fatalf("swap_year_day default: got %d", sym.SwapYearDay)
	}
	if len(symbolInsertArgs(sym)) != 122 {
		t.Fatalf("insert arg count: got %d want 122", len(symbolInsertArgs(sym)))
	}
}

func TestValidateMarketDepthSpreadAllowsWhenDepthOff(t *testing.T) {
	before := map[string]json.RawMessage{
		"tick_book_depth": json.RawMessage(`0`),
		"spread":          json.RawMessage(`16`),
		"spread_balance":  json.RawMessage(`0`),
	}
	spread := int32(20)
	body := UptSymbol{Spread: &spread}
	if err := validateMarketDepthSpread(before, &body); err != nil {
		t.Fatalf("expected spread change when depth off, got: %v", err)
	}
}

func TestValidateMarketDepthSpreadRejectsWhenDepthOn(t *testing.T) {
	before := map[string]json.RawMessage{
		"tick_book_depth": json.RawMessage(`16`),
		"spread":          json.RawMessage(`0`),
		"spread_balance":  json.RawMessage(`0`),
	}
	spread := int32(10)
	body := UptSymbol{Spread: &spread}
	if err := validateMarketDepthSpread(before, &body); err == nil {
		t.Fatal("expected spread patch to fail when market depth enabled")
	}
}

func TestValidateMarketDepthSpreadAllowsDepthChangeAlone(t *testing.T) {
	before := map[string]json.RawMessage{
		"tick_book_depth": json.RawMessage(`0`),
		"spread":          json.RawMessage(`0`),
		"spread_balance":  json.RawMessage(`0`),
	}
	depth := int32(16)
	body := UptSymbol{TickBookDepth: &depth}
	if err := validateMarketDepthSpread(before, &body); err != nil {
		t.Fatalf("expected enabling market depth without spread change, got: %v", err)
	}
}

func TestValidateMarketDepthSpreadAllowsSpreadWhenDisablingDepth(t *testing.T) {
	before := map[string]json.RawMessage{
		"tick_book_depth": json.RawMessage(`16`),
		"spread":          json.RawMessage(`0`),
		"spread_balance":  json.RawMessage(`0`),
	}
	depth := int32(0)
	spread := int32(12)
	body := UptSymbol{TickBookDepth: &depth, Spread: &spread}
	if err := validateMarketDepthSpread(before, &body); err != nil {
		t.Fatalf("expected spread change when disabling market depth in same patch, got: %v", err)
	}
}

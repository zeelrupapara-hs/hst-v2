package admin

import (
	"encoding/json"
	"regexp"
	"strings"

	"hstserver/model"
)

var (
	forexPairRE     = regexp.MustCompile(`^([A-Z]{3})([A-Z]{3})([._#&][A-Z0-9._#&]*)?$`)
	smallDigitQuote = map[string]bool{"JPY": true, "HUF": true}
	currencyDigits  = map[string]int32{
		"JPY": 0, "HUF": 0, "KRW": 0, "VND": 0, "CLP": 0, "ISK": 0, "TWD": 0,
		"BHD": 3, "JOD": 3, "KWD": 3, "OMR": 3, "TND": 3,
	}
)

type forexPair struct {
	Base   string
	Profit string
}

func isForexCalcMode(mode model.CalcMode) bool {
	return mode == model.CalcMode_forex || mode == model.CalcMode_forex_no_leverage
}

func isCFDCalcMode(mode model.CalcMode) bool {
	return mode == model.CalcMode_cfd || mode == model.CalcMode_cfd_leverage
}

func CurrencyDigits(code string) int32 {
	code = strings.TrimSpace(strings.ToUpper(code))
	if code == "" {
		return 2
	}
	if d, ok := currencyDigits[code]; ok {
		return d
	}
	return 2
}

func parseForexPair(symbol string) (forexPair, bool) {
	m := forexPairRE.FindStringSubmatch(strings.TrimSpace(strings.ToUpper(symbol)))
	if m == nil {
		return forexPair{}, false
	}
	return forexPair{Base: m[1], Profit: m[2]}, true
}

// derivedCurrencies holds auto-filled currency fields for a symbol.
type derivedCurrencies struct {
	CurrencyBase         string
	CurrencyBaseDigits   int32
	CurrencyProfit       string
	CurrencyProfitDigits int32
	CurrencyMargin       string
	CurrencyMarginDigits int32
	Digits               int32
	OK                   bool
}

func deriveSymbolCurrencies(symbol string, calcMode model.CalcMode, cur model.Symbol) derivedCurrencies {
	if isForexCalcMode(calcMode) {
		pair, ok := parseForexPair(symbol)
		if !ok {
			return derivedCurrencies{}
		}
		digits := int32(5)
		if smallDigitQuote[pair.Profit] {
			digits = 3
		}
		return derivedCurrencies{
			CurrencyBase:         pair.Base,
			CurrencyBaseDigits:   CurrencyDigits(pair.Base),
			CurrencyProfit:       pair.Profit,
			CurrencyProfitDigits: CurrencyDigits(pair.Profit),
			CurrencyMargin:       pair.Base,
			CurrencyMarginDigits: CurrencyDigits(pair.Base),
			Digits:               digits,
			OK:                   true,
		}
	}

	if isCFDCalcMode(calcMode) {
		profit := strings.TrimSpace(cur.CurrencyProfit)
		if profit == "" {
			profit = "USD"
		}
		out := derivedCurrencies{
			CurrencyProfit:       profit,
			CurrencyProfitDigits: CurrencyDigits(profit),
			CurrencyMargin:       profit,
			CurrencyMarginDigits: CurrencyDigits(profit),
			OK:                   true,
		}
		if pair, ok := parseForexPair(symbol); ok && strings.TrimSpace(cur.CurrencyBase) == "" {
			out.CurrencyBase = pair.Base
			out.CurrencyBaseDigits = CurrencyDigits(pair.Base)
		} else if base := strings.TrimSpace(cur.CurrencyBase); base != "" {
			out.CurrencyBase = base
			out.CurrencyBaseDigits = CurrencyDigits(base)
		}
		return out
	}

	return derivedCurrencies{}
}

func applyDerivedCurrencies(sym *model.Symbol) {
	if sym == nil {
		return
	}
	d := deriveSymbolCurrencies(sym.Symbol, sym.CalcMode, *sym)
	if !d.OK {
		return
	}
	sym.CurrencyBase = d.CurrencyBase
	sym.CurrencyBaseDigits = d.CurrencyBaseDigits
	sym.CurrencyProfit = d.CurrencyProfit
	sym.CurrencyProfitDigits = d.CurrencyProfitDigits
	sym.CurrencyMargin = d.CurrencyMargin
	sym.CurrencyMarginDigits = d.CurrencyMarginDigits
	if isForexCalcMode(sym.CalcMode) {
		sym.Digits = d.Digits
		point, multiply := pointMultiply(d.Digits)
		sym.Point = point
		sym.Multiply = multiply
	}
}

// applyDerivedCurrenciesToPatch fills currency fields on an update body when the client did not
// send them explicitly and symbol or calc_mode is changing.
func applyDerivedCurrenciesToPatch(before map[string]json.RawMessage, body *UptSymbol) {
	if body == nil {
		return
	}
	if body.Symbol == nil && body.CalcMode == nil {
		return
	}

	symbol := jsonString(before["symbol"])
	if body.Symbol != nil {
		symbol = *body.Symbol
	}
	var calcMode model.CalcMode
	_ = json.Unmarshal(before["calc_mode"], &calcMode)
	if body.CalcMode != nil {
		calcMode = *body.CalcMode
	}

	cur := model.Symbol{
		Symbol:         symbol,
		CalcMode:       calcMode,
		CurrencyBase:   jsonString(before["currency_base"]),
		CurrencyProfit: jsonString(before["currency_profit"]),
		CurrencyMargin: jsonString(before["currency_margin"]),
	}
	d := deriveSymbolCurrencies(symbol, calcMode, cur)
	if !d.OK {
		return
	}

	if body.CurrencyBase == nil {
		body.CurrencyBase = &d.CurrencyBase
	}
	if body.CurrencyBaseDigits == nil {
		body.CurrencyBaseDigits = &d.CurrencyBaseDigits
	}
	if body.CurrencyProfit == nil {
		body.CurrencyProfit = &d.CurrencyProfit
	}
	if body.CurrencyProfitDigits == nil {
		body.CurrencyProfitDigits = &d.CurrencyProfitDigits
	}
	if body.CurrencyMargin == nil {
		body.CurrencyMargin = &d.CurrencyMargin
	}
	if body.CurrencyMarginDigits == nil {
		body.CurrencyMarginDigits = &d.CurrencyMarginDigits
	}
	if isForexCalcMode(calcMode) && body.Digits == nil {
		body.Digits = &d.Digits
	}
}

package handler

import (
	"time"

	"hstcore/internal/settings"
	"hstcore/model"
)

const (
	SessionQuote = 0
	SessionTrade = 1
)

func (h *Handler) IsMarketOpen(r *settings.Rules, dir model.Direction) bool {
	if r.TradeMode == model.TradeDisabled {
		return false
	}
	if dir == model.DirectionOut {
		return true
	}

	now := time.Now().UTC()

	if h.IsHoliday(r, now) {
		return false
	}

	return h.InSession(r, now, SessionTrade)
}

func (h *Handler) InSession(r *settings.Rules, at time.Time, kind int16) bool {
	// #nosec G115 -- a weekday is 0..6
	windows := h.Sessions.For(r.Symbol.SymbolId, int16(at.Weekday()), kind)
	if len(windows) == 0 {
		return true
	}

	// #nosec G115 -- a minute of the day is 0..1439
	minutes := int32(at.Hour()*60 + at.Minute())

	for _, w := range windows {
		if minutes >= w.Open && minutes < w.Close {
			return true
		}
	}

	return false
}

func (h *Handler) IsHoliday(r *settings.Rules, at time.Time) bool {
	return h.Holidays.Covers(r.Symbol.Path, r.Symbol.Symbol, at)
}

func (h *Handler) ConvertCurrency(amount float64, from, to string, buy bool) float64 {
	if amount == 0 || from == "" || from == to {
		return amount
	}

	rate, ok := h.CrossRate(from, to, buy)
	if !ok {
		return amount
	}

	return amount * rate
}

func (h *Handler) CrossRate(from, to string, buy bool) (float64, bool) {
	if from == to {
		return 1, true
	}

	if t, ok := h.Quotes.Get(from + to); ok && t.Ok() {
		return t.OpenPrice(buy), true
	}

	if t, ok := h.Quotes.Get(to + from); ok && t.Ok() {
		if p := t.OpenPrice(buy); p > 0 {
			return 1 / p, true
		}
	}

	return 0, false
}

func (h *Handler) RateProfit(r *settings.Rules, a *model.Account, buy bool) float64 {
	rate, ok := h.CrossRate(r.CurrencyProfit, a.Currency, buy)
	if !ok {
		return 1
	}
	return rate
}

func (h *Handler) RateMargin(r *settings.Rules, a *model.Account, buy bool) float64 {
	rate, ok := h.CrossRate(r.CurrencyMargin, a.Currency, buy)
	if !ok {
		return 1
	}
	return rate
}

func GetPrice(t model.Tick, buy bool, dir model.Direction) float64 {
	if dir == model.DirectionOut {
		return t.ClosePrice(buy)
	}
	return t.OpenPrice(buy)
}

func GetStopLevel(point float64, level int32) float64 { return float64(level) * point }

func CheckPendingPriceHit(price, pending float64, kind model.OrderType) bool {
	switch kind {
	case model.OrderBuyLimit:
		return price <= pending
	case model.OrderSellLimit:
		return price >= pending
	case model.OrderBuyStop, model.OrderBuyStopLimit:
		return price >= pending
	case model.OrderSellStop, model.OrderSellStopLimit:
		return price <= pending
	}
	return false
}

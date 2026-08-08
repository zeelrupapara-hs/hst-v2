package handler

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"hstcore/internal/settings"
	"hstcore/model"
)

const (
	SessionQuote = 0
	SessionTrade = 1
)

func (h *Handler) IsMarketOpen(r *settings.Rules, dir model.Direction) bool {
	if r.TradeMode == model.TradeMode_disabled {
		return false
	}
	if dir == model.Direction_out {
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

func (h *Handler) CrossRate(from, to string, buy bool) (float64, bool) {
	return h.crossRate("", from, to, buy)
}

// crossRate prices the conversion leg; with a group it charges that group's own spread on the
// conversion pair, so a marked-up instrument converts at marked-up prices too.
func (h *Handler) crossRate(group, from, to string, buy bool) (float64, bool) {
	if from == to {
		return 1, true
	}

	if t, ok := h.quoteForPair(group, from+to); ok && t.Ok() {
		return t.OpenPrice(buy), true
	}

	if t, ok := h.quoteForPair(group, to+from); ok && t.Ok() {
		if p := t.OpenPrice(buy); p > 0 {
			return 1 / p, true
		}
	}

	return 0, false
}

func (h *Handler) quoteForPair(group, pair string) (model.Tick, bool) {
	t, ok := h.Quotes.Get(pair)
	if !ok {
		return t, false
	}
	if group == "" {
		return t, true
	}
	if r, ok := h.Settings.For(group, pair); ok {
		return CalculateAccountSpread(r, t), true
	}
	return t, true
}

func (h *Handler) RateProfit(r *settings.Rules, a *model.Account, buy bool) float64 {
	rate, ok := h.crossRate(a.Group, r.CurrencyProfit, a.Currency, buy)
	if !ok {
		return 1
	}
	return rate
}

func (h *Handler) RateMargin(r *settings.Rules, a *model.Account, buy bool) float64 {
	rate, ok := h.crossRate(a.Group, r.CurrencyMargin, a.Currency, buy)
	if !ok {
		return 1
	}
	return rate
}

// AccountSummary is where an account stands, as one line.
//
//	summary,login,balance,credit,equity,margin,free,level%,pl[,positionId,profit]...
//
// The trailing pairs are the positions on the instrument that just moved; every other position's
// result is already counted in pl.
func AccountSummary(a *model.Account, positions map[int64]float64) string {
	d := int(a.CurrencyDigits)
	if d < 0 {
		d = 2
	}

	var b strings.Builder

	b.WriteString("summary,")
	b.WriteString(strconv.FormatInt(a.Login, 10))

	for _, v := range []float64{a.Balance, a.Credit, a.Equity, a.Margin, a.MarginFree} {
		b.WriteByte(',')
		b.WriteString(strconv.FormatFloat(v, 'f', d, 64))
	}

	b.WriteByte(',')
	b.WriteString(strconv.FormatFloat(a.MarginLevel, 'f', 2, 64))
	b.WriteString("%,")
	b.WriteString(strconv.FormatFloat(a.Profit, 'f', d, 64))

	// sorted, so the same account state is always the same line
	ids := make([]int64, 0, len(positions))
	for id := range positions {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	for _, id := range ids {
		b.WriteByte(',')
		b.WriteString(strconv.FormatInt(id, 10))
		b.WriteByte(',')
		b.WriteString(strconv.FormatFloat(positions[id], 'f', d, 64))
	}

	return b.String()
}

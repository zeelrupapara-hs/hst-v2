package handler

import (
	"context"
	"time"

	"hstcore/internal/book"
	"hstcore/internal/settings"
	"hstcore/model"
	"hstcore/pkg/logger"
)

const (
	SwapDisabled       = 0
	SwapPoints         = 1
	SwapSymbolCurrency = 2
	SwapMarginCurrency = 3
	SwapDepositRate    = 4
	SwapInterestOpen   = 5
	SwapInterestCurr   = 6
	SwapPercentCurrent = 7
	SwapPercentOpen    = 8
)

// ChangeEndOfDayDate moves when the rollover runs. Every pod holds the same answer, so the
// change reaches all of them at once and is read back from the setting on the next boot.
func (h *Handler) ChangeEndOfDayDate(at string) {
	t, err := time.Parse(endOfDayLayout, at)
	if err != nil {
		h.Log.Log(logger.TypeSys, logger.CodeErr, "could not read the end of day time",
			"at", at, "error", err.Error())
		return
	}

	h.mu.Lock()
	// the rollover runs on the last second of the named minute
	h.endOfDay = t.Add(59 * time.Second)
	h.mu.Unlock()

	h.Log.Log(logger.TypeSys, logger.CodeOK, "end of day time changed",
		"at", h.EndOfDayAt().Format("15:04:05"))
}

// endOfDayLayout is how the time is written down, both on the wire and in the setting.
const endOfDayLayout = "15:04"

func (h *Handler) EndOfDayAt() time.Time {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.endOfDay
}

func (h *Handler) RunEndOfDay(ctx context.Context) {
	for {
		wait := h.untilEndOfDay(time.Now())

		h.Log.Log(logger.TypeSys, logger.CodeOK, "next end of day",
			"in", wait.Round(time.Minute).String(), "at", h.EndOfDayAt().Format("15:04:05"))

		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}

		h.EndOfDayProcess(ctx)
	}
}

func (h *Handler) EndOfDayProcess(ctx context.Context) {
	started := time.Now()

	h.Log.Log(logger.TypeSys, logger.CodeOK, "end of day started",
		"at", started.Format(time.DateTime))

	now := time.Now()

	h.CheckPendingOrdersExpiration(ctx)
	h.SwapsJob(ctx)
	h.ReleaseAccumulatedProfit(ctx)
	h.ChargeDailyCommissions(ctx, now)
	h.ChargeMonthlyCommissions(ctx, now)

	h.Log.Log(logger.TypeSys, logger.CodeOK, "end of day finished",
		"took_ms", time.Since(started).Milliseconds())
}

func (h *Handler) CheckPendingOrdersExpiration(ctx context.Context) {
	var swept int

	h.Accounts.Each(func(e *book.Entry) {
		for _, symbol := range e.Symbols() {
			h.ExpireOrders(ctx, e, symbol)
			swept++
		}
	})

	h.Log.Log(logger.TypeTrade, logger.CodeOK, "pending orders swept", "accounts", swept)
}

func (h *Handler) SwapsJob(ctx context.Context) {
	var accounts, changed int

	h.Accounts.Each(func(e *book.Entry) {
		accounts++
		if h.chargeAccountSwaps(ctx, e) {
			changed++
		}
	})

	h.Log.Log(logger.TypeTrade, logger.CodeOK, "swaps charged",
		"accounts", accounts, "changed", changed)
}

func (h *Handler) chargeAccountSwaps(ctx context.Context, e *book.Entry) bool {
	e.Lock()

	group := e.Account.Group

	// a group that does not charge swaps rolls its positions over for free
	if g, ok := h.Settings.Group(group); ok && g.TradeFlags&model.TradeFlagSwaps == 0 {
		e.Unlock()
		return false
	}

	var touched []*model.Position

	for _, p := range e.Positions {
		r, ok := h.Settings.For(group, p.Symbol)
		if !ok {
			continue
		}

		swap := h.CalculateSwaps(r, p, e.Account)
		if swap == 0 {
			continue
		}

		p.Storage += swap
		p.TimeUpdate = Now()
		touched = append(touched, p)
	}

	if len(touched) == 0 {
		e.Unlock()
		return false
	}

	h.CalculateAccountMargins(e).Apply(e.Account)

	account := *e.Account
	e.Unlock()

	for _, p := range touched {
		h.SavePositionAndPublishAsync(p)
	}

	h.PublishAccount(&account, nil)

	return true
}

// ReleaseAccumulatedProfit hands a day's held profit back to the balance, for the groups that
// asked the server to do it rather than leaving it to a gateway.
func (h *Handler) ReleaseAccumulatedProfit(ctx context.Context) {
	var released int

	h.Accounts.Each(func(e *book.Entry) {
		g, ok := h.Settings.Group(e.Account.Group)
		if !ok || g.MarginFlags&model.GroupMarginFlagClearAccumulated == 0 {
			return
		}

		e.Lock()

		held := e.Account.BlockedProfit
		if held == 0 {
			e.Unlock()
			return
		}

		e.Account.Balance += held
		e.Account.BlockedProfit = 0
		h.CalculateAccountMargins(e).Apply(e.Account)

		account := *e.Account
		e.Unlock()

		released++

		if err := h.SaveAccount(ctx, &account); err != nil {
			h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not release a held profit",
				"login", account.Login, "error", err.Error())
		}

		h.PublishAccount(&account, nil)
	})

	if released > 0 {
		h.Log.Log(logger.TypeTrade, logger.CodeOK, "held profit released", "accounts", released)
	}
}

func (h *Handler) CalculateSwaps(r *settings.Rules, p *model.Position, a *model.Account) float64 {
	if r.SwapMode == SwapDisabled {
		return 0
	}

	rate := r.SwapLong
	if !p.Buy() {
		rate = r.SwapShort
	}
	if rate == 0 {
		return 0
	}

	now := time.Now()

	factor := r.SwapRate[int(now.Weekday())]
	if factor == 0 {
		return 0
	}

	// with holidays taken into account nothing is charged on the holiday itself, and the day
	// before carries the charge for both
	if r.SwapFlags&model.SwapFlagConsiderHolidays != 0 {
		if h.IsHoliday(r, now) {
			return 0
		}
		if h.IsHoliday(r, now.AddDate(0, 0, 1)) {
			factor *= 2
		}
	}

	lots := p.Lots()
	var swap float64

	switch r.SwapMode {
	case SwapPoints:
		if r.TickSize > 0 {
			swap = rate * factor * r.Point / r.TickSize * r.TickValue * lots
			break
		}
		swap = rate * factor * r.Point * r.ContractSize * lots

	case SwapSymbolCurrency, SwapMarginCurrency, SwapDepositRate:
		swap = rate * factor * lots

	case SwapPercentCurrent, SwapInterestCurr:
		swap = factor * (lots * r.ContractSize * p.PriceCurrent) * (rate / 100) / float64(r.SwapYearDay)

	case SwapPercentOpen, SwapInterestOpen:
		swap = factor * (lots * r.ContractSize * p.PriceOpen) * (rate / 100) / float64(r.SwapYearDay)
	}

	swap = h.ConvertCurrency(swap, r.CurrencyProfit, a.Currency, p.Buy())

	return NormalisePrice(swap, a.CurrencyDigits)
}

func (h *Handler) untilEndOfDay(now time.Time) time.Duration {
	at := h.EndOfDayAt()

	next := time.Date(now.Year(), now.Month(), now.Day(),
		at.Hour(), at.Minute(), at.Second(), 0, now.Location())
	if !next.After(now) {
		next = next.AddDate(0, 0, 1)
	}

	return next.Sub(now)
}

func defaultEndOfDay() time.Time {
	t, _ := time.Parse("15:04:05", "23:59:59")
	return t
}

package handler

import (
	"context"
	"time"

	"hstcore/internal/book"
	"hstcore/internal/settings"
	"hstcore/model"
	"hstcore/pkg/logger"
)

// Swaps: the cost of holding a position overnight.
//
// Once a day, every open position is charged or paid for being carried. What is charged depends
// on the instrument's swap mode — some quote it in points, some as a yearly percentage of the
// position's value.
//
// Wednesday is charged three times. The market settles two days after the trade, so a position
// held through Wednesday night is really being carried over the weekend, and the industry
// charges for those three days at once.

// How the swap figure on an instrument should be read.
const (
	swapDisabled       = 0
	swapPoints         = 1
	swapSymbolCurrency = 2
	swapMarginCurrency = 3
	swapDepositRate    = 4
	swapInterestOpen   = 5
	swapInterestCurr   = 6
	swapPercentCurrent = 7
	swapPercentOpen    = 8
)

// tripleSwapDay is the weekday charged three times.
const tripleSwapDay = time.Wednesday

// ChargeSwaps walks every account this pod holds and charges a day's carry.
//
// Called once a day by the scheduler. It runs over the whole book rather than one symbol, which
// is the one job that legitimately touches every account.
func (h *Handler) ChargeSwaps(ctx context.Context, on time.Time) {
	days := 1
	if on.Weekday() == tripleSwapDay {
		days = 3
	}

	var accounts, charged int

	h.Accounts.Each(func(e *book.Entry) {
		accounts++
		if h.chargeAccountSwaps(ctx, e, days) {
			charged++
		}
	})

	h.Log.Log(logger.TypeTrade, logger.CodeOK, "swaps charged",
		"accounts", accounts, "changed", charged, "days", days)
}

// chargeAccountSwaps charges one account and reports whether anything moved.
func (h *Handler) chargeAccountSwaps(ctx context.Context, e *book.Entry, days int) bool {
	e.Lock()

	group := e.Account.Group
	var touched []*model.Position

	for _, p := range e.Positions {
		r, ok := h.Settings.For(group, p.Symbol)
		if !ok {
			continue
		}

		swap := h.swapFor(r, p, days)
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

	// storage is part of equity, so the account has to be added up again
	if r, ok := h.Settings.For(group, touched[0].Symbol); ok {
		Settle(e.Account, e.Positions, r.Group.MarginFreeProfit != 0).Apply(e.Account)
	}

	account := *e.Account
	e.Unlock()

	for _, p := range touched {
		if _, err := h.DB.DB.Exec(ctx,
			`UPDATE hst.positions SET storage = $1, time_update = $2, date_modified = $2
			  WHERE position_id = $3`, p.Storage, p.TimeUpdate, p.PositionId); err != nil {
			h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not save a swap",
				"login", p.Login, "position", p.PositionId, "error", err.Error())
		}
	}

	if _, err := h.DB.DB.Exec(ctx,
		`UPDATE hst.accounts SET storage = $1, equity = $2, margin_level = $3, updated_at = $4
		  WHERE login = $5`,
		account.Storage, account.Equity, account.MarginLevel, Now(), account.Login); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not save an account after swaps",
			"login", account.Login, "error", err.Error())
	}

	h.publishAccount(&account)

	return true
}

// swapFor is a day's carry on one position, in the deposit currency. A negative number is a
// charge, a positive one is paid to the client.
func (h *Handler) swapFor(r *settings.Rules, p *model.Position, days int) float64 {
	rate := r.SwapLong
	if !p.Buy() {
		rate = r.SwapShort
	}

	if rate == 0 || r.SwapMode == swapDisabled {
		return 0
	}

	lots := p.Lots()

	switch r.SwapMode {
	case swapPoints:
		// the figure is in points of the instrument, worth a tick each
		if r.TickSize > 0 {
			return rate * r.Point / r.TickSize * r.TickValue * lots * float64(days)
		}
		return rate * r.Point * r.ContractSize * lots * float64(days)

	case swapSymbolCurrency, swapMarginCurrency, swapDepositRate:
		// a flat amount per lot per day
		return rate * lots * float64(days)

	case swapPercentCurrent, swapInterestCurr:
		// a yearly percentage of what the position is worth now
		value := lots * r.ContractSize * p.PriceCurrent
		return value * rate / 100 / 360 * float64(days)

	case swapPercentOpen, swapInterestOpen:
		// the same, against what it was worth when opened
		value := lots * r.ContractSize * p.PriceOpen
		return value * rate / 100 / 360 * float64(days)
	}

	return 0
}

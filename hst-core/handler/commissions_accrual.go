package handler

import (
	"context"
	"time"

	"hstcore/internal/book"
	"hstcore/model"
	"hstcore/pkg/logger"
)

// A daily or monthly commission cannot be worked out when the deal happens, because the tier it
// falls in depends on everything the account traded over the whole period. So the trades are
// added up when the period closes and one charge is written for each commission.

type turnover struct {
	Lots  float64
	Money float64
	Deals int
}

func (h *Handler) ChargeDailyCommissions(ctx context.Context, at time.Time) {
	from := time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, at.Location())

	h.chargePeriod(ctx, chargeDaily, int32(model.DealCommissionDaily),
		from.UnixNano(), at.UnixNano(), "daily commission")
}

// ChargeMonthlyCommissions runs on the last end of day of the month.
func (h *Handler) ChargeMonthlyCommissions(ctx context.Context, at time.Time) {
	if !isMonthEnd(at) {
		return
	}

	from := time.Date(at.Year(), at.Month(), 1, 0, 0, 0, 0, at.Location())

	h.chargePeriod(ctx, chargeMonthly, int32(model.DealCommissionMonthly),
		from.UnixNano(), at.UnixNano(), "monthly commission")
}

func isMonthEnd(at time.Time) bool { return at.AddDate(0, 0, 1).Month() != at.Month() }

func (h *Handler) chargePeriod(ctx context.Context, mode int32, action int32,
	from, to int64, what string) {
	traded, err := h.turnoverBetween(ctx, from, to)
	if err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not read the period's turnover",
			"what", what, "error", err.Error())
		return
	}

	if len(traded) == 0 {
		return
	}

	var charged, accounts int

	h.Accounts.Each(func(e *book.Entry) {
		bySymbol := traded[e.Account.Login]
		if len(bySymbol) == 0 {
			return
		}

		// a period that was already closed must not be charged twice, which is what a restart
		// between the write and the next run would otherwise cause
		already, err := h.chargedBetween(ctx, e.Account.Login, action, from, to)
		if err != nil || already {
			return
		}

		amount := h.periodCommission(e, bySymbol, mode)
		if amount == 0 {
			return
		}

		accounts++
		if h.applyPeriodCharge(ctx, e, action, -amount, what) {
			charged++
		}
	})

	if accounts > 0 {
		h.Log.Log(logger.TypeTrade, logger.CodeOK, "period commissions charged",
			"what", what, "accounts", accounts, "written", charged)
	}
}

// periodCommission is what one account owes for the period, over every commission of this
// charge mode that covers something it traded.
func (h *Handler) periodCommission(e *book.Entry, bySymbol map[string]*turnover, mode int32) float64 {
	group, ok := h.Settings.Group(e.Account.Group)
	if !ok {
		return 0
	}

	h.mu.RLock()
	list := h.commissions[group.GroupId]
	h.mu.RUnlock()

	var total float64

	for i := range list {
		c := &list[i]
		if c.ChargeMode != mode {
			continue
		}

		var band turnover

		for symbol, t := range bySymbol {
			r, ok := h.Settings.For(e.Account.Group, symbol)
			if !ok || !pathCovers(c.Path, r) {
				continue
			}

			band.Lots += t.Lots
			band.Money += t.Money
			band.Deals += t.Deals
		}

		if band.Deals == 0 {
			continue
		}

		total += c.periodCharge(band)
	}

	return total
}

// periodCharge picks the tier the period's total falls in and applies it once.
func (c *Commission) periodCharge(t turnover) float64 {
	measure := t.Lots
	if c.RangeMode == rangeTurnoverMoney {
		measure = t.Money
	}

	var tier *CommissionTier
	for i := range c.Tiers {
		b := &c.Tiers[i]
		if measure >= b.RangeFrom && (b.RangeTo <= 0 || measure < b.RangeTo) {
			tier = b
			break
		}
	}

	if tier == nil {
		return 0
	}

	amount := tier.Value
	if tier.Type == tierPerVolume {
		amount = tier.Value * t.Lots
	}

	if tier.Minimal > 0 && amount < tier.Minimal {
		amount = tier.Minimal
	}

	return amount
}

// turnoverBetween is what every account on this pod traded in the window, by symbol.
func (h *Handler) turnoverBetween(ctx context.Context, from, to int64) (map[int64]map[string]*turnover, error) {
	rows, err := h.DB.DB.Query(ctx,
		`SELECT login, symbol, SUM(volume), SUM(ABS(price * volume * contract_size)), count(*)
		   FROM hst.deals
		  WHERE time >= $1 AND time < $2 AND action IN ($3, $4) AND symbol <> ''
		  GROUP BY login, symbol`,
		from, to, int32(model.DealBuy), int32(model.DealSell))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[int64]map[string]*turnover, 1024)

	for rows.Next() {
		var login int64
		var symbol string
		var volume int64
		var money float64
		var deals int

		if err := rows.Scan(&login, &symbol, &volume, &money, &deals); err != nil {
			return nil, err
		}

		if out[login] == nil {
			out[login] = make(map[string]*turnover, 4)
		}
		out[login][symbol] = &turnover{
			Lots:  model.Lots(volume),
			Money: money / model.VolumeUnit,
			Deals: deals,
		}
	}

	return out, rows.Err()
}

// chargedBetween reports whether this period was already settled for the account.
func (h *Handler) chargedBetween(ctx context.Context, login int64, action int32,
	from, to int64) (bool, error) {
	var n int

	if err := h.DB.DB.QueryRow(ctx,
		`SELECT count(*) FROM hst.deals
		  WHERE login = $1 AND action = $2 AND time >= $3 AND time < $4`,
		login, action, from, to).Scan(&n); err != nil {
		return false, err
	}

	return n > 0, nil
}

// applyPeriodCharge takes the money and writes the deal that explains it.
func (h *Handler) applyPeriodCharge(ctx context.Context, e *book.Entry, action int32,
	amount float64, comment string) bool {
	e.Lock()

	e.Account.Balance += amount

	h.SettleAccount(e).Apply(e.Account)

	deal := &model.Deal{
		Login:          e.Account.Login,
		Action:         action,
		Entry:          int32(model.EntryIn),
		DigitsCurrency: e.Account.CurrencyDigits,
		Time:           Now(),
		Profit:         amount,
		Value:          amount,
		Commission:     amount,
		RateProfit:     1,
		RateMargin:     1,
		Comment:        comment,
		Reason:         int32(model.ReasonDealer),
	}

	account := *e.Account
	e.Unlock()

	if err := h.SaveBalanceAndPublish(ctx, e, deal, &account); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not charge a period commission",
			"login", account.Login, "error", err.Error())
		return false
	}

	return true
}

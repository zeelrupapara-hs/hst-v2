package handler

import (
	"context"

	"hstcore/internal/book"
	"hstcore/model"

	"github.com/jackc/pgx/v5"
)

// Writing a trade down.
//
// One transaction per trade: the order, its deals, the positions it opened, changed or closed,
// and the account's new balance all land together or not at all. A deal written without the
// balance that goes with it would leave the ledger disagreeing with the account.

// save writes everything a fill produced.
func (h *Handler) save(ctx context.Context, e *book.Entry, o *model.Order, f *Fill,
	a *model.Account) error {
	tx, err := h.DB.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	now := Now()

	if err := tx.QueryRow(ctx,
		`INSERT INTO hst.orders
		   (login, dealer, symbol, digits, digits_currency, contract_size, state, reason,
		    time_setup, time_expiration, time_done, type, type_fill, type_time,
		    price_order, price_trigger, price_current, price_sl, price_tp,
		    volume_initial, volume_current, expert_id, position_id, comment, rate_margin,
		    date_created, date_modified)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,
		         $20,$21,$22,$23,$24,$25,$26,$26)
		 RETURNING order_id`,
		o.Login, o.Dealer, o.Symbol, o.Digits, o.DigitsCurrency, o.ContractSize,
		o.State, o.Reason, o.TimeSetup, o.TimeExpiration, o.TimeDone, o.Type,
		o.TypeFill, o.TypeTime, o.PriceOrder, o.PriceTrigger, o.PriceCurrent,
		o.PriceSL, o.PriceTP, o.VolumeInitial, o.VolumeCurrent, o.ExpertId,
		o.PositionId, o.Comment, o.RateMargin, now).Scan(&o.OrderId); err != nil {
		return err
	}

	return h.writeFill(ctx, tx, e, o, f, a, now)
}

// saveFill writes what a fill produced against an order row that already exists. Used when a
// pending order activates: its ticket was written when the client placed it.
func (h *Handler) saveFill(ctx context.Context, e *book.Entry, o *model.Order, f *Fill,
	a *model.Account) error {
	tx, err := h.DB.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	return h.writeFill(ctx, tx, e, o, f, a, Now())
}

// writeFill is the shared half: positions, deals and the account.
func (h *Handler) writeFill(ctx context.Context, tx pgx.Tx, e *book.Entry, o *model.Order,
	f *Fill, a *model.Account, now int64) error {
	// the position first, so its real id can go on the deals that reference it
	if f.Opened != nil {
		temp := f.Opened.PositionId

		if err := tx.QueryRow(ctx,
			`INSERT INTO hst.positions
			   (login, dealer, symbol, action, digits, digits_currency, reason, contract_size,
			    time_create, time_update, price_open, price_current, price_sl, price_tp,
			    volume, profit, storage, rate_profit, rate_margin, expert_id, comment,
			    date_created, date_modified)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,
			         $20,$21,$22,$22)
			 RETURNING position_id`,
			f.Opened.Login, f.Opened.Dealer, f.Opened.Symbol, f.Opened.Action,
			f.Opened.Digits, f.Opened.DigitsCurrency, f.Opened.Reason, f.Opened.ContractSize,
			f.Opened.TimeCreate, f.Opened.TimeUpdate, f.Opened.PriceOpen, f.Opened.PriceCurrent,
			f.Opened.PriceSL, f.Opened.PriceTP, f.Opened.Volume, f.Opened.Profit,
			f.Opened.Storage, f.Opened.RateProfit, f.Opened.RateMargin, f.Opened.ExpertId,
			f.Opened.Comment, now).Scan(&f.Opened.PositionId); err != nil {
			return err
		}

		// swap the temporary key for the real one now the row exists
		e.Lock()
		delete(e.Positions, temp)
		e.Positions[f.Opened.PositionId] = f.Opened
		e.Unlock()

		h.Accounts.Watch(f.Opened.Symbol, e)
	}

	for _, p := range f.Changed {
		if _, err := tx.Exec(ctx,
			`UPDATE hst.positions
			    SET volume = $1, price_open = $2, price_current = $3, profit = $4,
			        storage = $5, time_update = $6, date_modified = $6
			  WHERE position_id = $7`,
			p.Volume, p.PriceOpen, p.PriceCurrent, p.Profit, p.Storage, now, p.PositionId); err != nil {
			return err
		}
	}

	for _, p := range f.Closed {
		if _, err := tx.Exec(ctx, `DELETE FROM hst.positions WHERE position_id = $1`,
			p.PositionId); err != nil {
			return err
		}
		h.unwatchIfLast(e, p.Symbol)
	}

	for _, d := range f.Deals {
		d.OrderId = o.OrderId
		if d.PositionId <= 0 && f.Opened != nil {
			d.PositionId = f.Opened.PositionId
		}

		if err := tx.QueryRow(ctx,
			`INSERT INTO hst.deals
			   (login, dealer, order_id, action, entry, digits, digits_currency, contract_size,
			    time, symbol, price, price_sl, price_tp, volume, volume_closed, profit, value,
			    storage, commission, fee, rate_profit, rate_margin, expert_id, position_id,
			    comment, profit_raw, price_position, tick_value, tick_size, reason,
			    market_bid, market_ask, date_created)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,
			         $21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33)
			 RETURNING deal_id`,
			d.Login, d.Dealer, d.OrderId, d.Action, d.Entry, d.Digits, d.DigitsCurrency,
			d.ContractSize, d.Time, d.Symbol, d.Price, d.PriceSL, d.PriceTP, d.Volume,
			d.VolumeClosed, d.Profit, d.Value, d.Storage, d.Commission, d.Fee,
			d.RateProfit, d.RateMargin, d.ExpertId, d.PositionId, d.Comment, d.ProfitRaw,
			d.PricePosition, d.TickValue, d.TickSize, d.Reason, d.MarketBid, d.MarketAsk,
			now).Scan(&d.DealId); err != nil {
			return err
		}
	}

	if err := saveAccount(ctx, tx, a, now); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// saveAccount writes the money back. hst.accounts is the truth; hst.users.balance is kept in
// step behind it so the API can list accounts without a join.
func saveAccount(ctx context.Context, tx pgx.Tx, a *model.Account, now int64) error {
	if _, err := tx.Exec(ctx,
		`UPDATE hst.accounts
		    SET balance = $1, credit = $2, margin = $3, margin_free = $4, margin_level = $5,
		        profit = $6, storage = $7, floating = $8, equity = $9, updated_at = $10
		  WHERE login = $11`,
		a.Balance, a.Credit, a.Margin, a.MarginFree, a.MarginLevel,
		a.Profit, a.Storage, a.Floating, a.Equity, now, a.Login); err != nil {
		return err
	}

	_, err := tx.Exec(ctx, `UPDATE hst.users SET balance = $1 WHERE login = $2`,
		a.Balance, a.Login)

	return err
}

// unwatchIfLast drops the symbol from the index once the account holds nothing on it.
func (h *Handler) unwatchIfLast(e *book.Entry, symbol string) {
	e.Lock()
	defer e.Unlock()

	for _, p := range e.Positions {
		if p.Symbol == symbol {
			return
		}
	}
	for _, o := range e.Orders {
		if o.Symbol == symbol {
			return
		}
	}

	h.Accounts.Unwatch(symbol, e.Account.Login)
}

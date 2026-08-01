package handler

import (
	"context"

	"hstcore/internal/book"
	"hstcore/model"

	"github.com/jackc/pgx/v5"
)

// One transaction per trade.

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
		    volume_initial, volume_initial_ext, volume_current, volume_current_ext,
		    expert_id, position_id, position_by_id, comment, rate_margin,
		    activation_mode, activation_time, activation_price, activation_flags,
		    routing_id, date_created, date_modified)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,
		         $20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$34)
		 RETURNING order_id`,
		o.Login, o.Dealer, o.Symbol, o.Digits, o.DigitsCurrency, o.ContractSize,
		o.State, o.Reason, o.TimeSetup, o.TimeExpiration, o.TimeDone, o.Type,
		o.TypeFill, o.TypeTime, o.PriceOrder, o.PriceTrigger, o.PriceCurrent,
		o.PriceSL, o.PriceTP, o.VolumeInitial, o.VolumeExt, o.VolumeCurrent, o.VolumeExt,
		o.ExpertId, o.PositionId, o.PositionById, o.Comment, o.RateMargin,
		o.ActivationMode, o.ActivationTime, o.ActivationPrice, o.ActivationFlags,
		o.RoutingId, now).Scan(&o.OrderId); err != nil {
		return err
	}

	return h.writeFill(ctx, tx, e, o, f, a, now)
}

// saveFill writes a fill against an order row that already exists.
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
	if f.Opened != nil {
		temp := f.Opened.PositionId

		if err := tx.QueryRow(ctx,
			`INSERT INTO hst.positions
			   (login, dealer, symbol, action, digits, digits_currency, reason, contract_size,
			    time_create, time_update, price_open, price_current, price_sl, price_tp,
			    volume, volume_ext, profit, storage, rate_profit, rate_margin, expert_id,
			    comment, activation_flags, date_created, date_modified)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,
			         $20,$21,$22,$23,$24,$24)
			 RETURNING position_id`,
			f.Opened.Login, f.Opened.Dealer, f.Opened.Symbol, f.Opened.Action,
			f.Opened.Digits, f.Opened.DigitsCurrency, f.Opened.Reason, f.Opened.ContractSize,
			f.Opened.TimeCreate, f.Opened.TimeUpdate, f.Opened.PriceOpen, f.Opened.PriceCurrent,
			f.Opened.PriceSL, f.Opened.PriceTP, f.Opened.Volume, f.Opened.VolumeExt,
			f.Opened.Profit, f.Opened.Storage, f.Opened.RateProfit, f.Opened.RateMargin,
			f.Opened.ExpertId, f.Opened.Comment, f.Opened.ActivationFlags,
			now).Scan(&f.Opened.PositionId); err != nil {
			return err
		}

		e.Lock()
		delete(e.Positions, temp)
		e.Positions[f.Opened.PositionId] = f.Opened
		e.Unlock()

		h.Accounts.Watch(f.Opened.Symbol, e)
	}

	for _, p := range f.Changed {
		if _, err := tx.Exec(ctx,
			`UPDATE hst.positions
			    SET volume = $1, volume_ext = $2, price_open = $3, price_current = $4,
			        price_sl = $5, price_tp = $6, profit = $7, storage = $8,
			        time_update = $9, date_modified = $9
			  WHERE position_id = $10`,
			p.Volume, p.VolumeExt, p.PriceOpen, p.PriceCurrent, p.PriceSL, p.PriceTP,
			p.Profit, p.Storage, now, p.PositionId); err != nil {
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
			    time, symbol, price, price_sl, price_tp, volume, volume_ext, volume_closed,
			    profit, value, storage, commission, fee, rate_profit, rate_margin, expert_id,
			    position_id, comment, profit_raw, price_position, tick_value, tick_size, reason,
			    market_bid, market_ask, date_created)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,
			         $21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34)
			 RETURNING deal_id`,
			d.Login, d.Dealer, d.OrderId, d.Action, d.Entry, d.Digits, d.DigitsCurrency,
			d.ContractSize, d.Time, d.Symbol, d.Price, d.PriceSL, d.PriceTP, d.Volume,
			d.VolumeExt, d.VolumeClosed, d.Profit, d.Value, d.Storage, d.Commission, d.Fee,
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

// saveAccount writes the money back; hst.users.balance trails it so a list needs no join.
func saveAccount(ctx context.Context, tx pgx.Tx, a *model.Account, now int64) error {
	if _, err := tx.Exec(ctx,
		`UPDATE hst.accounts
		    SET balance = $1, credit = $2, margin = $3, margin_free = $4, margin_level = $5,
		        profit = $6, storage = $7, floating = $8, equity = $9, updated_at = $10,
		        margin_initial = $11, margin_maintenance = $12, blocked_profit = $13
		  WHERE login = $14`,
		a.Balance, a.Credit, a.Margin, a.MarginFree, a.MarginLevel,
		a.Profit, a.Storage, a.Floating, a.Equity, now,
		a.MarginInitial, a.MarginMaintenance, a.BlockedProfit, a.Login); err != nil {
		return err
	}

	_, err := tx.Exec(ctx, `UPDATE hst.users SET balance = $1 WHERE login = $2`,
		a.Balance, a.Login)

	return err
}

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

// updateOrder writes a modified working order back.
func updateOrder(ctx context.Context, tx pgx.Tx, o *model.Order) error {
	_, err := tx.Exec(ctx,
		`UPDATE hst.orders
		    SET state = $1, type = $2, type_fill = $3, type_time = $4,
		        time_expiration = $5, time_done = $6,
		        price_order = $7, price_trigger = $8, price_current = $9,
		        price_sl = $10, price_tp = $11,
		        volume_current = $12, volume_current_ext = $13,
		        position_id = $14, position_by_id = $15, comment = $16,
		        activation_mode = $17, activation_time = $18, activation_price = $19,
		        activation_flags = $20, date_modified = $21
		  WHERE order_id = $22`,
		o.State, o.Type, o.TypeFill, o.TypeTime, o.TimeExpiration, o.TimeDone,
		o.PriceOrder, o.PriceTrigger, o.PriceCurrent, o.PriceSL, o.PriceTP,
		o.VolumeCurrent, o.VolumeExt, o.PositionId, o.PositionById, o.Comment,
		o.ActivationMode, o.ActivationTime, o.ActivationPrice, o.ActivationFlags,
		Now(), o.OrderId)

	return err
}

// SaveAccount writes the money state on its own, for the paths that change no deal.
func (h *Handler) SaveAccount(ctx context.Context, a *model.Account) error {
	if _, err := h.DB.DB.Exec(ctx,
		`UPDATE hst.accounts
		    SET balance = $1, margin = $2, margin_free = $3, margin_level = $4,
		        margin_initial = $5, margin_maintenance = $6, blocked_profit = $7,
		        profit = $8, storage = $9, floating = $10, equity = $11, updated_at = $12
		  WHERE login = $13`,
		a.Balance, a.Margin, a.MarginFree, a.MarginLevel, a.MarginInitial, a.MarginMaintenance,
		a.BlockedProfit, a.Profit, a.Storage, a.Floating, a.Equity, Now(), a.Login); err != nil {
		return err
	}

	_, err := h.DB.DB.Exec(ctx, `UPDATE hst.users SET balance = $1 WHERE login = $2`,
		a.Balance, a.Login)

	return err
}

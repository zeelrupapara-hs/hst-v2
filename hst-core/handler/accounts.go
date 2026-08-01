package handler

import (
	"context"
	"encoding/json"

	"hstcore/internal/book"
	"hstcore/model"
	"hstcore/pkg/logger"

	natscore "github.com/nats-io/nats.go"
)

func (h *Handler) BalanceSystemEventHandler(msg *natscore.Msg) {
	var e model.BalanceEvent
	if err := json.Unmarshal(msg.Data, &e); err != nil || e.Data == nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "bad balance request")
		return
	}

	res := h.NewBalance(context.Background(), e.Data)

	if msg.Reply != "" {
		if body, err := json.Marshal(res); err == nil {
			if err := msg.Respond(body); err != nil {
				h.Log.Log(logger.TypeNet, logger.CodeWarn, "could not answer a balance request",
					"login", e.Data.Login, "error", err.Error())
			}
		}
	}
}

func (h *Handler) NewBalance(ctx context.Context, req *model.BalanceRequest) *model.TradeResult {
	res := &model.TradeResult{RequestId: req.RequestId, Login: req.Login}

	if !model.IsBalanceAction(req.Action) {
		return h.refuse(res, model.RetInvalidData, "")
	}
	if req.Amount == 0 {
		return h.refuse(res, model.RetInvalidData, "")
	}

	e, ok := h.Accounts.Get(req.Login)
	if !ok {
		return h.refuse(res, model.RetTradeWrongShard, "")
	}

	e.Lock()

	credit := model.AffectsCredit(req.Action)

	// only a correction may leave an account short; everything else has to be covered
	if !req.AllowNegative && req.Amount < 0 {
		available := e.Account.Balance
		if credit {
			available = e.Account.Credit
		}
		if available+req.Amount < 0 {
			e.Unlock()
			return h.refuse(res, model.RetTradeNoMoney, "")
		}
	}

	if credit {
		e.Account.Credit += req.Amount
	} else {
		e.Account.Balance += req.Amount
	}

	deal := h.balanceDeal(e, req)

	group, _ := h.Settings.Group(e.Account.Group)
	freeProfitOnly := group != nil && group.MarginFreeProfit != 0
	h.SettleAccount(e, freeProfitOnly).Apply(e.Account)

	account := *e.Account
	e.Unlock()

	if err := h.SaveBalanceAndPublish(ctx, e, deal, &account); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not save a balance operation",
			"login", req.Login, "error", err.Error())
		return h.refuse(res, model.RetError, "")
	}

	res.RetCode = int32(model.RetOK)
	res.Message = model.RetOK.String()
	res.DealId = deal.DealId
	res.Profit = req.Amount

	h.Log.Log(logger.TypeTrade, logger.CodeOK, "balance operation",
		"login", req.Login, "action", model.BalanceActionName(req.Action),
		"amount", req.Amount, "balance", account.Balance, "dealer", req.Dealer)

	return res
}

func (h *Handler) balanceDeal(e *book.Entry, req *model.BalanceRequest) *model.Deal {
	return &model.Deal{
		Login:          req.Login,
		Dealer:         req.Dealer,
		Action:         req.Action,
		Entry:          int32(model.EntryIn),
		DigitsCurrency: e.Account.CurrencyDigits,
		Time:           Now(),
		Profit:         req.Amount,
		Value:          req.Amount,
		RateProfit:     1,
		RateMargin:     1,
		ExpertId:       req.ExpertId,
		Comment:        req.Comment,
		Reason:         int32(model.ReasonDealer),
	}
}

func (h *Handler) SaveBalanceAndPublish(ctx context.Context, e *book.Entry, d *model.Deal,
	a *model.Account) error {
	tx, err := h.DB.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := tx.QueryRow(ctx,
		`INSERT INTO hst.deals
		   (login, dealer, action, entry, digits_currency, time, profit, value,
		    rate_profit, rate_margin, expert_id, comment, reason, date_created)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$6)
		 RETURNING deal_id`,
		d.Login, d.Dealer, d.Action, d.Entry, d.DigitsCurrency, d.Time,
		d.Profit, d.Value, d.RateProfit, d.RateMargin, d.ExpertId,
		d.Comment, d.Reason).Scan(&d.DealId); err != nil {
		return err
	}

	if err := saveAccount(ctx, tx, a, Now()); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	h.PublishWS(model.SubjectAccountDeals(d.Login), model.BalanceActionName(d.Action), d)
	h.PublishWS(model.SubjectAccountSummary(a.Login), "account", a)

	return nil
}

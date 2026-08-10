package handler

import (
	"encoding/json"
	"hstcore/internal/shardmap"

	"context"

	"hstcore/internal/book"
	"hstcore/model"
	"hstcore/pkg/logger"

	natscore "github.com/nats-io/nats.go"
)

// subscribe registers every consumer.
func (h *Handler) subscribe() error {
	if err := h.Subscribe(model.SubjectSystemMarketFeedAll, h.MarketSystemEventHandler); err != nil {
		return err
	}

	for _, subject := range []string{
		model.SubjectSystemGroups,
		model.SubjectSystemGroupSymbols,
		model.SubjectSystemSymbols,
		model.SubjectSystemCommissions,
		model.SubjectSystemRules,
		model.SubjectSystemLeverages,
		model.SubjectSystemManagers,
	} {
		if err := h.Subscribe(subject, h.ConfigSystemEventHandler); err != nil {
			return err
		}
	}

	if err := h.Subscribe(model.SubjectSystemAccounts, h.AccountSystemEventHandler); err != nil {
		return err
	}

	if err := h.Subscribe(model.SubjectSystemEndOfDay, h.EndOfDaySystemEventHandler); err != nil {
		return err
	}

	if err := h.Subscribe(model.SubjectSystemEndOfDayTime, h.EndOfDayTimeSystemEventHandler); err != nil {
		return err
	}

	// one queue per command, so a command reaches exactly one pod whichever holds the account
	for _, c := range commands {
		if err := h.QueueSubscribe(c.subject, c.queue, c.handler(h)); err != nil {
			return err
		}
	}

	// and the private inbox of each shard this pod holds, where a forwarded command arrives
	for _, shard := range h.Shards.Mine() {
		if err := h.subscribeOwned(shard); err != nil {
			return err
		}
	}

	h.Log.Log(logger.TypeNet, logger.CodeOK, "engine listening",
		"shards", len(h.Shards.Mine()), "ticks", model.SubjectSystemMarketFeedAll)

	return nil
}

// command is one thing the api can ask the engine to do.
type command struct {
	topic   string
	subject string
	queue   string
	handler func(*Handler) func(*natscore.Msg)
}

var commands = []command{
	{"orders", model.SubjectSystemOrders, model.QueueOrders,
		func(h *Handler) func(*natscore.Msg) { return h.OrderSystemEventHandler }},
	{"positions", model.SubjectSystemPositions, model.QueuePositions,
		func(h *Handler) func(*natscore.Msg) { return h.PositionSystemEventHandler }},
	{"dealing", model.SubjectSystemDealing, model.QueueDealing,
		func(h *Handler) func(*natscore.Msg) { return h.DealingSystemEventHandler }},
	{"balance", model.SubjectSystemBalance, model.QueueBalance,
		func(h *Handler) func(*natscore.Msg) { return h.BalanceSystemEventHandler }},
	{"query", model.SubjectSystemQuery, model.QueueQuery,
		func(h *Handler) func(*natscore.Msg) { return h.QuerySystemEventHandler }},
}

// subscribeOwned opens the private inbox of one shard this pod holds.
func (h *Handler) subscribeOwned(shard uint32) error {
	for _, c := range commands {
		if err := h.subscribeQuiet(model.SubjectOwner(shard, c.topic), c.handler(h)); err != nil {
			return err
		}
	}

	return nil
}

// forwarded hands a command to the pod holding the account and reports that it did.
//
// A pod that does not own an account must not refuse the trade: it passes the message on with the
// caller's reply subject intact, so the owner answers the caller directly and the client never
// learns that more than one pod was involved.
func (h *Handler) forwarded(msg *natscore.Msg, topic string, login int64) bool {
	if login == 0 || h.Shards.HoldsLogin(login) {
		return false
	}

	owner := model.SubjectOwner(shardmap.ShardOf(login), topic)
	if err := h.Nats.NC.PublishRequest(owner, msg.Reply, msg.Data); err != nil {
		h.Log.Log(logger.TypeNet, logger.CodeErr, "could not forward to the owning pod",
			"login", login, "topic", topic, "error", err.Error())
	}

	return true
}

// ConfigSystemEventHandler reads the configuration again.
func (h *Handler) ConfigSystemEventHandler(msg *natscore.Msg) {
	ctx := context.Background()

	if err := h.LoadSettings(ctx); err != nil {
		h.Log.Log(logger.TypeCfg, logger.CodeErr, "could not reload settings", "error", err.Error())
		return
	}
	if err := h.LoadRoutingRules(ctx); err != nil {
		h.Log.Log(logger.TypeCfg, logger.CodeErr, "could not reload routing rules", "error", err.Error())
		return
	}
	if err := h.LoadCommission(ctx); err != nil {
		h.Log.Log(logger.TypeCfg, logger.CodeErr, "could not reload commissions", "error", err.Error())
		return
	}
	if err := h.LoadManagerGroups(ctx); err != nil {
		h.Log.Log(logger.TypeCfg, logger.CodeErr, "could not reload manager groups", "error", err.Error())
		return
	}

	h.Log.Log(logger.TypeCfg, logger.CodeOK, "configuration reloaded",
		"subject", msg.Subject, "groups", h.Settings.Groups(), "symbols", h.Settings.Symbols())

	h.ResettleAccounts(ctx)
}

// ResettleAccounts works every account out again after a settings change and tells the terminal
// what it now stands at. A new margin rate or leverage is money the client can see, so it must
// not wait for their next trade to show up.
func (h *Handler) ResettleAccounts(ctx context.Context) {
	var changed int

	h.Accounts.Each(func(e *book.Entry) {
		e.Lock()

		before := *e.Account

		// the group decides what an account is denominated in and to how many digits, so a group
		// edit has to reach the accounts already in memory or they keep answering at the old one
		if g, ok := h.Settings.Group(e.Account.Group); ok {
			e.Account.Currency = g.Currency
			e.Account.CurrencyDigits = g.CurrencyDigits
		}

		h.CalculateAccountMargins(e).Apply(e.Account)

		groupPath := e.Account.Group
		e.Unlock()

		if g, ok := h.Settings.Group(groupPath); ok {
			// margin call and stop out must react to a profile change, not wait for the next tick
			h.checkStopOut(ctx, e, g, model.Tick{})
		}

		e.Lock()
		account := *e.Account
		e.Unlock()

		// the currency and its digits decide how the summary reads, so a change to either is worth
		// announcing even when the money itself has not moved
		if account.Margin == before.Margin && account.Equity == before.Equity &&
			account.MarginFree == before.MarginFree &&
			account.Currency == before.Currency &&
			account.CurrencyDigits == before.CurrencyDigits {
			return
		}

		changed++

		if err := h.SaveAccount(ctx, &account); err != nil {
			h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not save a resettled account",
				"login", account.Login, "error", err.Error())
		}

		h.PublishAccount(&account, nil)
	})

	if changed > 0 {
		h.Log.Log(logger.TypeCfg, logger.CodeOK, "accounts resettled after a settings change",
			"accounts", changed)
	}
}

// AccountSystemEventHandler picks up an account that was opened or moved while this pod was
// already running.
func (h *Handler) AccountSystemEventHandler(msg *natscore.Msg) {
	var ev struct {
		Login int64 `json:"login"`
	}
	if err := json.Unmarshal(msg.Data, &ev); err != nil || ev.Login == 0 {
		return
	}

	if !h.Shards.HoldsLogin(ev.Login) {
		return
	}
	// an account already held has had its group, leverage or rights changed instead
	if e, held := h.Accounts.Get(ev.Login); held {
		if err := h.RefreshAccount(context.Background(), e); err != nil {
			h.Log.Log(logger.TypeCfg, logger.CodeErr, "could not refresh an account",
				"login", ev.Login, "error", err.Error())
		}
		return
	}

	if err := h.LoadAccountById(context.Background(), ev.Login); err != nil {
		h.Log.Log(logger.TypeCfg, logger.CodeErr, "could not take on an account",
			"login", ev.Login, "error", err.Error())
		return
	}

	h.Log.Log(logger.TypeCfg, logger.CodeOK, "account taken on",
		"login", ev.Login, "subject", msg.Subject)
}

// EndOfDaySystemEventHandler runs the rollover on demand, for an operator who needs it now
// rather than at the scheduled hour.
func (h *Handler) EndOfDaySystemEventHandler(msg *natscore.Msg) {
	h.EndOfDayProcess(context.Background())
}

// EndOfDayTimeSystemEventHandler moves the hour the rollover runs at, on every pod at once.
func (h *Handler) EndOfDayTimeSystemEventHandler(msg *natscore.Msg) {
	var ev struct {
		At string `json:"at"`
	}
	if err := json.Unmarshal(msg.Data, &ev); err != nil || ev.At == "" {
		h.Log.Log(logger.TypeCfg, logger.CodeWarn, "bad end of day time", "subject", msg.Subject)
		return
	}

	h.ChangeEndOfDayDate(ev.At)
}

// MarketSystemEventHandler takes one quote off the wire.
func (h *Handler) MarketSystemEventHandler(msg *natscore.Msg) {
	var t model.Tick
	if err := json.Unmarshal(msg.Data, &t); err != nil {
		h.Log.Log(logger.TypeHst, logger.CodeWarn, "bad tick", "error", err.Error())
		return
	}

	if t.Symbol == "" || !t.Ok() {
		return
	}
	if t.Time == 0 {
		t.Time = Now()
	}

	h.NotifyAll(t)
}

// OrderSystemEventHandler takes one order request off the wire.
func (h *Handler) OrderSystemEventHandler(msg *natscore.Msg) {
	var e model.OrderEvent
	if err := json.Unmarshal(msg.Data, &e); err != nil || e.Data == nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "bad order event", "subject", msg.Subject)
		return
	}

	if h.forwarded(msg, "orders", e.Data.Login) {
		return
	}

	ctx := context.Background()
	res := &model.TradeResult{RequestId: e.Data.RequestId, Login: e.Data.Login}

	switch e.EventType {
	case model.OrderEvent_new_order:
		res = h.NewOrder(ctx, e.Data)
	case model.OrderEvent_update_order:
		res = h.UpdateOrder(ctx, e.Data)
	case model.OrderEvent_cancel_order:
		res = h.CancelOrder(ctx, e.Data)
	default:
		res = h.refuse(res, model.RetInvalidData, "unknown order event")
	}

	h.reply(msg, res)
}

// PositionSystemEventHandler takes one position request off the wire.
func (h *Handler) PositionSystemEventHandler(msg *natscore.Msg) {
	var e model.PositionEvent
	if err := json.Unmarshal(msg.Data, &e); err != nil || e.Data == nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "bad position event", "subject", msg.Subject)
		return
	}

	if h.forwarded(msg, "positions", e.Data.Login) {
		return
	}

	ctx := context.Background()
	res := &model.TradeResult{RequestId: e.Data.RequestId, Login: e.Data.Login}

	switch e.EventType {
	case model.PositionEvent_update:
		res = h.UpdatePosition(ctx, e.Data)
	case model.PositionEvent_close:
		res = h.ClosePosition(ctx, e.Data)
	case model.PositionEvent_close_by:
		res = h.CloseByPosition(ctx, e.Data)
	default:
		res = h.refuse(res, model.RetInvalidData, "unknown position event")
	}

	h.reply(msg, res)
}

// reply answers whoever asked and tells the account either way, so a dropped request still lands.
func (h *Handler) reply(msg *natscore.Msg, res *model.TradeResult) {
	if msg.Reply != "" {
		if body, err := json.Marshal(res); err == nil {
			if err := msg.Respond(body); err != nil {
				h.Log.Log(logger.TypeNet, logger.CodeWarn, "could not answer a request",
					"login", res.Login, "error", err.Error())
			}
		}
	}

	h.PublishRejected(res)
}

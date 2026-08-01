package handler

import (
	"encoding/json"

	"context"

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
	} {
		if err := h.Subscribe(subject, h.ConfigSystemEventHandler); err != nil {
			return err
		}
	}

	if err := h.Subscribe(model.SubjectSystemAccounts, h.AccountSystemEventHandler); err != nil {
		return err
	}

	for _, shard := range h.Shards.Mine() {
		if err := h.subscribeShard(shard); err != nil {
			return err
		}
	}

	h.Log.Log(logger.TypeNet, logger.CodeOK, "engine listening",
		"shards", len(h.Shards.Mine()), "ticks", model.SubjectSystemMarketFeedAll)

	return nil
}

// subscribeShard opens the subjects one owned shard carries.
func (h *Handler) subscribeShard(shard uint32) error {
	if err := h.subscribeQuiet(model.SubjectShardOrders(shard), h.OrderSystemEventHandler); err != nil {
		return err
	}
	if err := h.subscribeQuiet(model.SubjectShardPositions(shard), h.PositionSystemEventHandler); err != nil {
		return err
	}

	if err := h.subscribeQuiet(model.SubjectShardDealing(shard), h.DealingSystemEventHandler); err != nil {
		return err
	}

	return h.subscribeQuiet(model.SubjectShardBalance(shard), h.BalanceSystemEventHandler)
}

// ConfigSystemEventHandler reads the configuration again.
func (h *Handler) ConfigSystemEventHandler(msg *natscore.Msg) {
	ctx := context.Background()

	if err := h.loadSettings(ctx); err != nil {
		h.Log.Log(logger.TypeCfg, logger.CodeErr, "could not reload settings", "error", err.Error())
		return
	}
	if err := h.loadRules(ctx); err != nil {
		h.Log.Log(logger.TypeCfg, logger.CodeErr, "could not reload routing rules", "error", err.Error())
		return
	}
	if err := h.loadCommissions(ctx); err != nil {
		h.Log.Log(logger.TypeCfg, logger.CodeErr, "could not reload commissions", "error", err.Error())
		return
	}

	h.Log.Log(logger.TypeCfg, logger.CodeOK, "configuration reloaded",
		"subject", msg.Subject, "groups", h.Settings.Groups(), "symbols", h.Settings.Symbols())
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
	if _, held := h.Accounts.Get(ev.Login); held {
		return
	}

	if err := h.LoadAccount(context.Background(), ev.Login); err != nil {
		h.Log.Log(logger.TypeCfg, logger.CodeErr, "could not take on an account",
			"login", ev.Login, "error", err.Error())
		return
	}

	h.Log.Log(logger.TypeCfg, logger.CodeOK, "account taken on",
		"login", ev.Login, "subject", msg.Subject)
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

	ctx := context.Background()
	res := &model.TradeResult{RequestId: e.Data.RequestId, Login: e.Data.Login}

	switch e.EventType {
	case model.OrderEventNew:
		res = h.NewOrder(ctx, e.Data)
	case model.OrderEventUpdate:
		res = h.UpdateOrder(ctx, e.Data)
	case model.OrderEventCancel:
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

	ctx := context.Background()
	res := &model.TradeResult{RequestId: e.Data.RequestId, Login: e.Data.Login}

	switch e.EventType {
	case model.PositionEventUpdate:
		res = h.UpdatePosition(ctx, e.Data)
	case model.PositionEventClose:
		res = h.ClosePosition(ctx, e.Data)
	case model.PositionEventCloseBy:
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

	h.PublishResult(res)
}

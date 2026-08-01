package handler

import (
	"context"
	"encoding/json"
	"time"

	"hstcore/internal/book"
	"hstcore/internal/settings"
	"hstcore/model"
	"hstcore/pkg/logger"

	natscore "github.com/nats-io/nats.go"
)

// The trade request pipeline.
//
//	find the account   →  it must live in this pod
//	resolve settings   →  what this group may do with this instrument
//	check              →  the broker's own limits
//	route              →  what the broker wants to do about it
//	execute            →  deals and positions
//	save and publish   →  the database, then the client
//
// The account is locked from the check to the save. Everything in between reads balances and
// positions that must not move underneath it.

// TradeRequest is what the API server sends when a client asks to trade.
type TradeRequest struct {
	RequestId string  `json:"request_id"`
	Login     int64   `json:"login"`
	Symbol    string  `json:"symbol"`
	Type      int32   `json:"type"`
	Volume    int64   `json:"volume"`
	Price     float64 `json:"price"`
	PriceSL   float64 `json:"price_sl"`
	PriceTP   float64 `json:"price_tp"`
	TypeFill  int32   `json:"type_fill"`
	TypeTime  int32   `json:"type_time"`
	Expiry    int64   `json:"expiry"`
	Comment   string  `json:"comment"`
	ExpertId  int64   `json:"expert_id"`
	Reason    int32   `json:"reason"`
	Dealer    int64   `json:"dealer"`
}

// TradeResult is what goes back.
type TradeResult struct {
	RequestId  string  `json:"request_id"`
	Login      int64   `json:"login"`
	RetCode    int32   `json:"retcode"`
	Message    string  `json:"message"`
	OrderId    int64   `json:"order_id,omitempty"`
	PositionId int64   `json:"position_id,omitempty"`
	DealId     int64   `json:"deal_id,omitempty"`
	Price      float64 `json:"price,omitempty"`
	Volume     int64   `json:"volume,omitempty"`
	Profit     float64 `json:"profit,omitempty"`
	// Rule names the routing rule that settled the request, so a refusal can be explained.
	Rule string `json:"rule,omitempty"`
}

// onTradeRequest handles one request off the wire.
func (h *Handler) onTradeRequest(msg *natscore.Msg) {
	var req TradeRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "bad trade request", "error", err.Error())
		return
	}

	res := h.Trade(context.Background(), &req)

	if msg.Reply != "" {
		if body, err := json.Marshal(res); err == nil {
			if err := msg.Respond(body); err != nil {
				h.Log.Log(logger.TypeNet, logger.CodeWarn, "could not answer a trade request",
					"login", req.Login, "error", err.Error())
			}
		}
	}

	h.publishResult(res)
}

// Trade runs one request all the way through.
func (h *Handler) Trade(ctx context.Context, req *TradeRequest) *TradeResult {
	res := &TradeResult{RequestId: req.RequestId, Login: req.Login}

	e, ok := h.Accounts.Get(req.Login)
	if !ok {
		// the request reached the wrong pod, which means the shard map disagrees somewhere
		return h.refuse(res, model.RetTradeWrongShard, "")
	}

	r, ok := h.Settings.For(e.Account.Group, req.Symbol)
	if !ok {
		return h.refuse(res, model.RetTradeBadSymbol, "")
	}

	tick, ok := h.Quotes.Get(req.Symbol)
	if !ok {
		return h.refuse(res, model.RetTradeNoQuotes, "")
	}

	order := h.orderFrom(req, r, tick)

	e.Lock()

	if code := h.Check(e, order, r, tick); !code.OK() {
		e.Unlock()
		return h.refuse(res, code, "")
	}

	decision := h.Route(&Request{
		Kind:      kindOf(order, r),
		Order:     order,
		Entry:     e,
		Rules:     r,
		Tick:      tick,
		Gapped:    h.Quotes.Gapped(req.Symbol),
		Deviation: deviation(order, tick, r),
	})

	// a rule asked to wait before deciding; hold the account through it so the request keeps
	// its place, which is what MT5 does with a delay
	if decision.Delay > 0 {
		time.Sleep(decision.Delay)
	}

	if !decision.Executes() {
		e.Unlock()
		return h.refuseByRule(res, decision)
	}

	// confirm-by-request-price fills where the client asked; confirm-by-market fills here
	price := order.PriceOrder
	if decision.AtMarket() || price <= 0 {
		price = tick.OpenPrice(order.Kind().Buy())
	}
	price = NormalisePrice(price, r.Digits)

	fill := h.Execute(e, order, r, price, Now())

	h.settle(e, order, fill, r)

	account := *e.Account
	e.Unlock()

	if err := h.save(ctx, e, order, fill, &account); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not save a trade",
			"login", req.Login, "error", err.Error())
		return h.refuse(res, model.RetError, "")
	}

	res.RetCode = int32(model.RetOK)
	res.Message = model.RetOK.String()
	res.OrderId = order.OrderId
	res.Price = price
	res.Volume = order.VolumeCurrent
	res.Profit = fill.Profit
	res.Rule = decision.Rule.Name

	if fill.Opened != nil {
		res.PositionId = fill.Opened.PositionId
	}
	if len(fill.Deals) > 0 {
		res.DealId = fill.Deals[0].DealId
	}

	h.publishTrade(e, order, fill, &account)

	h.Log.Log(logger.TypeTrade, logger.CodeOK, "trade done",
		"login", req.Login, "symbol", req.Symbol, "volume", model.Lots(order.VolumeCurrent),
		"price", price, "rule", decision.Rule.Name)

	return res
}

// settle puts the fill onto the account and recomputes the money.
func (h *Handler) settle(e *book.Entry, o *model.Order, f *Fill, r *settings.Rules) {
	// commission comes off as the deal is booked, so the client sees the true cost of the trade
	// rather than a balance that moves again later
	for _, d := range f.Deals {
		d.Commission = -h.CommissionFor(d, r)
		e.Account.Balance += d.Commission
	}

	// Only the swap comes off a closed position here. Its Profit field is the floating value
	// the last tick worked out, and the same money is already in the fill's realised profit —
	// adding both would credit the client twice.
	for _, p := range f.Closed {
		delete(e.Positions, p.PositionId)
		e.Account.Balance += p.Storage
	}

	e.Account.Balance += f.Profit

	if f.Opened != nil {
		// the id is filled in when the row is written; until then it is keyed by a temporary
		// negative number so two opens in the same tick cannot collide
		f.Opened.PositionId = h.nextTempId()
		e.Positions[f.Opened.PositionId] = f.Opened
	}

	o.State = int32(model.StateFilled)
	o.TimeDone = Now()
	o.PriceCurrent = f.Price

	Settle(e.Account, e.Positions, r.Group.MarginFreeProfit != 0).Apply(e.Account)
}

// orderFrom builds the order record a request is asking for.
func (h *Handler) orderFrom(req *TradeRequest, r *settings.Rules, t model.Tick) *model.Order {
	now := Now()

	return &model.Order{
		Login:          req.Login,
		Dealer:         req.Dealer,
		Symbol:         req.Symbol,
		Digits:         r.Digits,
		ContractSize:   r.ContractSize,
		State:          int32(model.StateStarted),
		Reason:         req.Reason,
		TimeSetup:      now,
		TimeExpiration: req.Expiry,
		Type:           req.Type,
		TypeFill:       req.TypeFill,
		TypeTime:       req.TypeTime,
		PriceOrder:     req.Price,
		PriceCurrent:   t.OpenPrice(model.OrderType(req.Type).Buy()),
		PriceSL:        req.PriceSL,
		PriceTP:        req.PriceTP,
		VolumeInitial:  req.Volume,
		VolumeCurrent:  req.Volume,
		ExpertId:       req.ExpertId,
		Comment:        req.Comment,
		RateMargin:     1,
	}
}

// kindOf is which routing request type this order counts as, which decides what rules see it.
func kindOf(o *model.Order, r *settings.Rules) model.RouteFlags {
	if o.Kind().Pending() {
		return model.RoutePending
	}

	switch r.ExecMode {
	case model.ExecInstant:
		return model.RouteInstant
	case model.ExecRequest:
		return model.RouteRequest
	case model.ExecExchange:
		return model.RouteExchange
	default:
		return model.RouteMarket
	}
}

// deviation is how far the requested price is from the market, in points. Positive means the
// requested price favours the client.
func deviation(o *model.Order, t model.Tick, r *settings.Rules) float64 {
	if o.PriceOrder <= 0 || r.Point <= 0 {
		return 0
	}

	if o.Kind().Buy() {
		return Points(t.Ask-o.PriceOrder, r.Point)
	}

	return Points(o.PriceOrder-t.Bid, r.Point)
}

// refuse fills in a refusal with its return code.
//
// Every refusal is logged. A trade that quietly does not happen is the hardest thing to explain
// to a client afterwards, and the reason has to be in the journal before they ask.
func (h *Handler) refuse(res *TradeResult, code model.RetCode, message string) *TradeResult {
	res.RetCode = int32(code)
	res.Message = message
	if res.Message == "" {
		res.Message = code.String()
	}

	h.Log.Log(logger.TypeTrade, logger.CodeWarn, "trade refused",
		"login", res.Login, "request", res.RequestId,
		"retcode", int32(code), "reason", code.String(), "rule", res.Rule)

	h.publishResult(res)

	return res
}

// refuseByRule turns a routing decision that did not execute into a refusal.
//
// A request that no rule settled is refused with its own code, so an operator reading the log
// can tell "the rules said no" from "the rules said nothing".
func (h *Handler) refuseByRule(res *TradeResult, d Decision) *TradeResult {
	if d.Rule == nil {
		return h.refuse(res, model.RetTradeNotProcessed, "")
	}

	res.Rule = d.Rule.Name

	switch d.Action {
	case model.ActionReject:
		return h.refuse(res, model.RetTradeRejected, d.Reason)
	case model.ActionRequote:
		return h.refuse(res, model.RetTradeRequote, "")
	case model.ActionDealer, model.ActionDealerOnline:
		// the request is now a dealer's to settle, so it is neither done nor refused
		res.RetCode = int32(model.RetTradeTimeout)
		res.Message = "waiting for a dealer"
		return res
	case model.ActionCancelOrder:
		return h.refuse(res, model.RetTradeRejected, "order cancelled")
	}

	return h.refuse(res, model.RetTradeNotProcessed, "")
}

// nextTempId hands out the temporary keys a position is held under until its row is written.
func (h *Handler) nextTempId() int64 {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.tempId--

	return h.tempId
}

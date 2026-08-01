package handler

import (
	"context"
	"encoding/json"
	"math"

	"hstcore/internal/book"
	"hstcore/model"
	"hstcore/pkg/logger"

	natscore "github.com/nats-io/nats.go"
)

// What happens when a price arrives.
//
// A tick is the busiest thing in the system: thousands a second, every one of them potentially
// moving every account. The trick is not to ask every account whether it cares. The book keeps
// an index from symbol to the accounts holding it, so a EURUSD tick touches only the accounts
// with EURUSD open — usually a handful, never all of them.
//
// For each of those accounts, in order: reprice the positions, add up the money, then see
// whether anything has to happen because of it — a stop out, a stop loss, a pending order.

// gapPoints is how far a price has to jump before the symbol counts as gapped. Routing rules
// can then refuse or requote while it holds.
//
// ponytail: one threshold for every instrument, split it per symbol if a broker needs it
const gapPoints = 100

// onTick handles one quote off the wire.
func (h *Handler) onTick(msg *natscore.Msg) {
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

	h.Tick(t)
}

// Tick records a price and works through the accounts it affects.
func (h *Handler) Tick(t model.Tick) {
	previous, had := h.Quotes.Set(t)

	// a jump beyond the threshold puts the symbol in a gap, which the routing rules can see
	if had {
		h.checkGap(t, previous)
	}

	watching := h.Accounts.Watching(t.Symbol)
	if len(watching) == 0 {
		return
	}

	for _, e := range watching {
		entry := e
		h.Workers.Submit(func(ctx context.Context) {
			h.applyTick(ctx, entry, t)
		})
	}
}

// checkGap decides whether the price jumped far enough to count as a gap.
func (h *Handler) checkGap(t, previous model.Tick) {
	sym, ok := h.Settings.Symbol(t.Symbol)
	if !ok || sym.Point <= 0 {
		return
	}

	jump := math.Max(math.Abs(t.Bid-previous.Bid), math.Abs(t.Ask-previous.Ask))

	h.Quotes.SetGap(t.Symbol, Points(jump, sym.Point) >= gapPoints)
}

// applyTick brings one account up to date with a new price.
func (h *Handler) applyTick(ctx context.Context, e *book.Entry, t model.Tick) {
	e.Lock()

	group := e.Account.Group

	r, ok := h.Settings.For(group, t.Symbol)
	if !ok {
		e.Unlock()
		return
	}

	h.Revalue(e, t.Symbol, t)

	money := Settle(e.Account, e.Positions, r.Group.MarginFreeProfit != 0)
	money.Apply(e.Account)

	account := *e.Account
	level := money.MarginLevel
	margin := money.Margin

	// what the price crossed, gathered while the account is locked and acted on after
	hits := h.triggersFor(e, t)

	e.Unlock()

	h.publishAccount(&account)

	for _, hit := range hits {
		h.closeOnTrigger(ctx, e, hit, t)
	}

	// stop out last: closing a position for a stop loss may have already fixed the level
	if margin > 0 && level > 0 {
		h.checkStopOut(ctx, e, r.Group, t)
	}
}

// trigger is a level the price crossed.
type trigger struct {
	position *model.Position
	reason   model.Reason
}

// triggersFor collects the stop losses and take profits this tick crossed. The account's lock
// must be held.
func (h *Handler) triggersFor(e *book.Entry, t model.Tick) []trigger {
	var hits []trigger

	for _, p := range e.Positions {
		if p.Symbol != t.Symbol {
			continue
		}

		// a position closes at the opposite side to the one it opened on
		price := t.ClosePrice(p.Buy())

		switch {
		case p.PriceSL > 0 && p.Buy() && price <= p.PriceSL:
			hits = append(hits, trigger{p, model.ReasonSL})
		case p.PriceSL > 0 && !p.Buy() && price >= p.PriceSL:
			hits = append(hits, trigger{p, model.ReasonSL})
		case p.PriceTP > 0 && p.Buy() && price >= p.PriceTP:
			hits = append(hits, trigger{p, model.ReasonTP})
		case p.PriceTP > 0 && !p.Buy() && price <= p.PriceTP:
			hits = append(hits, trigger{p, model.ReasonTP})
		}
	}

	return hits
}

// closeOnTrigger closes a position because its stop loss or take profit was reached.
//
// The routing rules see this as its own kind of request, so a broker can hold a stop loss for a
// dealer during a gap rather than letting it fill at whatever the jump produced.
func (h *Handler) closeOnTrigger(ctx context.Context, e *book.Entry, hit trigger, t model.Tick) {
	kind := model.RouteSL
	if hit.reason == model.ReasonTP {
		kind = model.RouteTP
	}

	h.closePosition2(ctx, e, hit.position, t, hit.reason, kind)
}

// closePosition2 closes a whole position at the market, whatever asked for it.
func (h *Handler) closePosition2(ctx context.Context, e *book.Entry, p *model.Position,
	t model.Tick, reason model.Reason, kind model.RouteFlags) {
	e.Lock()

	// it may already have gone: two ticks can pick up the same level
	if _, still := e.Positions[p.PositionId]; !still {
		e.Unlock()
		return
	}

	r, ok := h.Settings.For(e.Account.Group, p.Symbol)
	if !ok {
		e.Unlock()
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "cannot close: the group has no settings for the symbol",
			"login", p.Login, "group", e.Account.Group, "symbol", p.Symbol)
		return
	}

	// the closing order is the opposite side to the position
	side := int32(model.OrderSell)
	if !p.Buy() {
		side = int32(model.OrderBuy)
	}

	o := &model.Order{
		Login:         p.Login,
		Symbol:        p.Symbol,
		Digits:        r.Digits,
		ContractSize:  r.ContractSize,
		State:         int32(model.StateStarted),
		Reason:        int32(reason),
		TimeSetup:     Now(),
		Type:          side,
		VolumeInitial: p.Volume,
		VolumeCurrent: p.Volume,
		PositionId:    p.PositionId,
		RateMargin:    p.RateMargin,
		Comment:       closeComment(reason),
	}

	decision := h.Route(&Request{
		Kind: kind, Order: o, Entry: e, Rules: r, Tick: t,
		Gapped: h.Quotes.Gapped(p.Symbol),
	})

	// the rules can refuse to close, which is a real thing a broker may want during a gap
	if !decision.Executes() {
		e.Unlock()
		h.Log.Log(logger.TypeTrade, logger.CodeWarn, "a close was not admitted by any rule",
			"login", p.Login, "position", p.PositionId, "reason", reason)
		return
	}

	price := NormalisePrice(t.ClosePrice(p.Buy()), r.Digits)
	fill := h.Execute(e, o, r, price, Now())

	h.settle(e, o, fill, r)

	account := *e.Account
	e.Unlock()

	if err := h.save(ctx, e, o, fill, &account); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not save a close",
			"login", p.Login, "position", p.PositionId, "error", err.Error())
		return
	}

	h.publishTrade(e, o, fill, &account)

	h.Log.Log(logger.TypeTrade, logger.CodeOK, "position closed",
		"login", p.Login, "position", p.PositionId, "reason", reason,
		"price", price, "profit", fill.Profit)
}

// closeComment is what shows against the deal in the client's history.
func closeComment(reason model.Reason) string {
	switch reason {
	case model.ReasonSL:
		return "[sl]"
	case model.ReasonTP:
		return "[tp]"
	case model.ReasonStopOut:
		return "[so]"
	}
	return ""
}

package handler

import (
	"strconv"
	"strings"
	"time"

	"hstcore/internal/book"
	"hstcore/internal/settings"
	"hstcore/model"
	"hstcore/pkg/logger"
)

// The routing gate.
//
// Every trade request walks the rule list from the top. A rule matches when its request mask,
// its order mask and every one of its conditions agree. The first rule with a terminal action
// settles the request; a rule with a non-terminal action changes the request and lets it carry
// on down the list.
//
// A request that reaches the end of the list without a terminal action is NOT processed. This
// is MT5's behaviour: the documentation is explicit that without a catch-all rule at the bottom
// "such requests will not be processed by the server". It means the rule list is a whitelist,
// and a broker who deletes the last rule stops all trading.

// Decision is what the rule list decided to do with a request.
type Decision struct {
	// Rule is the rule that settled it, or nil if none did.
	Rule *model.RoutingRule
	// Action is what to do. Only meaningful when Rule is set.
	Action model.RouteAction
	// Reason is the text shown to the client on a refusal, up to 31 characters.
	Reason string
	// Delay is how long the rules asked to wait before executing.
	Delay time.Duration
	// ClearSL and ClearTP record levels the rules stripped off the request.
	ClearSL bool
	ClearTP bool
	// Dealers is who to hand the request to, for the dealer actions.
	Dealers []int64
}

// Executes reports whether the decision fills the request.
func (d Decision) Executes() bool { return d.Rule != nil && d.Action.Executes() }

// AtMarket reports whether the fill should use the current price rather than the requested one.
func (d Decision) AtMarket() bool { return d.Action == model.ActionConfirmMarket }

// Request is everything the rules can look at.
type Request struct {
	Kind   model.RouteFlags
	Order  *model.Order
	Entry  *book.Entry
	Rules  *settings.Rules
	Tick   model.Tick
	Gapped bool

	// Deviation is how far the requested price sits from the market, in points. Positive is in
	// the client's favour.
	Deviation float64
}

// Route walks the rule list and returns what to do.
func (h *Handler) Route(req *Request) Decision {
	var d Decision

	for i := range h.rules {
		rule := &h.rules[i]

		if !rule.Enabled() || !h.ruleMatches(rule, req) {
			continue
		}

		action := model.RouteAction(rule.Action)

		// a non-terminal action changes the request and the walk carries on below it
		if !action.Terminal() {
			h.applySoftAction(action, rule, req, &d)
			continue
		}

		d.Rule = rule
		d.Action = action
		d.Reason = rule.ActionValue
		d.Dealers = rule.Dealers

		return d
	}

	// nothing settled it, so nothing happens to it
	return d
}

// applySoftAction handles the actions that let the request carry on: a delay, or stripping a
// level off the order.
func (h *Handler) applySoftAction(action model.RouteAction, rule *model.RoutingRule,
	req *Request, d *Decision) {
	switch action {
	case model.ActionDelayTime:
		if ms, err := strconv.Atoi(rule.ActionValue); err == nil && ms > 0 {
			d.Delay += time.Duration(ms) * time.Millisecond
		}

	case model.ActionDelayTick:
		// a tick delay is counted in ticks, not time; the engine holds the request until that
		// many quotes for the symbol have gone by
		if n, err := strconv.Atoi(rule.ActionValue); err == nil && n > 0 {
			d.Delay += time.Duration(n) * tickDelayEstimate
		}

	case model.ActionClearTP:
		d.ClearTP = true
		req.Order.PriceTP = 0

	case model.ActionClearSL:
		d.ClearSL = true
		req.Order.PriceSL = 0

	case model.ActionClearSLTP:
		d.ClearSL, d.ClearTP = true, true
		req.Order.PriceSL, req.Order.PriceTP = 0, 0
	}
}

// tickDelayEstimate stands in for how long one tick takes when a rule asks to wait a number of
// ticks. MT5 counts actual quotes; this is a first approximation until the tick counter exists.
//
// ponytail: fixed estimate, swap for a real per-symbol tick counter if delay-in-ticks is used
const tickDelayEstimate = 100 * time.Millisecond

// ruleMatches reports whether a rule covers this request. Masks first because they are cheap,
// then the conditions, which may have to read the account.
func (h *Handler) ruleMatches(rule *model.RoutingRule, req *Request) bool {
	// a zero mask means "all", which is how MT5 shows an unset "Where request is"
	if rule.Request != 0 && int32(req.Kind)&rule.Request == 0 {
		return false
	}

	if rule.Type != 0 && req.Order != nil {
		if int32(model.TypeFlagFor(req.Order.Kind()))&rule.Type == 0 {
			return false
		}
	}

	// every condition must hold: they are joined with AND
	for i := range rule.Conditions {
		if !h.conditionHolds(&rule.Conditions[i], req) {
			return false
		}
	}

	return true
}

// conditionHolds evaluates one condition against the request.
func (h *Handler) conditionHolds(c *model.RoutingCondition, req *Request) bool {
	switch model.RouteCondition(c.Condition) {

	// request
	case model.CondSymbol:
		return matchMask(c, symbolOf(req))
	case model.CondVolume:
		return compareNumber(c, model.Lots(volumeOf(req)))
	case model.CondDeviation:
		return compareNumber(c, req.Deviation)
	case model.CondDeviationSpread:
		spread := Points(req.Tick.Spread(), pointOf(req))
		if spread <= 0 {
			return false
		}
		return compareNumber(c, req.Deviation/spread)
	case model.CondCurrentSpread:
		return compareNumber(c, Points(req.Tick.Spread(), pointOf(req)))
	case model.CondGap:
		return compareBool(c, req.Gapped)
	case model.CondComment:
		return matchText(c, commentOf(req))
	case model.CondExpert:
		return compareBool(c, req.Order != nil && req.Order.ExpertId != 0)
	case model.CondRequestPrice:
		return compareNumber(c, priceOf(req))
	case model.CondReason:
		if req.Order == nil {
			return false
		}
		return compareNumber(c, float64(req.Order.Reason))
	case model.CondWeekday:
		return compareNumber(c, float64(time.Now().Weekday()))
	case model.CondTime:
		now := time.Now()
		return compareNumber(c, float64(now.Hour()*60+now.Minute()))
	case model.CondDatetime:
		return compareNumber(c, float64(time.Now().Unix()))

	// account
	case model.CondLogin:
		return compareNumber(c, float64(loginOf(req)))
	case model.CondGroup:
		return matchMask(c, groupOf(req))
	case model.CondLeverage:
		if req.Entry == nil {
			return false
		}
		return compareNumber(c, float64(req.Entry.Account.Leverage))
	case model.CondBalance:
		return compareNumber(c, accountNumber(req, func(a *model.Account) float64 { return a.Balance }))
	case model.CondEquity:
		return compareNumber(c, accountNumber(req, func(a *model.Account) float64 { return a.Equity }))
	case model.CondMargin:
		return compareNumber(c, accountNumber(req, func(a *model.Account) float64 { return a.Margin }))
	case model.CondMarginFree:
		return compareNumber(c, accountNumber(req, func(a *model.Account) float64 { return a.MarginFree }))
	case model.CondMarginLevel:
		return compareNumber(c, accountNumber(req, func(a *model.Account) float64 { return a.MarginLevel }))
	case model.CondProfit:
		return compareNumber(c, accountNumber(req, func(a *model.Account) float64 { return a.Floating }))

	// positions and orders
	case model.CondPositionTotal:
		if req.Entry == nil {
			return false
		}
		return compareNumber(c, float64(len(req.Entry.Positions)))
	case model.CondOrderTotal:
		if req.Entry == nil {
			return false
		}
		return compareNumber(c, float64(len(req.Entry.Orders)))
	case model.CondPositionTotalSymbol:
		return compareNumber(c, float64(countOnSymbol(req, true)))
	case model.CondOrderTotalSymbol:
		return compareNumber(c, float64(countOnSymbol(req, false)))
	case model.CondPositionVolume:
		return compareNumber(c, model.Lots(volumeOnSymbol(req)))
	case model.CondPositionProfit:
		return compareNumber(c, profitOnSymbol(req))
	}

	// a condition the engine does not know how to read must not silently pass: a rule that was
	// meant to restrict something would then admit everything
	h.Log.Log(logger.TypeTrade, logger.CodeWarn, "routing condition not supported",
		"condition", c.Condition, "value", c.Value)

	return false
}

// compareNumber applies the rule's operator to two numbers.
func compareNumber(c *model.RoutingCondition, actual float64) bool {
	want, err := strconv.ParseFloat(strings.TrimSpace(c.Value), 64)
	if err != nil {
		return false
	}

	switch model.ConditionRule(c.Rule) {
	case model.RuleEqual:
		return actual == want
	case model.RuleNotEqual:
		return actual != want
	case model.RuleGreater:
		return actual > want
	case model.RuleNotLess:
		return actual >= want
	case model.RuleLess:
		return actual < want
	case model.RuleNotGreater:
		return actual <= want
	}

	return false
}

// compareBool handles the true/false conditions, which MT5 stores as 1 and 0.
func compareBool(c *model.RoutingCondition, actual bool) bool {
	want := strings.TrimSpace(c.Value)
	yes := want == "1" || strings.EqualFold(want, "true")

	if model.ConditionRule(c.Rule) == model.RuleNotEqual {
		return actual != yes
	}

	return actual == yes
}

// matchText compares strings the way MT5 documents it: "=" is an exact match, "greater" looks
// for the rule's value inside the actual one, and "less" looks for the actual inside the rule's.
func matchText(c *model.RoutingCondition, actual string) bool {
	want := c.Value

	switch model.ConditionRule(c.Rule) {
	case model.RuleEqual:
		return actual == want
	case model.RuleNotEqual:
		return actual != want
	case model.RuleGreater, model.RuleNotLess:
		return strings.Contains(actual, want)
	case model.RuleLess, model.RuleNotGreater:
		return strings.Contains(want, actual)
	}

	return false
}

// matchMask compares against a comma separated list of masks, where `*` stands for any run of
// characters. Used for symbols and for groups.
func matchMask(c *model.RoutingCondition, actual string) bool {
	hit := false

	for _, mask := range strings.Split(c.Value, ",") {
		mask = strings.TrimSpace(mask)
		if mask == "" {
			continue
		}

		// a leading "!" excludes, which is how MT5 writes "everything but this"
		if strings.HasPrefix(mask, "!") {
			if globMatch(strings.TrimPrefix(mask, "!"), actual) {
				return false
			}
			continue
		}

		if globMatch(mask, actual) {
			hit = true
		}
	}

	if model.ConditionRule(c.Rule) == model.RuleNotEqual {
		return !hit
	}

	return hit
}

// globMatch matches a string against a mask containing any number of `*`.
func globMatch(mask, s string) bool {
	if mask == "*" || mask == "" {
		return true
	}

	parts := strings.Split(mask, "*")

	// no star at all is a plain comparison
	if len(parts) == 1 {
		return mask == s
	}

	if !strings.HasPrefix(s, parts[0]) {
		return false
	}
	s = s[len(parts[0]):]

	for i := 1; i < len(parts)-1; i++ {
		idx := strings.Index(s, parts[i])
		if idx < 0 {
			return false
		}
		s = s[idx+len(parts[i]):]
	}

	last := parts[len(parts)-1]

	return strings.HasSuffix(s, last)
}

// Small readers, so the condition switch above stays flat and readable.

func symbolOf(r *Request) string {
	if r.Order != nil {
		return r.Order.Symbol
	}
	return ""
}

func commentOf(r *Request) string {
	if r.Order != nil {
		return r.Order.Comment
	}
	return ""
}

func priceOf(r *Request) float64 {
	if r.Order != nil {
		return r.Order.PriceOrder
	}
	return 0
}

func volumeOf(r *Request) int64 {
	if r.Order != nil {
		return r.Order.VolumeCurrent
	}
	return 0
}

func loginOf(r *Request) int64 {
	if r.Order != nil {
		return r.Order.Login
	}
	if r.Entry != nil {
		return r.Entry.Account.Login
	}
	return 0
}

func groupOf(r *Request) string {
	if r.Entry != nil {
		return r.Entry.Account.Group
	}
	return ""
}

func pointOf(r *Request) float64 {
	if r.Rules != nil {
		return r.Rules.Point
	}
	return 0
}

func accountNumber(r *Request, pick func(*model.Account) float64) float64 {
	if r.Entry == nil {
		return 0
	}
	return pick(r.Entry.Account)
}

// countOnSymbol counts the account's positions or orders on the request's symbol.
func countOnSymbol(r *Request, positions bool) int {
	if r.Entry == nil {
		return 0
	}

	symbol := symbolOf(r)
	n := 0

	if positions {
		for _, p := range r.Entry.Positions {
			if p.Symbol == symbol {
				n++
			}
		}
		return n
	}

	for _, o := range r.Entry.Orders {
		if o.Symbol == symbol {
			n++
		}
	}

	return n
}

// volumeOnSymbol is the account's net open volume on the request's symbol.
func volumeOnSymbol(r *Request) int64 {
	if r.Entry == nil {
		return 0
	}

	symbol := symbolOf(r)
	var v int64

	for _, p := range r.Entry.Positions {
		if p.Symbol != symbol {
			continue
		}
		if p.Buy() {
			v += p.Volume
			continue
		}
		v -= p.Volume
	}

	if v < 0 {
		return -v
	}

	return v
}

// profitOnSymbol is the account's floating profit on the request's symbol.
func profitOnSymbol(r *Request) float64 {
	if r.Entry == nil {
		return 0
	}

	symbol := symbolOf(r)
	var p float64

	for _, pos := range r.Entry.Positions {
		if pos.Symbol == symbol {
			p += pos.Profit
		}
	}

	return p
}

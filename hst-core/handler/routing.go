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

// Every request walks the rule list from the top.

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
func (d Decision) AtMarket() bool { return d.Action == model.RouteAction_confirm_market }

// Request is everything the rules can look at.
type Request struct {
	Kind   model.RouteFlags
	Order  *model.Order
	Entry  *book.Entry
	Rules  *settings.Rules
	Tick   model.Tick
	Gapped bool

	// Deviation is how far the requested price sits from the market, in points.
	Deviation float64
}

// Route walks the rule list and returns what to do.
func (h *Handler) Route(req *Request) Decision {
	var d Decision

	// the reload swaps the slice whole, so a snapshot under the lock is a consistent list
	h.mu.RLock()
	rules := h.rules
	h.mu.RUnlock()

	for i := range rules {
		rule := &rules[i]

		if !rule.Enabled() || !h.ruleMatches(rule, req) {
			continue
		}

		action := model.RouteAction(rule.Action)

		// a non-terminal action marks the request and the walk carries on below it
		if !action.Terminal() {
			h.applySoftAction(action, rule, &d)
			continue
		}

		// a rule that sends work to the desk needs someone to send it to
		if action.ToDealer() {
			dealers, skip := h.deskFor(rule, action, groupOf(req))
			if skip {
				continue
			}
			d.Dealers = dealers
		}

		d.Rule = rule
		d.Action = action
		d.Reason = rule.ActionValue

		// the stripped levels land only now: a request no rule settles keeps what it asked for
		if req.Order != nil {
			if d.ClearSL {
				req.Order.PriceSL = 0
			}
			if d.ClearTP {
				req.Order.PriceTP = 0
			}
		}

		return d
	}

	// nothing settled it, so nothing happens to it
	return d
}

// applySoftAction handles the actions that let the request carry on.
func (h *Handler) applySoftAction(action model.RouteAction, rule *model.RoutingRule, d *Decision) {
	switch action {
	case model.RouteAction_delay_time:
		if ms, err := strconv.Atoi(rule.ActionValue); err == nil && ms > 0 {
			d.Delay += time.Duration(ms) * time.Millisecond
		}

	case model.RouteAction_delay_tick:
		// a tick delay is counted in ticks, not time.
		if n, err := strconv.Atoi(rule.ActionValue); err == nil && n > 0 {
			d.Delay += time.Duration(n) * tickDelayEstimate
		}

	case model.RouteAction_clear_tp:
		d.ClearTP = true

	case model.RouteAction_clear_sl:
		d.ClearSL = true

	case model.RouteAction_clear_sltp:
		d.ClearSL, d.ClearTP = true, true
	}
}

// ponytail: fixed estimate for one tick, swap for a per-symbol counter if delay-in-ticks is used
const tickDelayEstimate = 100 * time.Millisecond

// ruleMatches checks the cheap masks first, then the conditions that may read the account.
func (h *Handler) ruleMatches(rule *model.RoutingRule, req *Request) bool {
	// a zero mask means all
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
	case model.RouteCondition_symbol:
		return matchMask(c, symbolOf(req))
	case model.RouteCondition_volume:
		return compareNumber(c, model.Lots(volumeOf(req)))
	case model.RouteCondition_deviation:
		return compareNumber(c, req.Deviation)
	case model.RouteCondition_deviation_spread:
		spread := Points(req.Tick.Spread(), pointOf(req))
		if spread <= 0 {
			return false
		}
		return compareNumber(c, req.Deviation/spread)
	case model.RouteCondition_current_spread:
		return compareNumber(c, Points(req.Tick.Spread(), pointOf(req)))
	case model.RouteCondition_gap:
		return compareBool(c, req.Gapped)
	case model.RouteCondition_comment:
		return matchText(c, commentOf(req))
	case model.RouteCondition_expert:
		return compareBool(c, req.Order != nil && req.Order.ExpertId != 0)
	case model.RouteCondition_request_price:
		return compareNumber(c, priceOf(req))
	case model.RouteCondition_value:
		// the request's worth in the instrument's base currency: lots by contract size
		if req.Rules == nil {
			return false
		}
		return compareNumber(c, model.Lots(volumeOf(req))*req.Rules.ContractSize)
	case model.RouteCondition_dealer_login:
		if req.Order == nil {
			return false
		}
		return compareNumber(c, float64(req.Order.Dealer))
	case model.RouteCondition_reason:
		if req.Order == nil {
			return false
		}
		return compareNumber(c, float64(req.Order.Reason))
	case model.RouteCondition_weekday:
		return compareNumber(c, float64(time.Now().Weekday()))
	case model.RouteCondition_time:
		now := time.Now()
		return compareNumber(c, float64(now.Hour()*60+now.Minute()))
	case model.RouteCondition_datetime:
		return compareNumber(c, float64(time.Now().Unix()))

	// account
	case model.RouteCondition_login:
		return compareNumber(c, float64(loginOf(req)))
	case model.RouteCondition_group:
		return matchMask(c, groupOf(req))
	case model.RouteCondition_leverage:
		if req.Entry == nil {
			return false
		}
		return compareNumber(c, float64(req.Entry.Account.Leverage))
	case model.RouteCondition_balance:
		return compareNumber(c, accountNumber(req, func(a *model.Account) float64 { return a.Balance }))
	case model.RouteCondition_equity:
		return compareNumber(c, accountNumber(req, func(a *model.Account) float64 { return a.Equity }))
	case model.RouteCondition_margin:
		return compareNumber(c, accountNumber(req, func(a *model.Account) float64 { return a.Margin }))
	case model.RouteCondition_margin_free:
		return compareNumber(c, accountNumber(req, func(a *model.Account) float64 { return a.MarginFree }))
	case model.RouteCondition_margin_level:
		return compareNumber(c, accountNumber(req, func(a *model.Account) float64 { return a.MarginLevel }))
	case model.RouteCondition_profit:
		return compareNumber(c, accountNumber(req, func(a *model.Account) float64 { return a.Floating }))
	case model.RouteCondition_country:
		return matchText(c, accountText(req, func(a *model.Account) string { return a.Country }))
	case model.RouteCondition_city:
		return matchText(c, accountText(req, func(a *model.Account) string { return a.City }))
	case model.RouteCondition_status:
		return matchText(c, accountText(req, func(a *model.Account) string { return a.Status }))
	case model.RouteCondition_client_id:
		return compareNumber(c, accountNumber(req, func(a *model.Account) float64 { return float64(a.ClientId) }))

	// positions and orders
	case model.RouteCondition_position_total:
		if req.Entry == nil {
			return false
		}
		return compareNumber(c, float64(len(req.Entry.Positions)))
	case model.RouteCondition_order_total:
		if req.Entry == nil {
			return false
		}
		return compareNumber(c, float64(len(req.Entry.Orders)))
	case model.RouteCondition_position_total_symbol:
		return compareNumber(c, float64(countOnSymbol(req, true)))
	case model.RouteCondition_order_total_symbol:
		return compareNumber(c, float64(countOnSymbol(req, false)))
	case model.RouteCondition_position_volume:
		return compareNumber(c, model.Lots(volumeOnSymbol(req)))
	case model.RouteCondition_position_profit:
		return compareNumber(c, profitOnSymbol(req))
	case model.RouteCondition_position_value:
		return compareNumber(c, valueOnSymbol(req))
	case model.RouteCondition_position_age:
		return compareNumber(c, positionSeconds(req, func(p *model.Position) int64 { return p.TimeCreate }))
	case model.RouteCondition_position_modify_time:
		return compareNumber(c, positionSeconds(req, func(p *model.Position) int64 { return p.TimeUpdate }))
	}

	// a condition the engine does not know how to read must not silently pass.
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
	case model.ConditionRule_equal:
		return actual == want
	case model.ConditionRule_not_equal:
		return actual != want
	case model.ConditionRule_greater:
		return actual > want
	case model.ConditionRule_not_less:
		return actual >= want
	case model.ConditionRule_less:
		return actual < want
	case model.ConditionRule_not_greater:
		return actual <= want
	}

	return false
}

// compareBool handles the true/false conditions, stored as 1 and 0.
func compareBool(c *model.RoutingCondition, actual bool) bool {
	want := strings.TrimSpace(c.Value)
	yes := want == "1" || strings.EqualFold(want, "true")

	if model.ConditionRule(c.Rule) == model.ConditionRule_not_equal {
		return actual != yes
	}

	return actual == yes
}

// "=" is exact, "greater" looks for the rule's value inside the actual, "less" the other way.
func matchText(c *model.RoutingCondition, actual string) bool {
	want := c.Value

	switch model.ConditionRule(c.Rule) {
	case model.ConditionRule_equal:
		return actual == want
	case model.ConditionRule_not_equal:
		return actual != want
	case model.ConditionRule_greater, model.ConditionRule_not_less:
		return strings.Contains(actual, want)
	case model.ConditionRule_less, model.ConditionRule_not_greater:
		return strings.Contains(want, actual)
	}

	return false
}

// matchMask compares against a comma separated list of masks, where `*` is any run of characters.
func matchMask(c *model.RoutingCondition, actual string) bool {
	hit := false

	for _, mask := range strings.Split(c.Value, ",") {
		mask = strings.TrimSpace(mask)
		if mask == "" {
			continue
		}

		// a leading "!" excludes
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

	if model.ConditionRule(c.Rule) == model.ConditionRule_not_equal {
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

func accountText(r *Request, pick func(*model.Account) string) string {
	if r.Entry == nil {
		return ""
	}
	return pick(r.Entry.Account)
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
		if p.IsBuy() {
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

// valueOnSymbol is what the account's positions on the request's symbol are worth at the market.
func valueOnSymbol(r *Request) float64 {
	if r.Entry == nil {
		return 0
	}

	symbol := symbolOf(r)
	var v float64

	for _, p := range r.Entry.Positions {
		if p.Symbol == symbol {
			v += model.Lots(p.Volume) * contractOf(r) * p.PriceCurrent
		}
	}

	return v
}

func contractOf(r *Request) float64 {
	if r.Rules != nil {
		return r.Rules.ContractSize
	}
	return 0
}

// positionSeconds is the age in seconds of the oldest stamp among the symbol's positions.
func positionSeconds(r *Request, stamp func(*model.Position) int64) float64 {
	if r.Entry == nil {
		return 0
	}

	symbol := symbolOf(r)
	var oldest int64

	for _, p := range r.Entry.Positions {
		if p.Symbol != symbol {
			continue
		}
		if at := stamp(p); oldest == 0 || at < oldest {
			oldest = at
		}
	}

	if oldest == 0 {
		return 0
	}

	return float64(Now()-oldest) / 1e9
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

// deskFor is who a dealer rule hands the request to, and whether the rule steps aside instead.
//
// The list keeps the order the rule names it in, because that order is the order the request is
// offered in. Only dealers who service the account's group are on it: a request a dealer cannot
// even open is not theirs to answer.
//
// Either action can carry the "skip this rule if no dealers online" box, and when it does and
// nobody eligible has connected, the walk carries on to the rule below rather than queueing
// work nobody is sitting in front of. What the two actions differ in is delivery: the online
// action reaches only the dealers who have connected, the plain one reaches everyone eligible,
// who will see it when they do.
func (h *Handler) deskFor(rule *model.RoutingRule, action model.RouteAction,
	group string) ([]int64, bool) {
	eligible := h.servicing(rule.Dealers, group)
	online := h.onlineOf(eligible)

	if len(online) == 0 && skipWhenDeskEmpty(rule.ActionValue) {
		return nil, true
	}

	if action == model.RouteAction_dealer_online {
		return online, false
	}

	return eligible, false
}

// servicing keeps the dealers whose group masks cover this account.
func (h *Handler) servicing(dealers []int64, group string) []int64 {
	if len(dealers) == 0 {
		return nil
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	out := make([]int64, 0, len(dealers))

	for _, login := range dealers {
		if maskCovers(h.managerGroups[login], group) {
			out = append(out, login)
		}
	}

	return out
}

// maskCovers reads a manager's group masks the way the rest of the platform does: a bare match
// admits, a leading "!" refuses outright, and nothing granted is nothing seen.
func maskCovers(masks []string, group string) bool {
	hit := false

	for _, mask := range masks {
		if strings.HasPrefix(mask, "!") {
			if globMatch(strings.TrimPrefix(mask, "!"), group) {
				return false
			}
			continue
		}

		if globMatch(mask, group) {
			hit = true
		}
	}

	return hit
}

// skipWhenDeskEmpty reads the flag beside the dealer actions.
func skipWhenDeskEmpty(actionValue string) bool {
	n, err := strconv.Atoi(actionValue)

	return err == nil && n == 1
}

package handler

import (
	"testing"

	"hstcore/internal/book"
	"hstcore/internal/settings"
	"hstcore/model"
	"hstcore/pkg/logger"
)

func engine(rules ...model.RoutingRule) *Handler {
	return &Handler{Log: logger.NewNop(), rules: rules}
}

func rule(name string, action model.RouteAction, conds ...model.RoutingCondition) model.RoutingRule {
	return model.RoutingRule{Name: name, Mode: 1, Action: int32(action), Conditions: conds}
}

func cond(c model.RouteCondition, r model.ConditionRule, v string) model.RoutingCondition {
	return model.RoutingCondition{Condition: int32(c), Rule: int16(r), Value: v}
}

func buyRequest() *Request {
	return &Request{
		Kind:  model.RouteMarket,
		Order: &model.Order{Login: 1001, Symbol: "EURUSD", Type: int32(model.OrderBuy), VolumeCurrent: 10000},
		Entry: &book.Entry{
			Account:   &model.Account{Login: 1001, Group: `demo\forex`, Balance: 10000},
			Positions: map[int64]*model.Position{},
			Orders:    map[int64]*model.Order{},
		},
		Rules: &settings.Rules{Point: 0.00001},
		Tick:  model.Tick{Symbol: "EURUSD", Bid: 1.1000, Ask: 1.1001},
	}
}

// The rule that matters most: with no rules at all, nothing executes. MT5 does not fall through
// to a fill, and neither do we.
func TestNoRulesMeansNothingExecutes(t *testing.T) {
	d := engine().Route(buyRequest())

	if d.Rule != nil {
		t.Fatalf("a request matched %q with an empty rule list", d.Rule.Name)
	}
	if d.Executes() {
		t.Fatal("a request executed with no rule admitting it")
	}
}

// A catch-all at the bottom is what makes anything trade.
func TestCatchAllExecutes(t *testing.T) {
	d := engine(rule("Automate other requests", model.ActionConfirmMarket)).Route(buyRequest())

	if !d.Executes() {
		t.Fatal("the catch-all did not execute the request")
	}
	if !d.AtMarket() {
		t.Fatal("confirm by market price should fill at the market")
	}
}

// A rule that does not match must not settle the request, leaving it to fall to the catch-all.
func TestNonMatchingRuleFallsThrough(t *testing.T) {
	h := engine(
		rule("Big trades only", model.ActionReject, cond(model.CondVolume, model.RuleNotLess, "10")),
		rule("Automate other requests", model.ActionConfirmMarket),
	)

	d := h.Route(buyRequest()) // one lot, so the reject rule does not apply

	if !d.Executes() {
		t.Fatalf("expected the catch-all to take it, got rule %v", d.Rule)
	}
}

// The first terminal rule wins, and the ones below it never run.
func TestFirstTerminalRuleWins(t *testing.T) {
	h := engine(
		rule("Refuse", model.ActionReject),
		rule("Automate other requests", model.ActionConfirmMarket),
	)

	d := h.Route(buyRequest())

	if d.Executes() {
		t.Fatal("a rule below a reject was allowed to execute the request")
	}
	if d.Action != model.ActionReject {
		t.Fatalf("action = %v, want reject", d.Action)
	}
}

// A delay is not terminal: it holds the request and then lets the rules below decide.
func TestDelayFallsThroughToTheNextRule(t *testing.T) {
	delay := model.RoutingRule{
		Name: "Slow it down", Mode: 1,
		Action: int32(model.ActionDelayTime), ActionValue: "250",
	}

	d := engine(delay, rule("Automate other requests", model.ActionConfirmMarket)).Route(buyRequest())

	if !d.Executes() {
		t.Fatal("a delay should not stop the request reaching the catch-all")
	}
	if d.Delay.Milliseconds() != 250 {
		t.Fatalf("delay = %v, want 250ms", d.Delay)
	}
}

// Clearing a level changes the order and then carries on.
func TestClearSLTPStripsTheLevelsAndContinues(t *testing.T) {
	req := buyRequest()
	req.Order.PriceSL, req.Order.PriceTP = 1.09, 1.11

	clear := model.RoutingRule{Name: "No levels", Mode: 1, Action: int32(model.ActionClearSLTP)}
	d := engine(clear, rule("Automate other requests", model.ActionConfirmMarket)).Route(req)

	if !d.Executes() {
		t.Fatal("clearing levels should not stop the request")
	}
	if req.Order.PriceSL != 0 || req.Order.PriceTP != 0 {
		t.Fatalf("levels not cleared: sl=%v tp=%v", req.Order.PriceSL, req.Order.PriceTP)
	}
}

// Every condition on a rule must hold. One failing condition disqualifies the whole rule.
func TestConditionsAreAndedTogether(t *testing.T) {
	h := engine(
		rule("Big EURUSD",
			model.ActionReject,
			cond(model.CondSymbol, model.RuleEqual, "EURUSD"),
			cond(model.CondVolume, model.RuleNotLess, "10"), // one lot, so this fails
		),
		rule("Automate other requests", model.ActionConfirmMarket),
	)

	if d := h.Route(buyRequest()); !d.Executes() {
		t.Fatal("a rule matched even though one of its conditions did not hold")
	}
}

// The request-type mask decides which requests a rule sees at all.
func TestRequestMaskFilters(t *testing.T) {
	pendingOnly := model.RoutingRule{
		Name: "Pending only", Mode: 1,
		Request: int32(model.RoutePending), Action: int32(model.ActionReject),
	}

	// a market request must not be caught by a pending-only rule
	d := engine(pendingOnly, rule("Automate other requests", model.ActionConfirmMarket)).Route(buyRequest())

	if !d.Executes() {
		t.Fatal("a pending-only rule caught a market request")
	}
}

// The order-type mask does the same for buy against sell.
func TestOrderMaskFilters(t *testing.T) {
	sellOnly := model.RoutingRule{
		Name: "Sells only", Mode: 1,
		Type: int32(model.TypeSell), Action: int32(model.ActionReject),
	}

	d := engine(sellOnly, rule("Automate other requests", model.ActionConfirmMarket)).Route(buyRequest())

	if !d.Executes() {
		t.Fatal("a sell-only rule caught a buy")
	}
}

// A disabled rule takes no part.
func TestDisabledRuleIsSkipped(t *testing.T) {
	off := model.RoutingRule{Name: "Off", Mode: 0, Action: int32(model.ActionReject)}

	if d := engine(off, rule("Automate other requests", model.ActionConfirmMarket)).Route(buyRequest()); !d.Executes() {
		t.Fatal("a disabled rule was applied")
	}
}

// Group masks are how a broker writes "all the demo accounts".
func TestGroupMaskMatchesSubgroups(t *testing.T) {
	h := engine(
		rule("Demo only", model.ActionReject, cond(model.CondGroup, model.RuleEqual, `demo\*`)),
		rule("Automate other requests", model.ActionConfirmMarket),
	)

	if d := h.Route(buyRequest()); d.Action != model.ActionReject {
		t.Fatalf(`demo\forex did not match the mask demo\*, action = %v`, d.Action)
	}
}

// The reject reason travels back to the client.
func TestRejectCarriesItsReason(t *testing.T) {
	refuse := model.RoutingRule{
		Name: "No", Mode: 1,
		Action: int32(model.ActionReject), ActionValue: "Out of hours",
	}

	if d := engine(refuse).Route(buyRequest()); d.Reason != "Out of hours" {
		t.Fatalf("reason = %q, want %q", d.Reason, "Out of hours")
	}
}

// An unknown condition must refuse the rule rather than pass it. A rule written to restrict
// something must never end up admitting everything because the engine did not understand it.
func TestUnknownConditionDoesNotMatch(t *testing.T) {
	h := engine(rule("Odd", model.ActionConfirmMarket,
		model.RoutingCondition{Condition: 999999, Rule: 0, Value: "1"}))

	if d := h.Route(buyRequest()); d.Executes() {
		t.Fatal("a rule with an unreadable condition was allowed to execute a request")
	}
}

func TestGlobMatch(t *testing.T) {
	cases := []struct {
		mask, value string
		want        bool
	}{
		{"*", "anything", true},
		{`demo\*`, `demo\forex`, true},
		{`demo\*`, `real\forex`, false},
		{"*forex", `demo\forex`, true},
		{"EURUSD", "EURUSD", true},
		{"EURUSD", "GBPUSD", false},
		{"*USD*", "EURUSDm", true},
	}

	for _, c := range cases {
		if got := globMatch(c.mask, c.value); got != c.want {
			t.Errorf("globMatch(%q, %q) = %v, want %v", c.mask, c.value, got, c.want)
		}
	}
}

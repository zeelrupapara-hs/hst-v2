//go:build e2e

package admin

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"hste2e/harness"

	"github.com/nats-io/nats.go"
)

// ADM-1: the first administrator logs in, sees itself, holds every right, and leaves no token in the journal.
func TestAdminLogin(t *testing.T) {
	c := harness.NewClient(env.BaseURL)
	st, res := c.Login("/auth/v1/login", env.AdminLogin, env.AdminPassword, harness.ConnAdmin)
	harness.Status(t, st, 200, "login: "+res.Error)

	var me struct {
		Login int64 `json:"login"`
	}
	st, _ = c.Get(t, "/api/v1/auth/me", &me)
	harness.Status(t, st, 200, "me")
	if me.Login != env.AdminLogin {
		t.Fatalf("me login %d, want %d", me.Login, env.AdminLogin)
	}

	var nav struct {
		Can map[string]bool `json:"can"`
	}
	st, _ = c.Get(t, "/api/v1/navigation", &nav)
	harness.Status(t, st, 200, "navigation")
	if len(nav.Can) == 0 {
		t.Fatal("navigation has no can map")
	}
	for right, ok := range nav.Can {
		if !ok {
			t.Fatalf("admin lacks %s", right)
		}
	}

	var n int
	env.Scan(t, `SELECT count(*) FROM hst.journal WHERE login = $1 AND created_at > $2 AND (message LIKE '%' || $3 || '%' OR detail::text LIKE '%' || $3 || '%')`,
		[]any{env.AdminLogin, time.Now().Add(-time.Minute).UnixNano(), c.Token.AccessToken}, &n)
	if n != 0 {
		t.Fatalf("journal rows carry the access token: %d", n)
	}
	if n := env.Count(t, `SELECT count(*) FROM hst.journal WHERE login = $1 AND created_at > $2`, env.AdminLogin, time.Now().Add(-time.Minute).UnixNano()); n == 0 {
		t.Fatal("login wrote no journal row")
	}
}

// ADM-2: the admin builds a desk; each piece is created, announced on system.*, and the engine trades on it.
func TestAdminBuildDesk(t *testing.T) {
	a := env.Admin
	suffix := harness.Now()

	// every config change must go out to the other services
	events := make(chan string, 256)
	sub, err := env.NC.Subscribe("system.>", func(m *nats.Msg) { events <- m.Subject })
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Unsubscribe()
	announced := func(subject string) {
		t.Helper()
		deadline := time.After(harness.WaitTimeout)
		for {
			select {
			case s := <-events:
				if s == subject {
					return
				}
			case <-deadline:
				t.Fatalf("%s never published", subject)
			}
		}
	}

	group := `e2e\desk` + suffix
	var g struct {
		GroupId int64 `json:"group_id"`
	}
	st, res := a.Post(t, "/api/v1/groups", harness.M{
		"group": group, "currency": "USD", "currency_digits": 2, "margin_mode": 2, "margin_call": 100, "margin_stop_out": 50,
		"demo_leverage": 100, "trade_flags": 127, "permission_flags": 3,
	}, &g)
	harness.Status(t, st, 201, "create group: "+res.Error)
	t.Cleanup(func() { a.Delete(t, fmt.Sprintf("/api/v1/groups/%d", g.GroupId)) })
	announced("system.groups.created")

	var gs struct {
		SymbolId int64 `json:"symbol_id"`
	}
	st, res = a.Post(t, fmt.Sprintf("/api/v1/groups/%d/symbols", g.GroupId), harness.M{"path": `E2E\*`, "spread_diff": 2}, &gs)
	harness.Status(t, st, 201, "create group symbol: "+res.Error)
	announced("system.group_symbols.created")

	symbol := "E2EDSK" + suffix[len(suffix)-4:]
	var sym struct {
		SymbolId int64 `json:"symbol_id"`
	}
	st, res = a.Post(t, "/api/v1/symbols", harness.M{
		"symbol": symbol, "path": `E2E\` + symbol, "currency_base": "USD", "currency_profit": "USD", "currency_margin": "USD",
		"digits": 5, "contract_size": 100000, "exec_mode": 1, "trade_mode": 4, "fill_flags": 3, "expir_flags": 15, "order_flags": 127,
		"volume_min": 100, "volume_max": 1000000, "volume_step": 100,
	}, &sym)
	harness.Status(t, st, 201, "create symbol: "+res.Error)
	t.Cleanup(func() { a.Delete(t, fmt.Sprintf("/api/v1/symbols/%d", sym.SymbolId)) })
	announced("system.symbols.created")

	var lev struct {
		Id         int64 `json:"id"`
		LeverageId int64 `json:"leverage_id"`
	}
	st, res = a.Post(t, "/api/v1/leverage-profiles", harness.M{"name": "e2e leverage " + suffix, "rules": []harness.M{
		{"name": "forex", "path": `E2E\*`, "range_mode": 0, "tiers": []harness.M{{"range_to": 0, "margin_rate_initial": 1, "margin_rate_maintenance": 1}}},
	}}, &lev)
	harness.Status(t, st, 201, "create leverage profile: "+res.Error)
	levId := lev.Id
	if levId == 0 {
		levId = lev.LeverageId
	}
	t.Cleanup(func() { a.Delete(t, fmt.Sprintf("/api/v1/leverage-profiles/%d", levId)) })
	announced("system.leverages.created")

	st, res = a.Post(t, fmt.Sprintf("/api/v1/groups/%d/commissions", g.GroupId), harness.M{
		"name": "e2e commission", "path": `E2E\*`, "tiers": []harness.M{{"range_from": 0, "range_to": 0, "value": 0, "currency": "USD"}},
	}, nil)
	harness.Status(t, st, 201, "create commission: "+res.Error)
	announced("system.commissions.created")

	var hol struct {
		Id        int64 `json:"id"`
		HolidayId int64 `json:"holiday_id"`
	}
	st, res = a.Post(t, "/api/v1/holidays", harness.M{"day": 25, "month": 12, "year": 2099, "mode": 0, "from": 0, "to": 1439, "description": "e2e " + suffix, "symbols": []string{`E2E\` + symbol}}, &hol)
	harness.Status(t, st, 201, "create holiday: "+res.Error)
	holId := hol.Id
	if holId == 0 {
		holId = hol.HolidayId
	}
	t.Cleanup(func() { a.Delete(t, fmt.Sprintf("/api/v1/holidays/%d", holId)) })
	announced("system.holidays.created")

	var rule struct {
		Id        int64 `json:"id"`
		RoutingId int64 `json:"routing_id"`
	}
	st, res = a.Post(t, "/api/v1/routing", harness.M{"name": "e2e rule " + suffix, "mode": 1, "request": 0, "type": 0, "flags": 0, "action": 1001}, &rule)
	harness.Status(t, st, 201, "create routing: "+res.Error)
	ruleId := rule.Id
	if ruleId == 0 {
		ruleId = rule.RoutingId
	}
	t.Cleanup(func() { a.Delete(t, fmt.Sprintf("/api/v1/routing/%d", ruleId)) })
	announced("system.routing.created")
	st, res = a.Post(t, fmt.Sprintf("/api/v1/routing/%d/dealers", ruleId), harness.M{"login": env.AdminLogin}, nil)
	harness.Status(t, st, 201, "add dealer: "+res.Error)
	// a dealer rule at the bottom of the list must not catch the desk's market orders
	a.Post(t, fmt.Sprintf("/api/v1/routing/%d/move-down", ruleId), nil, nil)
	a.Post(t, fmt.Sprintf("/api/v1/routing/%d/move-down", ruleId), nil, nil)

	// the engine reloaded: an account in the new group trades the seeded symbol through the new override
	tr := env.Trader(t, harness.Persona{Group: group, Balance: 1000})
	w := env.Watch(t, tr.Login)
	env.Tick(t, env.Seed.Symbol, env.Seed.Bid, env.Seed.Ask)
	st, res = tr.Client.Post(t, "/api/trader/v1/orders", harness.M{"symbol": env.Seed.Symbol, "type": 0, "volume": 0.01, "price": env.Seed.Ask, "deviation": 10}, nil)
	harness.Status(t, st, 202, "order: "+res.Error)
	pos := w.Wait(t, "position_create", nil)
	id := int64(harness.Field(t, pos, "position_id"))
	tr.Client.Post(t, fmt.Sprintf("/api/trader/v1/positions/%d/close", id), harness.M{"position_id": id}, nil)
	w.Wait(t, "position_close", func(p []byte) bool { return int64(harness.Field(t, p, "position_id")) == id })
}

// ADM-3: a client is onboarded: client record, account in a group, deposit, and the audit trail of it.
func TestAdminOnboardClient(t *testing.T) {
	a := env.Admin
	var client struct {
		Id       int64 `json:"id"`
		ClientId int64 `json:"client_id"`
	}
	st, res := a.Post(t, "/api/v1/clients", harness.M{"person_name": "E2E", "person_last_name": "Client " + harness.Now(), "contact_email": "e2e" + harness.Now() + "@example.test"}, &client)
	harness.Status(t, st, 201, "create client: "+res.Error)
	clientId := client.Id
	if clientId == 0 {
		clientId = client.ClientId
	}
	t.Cleanup(func() { a.Delete(t, fmt.Sprintf("/api/v1/clients/%d", clientId)) })

	var user struct {
		Login int64 `json:"login"`
	}
	st, res = a.Post(t, "/api/v1/users", harness.M{
		"group": env.Seed.RealGroup, "name": "e2e onboarded " + harness.Now(), "client_id": clientId, "rights": 355,
		"password_main": harness.Password, "password_investor": harness.Password + "i",
	}, &user)
	harness.Status(t, st, 201, "create user: "+res.Error)
	t.Cleanup(func() {
		a.Post(t, "/api/v1/balance/withdrawal", harness.M{"login": user.Login, "amount": 10000, "comment": "e2e cleanup"}, nil)
		a.Delete(t, fmt.Sprintf("/api/v1/users/%d", user.Login))
	})

	w := env.Watch(t, user.Login)
	since := time.Now().UnixNano()
	env.Deposit(t, user.Login, 10000)
	w.Wait(t, "money_change", nil)

	var view struct {
		Balance float64 `json:"balance"`
		Group   string  `json:"group"`
	}
	st, _ = a.Get(t, fmt.Sprintf("/api/v1/users/%d", user.Login), &view)
	harness.Status(t, st, 200, "get user")
	harness.Approx(t, view.Balance, 10000, 0.001, "balance after deposit")
	if view.Group != env.Seed.RealGroup {
		t.Fatalf("group %q", view.Group)
	}

	if n := env.Count(t, `SELECT count(*) FROM hst.deals WHERE login = $1 AND action = 2 AND profit = 10000`, user.Login); n != 1 {
		t.Fatalf("balance deals: %d, want 1", n)
	}
	var messages []string
	rows, err := env.DB.Query(t.Context(), `SELECT message FROM hst.journal WHERE login = $1 AND created_at >= $2`, env.AdminLogin, since)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var m string
		_ = rows.Scan(&m)
		messages = append(messages, m)
	}
	rows.Close()
	found := false
	for _, m := range messages {
		if strings.Contains(m, fmt.Sprint(user.Login)) {
			found = true
		}
	}
	if !found {
		t.Fatalf("no journal row by actor %d naming %d; rows: %v", env.AdminLogin, user.Login, messages)
	}
}

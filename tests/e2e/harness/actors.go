package harness

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

// Persona is how a trader account is created.
type Persona struct {
	Group    string
	Balance  float64
	Leverage int32
	// Investor logs the session in with the investor password, a read-only scope.
	Investor bool
	// Disabled sets the trade-disabled right.
	Disabled bool
}

// Trader is a logged-in trading account the test owns.
type Trader struct {
	Login  int64
	Group  string
	Client *Client
	env    *Env
}

// Trader creates an account from the persona, funds it and logs it in; removed on cleanup.
func (e *Env) Trader(t *testing.T, p Persona) *Trader {
	t.Helper()
	if p.Group == "" {
		p.Group = e.Seed.RealGroup
	}
	if p.Leverage == 0 {
		p.Leverage = 100
	}
	rights := userRights
	if p.Disabled {
		rights |= 4
	}
	login := e.createUser(t, p.Group, "e2e trader "+Now(), rights)
	Status(t, statusOf(e.Admin.Patch(t, fmt.Sprintf("/api/v1/users/%d", login), M{"leverage": p.Leverage}, nil)), 200, "set leverage")
	if p.Balance != 0 {
		e.Deposit(t, login, p.Balance)
	}

	tr := &Trader{Login: login, Group: p.Group, Client: NewClient(e.BaseURL), env: e}
	if st, res := tr.Client.Login("/auth/trader/v1/login", login, Password, ConnTrader); st != 200 {
		t.Fatalf("trader login %d: %d %s", login, st, res.Error)
	}
	return tr
}

// Manager is a logged-in staff login with the rights and group masks the test asked for.
type Manager struct {
	Login  int64
	Client *Client
}

// Manager creates a user in the managers group, makes it staff and logs it in; removed on cleanup.
func (e *Env) Manager(t *testing.T, name string, rights, masks []string) *Manager {
	t.Helper()
	login := e.createUser(t, e.Seed.ManagerGroup, name+" "+Now(), userRights)
	st, res := e.Admin.Post(t, "/api/v1/managers", M{"login": login, "name": name, "groups": masks, "rights": rights}, nil)
	Status(t, st, 201, "create manager: "+res.Error)
	t.Cleanup(func() {
		e.Admin.Delete(t, fmt.Sprintf("/api/v1/managers/%d", login))
	})
	m := &Manager{Login: login, Client: NewClient(e.BaseURL)}
	if st, res := m.Client.Login("/auth/v1/login", login, Password, ConnManager); st != 200 {
		t.Fatalf("manager login %d: %d %s", login, st, res.Error)
	}
	return m
}

// DeskManager is the desk persona: accounts and trades inside e2e\real.
func (e *Env) DeskManager(t *testing.T) *Manager {
	t.Helper()
	return e.Manager(t, "e2e desk", []string{"right_manager", "right_acc_read", "right_acc_manager", "right_trades_read", "right_trades_manager",
		"right_clients_access", "right_clients_create", "right_clients_edit"}, []string{e.Seed.RealGroup + `\*`})
}

// createUser makes a user over the admin api and schedules its removal.
func (e *Env) createUser(t *testing.T, group, name string, rights int) int64 {
	t.Helper()
	var out idOut
	st, res := e.Admin.Post(t, "/api/v1/users", M{
		"group": group, "name": name, "rights": rights,
		"password_main": Password, "password_investor": Password + "i",
	}, &out)
	Status(t, st, 201, "create user: "+res.Error)
	t.Cleanup(func() { e.removeUser(t, out.Login) })
	return out.Login
}

// removeUser zeroes the balance, then deletes; a login that cannot be deleted is reported, not failed.
func (e *Env) removeUser(t *testing.T, login int64) {
	var balance float64
	_ = e.DB.QueryRow(context.Background(), `SELECT balance FROM hst.users WHERE login = $1`, login).Scan(&balance)
	// the exact amount, so float residue from the trades does not keep the account funded
	if balance != 0 {
		path, amount := "/api/v1/balance/withdrawal", balance
		if balance < 0 {
			path, amount = "/api/v1/balance/deposit", -balance
		}
		e.Admin.Post(t, path, M{"login": login, "amount": amount, "comment": "e2e cleanup"}, nil)
	}
	// hst-server refuses to delete any login with a row in hst.orders, filled history included, so a traded account stays
	if st, res := e.Admin.Delete(t, fmt.Sprintf("/api/v1/users/%d", login)); st != 200 && st != 204 {
		t.Logf("cleanup: user %d left behind: %d %s %s", login, st, res.Error, res.Message)
	}
}

// Deposit funds an account through the accountant route and fails on a refusal.
func (e *Env) Deposit(t *testing.T, login int64, amount float64) {
	t.Helper()
	var res struct {
		RetCode int `json:"retcode"`
	}
	// the engine learns of a new account from a nats event, so the first deposit may land before it did
	deadline := time.Now().Add(WaitTimeout)
	for {
		st, r := e.Admin.Post(t, "/api/v1/balance/deposit", M{"login": login, "amount": amount, "comment": "e2e deposit"}, &res)
		if st == 400 && strings.Contains(r.Message, "account not found") && time.Now().Before(deadline) {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		Status(t, st, 200, "deposit: "+r.Error+" "+r.Message)
		break
	}
	if res.RetCode != 0 {
		t.Fatalf("deposit refused, retcode %d", res.RetCode)
	}
}

func statusOf(st int, _ Response) int { return st }

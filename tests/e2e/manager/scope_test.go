//go:build e2e

package manager

import (
	"fmt"
	"testing"

	"hste2e/harness"
)

// MGR-2: the desk manager reaches accounts inside its masks and nothing outside them.
func TestManagerScope(t *testing.T) {
	m := env.DeskManager(t)
	inside := env.Trader(t, harness.Persona{Group: env.Seed.RealGroup})
	outside := env.Trader(t, harness.Persona{Group: env.Seed.NetGroup})

	var users []struct {
		Login int64 `json:"login"`
	}
	st, res := m.Client.Get(t, "/api/v1/users?limit=500", &users)
	harness.Status(t, st, 200, "list users: "+res.Error)
	seen := map[int64]bool{}
	for _, u := range users {
		seen[u.Login] = true
	}
	if !seen[inside.Login] || seen[outside.Login] {
		t.Fatalf("list: inside %v outside %v", seen[inside.Login], seen[outside.Login])
	}

	st, _ = m.Client.Get(t, fmt.Sprintf("/api/v1/users/%d", inside.Login), nil)
	harness.Status(t, st, 200, "get inside")
	st, _ = m.Client.Patch(t, fmt.Sprintf("/api/v1/users/%d", inside.Login), harness.M{"name": "renamed by desk"}, nil)
	harness.Status(t, st, 200, "update inside")
	st, _ = m.Client.Get(t, fmt.Sprintf("/api/v1/users/%d", outside.Login), nil)
	harness.Status(t, st, 404, "get outside")

	st, _ = m.Client.Post(t, "/api/v1/users", harness.M{"group": env.Seed.NetGroup, "name": "e2e intruder", "password_main": harness.Password, "password_investor": harness.Password + "i"}, nil)
	harness.Status(t, st, 403, "create outside")
	st, _ = m.Client.Patch(t, fmt.Sprintf("/api/v1/users/%d", inside.Login), harness.M{"group": env.Seed.NetGroup}, nil)
	harness.Status(t, st, 403, "move out of mask")

	st, _ = m.Client.Post(t, fmt.Sprintf("/api/v1/users/%d/password", inside.Login), harness.M{"kind": "main", "password": harness.Password + "2"}, nil)
	harness.Status(t, st, 204, "reset password inside")
	st, _ = m.Client.Post(t, fmt.Sprintf("/api/v1/users/%d/password", outside.Login), harness.M{"kind": "main", "password": harness.Password + "2"}, nil)
	harness.Status(t, st, 404, "reset password outside")
}

// MGR-4: every route checks its own right bit; the desk manager has none of these.
func TestManagerRights(t *testing.T) {
	m := env.DeskManager(t)
	tr := env.Trader(t, harness.Persona{Group: env.Seed.RealGroup})

	st, _ := m.Client.Post(t, "/api/v1/balance/deposit", harness.M{"login": tr.Login, "amount": 1}, nil)
	harness.Status(t, st, 403, "deposit without accountant")
	st, _ = m.Client.Post(t, "/api/v1/symbols", harness.M{"symbol": "E2ENOPE", "path": `E2E\E2ENOPE`, "currency_base": "USD", "currency_profit": "USD", "currency_margin": "USD"}, nil)
	harness.Status(t, st, 403, "create symbol without cfg symbols")
	st, _ = m.Client.Get(t, "/api/v1/journal", nil)
	harness.Status(t, st, 403, "journal without srv journals")
	st, _ = m.Client.Patch(t, fmt.Sprintf("/api/v1/managers/%d", m.Login), harness.M{"name": "self", "rights": []string{"right_admin"}}, nil)
	harness.Status(t, st, 403, "edit manager without cfg managers")
}

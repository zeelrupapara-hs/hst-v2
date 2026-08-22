package harness

import (
	"context"
	"fmt"
	"net/http"
)

// Seed is the baseline every run shares: fixed names, created once, never deleted.
type Seed struct {
	RealGroup    string
	NetGroup     string
	ManagerGroup string
	RealGroupId  int64
	NetGroupId   int64
	// Symbol is instant execution, MarketSymbol is market execution.
	Symbol         string
	MarketSymbol   string
	SymbolId       int64
	MarketSymbolId int64
	FeedId         int64
	// Bid and Ask are the prices the seed ticks at.
	Bid float64
	Ask float64
}

// Password every seeded account uses.
const Password = "E2e-Pass-1234!"

// userRights is enabled|password|trailing|expert|reports, the same as the panel's default user.
const userRights = 355

type idOut struct {
	Id       int64 `json:"id"`
	GroupId  int64 `json:"group_id"`
	SymbolId int64 `json:"symbol_id"`
	FeedId   int64 `json:"datafeed_id"`
	Login    int64 `json:"login"`
}

// seed makes the run baseline idempotently: groups, symbols, routing rule, feed.
func (e *Env) seed() error {
	s := &e.Seed
	s.RealGroup, s.NetGroup, s.ManagerGroup = `e2e\real`, `e2e\net`, `e2e\managers`
	s.Symbol, s.MarketSymbol = "E2EUSD", "E2EMKT"
	s.Bid, s.Ask = 1.10000, 1.10002

	var err error
	if s.RealGroupId, err = e.ensureGroup(s.RealGroup, 2); err != nil {
		return err
	}
	if s.NetGroupId, err = e.ensureGroup(s.NetGroup, 0); err != nil {
		return err
	}
	if _, err = e.ensureGroup(s.ManagerGroup, 2); err != nil {
		return err
	}
	if s.SymbolId, err = e.ensureSymbol(s.Symbol, 1); err != nil {
		return err
	}
	if s.MarketSymbolId, err = e.ensureSymbol(s.MarketSymbol, 2); err != nil {
		return err
	}
	for _, gid := range []int64{s.RealGroupId, s.NetGroupId} {
		if err := e.ensureGroupSymbol(gid, `E2E\*`); err != nil {
			return err
		}
	}
	if err := e.ensureRouting(); err != nil {
		return err
	}
	if s.FeedId, err = e.ensureFeed(); err != nil {
		return err
	}
	return nil
}

func (e *Env) ensureGroup(name string, marginMode int) (int64, error) {
	var id int64
	if err := e.DB.QueryRow(context.Background(), `SELECT group_id FROM hst.groups WHERE "group" = $1`, name).Scan(&id); err == nil {
		return id, nil
	}
	var out idOut
	st, res, _ := e.Admin.Request("POST", "/api/v1/groups", M{
		"group": name, "currency": "USD", "currency_digits": 2, "company": "e2e",
		"margin_mode": marginMode, "margin_call": 100, "margin_stop_out": 50, "margin_so_mode": 0,
		"demo_leverage": 100, "trade_flags": 127, "permission_flags": 3, "status": "enabled",
	}, &out)
	if st != http.StatusCreated {
		return 0, fmt.Errorf("create group %s: %d %s", name, st, res.Error)
	}
	return out.GroupId, nil
}

func (e *Env) ensureSymbol(name string, execMode int) (int64, error) {
	var id int64
	if err := e.DB.QueryRow(context.Background(), `SELECT symbol_id FROM hst.symbols WHERE symbol = $1`, name).Scan(&id); err == nil {
		return id, nil
	}
	sessions := []M{}
	for day := 0; day < 7; day++ {
		sessions = append(sessions, M{"type": 0, "day": day, "open": 0, "close": 1440}, M{"type": 1, "day": day, "open": 0, "close": 1440})
	}
	var out idOut
	st, res, _ := e.Admin.Request("POST", "/api/v1/symbols", M{
		"symbol": name, "path": `E2E\` + name, "description": "e2e test symbol",
		"currency_base": "USD", "currency_profit": "USD", "currency_margin": "USD",
		"digits": 5, "contract_size": 100000, "calc_mode": 0, "exec_mode": execMode, "trade_mode": 4,
		"fill_flags": 3, "expir_flags": 15, "order_flags": 127,
		"volume_min": 100, "volume_max": 1000000, "volume_step": 100,
		"stops_level": 10, "freeze_level": 5, "quotes_timeout": 0, "spread": 0,
		"sessions": sessions,
	}, &out)
	if st != http.StatusCreated {
		return 0, fmt.Errorf("create symbol %s: %d %s", name, st, res.Error)
	}
	return out.SymbolId, nil
}

func (e *Env) ensureGroupSymbol(groupId int64, path string) error {
	var n int
	_ = e.DB.QueryRow(context.Background(), `SELECT count(*) FROM hst.groups_symbols WHERE group_id = $1 AND path = $2`, groupId, path).Scan(&n)
	if n > 0 {
		return nil
	}
	st, res, _ := e.Admin.Request("POST", fmt.Sprintf("/api/v1/groups/%d/symbols", groupId), M{"path": path}, nil)
	if st != http.StatusCreated {
		return fmt.Errorf("create group symbol %d %s: %d %s", groupId, path, st, res.Error)
	}
	return nil
}

// ensureRouting makes sure some rule confirms at market, or every order is refused.
func (e *Env) ensureRouting() error {
	var n int
	_ = e.DB.QueryRow(context.Background(), `SELECT count(*) FROM hst.routing WHERE action = 1006`).Scan(&n)
	if n > 0 {
		return nil
	}
	st, res, _ := e.Admin.Request("POST", "/api/v1/routing", M{"name": "e2e confirm at market", "mode": 1, "request": 0, "type": 0, "flags": 0, "action": 1006}, nil)
	if st != http.StatusCreated {
		return fmt.Errorf("create routing: %d %s", st, res.Error)
	}
	return nil
}

// ensureFeed registers the fake DDE server as a quotes feed translating the run's symbols.
func (e *Env) ensureFeed() (int64, error) {
	ctx := context.Background()
	server := fmt.Sprintf("127.0.0.1:%d", e.FeedPort)
	var id int64
	if err := e.DB.QueryRow(ctx, `SELECT datafeed_id FROM hst.datafeeds WHERE name = 'e2e-dde'`).Scan(&id); err != nil {
		var out idOut
		st, res, _ := e.Admin.Request("POST", "/api/v1/datafeeds", M{
			"name": "e2e-dde", "module": "dde", "mode": 1, "enable": 0,
			"feed_server": server, "feed_login": "e2e", "feed_password": "e2e", "timeout_reconnect": 1,
		}, &out)
		if st != http.StatusCreated {
			return 0, fmt.Errorf("create feed: %d %s", st, res.Error)
		}
		id = out.FeedId
		if st, res, _ := e.Admin.Request("POST", fmt.Sprintf("/api/v1/datafeeds/%d/symbols", id), M{"path": `E2E\*`}, nil); st != http.StatusCreated {
			return 0, fmt.Errorf("feed symbols: %d %s", st, res.Error)
		}
		for _, sid := range []int64{e.Seed.SymbolId, e.Seed.MarketSymbolId} {
			var sym string
			_ = e.DB.QueryRow(ctx, `SELECT symbol FROM hst.symbols WHERE symbol_id = $1`, sid).Scan(&sym)
			if st, res, _ := e.Admin.Request("POST", fmt.Sprintf("/api/v1/datafeeds/%d/translates", id), M{"symbol_id": sid, "source": sym, "digits": 5}, nil); st != http.StatusCreated {
				return 0, fmt.Errorf("feed translate %s: %d %s", sym, st, res.Error)
			}
		}
	}
	// the port may differ between runs, and activating republishes the worker config
	if st, res, _ := e.Admin.Request("PATCH", fmt.Sprintf("/api/v1/datafeeds/%d", id), M{"feed_server": server}, nil); st != 200 {
		return 0, fmt.Errorf("feed server: %d %s", st, res.Error)
	}
	if st, res, _ := e.Admin.Request("POST", fmt.Sprintf("/api/v1/datafeeds/%d/activate", id), nil, nil); st != 200 {
		return 0, fmt.Errorf("feed activate: %d %s", st, res.Error)
	}
	return id, nil
}

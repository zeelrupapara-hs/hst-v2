//go:build e2e

package trader

import (
	"fmt"
	"testing"

	"hste2e/harness"
)

// vias runs one scenario over rest and then over the websocket.
var vias = []string{"rest", "ws"}

// trade is the trader plus the two command surfaces a sub-test drives.
type trade struct {
	t   *testing.T
	tr  *harness.Trader
	ws  *harness.WS
	via string
}

// open sends a market order and returns the position_create payload.
func (x *trade) open(w *harness.Watcher, symbol string, typ int, volume, price float64) []byte {
	x.t.Helper()
	body := harness.M{"symbol": symbol, "type": typ, "volume": volume, "price": price, "type_fill": 0, "type_time": 0, "deviation": 10}
	if x.via == "ws" {
		x.ws.Send(x.t, "order_create", body)
	} else {
		st, res := x.tr.Client.Post(x.t, "/api/trader/v1/orders", body, nil)
		harness.Status(x.t, st, 202, "order_create: "+res.Error)
	}
	return w.Wait(x.t, "position_create", func(p []byte) bool {
		return harness.Field(x.t, p, "action") == float64(typ) && harness.Field(x.t, p, "volume") == volume
	})
}

// close sends a close for a position, whole when volume is 0.
func (x *trade) close(positionId int64, volume, price float64) (int, harness.Response) {
	x.t.Helper()
	body := harness.M{"position_id": positionId, "volume": volume, "price": price, "deviation": 10}
	if x.via == "ws" {
		x.ws.Send(x.t, "position_close", body)
		return 202, harness.Response{}
	}
	return x.tr.Client.Post(x.t, fmt.Sprintf("/api/trader/v1/positions/%d/close", positionId), body, nil)
}

func newTrade(t *testing.T, via string) *trade {
	t.Helper()
	tr := env.Trader(t, harness.Persona{Balance: 10000})
	x := &trade{t: t, tr: tr, via: via}
	if via == "ws" {
		x.ws = tr.WS(t)
	}
	env.Tick(t, env.Seed.Symbol, env.Seed.Bid, env.Seed.Ask)
	return x
}

// TRD-2: market buy and sell open two positions; margin, equity and deals are right.
func TestMarketBuySell(t *testing.T) {
	for _, via := range vias {
		t.Run("via="+via, func(t *testing.T) {
			x := newTrade(t, via)
			w := env.Watch(t, x.tr.Login)
			s := env.Seed

			buy := x.open(w, s.Symbol, 0, 0.10, s.Ask)
			harness.Approx(t, harness.Field(t, buy, "price_open"), s.Ask, 1e-9, "buy opens at ask")
			// forex margin = lots * contract / leverage, one leg only
			sum := w.WaitSummary(t, func(m harness.Summary) bool { return m.Margin > 0 })
			harness.Approx(t, sum.Margin, 0.10*100000/100, 0.01, "margin after buy")
			harness.Approx(t, sum.Equity, sum.Balance+sum.Profit, 0.01, "equity = balance + profit")
			harness.Approx(t, sum.MarginFree, sum.Equity-sum.Margin, 0.01, "free = equity - margin")

			sell := x.open(w, s.Symbol, 1, 0.10, s.Bid)
			harness.Approx(t, harness.Field(t, sell, "price_open"), s.Bid, 1e-9, "sell opens at bid")

			if n := env.Count(t, `SELECT count(*) FROM hst.positions WHERE login = $1`, x.tr.Login); n != 2 {
				t.Fatalf("positions in db: %d, want 2", n)
			}
			if n := env.Count(t, `SELECT count(*) FROM hst.deals WHERE login = $1 AND entry = 0 AND action IN (0,1)`, x.tr.Login); n != 2 {
				t.Fatalf("in deals: %d, want 2", n)
			}
			if via == "ws" {
				x.ws.Wait(t, "position_create", func(p []byte) bool { return harness.Field(t, p, "action") == 1 })
			}

			// leave the account flat so cleanup can delete it
			for _, p := range [][]byte{buy, sell} {
				id := int64(harness.Field(t, p, "position_id"))
				x.close(id, 0, 0)
				w.Wait(t, "position_close", func(q []byte) bool { return int64(harness.Field(t, q, "position_id")) == id })
			}
		})
	}
}

// TRD-12: a full close pays profit = (bid - open) * lots * contract into the balance and removes the position.
func TestCloseFull(t *testing.T) {
	for _, via := range vias {
		t.Run("via="+via, func(t *testing.T) {
			x := newTrade(t, via)
			w := env.Watch(t, x.tr.Login)
			s := env.Seed

			buy := x.open(w, s.Symbol, 0, 0.10, s.Ask)
			id := int64(harness.Field(t, buy, "position_id"))
			open := harness.Field(t, buy, "price_open")

			// move the market up, then close at the new bid
			bid, ask := 1.10050, 1.10052
			env.Tick(t, s.Symbol, bid, ask)
			st, res := x.close(id, 0, bid)
			harness.Status(t, st, 202, "close: "+res.Error)

			closed := w.Wait(t, "position_close", func(p []byte) bool { return int64(harness.Field(t, p, "position_id")) == id })
			want := (bid - open) * 0.10 * 100000
			deal := w.Wait(t, "deal_create", func(p []byte) bool {
				return int64(harness.Field(t, p, "position_id")) == id && harness.Field(t, p, "entry") == 1
			})
			harness.Approx(t, harness.Field(t, deal, "profit"), want, 0.01, "out deal profit")
			harness.Approx(t, harness.Field(t, deal, "price"), bid, 1e-9, "closed at bid")
			_ = closed

			storage, commission := harness.Field(t, deal, "storage"), harness.Field(t, deal, "commission")
			sum := w.WaitSummary(t, func(m harness.Summary) bool { return m.Margin == 0 })
			harness.Approx(t, sum.Balance, 10000+want+storage+commission, 0.01, "balance after close")
			if via == "ws" {
				x.ws.Wait(t, "position_close", func(p []byte) bool { return int64(harness.Field(t, p, "position_id")) == id })
			}

			if n := env.Count(t, `SELECT count(*) FROM hst.positions WHERE position_id = $1`, id); n != 0 {
				t.Fatalf("position %d still in db", id)
			}
			var balance float64
			env.Scan(t, `SELECT COALESCE(SUM(profit + storage + commission), 0) FROM hst.deals WHERE login = $1`, []any{x.tr.Login}, &balance)
			harness.Approx(t, balance, sum.Balance, 0.01, "deals reconcile with balance")
		})
	}
}

// TRD-13: a partial close leaves the rest open; a volume off the step is refused.
func TestClosePartial(t *testing.T) {
	for _, via := range vias {
		t.Run("via="+via, func(t *testing.T) {
			x := newTrade(t, via)
			w := env.Watch(t, x.tr.Login)
			s := env.Seed

			buy := x.open(w, s.Symbol, 0, 0.10, s.Ask)
			id := int64(harness.Field(t, buy, "position_id"))

			x.close(id, 0.04, s.Bid)
			rest := w.Wait(t, "position_update", func(p []byte) bool { return int64(harness.Field(t, p, "position_id")) == id })
			harness.Approx(t, harness.Field(t, rest, "volume"), 0.06, 1e-9, "remaining volume")
			deal := w.Wait(t, "deal_create", func(p []byte) bool {
				return int64(harness.Field(t, p, "position_id")) == id && harness.Field(t, p, "entry") == 1
			})
			harness.Approx(t, harness.Field(t, deal, "volume"), 0.04, 1e-9, "out deal volume")

			var volume int64
			env.Scan(t, `SELECT volume FROM hst.positions WHERE position_id = $1`, []any{id}, &volume)
			if volume != 600 {
				t.Fatalf("position volume in db %d, want 600", volume)
			}

			// 0.015 is not a multiple of the 0.01 step
			if st, _ := x.close(id, 0.015, s.Bid); st == 202 {
				w.Wait(t, "order_rejected", func(p []byte) bool {
					harness.Retcode(t, p, 10008)
					return true
				})
			} else if st != 400 {
				t.Fatalf("bad step close: status %d, want 400 or a rejection", st)
			}

			x.close(id, 0, 0)
			w.Wait(t, "position_close", func(p []byte) bool { return int64(harness.Field(t, p, "position_id")) == id })
		})
	}
}

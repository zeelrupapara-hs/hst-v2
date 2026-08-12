package v1

import (
	"context"
	"encoding/json"
	"math"
	"sort"
	"strings"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
)

// ViewExposure is one currency's open exposure, split by who holds it.
type ViewExposure struct {
	Currency string  `json:"currency"`
	Clients  float64 `json:"clients"`
	Coverage float64 `json:"coverage"`
	Net      float64 `json:"net"`
	Rate     float64 `json:"rate"`
	NetUsd   float64 `json:"net_usd"`
}

// ViewExposureTotals sums the per-currency USD nets.
type ViewExposureTotals struct {
	NetUsd float64 `json:"net_usd"`
}

// ViewExposureReport is the whole answer: rows plus totals.
type ViewExposureReport struct {
	Currencies []ViewExposure     `json:"currencies"`
	Totals     ViewExposureTotals `json:"totals"`
}

// usdRate is currency->USD via the last redis quote: XUSD direct, USDX inverted, else 0.
func (s *HttpServer) usdRate(ctx context.Context, currency string) float64 {
	if currency == "USD" {
		return 1
	}
	if bid := s.lastBid(ctx, currency+"USD"); bid > 0 {
		return bid
	}
	if bid := s.lastBid(ctx, "USD"+currency); bid > 0 {
		return 1 / bid
	}
	return 0
}

// lastBid reads a symbol's last quote from redis; 0 when there is none.
func (s *HttpServer) lastBid(ctx context.Context, symbol string) float64 {
	raw, err := s.Redis.Client.Get(ctx, "hstquote:last:"+symbol).Result()
	if err != nil {
		return 0
	}
	var t model.Tick
	if json.Unmarshal([]byte(raw), &t) != nil {
		return 0
	}
	return t.Bid
}

// GetExposure splits every open position into its two currency legs and nets them per currency.
//
//	@Id			GetExposure
//	@Tags		Exposure
//	@Produce	json
//	@Success	200	{object}	Response{data=ViewExposureReport}
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/exposure [get]
func (s *HttpServer) GetExposure(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	where, args := groupWhere(snap.IsManager, snap.ManagerGroups, 1)

	rows, err := s.DB.DB.Query(c.UserContext(), `
		SELECT p.action, GREATEST(p.volume_ext, p.volume * 10000),
		       COALESCE(NULLIF(p.contract_size, 0), sym.contract_size), p.price_open,
		       sym.currency_base, sym.currency_profit, u."group" ILIKE 'coverage%'
		  FROM hst.positions p
		  JOIN hst.users u ON u.login = p.login
		  JOIN hst.symbols sym ON sym.symbol = p.symbol
		 WHERE `+where, args...)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	type sides struct{ clients, coverage float64 }
	byCurrency := map[string]*sides{}
	add := func(currency string, amount float64, coverage bool) {
		currency = strings.ToUpper(currency)
		if currency == "" {
			return
		}
		v := byCurrency[currency]
		if v == nil {
			v = &sides{}
			byCurrency[currency] = v
		}
		if coverage {
			v.coverage += amount
		} else {
			v.clients += amount
		}
	}

	for rows.Next() {
		var action int32
		var volExt int64
		var contractSize, priceOpen float64
		var base, profit string
		var coverage bool
		if err := rows.Scan(&action, &volExt, &contractSize, &priceOpen, &base, &profit, &coverage); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		units := model.ExtToLots(volExt) * contractSize
		if model.OrderType(action) == model.OrderType_sell {
			units = -units
		}
		add(base, units, coverage)
		// profit leg valued at open price; current price needs a per-symbol quote lookup
		add(profit, -units*priceOpen, coverage)
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}

	report := ViewExposureReport{Currencies: []ViewExposure{}}
	for currency, v := range byCurrency {
		rate := s.usdRate(c.UserContext(), currency)
		net := v.clients + v.coverage
		report.Currencies = append(report.Currencies, ViewExposure{
			Currency: currency, Clients: v.clients, Coverage: v.coverage,
			Net: net, Rate: rate, NetUsd: net * rate,
		})
		report.Totals.NetUsd += net * rate
	}
	sort.Slice(report.Currencies, func(i, j int) bool {
		return math.Abs(report.Currencies[i].NetUsd) > math.Abs(report.Currencies[j].NetUsd)
	})

	return s.App.HttpResponseOK(c, report)
}

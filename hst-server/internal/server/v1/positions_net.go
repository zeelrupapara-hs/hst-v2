package v1

import (
	"sort"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
)

// ViewNetSide is one side of the book on one instrument.
type ViewNetSide struct {
	Volume    float64 `json:"volume"`
	PriceOpen float64 `json:"price_open"`
	Value     float64 `json:"value"`
	Profit    float64 `json:"profit"`
	Positions int     `json:"positions"`
}

// ViewNetPosition is the caller's exposure on one instrument, both sides and the net between them.
type ViewNetPosition struct {
	Symbol       string      `json:"symbol"`
	Buy          ViewNetSide `json:"buy"`
	Sell         ViewNetSide `json:"sell"`
	NetVolume    float64     `json:"net_volume"`
	NetValue     float64     `json:"net_value"`
	Profit       float64     `json:"profit"`
	PriceCurrent float64     `json:"price_current"`
	Positions    int         `json:"positions"`
}

// netAccum sums in extended volume units so the lot rounding happens once, at the end.
type netAccum struct {
	buyExt, sellExt       int64
	buyNotional, sellNoti float64
	buyProfit, sellProfit float64
	buyValue, sellValue   float64
	buyCount, sellCount   int
	priceCurrent          float64
}

// GetMyNetPositions aggregates the calling account's open positions by symbol and side.
//
//	@Id			GetMyNetPositions
//	@Tags		Trader
//	@Produce	json
//	@Success	200	{object}	Response{data=[]ViewNetPosition}
//	@Failure	403	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/positions/net [get]
func (s *HttpServer) GetMyNetPositions(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	out, err := s.readPositions(c.UserContext(), "p.login = $1", []any{snap.Login})
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// the net must agree with the Positions tab on the same screen, so aggregate the live values
	s.overlayLive(c.UserContext(), snap.Login, out)

	return s.App.HttpResponseOK(c, netBySymbol(out))
}

// netBySymbol folds positions into one row per instrument, sorted by symbol.
func netBySymbol(in []ViewPosition) []ViewNetPosition {
	by := map[string]*netAccum{}

	for i := range in {
		p := &in[i]
		a := by[p.Symbol]
		if a == nil {
			a = &netAccum{}
			by[p.Symbol] = a
		}

		ext := model.ExtendedVolume(p.VolumeUnits, p.VolumeExt)
		lots := model.ExtToLots(ext)
		a.priceCurrent = p.PriceCurrent

		if p.Action.IsBuy() {
			a.buyExt += ext
			a.buyNotional += p.PriceOpen * lots
			a.buyValue += p.PriceCurrent * lots
			a.buyProfit += p.Profit
			a.buyCount++
			continue
		}
		a.sellExt += ext
		a.sellNoti += p.PriceOpen * lots
		a.sellValue += p.PriceCurrent * lots
		a.sellProfit += p.Profit
		a.sellCount++
	}

	out := make([]ViewNetPosition, 0, len(by))
	for symbol, a := range by {
		buyLots := model.ExtToLots(a.buyExt)
		sellLots := model.ExtToLots(a.sellExt)
		out = append(out, ViewNetPosition{
			Symbol:       symbol,
			Buy:          ViewNetSide{Volume: buyLots, PriceOpen: avg(a.buyNotional, buyLots), Value: a.buyValue, Profit: a.buyProfit, Positions: a.buyCount},
			Sell:         ViewNetSide{Volume: sellLots, PriceOpen: avg(a.sellNoti, sellLots), Value: a.sellValue, Profit: a.sellProfit, Positions: a.sellCount},
			NetVolume:    model.ExtToLots(a.buyExt - a.sellExt),
			NetValue:     a.buyValue - a.sellValue,
			Profit:       a.buyProfit + a.sellProfit,
			PriceCurrent: a.priceCurrent,
			Positions:    a.buyCount + a.sellCount,
		})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Symbol < out[j].Symbol })
	return out
}

func avg(notional, lots float64) float64 {
	if lots == 0 {
		return 0
	}
	return notional / lots
}

package handler

import (
	"hstcore/internal/book"
	"hstcore/internal/settings"
	"hstcore/model"
)

type marginExposure struct {
	symbol  string
	lots    float64
	base    float64
	isOrder bool
}

type accountMarginBreakdown struct {
	bySymbol            map[string]float64
	maintenanceBySymbol map[string]float64
	pendingInitial      float64
	pendingMaintenance  float64
}

func (b accountMarginBreakdown) totalInitial() float64 {
	var n float64
	for _, v := range b.bySymbol {
		n += v
	}
	return n + b.pendingInitial
}

func (b accountMarginBreakdown) totalMaintenance() float64 {
	var n float64
	for _, v := range b.maintenanceBySymbol {
		n += v
	}
	return n + b.pendingMaintenance
}

type floatingRuleKey struct {
	path string
	mode int32
}

// accountMarginBreakdown computes initial and maintenance margin for the whole account.
//
// extra adds one hypothetical order exposure for pre-trade checks. Floating leverage tiers are
// applied per rule: aggregate volume across all matching symbols, or per symbol separately.
func (h *Handler) accountMarginBreakdown(e *book.Entry, extra *marginExposure) accountMarginBreakdown {
	out := accountMarginBreakdown{
		bySymbol:            make(map[string]float64, 8),
		maintenanceBySymbol: make(map[string]float64, 8),
	}

	g, ok := h.Settings.Group(e.Account.Group)
	if !ok {
		return out
	}

	exposures := h.collectMarginExposures(e, extra)
	plain := make([]marginExposure, 0, len(exposures))
	byRule := make(map[floatingRuleKey][]marginExposure)

	for _, exp := range exposures {
		r, ok := h.Settings.For(e.Account.Group, exp.symbol)
		if !ok || r.Symbol == nil {
			continue
		}

		rule := model.MatchLeverageRule(g, exp.symbol, r.Symbol.Path)
		if rule == nil {
			plain = append(plain, exp)
			continue
		}

		key := floatingRuleKey{path: rule.Path, mode: rule.RangeMode}
		byRule[key] = append(byRule[key], exp)
	}

	for _, exp := range plain {
		r, ok := h.Settings.For(e.Account.Group, exp.symbol)
		if !ok {
			continue
		}
		if exp.isOrder {
			out.pendingInitial += exp.base
			out.pendingMaintenance += h.orderMaintenanceMargin(e, exp, r)
			continue
		}
		out.bySymbol[exp.symbol] = h.symbolBaseMargin(e, exp.symbol, r, false)
		out.maintenanceBySymbol[exp.symbol] = h.symbolBaseMargin(e, exp.symbol, r, true)
	}

	for key, group := range byRule {
		rule := h.findLeverageRule(g, key)
		if rule == nil {
			for _, exp := range group {
				plain = append(plain, exp)
			}
			continue
		}

		if rule.RangeMode == model.RangeModeVolumePerSymbol {
			h.applyFloatingPerSymbol(g, rule, group, &out)
		} else {
			h.applyFloatingAggregate(rule, group, &out)
		}
	}

	return out
}

func (h *Handler) findLeverageRule(g *model.Group, key floatingRuleKey) *model.LeverageRule {
	for _, r := range g.LeverageRules {
		if r.Path == key.path && r.RangeMode == key.mode {
			return r
		}
	}
	return nil
}

func (h *Handler) collectMarginExposures(e *book.Entry, extra *marginExposure) []marginExposure {
	seen := make(map[string]bool, len(e.Positions))
	out := make([]marginExposure, 0, len(e.Positions)+len(e.Orders)+1)

	for _, p := range e.Positions {
		if seen[p.Symbol] {
			continue
		}
		seen[p.Symbol] = true

		r, ok := h.Settings.For(e.Account.Group, p.Symbol)
		if !ok {
			continue
		}

		var lots int64
		for _, pos := range e.Positions {
			if pos.Symbol == p.Symbol {
				lots += pos.Volume
			}
		}

		out = append(out, marginExposure{
			symbol:  p.Symbol,
			lots:    model.Lots(lots),
			base:    h.symbolBaseMargin(e, p.Symbol, r, false),
			isOrder: false,
		})
	}

	for _, o := range e.Orders {
		kind := o.Kind()
		if !kind.IsPending() || !o.State.IsLive() {
			continue
		}

		r, ok := h.Settings.For(e.Account.Group, o.Symbol)
		if !ok || r.MarginRate.For(kind) <= 0 {
			continue
		}

		price := o.PriceOrder
		if price <= 0 {
			continue
		}

		lots := model.Lots(o.VolumeCurrent)
		out = append(out, marginExposure{
			symbol:  o.Symbol,
			lots:    lots,
			base:    MarginForTypePlain(r, lots, price, e.Account.Leverage, o.RateMargin, kind, false),
			isOrder: true,
		})
	}

	if extra != nil && extra.lots > 0 {
		out = append(out, *extra)
	}

	return out
}

func (h *Handler) symbolBaseMargin(e *book.Entry, symbol string, r *settings.Rules, maintenance bool) float64 {
	return h.MarginForSymbol(e, symbol, r, maintenance)
}

func (h *Handler) orderMaintenanceMargin(e *book.Entry, exp marginExposure, r *settings.Rules) float64 {
	for _, o := range e.Orders {
		if o.Symbol != exp.symbol {
			continue
		}
		kind := o.Kind()
		if !kind.IsPending() || !o.State.IsLive() {
			continue
		}
		price := o.PriceOrder
		if price <= 0 {
			continue
		}
		if model.Lots(o.VolumeCurrent) == exp.lots {
			return MarginForTypePlain(r, exp.lots, price, e.Account.Leverage, o.RateMargin, kind, true)
		}
	}
	return exp.base
}

func (h *Handler) applyFloatingAggregate(rule *model.LeverageRule, group []marginExposure,
	out *accountMarginBreakdown) {

	var totalLots, totalBase float64
	for _, exp := range group {
		totalLots += exp.lots
		totalBase += exp.base
	}
	if totalLots <= 0 || totalBase <= 0 {
		return
	}

	marginPerLot := totalBase / totalLots
	initial := model.MarginInTiers(totalLots, marginPerLot, rule.Tiers, false)
	maintenance := model.MarginInTiers(totalLots, marginPerLot, rule.Tiers, true)
	h.allocateFloatingMargin(group, totalBase, initial, maintenance, out)
}

func (h *Handler) applyFloatingPerSymbol(_ *model.Group, rule *model.LeverageRule,
	group []marginExposure, out *accountMarginBreakdown) {

	bySymbol := make(map[string][]marginExposure, 4)
	for _, exp := range group {
		bySymbol[exp.symbol] = append(bySymbol[exp.symbol], exp)
	}

	for _, items := range bySymbol {
		var totalLots, totalBase float64
		for _, exp := range items {
			totalLots += exp.lots
			totalBase += exp.base
		}
		if totalLots <= 0 || totalBase <= 0 {
			continue
		}

		marginPerLot := totalBase / totalLots
		initial := model.MarginInTiers(totalLots, marginPerLot, rule.Tiers, false)
		maintenance := model.MarginInTiers(totalLots, marginPerLot, rule.Tiers, true)
		h.allocateFloatingMargin(items, totalBase, initial, maintenance, out)
	}
}

func (h *Handler) allocateFloatingMargin(group []marginExposure, totalBase, initial, maintenance float64,
	out *accountMarginBreakdown) {

	if totalBase <= 0 {
		return
	}

	for _, exp := range group {
		share := exp.base / totalBase
		mInitial := initial * share
		mMaint := maintenance * share

		if exp.isOrder {
			out.pendingInitial += mInitial
			out.pendingMaintenance += mMaint
			continue
		}

		out.bySymbol[exp.symbol] += mInitial
		out.maintenanceBySymbol[exp.symbol] += mMaint
	}
}

// hypotheticalOrderExposure is the margin exposure a not-yet-booked order would add.
func (h *Handler) hypotheticalOrderExposure(e *book.Entry, o *model.Order, r *settings.Rules,
	t model.Tick) *marginExposure {

	opening := h.openingVolume(e, o, r)
	if opening <= 0 || r.MarginRate.For(o.Kind()) <= 0 {
		return nil
	}

	price := o.PriceOrder
	if price <= 0 {
		price = t.OpenPrice(o.Kind().IsBuy())
	}

	lots := model.Lots(opening)
	return &marginExposure{
		symbol:  o.Symbol,
		lots:    lots,
		base:    MarginForTypePlain(r, lots, price, e.Account.Leverage, o.RateMargin, o.Kind(), false),
		isOrder: true,
	}
}

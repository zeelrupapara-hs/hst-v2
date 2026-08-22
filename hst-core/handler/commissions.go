package handler

import (
	"context"
	"strings"

	"hstcore/internal/settings"
	"hstcore/model"
	"hstcore/pkg/logger"
)

// Charging commission on a deal.

// Commission is one rule from hst.commissions with its tiers.
type Commission struct {
	CommissionId int64
	GroupId      int64
	Name         string
	Path         string
	Mode         int32
	RangeMode    int32
	ChargeMode   int32
	EntryMode    int32
	ActionMode   int32
	ProfitMode   int32
	ReasonFlags  int32
	Currency     string

	Tiers []CommissionTier
}

// CommissionTier is one band of a commission.
type CommissionTier struct {
	TierId    int64
	Mode      int32
	Type      int32
	Value     float64
	RangeFrom float64
	RangeTo   float64
	Minimal   float64
	Currency  string
}

// How a tier's value is read.
const (
	tierPerTrade  = 0
	tierPerVolume = 1
)

// What the tier bands are measured in.
const (
	rangeVolume         = 0
	rangeTurnoverMoney  = 1
	rangeTurnoverVolume = 2
)

// When a commission is taken.
const (
	chargeDaily   = 0
	chargeMonthly = 1
	chargeInstant = 2
)

// Which deals a commission covers.
const (
	entryAll = 0
	entryIn  = 1
	entryOut = 2

	actionAll  = 0
	actionBuy  = 1
	actionSell = 2

	profitAll    = 0
	profitProfit = 1
	profitLoss   = 2
)

// CommissionFor is what one deal costs, in the deposit currency.
func (h *Handler) CommissionFor(d *model.Deal, r *settings.Rules, a *model.Account) float64 {
	var total float64

	h.mu.RLock()
	list := h.commissions[r.Group.GroupId]
	h.mu.RUnlock()

	for i := range list {
		c := &list[i]

		if !c.covers(d, r) {
			continue
		}
		if c.ChargeMode != chargeInstant {
			// gathered by the end-of-period job instead
			continue
		}

		amount, currency := c.charge(d)
		total += h.inDeposit(amount, currency, a)
	}

	return total
}

// inDeposit converts a commission from the currency it was written in to the account's.
func (h *Handler) inDeposit(amount float64, currency string, a *model.Account) float64 {
	if amount == 0 || currency == "" || currency == a.Currency {
		return amount
	}

	rate, ok := h.crossRate(a.Group, currency, a.Currency, true)
	if !ok {
		h.Log.Log(logger.TypeTrade, logger.CodeWarn, "no conversion rate, commission charged 1:1",
			"from", currency, "to", a.Currency, "login", a.Login)
		return amount
	}

	return NormalisePrice(amount*rate, a.CurrencyDigits)
}

// covers reports whether this commission applies to the deal.
func (c *Commission) covers(d *model.Deal, r *settings.Rules) bool {
	if !pathCovers(c.Path, r) {
		return false
	}

	switch c.EntryMode {
	case entryIn:
		if model.DealEntry(d.Entry) != model.DealEntry_in {
			return false
		}
	case entryOut:
		if model.DealEntry(d.Entry) != model.DealEntry_out {
			return false
		}
	}

	switch c.ActionMode {
	case actionBuy:
		if model.DealAction(d.Action) != model.DealAction_buy {
			return false
		}
	case actionSell:
		if model.DealAction(d.Action) != model.DealAction_sell {
			return false
		}
	}

	switch c.ProfitMode {
	case profitProfit:
		if d.Profit <= 0 {
			return false
		}
	case profitLoss:
		if d.Profit >= 0 {
			return false
		}
	}

	// a reason mask of zero covers every reason
	if c.ReasonFlags != 0 && c.ReasonFlags&reasonFlagFor(model.OrderReason(d.Reason)) == 0 {
		return false
	}

	return true
}

// charge works out what the deal costs under this commission, and the currency that is in.
func (c *Commission) charge(d *model.Deal) (float64, string) {
	lots := model.Lots(d.Volume)

	tier := c.tierFor(lots, d)
	if tier == nil {
		return 0, ""
	}

	var amount float64

	switch tier.Type {
	case tierPerVolume:
		amount = tier.Value * lots
	default:
		amount = tier.Value
	}

	// a minimum charge, so a tiny trade still pays its way
	if tier.Minimal > 0 && amount < tier.Minimal {
		amount = tier.Minimal
	}

	return amount, c.currencyOf(tier)
}

// currencyOf is the currency a tier is written in, falling back to the commission's own.
func (c *Commission) currencyOf(t *CommissionTier) string {
	if t.Currency != "" {
		return t.Currency
	}
	return c.Currency
}

// tierFor picks the band the trade falls into.
func (c *Commission) tierFor(lots float64, d *model.Deal) *CommissionTier {
	// what the bands are measured in
	measure := lots
	if c.RangeMode == rangeTurnoverMoney {
		measure = lots * d.ContractSize * d.Price
	}

	for i := range c.Tiers {
		t := &c.Tiers[i]

		// a band with no upper edge runs to infinity, which is how the last one is written
		if measure >= t.RangeFrom && (t.RangeTo <= 0 || measure < t.RangeTo) {
			return t
		}
	}

	return nil
}

// pathCovers reports whether a commission's path covers the instrument.
func pathCovers(path string, r *settings.Rules) bool {
	switch {
	case path == "" || path == "*":
		return true
	case strings.HasSuffix(path, "*"):
		prefix := strings.TrimSuffix(path, "*")
		return strings.HasPrefix(r.Symbol.Path, prefix) || strings.HasPrefix(r.Symbol.Symbol, prefix)
	default:
		return path == r.Symbol.Symbol || path == r.Symbol.Path
	}
}

// reasonFlagFor maps an order reason onto the commission's reason mask.
func reasonFlagFor(reason model.OrderReason) int32 {
	switch reason {
	case model.OrderReason_client:
		return 0x0001
	case model.OrderReason_expert:
		return 0x0002
	case model.OrderReason_dealer:
		return 0x0004
	case model.OrderReason_mobile:
		return 0x0010
	case model.OrderReason_web:
		return 0x0020
	}
	return 0x0001
}

// loadCommissions reads every group's commissions and their tiers.
func (h *Handler) LoadCommission(ctx context.Context) error {
	rows, err := h.DB.DB.Query(ctx,
		`SELECT commission_id, group_id, name, path, mode, mode_range, mode_charge,
		        mode_entry, mode_action, mode_profit, mode_reason, turnover_currency
		   FROM hst.commissions
		  ORDER BY group_id, commission_id`)
	if err != nil {
		return err
	}

	byGroup := make(map[int64][]Commission, 64)
	index := make(map[int64]*Commission, 128)

	for rows.Next() {
		var c Commission
		if err := rows.Scan(&c.CommissionId, &c.GroupId, &c.Name, &c.Path, &c.Mode,
			&c.RangeMode, &c.ChargeMode, &c.EntryMode, &c.ActionMode, &c.ProfitMode,
			&c.ReasonFlags, &c.Currency); err != nil {
			rows.Close()
			return err
		}
		byGroup[c.GroupId] = append(byGroup[c.GroupId], c)
	}
	rows.Close()
	if rows.Err() != nil {
		return rows.Err()
	}

	// index after the slices are final, or the pointers would move as they grew
	for gid := range byGroup {
		for i := range byGroup[gid] {
			index[byGroup[gid][i].CommissionId] = &byGroup[gid][i]
		}
	}

	rows, err = h.DB.DB.Query(ctx,
		`SELECT tier_id, commission_id, mode, type, value, range_from, range_to, minimal, currency
		   FROM hst.commissions_tiers ORDER BY commission_id, range_from`)
	if err != nil {
		return err
	}

	for rows.Next() {
		var commissionId int64
		var t CommissionTier
		if err := rows.Scan(&t.TierId, &commissionId, &t.Mode, &t.Type, &t.Value,
			&t.RangeFrom, &t.RangeTo, &t.Minimal, &t.Currency); err != nil {
			rows.Close()
			return err
		}
		if c, ok := index[commissionId]; ok {
			c.Tiers = append(c.Tiers, t)
		}
	}
	rows.Close()
	if rows.Err() != nil {
		return rows.Err()
	}

	h.mu.Lock()
	h.commissions = byGroup
	h.mu.Unlock()

	h.Log.Log(logger.TypeCfg, logger.CodeOK, "commissions loaded", "groups", len(byGroup))

	return nil
}

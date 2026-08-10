package model

import "strings"

// Notional range modes are stored but not yet applied by the engine.
const (
	RangeModeNotionalValue         int32 = 2
	RangeModeNotionalValuePerSymbol int32 = 3
)

// FloatingLeverageApplies reports whether a group may use floating leverage profiles.
func FloatingLeverageApplies(g *Group) bool {
	if g == nil || len(g.LeverageRules) == 0 {
		return false
	}
	return MarginMode(g.MarginMode) != MarginMode_exchange
}

// MatchLeverageRule returns the first volume-mode rule covering an instrument.
func MatchLeverageRule(g *Group, symbol, path string) *LeverageRule {
	if !FloatingLeverageApplies(g) {
		return nil
	}
	for _, r := range g.LeverageRules {
		if r.RangeMode != RangeModeVolume && r.RangeMode != RangeModeVolumePerSymbol {
			continue
		}
		if PathCovers(r.Path, symbol, path) {
			return r
		}
	}
	return nil
}

// PathCovers reports whether a mask covers an instrument, by name or by its place in the tree.
func PathCovers(mask, symbol, path string) bool {
	switch {
	case mask == "" || mask == "*":
		return true
	case strings.HasSuffix(mask, "*"):
		prefix := strings.TrimSuffix(mask, "*")
		return strings.HasPrefix(path, prefix) || strings.HasPrefix(symbol, prefix)
	default:
		return mask == symbol || mask == path
	}
}

// MarginInTiers splits totalLots across tiers and returns the weighted margin total.
//
// marginPerLot is the base margin for one lot before tier rates are applied. The last tier
// uses RangeTo 0 to mean infinity.
func MarginInTiers(totalLots, marginPerLot float64, tiers []LeverageTier, maintenance bool) float64 {
	if totalLots <= 0 || marginPerLot <= 0 || len(tiers) == 0 {
		return 0
	}

	var (
		total     float64
		remaining = totalLots
		cursor    float64
	)

	for i := range tiers {
		if remaining <= 0 {
			break
		}

		t := &tiers[i]
		var tierLots float64
		if t.RangeTo <= 0 || i == len(tiers)-1 {
			tierLots = remaining
		} else {
			tierLots = t.RangeTo - cursor
			if tierLots > remaining {
				tierLots = remaining
			}
		}

		rate := t.MarginRateInitial
		if maintenance {
			rate = t.MarginRateMaintenance
		}
		total += tierLots * marginPerLot * rate

		remaining -= tierLots
		if t.RangeTo > 0 {
			cursor = t.RangeTo
		}
	}

	return total
}

package model

import "testing"

func TestMarginInTiers_referenceVolumeExample(t *testing.T) {
	tiers := []LeverageTier{
		{RangeTo: 10, MarginRateInitial: 1, MarginRateMaintenance: 1},
		{RangeTo: 20, MarginRateInitial: 2, MarginRateMaintenance: 2},
		{RangeTo: 0, MarginRateInitial: 3, MarginRateMaintenance: 3},
	}

	got := MarginInTiers(22, 1000, tiers, false)
	const want = 36000.0
	if got != want {
		t.Fatalf("MarginInTiers(22 lots) = %v, want %v", got, want)
	}
}

func TestMarginInTiers_referenceMixedSymbols(t *testing.T) {
	tiers := []LeverageTier{
		{RangeTo: 10, MarginRateInitial: 1, MarginRateMaintenance: 1},
		{RangeTo: 0, MarginRateInitial: 2, MarginRateMaintenance: 2},
	}

	const marginPerLot = (1800.0 + 1000.0 + 15000.0) / (2 + 1 + 15)
	got := MarginInTiers(18, marginPerLot, tiers, false)
	want := 10*marginPerLot*1 + 8*marginPerLot*2
	if got != want {
		t.Fatalf("MarginInTiers(18 lots) = %v, want %v", got, want)
	}
}

func TestMarginInTiers_zeroMaintenanceRate(t *testing.T) {
	tiers := []LeverageTier{
		{RangeTo: 10, MarginRateInitial: 1, MarginRateMaintenance: 0},
		{RangeTo: 0, MarginRateInitial: 2, MarginRateMaintenance: 1},
	}

	got := MarginInTiers(15, 1000, tiers, true)
	const want = 5000.0 // only the 5 lots above 10 at maintenance rate 1
	if got != want {
		t.Fatalf("maintenance MarginInTiers = %v, want %v", got, want)
	}
}

package seed

import (
	"context"
	"errors"
	"time"

	"hstserver/pkg/logger"

	"github.com/jackc/pgx/v5"
)

// SeedLeverages creates the default profile on a fresh install and always ensures the
// High/Low demo profiles exist for margin-call testing.
func (s *Seeder) SeedLeverages(ctx context.Context) error {
	var n int
	if err := s.DB.DB.QueryRow(ctx, `SELECT count(*) FROM hst.leverages`).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		if err := s.insertLeverageProfile(ctx, "Default", defaultForexTiers()); err != nil {
			return err
		}
	}
	return s.ensureLeverageDemoProfiles(ctx)
}

type leverageTier struct {
	from, to       float64
	initial, maint float64
}

func defaultForexTiers() []leverageTier {
	return []leverageTier{{0, 0, 1, 1}}
}

// highLeverageTiers keeps margin multipliers at 1× — minimal floating margin on all volume.
func highLeverageTiers() []leverageTier {
	return []leverageTier{{0, 0, 1, 1}}
}

// lowLeverageTiers simulates tightening leverage as volume grows: first 10 lots at 1× (≈1:100
// account leverage unchanged), then 10+ lots at 100× base margin (≈1:1 effective).
func lowLeverageTiers() []leverageTier {
	return []leverageTier{
		{0, 10, 1, 1},
		{10, 0, 100, 100},
	}
}

func (s *Seeder) ensureLeverageDemoProfiles(ctx context.Context) error {
	for _, p := range []struct {
		name  string
		tiers []leverageTier
	}{
		{"High leverage", highLeverageTiers()},
		{"Low leverage", lowLeverageTiers()},
	} {
		exists, err := s.leverageExists(ctx, p.name)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		if err := s.insertLeverageProfile(ctx, p.name, p.tiers); err != nil {
			return err
		}
		s.Log.Log(logger.TypeSys, logger.CodeOK, "seed leverage profile created", "name", p.name)
	}
	return nil
}

func (s *Seeder) leverageExists(ctx context.Context, name string) (bool, error) {
	var id int64
	err := s.DB.DB.QueryRow(ctx,
		`SELECT leverage_id FROM hst.leverages WHERE name = $1`, name).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (s *Seeder) insertLeverageProfile(ctx context.Context, name string, tiers []leverageTier) error {
	now := time.Now().UnixNano()

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var leverageID int64
	if err := tx.QueryRow(ctx,
		`INSERT INTO hst.leverages (name, "timestamp", flags)
		 VALUES ($1, $2, 0) RETURNING leverage_id`,
		name, now).Scan(&leverageID); err != nil {
		return err
	}

	var ruleID int64
	if err := tx.QueryRow(ctx,
		`INSERT INTO hst.leverage_rules
		 (leverage_id, name, description, path, range_mode, range_value_currency,
		  range_value_currency_digits, config_index)
		 VALUES ($1, $2, $3, $4, 0, '', 2, 0) RETURNING rule_id`,
		leverageID, "Forex", name+" forex volume rule", `Forex\*`).Scan(&ruleID); err != nil {
		return err
	}

	for _, t := range tiers {
		if _, err := tx.Exec(ctx,
			`INSERT INTO hst.leverage_tiers
			 (rule_id, range_from, range_to, margin_rate_initial, margin_rate_maintenance)
			 VALUES ($1, $2, $3, $4, $5)`,
			ruleID, t.from, t.to, t.initial, t.maint); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

package seed

import (
	"context"
	"time"

	"hstserver/pkg/logger"
)

// SeedLeverages creates a default floating-leverage profile on a fresh install.
func (s *Seeder) SeedLeverages(ctx context.Context) error {
	var n int
	if err := s.DB.DB.QueryRow(ctx, `SELECT count(*) FROM hst.leverages`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}

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
		"Default", now).Scan(&leverageID); err != nil {
		return err
	}

	var ruleID int64
	if err := tx.QueryRow(ctx,
		`INSERT INTO hst.leverage_rules
		 (leverage_id, name, description, path, range_mode, range_value_currency,
		  range_value_currency_digits, config_index)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, 0) RETURNING rule_id`,
		leverageID, "Forex", "Default forex floating leverage", `Forex\*`, 0, "", 2).Scan(&ruleID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO hst.leverage_tiers
		 (rule_id, range_from, range_to, margin_rate_initial, margin_rate_maintenance)
		 VALUES ($1, 0, 0, 1, 1)`,
		ruleID); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	s.Log.Log(logger.TypeSys, logger.CodeOK, "seed leverage profile created",
		"name", "Default", "leverage_id", leverageID)

	return nil
}

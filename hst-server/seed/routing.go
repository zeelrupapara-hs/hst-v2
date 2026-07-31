package seed

import (
	"context"
	"time"
)

// SeedRouting fills an empty routing table with starter non-dealer rules.
func (s *Seeder) SeedRouting(ctx context.Context) error {
	var rules int
	if err := s.DB.DB.QueryRow(ctx, `SELECT count(*) FROM hst.routing`).Scan(&rules); err != nil {
		return err
	}
	if rules > 0 {
		return nil
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	now := time.Now().UnixNano()

	// confirm small instant/market requests at the server
	var confirmID int64
	if err := tx.QueryRow(ctx,
		`INSERT INTO hst.routing
		   (name, mode, request, type, flags, action, action_value,
		    routing_index, date_created, date_modified)
		 VALUES ($1,$2,$3,$4,0,$5,'',$6,$7,$7)
		 RETURNING routing_id`,
		"Auto confirm small market",
		1,
		0x00000004|0x00000008, // instant + market
		0x0001|0x0002,         // buy + sell
		int32(1006),           // confirm at market
		0, now).Scan(&confirmID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO hst.routing_conds (routing_id, condition, rule, value)
		 VALUES ($1, 2, 4, '1000000')`,
		confirmID); err != nil {
		return err
	}

	// reject everything while the symbol is in gap mode
	if _, err := tx.Exec(ctx,
		`INSERT INTO hst.routing
		   (name, mode, request, type, flags, action, action_value,
		    routing_index, date_created, date_modified)
		 VALUES ($1,$2,0,0,0,$3,$4,$5,$6,$6)`,
		"Reject during gap",
		1,
		int32(1003), "Market closed",
		1, now); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO hst.routing_conds (routing_id, condition, rule, value)
		 VALUES ((SELECT routing_id FROM hst.routing WHERE name = $1), 12, 0, '1')`,
		"Reject during gap"); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

package seed

import (
	"context"
	"time"

	"hstserver/pkg/logger"
)

// startingGroups are the trees a fresh install needs before anybody can be put in one.
var startingGroups = []struct {
	Group    string
	Currency string
	Company  string
	// Deposit and Leverage are what an account in the tree opens with.
	Deposit  *float64
	Leverage *int32
	// Call and StopOut are the margin levels, as a percentage of the margin in use.
	Call    float64
	StopOut float64
}{
	{Group: `demo`, Currency: "USD", Company: "HS Trading", Deposit: ptr(100000.0), Leverage: ptr(int32(100)), Call: 100, StopOut: 50},
	{Group: `preliminary`, Currency: "USD", Company: "HS Trading", Call: 100, StopOut: 50},
	{Group: `real`, Currency: "USD", Company: "HS Trading", Call: 100, StopOut: 50},
	{Group: `managers\admin`, Currency: "USD", Company: "HS Trading", Call: 100, StopOut: 50},
}

func ptr[T any](v T) *T { return &v }

// SeedGroups creates the starting trees, and is a no-op once they are there.
func (s *Seeder) SeedGroups(ctx context.Context) error {
	now := time.Now().UnixNano()

	for _, g := range startingGroups {
		tag, err := s.DB.DB.Exec(ctx,
			`INSERT INTO hst.groups ("group", currency, company, demo_deposit, demo_leverage,
			                         margin_call, margin_stop_out, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 ON CONFLICT ("group") DO NOTHING`,
			g.Group, g.Currency, g.Company, g.Deposit, g.Leverage, g.Call, g.StopOut, now)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			continue
		}

		s.Log.Log(logger.TypeSys, logger.CodeOK, "seed group created", "group", g.Group)
	}

	return nil
}

// SeedGroupSymbols gives every group a "*" entry, so a fresh install can trade.
func (s *Seeder) SeedGroupSymbols(ctx context.Context) error {
	var rows int
	if err := s.DB.DB.QueryRow(ctx, `SELECT count(*) FROM hst.groups_symbols`).Scan(&rows); err != nil {
		return err
	}
	if rows > 0 {
		return nil
	}

	tag, err := s.DB.DB.Exec(ctx,
		`INSERT INTO hst.groups_symbols (group_id, path, config_index, updated_at)
		 SELECT group_id, '*', 0, $1 FROM hst.groups`,
		time.Now().UnixNano())
	if err != nil {
		return err
	}

	s.Log.Log(logger.TypeSys, logger.CodeOK, "seed group symbols created",
		"groups", tag.RowsAffected(), "path", "*")

	return nil
}

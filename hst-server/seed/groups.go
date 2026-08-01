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
	// Deposit and Leverage are what an account in the tree opens with, and only the demo
	// tree opens on the house; nil is genuinely unset rather than zero.
	Deposit  *float64
	Leverage *int32
}{
	{Group: `demo`, Currency: "USD", Company: "HS Trading", Deposit: ptr(100000.0), Leverage: ptr(int32(100))},
	{Group: `preliminary`, Currency: "USD", Company: "HS Trading"},
	{Group: `real`, Currency: "USD", Company: "HS Trading"},
	{Group: `managers\admin`, Currency: "USD", Company: "HS Trading"},
}

func ptr[T any](v T) *T { return &v }

// SeedGroups creates the starting trees, and is a no-op once they are there.
//
// It runs before the administrator is seeded, because a login belongs to a group and the API
// refuses one that does not exist.
func (s *Seeder) SeedGroups(ctx context.Context) error {
	now := time.Now().UnixNano()

	for _, g := range startingGroups {
		tag, err := s.DB.DB.Exec(ctx,
			`INSERT INTO hst.groups ("group", currency, company, demo_deposit, demo_leverage, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 ON CONFLICT ("group") DO NOTHING`,
			g.Group, g.Currency, g.Company, g.Deposit, g.Leverage, now)
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

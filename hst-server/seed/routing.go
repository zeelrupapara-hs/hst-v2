package seed

import (
	"context"
	"time"

	"hstserver/model"
	"hstserver/pkg/logger"
)

// ruleCond is one line of a rule's "Where conditions are" table.
type ruleCond struct {
	Condition model.RouteCondition
	Rule      model.ConditionRule
	Value     string
}

// startingRules are the rules a fresh install needs to be able to trade at all.
//
// Order is everything: rules run top to bottom and the first terminal action wins, so a rule
// that refuses something must sit above the rule that would have accepted it. The last entry is
// the catch-all, and without it nothing executes — MT5 is explicit that a request matching no
// rule is not processed by the server.
var startingRules = []struct {
	Name    string
	Request model.RouteFlags
	Type    model.TypeFlags
	Action  model.RouteAction
	Value   string
	Conds   []ruleCond
}{
	{
		// a gap is not a market: refuse before anything below can confirm
		Name:   "Reject during gap",
		Action: model.RouteAction_reject,
		Value:  "Market gap",
		Conds:  []ruleCond{{model.RouteCondition_gap, model.ConditionRule_eq, "1"}},
	},
	{
		// anything not caught above executes at the market price. Deleting this rule stops all
		// trading on the server, which is why it ships enabled.
		Name:   "Automate other requests",
		Action: model.RouteAction_confirm_market,
	},
}

// SeedRouting fills an empty routing table with the starting rules, and is a no-op once any rule
// exists — a broker's own list is never rewritten.
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

	for i, r := range startingRules {
		var id int64
		if err := tx.QueryRow(ctx,
			`INSERT INTO hst.routing
			   (name, mode, request, type, flags, action, action_value,
			    routing_index, date_created, date_modified)
			 VALUES ($1, 1, $2, $3, 0, $4, $5, $6, $7, $7)
			 RETURNING routing_id`,
			r.Name, int32(r.Request), int32(r.Type), int32(r.Action), r.Value, i, now).
			Scan(&id); err != nil {
			return err
		}

		for _, c := range r.Conds {
			if _, err := tx.Exec(ctx,
				`INSERT INTO hst.routing_conds (routing_id, condition, rule, value)
				 VALUES ($1, $2, $3, $4)`,
				id, int32(c.Condition), int16(c.Rule), c.Value); err != nil {
				return err
			}
		}

		s.Log.Log(logger.TypeSys, logger.CodeOK, "seed routing rule created",
			"name", r.Name, "index", i)
	}

	return tx.Commit(ctx)
}

package v1

import (
	"context"
	"errors"
	"fmt"
	"time"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/logger"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// MT5 caps both of these at 1024.
const (
	maxLeverageProfiles = 1024
	maxLeverageRules    = 1024
)

// CrtLeverage creates a floating leverage configuration, optionally with its
// whole rule tree in one request.
type CrtLeverage struct {
	Name  string            `json:"name" validate:"required,max=128"`
	Flags int32             `json:"flags"`
	Rules []CrtLeverageRule `json:"rules" validate:"omitempty,max=1024,dive"`
}

// UptLeverage patches a configuration. Every field is a pointer, so an absent
// one keeps its value and an explicit one overwrites it. Rules is a pointer to
// a slice for the same reason: absent keeps the existing rules, [] clears them,
// and a populated array replaces them. leverage_id and timestamp are server
// owned and not settable.
type UptLeverage struct {
	Name  *string            `json:"name" validate:"omitempty,max=128"`
	Flags *int32             `json:"flags"`
	Rules *[]CrtLeverageRule `json:"rules" validate:"omitempty,max=1024,dive"`
}

// CrtLeverageRule is one rule with its levels. Rules arrive as whole documents,
// tiers have no endpoints of their own.
type CrtLeverageRule struct {
	Name                     string            `json:"name" validate:"required,max=128"`
	Description              string            `json:"description" validate:"max=255"`
	Path                     string            `json:"path" validate:"required,max=128"`
	RangeMode                int32             `json:"range_mode" validate:"gte=0,lte=3"`
	RangeValueCurrency       string            `json:"range_value_currency" validate:"omitempty,max=8"`
	RangeValueCurrencyDigits int32             `json:"range_value_currency_digits" validate:"gte=0,lte=8"`
	Tiers                    []CrtLeverageTier `json:"tiers" validate:"required,min=1,dive"`
}

// UptLeverageRule patches one rule. A non nil Tiers replaces the whole level set.
type UptLeverageRule struct {
	Name                     *string            `json:"name" validate:"omitempty,max=128"`
	Description              *string            `json:"description" validate:"omitempty,max=255"`
	Path                     *string            `json:"path" validate:"omitempty,max=128"`
	RangeMode                *int32             `json:"range_mode" validate:"omitempty,gte=0,lte=3"`
	RangeValueCurrency       *string            `json:"range_value_currency" validate:"omitempty,max=8"`
	RangeValueCurrencyDigits *int32             `json:"range_value_currency_digits" validate:"omitempty,gte=0,lte=8"`
	Tiers                    *[]CrtLeverageTier `json:"tiers" validate:"omitempty,min=1,dive"`
}

// CrtLeverageTier is one level. Only the upper bound is given: MT5 sets the
// minimum from the previous level, so the server derives range_from.
type CrtLeverageTier struct {
	RangeTo               float64 `json:"range_to" validate:"gte=0"`
	MarginRateInitial     float64 `json:"margin_rate_initial" validate:"gte=0"`
	MarginRateMaintenance float64 `json:"margin_rate_maintenance" validate:"gte=0"`
}

// ReorderLeverageRules carries the new evaluation order, most significant first.
type ReorderLeverageRules struct {
	RuleIds []int64 `json:"rule_ids" validate:"required,min=1"`
}

// ViewLeverageRule is a rule with its levels attached.
type ViewLeverageRule struct {
	model.LeverageRule
	Tiers []model.LeverageTier `json:"tiers"`
}

// ViewLeverageDetail is the whole tree, what the profile dialog renders.
type ViewLeverageDetail struct {
	model.Leverage
	Rules []ViewLeverageRule `json:"rules"`
}

// leveragesSortable are the real columns of leverages.
var leveragesSortable = utils.NewSortable("leverage_id", "name", "timestamp")

// generated from the structs so the select and the scan cannot drift apart
const leverageColumns = `leverage_id, name, "timestamp", flags`

const leverageRuleColumns = `rule_id, leverage_id, name, description, path,
	range_mode, range_value_currency, range_value_currency_digits, config_index`

const leverageTierColumns = `tier_id, rule_id, range_from, range_to,
	margin_rate_initial, margin_rate_maintenance`

// CreateLeverageProfile creates a configuration and its rule tree in one
// transaction.
//
//	@Id			CreateLeverageProfile
//	@Tags		Leverages
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CrtLeverage	true	"the profile to create, rules and tiers may be inlined"
//	@Success	201		{object}	Response{data=ViewLeverageDetail}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	409		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/leverage-profiles [post]
func (s *HttpServer) CreateLeverageProfile(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var body CrtLeverage
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}
	if err := validateRules(body.Rules); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	var profiles int
	if err := s.DB.DB.QueryRow(ctx, `SELECT count(*) FROM hst.leverages`).Scan(&profiles); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if profiles >= maxLeverageProfiles {
		return s.App.HttpResponseConflict(c, errs.ErrLeverageLimitReached)
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var id int64
	if err := tx.QueryRow(ctx,
		`INSERT INTO hst.leverages (name, "timestamp", flags)
		 VALUES ($1, $2, $3) RETURNING leverage_id`,
		body.Name, time.Now().UnixNano(), body.Flags).Scan(&id); err != nil {
		if isUniqueViolation(err) {
			return s.App.HttpResponseConflict(c, errs.ErrLeverageNameExists)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if err := insertRules(ctx, tx, id, 0, body.Rules); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "leverage profile created",
		"actor", snap.Login, "target", id, "rules", len(body.Rules))

	return s.getLeverageProfile(c, id, s.App.HttpResponseCreated)
}

// ListLeverageProfiles returns a page of configurations without their rules.
//
//	@Id			ListLeverageProfiles
//	@Tags		Leverages
//	@Produce	json
//	@Param		page	query		int		false	"page number, from 1"
//	@Param		limit	query		int		false	"rows per page, max 500"
//	@Param		search	query		string	false	"matches name"
//	@Param		sort_by	query		string	false	"leverage_id, name, timestamp"	Enums(leverage_id, name, timestamp)
//	@Param		order	query		string	false	"asc or desc"					Enums(asc, desc)
//	@Success	200		{object}	Response{data=[]model.Leverage}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/leverage-profiles [get]
func (s *HttpServer) ListLeverageProfiles(c *fiber.Ctx) error {
	q, err := utils.QueryFilter(c, leveragesSortable, "leverage_id")
	if err != nil {
		return s.App.HttpResponseBadQueryParams(c, err)
	}

	// sort_by is validated against an allowlist in QueryFilter; a bind
	// parameter cannot carry an ORDER BY clause
	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT `+leverageColumns+`
		   FROM hst.leverages
		  WHERE ($1 = '' OR name ILIKE '%'||$1||'%')
		  ORDER BY `+q.SortBy+`
		  LIMIT $2 OFFSET $3`, q.Search, q.Limit, q.Offset)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := []model.Leverage{}
	for rows.Next() {
		var v model.Leverage
		if err := rows.Scan(&v.LeverageId, &v.Name, &v.Timestamp, &v.Flags); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		out = append(out, v)
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}

	return s.App.HttpResponseOK(c, out)
}

// GetLeverageProfile returns one configuration with its rules and levels.
//
//	@Id			GetLeverageProfile
//	@Tags		Leverages
//	@Produce	json
//	@Param		id	path		int	true	"leverage profile id"
//	@Success	200	{object}	Response{data=ViewLeverageDetail}
//	@Failure	400	{object}	Response
//	@Failure	403	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/leverage-profiles/{id} [get]
func (s *HttpServer) GetLeverageProfile(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	return s.getLeverageProfile(c, int64(id), s.App.HttpResponseOK)
}

// UpdateLeverageProfile patches a configuration. A rules array replaces the
// whole tree, since the rules are ordered and only meaningful as a set.
//
//	@Id			UpdateLeverageProfile
//	@Tags		Leverages
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int			true	"leverage profile id"
//	@Param		body	body		UptLeverage	true	"only the fields to change. rules absent keeps them, [] clears them, a list replaces them"
//	@Success	200		{object}	Response{data=ViewLeverageDetail}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	409		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/leverage-profiles/{id} [put]
func (s *HttpServer) UpdateLeverageProfile(c *fiber.Ctx) error {
	ctx := c.UserContext()

	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	var body UptLeverage
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}
	if body.Rules != nil {
		if err := validateRules(*body.Rules); err != nil {
			return s.App.HttpResponseBadRequest(c, err)
		}
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx,
		`UPDATE hst.leverages SET
		    name        = COALESCE($2, name),
		    flags       = COALESCE($3, flags),
		    "timestamp" = $4
		  WHERE leverage_id = $1`,
		id, body.Name, body.Flags, time.Now().UnixNano())
	if err != nil {
		if isUniqueViolation(err) {
			return s.App.HttpResponseConflict(c, errs.ErrLeverageNameExists)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if tag.RowsAffected() == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	// the levels hang off the rules, so the cascade clears them too
	if body.Rules != nil {
		if _, err := tx.Exec(ctx,
			`DELETE FROM hst.leverage_rules WHERE leverage_id = $1`, id); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		if err := insertRules(ctx, tx, int64(id), 0, *body.Rules); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "leverage profile updated",
		"actor", snap.Login, "target", id, "rules_replaced", body.Rules != nil)

	return s.getLeverageProfile(c, int64(id), s.App.HttpResponseOK)
}

// DeleteLeverageProfile removes a configuration. Its rules and levels cascade.
//
//	@Id			DeleteLeverageProfile
//	@Tags		Leverages
//	@Produce	json
//	@Param		id	path		int	true	"leverage profile id"
//	@Success	204	{object}	Response
//	@Failure	400	{object}	Response
//	@Failure	403	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/leverage-profiles/{id} [delete]
func (s *HttpServer) DeleteLeverageProfile(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	tag, err := s.DB.DB.Exec(c.UserContext(),
		`DELETE FROM hst.leverages WHERE leverage_id = $1`, id)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if tag.RowsAffected() == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeWarn, "leverage profile deleted",
		"actor", snap.Login, "target", id)

	return s.App.HttpResponseNoContent(c)
}

// CreateLeverageRule appends a rule to a configuration. It lands last, so it is
// matched only after every existing rule.
//
//	@Id			CreateLeverageRule
//	@Tags		Leverages
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int				true	"leverage profile id"
//	@Param		body	body		CrtLeverageRule	true	"the rule to append, with its tiers"
//	@Success	201		{object}	Response{data=ViewLeverageDetail}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	409		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/leverage-profiles/{id}/rules [post]
func (s *HttpServer) CreateLeverageRule(c *fiber.Ctx) error {
	ctx := c.UserContext()

	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	var body CrtLeverageRule
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}
	if err := validateRules([]CrtLeverageRule{body}); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// lock the parent so two appends cannot pick the same config_index
	var exists bool
	err = tx.QueryRow(ctx,
		`SELECT true FROM hst.leverages WHERE leverage_id = $1 FOR UPDATE`, id).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	var next, count int32
	if err := tx.QueryRow(ctx,
		`SELECT COALESCE(MAX(config_index) + 1, 0), count(*)
		   FROM hst.leverage_rules WHERE leverage_id = $1`, id).Scan(&next, &count); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if count >= maxLeverageRules {
		return s.App.HttpResponseConflict(c, errs.ErrLeverageRuleLimitReached)
	}

	if err := insertRules(ctx, tx, int64(id), next, []CrtLeverageRule{body}); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if err := touchLeverage(ctx, tx, int64(id)); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "leverage rule created",
		"actor", snap.Login, "target", id, "config_index", next)

	return s.getLeverageProfile(c, int64(id), s.App.HttpResponseCreated)
}

// UpdateLeverageRule patches one rule. A tiers array replaces its whole level set.
//
//	@Id			UpdateLeverageRule
//	@Tags		Leverages
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int				true	"leverage profile id"
//	@Param		ruleId	path		int				true	"rule id, must belong to the profile"
//	@Param		body	body		UptLeverageRule	true	"only the fields to change. a tiers list replaces the whole level set"
//	@Success	200		{object}	Response{data=ViewLeverageDetail}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/leverage-profiles/{id}/rules/{ruleId} [put]
func (s *HttpServer) UpdateLeverageRule(c *fiber.Ctx) error {
	ctx := c.UserContext()

	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	ruleId, err := c.ParamsInt("ruleId")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	var body UptLeverageRule
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}
	if body.Tiers != nil {
		if err := validateTiers(*body.Tiers); err != nil {
			return s.App.HttpResponseBadRequest(c, err)
		}
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// the currency requirement depends on the mode, which may itself be changing
	var mode int32
	var currency string
	err = tx.QueryRow(ctx,
		`SELECT range_mode, range_value_currency
		   FROM hst.leverage_rules WHERE rule_id = $1 AND leverage_id = $2`,
		ruleId, id).Scan(&mode, &currency)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrRuleNotInProfile)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if body.RangeMode != nil {
		mode = *body.RangeMode
	}
	if body.RangeValueCurrency != nil {
		currency = *body.RangeValueCurrency
	}
	if model.RangeMode(mode).NeedsCurrency() && currency == "" {
		return s.App.HttpResponseBadRequest(c, fmt.Errorf(
			"range_value_currency is required for range_mode %d (%s)",
			mode, model.RangeMode_name[mode]))
	}

	if _, err := tx.Exec(ctx,
		`UPDATE hst.leverage_rules SET
		    name                        = COALESCE($3, name),
		    description                 = COALESCE($4, description),
		    path                        = COALESCE($5, path),
		    range_mode                  = COALESCE($6, range_mode),
		    range_value_currency        = COALESCE($7, range_value_currency),
		    range_value_currency_digits = COALESCE($8, range_value_currency_digits)
		  WHERE rule_id = $1 AND leverage_id = $2`,
		ruleId, id, body.Name, body.Description, body.Path, body.RangeMode,
		body.RangeValueCurrency, body.RangeValueCurrencyDigits); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if body.Tiers != nil {
		if _, err := tx.Exec(ctx,
			`DELETE FROM hst.leverage_tiers WHERE rule_id = $1`, ruleId); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		if err := insertTiers(ctx, tx, int64(ruleId), *body.Tiers); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}

	if err := touchLeverage(ctx, tx, int64(id)); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "leverage rule updated",
		"actor", snap.Login, "target", id, "rule_id", ruleId)

	return s.getLeverageProfile(c, int64(id), s.App.HttpResponseOK)
}

// DeleteLeverageRule removes one rule and closes the gap it leaves, so the
// evaluation order stays a dense 0..n-1 sequence.
//
//	@Id			DeleteLeverageRule
//	@Tags		Leverages
//	@Produce	json
//	@Param		id		path		int	true	"leverage profile id"
//	@Param		ruleId	path		int	true	"rule id, must belong to the profile"
//	@Success	200		{object}	Response{data=ViewLeverageDetail}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/leverage-profiles/{id}/rules/{ruleId} [delete]
func (s *HttpServer) DeleteLeverageRule(c *fiber.Ctx) error {
	ctx := c.UserContext()

	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	ruleId, err := c.ParamsInt("ruleId")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var gone int32
	err = tx.QueryRow(ctx,
		`DELETE FROM hst.leverage_rules
		  WHERE rule_id = $1 AND leverage_id = $2
		 RETURNING config_index`, ruleId, id).Scan(&gone)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrRuleNotInProfile)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// shifting down cannot collide: the vacated index is always free
	if _, err := tx.Exec(ctx,
		`UPDATE hst.leverage_rules SET config_index = config_index - 1
		  WHERE leverage_id = $1 AND config_index > $2`, id, gone); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if err := touchLeverage(ctx, tx, int64(id)); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeWarn, "leverage rule deleted",
		"actor", snap.Login, "target", id, "rule_id", ruleId)

	return s.getLeverageProfile(c, int64(id), s.App.HttpResponseOK)
}

// ReorderLeverageRules rewrites the evaluation order. The body must list every
// rule of the profile exactly once, because a partial order is ambiguous.
//
//	@Id			ReorderLeverageRules
//	@Tags		Leverages
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int						true	"leverage profile id"
//	@Param		body	body		ReorderLeverageRules	true	"every rule id of the profile, exactly once, in the new evaluation order"
//	@Success	200		{object}	Response{data=ViewLeverageDetail}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/leverage-profiles/{id}/rules/reorder [put]
func (s *HttpServer) ReorderLeverageRules(c *fiber.Ctx) error {
	ctx := c.UserContext()

	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	var body ReorderLeverageRules
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	seen := make(map[int64]struct{}, len(body.RuleIds))
	for _, rid := range body.RuleIds {
		if _, dup := seen[rid]; dup {
			return s.App.HttpResponseBadRequest(c, errs.ErrReorderMustListEveryRule)
		}
		seen[rid] = struct{}{}
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// lock the parent, so a concurrent append cannot slip in between the count
	// and the rewrite and leave a rule without an index. The lock has to be on
	// leverages: FOR UPDATE cannot be combined with an aggregate.
	var exists bool
	err = tx.QueryRow(ctx,
		`SELECT true FROM hst.leverages WHERE leverage_id = $1 FOR UPDATE`, id).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	var total int
	if err := tx.QueryRow(ctx,
		`SELECT count(*) FROM hst.leverage_rules WHERE leverage_id = $1`,
		id).Scan(&total); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if total != len(body.RuleIds) {
		return s.App.HttpResponseBadRequest(c, errs.ErrReorderMustListEveryRule)
	}

	// every id must belong to this profile; with the counts equal and no
	// duplicates, that also proves the two sets are identical
	var owned int
	if err := tx.QueryRow(ctx,
		`SELECT count(*) FROM hst.leverage_rules
		  WHERE leverage_id = $1 AND rule_id = ANY($2)`, id, body.RuleIds).Scan(&owned); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if owned != len(body.RuleIds) {
		return s.App.HttpResponseBadRequest(c, errs.ErrRuleNotInProfile)
	}

	// park the indexes out of range first, otherwise the unique constraint
	// fires the moment two rules swap places
	if _, err := tx.Exec(ctx,
		`UPDATE hst.leverage_rules SET config_index = -(config_index + 1)
		  WHERE leverage_id = $1`, id); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	for i, rid := range body.RuleIds {
		if _, err := tx.Exec(ctx,
			`UPDATE hst.leverage_rules SET config_index = $1
			  WHERE rule_id = $2 AND leverage_id = $3`, i, rid, id); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}

	if err := touchLeverage(ctx, tx, int64(id)); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "leverage rules reordered",
		"actor", snap.Login, "target", id, "rules", len(body.RuleIds))

	return s.getLeverageProfile(c, int64(id), s.App.HttpResponseOK)
}

// getLeverageProfile reads the whole tree and answers with the given responder.
// Three queries rather than one per rule, the tree is small but nested.
func (s *HttpServer) getLeverageProfile(c *fiber.Ctx, id int64,
	respond func(*fiber.Ctx, interface{}) error) error {

	ctx := c.UserContext()

	out := &ViewLeverageDetail{Rules: []ViewLeverageRule{}}
	err := s.DB.DB.QueryRow(ctx,
		`SELECT `+leverageColumns+` FROM hst.leverages WHERE leverage_id = $1`, id).
		Scan(&out.LeverageId, &out.Name, &out.Timestamp, &out.Flags)

	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	rows, err := s.DB.DB.Query(ctx,
		`SELECT `+leverageRuleColumns+`
		   FROM hst.leverage_rules WHERE leverage_id = $1 ORDER BY config_index`, id)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	at := map[int64]int{}
	ids := []int64{}
	for rows.Next() {
		var r model.LeverageRule
		if err := rows.Scan(&r.RuleId, &r.LeverageId, &r.Name, &r.Description,
			&r.Path, &r.RangeMode, &r.RangeValueCurrency,
			&r.RangeValueCurrencyDigits, &r.ConfigIndex); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		at[r.RuleId] = len(out.Rules)
		ids = append(ids, r.RuleId)
		out.Rules = append(out.Rules, ViewLeverageRule{
			LeverageRule: r,
			Tiers:        []model.LeverageTier{},
		})
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}
	if len(ids) == 0 {
		return respond(c, out)
	}

	tiers, err := s.DB.DB.Query(ctx,
		`SELECT `+leverageTierColumns+`
		   FROM hst.leverage_tiers WHERE rule_id = ANY($1) ORDER BY rule_id, range_from`, ids)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer tiers.Close()

	for tiers.Next() {
		var t model.LeverageTier
		if err := tiers.Scan(&t.TierId, &t.RuleId, &t.RangeFrom, &t.RangeTo,
			&t.MarginRateInitial, &t.MarginRateMaintenance); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		i, ok := at[t.RuleId]
		if !ok {
			continue
		}
		out.Rules[i].Tiers = append(out.Rules[i].Tiers, t)
	}
	if tiers.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, tiers.Err())
	}

	return respond(c, out)
}

// insertRules writes rules starting at the given index, each with its levels.
func insertRules(ctx context.Context, tx pgx.Tx, leverageId int64, from int32,
	rules []CrtLeverageRule) error {

	for i, r := range rules {
		var ruleId int64
		if err := tx.QueryRow(ctx,
			`INSERT INTO hst.leverage_rules (leverage_id, name, description, path,
			    range_mode, range_value_currency, range_value_currency_digits, config_index)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING rule_id`,
			leverageId, r.Name, r.Description, r.Path, r.RangeMode,
			r.RangeValueCurrency, r.RangeValueCurrencyDigits, from+int32(i)).Scan(&ruleId); err != nil {
			return err
		}
		if err := insertTiers(ctx, tx, ruleId, r.Tiers); err != nil {
			return err
		}
	}

	return nil
}

// insertTiers writes the levels of one rule. MT5 takes only the upper bound in
// the interface and reads the lower one off the previous level, so range_from
// is derived here rather than trusted from the request.
func insertTiers(ctx context.Context, tx pgx.Tx, ruleId int64, tiers []CrtLeverageTier) error {
	var from float64

	for _, t := range tiers {
		if _, err := tx.Exec(ctx,
			`INSERT INTO hst.leverage_tiers (rule_id, range_from, range_to,
			    margin_rate_initial, margin_rate_maintenance)
			 VALUES ($1,$2,$3,$4,$5)`,
			ruleId, from, t.RangeTo, t.MarginRateInitial, t.MarginRateMaintenance); err != nil {
			return err
		}
		from = t.RangeTo
	}

	return nil
}

// touchLeverage moves the profile timestamp, MT5's marker that a record changed.
// Rule and level edits change the profile as a whole, so they move it too.
func touchLeverage(ctx context.Context, tx pgx.Tx, id int64) error {
	_, err := tx.Exec(ctx,
		`UPDATE hst.leverages SET "timestamp" = $2 WHERE leverage_id = $1`,
		id, time.Now().UnixNano())
	return err
}

// validateRules checks the rule level invariants the schema cannot express.
func validateRules(rules []CrtLeverageRule) error {
	if len(rules) > maxLeverageRules {
		return errs.ErrLeverageRuleLimitReached
	}

	for i, r := range rules {
		if model.RangeMode(r.RangeMode).NeedsCurrency() && r.RangeValueCurrency == "" {
			return fmt.Errorf("rule %d: range_value_currency is required for range_mode %d (%s)",
				i, r.RangeMode, model.RangeMode_name[r.RangeMode])
		}
		if err := validateTiers(r.Tiers); err != nil {
			return fmt.Errorf("rule %d: %w", i, err)
		}
	}

	return nil
}

// validateTiers enforces MT5's level rules: the levels are contiguous brackets
// in ascending order, and the last one is open ended, which it signals with a
// zero upper bound.
func validateTiers(tiers []CrtLeverageTier) error {
	if len(tiers) == 0 {
		return errors.New("at least one tier is required")
	}

	last := len(tiers) - 1
	for i, t := range tiers {
		if i == last {
			if t.RangeTo != 0 {
				return fmt.Errorf("tier %d: the last tier must have range_to 0, meaning infinity", i)
			}
			continue
		}
		if t.RangeTo <= 0 {
			return fmt.Errorf("tier %d: range_to must be greater than 0, only the last tier may be 0", i)
		}
		if i > 0 && t.RangeTo <= tiers[i-1].RangeTo {
			return fmt.Errorf("tier %d: range_to must be greater than the previous tier", i)
		}
	}

	return nil
}

// isUniqueViolation reports whether the error is a duplicate key, so a clashing
// profile name answers 409 rather than 500.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

package admin

import (
	"context"
	"errors"
	v1 "hstserver/internal/server/v1"
	"strings"
	"time"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/journal"
	"hstserver/pkg/logger"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// CrtCommissionTier is one ladder level on create/replace.
type CrtCommissionTier struct {
	Mode      model.CommissionTierMode `json:"mode"`
	Type      model.CommissionTierType `json:"type"`
	Value     float64                  `json:"value"`
	RangeFrom float64                  `json:"range_from"`
	RangeTo   float64                  `json:"range_to"`
	Minimal   float64                  `json:"minimal"`
	Currency  string                   `json:"currency" validate:"max=16"`
}

// CrtCommission creates a commission header; tiers are inserted in the same request.
type CrtCommission struct {
	Name             string                       `json:"name" validate:"required,max=64"`
	Description      string                       `json:"description" validate:"max=64"`
	Path             string                       `json:"path" validate:"max=255"`
	Mode             *model.CommissionMode        `json:"mode"`
	ModeRange        *model.CommissionRangeMode   `json:"mode_range"`
	ModeCharge       *model.CommissionChargeMode  `json:"mode_charge"`
	TurnoverCurrency string                       `json:"turnover_currency" validate:"max=16"`
	ModeEntry        *model.CommissionEntryMode   `json:"mode_entry"`
	ModeAction       *model.CommissionActionMode  `json:"mode_action"`
	ModeProfit       *model.CommissionProfitMode  `json:"mode_profit"`
	ModeReason       *model.CommissionReasonFlags `json:"mode_reason"`
	Tiers            []CrtCommissionTier          `json:"tiers" validate:"dive"`
}

// UptCommission patches a commission; when tiers is present it replaces every tier.
type UptCommission struct {
	Name             *string                      `json:"name" validate:"omitempty,max=64"`
	Description      *string                      `json:"description" validate:"omitempty,max=64"`
	Path             *string                      `json:"path" validate:"omitempty,max=255"`
	Mode             *model.CommissionMode        `json:"mode"`
	ModeRange        *model.CommissionRangeMode   `json:"mode_range"`
	ModeCharge       *model.CommissionChargeMode  `json:"mode_charge"`
	TurnoverCurrency *string                      `json:"turnover_currency" validate:"omitempty,max=16"`
	ModeEntry        *model.CommissionEntryMode   `json:"mode_entry"`
	ModeAction       *model.CommissionActionMode  `json:"mode_action"`
	ModeProfit       *model.CommissionProfitMode  `json:"mode_profit"`
	ModeReason       *model.CommissionReasonFlags `json:"mode_reason"`
	Tiers            *[]CrtCommissionTier         `json:"tiers" validate:"omitempty,dive"`
}

// ViewCommissionTier is one ladder level in the response.
type ViewCommissionTier struct {
	TierID       int64                    `json:"tier_id"`
	CommissionID int64                    `json:"commission_id"`
	Mode         model.CommissionTierMode `json:"mode"`
	Type         model.CommissionTierType `json:"type"`
	Value        float64                  `json:"value"`
	RangeFrom    float64                  `json:"range_from"`
	RangeTo      float64                  `json:"range_to"`
	Minimal      float64                  `json:"minimal"`
	Currency     string                   `json:"currency"`
}

// ViewCommission is a commission header with its tiers.
type ViewCommission struct {
	CommissionID     int64                       `json:"commission_id"`
	GroupID          int64                       `json:"group_id"`
	UpdatedAt        int64                       `json:"updated_at"`
	Name             string                      `json:"name"`
	Description      string                      `json:"description"`
	Path             string                      `json:"path"`
	Mode             model.CommissionMode        `json:"mode"`
	ModeRange        model.CommissionRangeMode   `json:"mode_range"`
	ModeCharge       model.CommissionChargeMode  `json:"mode_charge"`
	TurnoverCurrency string                      `json:"turnover_currency"`
	ModeEntry        model.CommissionEntryMode   `json:"mode_entry"`
	ModeAction       model.CommissionActionMode  `json:"mode_action"`
	ModeProfit       model.CommissionProfitMode  `json:"mode_profit"`
	ModeReason       model.CommissionReasonFlags `json:"mode_reason"`
	Tiers            []ViewCommissionTier        `json:"tiers"`
}

const commissionColumns = `commission_id, group_id, updated_at, name, description, path,
	mode, mode_range, mode_charge, turnover_currency,
	mode_entry, mode_action, mode_profit, mode_reason`

const commissionTierColumns = `tier_id, commission_id, mode, type, value,
	range_from, range_to, minimal, currency`

func scanViewCommission(row pgx.Row) (*ViewCommission, error) {
	v := &ViewCommission{Tiers: []ViewCommissionTier{}}
	err := row.Scan(
		&v.CommissionID, &v.GroupID, &v.UpdatedAt, &v.Name, &v.Description, &v.Path,
		&v.Mode, &v.ModeRange, &v.ModeCharge, &v.TurnoverCurrency,
		&v.ModeEntry, &v.ModeAction, &v.ModeProfit, &v.ModeReason,
	)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func loadCommissionTiers(ctx context.Context, s *Server, commissionID int64) ([]ViewCommissionTier, error) {
	rows, err := s.DB.DB.Query(ctx,
		`SELECT `+commissionTierColumns+` FROM hst.commissions_tiers
		  WHERE commission_id = $1 ORDER BY range_from, tier_id`, commissionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ViewCommissionTier{}
	for rows.Next() {
		var t ViewCommissionTier
		if err := rows.Scan(
			&t.TierID, &t.CommissionID, &t.Mode, &t.Type, &t.Value,
			&t.RangeFrom, &t.RangeTo, &t.Minimal, &t.Currency,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// ListGroupCommissions lists commission headers for a group (tiers included).
//
//	@Id			ListGroupCommissions
//	@Tags		Groups
//	@Produce	json
//	@Param		id	path		int	true	"group id"
//	@Success	200	{object}	Response{data=[]ViewCommission}
//	@Failure	400	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/groups/{id}/commissions [get]
func (s *Server) ListGroupCommissions(c *fiber.Ctx) error {
	groupID, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	if err := groupExists(c, s, groupID); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	ctx := c.UserContext()
	rows, err := s.DB.DB.Query(ctx,
		`SELECT `+commissionColumns+` FROM hst.commissions
		  WHERE group_id = $1 ORDER BY name, commission_id`, groupID)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := []ViewCommission{}
	for rows.Next() {
		v, err := scanViewCommission(rows)
		if err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		tiers, err := loadCommissionTiers(ctx, s, v.CommissionID)
		if err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		v.Tiers = tiers
		out = append(out, *v)
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}
	return s.App.HttpResponseOK(c, out)
}

// GetGroupCommission returns one commission with tiers.
//
//	@Id			GetGroupCommission
//	@Tags		Groups
//	@Produce	json
//	@Param		id				path		int	true	"group id"
//	@Param		commissionId	path		int	true	"commission id"
//	@Success	200				{object}	Response{data=ViewCommission}
//	@Failure	400				{object}	Response
//	@Failure	404				{object}	Response
//	@Failure	500				{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/groups/{id}/commissions/{commissionId} [get]
func (s *Server) GetGroupCommission(c *fiber.Ctx) error {
	groupID, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	commissionID, err := c.ParamsInt("commissionId")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	ctx := c.UserContext()
	v, err := scanViewCommission(s.DB.DB.QueryRow(ctx,
		`SELECT `+commissionColumns+` FROM hst.commissions
		  WHERE group_id = $1 AND commission_id = $2`, groupID, commissionID))
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	tiers, err := loadCommissionTiers(ctx, s, v.CommissionID)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	v.Tiers = tiers
	return s.App.HttpResponseOK(c, v)
}

// CreateGroupCommission inserts a commission header and optional tiers.
//
//	@Id			CreateGroupCommission
//	@Tags		Groups
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int				true	"group id"
//	@Param		body	body		CrtCommission	true	"name required; tiers optional"
//	@Success	201		{object}	Response{data=ViewCommission}
//	@Failure	400		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/groups/{id}/commissions [post]
func (s *Server) CreateGroupCommission(c *fiber.Ctx) error {
	groupID, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	if err := groupExists(c, s, groupID); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	var body CrtCommission
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	ctx := c.UserContext()
	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	now := time.Now().UnixNano()
	v, err := scanViewCommission(tx.QueryRow(ctx,
		`INSERT INTO hst.commissions (
		    group_id, updated_at, name, description, path,
		    mode, mode_range, mode_charge, turnover_currency,
		    mode_entry, mode_action, mode_profit, mode_reason
		 ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		 RETURNING `+commissionColumns,
		groupID, now, name, body.Description, body.Path,
		v1.PtrOr(body.Mode, model.CommissionMode_standard),
		v1.PtrOr(body.ModeRange, model.CommissionRangeMode_volume),
		v1.PtrOr(body.ModeCharge, model.CommissionChargeMode_daily),
		body.TurnoverCurrency,
		v1.PtrOr(body.ModeEntry, model.CommissionEntryMode_all),
		v1.PtrOr(body.ModeAction, model.CommissionActionMode_all),
		v1.PtrOr(body.ModeProfit, model.CommissionProfitMode_all),
		v1.PtrOr(body.ModeReason, model.CommissionReasonFlags_none),
	))
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	for _, t := range body.Tiers {
		if _, err := tx.Exec(ctx,
			`INSERT INTO hst.commissions_tiers
			    (commission_id, mode, type, value, range_from, range_to, minimal, currency)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			v.CommissionID, t.Mode, t.Type, t.Value, t.RangeFrom, t.RangeTo, t.Minimal, t.Currency); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	tiers, err := loadCommissionTiers(ctx, s, v.CommissionID)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	v.Tiers = tiers

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "group commission created",
		"actor", snap.Login, "group_id", groupID, "commission_id", v.CommissionID)

	path := s.GroupPath(c, groupID)
	s.NotifyWS(model.SubjectGroupCommission(path), model.EventGroupCommissionCreated, v)
	s.NotifySystem(model.SubjectSystemGroupCommissionCreated, v)
	s.JournalEntry(c, logger.TypeCfg, logger.CodeOK, journal.GroupCommissionCreatedMsg(snap.Login, path), v)

	return s.App.HttpResponseCreated(c, v)
}

// UpdateGroupCommission patches a commission; sending tiers replaces all levels.
//
//	@Id			UpdateGroupCommission
//	@Tags		Groups
//	@Accept		json
//	@Produce	json
//	@Param		id				path		int				true	"group id"
//	@Param		commissionId	path		int				true	"commission id"
//	@Param		body			body		UptCommission	true	"fields to change"
//	@Success	200				{object}	Response{data=ViewCommission}
//	@Failure	400				{object}	Response
//	@Failure	404				{object}	Response
//	@Failure	500				{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/groups/{id}/commissions/{commissionId} [patch]
func (s *Server) UpdateGroupCommission(c *fiber.Ctx) error {
	groupID, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	commissionID, err := c.ParamsInt("commissionId")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	var body UptCommission
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	ctx := c.UserContext()
	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	now := time.Now().UnixNano()
	v, err := scanViewCommission(tx.QueryRow(ctx,
		`UPDATE hst.commissions SET
		    name              = COALESCE($3, name),
		    description       = COALESCE($4, description),
		    path              = COALESCE($5, path),
		    mode              = COALESCE($6, mode),
		    mode_range        = COALESCE($7, mode_range),
		    mode_charge       = COALESCE($8, mode_charge),
		    turnover_currency = COALESCE($9, turnover_currency),
		    mode_entry        = COALESCE($10, mode_entry),
		    mode_action       = COALESCE($11, mode_action),
		    mode_profit       = COALESCE($12, mode_profit),
		    mode_reason       = COALESCE($13, mode_reason),
		    updated_at        = $14
		  WHERE group_id = $1 AND commission_id = $2
		  RETURNING `+commissionColumns,
		groupID, commissionID,
		body.Name, body.Description, body.Path,
		body.Mode, body.ModeRange, body.ModeCharge, body.TurnoverCurrency,
		body.ModeEntry, body.ModeAction, body.ModeProfit, body.ModeReason, now,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if body.Tiers != nil {
		if _, err := tx.Exec(ctx,
			`DELETE FROM hst.commissions_tiers WHERE commission_id = $1`, commissionID); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		for _, t := range *body.Tiers {
			if _, err := tx.Exec(ctx,
				`INSERT INTO hst.commissions_tiers
				    (commission_id, mode, type, value, range_from, range_to, minimal, currency)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
				commissionID, t.Mode, t.Type, t.Value, t.RangeFrom, t.RangeTo, t.Minimal, t.Currency); err != nil {
				return s.App.HttpResponseInternalServerErrorRequest(c, err)
			}
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	tiers, err := loadCommissionTiers(ctx, s, v.CommissionID)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	v.Tiers = tiers

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "group commission updated",
		"actor", snap.Login, "group_id", groupID, "commission_id", commissionID)

	path := s.GroupPath(c, groupID)
	s.NotifyWS(model.SubjectGroupCommission(path), model.EventGroupCommissionUpdated, v)
	s.NotifySystem(model.SubjectSystemGroupCommissionUpdated, v)
	s.JournalEntry(c, logger.TypeCfg, logger.CodeOK, journal.GroupCommissionUpdatedMsg(snap.Login, path), v)

	return s.App.HttpResponseOK(c, v)
}

// DeleteGroupCommission removes a commission and its tiers.
//
//	@Id			DeleteGroupCommission
//	@Tags		Groups
//	@Produce	json
//	@Param		id				path		int	true	"group id"
//	@Param		commissionId	path		int	true	"commission id"
//	@Success	204				{object}	Response
//	@Failure	400				{object}	Response
//	@Failure	404				{object}	Response
//	@Failure	500				{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/groups/{id}/commissions/{commissionId} [delete]
func (s *Server) DeleteGroupCommission(c *fiber.Ctx) error {
	groupID, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	commissionID, err := c.ParamsInt("commissionId")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	ct, err := s.DB.DB.Exec(c.UserContext(),
		`DELETE FROM hst.commissions WHERE group_id = $1 AND commission_id = $2`,
		groupID, commissionID)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if ct.RowsAffected() == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "group commission deleted",
		"actor", snap.Login, "group_id", groupID, "commission_id", commissionID)

	path := s.GroupPath(c, groupID)
	ref := v1.ViewCommissionRef{GroupID: groupID, CommissionID: commissionID}
	s.NotifyWS(model.SubjectGroupCommission(path), model.EventGroupCommissionDeleted, ref)
	s.NotifySystem(model.SubjectSystemGroupCommissionDeleted, ref)
	s.JournalEntry(c, logger.TypeCfg, logger.CodeWarn, journal.GroupCommissionDeletedMsg(snap.Login, path), ref)

	return s.App.HttpResponseNoContent(c)
}

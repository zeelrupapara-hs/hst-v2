package admin

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/logger"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

const (
	routeActionDealer       = 1001
	routeActionDealerOnline = 1002
	maxRejectReasonLen      = 31
	maxDelayTicks           = 60
)

// CrtRoutingCondition is one additional filter on a rule.
type CrtRoutingCondition struct {
	Condition int32  `json:"condition"`
	Rule      int16  `json:"rule" validate:"gte=0,lte=5"`
	Value     string `json:"value" validate:"max=1024"`
}

// CrtRouting creates a routing rule, optionally with its conditions.
type CrtRouting struct {
	Name         string                `json:"name" validate:"required,max=128"`
	Mode         *int16                `json:"mode" validate:"omitempty,gte=0,lte=1"`
	Request      int32                 `json:"request"`
	Type         int32                 `json:"type"`
	Flags        int32                 `json:"flags"`
	Action       *int32                `json:"action" validate:"required"`
	ActionValue  string                `json:"action_value" validate:"max=1024"`
	RoutingIndex *int32                `json:"routing_index" validate:"omitempty,gte=0"`
	Conditions   []CrtRoutingCondition `json:"conditions" validate:"dive"`
}

// UptRouting patches a rule. conditions replaces all rows when present.
type UptRouting struct {
	Name         *string                `json:"name" validate:"omitempty,max=128"`
	Mode         *int16                 `json:"mode" validate:"omitempty,gte=0,lte=1"`
	Request      *int32                 `json:"request"`
	Type         *int32                 `json:"type"`
	Flags        *int32                 `json:"flags"`
	Action       *int32                 `json:"action"`
	ActionValue  *string                `json:"action_value" validate:"omitempty,max=1024"`
	RoutingIndex *int32                 `json:"routing_index" validate:"omitempty,gte=0"`
	Conditions   *[]CrtRoutingCondition `json:"conditions" validate:"omitempty,dive"`
}

// ReorderRouting carries the new top-to-bottom evaluation order.
type ReorderRouting struct {
	RoutingIds []int64 `json:"routing_ids" validate:"required,min=1"`
}

// ViewRoutingRef identifies the rule a change was about; the engine reloads the whole list either way.
type ViewRoutingRef struct {
	RoutingId int64 `json:"routing_id"`
}

// ViewRouting is one row in the list.
type ViewRouting struct {
	RoutingId    int64  `json:"routing_id"`
	Name         string `json:"name"`
	Mode         int16  `json:"mode"`
	Action       int32  `json:"action"`
	RoutingIndex int32  `json:"routing_index"`
	DateModified int64  `json:"date_modified"`
}

// ViewRoutingDetail is a rule with its conditions attached.
type ViewRoutingDetail struct {
	model.RoutingRule
	Conditions []model.RoutingCondition `json:"conditions"`
}

var routingSortable = utils.NewSortable(
	"routing_id", "name", "mode", "action", "routing_index", "date_created", "date_modified")

const routingListColumns = `routing_id, name, mode, action, routing_index, date_modified`

const routingAllColumns = `
	routing_id, name, mode, request, type, flags,
	action, action_value,
	routing_index, date_created, date_modified`

const routingCondColumns = `condition_id, routing_id, condition, rule, value`

// CreateRouting registers a trade request routing rule.
//
//	@Id			CreateRouting
//	@Tags		Routing
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CrtRouting	true	"the rule to create"
//	@Success	201		{object}	Response{data=ViewRoutingDetail}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	409		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/routing [post]
func (s *Server) CreateRouting(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var body CrtRouting
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}
	if err := validateRoutingAction(*body.Action, body.ActionValue); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := validateRoutingConditions(body.Conditions); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	mode := int16(1)
	if body.Mode != nil {
		mode = *body.Mode
	}
	action := *body.Action

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var index int32
	if body.RoutingIndex != nil {
		index = *body.RoutingIndex
		if err := shiftRoutingIndices(ctx, tx, index, 1); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	} else {
		if err := tx.QueryRow(ctx,
			`SELECT COALESCE(MAX(routing_index) + 1, 0) FROM hst.routing`).Scan(&index); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}

	now := time.Now().UnixNano()
	var id int64
	err = tx.QueryRow(ctx,
		`INSERT INTO hst.routing
		   (name, mode, request, type, flags, action, action_value,
		    routing_index, date_created, date_modified)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$9)
		 RETURNING routing_id`,
		body.Name, mode, body.Request, body.Type, body.Flags,
		action, body.ActionValue,
		index, now).Scan(&id)
	if err != nil {
		if utils.IsUniqueViolation(err) {
			return s.App.HttpResponseConflict(c, errs.ErrAlreadyExists)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if err := insertRoutingConditions(ctx, tx, id, body.Conditions); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "routing rule created",
		"actor", snap.Login, "target", body.Name, "conditions", len(body.Conditions))

	s.NotifySystem(model.SubjectSystemRoutingCreated, ViewRoutingRef{RoutingId: id})

	return s.getRoutingDetail(c, id, s.App.HttpResponseCreated)
}

// ListRouting returns routing rules in evaluation order.
//
//	@Id			ListRouting
//	@Tags		Routing
//	@Produce	json
//	@Param		page	query		int		false	"page number, from 1"
//	@Param		limit	query		int		false	"rows per page, max 500"
//	@Param		search	query		string	false	"matches name"
//	@Param		sort_by	query		string	false	"routing_id, name, mode, action, routing_index, date_created, date_modified"	Enums(routing_id, name, mode, action, routing_index, date_created, date_modified)
//	@Param		order	query		string	false	"asc or desc"																	Enums(asc, desc)
//	@Success	200		{object}	Response{data=[]ViewRouting}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/routing [get]
func (s *Server) ListRouting(c *fiber.Ctx) error {
	q, err := utils.QueryFilter(c, routingSortable, "routing_index")
	if err != nil {
		return s.App.HttpResponseBadQueryParams(c, err)
	}

	// always ascending: routing_index is the evaluation order, not a user preference, and a
	// list shown newest-first would read as the rules running backwards
	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT `+routingListColumns+`
		   FROM hst.routing
		  WHERE ($1 = '' OR name ILIKE '%'||$1||'%')
		  ORDER BY routing_index
		  LIMIT $2 OFFSET $3`, q.Search, q.Limit, q.Offset)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := []ViewRouting{}
	for rows.Next() {
		var v ViewRouting
		if err := rows.Scan(&v.RoutingId, &v.Name, &v.Mode, &v.Action,
			&v.RoutingIndex, &v.DateModified); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		out = append(out, v)
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}

	return s.App.HttpResponseOK(c, out)
}

// GetRouting returns one rule with its conditions.
//
//	@Id			GetRouting
//	@Tags		Routing
//	@Produce	json
//	@Param		id	path		int	true	"routing rule id"
//	@Success	200	{object}	Response{data=ViewRoutingDetail}
//	@Failure	400	{object}	Response
//	@Failure	403	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/routing/{id} [get]
func (s *Server) GetRouting(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	return s.getRoutingDetail(c, int64(id), s.App.HttpResponseOK)
}

// UpdateRouting patches fields present in the body.
//
//	@Id			UpdateRouting
//	@Tags		Routing
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int			true	"routing rule id"
//	@Param		body	body		UptRouting	true	"only the fields to change"
//	@Success	200		{object}	Response{data=ViewRoutingDetail}
//	@Failure	400		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	409		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/routing/{id} [patch]
func (s *Server) UpdateRouting(c *fiber.Ctx) error {
	ctx := c.UserContext()

	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	var body UptRouting
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}
	if body.Conditions != nil {
		if err := validateRoutingConditions(*body.Conditions); err != nil {
			return s.App.HttpResponseBadRequest(c, err)
		}
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var name string
	var currentAction int32
	var currentActionValue string
	err = tx.QueryRow(ctx,
		`SELECT name, action, action_value FROM hst.routing WHERE routing_id = $1 FOR UPDATE`, id).
		Scan(&name, &currentAction, &currentActionValue)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if body.Action != nil || body.ActionValue != nil {
		action := currentAction
		if body.Action != nil {
			action = *body.Action
		}
		actionValue := currentActionValue
		if body.ActionValue != nil {
			actionValue = *body.ActionValue
		}
		if err := validateRoutingAction(action, actionValue); err != nil {
			return s.App.HttpResponseBadRequest(c, err)
		}
	}

	if body.RoutingIndex != nil {
		var current int32
		if err := tx.QueryRow(ctx,
			`SELECT routing_index FROM hst.routing WHERE routing_id = $1`, id).Scan(&current); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		if *body.RoutingIndex != current {
			if err := moveRoutingIndex(ctx, tx, int64(id), current, *body.RoutingIndex); err != nil {
				return s.App.HttpResponseInternalServerErrorRequest(c, err)
			}
		}
	}

	tag, err := tx.Exec(ctx,
		`UPDATE hst.routing SET
		    name          = COALESCE($2, name),
		    mode          = COALESCE($3, mode),
		    request       = COALESCE($4, request),
		    type          = COALESCE($5, type),
		    flags         = COALESCE($6, flags),
		    action        = COALESCE($7, action),
		    action_value  = COALESCE($8, action_value),
		    date_modified = $9
		  WHERE routing_id = $1`,
		id, body.Name, body.Mode, body.Request, body.Type, body.Flags,
		body.Action, body.ActionValue,
		time.Now().UnixNano())
	if err != nil {
		if utils.IsUniqueViolation(err) {
			return s.App.HttpResponseConflict(c, errs.ErrAlreadyExists)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if tag.RowsAffected() == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	if body.Conditions != nil {
		if _, err := tx.Exec(ctx,
			`DELETE FROM hst.routing_conds WHERE routing_id = $1`, id); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		if err := insertRoutingConditions(ctx, tx, int64(id), *body.Conditions); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	target := name
	if body.Name != nil {
		target = *body.Name
	}
	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "routing rule updated",
		"actor", snap.Login, "target", target, "conditions_replaced", body.Conditions != nil)

	s.NotifySystem(model.SubjectSystemRoutingUpdated, ViewRoutingRef{RoutingId: int64(id)})

	return s.getRoutingDetail(c, int64(id), s.App.HttpResponseOK)
}

// DeleteRouting removes a rule and compacts the remaining indices.
//
//	@Id			DeleteRouting
//	@Tags		Routing
//	@Produce	json
//	@Param		id	path		int	true	"routing rule id"
//	@Success	204	{object}	Response
//	@Failure	400	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/routing/{id} [delete]
func (s *Server) DeleteRouting(c *fiber.Ctx) error {
	ctx := c.UserContext()

	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var name string
	var gone int32
	err = tx.QueryRow(ctx,
		`DELETE FROM hst.routing WHERE routing_id = $1
		 RETURNING name, routing_index`, id).Scan(&name, &gone)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// routing_index is unique, so closing the gap in one pass can collide with a row that has
	// not moved yet; parking the tail on negative indexes first keeps every step unique.
	if _, err := tx.Exec(ctx,
		`UPDATE hst.routing SET routing_index = -routing_index - 1
		  WHERE routing_index > $1`, gone); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if _, err := tx.Exec(ctx,
		`UPDATE hst.routing SET routing_index = -routing_index - 2
		  WHERE routing_index < 0`); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeWarn, "routing rule deleted",
		"actor", snap.Login, "target", name)

	s.NotifySystem(model.SubjectSystemRoutingDeleted, ViewRoutingRef{RoutingId: int64(id)})

	return s.App.HttpResponseNoContent(c)
}

// ReorderRouting rewrites the evaluation order for every rule.
//
//	@Id			ReorderRouting
//	@Tags		Routing
//	@Accept		json
//	@Produce	json
//	@Param		body	body		ReorderRouting	true	"every routing id exactly once, in the new order"
//	@Success	200		{object}	Response{data=[]ViewRouting}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/routing/order [put]
func (s *Server) ReorderRouting(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var body ReorderRouting
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	seen := make(map[int64]struct{}, len(body.RoutingIds))
	for _, rid := range body.RoutingIds {
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

	var total int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM hst.routing`).Scan(&total); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if total != len(body.RoutingIds) {
		return s.App.HttpResponseBadRequest(c, errs.ErrReorderMustListEveryRule)
	}

	var owned int
	if err := tx.QueryRow(ctx,
		`SELECT count(*) FROM hst.routing WHERE routing_id = ANY($1)`,
		body.RoutingIds).Scan(&owned); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if owned != len(body.RoutingIds) {
		return s.App.HttpResponseBadRequest(c, errs.ErrReorderMustListEveryRule)
	}

	if _, err := tx.Exec(ctx,
		`UPDATE hst.routing SET routing_index = -(routing_index + 1)`); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	now := time.Now().UnixNano()
	for i, rid := range body.RoutingIds {
		if _, err := tx.Exec(ctx,
			`UPDATE hst.routing SET routing_index = $1, date_modified = $2
			  WHERE routing_id = $3`, i, now, rid); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "routing rules reordered",
		"actor", snap.Login, "rules", len(body.RoutingIds))

	s.NotifySystem(model.SubjectSystemRoutingUpdated, ViewRoutingRef{})

	out, err := s.listRoutingOrdered(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, out)
}

// MoveRoutingUp swaps the rule one step toward the top of the evaluation list.
//
//	@Id			MoveRoutingUp
//	@Tags		Routing
//	@Produce	json
//	@Param		id	path		int	true	"routing rule id"
//	@Success	200	{object}	Response{data=[]ViewRouting}
//	@Failure	400	{object}	Response
//	@Failure	403	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/routing/{id}/move-up [post]
func (s *Server) MoveRoutingUp(c *fiber.Ctx) error {
	return s.moveRoutingOneStep(c, -1)
}

// MoveRoutingDown swaps the rule one step toward the bottom of the evaluation list.
//
//	@Id			MoveRoutingDown
//	@Tags		Routing
//	@Produce	json
//	@Param		id	path		int	true	"routing rule id"
//	@Success	200	{object}	Response{data=[]ViewRouting}
//	@Failure	400	{object}	Response
//	@Failure	403	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/routing/{id}/move-down [post]
func (s *Server) MoveRoutingDown(c *fiber.Ctx) error {
	return s.moveRoutingOneStep(c, 1)
}

func (s *Server) moveRoutingOneStep(c *fiber.Ctx, delta int32) error {
	ctx := c.UserContext()

	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var name string
	var current int32
	err = tx.QueryRow(ctx,
		`SELECT name, routing_index FROM hst.routing WHERE routing_id = $1 FOR UPDATE`, id).
		Scan(&name, &current)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	target := current + delta
	if delta < 0 && current == 0 {
		return s.App.HttpResponseBadRequest(c, errs.ErrRoutingAlreadyFirst)
	}
	if delta > 0 {
		var maxIndex int32
		if err := tx.QueryRow(ctx, `SELECT COALESCE(MAX(routing_index), 0) FROM hst.routing`).Scan(&maxIndex); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		if current >= maxIndex {
			return s.App.HttpResponseBadRequest(c, errs.ErrRoutingAlreadyLast)
		}
	}

	var neighborId int64
	err = tx.QueryRow(ctx,
		`SELECT routing_id FROM hst.routing WHERE routing_index = $1`, target).Scan(&neighborId)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseInternalServerErrorRequest(c, fmt.Errorf("routing neighbor at index %d not found", target))
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	now := time.Now().UnixNano()
	// Swap through a unique negative slot so routing_index stays unique mid-update.
	if _, err := tx.Exec(ctx,
		`UPDATE hst.routing SET routing_index = -routing_id WHERE routing_id = $1`, id); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if _, err := tx.Exec(ctx,
		`UPDATE hst.routing SET routing_index = $1, date_modified = $2
		  WHERE routing_id = $3`, current, now, neighborId); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if _, err := tx.Exec(ctx,
		`UPDATE hst.routing SET routing_index = $1, date_modified = $2
		  WHERE routing_id = $3`, target, now, id); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	dir := "down"
	if delta < 0 {
		dir = "up"
	}
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "routing rule moved "+dir,
		"actor", snap.Login, "target", name, "from", current, "to", target)

	s.NotifySystem(model.SubjectSystemRoutingUpdated, ViewRoutingRef{RoutingId: int64(id)})

	out, err := s.listRoutingOrdered(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, out)
}

func (s *Server) listRoutingOrdered(ctx context.Context) ([]ViewRouting, error) {
	rows, err := s.DB.DB.Query(ctx,
		`SELECT `+routingListColumns+` FROM hst.routing ORDER BY routing_index`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ViewRouting{}
	for rows.Next() {
		var v ViewRouting
		if err := rows.Scan(&v.RoutingId, &v.Name, &v.Mode, &v.Action,
			&v.RoutingIndex, &v.DateModified); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Server) getRoutingDetail(c *fiber.Ctx, id int64,
	respond func(*fiber.Ctx, interface{}) error) error {

	ctx := c.UserContext()
	out := &ViewRoutingDetail{Conditions: []model.RoutingCondition{}}

	err := s.DB.DB.QueryRow(ctx,
		`SELECT `+routingAllColumns+` FROM hst.routing WHERE routing_id = $1`, id).
		Scan(&out.RoutingId, &out.Name, &out.Mode, &out.Request, &out.Type, &out.Flags,
			&out.Action, &out.ActionValue,
			&out.RoutingIndex, &out.DateCreated, &out.DateModified)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	rows, err := s.DB.DB.Query(ctx,
		`SELECT `+routingCondColumns+`
		   FROM hst.routing_conds WHERE routing_id = $1 ORDER BY condition_id`, id)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	for rows.Next() {
		var cond model.RoutingCondition
		if err := rows.Scan(&cond.ConditionId, &cond.RoutingId, &cond.Condition, &cond.Rule,
			&cond.Value); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		out.Conditions = append(out.Conditions, cond)
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}

	return respond(c, out)
}

func insertRoutingConditions(ctx context.Context, tx pgx.Tx, routingId int64,
	conds []CrtRoutingCondition) error {

	for _, c := range conds {
		if _, err := tx.Exec(ctx,
			`INSERT INTO hst.routing_conds (routing_id, condition, rule, value)
			 VALUES ($1,$2,$3,$4)`,
			routingId, c.Condition, c.Rule, c.Value); err != nil {
			return err
		}
	}

	return nil
}

func validateRoutingAction(action int32, actionValue string) error {
	switch action {
	case routeActionDealer, routeActionDealerOnline:
		// the extra parameter is the "skip this rule if no dealers online" box
		if actionValue == "" {
			return nil
		}
		n, err := strconv.ParseInt(actionValue, 10, 32)
		if err != nil || n < 0 || n > 1 {
			return fmt.Errorf("skip if no dealers online must be 0 or 1")
		}
	case int32(model.RouteAction_delay_tick):
		if actionValue == "" {
			return fmt.Errorf("delay tick count is required")
		}
		n, err := strconv.ParseInt(actionValue, 10, 32)
		if err != nil || n < 1 || n > maxDelayTicks {
			return fmt.Errorf("delay tick count must be between 1 and %d", maxDelayTicks)
		}
	case int32(model.RouteAction_delay_time):
		if actionValue != "" {
			n, err := strconv.ParseInt(actionValue, 10, 32)
			if err != nil || n < 0 {
				return fmt.Errorf("delay milliseconds must be >= 0")
			}
		}
	case int32(model.RouteAction_reject):
		if len(actionValue) > maxRejectReasonLen {
			return fmt.Errorf("reject reason must be at most %d characters", maxRejectReasonLen)
		}
	}

	return nil
}

func validateRoutingConditions(conds []CrtRoutingCondition) error {
	for i, c := range conds {
		if err := validateConditionRulePair(c.Condition, c.Rule); err != nil {
			return fmt.Errorf("conditions[%d]: %w", i, err)
		}
	}
	return nil
}

func validateConditionRulePair(condition int32, rule int16) error {
	if rule < 0 || rule > 5 {
		return fmt.Errorf("rule must be between 0 and 5")
	}

	switch model.RouteCondition(condition) {
	case model.RouteCondition_symbol, model.RouteCondition_group,
		model.RouteCondition_comment, model.RouteCondition_comment_client,
		model.RouteCondition_country, model.RouteCondition_city:
		if rule != int16(model.ConditionRule_equal) && rule != int16(model.ConditionRule_not_equal) {
			return fmt.Errorf("condition %d only supports = and != comparisons", condition)
		}
	case model.RouteCondition_expert, model.RouteCondition_signal, model.RouteCondition_gap,
		model.RouteCondition_position_sl_touched, model.RouteCondition_position_tp_touched,
		model.RouteCondition_order_sl_touched, model.RouteCondition_order_tp_touched:
		if rule != int16(model.ConditionRule_equal) && rule != int16(model.ConditionRule_not_equal) {
			return fmt.Errorf("condition %d only supports = and != comparisons", condition)
		}
	}

	return nil
}

// shiftRoutingIndices opens a slot at index by moving rules at or after it down.
// Indices pass through negative space so the unique routing_index constraint cannot collide mid-update.
func shiftRoutingIndices(ctx context.Context, tx pgx.Tx, at int32, delta int32) error {
	if _, err := tx.Exec(ctx,
		`UPDATE hst.routing SET routing_index = -(routing_index + 1)
		  WHERE routing_index >= $1`, at); err != nil {
		return err
	}
	_, err := tx.Exec(ctx,
		`UPDATE hst.routing SET routing_index = -routing_index + $1
		  WHERE routing_index < 0`, delta-1)
	return err
}

// moveRoutingIndex repositions one rule within the ordered list.
func moveRoutingIndex(ctx context.Context, tx pgx.Tx, id int64, from, to int32) error {
	if from == to {
		return nil
	}
	if _, err := tx.Exec(ctx,
		`UPDATE hst.routing SET routing_index = -routing_id WHERE routing_id = $1`, id); err != nil {
		return err
	}
	if from < to {
		if _, err := tx.Exec(ctx,
			`UPDATE hst.routing SET routing_index = routing_index - 1
			  WHERE routing_index > $1 AND routing_index <= $2`, from, to); err != nil {
			return err
		}
	} else {
		if _, err := tx.Exec(ctx,
			`UPDATE hst.routing SET routing_index = routing_index + 1
			  WHERE routing_index >= $2 AND routing_index < $1`, from, to); err != nil {
			return err
		}
	}
	_, err := tx.Exec(ctx,
		`UPDATE hst.routing SET routing_index = $2 WHERE routing_id = $1`, id, to)
	return err
}

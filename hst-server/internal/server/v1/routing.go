package v1

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/journal"
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

// SwapRoutingRulePriorityRequest swaps evaluation order between two rules.
type SwapRoutingRulePriorityRequest struct {
	RoutingId1 int64 `json:"routing_id_1" validate:"required"`
	RoutingId2 int64 `json:"routing_id_2" validate:"required"`
}

// UptRouting patches a rule. conditions replaces all rows when present.
type UptRouting struct {
	Name        *string                `json:"name" validate:"omitempty,max=128"`
	Mode        *int16                 `json:"mode" validate:"omitempty,gte=0,lte=1"`
	Request     *int32                 `json:"request"`
	Type        *int32                 `json:"type"`
	Flags       *int32                 `json:"flags"`
	Action      *int32                 `json:"action"`
	ActionValue *string                `json:"action_value" validate:"omitempty,max=1024"`
	Conditions  *[]CrtRoutingCondition `json:"conditions" validate:"omitempty,dive"`
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
func (s *HttpServer) CreateRouting(c *fiber.Ctx) error {
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

	detail, err := s.loadRoutingDetail(ctx, id)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	s.NotifyWS(model.SubjectRouting, model.EventRoutingCreated, detail)
	s.NotifySystem(model.SubjectSystemRoutingCreated, detail)
	s.JournalEntry(c, logger.CodeOK, journal.RoutingCreatedMsg(snap.Login, detail.Name), detail)

	return s.App.HttpResponseCreated(c, detail)
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
func (s *HttpServer) ListRouting(c *fiber.Ctx) error {
	q, err := utils.QueryFilter(c, routingSortable, "routing_index")
	if err != nil {
		return s.App.HttpResponseBadQueryParams(c, err)
	}

	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT `+routingListColumns+`
		   FROM hst.routing
		  WHERE ($1 = '' OR name ILIKE '%'||$1||'%')
		  ORDER BY `+q.SortBy+`
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
func (s *HttpServer) GetRouting(c *fiber.Ctx) error {
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
func (s *HttpServer) UpdateRouting(c *fiber.Ctx) error {
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

	detail, err := s.loadRoutingDetail(ctx, int64(id))
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	s.NotifyWS(model.SubjectRouting, model.EventRoutingUpdated, detail)
	s.NotifySystem(model.SubjectSystemRoutingUpdated, detail)
	s.JournalEntry(c, logger.CodeOK, journal.RoutingUpdatedMsg(snap.Login, detail.Name), detail)

	return s.App.HttpResponseOK(c, detail)
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
func (s *HttpServer) DeleteRouting(c *fiber.Ctx) error {
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

	if _, err := tx.Exec(ctx,
		`UPDATE hst.routing SET routing_index = routing_index - 1
		  WHERE routing_index > $1`, gone); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeWarn, "routing rule deleted",
		"actor", snap.Login, "target", name)

	ref := ViewRoutingRef{RoutingId: int64(id), Name: name}
	s.NotifyWS(model.SubjectRouting, model.EventRoutingDeleted, ref)
	s.NotifySystem(model.SubjectSystemRoutingDeleted, ref)
	s.JournalEntry(c, logger.CodeWarn, journal.RoutingDeletedMsg(snap.Login, name), ref)

	return s.App.HttpResponseNoContent(c)
}

// SwapRoutingRulePriority swaps evaluation order between two rules.
//
//	@Id			SwapRoutingRulePriority
//	@Tags		Routing
//	@Accept		json
//	@Produce	json
//	@Param		body	body		SwapRoutingRulePriorityRequest	true	"two routing rule ids to swap"
//	@Success	200		{object}	Response{data=[]ViewRouting}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/routing/swap [put]
func (s *HttpServer) SwapRoutingRulePriority(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var body SwapRoutingRulePriorityRequest
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}
	if body.RoutingId1 == body.RoutingId2 {
		return s.App.HttpResponseBadRequest(c, errs.ErrRoutingSwapRule)
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	name1, idx1, name2, idx2, err := swapRoutingRulePriority(ctx, tx, body.RoutingId1, body.RoutingId2)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "routing rule priorities swapped",
		"actor", snap.Login,
		"rule1", name1, "index1", idx1,
		"rule2", name2, "index2", idx2)

	out, err := s.listRoutingOrdered(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	s.NotifyWS(model.SubjectRouting, model.EventRoutingSwapped, out)
	s.NotifySystem(model.SubjectSystemRoutingSwapped, out)
	s.JournalEntry(c, logger.CodeOK, journal.RoutingSwappedMsg(snap.Login, body.RoutingId1, body.RoutingId2), out)

	return s.App.HttpResponseOK(c, out)
}

func swapRoutingRulePriority(ctx context.Context, tx pgx.Tx, id1, id2 int64) (name1 string, idx1 int32, name2 string, idx2 int32, err error) {
	rows, err := tx.Query(ctx,
		`SELECT routing_id, name, routing_index FROM hst.routing
		  WHERE routing_id IN ($1, $2)
		  ORDER BY routing_id FOR UPDATE`, id1, id2)
	if err != nil {
		return "", 0, "", 0, err
	}
	defer rows.Close()

	type ruleRow struct {
		id    int64
		name  string
		index int32
	}
	found := make(map[int64]ruleRow, 2)
	for rows.Next() {
		var r ruleRow
		if err := rows.Scan(&r.id, &r.name, &r.index); err != nil {
			return "", 0, "", 0, err
		}
		found[r.id] = r
	}
	if err := rows.Err(); err != nil {
		return "", 0, "", 0, err
	}
	r1, ok1 := found[id1]
	r2, ok2 := found[id2]
	if !ok1 || !ok2 {
		return "", 0, "", 0, pgx.ErrNoRows
	}
	if r1.index == r2.index {
		return r1.name, r1.index, r2.name, r2.index, nil
	}

	now := time.Now().UnixNano()
	if _, err := tx.Exec(ctx,
		`UPDATE hst.routing SET routing_index = -routing_id WHERE routing_id = $1`, id1); err != nil {
		return "", 0, "", 0, err
	}
	if _, err := tx.Exec(ctx,
		`UPDATE hst.routing SET routing_index = $1, date_modified = $2
		  WHERE routing_id = $3`, r1.index, now, id2); err != nil {
		return "", 0, "", 0, err
	}
	if _, err := tx.Exec(ctx,
		`UPDATE hst.routing SET routing_index = $1, date_modified = $2
		  WHERE routing_id = $3`, r2.index, now, id1); err != nil {
		return "", 0, "", 0, err
	}

	return r1.name, r2.index, r2.name, r1.index, nil
}

func (s *HttpServer) listRoutingOrdered(ctx context.Context) ([]ViewRouting, error) {
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

func (s *HttpServer) getRoutingDetail(c *fiber.Ctx, id int64,
	respond func(*fiber.Ctx, interface{}) error) error {

	out, err := s.loadRoutingDetail(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return respond(c, out)
}

func (s *HttpServer) loadRoutingDetail(ctx context.Context, id int64) (*ViewRoutingDetail, error) {
	out := &ViewRoutingDetail{Conditions: []model.RoutingCondition{}}

	err := s.DB.DB.QueryRow(ctx,
		`SELECT `+routingAllColumns+` FROM hst.routing WHERE routing_id = $1`, id).
		Scan(&out.RoutingId, &out.Name, &out.Mode, &out.Request, &out.Type, &out.Flags,
			&out.Action, &out.ActionValue,
			&out.RoutingIndex, &out.DateCreated, &out.DateModified)
	if err != nil {
		return nil, err
	}

	rows, err := s.DB.DB.Query(ctx,
		`SELECT `+routingCondColumns+`
		   FROM hst.routing_conds WHERE routing_id = $1 ORDER BY condition_id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var cond model.RoutingCondition
		if err := rows.Scan(&cond.ConditionId, &cond.RoutingId, &cond.Condition, &cond.Rule,
			&cond.Value); err != nil {
			return nil, err
		}
		out.Conditions = append(out.Conditions, cond)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
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
		return fmt.Errorf("action %d is not supported", action)
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
		if rule != int16(model.ConditionRule_eq) && rule != int16(model.ConditionRule_not_eq) {
			return fmt.Errorf("condition %d only supports = and != comparisons", condition)
		}
	case model.RouteCondition_expert, model.RouteCondition_signal, model.RouteCondition_gap,
		model.RouteCondition_position_sl_touched, model.RouteCondition_position_tp_touched,
		model.RouteCondition_order_sl_touched, model.RouteCondition_order_tp_touched:
		if rule != int16(model.ConditionRule_eq) && rule != int16(model.ConditionRule_not_eq) {
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

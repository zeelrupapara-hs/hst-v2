package v1

import (
	"errors"
	"strings"
	"time"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/logger"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// CrtGroup creates a group template. Path is required.
type CrtGroup struct {
	Group          string   `json:"group" validate:"required,max=255"`
	ParentID       *int64   `json:"parent_id"`
	Root           *bool    `json:"root"`
	Currency       string   `json:"currency" validate:"omitempty,max=16"`
	CurrencyDigits *int32   `json:"currency_digits"`
	Company        string   `json:"company" validate:"max=255"`
	MarginCall     *float64 `json:"margin_call"`
	MarginStopOut  *float64 `json:"margin_stop_out"`
	AuthMode       *int32   `json:"auth_mode"`
	Status         string   `json:"status" validate:"omitempty,oneof=active inactive enabled disabled"`
}

// UptGroup patches mutable group fields. Pointers + COALESCE keep absent fields.
type UptGroup struct {
	Currency        *string  `json:"currency" validate:"omitempty,max=16"`
	CurrencyDigits  *int32   `json:"currency_digits"`
	Company         *string  `json:"company" validate:"omitempty,max=255"`
	MarginCall      *float64 `json:"margin_call"`
	MarginStopOut   *float64 `json:"margin_stop_out"`
	AuthMode        *int32   `json:"auth_mode"`
	PermissionFlags *int32   `json:"permission_flags"`
	Status          *string  `json:"status" validate:"omitempty,oneof=active inactive enabled disabled"`
}

// ViewGroup is what the panel renders (flat or nested under Groups).
type ViewGroup struct {
	GroupID         int64        `json:"group_id"`
	UpdatedAt       int64        `json:"updated_at"`
	Group           string       `json:"group"`
	Root            bool         `json:"root"`
	ParentID        *int64       `json:"parent_id,omitempty"`
	PermissionFlags int32        `json:"permission_flags"`
	Status          string       `json:"status"`
	AuthMode        int32        `json:"auth_mode"`
	Currency        string       `json:"currency"`
	CurrencyDigits  int32        `json:"currency_digits"`
	Company         string       `json:"company"`
	MarginCall      float64      `json:"margin_call"`
	MarginStopOut   float64      `json:"margin_stop_out"`
	Groups          []*ViewGroup `json:"groups,omitempty"`
}

// ViewGroupSymbol is a slim list row for group symbol overrides.
type ViewGroupSymbol struct {
	SymbolID  int64  `json:"symbol_id"`
	GroupID   int64  `json:"group_id"`
	UpdatedAt int64  `json:"updated_at"`
	Path      string `json:"path"`
	TradeMode *int32 `json:"trade_mode,omitempty"`
	ExecMode  *int32 `json:"exec_mode,omitempty"`
}

const groupColumns = `group_id, updated_at, "group", root, parent_id,
	permission_flags, auth_mode, currency, currency_digits, company,
	margin_call, margin_stop_out`

func scanViewGroup(row pgx.Row) (*ViewGroup, error) {
	v := &ViewGroup{}
	err := row.Scan(
		&v.GroupID, &v.UpdatedAt, &v.Group, &v.Root, &v.ParentID,
		&v.PermissionFlags, &v.AuthMode, &v.Currency, &v.CurrencyDigits,
		&v.Company, &v.MarginCall, &v.MarginStopOut,
	)
	if err != nil {
		return nil, err
	}
	v.Status = model.GroupStatusFromFlags(v.PermissionFlags)
	return v, nil
}

func buildGroupTree(flat []ViewGroup) []*ViewGroup {
	byID := make(map[int64]*ViewGroup, len(flat))
	nodes := make([]*ViewGroup, len(flat))
	for i := range flat {
		g := flat[i]
		g.Groups = nil
		nodes[i] = &g
		byID[g.GroupID] = nodes[i]
	}

	var roots []*ViewGroup
	for _, g := range nodes {
		if g.ParentID == nil {
			roots = append(roots, g)
			continue
		}
		if p, ok := byID[*g.ParentID]; ok {
			p.Groups = append(p.Groups, g)
		} else {
			roots = append(roots, g)
		}
	}
	return roots
}

// ListGroups returns the group tree, or a flat list when ?flat=1.
//
//	@Id			ListGroups
//	@Tags		Groups
//	@Produce	json
//	@Param		flat	query	int	false	"1 = flat list"
//	@Success	200	{object}	Response{data=[]ViewGroup}
//	@Failure	403	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/groups [get]
func (s *HttpServer) ListGroups(c *fiber.Ctx) error {
	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT `+groupColumns+` FROM hst.groups ORDER BY "group"`)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	flat := []ViewGroup{}
	for rows.Next() {
		v, err := scanViewGroup(rows)
		if err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		flat = append(flat, *v)
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}

	if c.Query("flat") == "1" {
		return s.App.HttpResponseOK(c, flat)
	}
	return s.App.HttpResponseOK(c, buildGroupTree(flat))
}

// GetGroup returns one group template.
//
//	@Id			GetGroup
//	@Tags		Groups
//	@Produce	json
//	@Success	200	{object}	Response{data=ViewGroup}
//	@Failure	400	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/groups/{id} [get]
func (s *HttpServer) GetGroup(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	v, err := scanViewGroup(s.DB.DB.QueryRow(c.UserContext(),
		`SELECT `+groupColumns+` FROM hst.groups WHERE group_id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	return s.App.HttpResponseOK(c, v)
}

// CreateGroup inserts a group template.
//
//	@Id			CreateGroup
//	@Tags		Groups
//	@Accept		json
//	@Produce	json
//	@Success	201	{object}	Response{data=ViewGroup}
//	@Failure	400	{object}	Response
//	@Failure	403	{object}	Response
//	@Failure	409	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/groups [post]
func (s *HttpServer) CreateGroup(c *fiber.Ctx) error {
	var body CrtGroup
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	path := strings.TrimSpace(body.Group)
	if path == "" {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	root := body.ParentID == nil
	if body.Root != nil {
		root = *body.Root
	}
	if root {
		body.ParentID = nil
	}

	flags := model.PermissionFlagsFromStatus("active")
	if body.Status != "" {
		flags = model.PermissionFlagsFromStatus(body.Status)
	}

	currency := body.Currency
	if currency == "" {
		currency = "USD"
	}
	digits := int32(2)
	if body.CurrencyDigits != nil {
		digits = *body.CurrencyDigits
	}
	var marginCall, marginStopOut float64
	if body.MarginCall != nil {
		marginCall = *body.MarginCall
	}
	if body.MarginStopOut != nil {
		marginStopOut = *body.MarginStopOut
	}
	var authMode int32
	if body.AuthMode != nil {
		authMode = *body.AuthMode
	}

	snap, _ := utils.GetClient(c)
	now := time.Now().UnixNano()

	v, err := scanViewGroup(s.DB.DB.QueryRow(c.UserContext(),
		`INSERT INTO hst.groups
		   ("group", root, parent_id, permission_flags, auth_mode,
		    currency, currency_digits, company, margin_call, margin_stop_out, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		 RETURNING `+groupColumns,
		path, root, body.ParentID, flags, authMode,
		currency, digits, body.Company, marginCall, marginStopOut, now))
	if err != nil {
		if isUniqueViolation(err) {
			return s.App.HttpResponseConflict(c, errs.ErrAlreadyExists)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	s.Log.Log(logger.TypeCfg, logger.CodeOK, "group created",
		"actor", snap.Login, "group_id", v.GroupID, "group", v.Group)

	return s.App.HttpResponseCreated(c, v)
}

// UpdateGroup patches currency / margins / company / enable status.
//
//	@Id			UpdateGroup
//	@Tags		Groups
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	Response{data=ViewGroup}
//	@Failure	400	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/groups/{id} [patch]
func (s *HttpServer) UpdateGroup(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	var body UptGroup
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	flags := body.PermissionFlags
	if body.Status != nil && *body.Status != "" {
		f := model.PermissionFlagsFromStatus(*body.Status)
		flags = &f
	}

	snap, _ := utils.GetClient(c)
	now := time.Now().UnixNano()

	v, err := scanViewGroup(s.DB.DB.QueryRow(c.UserContext(),
		`UPDATE hst.groups SET
		    currency         = COALESCE($2, currency),
		    currency_digits  = COALESCE($3, currency_digits),
		    company          = COALESCE($4, company),
		    margin_call      = COALESCE($5, margin_call),
		    margin_stop_out  = COALESCE($6, margin_stop_out),
		    auth_mode        = COALESCE($7, auth_mode),
		    permission_flags = COALESCE($8, permission_flags),
		    updated_at       = $9
		  WHERE group_id = $1
		  RETURNING `+groupColumns,
		id, body.Currency, body.CurrencyDigits, body.Company,
		body.MarginCall, body.MarginStopOut, body.AuthMode, flags, now))
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	s.Log.Log(logger.TypeCfg, logger.CodeOK, "group updated",
		"actor", snap.Login, "group_id", v.GroupID)

	return s.App.HttpResponseOK(c, v)
}

// DeleteGroup removes a leaf group with no users on that path.
//
//	@Id			DeleteGroup
//	@Tags		Groups
//	@Produce	json
//	@Success	204	{object}	Response
//	@Failure	400	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	409	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/groups/{id} [delete]
func (s *HttpServer) DeleteGroup(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	ctx := c.UserContext()
	var path string
	err = s.DB.DB.QueryRow(ctx, `SELECT "group" FROM hst.groups WHERE group_id = $1`, id).Scan(&path)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	var children int
	if err := s.DB.DB.QueryRow(ctx, `SELECT COUNT(*) FROM hst.groups WHERE parent_id = $1`, id).Scan(&children); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if children > 0 {
		return s.App.HttpResponseConflict(c, errs.ErrDeleteWhileNotEmpty)
	}

	var users int
	if err := s.DB.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM hst.users WHERE "group" = $1`, path).Scan(&users); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if users > 0 {
		return s.App.HttpResponseConflict(c, errs.ErrDeleteWhileNotEmpty)
	}

	ct, err := s.DB.DB.Exec(ctx, `DELETE FROM hst.groups WHERE group_id = $1`, id)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if ct.RowsAffected() == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "group deleted",
		"actor", snap.Login, "group_id", id, "group", path)

	return s.App.HttpResponseNoContent(c)
}

// ListGroupSymbols lists symbol overrides for one group.
//
//	@Id			ListGroupSymbols
//	@Tags		Groups
//	@Produce	json
//	@Success	200	{object}	Response{data=[]ViewGroupSymbol}
//	@Failure	400	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/groups/{id}/symbols [get]
func (s *HttpServer) ListGroupSymbols(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	var exists int
	if err := s.DB.DB.QueryRow(c.UserContext(),
		`SELECT 1 FROM hst.groups WHERE group_id = $1`, id).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT symbol_id, group_id, updated_at, path, trade_mode, exec_mode
		   FROM hst.groups_symbols WHERE group_id = $1 ORDER BY path`, id)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := []ViewGroupSymbol{}
	for rows.Next() {
		var v ViewGroupSymbol
		if err := rows.Scan(&v.SymbolID, &v.GroupID, &v.UpdatedAt, &v.Path, &v.TradeMode, &v.ExecMode); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		out = append(out, v)
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}
	return s.App.HttpResponseOK(c, out)
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate key")
}

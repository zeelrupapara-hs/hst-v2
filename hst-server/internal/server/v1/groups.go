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

// CrtGroup creates a group template. Path is required; other fields optional.
type CrtGroup struct {
	Group    string `json:"group" validate:"required,max=255"`
	ParentID *int64 `json:"parent_id"`
	Root     *bool  `json:"root"`
	Status   string `json:"status" validate:"omitempty,oneof=active inactive enabled disabled"`

	PermissionFlags *model.PermissionsFlags `json:"permission_flags"`
	AuthMode        *model.AuthMode         `json:"auth_mode"`
	AuthPasswordMin *int32                  `json:"auth_password_min"`

	Company             string `json:"company" validate:"max=255"`
	CompanyPage         string `json:"company_page"`
	CompanyEmail        string `json:"company_email" validate:"max=255"`
	CompanySupportPage  string `json:"company_support_page"`
	CompanySupportEmail string `json:"company_support_email" validate:"max=255"`
	CompanyCatalog      string `json:"company_catalog" validate:"max=255"`

	Currency       string `json:"currency" validate:"omitempty,max=16"`
	CurrencyDigits *int32 `json:"currency_digits"`

	ReportsMode      *model.ReportsMode  `json:"reports_mode"`
	ReportsFlags     *model.ReportsFlags `json:"reports_flags"`
	ReportsEmail     string              `json:"reports_email" validate:"max=255"`
	ReportsSMTP      string              `json:"reports_smtp"`
	ReportsSMTPLogin string              `json:"reports_smtp_login"`

	NewsMode     *model.NewsMode `json:"news_mode"`
	NewsCategory string          `json:"news_category"`
	NewsLangs    []int32         `json:"news_langs"`
	MailMode     *model.MailMode `json:"mail_mode"`

	TradeFlags         *model.GroupTradeFlags `json:"trade_flags"`
	TradeInterestRate  *float64               `json:"trade_interest_rate"`
	TradeVirtualCredit *float64               `json:"trade_virtual_credit"`
	TradeTransferMode  *model.TransferMode    `json:"trade_transfer_mode"`

	MarginFreeMode       *model.FreeMarginMode       `json:"margin_free_mode"`
	MarginSOMode         *model.StopOutMode          `json:"margin_so_mode"`
	MarginCall           *float64                    `json:"margin_call"`
	MarginStopOut        *float64                    `json:"margin_stop_out"`
	MarginFreeProfitMode *model.MarginFreeProfitMode `json:"margin_free_profit_mode"`
	MarginMode           *model.MarginMode           `json:"margin_mode"`
	MarginFlags          *model.GroupMarginFlags     `json:"margin_flags"`

	DemoLeverage *int32   `json:"demo_leverage"`
	DemoDeposit  *float64 `json:"demo_deposit"`

	LimitHistory         *model.HistoryLimit `json:"limit_history"`
	LimitOrders          *int32              `json:"limit_orders"`
	LimitSymbols         *int32              `json:"limit_symbols"`
	LimitPositions       *int32              `json:"limit_positions"`
	LimitPositionsVolume *float64            `json:"limit_positions_volume"`
}

// UptGroup patches mutable group fields. Pointers + COALESCE keep absent fields.
type UptGroup struct {
	Status *string `json:"status" validate:"omitempty,oneof=active inactive enabled disabled"`

	PermissionFlags *model.PermissionsFlags `json:"permission_flags"`
	AuthMode        *model.AuthMode         `json:"auth_mode"`
	AuthPasswordMin *int32                  `json:"auth_password_min"`

	Company             *string `json:"company" validate:"omitempty,max=255"`
	CompanyPage         *string `json:"company_page"`
	CompanyEmail        *string `json:"company_email" validate:"omitempty,max=255"`
	CompanySupportPage  *string `json:"company_support_page"`
	CompanySupportEmail *string `json:"company_support_email" validate:"omitempty,max=255"`
	CompanyCatalog      *string `json:"company_catalog" validate:"omitempty,max=255"`

	Currency       *string `json:"currency" validate:"omitempty,max=16"`
	CurrencyDigits *int32  `json:"currency_digits"`

	ReportsMode      *model.ReportsMode  `json:"reports_mode"`
	ReportsFlags     *model.ReportsFlags `json:"reports_flags"`
	ReportsEmail     *string             `json:"reports_email" validate:"omitempty,max=255"`
	ReportsSMTP      *string             `json:"reports_smtp"`
	ReportsSMTPLogin *string             `json:"reports_smtp_login"`

	NewsMode     *model.NewsMode `json:"news_mode"`
	NewsCategory *string         `json:"news_category"`
	NewsLangs    *[]int32        `json:"news_langs"`
	MailMode     *model.MailMode `json:"mail_mode"`

	TradeFlags         *model.GroupTradeFlags `json:"trade_flags"`
	TradeInterestRate  *float64               `json:"trade_interest_rate"`
	TradeVirtualCredit *float64               `json:"trade_virtual_credit"`
	TradeTransferMode  *model.TransferMode    `json:"trade_transfer_mode"`

	MarginFreeMode       *model.FreeMarginMode       `json:"margin_free_mode"`
	MarginSOMode         *model.StopOutMode          `json:"margin_so_mode"`
	MarginCall           *float64                    `json:"margin_call"`
	MarginStopOut        *float64                    `json:"margin_stop_out"`
	MarginFreeProfitMode *model.MarginFreeProfitMode `json:"margin_free_profit_mode"`
	MarginMode           *model.MarginMode           `json:"margin_mode"`
	MarginFlags          *model.GroupMarginFlags     `json:"margin_flags"`

	DemoLeverage *int32   `json:"demo_leverage"`
	DemoDeposit  *float64 `json:"demo_deposit"`

	LimitHistory         *model.HistoryLimit `json:"limit_history"`
	LimitOrders          *int32              `json:"limit_orders"`
	LimitSymbols         *int32              `json:"limit_symbols"`
	LimitPositions       *int32              `json:"limit_positions"`
	LimitPositionsVolume *float64            `json:"limit_positions_volume"`
}

// ViewGroup is what the panel renders (flat or nested under Groups).
type ViewGroup struct {
	GroupID   int64  `json:"group_id"`
	UpdatedAt int64  `json:"updated_at"`
	Group     string `json:"group"`
	Root      bool   `json:"root"`
	ParentID  *int64 `json:"parent_id,omitempty"`
	Status    string `json:"status"`

	PermissionFlags model.PermissionsFlags `json:"permission_flags"`
	AuthMode        model.AuthMode         `json:"auth_mode"`
	AuthPasswordMin int32                  `json:"auth_password_min"`

	Company             string `json:"company"`
	CompanyPage         string `json:"company_page"`
	CompanyEmail        string `json:"company_email"`
	CompanySupportPage  string `json:"company_support_page"`
	CompanySupportEmail string `json:"company_support_email"`
	CompanyCatalog      string `json:"company_catalog"`

	Currency       string `json:"currency"`
	CurrencyDigits int32  `json:"currency_digits"`

	ReportsMode      model.ReportsMode  `json:"reports_mode"`
	ReportsFlags     model.ReportsFlags `json:"reports_flags"`
	ReportsEmail     string             `json:"reports_email"`
	ReportsSMTP      string             `json:"reports_smtp"`
	ReportsSMTPLogin string             `json:"reports_smtp_login"`

	NewsMode     model.NewsMode `json:"news_mode"`
	NewsCategory string         `json:"news_category"`
	NewsLangs    []int32        `json:"news_langs"`
	MailMode     model.MailMode `json:"mail_mode"`

	TradeFlags         model.GroupTradeFlags `json:"trade_flags"`
	TradeInterestRate  float64               `json:"trade_interest_rate"`
	TradeVirtualCredit float64               `json:"trade_virtual_credit"`
	TradeTransferMode  model.TransferMode    `json:"trade_transfer_mode"`

	MarginFreeMode       model.FreeMarginMode       `json:"margin_free_mode"`
	MarginSOMode         model.StopOutMode          `json:"margin_so_mode"`
	MarginCall           float64                    `json:"margin_call"`
	MarginStopOut        float64                    `json:"margin_stop_out"`
	MarginFreeProfitMode model.MarginFreeProfitMode `json:"margin_free_profit_mode"`
	MarginMode           model.MarginMode           `json:"margin_mode"`
	MarginFlags          model.GroupMarginFlags     `json:"margin_flags"`

	DemoLeverage int32   `json:"demo_leverage"`
	DemoDeposit  float64 `json:"demo_deposit"`

	LimitHistory         model.HistoryLimit `json:"limit_history"`
	LimitOrders          int32              `json:"limit_orders"`
	LimitSymbols         int32              `json:"limit_symbols"`
	LimitPositions       int32              `json:"limit_positions"`
	LimitPositionsVolume float64            `json:"limit_positions_volume"`

	Groups []*ViewGroup `json:"groups,omitempty"`
}

const groupColumns = `group_id, updated_at, "group", root, parent_id,
	permission_flags, auth_mode, auth_password_min,
	company, company_page, company_email, company_support_page, company_support_email, company_catalog,
	currency, currency_digits,
	reports_mode, reports_flags, reports_email, reports_smtp, reports_smtp_login,
	news_mode, news_category, news_langs, mail_mode,
	trade_flags, trade_interest_rate, trade_virtual_credit, trade_transfer_mode,
	margin_free_mode, margin_so_mode, margin_call, margin_stop_out,
	margin_free_profit_mode, margin_mode, margin_flags,
	demo_leverage, demo_deposit,
	limit_history, limit_orders, limit_symbols, limit_positions, limit_positions_volume`

func scanViewGroup(row pgx.Row) (*ViewGroup, error) {
	v := &ViewGroup{}
	err := row.Scan(
		&v.GroupID, &v.UpdatedAt, &v.Group, &v.Root, &v.ParentID,
		&v.PermissionFlags, &v.AuthMode, &v.AuthPasswordMin,
		&v.Company, &v.CompanyPage, &v.CompanyEmail, &v.CompanySupportPage, &v.CompanySupportEmail, &v.CompanyCatalog,
		&v.Currency, &v.CurrencyDigits,
		&v.ReportsMode, &v.ReportsFlags, &v.ReportsEmail, &v.ReportsSMTP, &v.ReportsSMTPLogin,
		&v.NewsMode, &v.NewsCategory, &v.NewsLangs, &v.MailMode,
		&v.TradeFlags, &v.TradeInterestRate, &v.TradeVirtualCredit, &v.TradeTransferMode,
		&v.MarginFreeMode, &v.MarginSOMode, &v.MarginCall, &v.MarginStopOut,
		&v.MarginFreeProfitMode, &v.MarginMode, &v.MarginFlags,
		&v.DemoLeverage, &v.DemoDeposit,
		&v.LimitHistory, &v.LimitOrders, &v.LimitSymbols, &v.LimitPositions, &v.LimitPositionsVolume,
	)
	if err != nil {
		return nil, err
	}
	if v.NewsLangs == nil {
		v.NewsLangs = []int32{}
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

func ptrOr[T any](p *T, def T) T {
	if p != nil {
		return *p
	}
	return def
}

// ListGroups returns the group tree, or a flat list when ?flat=1.
//
//	@Id			ListGroups
//	@Tags		Groups
//	@Produce	json
//	@Param		flat	query		int	false	"1 = flat list"
//	@Success	200		{object}	Response{data=[]ViewGroup}
//	@Failure	403		{object}	Response
//	@Failure	500		{object}	Response
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
//	@Param		id	path		int	true	"group id"
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
//	@Param		body	body		CrtGroup	true	"group path is required; other fields optional"
//	@Success	201		{object}	Response{data=ViewGroup}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	409		{object}	Response
//	@Failure	500		{object}	Response
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
	if body.PermissionFlags != nil {
		flags = *body.PermissionFlags
	} else if body.Status != "" {
		flags = model.PermissionFlagsFromStatus(body.Status)
	}

	currency := body.Currency
	if currency == "" {
		currency = "USD"
	}
	newsLangs := body.NewsLangs
	if newsLangs == nil {
		newsLangs = []int32{}
	}

	snap, _ := utils.GetClient(c)
	now := time.Now().UnixNano()

	v, err := scanViewGroup(s.DB.DB.QueryRow(c.UserContext(),
		`INSERT INTO hst.groups (
		    "group", root, parent_id,
		    permission_flags, auth_mode, auth_password_min,
		    company, company_page, company_email, company_support_page, company_support_email, company_catalog,
		    currency, currency_digits,
		    reports_mode, reports_flags, reports_email, reports_smtp, reports_smtp_login,
		    news_mode, news_category, news_langs, mail_mode,
		    trade_flags, trade_interest_rate, trade_virtual_credit, trade_transfer_mode,
		    margin_free_mode, margin_so_mode, margin_call, margin_stop_out,
		    margin_free_profit_mode, margin_mode, margin_flags,
		    demo_leverage, demo_deposit,
		    limit_history, limit_orders, limit_symbols, limit_positions, limit_positions_volume,
		    updated_at
		 ) VALUES (
		    $1,$2,$3,
		    $4,$5,$6,
		    $7,$8,$9,$10,$11,$12,
		    $13,$14,
		    $15,$16,$17,$18,$19,
		    $20,$21,$22,$23,
		    $24,$25,$26,$27,
		    $28,$29,$30,$31,
		    $32,$33,$34,
		    $35,$36,
		    $37,$38,$39,$40,$41,
		    $42
		 ) RETURNING `+groupColumns,
		path, root, body.ParentID,
		flags, ptrOr(body.AuthMode, model.AuthMode_standard), ptrOr(body.AuthPasswordMin, int32(0)),
		body.Company, body.CompanyPage, body.CompanyEmail, body.CompanySupportPage, body.CompanySupportEmail, body.CompanyCatalog,
		currency, ptrOr(body.CurrencyDigits, int32(2)),
		ptrOr(body.ReportsMode, model.ReportsMode_disabled), ptrOr(body.ReportsFlags, model.ReportsFlags_none),
		body.ReportsEmail, body.ReportsSMTP, body.ReportsSMTPLogin,
		ptrOr(body.NewsMode, model.NewsMode_disabled), body.NewsCategory, newsLangs, ptrOr(body.MailMode, model.MailMode_disabled),
		ptrOr(body.TradeFlags, model.GroupTradeFlags_none), ptrOr(body.TradeInterestRate, 0.0),
		ptrOr(body.TradeVirtualCredit, 0.0), ptrOr(body.TradeTransferMode, model.TransferMode_disabled),
		ptrOr(body.MarginFreeMode, model.FreeMarginMode_not_use_pl), ptrOr(body.MarginSOMode, model.StopOutMode_percent),
		ptrOr(body.MarginCall, 0.0), ptrOr(body.MarginStopOut, 0.0),
		ptrOr(body.MarginFreeProfitMode, model.MarginFreeProfitMode_pl),
		ptrOr(body.MarginMode, model.MarginMode_retail), ptrOr(body.MarginFlags, model.GroupMarginFlags_none),
		ptrOr(body.DemoLeverage, int32(0)), ptrOr(body.DemoDeposit, 0.0),
		ptrOr(body.LimitHistory, model.HistoryLimit_all), ptrOr(body.LimitOrders, int32(0)),
		ptrOr(body.LimitSymbols, int32(0)), ptrOr(body.LimitPositions, int32(0)), ptrOr(body.LimitPositionsVolume, 0.0),
		now))
	if err != nil {
		if utils.IsUniqueViolation(err) {
			return s.App.HttpResponseConflict(c, errs.ErrAlreadyExists)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	s.Log.Log(logger.TypeCfg, logger.CodeOK, "group created",
		"actor", snap.Login, "group_id", v.GroupID, "group", v.Group)

	return s.App.HttpResponseCreated(c, v)
}

// UpdateGroup patches group config fields.
//
//	@Id			UpdateGroup
//	@Tags		Groups
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int			true	"group id"
//	@Param		body	body		UptGroup	true	"only the fields to change"
//	@Success	200		{object}	Response{data=ViewGroup}
//	@Failure	400		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	500		{object}	Response
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
		    permission_flags         = COALESCE($2, permission_flags),
		    auth_mode                = COALESCE($3, auth_mode),
		    auth_password_min        = COALESCE($4, auth_password_min),
		    company                  = COALESCE($5, company),
		    company_page             = COALESCE($6, company_page),
		    company_email            = COALESCE($7, company_email),
		    company_support_page     = COALESCE($8, company_support_page),
		    company_support_email    = COALESCE($9, company_support_email),
		    company_catalog          = COALESCE($10, company_catalog),
		    currency                 = COALESCE($11, currency),
		    currency_digits          = COALESCE($12, currency_digits),
		    reports_mode             = COALESCE($13, reports_mode),
		    reports_flags            = COALESCE($14, reports_flags),
		    reports_email            = COALESCE($15, reports_email),
		    reports_smtp             = COALESCE($16, reports_smtp),
		    reports_smtp_login       = COALESCE($17, reports_smtp_login),
		    news_mode                = COALESCE($18, news_mode),
		    news_category            = COALESCE($19, news_category),
		    news_langs               = COALESCE($20, news_langs),
		    mail_mode                = COALESCE($21, mail_mode),
		    trade_flags              = COALESCE($22, trade_flags),
		    trade_interest_rate      = COALESCE($23, trade_interest_rate),
		    trade_virtual_credit     = COALESCE($24, trade_virtual_credit),
		    trade_transfer_mode      = COALESCE($25, trade_transfer_mode),
		    margin_free_mode         = COALESCE($26, margin_free_mode),
		    margin_so_mode           = COALESCE($27, margin_so_mode),
		    margin_call              = COALESCE($28, margin_call),
		    margin_stop_out          = COALESCE($29, margin_stop_out),
		    margin_free_profit_mode  = COALESCE($30, margin_free_profit_mode),
		    margin_mode              = COALESCE($31, margin_mode),
		    margin_flags             = COALESCE($32, margin_flags),
		    demo_leverage            = COALESCE($33, demo_leverage),
		    demo_deposit             = COALESCE($34, demo_deposit),
		    limit_history            = COALESCE($35, limit_history),
		    limit_orders             = COALESCE($36, limit_orders),
		    limit_symbols            = COALESCE($37, limit_symbols),
		    limit_positions          = COALESCE($38, limit_positions),
		    limit_positions_volume   = COALESCE($39, limit_positions_volume),
		    updated_at               = $40
		  WHERE group_id = $1
		  RETURNING `+groupColumns,
		id, flags, body.AuthMode, body.AuthPasswordMin,
		body.Company, body.CompanyPage, body.CompanyEmail, body.CompanySupportPage, body.CompanySupportEmail, body.CompanyCatalog,
		body.Currency, body.CurrencyDigits,
		body.ReportsMode, body.ReportsFlags, body.ReportsEmail, body.ReportsSMTP, body.ReportsSMTPLogin,
		body.NewsMode, body.NewsCategory, body.NewsLangs, body.MailMode,
		body.TradeFlags, body.TradeInterestRate, body.TradeVirtualCredit, body.TradeTransferMode,
		body.MarginFreeMode, body.MarginSOMode, body.MarginCall, body.MarginStopOut,
		body.MarginFreeProfitMode, body.MarginMode, body.MarginFlags,
		body.DemoLeverage, body.DemoDeposit,
		body.LimitHistory, body.LimitOrders, body.LimitSymbols, body.LimitPositions, body.LimitPositionsVolume,
		now))
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
//	@Param		id	path		int	true	"group id"
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

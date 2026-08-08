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

// CrtGroup creates a group template. Path is required; other fields optional.
type CrtGroup struct {
	Group  string `json:"group" validate:"required,max=255"`
	Status string `json:"status" validate:"omitempty,oneof=active inactive enabled disabled"`

	PermissionFlags *model.PermissionsFlags `json:"permission_flags"`
	AuthMode        *model.AuthMode         `json:"auth_mode"`
	AuthPasswordMin *int32                  `json:"auth_password_min"`

	Company             string `json:"company" validate:"max=255"`
	CompanyPage         string `json:"company_page"`
	CompanyEmail        string `json:"company_email" validate:"max=255"`
	CompanySupportPage  string `json:"company_support_page"`
	CompanySupportEmail string `json:"company_support_email" validate:"max=255"`
	CompanyCatalog      string `json:"company_catalog" validate:"max=255"`
	CompanyDeposit      string `json:"company_deposit" validate:"max=255"`
	CompanyWithdrawal   string `json:"company_withdrawal" validate:"max=255"`

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
	MarginLeverageId     *int64                      `json:"margin_leverage_id"`

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
	// Group renames in place; the sections above it are created as needed, as on create.
	Group  *string `json:"group" validate:"omitempty,max=255"`
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
	CompanyDeposit      *string `json:"company_deposit" validate:"omitempty,max=255"`
	CompanyWithdrawal   *string `json:"company_withdrawal" validate:"omitempty,max=255"`

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
	MarginLeverageId     *int64                      `json:"margin_leverage_id"`

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
	// Name is the last segment of the path, what a tree renders on the node.
	Name string `json:"name"`
	// Exists is false for a section: a node the path of some group passes through, with no group of its own.
	Exists bool   `json:"exists"`
	Status string `json:"status"`

	PermissionFlags model.PermissionsFlags `json:"permission_flags"`
	AuthMode        model.AuthMode         `json:"auth_mode"`
	AuthPasswordMin int32                  `json:"auth_password_min"`

	Company             string `json:"company"`
	CompanyPage         string `json:"company_page"`
	CompanyEmail        string `json:"company_email"`
	CompanySupportPage  string `json:"company_support_page"`
	CompanySupportEmail string `json:"company_support_email"`
	CompanyCatalog      string `json:"company_catalog"`
	CompanyDeposit      string `json:"company_deposit"`
	CompanyWithdrawal   string `json:"company_withdrawal"`

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
	MarginLeverageId     *int64                     `json:"margin_leverage_id"`

	DemoLeverage *int32   `json:"demo_leverage"`
	DemoDeposit  *float64 `json:"demo_deposit"`

	LimitHistory         model.HistoryLimit `json:"limit_history"`
	LimitOrders          int32              `json:"limit_orders"`
	LimitSymbols         int32              `json:"limit_symbols"`
	LimitPositions       int32              `json:"limit_positions"`
	LimitPositionsVolume float64            `json:"limit_positions_volume"`

	Groups []*ViewGroup `json:"groups,omitempty"`
}

const groupColumns = `group_id, updated_at, "group",
	permission_flags, auth_mode, auth_password_min,
	company, company_page, company_email, company_support_page, company_support_email, company_catalog,
	company_deposit, company_withdrawal,
	currency, currency_digits,
	reports_mode, reports_flags, reports_email, reports_smtp, reports_smtp_login,
	news_mode, news_category, news_langs, mail_mode,
	trade_flags, trade_interest_rate, trade_virtual_credit, trade_transfer_mode,
	margin_free_mode, margin_so_mode, margin_call, margin_stop_out,
	margin_free_profit_mode, margin_mode, margin_flags, margin_leverage_id,
	demo_leverage, demo_deposit,
	limit_history, limit_orders, limit_symbols, limit_positions, limit_positions_volume`

// groupPath is the name as stored: trimmed, and refused when empty or ending on a separator.
func groupPath(name string) (string, error) {
	path := strings.TrimSpace(name)
	if path == "" || strings.HasSuffix(path, `\`) {
		return "", errs.ErrRequiredParams
	}

	return path, nil
}

func scanViewGroup(row pgx.Row) (*ViewGroup, error) {
	v := &ViewGroup{}
	err := row.Scan(
		&v.GroupID, &v.UpdatedAt, &v.Group,
		&v.PermissionFlags, &v.AuthMode, &v.AuthPasswordMin,
		&v.Company, &v.CompanyPage, &v.CompanyEmail, &v.CompanySupportPage, &v.CompanySupportEmail, &v.CompanyCatalog,
		&v.CompanyDeposit, &v.CompanyWithdrawal,
		&v.Currency, &v.CurrencyDigits,
		&v.ReportsMode, &v.ReportsFlags, &v.ReportsEmail, &v.ReportsSMTP, &v.ReportsSMTPLogin,
		&v.NewsMode, &v.NewsCategory, &v.NewsLangs, &v.MailMode,
		&v.TradeFlags, &v.TradeInterestRate, &v.TradeVirtualCredit, &v.TradeTransferMode,
		&v.MarginFreeMode, &v.MarginSOMode, &v.MarginCall, &v.MarginStopOut,
		&v.MarginFreeProfitMode, &v.MarginMode, &v.MarginFlags, &v.MarginLeverageId,
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

// ViewGroupTree derives the tree from the paths, which are the only thing that describes the hierarchy.
func ViewGroupTree(flat []ViewGroup) []*ViewGroup {
	byPath := make(map[string]*ViewGroup, len(flat)*2)
	var roots []*ViewGroup

	// node returns the tree node for a path, synthesising the sections above it on the way down
	var node func(path string) *ViewGroup
	node = func(path string) *ViewGroup {
		if n, ok := byPath[path]; ok {
			return n
		}

		n := &ViewGroup{Group: path, Name: path, Exists: false}
		byPath[path] = n

		if cut := strings.LastIndex(path, model.GroupSep); cut >= 0 {
			n.Name = path[cut+len(model.GroupSep):]
			parent := node(path[:cut])
			parent.Groups = append(parent.Groups, n)
		} else {
			roots = append(roots, n)
		}

		return n
	}

	for i := range flat {
		g := flat[i]
		n := node(g.Group)

		// keep the children collected so far: a section may be reached before the group that sits at the same path
		children := n.Groups
		*n = g
		n.Groups = children
		n.Exists = true
		n.Name = g.Group
		if cut := strings.LastIndex(g.Group, model.GroupSep); cut >= 0 {
			n.Name = g.Group[cut+len(model.GroupSep):]
		}
	}

	return roots
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
func (s *Server) ListGroups(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	// a node can be a group and a section at once, so Exists says which are real
	where, args := utils.GroupAccessFor(snap.IsManager, snap.ManagerGroups, `"group"`, 1)

	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT `+groupColumns+` FROM hst.groups WHERE `+where+` ORDER BY "group"`, args...)
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
	return s.App.HttpResponseOK(c, ViewGroupTree(flat))
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
func (s *Server) GetGroup(c *fiber.Ctx) error {
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

// the one place a new group's defaults are stated; the groups table DEFAULTs mirror it
var groupDefaults = struct {
	PermissionFlags      model.PermissionsFlags
	AuthMode             model.AuthMode
	AuthPasswordMin      int32
	Currency             string
	CurrencyDigits       int32
	ReportsMode          model.ReportsMode
	ReportsFlags         model.ReportsFlags
	NewsMode             model.NewsMode
	MailMode             model.MailMode
	TradeFlags           model.GroupTradeFlags
	TradeInterestRate    float64
	TradeVirtualCredit   float64
	TradeTransferMode    model.TransferMode
	MarginFreeMode       model.FreeMarginMode
	MarginSOMode         model.StopOutMode
	MarginCall           float64
	MarginStopOut        float64
	MarginFreeProfitMode model.MarginFreeProfitMode
	MarginMode           model.MarginMode
	MarginFlags          model.GroupMarginFlags
	LimitHistory         model.HistoryLimit
	LimitOrders          int32
	LimitSymbols         int32
	LimitPositions       int32
	LimitPositionsVolume float64
	SymbolPath           string
}{
	PermissionFlags:      model.PermissionsFlags_group_default,
	AuthMode:             model.AuthMode_standard,
	AuthPasswordMin:      8,
	Currency:             "USD",
	CurrencyDigits:       2,
	ReportsMode:          model.ReportsMode_disabled,
	ReportsFlags:         model.ReportsFlags_none,
	NewsMode:             model.NewsMode_full,
	MailMode:             model.MailMode_full,
	TradeFlags:           model.GroupTradeFlags_swaps | model.GroupTradeFlags_trailing | model.GroupTradeFlags_experts | model.GroupTradeFlags_signals_all,
	TradeInterestRate:    0,
	TradeVirtualCredit:   0,
	TradeTransferMode:    model.TransferMode_disabled,
	MarginFreeMode:       model.FreeMarginMode_use_pl,
	MarginSOMode:         model.StopOutMode_percent,
	MarginCall:           50,
	MarginStopOut:        30,
	MarginFreeProfitMode: model.MarginFreeProfitMode_pl,
	MarginMode:           model.MarginMode_retail_netting,
	MarginFlags:          model.GroupMarginFlags_none,
	LimitHistory:         model.HistoryLimit_all,
	LimitOrders:          0,
	LimitSymbols:         0,
	LimitPositions:       0,
	LimitPositionsVolume: 0,
	SymbolPath:           "*",
}

// validateSOLevels refuses a margin pair the stop out would make nonsense of: in both modes the
// level falls as risk grows, so the call must sit at or above the stop, and neither may be negative.
func validateSOLevels(call, stop float64) error {
	if call < 0 || stop < 0 {
		return errors.New("margin_call and margin_stop_out must not be negative")
	}
	if stop > call {
		return errors.New("margin_stop_out must not exceed margin_call: the warning must come before the liquidation")
	}
	return nil
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
func (s *Server) CreateGroup(c *fiber.Ctx) error {
	var body CrtGroup
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	path, err := groupPath(body.Group)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	// exchange margin is accepted by the schema but the engine still nets, so it would silently lie
	if body.MarginMode != nil && *body.MarginMode == model.MarginMode_exchange {
		return s.App.HttpResponseBadRequest(c, errors.New("margin_mode exchange is not supported yet; use retail netting or retail hedging"))
	}

	if err := validateSOLevels(v1.PtrOr(body.MarginCall, groupDefaults.MarginCall),
		v1.PtrOr(body.MarginStopOut, groupDefaults.MarginStopOut)); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	flags := groupDefaults.PermissionFlags
	if body.PermissionFlags != nil {
		flags = *body.PermissionFlags
	} else if body.Status != "" {
		flags = model.PermissionFlagsFromStatus(body.Status)
	}

	currency := body.Currency
	if currency == "" {
		currency = groupDefaults.Currency
	}
	newsLangs := body.NewsLangs
	if newsLangs == nil {
		newsLangs = []int32{}
	}

	snap, _ := utils.GetClient(c)
	now := time.Now().UnixNano()

	// the group and its first scope rule are one act: a group that trades nothing is not a group
	tx, err := s.DB.DB.Begin(c.UserContext())
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(c.UserContext()) }()

	v, err := scanViewGroup(tx.QueryRow(c.UserContext(),
		`INSERT INTO hst.groups (
		    "group",
		    permission_flags, auth_mode, auth_password_min,
		    company, company_page, company_email, company_support_page, company_support_email, company_catalog,
		    company_deposit, company_withdrawal,
		    currency, currency_digits,
		    reports_mode, reports_flags, reports_email, reports_smtp, reports_smtp_login,
		    news_mode, news_category, news_langs, mail_mode,
		    trade_flags, trade_interest_rate, trade_virtual_credit, trade_transfer_mode,
		    margin_free_mode, margin_so_mode, margin_call, margin_stop_out,
		    margin_free_profit_mode, margin_mode, margin_flags, margin_leverage_id,
		    demo_leverage, demo_deposit,
		    limit_history, limit_orders, limit_symbols, limit_positions, limit_positions_volume,
		    updated_at
		 ) VALUES (
		    $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,
		    $11,$12,$13,$14,$15,$16,$17,$18,$19,$20,
		    $21,$22,$23,$24,$25,$26,$27,$28,$29,$30,
		    $31,$32,$33,$34,
		    -- zero is not a leverage profile, it is an invalid foreign key
		    CASE WHEN $35::bigint = 0 THEN NULL ELSE $35::bigint END,
		    $36,$37,$38,$39,$40,
		    $41,$42,$43
		 ) RETURNING `+groupColumns,
		path,
		flags, v1.PtrOr(body.AuthMode, groupDefaults.AuthMode), v1.PtrOr(body.AuthPasswordMin, groupDefaults.AuthPasswordMin),
		body.Company, body.CompanyPage, body.CompanyEmail, body.CompanySupportPage, body.CompanySupportEmail, body.CompanyCatalog,
		body.CompanyDeposit, body.CompanyWithdrawal,
		currency, v1.PtrOr(body.CurrencyDigits, groupDefaults.CurrencyDigits),
		v1.PtrOr(body.ReportsMode, groupDefaults.ReportsMode), v1.PtrOr(body.ReportsFlags, groupDefaults.ReportsFlags),
		body.ReportsEmail, body.ReportsSMTP, body.ReportsSMTPLogin,
		v1.PtrOr(body.NewsMode, groupDefaults.NewsMode), body.NewsCategory, newsLangs, v1.PtrOr(body.MailMode, groupDefaults.MailMode),
		v1.PtrOr(body.TradeFlags, groupDefaults.TradeFlags), v1.PtrOr(body.TradeInterestRate, groupDefaults.TradeInterestRate),
		v1.PtrOr(body.TradeVirtualCredit, groupDefaults.TradeVirtualCredit), v1.PtrOr(body.TradeTransferMode, groupDefaults.TradeTransferMode),
		v1.PtrOr(body.MarginFreeMode, groupDefaults.MarginFreeMode), v1.PtrOr(body.MarginSOMode, groupDefaults.MarginSOMode),
		v1.PtrOr(body.MarginCall, groupDefaults.MarginCall), v1.PtrOr(body.MarginStopOut, groupDefaults.MarginStopOut),
		v1.PtrOr(body.MarginFreeProfitMode, groupDefaults.MarginFreeProfitMode),
		v1.PtrOr(body.MarginMode, groupDefaults.MarginMode), v1.PtrOr(body.MarginFlags, groupDefaults.MarginFlags),
		body.MarginLeverageId,
		body.DemoLeverage, body.DemoDeposit,
		v1.PtrOr(body.LimitHistory, groupDefaults.LimitHistory), v1.PtrOr(body.LimitOrders, groupDefaults.LimitOrders),
		v1.PtrOr(body.LimitSymbols, groupDefaults.LimitSymbols), v1.PtrOr(body.LimitPositions, groupDefaults.LimitPositions),
		v1.PtrOr(body.LimitPositionsVolume, groupDefaults.LimitPositionsVolume),
		now))
	if err != nil {
		if utils.IsUniqueViolation(err) {
			return s.App.HttpResponseConflict(c, errs.ErrAlreadyExists)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// without one scope rule the group trades nothing; "*" means every instrument, all settings inherited
	if _, err := tx.Exec(c.UserContext(),
		`INSERT INTO hst.groups_symbols (group_id, updated_at, path, config_index) VALUES ($1,$2,$3,0)`,
		v.GroupID, now, groupDefaults.SymbolPath); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if err := tx.Commit(c.UserContext()); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	s.Log.Log(logger.TypeCfg, logger.CodeOK, "group created",
		"actor", snap.Login, "group_id", v.GroupID, "group", v.Group)

	// a manager allowed to create groups but granted none would otherwise create one and immediately be unable to see it
	s.grantCreatorAccess(c.UserContext(), snap.Login, v.Group)

	// every manager whose access covers this path hears about it, including the ones granted a parent long before this group existed
	s.NotifyWS(model.SubjectGroup(v.Group), model.EventGroupCreated, v)
	s.NotifySystem(model.SubjectSystemGroupCreated, v)
	s.JournalEntry(c, logger.TypeCfg, logger.CodeOK, journal.GroupCreatedMsg(snap.Login, v.Group), v)

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
func (s *Server) UpdateGroup(c *fiber.Ctx) error {
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

	// exchange margin is accepted by the schema but the engine still nets, so it would silently lie
	if body.MarginMode != nil && *body.MarginMode == model.MarginMode_exchange {
		return s.App.HttpResponseBadRequest(c, errors.New("margin_mode exchange is not supported yet; use retail netting or retail hedging"))
	}

	// the pair must stay sane after the patch, whichever half of it the patch carries
	if body.MarginCall != nil || body.MarginStopOut != nil {
		var call, stop float64
		if err := s.DB.DB.QueryRow(c.UserContext(),
			`SELECT margin_call, margin_stop_out FROM hst.groups WHERE group_id = $1`,
			id).Scan(&call, &stop); err == nil {
			if err := validateSOLevels(v1.PtrOr(body.MarginCall, call),
				v1.PtrOr(body.MarginStopOut, stop)); err != nil {
				return s.App.HttpResponseBadRequest(c, err)
			}
		}
	}

	flags := body.PermissionFlags
	if body.Status != nil && *body.Status != "" {
		f := model.PermissionFlagsFromStatus(*body.Status)
		flags = &f
	}

	snap, _ := utils.GetClient(c)
	now := time.Now().UnixNano()

	var renamed *string
	if body.Group != nil {
		path, err := groupPath(*body.Group)
		if err != nil {
			return s.App.HttpResponseBadRequest(c, err)
		}
		renamed = &path
	}

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
		    "group"                  = COALESCE($41, "group"),
		    company_deposit          = COALESCE($42, company_deposit),
		    company_withdrawal       = COALESCE($43, company_withdrawal),
		    -- zero clears the profile, which COALESCE alone cannot express
		    margin_leverage_id       = CASE WHEN $44::bigint IS NULL THEN margin_leverage_id
		                                   WHEN $44::bigint = 0 THEN NULL ELSE $44::bigint END,
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
		now, renamed, body.CompanyDeposit, body.CompanyWithdrawal, body.MarginLeverageId))
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	s.Log.Log(logger.TypeCfg, logger.CodeOK, "group updated",
		"actor", snap.Login, "group_id", v.GroupID)

	s.NotifyWS(model.SubjectGroup(v.Group), model.EventGroupUpdated, v)
	s.NotifySystem(model.SubjectSystemGroupUpdated, v)
	s.JournalEntry(c, logger.TypeCfg, logger.CodeOK, journal.GroupUpdatedMsg(snap.Login, v.Group), v)

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
func (s *Server) DeleteGroup(c *fiber.Ctx) error {
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

	// a section disappears when the last group under it goes
	var children int
	if err := s.DB.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM hst.groups WHERE starts_with("group", $1)`,
		path+`\`).Scan(&children); err != nil {
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

	ref := v1.ViewGroupRef{GroupID: id, Group: path}
	s.NotifyWS(model.SubjectGroup(path), model.EventGroupDeleted, ref)
	s.NotifySystem(model.SubjectSystemGroupDeleted, ref)
	s.JournalEntry(c, logger.TypeCfg, logger.CodeWarn, journal.GroupDeletedMsg(snap.Login, path), ref)

	return s.App.HttpResponseNoContent(c)
}

// a manager permitted to create groups but granted none would not see what it made
func (s *Server) grantCreatorAccess(ctx context.Context, login int64, path string) {
	tag, err := s.DB.DB.Exec(ctx,
		`UPDATE hst.managers
		    SET groups = ARRAY[$2], updated_at = $3
		  WHERE login = $1 AND COALESCE(cardinality(groups), 0) = 0`,
		login, path+model.GroupSep+"*", time.Now().UnixNano())
	if err != nil {
		s.Log.Log(logger.TypeCfg, logger.CodeWarn, "could not grant the creator access to its group",
			"login", login, "group", path, "error", err.Error())
		return
	}
	if tag.RowsAffected() == 0 {
		// the manager already had access somewhere, so nothing is assumed
		return
	}

	s.Log.Log(logger.TypeCfg, logger.CodeOK, "granted the creator access to its first group",
		"login", login, "group", path)

	// the access just widened, and the session carries a copy of it.
	mgr, err := s.SelectManager(ctx, login)
	if err != nil {
		s.Log.Log(logger.TypeUser, logger.CodeWarn, "could not reload the manager after granting access",
			"login", login, "error", err.Error())
		return
	}
	if err := s.OAuth2.RefreshLogin(ctx, login, mgr); err != nil {
		s.Log.Log(logger.TypeUser, logger.CodeWarn, "could not refresh the session after granting access",
			"login", login, "error", err.Error())
	}
}

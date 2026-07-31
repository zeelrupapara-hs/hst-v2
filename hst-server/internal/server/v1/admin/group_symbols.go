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

// ViewGroupSymbol is one groups_symbols override row.
type ViewGroupSymbol struct {
	SymbolID    int64  `json:"symbol_id"`
	GroupID     int64  `json:"group_id"`
	UpdatedAt   int64  `json:"updated_at"`
	Path        string `json:"path"`
	ConfigIndex int32  `json:"config_index"`

	TradeMode                      *model.TradeMode         `json:"trade_mode,omitempty"`
	ExecMode                       *model.ExecMode          `json:"exec_mode,omitempty"`
	FillFlags                      *model.FillingFlags      `json:"fill_flags,omitempty"`
	ExpirFlags                     *model.ExpirationFlags   `json:"expir_flags,omitempty"`
	SpreadDiff                     *int32                   `json:"spread_diff,omitempty"`
	SpreadDiffBalance              *int32                   `json:"spread_diff_balance,omitempty"`
	StopsLevel                     *int32                   `json:"stops_level,omitempty"`
	FreezeLevel                    *int32                   `json:"freeze_level,omitempty"`
	VolumeMin                      *int64                   `json:"volume_min,omitempty"`
	VolumeMinExt                   *int64                   `json:"volume_min_ext,omitempty"`
	VolumeMax                      *int64                   `json:"volume_max,omitempty"`
	VolumeMaxExt                   *int64                   `json:"volume_max_ext,omitempty"`
	VolumeStep                     *int64                   `json:"volume_step,omitempty"`
	VolumeStepExt                  *int64                   `json:"volume_step_ext,omitempty"`
	VolumeLimit                    *int64                   `json:"volume_limit,omitempty"`
	VolumeLimitExt                 *int64                   `json:"volume_limit_ext,omitempty"`
	MarginFlags                    *model.SymbolMarginFlags `json:"margin_flags,omitempty"`
	MarginInitial                  *float64                 `json:"margin_initial,omitempty"`
	MarginMaintenance              *float64                 `json:"margin_maintenance,omitempty"`
	MarginInitialBuy               *float64                 `json:"margin_initial_buy,omitempty"`
	MarginInitialSell              *float64                 `json:"margin_initial_sell,omitempty"`
	MarginInitialBuyLimit          *float64                 `json:"margin_initial_buy_limit,omitempty"`
	MarginInitialSellLimit         *float64                 `json:"margin_initial_sell_limit,omitempty"`
	MarginInitialBuyStop           *float64                 `json:"margin_initial_buy_stop,omitempty"`
	MarginInitialSellStop          *float64                 `json:"margin_initial_sell_stop,omitempty"`
	MarginInitialBuyStopLimit      *float64                 `json:"margin_initial_buy_stop_limit,omitempty"`
	MarginInitialSellStopLimit     *float64                 `json:"margin_initial_sell_stop_limit,omitempty"`
	MarginMaintenanceBuy           *float64                 `json:"margin_maintenance_buy,omitempty"`
	MarginMaintenanceSell          *float64                 `json:"margin_maintenance_sell,omitempty"`
	MarginMaintenanceBuyLimit      *float64                 `json:"margin_maintenance_buy_limit,omitempty"`
	MarginMaintenanceSellLimit     *float64                 `json:"margin_maintenance_sell_limit,omitempty"`
	MarginMaintenanceBuyStop       *float64                 `json:"margin_maintenance_buy_stop,omitempty"`
	MarginMaintenanceSellStop      *float64                 `json:"margin_maintenance_sell_stop,omitempty"`
	MarginMaintenanceBuyStopLimit  *float64                 `json:"margin_maintenance_buy_stop_limit,omitempty"`
	MarginMaintenanceSellStopLimit *float64                 `json:"margin_maintenance_sell_stop_limit,omitempty"`
	MarginCurrency                 *string                  `json:"margin_currency,omitempty"`
	MarginLiquidity                *float64                 `json:"margin_liquidity,omitempty"`
	MarginHedged                   *float64                 `json:"margin_hedged,omitempty"`
	SwapMode                       *model.SwapMode          `json:"swap_mode,omitempty"`
	SwapLong                       *float64                 `json:"swap_long,omitempty"`
	SwapShort                      *float64                 `json:"swap_short,omitempty"`
	SwapYearDay                    *int32                   `json:"swap_year_day,omitempty"`
	SwapFlags                      *model.SwapFlags         `json:"swap_flags,omitempty"`
	SwapRateSunday                 *float64                 `json:"swap_rate_sunday,omitempty"`
	SwapRateMonday                 *float64                 `json:"swap_rate_monday,omitempty"`
	SwapRateTuesday                *float64                 `json:"swap_rate_tuesday,omitempty"`
	SwapRateWednesday              *float64                 `json:"swap_rate_wednesday,omitempty"`
	SwapRateThursday               *float64                 `json:"swap_rate_thursday,omitempty"`
	SwapRateFriday                 *float64                 `json:"swap_rate_friday,omitempty"`
	SwapRateSaturday               *float64                 `json:"swap_rate_saturday,omitempty"`
	RETimeout                      *int32                   `json:"re_timeout,omitempty"`
	IECheckMode                    *model.InstantMode       `json:"ie_check_mode,omitempty"`
	IETimeout                      *int32                   `json:"ie_timeout,omitempty"`
	IESlipProfit                   *int32                   `json:"ie_slip_profit,omitempty"`
	IESlipLosing                   *int32                   `json:"ie_slip_losing,omitempty"`
	IEVolumeMax                    *int64                   `json:"ie_volume_max,omitempty"`
	IEVolumeMaxExt                 *int64                   `json:"ie_volume_max_ext,omitempty"`
	IEFlags                        *int32                   `json:"ie_flags,omitempty"`
	OrderFlags                     *model.OrderFlags        `json:"order_flags,omitempty"`
	PermissionsFlags               *int32                   `json:"permissions_flags,omitempty"`
	PermissionsBookDepth           *int32                   `json:"permissions_book_depth,omitempty"`
	REFlags                        *model.RequestFlags      `json:"re_flags,omitempty"`
}

// CrtGroupSymbol creates a path-mask override. Omitted override fields stay NULL (inherit).
// Provide path directly, or base_symbol_id / symbol to assign from the global catalog.
type CrtGroupSymbol struct {
	Path         string `json:"path" validate:"omitempty,max=255"`
	BaseSymbolID *int64 `json:"base_symbol_id"`
	Symbol       string `json:"symbol" validate:"omitempty,max=64"`
	ConfigIndex  *int32 `json:"config_index"`
	groupSymbolOverrides
}

// UptGroupSymbol patches an override. use_default_* clears a UI cluster back to inherit (NULL).
type UptGroupSymbol struct {
	Path        *string `json:"path" validate:"omitempty,max=255"`
	ConfigIndex *int32  `json:"config_index"`

	UseDefaultCommon     *bool `json:"use_default_common"`
	UseDefaultTrade      *bool `json:"use_default_trade"`
	UseDefaultExecution  *bool `json:"use_default_execution"`
	UseDefaultMargin     *bool `json:"use_default_margin"`
	UseDefaultMarginRate *bool `json:"use_default_margin_rate"`
	UseDefaultSwaps      *bool `json:"use_default_swaps"`

	groupSymbolOverrides
}

// groupSymbolOverrides are the sparse override columns (NULL = inherit).
type groupSymbolOverrides struct {
	TradeMode                      *model.TradeMode         `json:"trade_mode"`
	ExecMode                       *model.ExecMode          `json:"exec_mode"`
	FillFlags                      *model.FillingFlags      `json:"fill_flags"`
	ExpirFlags                     *model.ExpirationFlags   `json:"expir_flags"`
	SpreadDiff                     *int32                   `json:"spread_diff"`
	SpreadDiffBalance              *int32                   `json:"spread_diff_balance"`
	StopsLevel                     *int32                   `json:"stops_level"`
	FreezeLevel                    *int32                   `json:"freeze_level"`
	VolumeMin                      *int64                   `json:"volume_min"`
	VolumeMinExt                   *int64                   `json:"volume_min_ext"`
	VolumeMax                      *int64                   `json:"volume_max"`
	VolumeMaxExt                   *int64                   `json:"volume_max_ext"`
	VolumeStep                     *int64                   `json:"volume_step"`
	VolumeStepExt                  *int64                   `json:"volume_step_ext"`
	VolumeLimit                    *int64                   `json:"volume_limit"`
	VolumeLimitExt                 *int64                   `json:"volume_limit_ext"`
	MarginFlags                    *model.SymbolMarginFlags `json:"margin_flags"`
	MarginInitial                  *float64                 `json:"margin_initial"`
	MarginMaintenance              *float64                 `json:"margin_maintenance"`
	MarginInitialBuy               *float64                 `json:"margin_initial_buy"`
	MarginInitialSell              *float64                 `json:"margin_initial_sell"`
	MarginInitialBuyLimit          *float64                 `json:"margin_initial_buy_limit"`
	MarginInitialSellLimit         *float64                 `json:"margin_initial_sell_limit"`
	MarginInitialBuyStop           *float64                 `json:"margin_initial_buy_stop"`
	MarginInitialSellStop          *float64                 `json:"margin_initial_sell_stop"`
	MarginInitialBuyStopLimit      *float64                 `json:"margin_initial_buy_stop_limit"`
	MarginInitialSellStopLimit     *float64                 `json:"margin_initial_sell_stop_limit"`
	MarginMaintenanceBuy           *float64                 `json:"margin_maintenance_buy"`
	MarginMaintenanceSell          *float64                 `json:"margin_maintenance_sell"`
	MarginMaintenanceBuyLimit      *float64                 `json:"margin_maintenance_buy_limit"`
	MarginMaintenanceSellLimit     *float64                 `json:"margin_maintenance_sell_limit"`
	MarginMaintenanceBuyStop       *float64                 `json:"margin_maintenance_buy_stop"`
	MarginMaintenanceSellStop      *float64                 `json:"margin_maintenance_sell_stop"`
	MarginMaintenanceBuyStopLimit  *float64                 `json:"margin_maintenance_buy_stop_limit"`
	MarginMaintenanceSellStopLimit *float64                 `json:"margin_maintenance_sell_stop_limit"`
	MarginCurrency                 *string                  `json:"margin_currency"`
	MarginLiquidity                *float64                 `json:"margin_liquidity"`
	MarginHedged                   *float64                 `json:"margin_hedged"`
	SwapMode                       *model.SwapMode          `json:"swap_mode"`
	SwapLong                       *float64                 `json:"swap_long"`
	SwapShort                      *float64                 `json:"swap_short"`
	SwapYearDay                    *int32                   `json:"swap_year_day"`
	SwapFlags                      *model.SwapFlags         `json:"swap_flags"`
	SwapRateSunday                 *float64                 `json:"swap_rate_sunday"`
	SwapRateMonday                 *float64                 `json:"swap_rate_monday"`
	SwapRateTuesday                *float64                 `json:"swap_rate_tuesday"`
	SwapRateWednesday              *float64                 `json:"swap_rate_wednesday"`
	SwapRateThursday               *float64                 `json:"swap_rate_thursday"`
	SwapRateFriday                 *float64                 `json:"swap_rate_friday"`
	SwapRateSaturday               *float64                 `json:"swap_rate_saturday"`
	RETimeout                      *int32                   `json:"re_timeout"`
	IECheckMode                    *model.InstantMode       `json:"ie_check_mode"`
	IETimeout                      *int32                   `json:"ie_timeout"`
	IESlipProfit                   *int32                   `json:"ie_slip_profit"`
	IESlipLosing                   *int32                   `json:"ie_slip_losing"`
	IEVolumeMax                    *int64                   `json:"ie_volume_max"`
	IEVolumeMaxExt                 *int64                   `json:"ie_volume_max_ext"`
	IEFlags                        *int32                   `json:"ie_flags"`
	OrderFlags                     *model.OrderFlags        `json:"order_flags"`
	PermissionsFlags               *int32                   `json:"permissions_flags"`
	PermissionsBookDepth           *int32                   `json:"permissions_book_depth"`
	REFlags                        *model.RequestFlags      `json:"re_flags"`
}

const groupSymbolColumns = `symbol_id, group_id, updated_at, path, config_index,
	trade_mode, exec_mode, fill_flags, expir_flags,
	spread_diff, spread_diff_balance, stops_level, freeze_level,
	volume_min, volume_min_ext, volume_max, volume_max_ext,
	volume_step, volume_step_ext, volume_limit, volume_limit_ext,
	margin_flags, margin_initial, margin_maintenance,
	margin_initial_buy, margin_initial_sell,
	margin_initial_buy_limit, margin_initial_sell_limit,
	margin_initial_buy_stop, margin_initial_sell_stop,
	margin_initial_buy_stop_limit, margin_initial_sell_stop_limit,
	margin_maintenance_buy, margin_maintenance_sell,
	margin_maintenance_buy_limit, margin_maintenance_sell_limit,
	margin_maintenance_buy_stop, margin_maintenance_sell_stop,
	margin_maintenance_buy_stop_limit, margin_maintenance_sell_stop_limit,
	margin_currency, margin_liquidity, margin_hedged,
	swap_mode, swap_long, swap_short, swap_year_day, swap_flags,
	swap_rate_sunday, swap_rate_monday, swap_rate_tuesday, swap_rate_wednesday,
	swap_rate_thursday, swap_rate_friday, swap_rate_saturday,
	re_timeout, ie_check_mode, ie_timeout, ie_slip_profit, ie_slip_losing,
	ie_volume_max, ie_volume_max_ext, ie_flags, order_flags,
	permissions_flags, permissions_book_depth, re_flags`

func scanViewGroupSymbol(row pgx.Row) (*ViewGroupSymbol, error) {
	v := &ViewGroupSymbol{}
	err := row.Scan(
		&v.SymbolID, &v.GroupID, &v.UpdatedAt, &v.Path, &v.ConfigIndex,
		&v.TradeMode, &v.ExecMode, &v.FillFlags, &v.ExpirFlags,
		&v.SpreadDiff, &v.SpreadDiffBalance, &v.StopsLevel, &v.FreezeLevel,
		&v.VolumeMin, &v.VolumeMinExt, &v.VolumeMax, &v.VolumeMaxExt,
		&v.VolumeStep, &v.VolumeStepExt, &v.VolumeLimit, &v.VolumeLimitExt,
		&v.MarginFlags, &v.MarginInitial, &v.MarginMaintenance,
		&v.MarginInitialBuy, &v.MarginInitialSell,
		&v.MarginInitialBuyLimit, &v.MarginInitialSellLimit,
		&v.MarginInitialBuyStop, &v.MarginInitialSellStop,
		&v.MarginInitialBuyStopLimit, &v.MarginInitialSellStopLimit,
		&v.MarginMaintenanceBuy, &v.MarginMaintenanceSell,
		&v.MarginMaintenanceBuyLimit, &v.MarginMaintenanceSellLimit,
		&v.MarginMaintenanceBuyStop, &v.MarginMaintenanceSellStop,
		&v.MarginMaintenanceBuyStopLimit, &v.MarginMaintenanceSellStopLimit,
		&v.MarginCurrency, &v.MarginLiquidity, &v.MarginHedged,
		&v.SwapMode, &v.SwapLong, &v.SwapShort, &v.SwapYearDay, &v.SwapFlags,
		&v.SwapRateSunday, &v.SwapRateMonday, &v.SwapRateTuesday, &v.SwapRateWednesday,
		&v.SwapRateThursday, &v.SwapRateFriday, &v.SwapRateSaturday,
		&v.RETimeout, &v.IECheckMode, &v.IETimeout, &v.IESlipProfit, &v.IESlipLosing,
		&v.IEVolumeMax, &v.IEVolumeMaxExt, &v.IEFlags, &v.OrderFlags,
		&v.PermissionsFlags, &v.PermissionsBookDepth, &v.REFlags,
	)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// groupExists reports whether the group is there and this manager may see it.
func groupExists(c *fiber.Ctx, s *Server, groupID int) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return errs.ErrCouldNotParseClientCfg
	}

	access, args := utils.GroupAccessFor(snap.IsManager, snap.ManagerGroups, `"group"`, 2)

	var exists int
	err := s.DB.DB.QueryRow(c.UserContext(),
		`SELECT 1 FROM hst.groups WHERE group_id = $1 AND `+access,
		append([]any{groupID}, args...)...).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.ErrNotFound
	}
	return err
}

// GroupPath is the path of a group, for announcing a change to its symbols.
func (s *Server) GroupPath(c *fiber.Ctx, groupID int) string {
	var path string
	if err := s.DB.DB.QueryRow(c.UserContext(),
		`SELECT "group" FROM hst.groups WHERE group_id = $1`, groupID).Scan(&path); err != nil {
		return ""
	}
	return path
}

// ListGroupSymbols lists sparse symbol overrides for one group (NULL fields inherit from base symbol).
//
//	@Id			ListGroupSymbols
//	@Tags		Groups
//	@Produce	json
//	@Param		id	path		int	true	"group id"
//	@Success	200	{object}	Response{data=[]ViewGroupSymbol}
//	@Failure	400	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/groups/{id}/symbols [get]
func (s *Server) ListGroupSymbols(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	if err := groupExists(c, s, id); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT `+groupSymbolColumns+` FROM hst.groups_symbols
		  WHERE group_id = $1 ORDER BY config_index, path`, id)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := []ViewGroupSymbol{}
	for rows.Next() {
		v, err := scanViewGroupSymbol(rows)
		if err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		out = append(out, *v)
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}
	return s.App.HttpResponseOK(c, out)
}

// GetGroupSymbol returns one sparse override row (merge with GET /symbols/:id on the client).
//
//	@Id			GetGroupSymbol
//	@Tags		Groups
//	@Produce	json
//	@Param		id			path		int	true	"group id"
//	@Param		symbolId	path		int	true	"groups_symbols row id"
//	@Success	200			{object}	Response{data=ViewGroupSymbol}
//	@Failure	400			{object}	Response
//	@Failure	404			{object}	Response
//	@Failure	500			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/groups/{id}/symbols/{symbolId} [get]
func (s *Server) GetGroupSymbol(c *fiber.Ctx) error {
	groupID, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	symbolID, err := c.ParamsInt("symbolId")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	v, err := scanViewGroupSymbol(s.DB.DB.QueryRow(c.UserContext(),
		`SELECT `+groupSymbolColumns+` FROM hst.groups_symbols
		  WHERE group_id = $1 AND symbol_id = $2`, groupID, symbolID))
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, v)
}

func normalizeSymbolPath(p string) string {
	return strings.ReplaceAll(p, "/", `\`)
}

func (s *Server) loadBaseSymbolByName(ctx context.Context, name string) (*model.Symbol, error) {
	var id int64
	err := s.DB.DB.QueryRow(ctx,
		`SELECT symbol_id FROM hst.symbols WHERE symbol = $1`, strings.TrimSpace(name)).Scan(&id)
	if err != nil {
		return nil, err
	}
	return s.loadBaseSymbolByID(ctx, id)
}

func (s *Server) loadBaseSymbolByID(ctx context.Context, id int64) (*model.Symbol, error) {
	detail, err := s.selectSymbolDetail(ctx, id)
	if err != nil {
		return nil, err
	}
	return &detail.Symbol, nil
}

func resolveCreateGroupSymbolPath(c *fiber.Ctx, s *Server, body CrtGroupSymbol) (string, error) {
	path := normalizeSymbolPath(strings.TrimSpace(body.Path))
	if path != "" {
		return path, nil
	}
	if body.BaseSymbolID != nil {
		base, err := s.loadBaseSymbolByID(c.UserContext(), *body.BaseSymbolID)
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errs.ErrNotFound
		}
		if err != nil {
			return "", err
		}
		return base.Path, nil
	}
	if sym := strings.TrimSpace(body.Symbol); sym != "" {
		base, err := s.loadBaseSymbolByName(c.UserContext(), sym)
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errs.ErrNotFound
		}
		if err != nil {
			return "", err
		}
		return base.Path, nil
	}
	return "", errs.ErrRequiredParams
}

// CreateGroupSymbol inserts a path-mask override under a group.
//
//	@Id			CreateGroupSymbol
//	@Tags		Groups
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int				true	"group id"
//	@Param		body	body		CrtGroupSymbol	true	"path, symbol, or base_symbol_id; omitted override fields inherit"
//	@Success	201		{object}	Response{data=ViewGroupSymbol}
//	@Failure	400		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/groups/{id}/symbols [post]
func (s *Server) CreateGroupSymbol(c *fiber.Ctx) error {
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

	var body CrtGroupSymbol
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}
	path, err := resolveCreateGroupSymbolPath(c, s, body)
	if errors.Is(err, errs.ErrNotFound) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if errors.Is(err, errs.ErrRequiredParams) {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	o := body.groupSymbolOverrides
	now := time.Now().UnixNano()
	v, err := scanViewGroupSymbol(s.DB.DB.QueryRow(c.UserContext(),
		`INSERT INTO hst.groups_symbols (
		    group_id, updated_at, path, config_index,
		    trade_mode, exec_mode, fill_flags, expir_flags,
		    spread_diff, spread_diff_balance, stops_level, freeze_level,
		    volume_min, volume_min_ext, volume_max, volume_max_ext,
		    volume_step, volume_step_ext, volume_limit, volume_limit_ext,
		    margin_flags, margin_initial, margin_maintenance,
		    margin_initial_buy, margin_initial_sell,
		    margin_initial_buy_limit, margin_initial_sell_limit,
		    margin_initial_buy_stop, margin_initial_sell_stop,
		    margin_initial_buy_stop_limit, margin_initial_sell_stop_limit,
		    margin_maintenance_buy, margin_maintenance_sell,
		    margin_maintenance_buy_limit, margin_maintenance_sell_limit,
		    margin_maintenance_buy_stop, margin_maintenance_sell_stop,
		    margin_maintenance_buy_stop_limit, margin_maintenance_sell_stop_limit,
		    margin_currency, margin_liquidity, margin_hedged,
		    swap_mode, swap_long, swap_short, swap_year_day, swap_flags,
		    swap_rate_sunday, swap_rate_monday, swap_rate_tuesday, swap_rate_wednesday,
		    swap_rate_thursday, swap_rate_friday, swap_rate_saturday,
		    re_timeout, ie_check_mode, ie_timeout, ie_slip_profit, ie_slip_losing,
		    ie_volume_max, ie_volume_max_ext, ie_flags, order_flags,
		    permissions_flags, permissions_book_depth, re_flags
		 ) VALUES (
		    $1,$2,$3,$4,
		    $5,$6,$7,$8,
		    $9,$10,$11,$12,
		    $13,$14,$15,$16,
		    $17,$18,$19,$20,
		    $21,$22,$23,$24,$25,
		    $26,$27,$28,$29,$30,$31,
		    $32,$33,$34,$35,$36,$37,
		    $38,$39,$40,
		    $41,$42,$43,$44,$45,
		    $46,$47,$48,$49,$50,$51,$52,
		    $53,$54,$55,$56,$57,$58,$59,$60,$61,
		    $62,$63,$64,$65,$66
		 ) RETURNING `+groupSymbolColumns,
		groupID, now, path, v1.PtrOr(body.ConfigIndex, int32(0)),
		o.TradeMode, o.ExecMode, o.FillFlags, o.ExpirFlags,
		o.SpreadDiff, o.SpreadDiffBalance, o.StopsLevel, o.FreezeLevel,
		o.VolumeMin, o.VolumeMinExt, o.VolumeMax, o.VolumeMaxExt,
		o.VolumeStep, o.VolumeStepExt, o.VolumeLimit, o.VolumeLimitExt,
		o.MarginFlags, o.MarginInitial, o.MarginMaintenance,
		o.MarginInitialBuy, o.MarginInitialSell,
		o.MarginInitialBuyLimit, o.MarginInitialSellLimit,
		o.MarginInitialBuyStop, o.MarginInitialSellStop,
		o.MarginInitialBuyStopLimit, o.MarginInitialSellStopLimit,
		o.MarginMaintenanceBuy, o.MarginMaintenanceSell,
		o.MarginMaintenanceBuyLimit, o.MarginMaintenanceSellLimit,
		o.MarginMaintenanceBuyStop, o.MarginMaintenanceSellStop,
		o.MarginMaintenanceBuyStopLimit, o.MarginMaintenanceSellStopLimit,
		o.MarginCurrency, o.MarginLiquidity, o.MarginHedged,
		o.SwapMode, o.SwapLong, o.SwapShort, o.SwapYearDay, o.SwapFlags,
		o.SwapRateSunday, o.SwapRateMonday, o.SwapRateTuesday, o.SwapRateWednesday,
		o.SwapRateThursday, o.SwapRateFriday, o.SwapRateSaturday,
		o.RETimeout, o.IECheckMode, o.IETimeout, o.IESlipProfit, o.IESlipLosing,
		o.IEVolumeMax, o.IEVolumeMaxExt, o.IEFlags, o.OrderFlags,
		o.PermissionsFlags, o.PermissionsBookDepth, o.REFlags,
	))
	if err != nil {
		if utils.IsUniqueViolation(err) {
			return s.App.HttpResponseConflict(c, errs.ErrAlreadyExists)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "group symbol created",
		"actor", snap.Login, "group_id", groupID, "symbol_id", v.SymbolID, "path", v.Path)

	GroupPath := s.GroupPath(c, groupID)
	s.NotifyWS(model.SubjectGroupSymbol(GroupPath), model.EventGroupSymbolCreated, v)
	s.NotifySystem(model.SubjectSystemGroupSymbolCreated, v)
	s.JournalEntry(c, logger.CodeOK, journal.GroupSymbolCreatedMsg(snap.Login, GroupPath), v)

	return s.App.HttpResponseCreated(c, v)
}

// UpdateGroupSymbol patches an override; use_default_* writes NULL for that cluster.
//
//	@Id			UpdateGroupSymbol
//	@Tags		Groups
//	@Accept		json
//	@Produce	json
//	@Param		id			path		int				true	"group id"
//	@Param		symbolId	path		int				true	"groups_symbols row id"
//	@Param		body		body		UptGroupSymbol	true	"fields to change; use_default_* clears inherit clusters"
//	@Success	200			{object}	Response{data=ViewGroupSymbol}
//	@Failure	400			{object}	Response
//	@Failure	404			{object}	Response
//	@Failure	500			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/groups/{id}/symbols/{symbolId} [patch]
func (s *Server) UpdateGroupSymbol(c *fiber.Ctx) error {
	groupID, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	symbolID, err := c.ParamsInt("symbolId")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	var body UptGroupSymbol
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	o := body.groupSymbolOverrides
	now := time.Now().UnixNano()

	// CASE WHEN use_default: force NULL; else COALESCE keeps absent fields.
	v, err := scanViewGroupSymbol(s.DB.DB.QueryRow(c.UserContext(),
		`UPDATE hst.groups_symbols SET
		    path = COALESCE($3, path),
		    config_index = COALESCE($4, config_index),
		    permissions_flags = CASE WHEN $5::boolean THEN NULL ELSE COALESCE($6, permissions_flags) END,
		    permissions_book_depth = CASE WHEN $5::boolean THEN NULL ELSE COALESCE($7, permissions_book_depth) END,
		    spread_diff = CASE WHEN $5::boolean THEN NULL ELSE COALESCE($8, spread_diff) END,
		    spread_diff_balance = CASE WHEN $5::boolean THEN NULL ELSE COALESCE($9, spread_diff_balance) END,
		    volume_min = CASE WHEN $5::boolean THEN NULL ELSE COALESCE($10, volume_min) END,
		    volume_min_ext = CASE WHEN $5::boolean THEN NULL ELSE COALESCE($11, volume_min_ext) END,
		    volume_max = CASE WHEN $5::boolean THEN NULL ELSE COALESCE($12, volume_max) END,
		    volume_max_ext = CASE WHEN $5::boolean THEN NULL ELSE COALESCE($13, volume_max_ext) END,
		    volume_step = CASE WHEN $5::boolean THEN NULL ELSE COALESCE($14, volume_step) END,
		    volume_step_ext = CASE WHEN $5::boolean THEN NULL ELSE COALESCE($15, volume_step_ext) END,
		    volume_limit = CASE WHEN $5::boolean THEN NULL ELSE COALESCE($16, volume_limit) END,
		    volume_limit_ext = CASE WHEN $5::boolean THEN NULL ELSE COALESCE($17, volume_limit_ext) END,
		    trade_mode = CASE WHEN $18::boolean THEN NULL ELSE COALESCE($19, trade_mode) END,
		    fill_flags = CASE WHEN $18::boolean THEN NULL ELSE COALESCE($20, fill_flags) END,
		    expir_flags = CASE WHEN $18::boolean THEN NULL ELSE COALESCE($21, expir_flags) END,
		    order_flags = CASE WHEN $18::boolean THEN NULL ELSE COALESCE($22, order_flags) END,
		    stops_level = CASE WHEN $18::boolean THEN NULL ELSE COALESCE($23, stops_level) END,
		    freeze_level = CASE WHEN $18::boolean THEN NULL ELSE COALESCE($24, freeze_level) END,
		    exec_mode = CASE WHEN $25::boolean THEN NULL ELSE COALESCE($26, exec_mode) END,
		    re_timeout = CASE WHEN $25::boolean THEN NULL ELSE COALESCE($27, re_timeout) END,
		    re_flags = CASE WHEN $25::boolean THEN NULL ELSE COALESCE($28, re_flags) END,
		    ie_check_mode = CASE WHEN $25::boolean THEN NULL ELSE COALESCE($29, ie_check_mode) END,
		    ie_timeout = CASE WHEN $25::boolean THEN NULL ELSE COALESCE($30, ie_timeout) END,
		    ie_slip_profit = CASE WHEN $25::boolean THEN NULL ELSE COALESCE($31, ie_slip_profit) END,
		    ie_slip_losing = CASE WHEN $25::boolean THEN NULL ELSE COALESCE($32, ie_slip_losing) END,
		    ie_volume_max = CASE WHEN $25::boolean THEN NULL ELSE COALESCE($33, ie_volume_max) END,
		    ie_volume_max_ext = CASE WHEN $25::boolean THEN NULL ELSE COALESCE($34, ie_volume_max_ext) END,
		    ie_flags = CASE WHEN $25::boolean THEN NULL ELSE COALESCE($35, ie_flags) END,
		    margin_flags = CASE WHEN $36::boolean THEN NULL ELSE COALESCE($37, margin_flags) END,
		    margin_initial = CASE WHEN $36::boolean THEN NULL ELSE COALESCE($38, margin_initial) END,
		    margin_maintenance = CASE WHEN $36::boolean THEN NULL ELSE COALESCE($39, margin_maintenance) END,
		    margin_hedged = CASE WHEN $36::boolean THEN NULL ELSE COALESCE($40, margin_hedged) END,
		    margin_liquidity = CASE WHEN $41::boolean THEN NULL ELSE COALESCE($42, margin_liquidity) END,
		    margin_currency = CASE WHEN $41::boolean THEN NULL ELSE COALESCE($43, margin_currency) END,
		    margin_maintenance_buy = CASE WHEN $41::boolean THEN NULL ELSE COALESCE($44, margin_maintenance_buy) END,
		    margin_maintenance_sell = CASE WHEN $41::boolean THEN NULL ELSE COALESCE($45, margin_maintenance_sell) END,
		    margin_initial_buy = CASE WHEN $41::boolean THEN NULL ELSE COALESCE($46, margin_initial_buy) END,
		    margin_initial_sell = CASE WHEN $41::boolean THEN NULL ELSE COALESCE($47, margin_initial_sell) END,
		    margin_initial_buy_limit = CASE WHEN $41::boolean THEN NULL ELSE COALESCE($48, margin_initial_buy_limit) END,
		    margin_initial_sell_limit = CASE WHEN $41::boolean THEN NULL ELSE COALESCE($49, margin_initial_sell_limit) END,
		    margin_initial_buy_stop = CASE WHEN $41::boolean THEN NULL ELSE COALESCE($50, margin_initial_buy_stop) END,
		    margin_initial_sell_stop = CASE WHEN $41::boolean THEN NULL ELSE COALESCE($51, margin_initial_sell_stop) END,
		    margin_initial_buy_stop_limit = CASE WHEN $41::boolean THEN NULL ELSE COALESCE($52, margin_initial_buy_stop_limit) END,
		    margin_initial_sell_stop_limit = CASE WHEN $41::boolean THEN NULL ELSE COALESCE($53, margin_initial_sell_stop_limit) END,
		    margin_maintenance_buy_limit = CASE WHEN $41::boolean THEN NULL ELSE COALESCE($54, margin_maintenance_buy_limit) END,
		    margin_maintenance_sell_limit = CASE WHEN $41::boolean THEN NULL ELSE COALESCE($55, margin_maintenance_sell_limit) END,
		    margin_maintenance_buy_stop = CASE WHEN $41::boolean THEN NULL ELSE COALESCE($56, margin_maintenance_buy_stop) END,
		    margin_maintenance_sell_stop = CASE WHEN $41::boolean THEN NULL ELSE COALESCE($57, margin_maintenance_sell_stop) END,
		    margin_maintenance_buy_stop_limit = CASE WHEN $41::boolean THEN NULL ELSE COALESCE($58, margin_maintenance_buy_stop_limit) END,
		    margin_maintenance_sell_stop_limit = CASE WHEN $41::boolean THEN NULL ELSE COALESCE($59, margin_maintenance_sell_stop_limit) END,
		    swap_mode = CASE WHEN $60::boolean THEN NULL ELSE COALESCE($61, swap_mode) END,
		    swap_long = CASE WHEN $60::boolean THEN NULL ELSE COALESCE($62, swap_long) END,
		    swap_short = CASE WHEN $60::boolean THEN NULL ELSE COALESCE($63, swap_short) END,
		    swap_year_day = CASE WHEN $60::boolean THEN NULL ELSE COALESCE($64, swap_year_day) END,
		    swap_flags = CASE WHEN $60::boolean THEN NULL ELSE COALESCE($65, swap_flags) END,
		    swap_rate_sunday = CASE WHEN $60::boolean THEN NULL ELSE COALESCE($66, swap_rate_sunday) END,
		    swap_rate_monday = CASE WHEN $60::boolean THEN NULL ELSE COALESCE($67, swap_rate_monday) END,
		    swap_rate_tuesday = CASE WHEN $60::boolean THEN NULL ELSE COALESCE($68, swap_rate_tuesday) END,
		    swap_rate_wednesday = CASE WHEN $60::boolean THEN NULL ELSE COALESCE($69, swap_rate_wednesday) END,
		    swap_rate_thursday = CASE WHEN $60::boolean THEN NULL ELSE COALESCE($70, swap_rate_thursday) END,
		    swap_rate_friday = CASE WHEN $60::boolean THEN NULL ELSE COALESCE($71, swap_rate_friday) END,
		    swap_rate_saturday = CASE WHEN $60::boolean THEN NULL ELSE COALESCE($72, swap_rate_saturday) END,
		    updated_at = $73
		  WHERE group_id = $1 AND symbol_id = $2
		  RETURNING `+groupSymbolColumns,
		groupID, symbolID, body.Path, body.ConfigIndex,
		v1.PtrOr(body.UseDefaultCommon, false),
		o.PermissionsFlags, o.PermissionsBookDepth,
		o.SpreadDiff, o.SpreadDiffBalance,
		o.VolumeMin, o.VolumeMinExt, o.VolumeMax, o.VolumeMaxExt,
		o.VolumeStep, o.VolumeStepExt, o.VolumeLimit, o.VolumeLimitExt,
		v1.PtrOr(body.UseDefaultTrade, false),
		o.TradeMode, o.FillFlags, o.ExpirFlags, o.OrderFlags, o.StopsLevel, o.FreezeLevel,
		v1.PtrOr(body.UseDefaultExecution, false),
		o.ExecMode, o.RETimeout, o.REFlags, o.IECheckMode, o.IETimeout,
		o.IESlipProfit, o.IESlipLosing, o.IEVolumeMax, o.IEVolumeMaxExt, o.IEFlags,
		v1.PtrOr(body.UseDefaultMargin, false),
		o.MarginFlags, o.MarginInitial, o.MarginMaintenance, o.MarginHedged,
		v1.PtrOr(body.UseDefaultMarginRate, false),
		o.MarginLiquidity, o.MarginCurrency,
		o.MarginMaintenanceBuy, o.MarginMaintenanceSell,
		o.MarginInitialBuy, o.MarginInitialSell,
		o.MarginInitialBuyLimit, o.MarginInitialSellLimit,
		o.MarginInitialBuyStop, o.MarginInitialSellStop,
		o.MarginInitialBuyStopLimit, o.MarginInitialSellStopLimit,
		o.MarginMaintenanceBuyLimit, o.MarginMaintenanceSellLimit,
		o.MarginMaintenanceBuyStop, o.MarginMaintenanceSellStop,
		o.MarginMaintenanceBuyStopLimit, o.MarginMaintenanceSellStopLimit,
		v1.PtrOr(body.UseDefaultSwaps, false),
		o.SwapMode, o.SwapLong, o.SwapShort, o.SwapYearDay, o.SwapFlags,
		o.SwapRateSunday, o.SwapRateMonday, o.SwapRateTuesday, o.SwapRateWednesday,
		o.SwapRateThursday, o.SwapRateFriday, o.SwapRateSaturday,
		now,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "group symbol updated",
		"actor", snap.Login, "group_id", groupID, "symbol_id", symbolID)

	path := s.GroupPath(c, groupID)
	s.NotifyWS(model.SubjectGroupSymbol(path), model.EventGroupSymbolUpdated, v)
	s.NotifySystem(model.SubjectSystemGroupSymbolUpdated, v)
	s.JournalEntry(c, logger.CodeOK, journal.GroupSymbolUpdatedMsg(snap.Login, path), v)

	return s.App.HttpResponseOK(c, v)
}

// DeleteGroupSymbol removes one override row.
//
//	@Id			DeleteGroupSymbol
//	@Tags		Groups
//	@Produce	json
//	@Param		id			path		int	true	"group id"
//	@Param		symbolId	path		int	true	"groups_symbols row id"
//	@Success	204			{object}	Response
//	@Failure	400			{object}	Response
//	@Failure	404			{object}	Response
//	@Failure	500			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/groups/{id}/symbols/{symbolId} [delete]
func (s *Server) DeleteGroupSymbol(c *fiber.Ctx) error {
	groupID, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	symbolID, err := c.ParamsInt("symbolId")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	ct, err := s.DB.DB.Exec(c.UserContext(),
		`DELETE FROM hst.groups_symbols WHERE group_id = $1 AND symbol_id = $2`,
		groupID, symbolID)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if ct.RowsAffected() == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "group symbol deleted",
		"actor", snap.Login, "group_id", groupID, "symbol_id", symbolID)

	path := s.GroupPath(c, groupID)
	ref := v1.ViewGroupSymbolRef{GroupID: groupID, SymbolID: symbolID}
	s.NotifyWS(model.SubjectGroupSymbol(path), model.EventGroupSymbolDeleted, ref)
	s.NotifySystem(model.SubjectSystemGroupSymbolDeleted, ref)
	s.JournalEntry(c, logger.CodeWarn, journal.GroupSymbolDeletedMsg(snap.Login, path), ref)

	return s.App.HttpResponseNoContent(c)
}

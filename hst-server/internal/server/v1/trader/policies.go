package trader

import (
	"errors"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// ViewUIPolicy is what the terminal shell may render, decoded from rights and the group's symbol flags.
type ViewUIPolicy struct {
	Login    int64 `json:"login"`
	ReadOnly bool  `json:"read_only"`
	CanTrade bool  `json:"can_trade"`

	Enabled            bool              `json:"enabled"`
	TradeMode_disabled bool              `json:"trade_disabled"`
	ExpertAllowed      bool              `json:"expert_allowed"`
	TrailingAllowed    bool              `json:"trailing_allowed"`
	ApiAllowed         bool              `json:"api_allowed"`
	ReportsAllowed     bool              `json:"reports_allowed"`
	Rights             model.UsersRights `json:"rights"`

	CanOpenMarket    bool  `json:"can_open_market"`
	CanOpenLimit     bool  `json:"can_open_limit"`
	CanOpenStop      bool  `json:"can_open_stop"`
	CanOpenStopLimit bool  `json:"can_open_stop_limit"`
	CanSetSL         bool  `json:"can_set_sl"`
	CanSetTP         bool  `json:"can_set_tp"`
	CanCloseBy       bool  `json:"can_close_by"`
	CanClose         bool  `json:"can_close"`
	CanModify        bool  `json:"can_modify"`
	OrderFlags       int32 `json:"order_flags"`

	FillFOK   bool  `json:"fill_fok"`
	FillIOC   bool  `json:"fill_ioc"`
	FillFlags int32 `json:"fill_flags"`

	ExpiryGTC          bool  `json:"expiry_gtc"`
	ExpiryDay          bool  `json:"expiry_day"`
	ExpirySpecified    bool  `json:"expiry_specified"`
	ExpirySpecifiedDay bool  `json:"expiry_specified_day"`
	ExpiryFlags        int32 `json:"expiry_flags"`

	SymbolsCount int `json:"symbols_count"`
}

// MyUIPolicies is the permission map the terminal reads before it draws a control.
//
//	@Id			MyUIPolicies
//	@Tags		Trader
//	@Produce	json
//	@Success	200	{object}	Response{data=ViewUIPolicy}
//	@Failure	403	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/accounts/me/policies/ui [get]
func (s *Server) MyUIPolicies(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	v := ViewUIPolicy{Login: snap.Login}
	err := s.DB.DB.QueryRow(c.UserContext(),
		`SELECT rights FROM hst.users WHERE login = $1`, snap.Login).Scan(&v.Rights)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// an investor session sees everything and trades nothing, same check MyProfile makes
	v.ReadOnly = snap.Scope == int32(model.UsersPasswords_investor) || v.Rights.Has(model.UsersRights_readonly)
	v.Enabled = v.Rights.CanConnect()
	v.TradeMode_disabled = v.Rights.Has(model.UsersRights_trade_disabled)
	v.ExpertAllowed = v.Rights.Has(model.UsersRights_expert)
	v.TrailingAllowed = v.Rights.Has(model.UsersRights_trailing)
	v.ApiAllowed = v.Rights.Has(model.UsersRights_api_enabled)
	v.ReportsAllowed = v.Rights.Has(model.UsersRights_reports)
	v.CanTrade = v.Enabled && !v.TradeMode_disabled && !v.ReadOnly

	// the flags the group actually grants are per symbol, so the shell gets the union of what it may see
	symbols := s.AskEngineSymbols(c.UserContext(), snap.Login)
	v.SymbolsCount = len(symbols)
	var order, fill, expir, tradable int32
	for i := range symbols {
		order |= symbols[i].OrderFlags
		fill |= symbols[i].FillFlags
		expir |= symbols[i].ExpirFlags
		if symbols[i].TradeMode != int32(model.TradeMode_disabled) {
			tradable++
		}
	}
	v.OrderFlags, v.FillFlags, v.ExpiryFlags = order, fill, expir

	has := func(mask, flag int32) bool { return mask&flag == flag }
	v.CanOpenMarket = v.CanTrade && has(order, int32(model.OrderFlags_market))
	v.CanOpenLimit = v.CanTrade && has(order, int32(model.OrderFlags_limit))
	v.CanOpenStop = v.CanTrade && has(order, int32(model.OrderFlags_stop))
	v.CanOpenStopLimit = v.CanTrade && has(order, int32(model.OrderFlags_stop_limit))
	v.CanSetSL = v.CanTrade && has(order, int32(model.OrderFlags_sl))
	v.CanSetTP = v.CanTrade && has(order, int32(model.OrderFlags_tp))
	v.CanCloseBy = v.CanTrade && has(order, int32(model.OrderFlags_closeby))
	// closing is allowed wherever any symbol still trades, including close-only ones
	v.CanClose = v.CanTrade && tradable > 0
	v.CanModify = v.CanTrade && (v.CanSetSL || v.CanSetTP)

	v.FillFOK = has(fill, int32(model.FillingFlags_fok))
	v.FillIOC = has(fill, int32(model.FillingFlags_ioc))
	v.ExpiryGTC = has(expir, int32(model.ExpirationFlags_gtc))
	v.ExpiryDay = has(expir, int32(model.ExpirationFlags_day))
	v.ExpirySpecified = has(expir, int32(model.ExpirationFlags_specified))
	v.ExpirySpecifiedDay = has(expir, int32(model.ExpirationFlags_specified_day))

	return s.App.HttpResponseOK(c, v)
}

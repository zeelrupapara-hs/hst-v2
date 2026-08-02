package trader

import (
	"context"
	"errors"
	v1 "hstserver/internal/server/v1"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	nethttp "hstserver/pkg/http"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// ViewTraderAccount is the money state of one login, as its owner sees it.
type ViewTraderAccount struct {
	Login             int64   `json:"login"`
	Group             string  `json:"group"`
	Currency          string  `json:"currency"`
	CurrencyDigits    int32   `json:"currency_digits"`
	Balance           float64 `json:"balance"`
	Credit            float64 `json:"credit"`
	Margin            float64 `json:"margin"`
	MarginFree        float64 `json:"margin_free"`
	MarginLevel       float64 `json:"margin_level"`
	MarginLeverage    int32   `json:"margin_leverage"`
	MarginInitial     float64 `json:"margin_initial"`
	MarginMaintenance float64 `json:"margin_maintenance"`
	Profit            float64 `json:"profit"`
	Storage           float64 `json:"storage"`
	Floating          float64 `json:"floating"`
	Equity            float64 `json:"equity"`
	Assets            float64 `json:"assets"`
	Liabilities       float64 `json:"liabilities"`
	UpdatedAt         int64   `json:"updated_at"`
}

// ViewTraderProfile is what a trader may see about itself. No password column appears.
type ViewTraderProfile struct {
	Login        int64             `json:"login"`
	Group        string            `json:"group"`
	AccountType  string            `json:"account_type"`
	Name         string            `json:"name"`
	Email        string            `json:"email"`
	Phone        string            `json:"phone"`
	Country      string            `json:"country"`
	City         string            `json:"city"`
	Leverage     int32             `json:"leverage"`
	Rights       model.UsersRights `json:"rights"`
	ReadOnly     bool              `json:"read_only"`
	Registration int64             `json:"registration"`
	LastAccess   int64             `json:"last_access"`
}

// MyAccount returns the caller's own money state.
//
//	@Id			MyAccount
//	@Tags		Trader
//	@Produce	json
//	@Success	200	{object}	Response{data=ViewTraderAccount}
//	@Failure	403	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/account [get]
func (s *Server) MyAccount(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	v, status, err := s.TraderAccount(c.UserContext(), snap.Login)
	if err != nil {
		return s.App.HttpResponseStatus(c, status, err)
	}

	// the money the account is worth right now lives in the engine, not the database
	if live := s.AskEngineAccount(c.UserContext(), snap.Login); live != nil {
		v.Balance = live.Balance
		v.Credit = live.Credit
		v.Margin = live.Margin
		v.MarginFree = live.MarginFree
		v.MarginLevel = live.MarginLevel
		v.Profit = live.Profit
		v.Storage = live.Storage
		v.Floating = live.Floating
		v.Equity = live.Equity
	}

	return s.App.HttpResponseOK(c, v)
}

// TraderAccount reads one login's money state.
func (s *Server) TraderAccount(ctx context.Context, login int64) (*ViewTraderAccount, int, error) {
	v := &ViewTraderAccount{}
	err := s.DB.DB.QueryRow(ctx,
		`SELECT a.login, u."group", COALESCE(g.currency, ''), a.currency_digits,
		        a.balance, a.credit, a.margin, a.margin_free, a.margin_level, a.margin_leverage,
		        a.margin_initial, a.margin_maintenance, a.profit, a.storage, a.floating,
		        a.equity, a.assets, a.liabilities, a.updated_at
		   FROM hst.accounts a
		   JOIN hst.users u ON u.login = a.login
		   LEFT JOIN hst.groups g ON g."group" = u."group"
		  WHERE a.login = $1`, login).
		Scan(&v.Login, &v.Group, &v.Currency, &v.CurrencyDigits,
			&v.Balance, &v.Credit, &v.Margin, &v.MarginFree, &v.MarginLevel, &v.MarginLeverage,
			&v.MarginInitial, &v.MarginMaintenance, &v.Profit, &v.Storage, &v.Floating,
			&v.Equity, &v.Assets, &v.Liabilities, &v.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nethttp.StatusNotFound, errs.ErrNotFound
	}
	if err != nil {
		return nil, nethttp.StatusInternalServerError, err
	}

	return v, nethttp.StatusOK, nil
}

// MyProfile returns the caller's own record.
//
//	@Id			MyProfile
//	@Tags		Trader
//	@Produce	json
//	@Success	200	{object}	Response{data=ViewTraderProfile}
//	@Failure	403	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/profile [get]
func (s *Server) MyProfile(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	v := &ViewTraderProfile{}
	err := s.DB.DB.QueryRow(c.UserContext(),
		`SELECT login, "group", name, email, phone, country, city, leverage, rights, registration, last_access
		   FROM hst.users WHERE login = $1`, snap.Login).
		Scan(&v.Login, &v.Group, &v.Name, &v.Email, &v.Phone, &v.Country, &v.City,
			&v.Leverage, &v.Rights, &v.Registration, &v.LastAccess)

	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	v.AccountType = AccountTypeOf(v.Group)
	// an investor session sees everything and trades nothing, so the panel is told before it tries
	v.ReadOnly = snap.Scope == int32(model.UsersPasswords_investor)

	return s.App.HttpResponseOK(c, v)
}

// MySymbols falls back to the database, without limits or prices, when the engine cannot be reached.
//
//	@Id			MySymbols
//	@Tags		Trader
//	@Produce	json
//	@Success	200	{object}	Response{data=[]model.SymbolInfo}
//	@Failure	403	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/symbols [get]
func (s *Server) MySymbols(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	out, err := s.resolveMySymbols(c.UserContext(), snap.Login)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, out)
}

// AccountTypeOf is the kind of account a group path holds.
func AccountTypeOf(group string) string {
	switch {
	case v1.IsDemoGroup(group):
		return AccountTypeDemo
	case v1.IsPreliminaryGroup(group):
		return "preliminary"
	default:
		return AccountTypeReal
	}
}

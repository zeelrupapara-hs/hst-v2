package admin

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/events"
	"hstserver/pkg/logger"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// CloneSymbols copies symbols under a postfix, the way a broker opens a second book on the
// same instruments. Either a selection or a whole folder is copied.
type CloneSymbols struct {
	Postfix string  `json:"postfix" validate:"required,max=16"`
	Path    string  `json:"path" validate:"max=255"`
	CopyTo  string  `json:"copy_to" validate:"max=255"`
	Symbols []int64 `json:"symbols"`
}

// ViewClonedSymbol names one copy that was made.
type ViewClonedSymbol struct {
	SymbolId int64  `json:"symbol_id"`
	Symbol   string `json:"symbol"`
	Path     string `json:"path"`
}

// cloneColumns are every column a copy inherits; symbol_id and the two names are set by the copy.
const cloneColumns = `isin, description, international, category, exchange, cfi, sector, industry,
	country, basis, source, page, currency_base, currency_base_digits, currency_profit,
	currency_profit_digits, currency_margin, currency_margin_digits, color, color_background,
	digits, point, multiply, tick_flags, tick_book_depth, tick_book_volume, filter_soft,
	filter_soft_ticks, filter_hard, filter_hard_ticks, filter_discard, filter_spread_max,
	filter_spread_min, subscriptions_delay, trade_mode, calc_mode, exec_mode, gtc_mode,
	fill_flags, expir_flags, spread, spread_balance, spread_diff, spread_diff_balance,
	tick_value, tick_size, contract_size, stops_level, freeze_level, quotes_timeout,
	volume_min, volume_max, volume_step, volume_limit, margin_flags, margin_initial,
	margin_maintenance, margin_hedged, swap_mode, swap_long, swap_short, swap_year_day, swap_flags`

// CloneSymbol copies the selected symbols, or a folder of them, under a new postfix.
//
//	@Id			CloneSymbol
//	@Tags		Symbols
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CloneSymbols	true	"what to copy and the postfix to copy it under"
//	@Success	201		{object}	Response{data=[]ViewClonedSymbol}
//	@Failure	400		{object}	Response
//	@Failure	409		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/symbols/clone [post]
func (s *Server) CloneSymbol(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var body CloneSymbols
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	postfix := strings.TrimSpace(body.Postfix)
	if postfix == "" || (len(body.Symbols) == 0 && body.Path == "") {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	where, args := "symbol_id = ANY($1)", []any{body.Symbols}
	if len(body.Symbols) == 0 {
		where, args = "path LIKE $1", []any{body.Path + `\%`}
	}

	copyTo := strings.TrimSpace(body.CopyTo)
	args = append(args, time.Now().UnixNano())
	modIdx := len(args)
	mod := "$" + strconv.Itoa(modIdx)

	const pathSep = " || chr(92) || "

	var pathExpr, symbolExpr string
	if copyTo != "" {
		args = append(args, copyTo, postfix)
		dest, fix := "$"+strconv.Itoa(modIdx+1), "$"+strconv.Itoa(modIdx+2)
		symbolExpr = "symbol || " + fix
		pathExpr = dest + " || " + fix + pathSep + "symbol || " + fix
	} else {
		args = append(args, postfix)
		fix := "$" + strconv.Itoa(modIdx+1)
		symbolExpr = "symbol || " + fix
		pathExpr = "regexp_replace(path, '\\\\' || symbol || '$', '') || " + fix + pathSep + "symbol || " + fix
	}

	rows, err := s.DB.DB.Query(ctx,
		`INSERT INTO hst.symbols (symbol, path, date_modified, `+cloneColumns+`)
		 SELECT `+symbolExpr+`, `+pathExpr+`, `+mod+`, `+cloneColumns+`
		   FROM hst.symbols WHERE `+where+`
		 RETURNING symbol_id, symbol, path`, args...)
	if err != nil {
		if utils.IsUniqueViolation(err) {
			return s.App.HttpResponseConflict(c, errs.ErrAlreadyExists)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := []ViewClonedSymbol{}
	for rows.Next() {
		var v ViewClonedSymbol
		if err := rows.Scan(&v.SymbolId, &v.Symbol, &v.Path); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		out = append(out, v)
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "symbols cloned",
		"actor", snap.Login, "count", len(out), "postfix", postfix)

	return s.App.HttpResponseCreated(c, out)
}

// ReorderDatafeeds carries the new priority order, the feed tried first at the front.
type ReorderDatafeeds struct {
	DatafeedIds []int64 `json:"datafeed_ids" validate:"required,min=1"`
}

// ReorderDatafeed rewrites the order the feeds are tried in.
//
//	@Id			ReorderDatafeed
//	@Tags		Datafeeds
//	@Accept		json
//	@Produce	json
//	@Param		body	body		ReorderDatafeeds	true	"every feed id exactly once, in the new order"
//	@Success	204		{object}	Response
//	@Failure	400		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/datafeeds/order [put]
func (s *Server) ReorderDatafeed(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var body ReorderDatafeeds
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// feed_index is unique, so the whole list parks on negatives before taking its new places
	if _, err := tx.Exec(ctx, `UPDATE hst.datafeeds SET feed_index = -feed_index - 1`); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	now := time.Now().UnixNano()
	for i, id := range body.DatafeedIds {
		if _, err := tx.Exec(ctx,
			`UPDATE hst.datafeeds SET feed_index = $2, updated_at = $3 WHERE datafeed_id = $1`,
			id, i, now); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "data feed order changed",
		"actor", snap.Login, "count", len(body.DatafeedIds))

	s.NotifySystem(events.SubjectDatafeedUpdated, body)
	s.NotifyWS(model.SubjectDatafeed, model.EventDatafeedUpdated, body)

	return s.App.HttpResponseNoContent(c)
}

// UptTimeSettings moves the server clock.
type UptTimeSettings struct {
	TimeZone  *string `json:"time_zone" validate:"omitempty,max=64"`
	Dst       *string `json:"time_dst" validate:"omitempty,oneof=none europe usa australia"`
	NtpServer *string `json:"time_ntp_server" validate:"omitempty,max=128"`
}

// ViewTimeSettings is the server clock as configured, with the clock itself for comparison.
type ViewTimeSettings struct {
	TimeZone  string `json:"time_zone"`
	Dst       string `json:"time_dst"`
	NtpServer string `json:"time_ntp_server"`
	UpdatedAt int64  `json:"updated_at"`
	ServerNow int64  `json:"server_now"`
}

// timeSettingKeys are the clock settings, kept in hst.settings beside the end of day hour.
var timeSettingKeys = []string{"time_zone", "time_dst", "time_ntp_server"}

// GetTimeSettings reports the server time zone, daylight saving rule and time source.
//
//	@Id			GetTimeSettings
//	@Tags		System
//	@Produce	json
//	@Success	200	{object}	Response{data=ViewTimeSettings}
//	@Failure	403	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/system/time [get]
func (s *Server) GetTimeSettings(c *fiber.Ctx) error {
	v := ViewTimeSettings{TimeZone: "UTC", Dst: "none", ServerNow: time.Now().UnixNano()}

	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT key, value, updated_at FROM hst.settings WHERE key = ANY($1)`, timeSettingKeys)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		var at int64
		if err := rows.Scan(&key, &value, &at); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}

		switch key {
		case "time_zone":
			v.TimeZone = value
		case "time_dst":
			v.Dst = value
		case "time_ntp_server":
			v.NtpServer = value
		}

		if at > v.UpdatedAt {
			v.UpdatedAt = at
		}
	}

	return s.App.HttpResponseOK(c, v)
}

// UpdateTimeSettings changes the server clock settings.
//
//	@Id			UpdateTimeSettings
//	@Tags		System
//	@Accept		json
//	@Produce	json
//	@Param		body	body		UptTimeSettings	true	"the settings to change"
//	@Success	200		{object}	Response{data=ViewTimeSettings}
//	@Failure	400		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/system/time [put]
func (s *Server) UpdateTimeSettings(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var body UptTimeSettings
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	// a zone the server cannot load would silently mis-stamp every record that follows
	if body.TimeZone != nil {
		if _, err := time.LoadLocation(*body.TimeZone); err != nil {
			return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
		}
	}

	now := time.Now().UnixNano()
	for key, value := range map[string]*string{
		"time_zone": body.TimeZone, "time_dst": body.Dst, "time_ntp_server": body.NtpServer,
	} {
		if value == nil {
			continue
		}
		if _, err := s.DB.DB.Exec(ctx,
			`INSERT INTO hst.settings (key, value, updated_at) VALUES ($1,$2,$3)
			 ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at`,
			key, *value, now); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "server time settings changed", "actor", snap.Login)

	return s.GetTimeSettings(c)
}

// ResetPassword sets one of an account's passwords on the owner's behalf.
type ResetPassword struct {
	Kind     string `json:"kind" validate:"required,oneof=main investor api"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

// ResetUserPassword sets an account password. A new master password ends the account's sessions.
//
//	@Id			ResetUserPassword
//	@Tags		Users
//	@Accept		json
//	@Produce	json
//	@Param		login	path		int				true	"the account"
//	@Param		body	body		ResetPassword	true	"which password, and the new one"
//	@Success	204		{object}	Response
//	@Failure	400		{object}	Response
//	@Failure	404		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/users/{login}/password [post]
func (s *Server) ResetUserPassword(c *fiber.Ctx) error {
	ctx := c.UserContext()

	login, err := c.ParamsInt("login")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	var body ResetPassword
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	if err := s.RequirePasswordLength(ctx, int64(login), "", body.Password); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	hash, err := s.OAuth2.Hasher.HashPassword(body.Password)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	column := map[string]string{
		"main": "password_main", "investor": "password_investor", "api": "password_api",
	}[body.Kind]

	var found int64
	err = s.DB.DB.QueryRow(ctx,
		`UPDATE hst.users SET `+column+` = $2, last_pass_change = $3, updated_at = $3
		  WHERE login = $1 RETURNING login`,
		login, hash, time.Now().UnixNano()).Scan(&found)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeUser, logger.CodeWarn, "password reset by staff",
		"actor", snap.Login, "login", login, "kind", body.Kind)
	s.JournalEntry(c, model.JournalType_auth, logger.CodeWarn,
		body.Kind+" password of account #"+strconv.FormatInt(int64(login), 10)+" was reset", nil)

	return s.App.HttpResponseNoContent(c)
}

package v1

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"hstserver/model"
	"hstserver/pkg/datafeedmodules"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/events"
	"hstserver/pkg/logger"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

var (
	datafeedsSortable = utils.NewSortable("datafeed_id", "name", "enable", "updated_at", "module")
	datafeedNameBad   = regexp.MustCompile(`[?*<>]`)
	feederModeMask    = int32(model.FeederFlags_quotes | model.FeederFlags_news | model.FeederFlags_remote)
)

// CrtDatafeed creates a data feed configuration.
type CrtDatafeed struct {
	Name             string                `json:"name" validate:"required,max=64"`
	Module           string                `json:"module" validate:"required,max=128"`
	Enable           *model.DatafeedEnable `json:"enable"`
	Mode             *model.FeederFlags    `json:"mode"`
	GatewayServer    string                `json:"gateway_server" validate:"max=255"`
	FeedServer       string                `json:"feed_server" validate:"max=255"`
	FeedLogin        int64                 `json:"feed_login"`
	FeedPassword     string                `json:"feed_password"`
	GatewayLogin     int64                 `json:"gateway_login"`
	GatewayPassword  string                `json:"gateway_password"`
	Symbols          string                `json:"symbols"`
	Timeout          *int32                `json:"timeout"`
	TimeoutReconnect *int32                `json:"timeout_reconnect"`
	TimeoutSleep     *int32                `json:"timeout_sleep"`
	AttemptsSleep    *int32                `json:"attempts_sleep"`
	Company          string                `json:"company" validate:"max=255"`
	Issuer           string                `json:"issuer" validate:"max=255"`
}

// UptDatafeed patches a data feed. Passwords update only when sent.
type UptDatafeed struct {
	Name             *string               `json:"name" validate:"omitempty,max=64"`
	Module           *string               `json:"module" validate:"omitempty,max=128"`
	Enable           *model.DatafeedEnable `json:"enable"`
	Mode             *model.FeederFlags    `json:"mode"`
	GatewayServer    *string               `json:"gateway_server" validate:"omitempty,max=255"`
	FeedServer       *string               `json:"feed_server" validate:"omitempty,max=255"`
	FeedLogin        *int64                `json:"feed_login"`
	FeedPassword     *string               `json:"feed_password"`
	GatewayLogin     *int64                `json:"gateway_login"`
	GatewayPassword  *string               `json:"gateway_password"`
	Symbols          *string               `json:"symbols"`
	Timeout          *int32                `json:"timeout"`
	TimeoutReconnect *int32                `json:"timeout_reconnect"`
	TimeoutSleep     *int32                `json:"timeout_sleep"`
	AttemptsSleep    *int32                `json:"attempts_sleep"`
	Company          *string               `json:"company" validate:"omitempty,max=255"`
	Issuer           *string               `json:"issuer" validate:"omitempty,max=255"`
}

// ViewDatafeed is a list row without nested children or passwords.
type ViewDatafeed struct {
	DatafeedID       int64                       `json:"datafeed_id"`
	Name             string                      `json:"name"`
	Module           string                      `json:"module"`
	Enable           model.DatafeedEnable        `json:"enable"`
	Mode             model.FeederFlags           `json:"mode"`
	GatewayServer    string                      `json:"gateway_server"`
	FeedServer       string                      `json:"feed_server"`
	FeedLogin        int64                       `json:"feed_login"`
	GatewayLogin     int64                       `json:"gateway_login"`
	Symbols          string                      `json:"symbols"`
	Timeout          int32                       `json:"timeout"`
	TimeoutReconnect int32                       `json:"timeout_reconnect"`
	TimeoutSleep     int32                       `json:"timeout_sleep"`
	AttemptsSleep    int32                       `json:"attempts_sleep"`
	UpdatedAt        int64                       `json:"updated_at"`
	Company          string                      `json:"company"`
	Issuer           string                      `json:"issuer"`
	SysConnection    model.DatafeedSysConnection `json:"sys_connection"`
	SysLastTime      int64                       `json:"sys_last_time"`
	TickStatsCount   int64                       `json:"tick_stats_count"`
	TicksCount       int64                       `json:"ticks_count"`
	BooksCount       int64                       `json:"books_count"`
	NewsCount        int64                       `json:"news_count"`
	BytesReceived    int64                       `json:"bytes_received"`
	BytesSent        int64                       `json:"bytes_sent"`
	StateFlags       int32                       `json:"state_flags"`
}

// ViewDatafeedDetail includes nested params and translations.
type ViewDatafeedDetail struct {
	ViewDatafeed
	Params     []ViewDatafeedParam     `json:"params"`
	Translates []ViewDatafeedTranslate `json:"translates"`
}

// CrtDatafeedParam creates an additional feed parameter.
type CrtDatafeedParam struct {
	ParamKey string                 `json:"param_key" validate:"required,max=64"`
	Type     *model.FeederParamType `json:"type"`
	Name     string                 `json:"name" validate:"required,max=128"`
	Value    string                 `json:"value"`
}

// UptDatafeedParam patches a feed parameter.
type UptDatafeedParam struct {
	ParamKey *string                `json:"param_key" validate:"omitempty,max=64"`
	Type     *model.FeederParamType `json:"type"`
	Name     *string                `json:"name" validate:"omitempty,max=128"`
	Value    *string                `json:"value"`
}

// ViewDatafeedParam is one parameter row.
type ViewDatafeedParam struct {
	ParamID    int64                 `json:"param_id"`
	DatafeedID int64                 `json:"datafeed_id"`
	ParamKey   string                `json:"param_key"`
	Type       model.FeederParamType `json:"type"`
	Name       string                `json:"name"`
	Value      string                `json:"value"`
}

// CrtDatafeedTranslate creates a symbol translation rule.
type CrtDatafeedTranslate struct {
	Symbol    string `json:"symbol" validate:"required,max=255"`
	Source    string `json:"source" validate:"max=255"`
	BidMarkup int32  `json:"bid_markup"`
	AskMarkup int32  `json:"ask_markup"`
	Digits    int16  `json:"digits" validate:"gte=0,lte=12"`
}

// UptDatafeedTranslate patches a symbol translation rule.
type UptDatafeedTranslate struct {
	Symbol    *string `json:"symbol" validate:"omitempty,max=255"`
	Source    *string `json:"source" validate:"omitempty,max=255"`
	BidMarkup *int32  `json:"bid_markup"`
	AskMarkup *int32  `json:"ask_markup"`
	Digits    *int16  `json:"digits" validate:"omitempty,gte=0,lte=12"`
}

// ViewDatafeedTranslate is one translation row.
type ViewDatafeedTranslate struct {
	TranslateID int64  `json:"translate_id"`
	DatafeedID  int64  `json:"datafeed_id"`
	Symbol      string `json:"symbol"`
	Source      string `json:"source"`
	BidMarkup   int32  `json:"bid_markup"`
	AskMarkup   int32  `json:"ask_markup"`
	Digits      int16  `json:"digits"`
}

const datafeedColumns = `datafeed_id, name, module, enable, mode,
	gateway_server, feed_server, feed_login, gateway_login, symbols,
	timeout, timeout_reconnect, timeout_sleep, attempts_sleep, updated_at,
	company, issuer, sys_connection, sys_last_time,
	tick_stats_count, ticks_count, books_count, news_count,
	bytes_received, bytes_sent, state_flags`

const datafeedParamColumns = `param_id, datafeed_id, param_key, type, name, value`

const datafeedTranslateColumns = `translate_id, datafeed_id, symbol, source, bid_markup, ask_markup, digits`

func scanViewDatafeed(row pgx.Row) (*ViewDatafeed, error) {
	v := &ViewDatafeed{}
	err := row.Scan(
		&v.DatafeedID, &v.Name, &v.Module, &v.Enable, &v.Mode,
		&v.GatewayServer, &v.FeedServer, &v.FeedLogin, &v.GatewayLogin, &v.Symbols,
		&v.Timeout, &v.TimeoutReconnect, &v.TimeoutSleep, &v.AttemptsSleep, &v.UpdatedAt,
		&v.Company, &v.Issuer, &v.SysConnection, &v.SysLastTime,
		&v.TickStatsCount, &v.TicksCount, &v.BooksCount, &v.NewsCount,
		&v.BytesReceived, &v.BytesSent, &v.StateFlags,
	)
	return v, err
}

func scanViewDatafeedParam(row pgx.Row) (*ViewDatafeedParam, error) {
	v := &ViewDatafeedParam{}
	err := row.Scan(&v.ParamID, &v.DatafeedID, &v.ParamKey, &v.Type, &v.Name, &v.Value)
	return v, err
}

func scanViewDatafeedTranslate(row pgx.Row) (*ViewDatafeedTranslate, error) {
	v := &ViewDatafeedTranslate{}
	err := row.Scan(
		&v.TranslateID, &v.DatafeedID, &v.Symbol, &v.Source,
		&v.BidMarkup, &v.AskMarkup, &v.Digits,
	)
	return v, err
}

func validateDatafeedName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errs.ErrRequiredParams
	}
	if datafeedNameBad.MatchString(name) {
		return errors.New("name must not contain ? * < >")
	}
	return nil
}

func validateFeederMode(mode model.FeederFlags) error {
	if int32(mode)&^feederModeMask != 0 {
		return errors.New("mode contains unknown flag bits")
	}
	if mode == 0 {
		return errors.New("mode must include at least one of quotes, news, or remote")
	}
	return nil
}

func validateDatafeedModule(module string, mode model.FeederFlags) (string, error) {
	module = strings.TrimSpace(module)
	if err := datafeedmodules.Validate(module, mode); err != nil {
		return "", err
	}
	if canon, ok := datafeedmodules.Canonical(module); ok {
		return canon, nil
	}
	return module, nil
}

func (s *HttpServer) resolveDatafeedModuleMode(ctx context.Context, id int64, bodyModule *string, bodyMode *model.FeederFlags) (string, model.FeederFlags, error) {
	var curModule string
	var curMode model.FeederFlags
	err := s.DB.DB.QueryRow(ctx,
		`SELECT module, mode FROM hst.datafeeds WHERE datafeed_id = $1`, id).
		Scan(&curModule, &curMode)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", 0, errs.ErrNotFound
	}
	if err != nil {
		return "", 0, err
	}

	module := curModule
	if bodyModule != nil {
		module = strings.TrimSpace(*bodyModule)
	}
	mode := curMode
	if bodyMode != nil {
		mode = *bodyMode
	}
	return module, mode, nil
}

func validateGatewayPassword(pw string) error {
	if pw == "" {
		return nil
	}
	if len(pw) < 6 {
		return errors.New("gateway_password must be at least 6 characters")
	}
	var upper, lower, digit bool
	for _, r := range pw {
		switch {
		case unicode.IsUpper(r):
			upper = true
		case unicode.IsLower(r):
			lower = true
		case unicode.IsDigit(r):
			digit = true
		}
	}
	kinds := 0
	if upper {
		kinds++
	}
	if lower {
		kinds++
	}
	if digit {
		kinds++
	}
	if kinds < 2 {
		return errors.New("gateway_password must include two of upper, lower, and digit")
	}
	return nil
}

func validateGatewayLogin(login int64) error {
	if login < 0 {
		return errors.New("gateway_login must be non-negative")
	}
	return nil
}

func datafeedExists(c *fiber.Ctx, s *HttpServer, datafeedID int) error {
	var exists int
	err := s.DB.DB.QueryRow(c.UserContext(),
		`SELECT 1 FROM hst.datafeeds WHERE datafeed_id = $1`, datafeedID).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.ErrNotFound
	}
	return err
}

func loadDatafeedParams(ctx context.Context, s *HttpServer, datafeedID int64) ([]ViewDatafeedParam, error) {
	rows, err := s.DB.DB.Query(ctx,
		`SELECT `+datafeedParamColumns+` FROM hst.datafeed_params
		  WHERE datafeed_id = $1 ORDER BY param_id`, datafeedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ViewDatafeedParam{}
	for rows.Next() {
		v, err := scanViewDatafeedParam(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *v)
	}
	return out, rows.Err()
}

func loadDatafeedTranslates(ctx context.Context, s *HttpServer, datafeedID int64) ([]ViewDatafeedTranslate, error) {
	rows, err := s.DB.DB.Query(ctx,
		`SELECT `+datafeedTranslateColumns+` FROM hst.datafeed_translates
		  WHERE datafeed_id = $1 ORDER BY translate_id`, datafeedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ViewDatafeedTranslate{}
	for rows.Next() {
		v, err := scanViewDatafeedTranslate(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *v)
	}
	return out, rows.Err()
}

func selectDatafeedDetail(ctx context.Context, s *HttpServer, datafeedID int64) (*ViewDatafeedDetail, error) {
	v, err := scanViewDatafeed(s.DB.DB.QueryRow(ctx,
		`SELECT `+datafeedColumns+` FROM hst.datafeeds WHERE datafeed_id = $1`, datafeedID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	params, err := loadDatafeedParams(ctx, s, datafeedID)
	if err != nil {
		return nil, err
	}
	translates, err := loadDatafeedTranslates(ctx, s, datafeedID)
	if err != nil {
		return nil, err
	}

	return &ViewDatafeedDetail{
		ViewDatafeed: *v,
		Params:       params,
		Translates:   translates,
	}, nil
}

// ListDatafeeds returns a page of data feed configurations.
//
//	@Id			ListDatafeeds
//	@Tags		Datafeeds
//	@Produce	json
//	@Param		page	query		int		false	"page number, from 1"
//	@Param		limit	query		int		false	"rows per page, max 500"
//	@Param		search	query		string	false	"matches name or module"
//	@Param		sort_by	query		string	false	"datafeed_id, name, enable, updated_at, module"	Enums(datafeed_id, name, enable, updated_at, module)
//	@Param		order	query		string	false	"asc or desc"									Enums(asc, desc)
//	@Success	200		{object}	Response{data=[]ViewDatafeed}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/datafeeds [get]
func (s *HttpServer) ListDatafeeds(c *fiber.Ctx) error {
	q, err := utils.QueryFilter(c, datafeedsSortable, "datafeed_id")
	if err != nil {
		return s.App.HttpResponseBadQueryParams(c, err)
	}

	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT `+datafeedColumns+`
		   FROM hst.datafeeds
		  WHERE ($1 = '' OR name ILIKE '%'||$1||'%' OR module ILIKE '%'||$1||'%')
		  ORDER BY `+q.SortBy+`
		  LIMIT $2 OFFSET $3`, q.Search, q.Limit, q.Offset)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := []ViewDatafeed{}
	for rows.Next() {
		v, err := scanViewDatafeed(rows)
		if err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		out = append(out, *v)
	}
	if err := rows.Err(); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, out)
}

// ListDatafeedModules returns supported feeder modules for a mode (quotes or news).
//
//	@Id			ListDatafeedModules
//	@Tags		Datafeeds
//	@Produce	json
//	@Param		mode	query		int	true	"1=quotes 2=news"
//	@Success	200		{object}	Response{data=[]datafeedmodules.Option}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/datafeeds/modules [get]
func (s *HttpServer) ListDatafeedModules(c *fiber.Ctx) error {
	raw := c.Query("mode")
	if raw == "" {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	v, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return s.App.HttpResponseBadQueryParams(c, err)
	}
	mode := model.FeederFlags(v)
	if err := validateFeederMode(mode); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	opts := datafeedmodules.Options(mode)
	if len(opts) == 0 {
		return s.App.HttpResponseBadRequest(c,
			errors.New("mode must be quotes-only (1) or news-only (2)"))
	}
	return s.App.HttpResponseOK(c, opts)
}

// CreateDatafeed registers a data feed configuration.
//
//	@Id			CreateDatafeed
//	@Tags		Datafeeds
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CrtDatafeed	true	"name and module required"
//	@Success	201		{object}	Response{data=ViewDatafeedDetail}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	409		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/datafeeds [post]
func (s *HttpServer) CreateDatafeed(c *fiber.Ctx) error {
	var body CrtDatafeed
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	name := strings.TrimSpace(body.Name)
	if err := validateDatafeedName(name); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	module := strings.TrimSpace(body.Module)
	if module == "" {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	mode := ptrOr(body.Mode, model.FeederFlags_quotes)
	if err := validateFeederMode(mode); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := validateGatewayLogin(body.GatewayLogin); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := validateGatewayPassword(body.GatewayPassword); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	canonicalModule, err := validateDatafeedModule(module, mode)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	module = canonicalModule

	ctx := c.UserContext()
	now := time.Now().UnixNano()
	var id int64
	err = s.DB.DB.QueryRow(ctx,
		`INSERT INTO hst.datafeeds (
		    name, module, enable, mode,
		    gateway_server, feed_server, feed_login, feed_password,
		    gateway_login, gateway_password, symbols,
		    timeout, timeout_reconnect, timeout_sleep, attempts_sleep,
		    updated_at, company, issuer
		 ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
		 RETURNING datafeed_id`,
		name, module,
		ptrOr(body.Enable, model.DatafeedEnable_enabled),
		mode,
		body.GatewayServer, body.FeedServer,
		body.FeedLogin, body.FeedPassword,
		body.GatewayLogin, body.GatewayPassword,
		body.Symbols,
		ptrOr(body.Timeout, 0),
		ptrOr(body.TimeoutReconnect, 5),
		ptrOr(body.TimeoutSleep, 60),
		ptrOr(body.AttemptsSleep, 10),
		now, body.Company, body.Issuer,
	).Scan(&id)
	if err != nil {
		if utils.IsUniqueViolation(err) {
			return s.App.HttpResponseConflict(c, errs.ErrAlreadyExists)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	detail, err := selectDatafeedDetail(ctx, s, id)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "datafeed created",
		"actor", snap.Login, "target", name, "datafeed_id", id)

	s.publishDatafeedEvent(events.SubjectDatafeedCreated, int64(id), detail.Enable, detail.Mode)
	s.publishWorkerConfigSnapshot(ctx, int64(id))

	return s.App.HttpResponseCreated(c, detail)
}

// GetDatafeed returns one data feed with params and translations.
//
//	@Id			GetDatafeed
//	@Tags		Datafeeds
//	@Produce	json
//	@Param		id	path		int	true	"datafeed id"
//	@Success	200	{object}	Response{data=ViewDatafeedDetail}
//	@Failure	400	{object}	Response
//	@Failure	403	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/datafeeds/{id} [get]
func (s *HttpServer) GetDatafeed(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	detail, err := selectDatafeedDetail(c.UserContext(), s, int64(id))
	if errors.Is(err, errs.ErrNotFound) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, detail)
}

// UpdateDatafeed patches a data feed configuration.
//
//	@Id			UpdateDatafeed
//	@Tags		Datafeeds
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int			true	"datafeed id"
//	@Param		body	body		UptDatafeed	true	"only the fields to change"
//	@Success	200		{object}	Response{data=ViewDatafeedDetail}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	409		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/datafeeds/{id} [patch]
func (s *HttpServer) UpdateDatafeed(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	var body UptDatafeed
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	if body.Name != nil {
		if err := validateDatafeedName(*body.Name); err != nil {
			return s.App.HttpResponseBadRequest(c, err)
		}
	}
	if body.Mode != nil {
		if err := validateFeederMode(*body.Mode); err != nil {
			return s.App.HttpResponseBadRequest(c, err)
		}
	}
	if body.GatewayLogin != nil {
		if err := validateGatewayLogin(*body.GatewayLogin); err != nil {
			return s.App.HttpResponseBadRequest(c, err)
		}
	}
	if body.GatewayPassword != nil {
		if err := validateGatewayPassword(*body.GatewayPassword); err != nil {
			return s.App.HttpResponseBadRequest(c, err)
		}
	}

	ctx := c.UserContext()
	var module string
	var mode model.FeederFlags
	module, mode, err = s.resolveDatafeedModuleMode(ctx, int64(id), body.Module, body.Mode)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if _, err := validateDatafeedModule(module, mode); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	var modulePatch *string
	if body.Module != nil {
		canon, _ := datafeedmodules.Canonical(strings.TrimSpace(*body.Module))
		modulePatch = &canon
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx,
		`UPDATE hst.datafeeds SET
		    name               = COALESCE($2, name),
		    module             = COALESCE($3, module),
		    enable             = COALESCE($4, enable),
		    mode               = COALESCE($5, mode),
		    gateway_server     = COALESCE($6, gateway_server),
		    feed_server        = COALESCE($7, feed_server),
		    feed_login         = COALESCE($8, feed_login),
		    gateway_login      = COALESCE($9, gateway_login),
		    symbols            = COALESCE($10, symbols),
		    timeout            = COALESCE($11, timeout),
		    timeout_reconnect  = COALESCE($12, timeout_reconnect),
		    timeout_sleep      = COALESCE($13, timeout_sleep),
		    attempts_sleep     = COALESCE($14, attempts_sleep),
		    company            = COALESCE($15, company),
		    issuer             = COALESCE($16, issuer),
		    updated_at         = $17
		  WHERE datafeed_id = $1`,
		id, body.Name, modulePatch, body.Enable, body.Mode,
		body.GatewayServer, body.FeedServer, body.FeedLogin, body.GatewayLogin,
		body.Symbols, body.Timeout, body.TimeoutReconnect, body.TimeoutSleep, body.AttemptsSleep,
		body.Company, body.Issuer, time.Now().UnixNano(),
	)
	if err != nil {
		if utils.IsUniqueViolation(err) {
			return s.App.HttpResponseConflict(c, errs.ErrAlreadyExists)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if tag.RowsAffected() == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	if body.FeedPassword != nil {
		if _, err := tx.Exec(ctx,
			`UPDATE hst.datafeeds SET feed_password = $2 WHERE datafeed_id = $1`,
			id, *body.FeedPassword); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}
	if body.GatewayPassword != nil {
		if _, err := tx.Exec(ctx,
			`UPDATE hst.datafeeds SET gateway_password = $2 WHERE datafeed_id = $1`,
			id, *body.GatewayPassword); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	detail, err := selectDatafeedDetail(ctx, s, int64(id))
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "datafeed updated",
		"actor", snap.Login, "target", detail.Name, "datafeed_id", id)

	s.publishDatafeedEvent(events.SubjectDatafeedUpdated, int64(id), detail.Enable, detail.Mode)
	s.publishWorkerConfigSnapshot(ctx, int64(id))

	return s.App.HttpResponseOK(c, detail)
}

// DeleteDatafeed removes a data feed and its params and translations.
//
//	@Id			DeleteDatafeed
//	@Tags		Datafeeds
//	@Produce	json
//	@Param		id	path		int	true	"datafeed id"
//	@Success	204	{object}	Response
//	@Failure	400	{object}	Response
//	@Failure	403	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/datafeeds/{id} [delete]
func (s *HttpServer) DeleteDatafeed(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	var name string
	var mode model.FeederFlags
	err = s.DB.DB.QueryRow(c.UserContext(),
		`DELETE FROM hst.datafeeds WHERE datafeed_id = $1 RETURNING name, mode`, id).Scan(&name, &mode)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeWarn, "datafeed deleted",
		"actor", snap.Login, "target", name, "datafeed_id", id)

	s.publishDatafeedEvent(events.SubjectDatafeedDeleted, int64(id), model.DatafeedEnable_disabled, mode)

	return s.App.HttpResponseNoContent(c)
}

// ActivateDatafeed enables a data feed and notifies ingestion workers.
//
//	@Id			ActivateDatafeed
//	@Tags		Datafeeds
//	@Produce	json
//	@Param		id	path		int	true	"datafeed id"
//	@Success	200	{object}	Response{data=ViewDatafeedDetail}
//	@Failure	400	{object}	Response
//	@Failure	403	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/datafeeds/{id}/activate [post]
func (s *HttpServer) ActivateDatafeed(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	ctx := c.UserContext()
	tag, err := s.DB.DB.Exec(ctx,
		`UPDATE hst.datafeeds SET enable = $2, updated_at = $3 WHERE datafeed_id = $1`,
		id, model.DatafeedEnable_enabled, time.Now().UnixNano(),
	)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if tag.RowsAffected() == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	detail, err := selectDatafeedDetail(ctx, s, int64(id))
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "datafeed activated",
		"actor", snap.Login, "target", detail.Name, "datafeed_id", id)

	s.publishDatafeedEvent(events.SubjectDatafeedUpdated, int64(id), detail.Enable, detail.Mode)
	s.publishWorkerConfigSnapshot(ctx, int64(id))

	return s.App.HttpResponseOK(c, detail)
}

func (s *HttpServer) publishDatafeedEvent(subject string, datafeedID int64, enable model.DatafeedEnable, mode model.FeederFlags) {
	if err := events.PublishDatafeed(s.Nats, subject, datafeedID, mode, enable); err != nil {
		s.Log.Log(logger.TypeNet, logger.CodeWarn, "datafeed event publish failed",
			"subject", subject, "datafeed_id", datafeedID, "error", err.Error())
	}
}

func (s *HttpServer) notifyDatafeedConfigChanged(datafeedID int64) {
	ctx := context.Background()
	v, err := scanViewDatafeed(s.DB.DB.QueryRow(ctx,
		`SELECT `+datafeedColumns+` FROM hst.datafeeds WHERE datafeed_id = $1`, datafeedID))
	if err != nil {
		return
	}
	s.publishDatafeedEvent(events.SubjectDatafeedUpdated, datafeedID, v.Enable, v.Mode)
	s.publishWorkerConfigSnapshot(ctx, datafeedID)
}

// ListDatafeedParams lists parameters for one data feed.
//
//	@Id			ListDatafeedParams
//	@Tags		Datafeeds
//	@Produce	json
//	@Param		id	path		int	true	"datafeed id"
//	@Success	200	{object}	Response{data=[]ViewDatafeedParam}
//	@Failure	400	{object}	Response
//	@Failure	403	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/datafeeds/{id}/params [get]
func (s *HttpServer) ListDatafeedParams(c *fiber.Ctx) error {
	datafeedID, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	if err := datafeedExists(c, s, datafeedID); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	params, err := loadDatafeedParams(c.UserContext(), s, int64(datafeedID))
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, params)
}

// CreateDatafeedParam adds a parameter to a data feed.
//
//	@Id			CreateDatafeedParam
//	@Tags		Datafeeds
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int					true	"datafeed id"
//	@Param		body	body		CrtDatafeedParam	true	"param_key and name required"
//	@Success	201		{object}	Response{data=ViewDatafeedParam}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	409		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/datafeeds/{id}/params [post]
func (s *HttpServer) CreateDatafeedParam(c *fiber.Ctx) error {
	datafeedID, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	if err := datafeedExists(c, s, datafeedID); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	var body CrtDatafeedParam
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	paramKey := strings.TrimSpace(body.ParamKey)
	if paramKey == "" {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	v, err := scanViewDatafeedParam(s.DB.DB.QueryRow(c.UserContext(),
		`INSERT INTO hst.datafeed_params (datafeed_id, param_key, type, name, value)
		 VALUES ($1,$2,$3,$4,$5)
		 RETURNING `+datafeedParamColumns,
		datafeedID, paramKey,
		ptrOr(body.Type, model.FeederParamType_string),
		name, body.Value,
	))
	if err != nil {
		if utils.IsUniqueViolation(err) {
			return s.App.HttpResponseConflict(c, errs.ErrAlreadyExists)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "datafeed param created",
		"actor", snap.Login, "datafeed_id", datafeedID, "param_id", v.ParamID)

	s.notifyDatafeedConfigChanged(int64(datafeedID))

	return s.App.HttpResponseCreated(c, v)
}

// GetDatafeedParam returns one parameter row.
//
//	@Id			GetDatafeedParam
//	@Tags		Datafeeds
//	@Produce	json
//	@Param		id		path		int	true	"datafeed id"
//	@Param		paramId	path		int	true	"param id"
//	@Success	200		{object}	Response{data=ViewDatafeedParam}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/datafeeds/{id}/params/{paramId} [get]
func (s *HttpServer) GetDatafeedParam(c *fiber.Ctx) error {
	datafeedID, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	paramID, err := c.ParamsInt("paramId")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	v, err := scanViewDatafeedParam(s.DB.DB.QueryRow(c.UserContext(),
		`SELECT `+datafeedParamColumns+` FROM hst.datafeed_params
		  WHERE datafeed_id = $1 AND param_id = $2`, datafeedID, paramID))
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, v)
}

// UpdateDatafeedParam patches one parameter row.
//
//	@Id			UpdateDatafeedParam
//	@Tags		Datafeeds
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int					true	"datafeed id"
//	@Param		paramId	path		int					true	"param id"
//	@Param		body	body		UptDatafeedParam	true	"only the fields to change"
//	@Success	200		{object}	Response{data=ViewDatafeedParam}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	409		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/datafeeds/{id}/params/{paramId} [patch]
func (s *HttpServer) UpdateDatafeedParam(c *fiber.Ctx) error {
	datafeedID, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	paramID, err := c.ParamsInt("paramId")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	var body UptDatafeedParam
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	tag, err := s.DB.DB.Exec(c.UserContext(),
		`UPDATE hst.datafeed_params SET
		    param_key = COALESCE($3, param_key),
		    type      = COALESCE($4, type),
		    name      = COALESCE($5, name),
		    value     = COALESCE($6, value)
		  WHERE datafeed_id = $1 AND param_id = $2`,
		datafeedID, paramID, body.ParamKey, body.Type, body.Name, body.Value,
	)
	if err != nil {
		if utils.IsUniqueViolation(err) {
			return s.App.HttpResponseConflict(c, errs.ErrAlreadyExists)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if tag.RowsAffected() == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	v, err := scanViewDatafeedParam(s.DB.DB.QueryRow(c.UserContext(),
		`SELECT `+datafeedParamColumns+` FROM hst.datafeed_params
		  WHERE datafeed_id = $1 AND param_id = $2`, datafeedID, paramID))
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "datafeed param updated",
		"actor", snap.Login, "datafeed_id", datafeedID, "param_id", paramID)

	s.notifyDatafeedConfigChanged(int64(datafeedID))

	return s.App.HttpResponseOK(c, v)
}

// DeleteDatafeedParam removes one parameter row.
//
//	@Id			DeleteDatafeedParam
//	@Tags		Datafeeds
//	@Produce	json
//	@Param		id		path		int	true	"datafeed id"
//	@Param		paramId	path		int	true	"param id"
//	@Success	204		{object}	Response
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/datafeeds/{id}/params/{paramId} [delete]
func (s *HttpServer) DeleteDatafeedParam(c *fiber.Ctx) error {
	datafeedID, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	paramID, err := c.ParamsInt("paramId")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	tag, err := s.DB.DB.Exec(c.UserContext(),
		`DELETE FROM hst.datafeed_params WHERE datafeed_id = $1 AND param_id = $2`,
		datafeedID, paramID)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if tag.RowsAffected() == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeWarn, "datafeed param deleted",
		"actor", snap.Login, "datafeed_id", datafeedID, "param_id", paramID)

	s.notifyDatafeedConfigChanged(int64(datafeedID))

	return s.App.HttpResponseNoContent(c)
}

// ListDatafeedTranslates lists symbol translations for one data feed.
//
//	@Id			ListDatafeedTranslates
//	@Tags		Datafeeds
//	@Produce	json
//	@Param		id	path		int	true	"datafeed id"
//	@Success	200	{object}	Response{data=[]ViewDatafeedTranslate}
//	@Failure	400	{object}	Response
//	@Failure	403	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/datafeeds/{id}/translates [get]
func (s *HttpServer) ListDatafeedTranslates(c *fiber.Ctx) error {
	datafeedID, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	if err := datafeedExists(c, s, datafeedID); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	translates, err := loadDatafeedTranslates(c.UserContext(), s, int64(datafeedID))
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, translates)
}

// CreateDatafeedTranslate adds a symbol translation rule.
//
//	@Id			CreateDatafeedTranslate
//	@Tags		Datafeeds
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int						true	"datafeed id"
//	@Param		body	body		CrtDatafeedTranslate	true	"symbol required"
//	@Success	201		{object}	Response{data=ViewDatafeedTranslate}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	409		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/datafeeds/{id}/translates [post]
func (s *HttpServer) CreateDatafeedTranslate(c *fiber.Ctx) error {
	datafeedID, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	if err := datafeedExists(c, s, datafeedID); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	var body CrtDatafeedTranslate
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	symbol := strings.TrimSpace(body.Symbol)
	if symbol == "" {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	v, err := scanViewDatafeedTranslate(s.DB.DB.QueryRow(c.UserContext(),
		`INSERT INTO hst.datafeed_translates
		    (datafeed_id, symbol, source, bid_markup, ask_markup, digits)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 RETURNING `+datafeedTranslateColumns,
		datafeedID, symbol, body.Source, body.BidMarkup, body.AskMarkup, body.Digits,
	))
	if err != nil {
		if utils.IsUniqueViolation(err) {
			return s.App.HttpResponseConflict(c, errs.ErrAlreadyExists)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "datafeed translate created",
		"actor", snap.Login, "datafeed_id", datafeedID, "translate_id", v.TranslateID)

	s.notifyDatafeedConfigChanged(int64(datafeedID))

	return s.App.HttpResponseCreated(c, v)
}

// GetDatafeedTranslate returns one translation row.
//
//	@Id			GetDatafeedTranslate
//	@Tags		Datafeeds
//	@Produce	json
//	@Param		id			path		int	true	"datafeed id"
//	@Param		translateId	path		int	true	"translate id"
//	@Success	200			{object}	Response{data=ViewDatafeedTranslate}
//	@Failure	400			{object}	Response
//	@Failure	403			{object}	Response
//	@Failure	404			{object}	Response
//	@Failure	500			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/datafeeds/{id}/translates/{translateId} [get]
func (s *HttpServer) GetDatafeedTranslate(c *fiber.Ctx) error {
	datafeedID, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	translateID, err := c.ParamsInt("translateId")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	v, err := scanViewDatafeedTranslate(s.DB.DB.QueryRow(c.UserContext(),
		`SELECT `+datafeedTranslateColumns+` FROM hst.datafeed_translates
		  WHERE datafeed_id = $1 AND translate_id = $2`, datafeedID, translateID))
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, v)
}

// UpdateDatafeedTranslate patches one translation row.
//
//	@Id			UpdateDatafeedTranslate
//	@Tags		Datafeeds
//	@Accept		json
//	@Produce	json
//	@Param		id			path		int						true	"datafeed id"
//	@Param		translateId	path		int						true	"translate id"
//	@Param		body		body		UptDatafeedTranslate	true	"only the fields to change"
//	@Success	200			{object}	Response{data=ViewDatafeedTranslate}
//	@Failure	400			{object}	Response
//	@Failure	403			{object}	Response
//	@Failure	404			{object}	Response
//	@Failure	409			{object}	Response
//	@Failure	500			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/datafeeds/{id}/translates/{translateId} [patch]
func (s *HttpServer) UpdateDatafeedTranslate(c *fiber.Ctx) error {
	datafeedID, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	translateID, err := c.ParamsInt("translateId")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	var body UptDatafeedTranslate
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	tag, err := s.DB.DB.Exec(c.UserContext(),
		`UPDATE hst.datafeed_translates SET
		    symbol     = COALESCE($3, symbol),
		    source     = COALESCE($4, source),
		    bid_markup = COALESCE($5, bid_markup),
		    ask_markup = COALESCE($6, ask_markup),
		    digits     = COALESCE($7, digits)
		  WHERE datafeed_id = $1 AND translate_id = $2`,
		datafeedID, translateID,
		body.Symbol, body.Source, body.BidMarkup, body.AskMarkup, body.Digits,
	)
	if err != nil {
		if utils.IsUniqueViolation(err) {
			return s.App.HttpResponseConflict(c, errs.ErrAlreadyExists)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if tag.RowsAffected() == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	v, err := scanViewDatafeedTranslate(s.DB.DB.QueryRow(c.UserContext(),
		`SELECT `+datafeedTranslateColumns+` FROM hst.datafeed_translates
		  WHERE datafeed_id = $1 AND translate_id = $2`, datafeedID, translateID))
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "datafeed translate updated",
		"actor", snap.Login, "datafeed_id", datafeedID, "translate_id", translateID)

	s.notifyDatafeedConfigChanged(int64(datafeedID))

	return s.App.HttpResponseOK(c, v)
}

// DeleteDatafeedTranslate removes one translation row.
//
//	@Id			DeleteDatafeedTranslate
//	@Tags		Datafeeds
//	@Produce	json
//	@Param		id			path		int	true	"datafeed id"
//	@Param		translateId	path		int	true	"translate id"
//	@Success	204			{object}	Response
//	@Failure	400			{object}	Response
//	@Failure	403			{object}	Response
//	@Failure	404			{object}	Response
//	@Failure	500			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/datafeeds/{id}/translates/{translateId} [delete]
func (s *HttpServer) DeleteDatafeedTranslate(c *fiber.Ctx) error {
	datafeedID, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	translateID, err := c.ParamsInt("translateId")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	tag, err := s.DB.DB.Exec(c.UserContext(),
		`DELETE FROM hst.datafeed_translates WHERE datafeed_id = $1 AND translate_id = $2`,
		datafeedID, translateID)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if tag.RowsAffected() == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeWarn, "datafeed translate deleted",
		"actor", snap.Login, "datafeed_id", datafeedID, "translate_id", translateID)

	s.notifyDatafeedConfigChanged(int64(datafeedID))

	return s.App.HttpResponseNoContent(c)
}

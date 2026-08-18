package admin

import (
	"errors"
	"strings"
	"time"

	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// The default Market Watch: the instruments a trading account sees on its first launch, before
// it has customized anything — the platform's counterpart of the symbol set a stock MT5 build
// ships. The trader api reads it on demand when it finds an account with no watchlist, so a
// change here only affects accounts that have not launched yet.

// defaultMarketWatchKey names the setting in hst.settings.
const defaultMarketWatchKey = "default_market_watch"

// defaultMarketWatchMax is the settings column width; the names are stored comma separated.
const defaultMarketWatchMax = 255

// UptDefaultMarketWatch replaces the default set, in the order given.
type UptDefaultMarketWatch struct {
	Symbols []string `json:"symbols" validate:"required,dive,min=1,max=32"`
}

// ViewDefaultMarketWatch is the default set as stored.
type ViewDefaultMarketWatch struct {
	Symbols   []string `json:"symbols"`
	UpdatedAt int64    `json:"updated_at"`
}

var errDefaultMarketWatchTooLong = errors.New("default market watch holds too many symbols")
var errDefaultMarketWatchUnknown = errors.New("default market watch names a symbol that does not exist")

// GetDefaultMarketWatch reports the instruments a new account starts with.
//
//	@Id			GetDefaultMarketWatch
//	@Tags		System
//	@Produce	json
//	@Success	200	{object}	Response{data=ViewDefaultMarketWatch}
//	@Failure	403	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/system/market-watch [get]
func (s *Server) GetDefaultMarketWatch(c *fiber.Ctx) error {
	v := ViewDefaultMarketWatch{Symbols: []string{}}

	var value string
	err := s.DB.DB.QueryRow(c.UserContext(),
		`SELECT value, updated_at FROM hst.settings WHERE key = $1`, defaultMarketWatchKey).
		Scan(&value, &v.UpdatedAt)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if value != "" {
		v.Symbols = strings.Split(value, ",")
	}

	return s.App.HttpResponseOK(c, v)
}

// UpdateDefaultMarketWatch replaces the instruments a new account starts with.
//
//	@Id			UpdateDefaultMarketWatch
//	@Tags		System
//	@Accept		json
//	@Produce	json
//	@Param		body	body		UptDefaultMarketWatch	true	"the symbol names, in display order"
//	@Success	200		{object}	Response{data=ViewDefaultMarketWatch}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/system/market-watch [put]
func (s *Server) UpdateDefaultMarketWatch(c *fiber.Ctx) error {
	var body UptDefaultMarketWatch
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	names := make([]string, 0, len(body.Symbols))
	seen := map[string]bool{}
	for _, name := range body.Symbols {
		name = strings.TrimSpace(name)
		if name != "" && !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}

	value := strings.Join(names, ",")
	if len(value) > defaultMarketWatchMax {
		return s.App.HttpResponseBadRequest(c, errDefaultMarketWatchTooLong)
	}

	// a default naming a symbol the platform does not carry would seed dead entries
	if len(names) > 0 {
		var known int
		err := s.DB.DB.QueryRow(c.UserContext(),
			`SELECT count(*) FROM hst.symbols WHERE symbol = ANY($1)`, names).Scan(&known)
		if err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		if known != len(names) {
			return s.App.HttpResponseBadRequest(c, errDefaultMarketWatchUnknown)
		}
	}

	v := ViewDefaultMarketWatch{Symbols: names}
	err := s.DB.DB.QueryRow(c.UserContext(),
		`INSERT INTO hst.settings (key, value, updated_at) VALUES ($1,$2,$3)
		 ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = $3
		 RETURNING updated_at`, defaultMarketWatchKey, value, time.Now().UnixNano()).
		Scan(&v.UpdatedAt)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, v)
}

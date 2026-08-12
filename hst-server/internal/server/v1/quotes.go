package v1

import (
	"encoding/json"
	"errors"
	"math"
	"time"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/journal"
	"hstserver/pkg/logger"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// CrtQuote is a dealer-thrown quote for one symbol.
type CrtQuote struct {
	Symbol string  `json:"symbol" validate:"required,max=32"`
	Bid    float64 `json:"bid" validate:"required,gt=0"`
	Ask    float64 `json:"ask" validate:"required,gt=0"`
}

// wireTick mirrors hst-quote's published tick so every consumer parses it the same way.
type wireTick struct {
	DatafeedID int64     `json:"datafeed_id"`
	SymbolID   int64     `json:"symbol_id"`
	Symbol     string    `json:"symbol"`
	Source     string    `json:"source"`
	Bid        float64   `json:"bid"`
	Ask        float64   `json:"ask"`
	Time       time.Time `json:"time"`
}

// ThrowQuote publishes a manual quote on the tick stream, as if the feed had sent it.
//
//	@Id			ThrowQuote
//	@Tags		Quotes
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CrtQuote	true	"the quote"
//	@Success	200		{object}	Response{data=CrtQuote}
//	@Failure	400		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/quotes [post]
func (s *HttpServer) ThrowQuote(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var body CrtQuote
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}
	if body.Ask < body.Bid {
		return s.App.HttpResponseBadRequest(c, errors.New("ask must be greater than or equal to bid"))
	}

	var symbolId int64
	var digits int32
	err := s.DB.DB.QueryRow(c.UserContext(),
		`SELECT symbol_id, digits FROM hst.symbols WHERE symbol = $1`, body.Symbol).
		Scan(&symbolId, &digits)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	pow := math.Pow10(int(digits))
	body.Bid = math.Round(body.Bid*pow) / pow
	body.Ask = math.Round(body.Ask*pow) / pow

	tick := &wireTick{
		SymbolID: symbolId,
		Symbol:   body.Symbol,
		Source:   body.Symbol,
		Bid:      body.Bid,
		Ask:      body.Ask,
		Time:     time.Now().UTC(),
	}
	if _, err := s.publish(model.RootQuote+".tick."+body.Symbol, tick); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// the last-quote cache feeds liveness and exposure rates, so a thrown quote lands there too
	if raw, err := json.Marshal(tick); err == nil {
		s.Redis.Client.Set(c.UserContext(), "hstquote:last:"+body.Symbol, raw, 0)
	}

	s.Log.Log(logger.TypeTrade, logger.CodeAtt, "quote thrown",
		"actor", snap.Login, "symbol", body.Symbol, "bid", body.Bid, "ask", body.Ask)
	s.JournalEntry(c, model.JournalType_symbols, logger.CodeAtt,
		journal.QuoteThrownMsg(body.Symbol, int(digits), body.Bid, body.Ask), body)

	return s.App.HttpResponseOK(c, body)
}

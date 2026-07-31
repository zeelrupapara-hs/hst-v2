package v1

import (
	"context"
	"errors"
	"strconv"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/events"
	"hstserver/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

const workerDatafeedColumns = `datafeed_id, name, module, enable, mode,
	feed_server, feed_login, feed_password, timeout_reconnect`

type workerRowScanner interface {
	Scan(dest ...any) error
}

func scanWorkerDatafeed(row workerRowScanner) (*events.WorkerDatafeedConfig, error) {
	cfg := &events.WorkerDatafeedConfig{}
	err := row.Scan(
		&cfg.DatafeedID, &cfg.Name, &cfg.Module, &cfg.Enable, &cfg.Mode,
		&cfg.FeedServer, &cfg.FeedLogin, &cfg.FeedPassword, &cfg.TimeoutReconnect,
	)
	return cfg, err
}

func (s *HttpServer) loadWorkerDatafeedConfig(ctx context.Context, datafeedID int64) (*events.WorkerDatafeedConfig, error) {
	cfg, err := scanWorkerDatafeed(s.DB.DB.QueryRow(ctx,
		`SELECT `+workerDatafeedColumns+` FROM hst.datafeeds WHERE datafeed_id = $1`, datafeedID))
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
	cfg.Params = make([]events.WorkerParam, 0, len(params))
	for _, p := range params {
		cfg.Params = append(cfg.Params, events.WorkerParam{
			ParamKey: p.ParamKey,
			Value:    p.Value,
		})
	}

	translates, err := loadWorkerTranslates(ctx, s, datafeedID)
	if err != nil {
		return nil, err
	}
	cfg.Translates = translates
	return cfg, nil
}

func (s *HttpServer) listWorkerDatafeedConfigs(ctx context.Context, mode int32) ([]events.WorkerDatafeedConfig, error) {
	rows, err := s.DB.DB.Query(ctx,
		`SELECT `+workerDatafeedColumns+`
		   FROM hst.datafeeds
		  WHERE enable = 1
		    AND ($1 = 0 OR (mode & $1) = $1)
		  ORDER BY datafeed_id`, mode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	out := []events.WorkerDatafeedConfig{}
	for rows.Next() {
		cfg, err := scanWorkerDatafeed(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *cfg)
		ids = append(ids, cfg.DatafeedID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return out, nil
	}

	paramRows, err := s.DB.DB.Query(ctx,
		`SELECT datafeed_id, param_key, value
		   FROM hst.datafeed_params
		  WHERE datafeed_id = ANY($1)
		  ORDER BY datafeed_id, param_id`, ids)
	if err != nil {
		return nil, err
	}
	defer paramRows.Close()

	paramsByFeed := map[int64][]events.WorkerParam{}
	for paramRows.Next() {
		var feedID int64
		var p events.WorkerParam
		if err := paramRows.Scan(&feedID, &p.ParamKey, &p.Value); err != nil {
			return nil, err
		}
		paramsByFeed[feedID] = append(paramsByFeed[feedID], p)
	}
	if err := paramRows.Err(); err != nil {
		return nil, err
	}

	trByFeed, err := loadWorkerTranslatesByFeedIDs(ctx, s, ids)
	if err != nil {
		return nil, err
	}

	for i := range out {
		id := out[i].DatafeedID
		out[i].Params = paramsByFeed[id]
		out[i].Translates = trByFeed[id]
	}
	return out, nil
}

func (s *HttpServer) publishWorkerConfigSnapshot(ctx context.Context, datafeedID int64) {
	cfg, err := s.loadWorkerDatafeedConfig(ctx, datafeedID)
	if err != nil {
		if !errors.Is(err, errs.ErrNotFound) {
			s.Log.Log(logger.TypeNet, logger.CodeWarn, "worker config load failed",
				"datafeed_id", datafeedID, "error", err.Error())
		}
		return
	}
	if err := events.PublishConfigSnapshot(s.Nats, *cfg); err != nil {
		s.Log.Log(logger.TypeNet, logger.CodeWarn, "worker config publish failed",
			"datafeed_id", datafeedID, "error", err.Error())
	}
}

// ListInternalDatafeeds returns enabled feeds for worker services.
func (s *HttpServer) ListInternalDatafeeds(c *fiber.Ctx) error {
	mode := int32(0)
	if raw := c.Query("mode"); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			return s.App.HttpResponseBadQueryParams(c, err)
		}
		mode = int32(v)
		if mode != 0 {
			if err := validateFeederMode(model.FeederFlags(mode)); err != nil {
				return s.App.HttpResponseBadRequest(c, err)
			}
		}
	}

	out, err := s.listWorkerDatafeedConfigs(c.UserContext(), mode)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	return s.App.HttpResponseOK(c, out)
}

// GetInternalDatafeed returns one feed config for worker services.
func (s *HttpServer) GetInternalDatafeed(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	cfg, err := s.loadWorkerDatafeedConfig(c.UserContext(), int64(id))
	if errors.Is(err, errs.ErrNotFound) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	return s.App.HttpResponseOK(c, cfg)
}

func loadWorkerTranslates(ctx context.Context, s *HttpServer, datafeedID int64) ([]events.WorkerTranslate, error) {
	rows, err := s.DB.DB.Query(ctx,
		`SELECT t.translate_id, COALESCE(s.symbol_id, 0), t.symbol, t.source,
		        t.bid_markup, t.ask_markup, t.digits
		   FROM hst.datafeed_translates t
		   LEFT JOIN hst.symbols s ON s.symbol = t.symbol
		  WHERE t.datafeed_id = $1
		  ORDER BY t.translate_id`, datafeedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []events.WorkerTranslate{}
	for rows.Next() {
		var t events.WorkerTranslate
		if err := rows.Scan(&t.TranslateID, &t.SymbolID, &t.Symbol, &t.Source,
			&t.BidMarkup, &t.AskMarkup, &t.Digits); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func loadWorkerTranslatesByFeedIDs(ctx context.Context, s *HttpServer, ids []int64) (map[int64][]events.WorkerTranslate, error) {
	rows, err := s.DB.DB.Query(ctx,
		`SELECT t.datafeed_id, t.translate_id, COALESCE(s.symbol_id, 0), t.symbol, t.source,
		        t.bid_markup, t.ask_markup, t.digits
		   FROM hst.datafeed_translates t
		   LEFT JOIN hst.symbols s ON s.symbol = t.symbol
		  WHERE t.datafeed_id = ANY($1)
		  ORDER BY t.datafeed_id, t.translate_id`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[int64][]events.WorkerTranslate{}
	for rows.Next() {
		var feedID int64
		var t events.WorkerTranslate
		if err := rows.Scan(&feedID, &t.TranslateID, &t.SymbolID, &t.Symbol, &t.Source,
			&t.BidMarkup, &t.AskMarkup, &t.Digits); err != nil {
			return nil, err
		}
		out[feedID] = append(out[feedID], t)
	}
	return out, rows.Err()
}

package v1

import (
	"hstserver/config"
	"hstserver/internal/middleware"
	"hstserver/pkg/cache"
	"hstserver/pkg/db"
	"hstserver/pkg/http"
	"hstserver/pkg/influxdb"
	"hstserver/pkg/journal"
	"hstserver/pkg/logger"
	"hstserver/pkg/mailer"
	"hstserver/pkg/nats"
	"hstserver/pkg/oauth2"
	"hstserver/pkg/redis"
	"hstserver/pkg/ws"

	"github.com/go-playground/validator/v10"
)

// Response is the envelope every endpoint returns, aliased here for swagger.
type Response = http.HttpResponse

// CacheStats is the session cache report, aliased for the same reason.
type CacheStats = cache.Stats

type HttpServer struct {
	// Config
	Cfg *config.Config
	// Middleware
	Middleware *middleware.Middleware
	// Fiber app
	App *http.App
	// Postgres DB
	DB *db.PostgresDB
	// zab logger for log to files and stdout
	Log *logger.Logger
	// Nats
	Nats *nats.Nats
	// Redis
	Redis *redis.Redis
	// OAuth2.0
	OAuth2 *oauth2.OAuth2
	// Validator
	Validate *validator.Validate
	// Hub holds the live websocket connections
	Hub *ws.Hub
	// Journal records what happened, for the back office to query
	Journal *journal.Journal
	// History reads the tick store back as chart bars
	History *influxdb.Reader
	// Mailer queues and sends outgoing email through the configured mail server
	Mailer *mailer.Mailer
}

func NewHTTP(app *http.App, database *db.PostgresDB, log *logger.Logger, nats *nats.Nats, rds *redis.Redis, middleware *middleware.Middleware, oauth *oauth2.OAuth2, cfg *config.Config, validate *validator.Validate) *HttpServer {

	h := &HttpServer{
		Middleware: middleware,
		Hub:        ws.NewHub(log),
		Journal:    journal.New(database, nats, log),
		Mailer:     mailer.New(database, log, cfg.Mail.TemplatesDir, cfg.Mail.DrainInterval),
		App:        app,
		DB:         database,
		Log:        log,
		Nats:       nats,
		Redis:      rds,
		OAuth2:     oauth,
		Validate:   validate,
		Cfg:        cfg,
	}

	reader, err := influxdb.NewReader(cfg, log)
	if err != nil {
		// a chart that cannot load is not a reason to refuse to trade
		log.Log(logger.TypeSys, logger.CodeErr, "chart history unavailable", "error", err.Error())
	}
	h.History = reader

	return h
}

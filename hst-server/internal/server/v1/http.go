package v1

import (
	"hstserver/config"
	"hstserver/internal/middleware"
	"hstserver/pkg/db"
	"hstserver/pkg/http"
	"hstserver/pkg/logger"
	"hstserver/pkg/nats"
	"hstserver/pkg/redis"

	"github.com/go-playground/validator/v10"
)

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
	// Validator
	Validate *validator.Validate
}

func NewHTTP(app *http.App, database *db.PostgresDB, log *logger.Logger, nats *nats.Nats, rds *redis.Redis, middleware *middleware.Middleware, cfg *config.Config, validate *validator.Validate) *HttpServer {

	h := &HttpServer{
		Middleware: middleware,
		App:        app,
		DB:         database,
		Log:        log,
		Nats:       nats,
		Redis:      rds,
		Validate:   validate,
		Cfg:        cfg,
	}

	return h
}

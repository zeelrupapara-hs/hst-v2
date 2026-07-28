package v1

import (
	"hstserver/config"
	"hstserver/internal/middleware"
	"hstserver/pkg/db"
	"hstserver/pkg/http"
	"hstserver/pkg/logger"
	"hstserver/pkg/nats"
	"hstserver/pkg/oauth2"
	"hstserver/pkg/redis"

	"github.com/go-playground/validator/v10"
)

// ErrorResponse is the failure envelope, aliased into this package so the
// swagger annotations can resolve it without every handler file importing
// pkg/http purely for a comment.
type ErrorResponse = http.HttpResponse

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
}

func NewHTTP(app *http.App, database *db.PostgresDB, log *logger.Logger, nats *nats.Nats, rds *redis.Redis, middleware *middleware.Middleware, oauth *oauth2.OAuth2, cfg *config.Config, validate *validator.Validate) *HttpServer {

	h := &HttpServer{
		Middleware: middleware,
		App:        app,
		DB:         database,
		Log:        log,
		Nats:       nats,
		Redis:      rds,
		OAuth2:     oauth,
		Validate:   validate,
		Cfg:        cfg,
	}

	return h
}

package middleware

import (
	"hstserver/config"
	"hstserver/pkg/db"
	"hstserver/pkg/http"
	"hstserver/pkg/logger"
	"hstserver/pkg/nats"
	"hstserver/pkg/oauth2"
	"hstserver/pkg/redis"
)

type Middleware struct {
	// Fiber App
	App *http.App
	// Postgres DB
	DB *db.PostgresDB
	// Authentication
	OAuth2 *oauth2.OAuth2
	// Redis
	Redis *redis.Redis
	// zab logger for log to files and stdout
	Log *logger.Logger
	// Nats
	Nats *nats.Nats
	// Config
	Cfg *config.Config
}

func NewMiddleware(app *http.App, db *db.PostgresDB, oauth *oauth2.OAuth2, rds *redis.Redis, log *logger.Logger, nats *nats.Nats, cfg *config.Config) *Middleware {

	m := &Middleware{
		App:    app,
		DB:     db,
		OAuth2: oauth,
		Redis:  rds,
		Log:    log,
		Nats:   nats,
		Cfg:    cfg,
	}

	return m
}

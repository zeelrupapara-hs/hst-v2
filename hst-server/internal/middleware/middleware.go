package middleware

import (
	"hstserver/pkg/db"
	"hstserver/pkg/http"
	"hstserver/pkg/logger"
	"hstserver/pkg/nats"
)

type Middleware struct {
	// Fiber App
	App *http.App
	// Postgres DB
	DB *db.PostgresDB
	// zab logger for log to files and stdout
	Log *logger.Logger
	// Nats
	Nats *nats.Nats
}

func NewMiddleware(app *http.App, db *db.PostgresDB, log *logger.Logger, nats *nats.Nats) *Middleware {

	m := &Middleware{
		App:  app,
		DB:   db,
		Log:  log,
		Nats: nats,
	}

	return m
}

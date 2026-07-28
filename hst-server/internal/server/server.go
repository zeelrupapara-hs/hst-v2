package server

import (
	"github.com/go-playground/validator/v10"

	// import local pkg
	"hstserver/config"
	"hstserver/internal/middleware"
	v1 "hstserver/internal/server/v1"
	"hstserver/pkg/db"
	"hstserver/pkg/http"
	"hstserver/pkg/logger"
	"hstserver/pkg/nats"
	"hstserver/pkg/redis"
)

// Now server is a generic builder with http server as v1 every http, ws or any public interface should
// move to v1 including routes,controlers
// also model/v1 the same
// all resources
// remove any resources here that was in v1 and duplicated
type Server struct {
	App *http.App
	// Middleware
	Middleware *middleware.Middleware
	// Our v1 Http API
	Web *v1.HttpServer
	// zab logger for log to files and stdout
	Log *logger.Logger
	// DB
	DB *db.PostgresDB
	// Nats
	Nats *nats.Nats
	// Redis
	Redis *redis.Redis
	// Config
	Cfg *config.Config
}

func NewServer(log *logger.Logger, database *db.PostgresDB, nats *nats.Nats, rds *redis.Redis, validate *validator.Validate, cfg *config.Config) *Server {
	// fiber instence
	app := http.NewApp(cfg, log)

	// Middleware
	newMiddleware := middleware.NewMiddleware(app, database, log, nats)

	// v1 http server
	web := v1.NewHTTP(app, database, log, nats, rds, newMiddleware, cfg, validate)

	return &Server{
		App:        app,
		Middleware: newMiddleware,
		Web:        web,
		Log:        log,
		DB:         database,
		Nats:       nats,
		Redis:      rds,
		Cfg:        cfg,
	}
}

// Run will register the routes and start listening
func (s *Server) Run() error {
	// register all routes
	s.RegisterRoutes()

	addr := s.Cfg.HTTP.Host + ":" + s.Cfg.HTTP.Port
	s.Log.Logger.Info("http server listening on ", addr)

	return s.App.Listen(addr)
}

// Shutdown drains in-flight requests, capped by ShutdownTimeout.
// A plain Shutdown() would wait forever on a hung client.
func (s *Server) Shutdown() error {
	s.Log.Logger.Infow("draining in-flight requests",
		"timeout", s.Cfg.HTTP.ShutdownTimeout.String())
	return s.App.ShutdownWithTimeout(s.Cfg.HTTP.ShutdownTimeout)
}

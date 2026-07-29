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
	"hstserver/pkg/oauth2"
	"hstserver/pkg/redis"
)

// Server wires the app together; every public interface lives under v1.
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
	// OAuth2.0
	OAuth2 *oauth2.OAuth2
	// Config
	Cfg *config.Config
}

func NewServer(log *logger.Logger, database *db.PostgresDB, nats *nats.Nats, rds *redis.Redis, oauth *oauth2.OAuth2, validate *validator.Validate, cfg *config.Config) *Server {
	// fiber instence
	app := http.NewApp(cfg, log)

	// Middleware
	newMiddleware := middleware.NewMiddleware(app, database, oauth, rds, log, nats, cfg)

	// v1 http server
	web := v1.NewHTTP(app, database, log, nats, rds, newMiddleware, oauth, cfg, validate)

	return &Server{
		App:        app,
		Middleware: newMiddleware,
		Web:        web,
		Log:        log,
		DB:         database,
		Nats:       nats,
		Redis:      rds,
		OAuth2:     oauth,
		Cfg:        cfg,
	}
}

// Run will register the routes and start listening
func (s *Server) Run() error {
	// register all routes
	s.RegisterRoutes()

	addr := s.Cfg.HTTP.Host + ":" + s.Cfg.HTTP.Port

	// tls here is for serving https directly; behind an ingress that terminates
	// it, leave the pair empty and let the proxy do it
	if s.Cfg.HTTP.TlsCert != "" {
		s.Log.Logger.Info("https server listening on ", addr)
		return s.App.ListenTLS(addr, s.Cfg.HTTP.TlsCert, s.Cfg.HTTP.TlsKey)
	}

	s.Log.Logger.Info("http server listening on ", addr)

	return s.App.Listen(addr)
}

// Shutdown drains in-flight requests, capped by ShutdownTimeout.
func (s *Server) Shutdown() error {
	s.Log.Logger.Infow("draining in-flight requests",
		"timeout", s.Cfg.HTTP.ShutdownTimeout.String())
	return s.App.ShutdownWithTimeout(s.Cfg.HTTP.ShutdownTimeout)
}

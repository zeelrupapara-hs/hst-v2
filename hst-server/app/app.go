package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"hstserver/config"
	"hstserver/internal/server"
	"hstserver/pkg/db"
	"hstserver/pkg/logger"
	"hstserver/pkg/nats"
	"hstserver/pkg/oauth2"
	"hstserver/pkg/redis"
	"hstserver/pkg/seed"

	"github.com/go-playground/validator/v10"
)

var (
	// this name is one time only
	service = "hstserver"

	// this change as per git -tag -v everytime this will go into testing
	// v1.0.0 Major.Minor.Batch or bug
	version = "v1.0.0"
)

// Start boots the service and blocks until it has shut down cleanly.
func Start() {
	// run() owns the resources so its defers always run.
	// Fatalf deeper down would skip them.
	if code := Run(); code != 0 {
		os.Exit(code)
	}
}

func Run() int {
	// config instant
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid configuration:", err)
		return 1
	}

	// pass to logger handler instant
	log, err := logger.NewLogger(cfg)
	if err != nil {
		// no logger yet, so stderr is all we have
		fmt.Fprintln(os.Stderr, "failed to init logger:", err)
		return 1
	}
	// flush buffered entries and release the day file last of all
	defer func() {
		_ = log.Sync()
		_ = log.Close()
	}()

	// Start logging
	log.Logger.Info("Logging started for service: ", service+"@"+version)

	// Postgres DB
	database, err := db.NewPostgresDB(cfg)
	if err != nil {
		log.Logger.Errorf("failed to connect to postgres %v", err)
		return 1
	}
	// closed after the http server drains
	defer func() {
		log.Logger.Info("closing postgres pool")
		database.DB.Close()
	}()

	log.Logger.Info("postgres connected")

	// Migrate the schema
	if err := database.Migrate(); err != nil {
		log.Logger.Errorf("failed to migrate postgres %v", err)
		return 1
	}

	// Nats
	natsClient, err := nats.NewNatClient(cfg, log)
	if err != nil {
		log.Logger.Errorf("failed to connect to nats %v", err)
		return 1
	}
	// drained before the pool closes
	defer func() {
		log.Logger.Info("draining nats connection")
		if err := natsClient.Close(); err != nil {
			log.Logger.Errorf("failed to drain nats %v", err)
		}
	}()

	log.Logger.Info("nats connected")

	// Redis
	redisClient, err := redis.NewRedisClient(cfg, log)
	if err != nil {
		log.Logger.Errorf("failed to connect to redis %v", err)
		return 1
	}
	defer func() {
		log.Logger.Info("closing redis pool")
		if err := redisClient.Close(); err != nil {
			log.Logger.Errorf("failed to close redis %v", err)
		}
	}()

	log.Logger.Info("redis connected")

	// Authentication
	oauth, err := oauth2.NewOAuth2(redisClient, database, cfg, log)
	if err != nil {
		log.Logger.Errorf("failed to init auth %v", err)
		return 1
	}
	// stopped before redis closes, its workers publish on shutdown
	defer func() {
		log.Logger.Info("stopping auth workers")
		oauth.Close()
	}()

	// starting rows a fresh install needs, each one a no op if already applied
	if err := seed.New(database, oauth.Hasher, log, cfg).Run(context.Background()); err != nil {
		log.Logger.Errorf("failed to seed %v", err)
		return 1
	}

	// Validator
	validate := validator.New()

	// build the server
	srv := server.NewServer(log, database, natsClient, redisClient, oauth, validate, cfg)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// buffered so the goroutine never leaks on the signal path
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Run()
	}()

	select {
	case err := <-errCh:
		if err != nil {
			log.Logger.Errorf("failed to run the server %v", err)
			return 1
		}

	case signal := <-quit:
		log.Logger.Infow("shutdown signal received", "signal", signal.String())

		if err := srv.Shutdown(); err != nil {
			// timeout expired with requests still open
			log.Logger.Errorf("graceful shutdown failed %v", err)
			return 1
		}

		// wait for Listen to unblock before closing the pool
		if err := <-errCh; err != nil {
			log.Logger.Errorf("server exited with error %v", err)
			return 1
		}
		log.Logger.Info("all in-flight requests drained")
	}

	log.Logger.Info("server stopped for service: ", service+"@"+version)
	return 0
}
